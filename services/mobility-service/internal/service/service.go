package service

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/adedaryorh/logistics-platform/pkg/observability"
	platformredis "github.com/adedaryorh/logistics-platform/pkg/redis"
	mobilityh3 "github.com/adedaryorh/logistics-platform/services/mobility-service/internal/h3"
	"github.com/adedaryorh/logistics-platform/services/mobility-service/internal/routing"
	"github.com/adedaryorh/logistics-platform/services/mobility-service/internal/tracking"
)

type DriverLocation struct {
	DriverID    string    `json:"driver_id"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	H3Cell      string    `json:"h3_cell"`
	Rating      float64   `json:"rating"`
	TotalTrips  int       `json:"total_trips"`
	VehicleType string    `json:"vehicle_type"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TrackingSession struct {
	SessionID     string         `json:"session_id"`
	OrderID       string         `json:"order_id"`
	Status        string         `json:"status"`
	Route         *routing.Route `json:"route,omitempty"`
	StartedAt     time.Time      `json:"started_at"`
	LastUpdatedAt time.Time      `json:"last_updated_at"`
	EndedAt       *time.Time     `json:"ended_at,omitempty"`
}

type Service struct {
	cfg         *platformconfig.Config
	hub         *tracking.Hub
	mu          sync.RWMutex
	drivers     map[string]DriverLocation
	geofences   map[string][]string
	sessions    map[string]string
	sessionData map[string]*TrackingSession
	graph       *routing.Graph
	redis       platformredis.Store
}

func New(cfg *platformconfig.Config) *Service {
	a := &routing.Node{ID: "a", Lat: 6.5244, Lng: 3.3792, H3: mobilityh3.LatLngToCell(6.5244, 3.3792, 9)}
	b := &routing.Node{ID: "b", Lat: 6.5300, Lng: 3.3840, H3: mobilityh3.LatLngToCell(6.5300, 3.3840, 9)}
	c := &routing.Node{ID: "c", Lat: 6.5360, Lng: 3.3900, H3: mobilityh3.LatLngToCell(6.5360, 3.3900, 9)}

	return &Service{
		cfg: cfg,
		hub: tracking.NewHub(),
		drivers: map[string]DriverLocation{
			"driver-1": {DriverID: "driver-1", Lat: 6.5244, Lng: 3.3792, H3Cell: a.H3, Rating: 4.8, TotalTrips: 300, VehicleType: "ride", UpdatedAt: time.Now().UTC()},
			"driver-2": {DriverID: "driver-2", Lat: 6.5300, Lng: 3.3840, H3Cell: b.H3, Rating: 4.6, TotalTrips: 200, VehicleType: "ride", UpdatedAt: time.Now().UTC()},
		},
		geofences: map[string][]string{
			"lagos-service-area": {mobilityh3.LatLngToCell(6.5244, 3.3792, 8), mobilityh3.LatLngToCell(6.5300, 3.3840, 8)},
		},
		sessions:    map[string]string{},
		sessionData: map[string]*TrackingSession{},
		graph: &routing.Graph{
			Nodes: map[string]*routing.Node{"a": a, "b": b, "c": c},
			Edges: map[string][]*routing.Edge{
				"a": {{To: b, WeightS: 40, DistanceM: 800}},
				"b": {{To: c, WeightS: 55, DistanceM: 1100}},
				"c": {},
			},
		},
		redis: platformredis.NewStoreFromConfig(cfg),
	}
}

func (s *Service) FindNearbyDrivers(ctx context.Context, lat, lng float64, k int, vehicleType string) ([]DriverLocation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cell := mobilityh3.LatLngToCell(lat, lng, 8)
	allowed := map[string]struct{}{}
	for _, nearbyCell := range mobilityh3.KRingCells(cell, k) {
		allowed[nearbyCell] = struct{}{}
	}

	drivers := make([]DriverLocation, 0)
	seen := map[string]struct{}{}
	if s.redis != nil {
		for _, nearbyCell := range mobilityh3.KRingCells(cell, k) {
			ids, err := s.redis.SMembers(ctx, "h3:8:"+nearbyCell)
			if err == nil {
				for _, id := range ids {
					seen[id] = struct{}{}
				}
			}
		}
	}
	for _, driver := range s.drivers {
		if vehicleType != "" && driver.VehicleType != vehicleType {
			continue
		}
		if _, ok := allowed[mobilityh3.LatLngToCell(driver.Lat, driver.Lng, 8)]; !ok && len(seen) == 0 {
			continue
		}
		if len(seen) > 0 {
			if _, ok := seen[driver.DriverID]; !ok {
				continue
			}
		}
		drivers = append(drivers, driver)
	}

	sort.SliceStable(drivers, func(i, j int) bool {
		if drivers[i].Rating == drivers[j].Rating {
			return drivers[i].TotalTrips > drivers[j].TotalTrips
		}
		return drivers[i].Rating > drivers[j].Rating
	})
	observability.IncCounter("nearby_driver_queries_total", 1, map[string]string{"vehicle_type": vehicleType})
	observability.ObserveHistogram("nearby_driver_query_results", float64(len(drivers)), map[string]string{"vehicle_type": vehicleType})

	return drivers, nil
}

func (s *Service) GetH3Cell(ctx context.Context, lat, lng float64, resolution int) string {
	return mobilityh3.LatLngToCell(lat, lng, resolution)
}

func (s *Service) IsInGeofence(ctx context.Context, cell string, geofenceName string) bool {
	return mobilityh3.IsInGeofenceCells(cell, s.geofences[geofenceName])
}

func (s *Service) GetSurgeMultiplier(ctx context.Context, cell string) float64 {
	if mobilityh3.IsInGeofenceCells(cell, s.geofences["lagos-service-area"]) {
		return 1.15
	}
	return 1
}

func (s *Service) GetRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (*routing.Route, error) {
	from := s.closestNode(fromLat, fromLng)
	to := s.closestNode(toLat, toLng)
	if from == nil || to == nil {
		return nil, platformerrors.ErrNotFound
	}
	return s.graph.AStar(from, to)
}

func (s *Service) StartTrackingSession(ctx context.Context, orderID string) (string, error) {
	if orderID == "" {
		return "", platformerrors.ErrBadRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionID := "tracking-" + orderID
	s.sessions[orderID] = sessionID
	now := time.Now().UTC()
	s.sessionData[orderID] = &TrackingSession{
		SessionID:     sessionID,
		OrderID:       orderID,
		Status:        "active",
		StartedAt:     now,
		LastUpdatedAt: now,
	}
	observability.IncCounter("tracking_sessions_started_total", 1, nil)
	observability.SetGauge("tracking_sessions_active", float64(len(s.sessions)), nil)
	return sessionID, nil
}

func (s *Service) EndTrackingSession(ctx context.Context, orderID string) error {
	if orderID == "" {
		return platformerrors.ErrBadRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, orderID)
	if session, ok := s.sessionData[orderID]; ok {
		now := time.Now().UTC()
		session.Status = "ended"
		session.EndedAt = &now
		session.LastUpdatedAt = now
	}
	observability.SetGauge("tracking_sessions_active", float64(len(s.sessions)), nil)
	return nil
}

func (s *Service) Subscribe(orderID string) (<-chan tracking.DriverLocationMsg, func()) {
	ch := make(chan tracking.DriverLocationMsg, 8)
	s.hub.Subscribe(orderID, ch)
	return ch, func() {
		s.hub.Unsubscribe(orderID, ch)
		close(ch)
	}
}

func (s *Service) UpdateDriverLocation(ctx context.Context, msg tracking.DriverLocationMsg) error {
	if msg.OrderID == "" || msg.DriverID == "" {
		return platformerrors.ErrBadRequest
	}
	s.mu.Lock()
	driver, ok := s.drivers[msg.DriverID]
	if !ok {
		s.mu.Unlock()
		return platformerrors.ErrNotFound
	}
	filter := tracking.NewKalmanFilter(driver.Lat, driver.Lng)
	lat, lng := filter.Update(msg.Lat, msg.Lng, 1)
	driver.Lat = lat
	driver.Lng = lng
	newCell := mobilityh3.LatLngToCell(lat, lng, 8)
	if s.redis != nil && driver.H3Cell != newCell {
		_ = s.redis.SRem(ctx, "h3:8:"+driver.H3Cell, msg.DriverID)
		_ = s.redis.SAdd(ctx, "h3:8:"+newCell, msg.DriverID)
		_ = s.redis.Set(ctx, "driver:"+msg.DriverID+":cell", newCell, 0)
	}
	driver.H3Cell = newCell
	driver.UpdatedAt = time.Now().UTC()
	s.drivers[msg.DriverID] = driver
	observability.IncCounter("tracking_location_updates_total", 1, map[string]string{"driver_id": msg.DriverID})
	s.mu.Unlock()

	msg.Lat = lat
	msg.Lng = lng
	s.hub.BroadcastDriverLocation(msg.OrderID, msg)
	return nil
}

func (s *Service) closestNode(lat, lng float64) *routing.Node {
	var best *routing.Node
	bestDistance := 1e18
	for _, node := range s.graph.Nodes {
		distance := square(node.Lat-lat) + square(node.Lng-lng)
		if distance < bestDistance {
			best = node
			bestDistance = distance
		}
	}
	return best
}

func square(value float64) float64 {
	return value * value
}

func (s *Service) Ready(ctx context.Context) error {
	if s.cfg == nil {
		return fmt.Errorf("config is required")
	}
	if s.redis != nil {
		if err := s.redis.Ping(ctx); err != nil {
			return fmt.Errorf("redis readiness: %w", err)
		}
	}
	return nil
}

func (s *Service) GetDriverLocation(ctx context.Context, driverID string) (*DriverLocation, error) {
	if driverID == "" {
		return nil, platformerrors.ErrBadRequest
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	driver, ok := s.drivers[driverID]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	copyDriver := driver
	return &copyDriver, nil
}

func (s *Service) prepareTrackingSession(ctx context.Context, orderID string, pickupLat, pickupLng, dropoffLat, dropoffLng *float64) error {
	if orderID == "" {
		return platformerrors.ErrBadRequest
	}
	now := time.Now().UTC()
	session := &TrackingSession{
		SessionID:     "tracking-" + orderID,
		OrderID:       orderID,
		Status:        "pending_assignment",
		StartedAt:     now,
		LastUpdatedAt: now,
	}
	if pickupLat != nil && pickupLng != nil && dropoffLat != nil && dropoffLng != nil {
		if route, err := s.GetRoute(ctx, *pickupLat, *pickupLng, *dropoffLat, *dropoffLng); err == nil {
			session.Route = route
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessionData[orderID] = session
	observability.IncCounter("tracking_sessions_prepared_total", 1, nil)
	return nil
}

func (s *Service) activateTrackingSession(orderID string) error {
	if orderID == "" {
		return platformerrors.ErrBadRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	sessionID := "tracking-" + orderID
	s.sessions[orderID] = sessionID
	session, ok := s.sessionData[orderID]
	if !ok {
		session = &TrackingSession{
			SessionID: sessionID,
			OrderID:   orderID,
			StartedAt: now,
		}
		s.sessionData[orderID] = session
	}
	session.Status = "active"
	session.LastUpdatedAt = now
	if session.StartedAt.IsZero() {
		session.StartedAt = now
	}
	observability.IncCounter("tracking_sessions_started_total", 1, nil)
	observability.SetGauge("tracking_sessions_active", float64(len(s.sessions)), nil)
	return nil
}
