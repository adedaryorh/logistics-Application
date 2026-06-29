package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/adedaryorh/logistics-platform/pkg/observability"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/application/dispatch"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
	logisticstemporal "github.com/adedaryorh/logistics-platform/services/logistics-service/internal/temporal"
)

type Service struct {
	cfg         *platformconfig.Config
	coordinator *logisticstemporal.Coordinator

	mu               sync.RWMutex
	orders           map[string]*model.Order
	orderIDs         []string
	orderHistory     map[string][]model.OrderStatusHistory
	orderAssignments map[string][]model.Assignment
	drivers          map[string]*model.Driver
	driversByUserID  map[string]string
	merchants        map[string]*model.Merchant
	menuItems        map[string][]model.MenuItem
	ratings          map[string][]model.Rating
	outboxEvents     []model.OutboxEvent
	watchers         map[string]map[chan model.Order]struct{}
}

type CreateOrderInput struct {
	CustomerID     string
	Type           string
	Pickup         model.Coordinate
	Dropoff        model.Coordinate
	Items          []model.OrderItem
	IdempotencyKey string
	MerchantID     *string
}

type CancelOrderInput struct {
	OrderID string
	ActorID string
	Reason  string
}

type ListOrdersInput struct {
	Cursor string
	Limit  int
	Status string
}

type DriverOnboardingInput struct {
	UserID   string
	FullName string
	Phone    string
	Type     string
	Lat      float64
	Lng      float64
	H3Cell   string
}

type DriverAvailabilityInput struct {
	UserID string
	Online bool
}

type NearbyDriversInput struct {
	Lat  float64
	Lng  float64
	Type string
}

type MerchantInput struct {
	Name    string
	Type    string
	Address string
	Lat     float64
	Lng     float64
	H3Cell  string
}

type MenuItemInput struct {
	Name        string
	Description string
	PriceMinor  int64
	Currency    string
	IsAvailable bool
}

type AssignmentResponseInput struct {
	AssignmentID string
	DriverUserID string
	Accepted     bool
}

type RatingInput struct {
	OrderID   string
	RaterID   string
	RateeID   string
	RateeType string
	Score     int
	Comment   string
}

func New(cfg *platformconfig.Config) *Service {
	return &Service{
		cfg:              cfg,
		coordinator:      logisticstemporal.NewCoordinator(),
		orders:           map[string]*model.Order{},
		orderIDs:         []string{},
		orderHistory:     map[string][]model.OrderStatusHistory{},
		orderAssignments: map[string][]model.Assignment{},
		drivers:          map[string]*model.Driver{},
		driversByUserID:  map[string]string{},
		merchants:        map[string]*model.Merchant{},
		menuItems:        map[string][]model.MenuItem{},
		ratings:          map[string][]model.Rating{},
		outboxEvents:     []model.OutboxEvent{},
		watchers:         map[string]map[chan model.Order]struct{}{},
	}
}

func (s *Service) CreateOrder(ctx context.Context, input CreateOrderInput) (*model.Order, error) {
	if input.CustomerID == "" || input.IdempotencyKey == "" {
		return nil, platformerrors.ErrBadRequest
	}
	if !validOrderType(input.Type) {
		return nil, platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range s.orderIDs {
		if s.orders[id].IdempotencyKey == input.IdempotencyKey {
			orderCopy := *s.orders[id]
			return &orderCopy, nil
		}
	}

	orderID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate order id: %w", err)
	}
	workflowID, err := s.coordinator.StartOrderSaga(ctx, logisticstemporal.OrderSagaInput{
		OrderID:    orderID.String(),
		CustomerID: input.CustomerID,
		Type:       input.Type,
	})
	if err != nil {
		return nil, fmt.Errorf("start order saga: %w", err)
	}

	now := time.Now().UTC()
	order := &model.Order{
		ID:                 orderID.String(),
		CustomerID:         input.CustomerID,
		MerchantID:         input.MerchantID,
		Type:               model.OrderType(input.Type),
		Status:             model.OrderStatusDispatching,
		Pickup:             withDerivedH3(input.Pickup),
		Dropoff:            withDerivedH3(input.Dropoff),
		Items:              normalizeItems(input.Items),
		PriceMinor:         estimateFare(input.Pickup, input.Dropoff, input.Items),
		Currency:           "NGN",
		SurgeMultiplier:    1,
		IdempotencyKey:     input.IdempotencyKey,
		TemporalWorkflowID: workflowID,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	s.orders[order.ID] = order
	s.orderIDs = append(s.orderIDs, order.ID)
	s.appendOrderHistoryLocked(order.ID, nil, order.Status, nil, &input.CustomerID, now)
	s.appendOutboxLocked(order.ID, "order.created", map[string]any{
		"order_id":        order.ID,
		"customer_id":     order.CustomerID,
		"type":            order.Type,
		"status":          order.Status,
		"price_minor":     order.PriceMinor,
		"workflow_id":     order.TemporalWorkflowID,
		"idempotency_key": order.IdempotencyKey,
		"pickup_lat":      order.Pickup.Lat,
		"pickup_lng":      order.Pickup.Lng,
		"dropoff_lat":     order.Dropoff.Lat,
		"dropoff_lng":     order.Dropoff.Lng,
	})

	s.seedAssignmentsLocked(order)
	s.broadcastLocked(order)
	observability.IncCounter("orders_created_total", 1, map[string]string{"type": input.Type, "status": string(order.Status)})

	orderCopy := *order
	return &orderCopy, nil
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[orderID]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	orderCopy := *order
	return &orderCopy, nil
}

func (s *Service) ListOrders(ctx context.Context, input ListOrdersInput) ([]model.Order, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := input.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	start := 0
	if input.Cursor != "" {
		idx, err := strconv.Atoi(input.Cursor)
		if err == nil && idx >= 0 {
			start = idx
		}
	}

	filtered := make([]model.Order, 0, len(s.orderIDs))
	for _, id := range s.orderIDs {
		order := s.orders[id]
		if input.Status != "" && string(order.Status) != input.Status {
			continue
		}
		filtered = append(filtered, *order)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	if start >= len(filtered) {
		return []model.Order{}, "", nil
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	next := ""
	if end < len(filtered) {
		next = strconv.Itoa(end)
	}

	return filtered[start:end], next, nil
}

func (s *Service) CancelOrder(ctx context.Context, input CancelOrderInput) (*model.Order, error) {
	if input.OrderID == "" || input.ActorID == "" || input.Reason == "" {
		return nil, platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[input.OrderID]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	if order.Status == model.OrderStatusDelivered || order.Status == model.OrderStatusCancelled {
		return nil, platformerrors.ErrConflict
	}

	prev := order.Status
	order.Status = model.OrderStatusCancelled
	order.CancellationReason = &input.Reason
	order.UpdatedAt = time.Now().UTC()
	s.appendOrderHistoryLocked(order.ID, &prev, order.Status, &input.Reason, &input.ActorID, order.UpdatedAt)
	s.appendOutboxLocked(order.ID, "order.cancelled", map[string]any{
		"order_id": order.ID,
		"reason":   input.Reason,
	})
	s.broadcastLocked(order)
	observability.IncCounter("orders_cancelled_total", 1, map[string]string{"reason": input.Reason})

	orderCopy := *order
	return &orderCopy, nil
}

func (s *Service) SubscribeOrder(orderID string) (<-chan model.Order, func(), error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return nil, nil, platformerrors.ErrNotFound
	}

	ch := make(chan model.Order, 8)
	if s.watchers[orderID] == nil {
		s.watchers[orderID] = map[chan model.Order]struct{}{}
	}
	s.watchers[orderID][ch] = struct{}{}
	ch <- *order

	cancel := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.watchers[orderID], ch)
		close(ch)
	}

	return ch, cancel, nil
}

func (s *Service) CreateDriver(ctx context.Context, input DriverOnboardingInput) (*model.Driver, error) {
	if input.UserID == "" || input.FullName == "" || input.Phone == "" || input.Type == "" {
		return nil, platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existingID, ok := s.driversByUserID[input.UserID]; ok {
		driverCopy := *s.drivers[existingID]
		return &driverCopy, nil
	}

	driverID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate driver id: %w", err)
	}
	now := time.Now().UTC()
	driver := &model.Driver{
		ID:         driverID.String(),
		UserID:     input.UserID,
		FullName:   input.FullName,
		Phone:      input.Phone,
		Type:       input.Type,
		Lat:        input.Lat,
		Lng:        input.Lng,
		H3Cell:     deriveH3Cell(input.Lat, input.Lng),
		Rating:     5,
		TotalTrips: 0,
		Online:     false,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if input.H3Cell != "" {
		driver.H3Cell = input.H3Cell
	}

	s.drivers[driver.ID] = driver
	s.driversByUserID[input.UserID] = driver.ID
	observability.IncCounter("drivers_created_total", 1, map[string]string{"type": driver.Type})

	driverCopy := *driver
	return &driverCopy, nil
}

func (s *Service) GetDriverByUserID(ctx context.Context, userID string) (*model.Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	driverID, ok := s.driversByUserID[userID]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	driverCopy := *s.drivers[driverID]
	return &driverCopy, nil
}

func (s *Service) SetDriverAvailability(ctx context.Context, input DriverAvailabilityInput) (*model.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	driverID, ok := s.driversByUserID[input.UserID]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	driver := s.drivers[driverID]
	driver.Online = input.Online
	driver.UpdatedAt = time.Now().UTC()

	eventType := "driver.offline"
	if input.Online {
		eventType = "driver.online"
	}
	s.appendOutboxLocked(driver.ID, eventType, map[string]any{
		"driver_id": driver.ID,
		"user_id":   driver.UserID,
		"online":    driver.Online,
	})
	observability.IncCounter("driver_availability_changes_total", 1, map[string]string{"online": fmt.Sprintf("%t", driver.Online), "type": driver.Type})
	observability.SetGauge("drivers_online_total", float64(s.countDriversOnlineLocked()), map[string]string{"type": driver.Type})

	driverCopy := *driver
	return &driverCopy, nil
}

func (s *Service) NearbyDrivers(ctx context.Context, input NearbyDriversInput) ([]model.Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type scored struct {
		driver model.Driver
		score  float64
	}
	scoredDrivers := make([]scored, 0)
	for _, driver := range s.drivers {
		if !driver.IsActive || !driver.Online {
			continue
		}
		if input.Type != "" && driver.Type != input.Type {
			continue
		}
		distance := haversineMeters(input.Lat, input.Lng, driver.Lat, driver.Lng)
		score := dispatch.ScoreDriver(dispatch.DriverCandidate{
			DriverID:   driver.ID,
			DistanceM:  distance,
			Rating:     driver.Rating,
			TotalTrips: driver.TotalTrips,
			H3Cell:     driver.H3Cell,
		}, 15000)
		scoredDrivers = append(scoredDrivers, scored{driver: *driver, score: score})
	}

	sort.SliceStable(scoredDrivers, func(i, j int) bool {
		return scoredDrivers[i].score > scoredDrivers[j].score
	})

	result := make([]model.Driver, 0, len(scoredDrivers))
	for _, item := range scoredDrivers {
		result = append(result, item.driver)
	}
	observability.ObserveHistogram("nearby_driver_results", float64(len(result)), map[string]string{"type": input.Type})
	return result, nil
}

func (s *Service) RespondToAssignment(ctx context.Context, input AssignmentResponseInput) (*model.Assignment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for orderID, assignments := range s.orderAssignments {
		for idx := range assignments {
			if assignments[idx].ID != input.AssignmentID {
				continue
			}
			driverID, ok := s.driversByUserID[input.DriverUserID]
			if !ok || assignments[idx].DriverID != driverID {
				return nil, platformerrors.ErrUnauthorized
			}
			if assignments[idx].Status != model.AssignmentStatusOffered {
				return nil, platformerrors.ErrConflict
			}

			now := time.Now().UTC()
			assignments[idx].RespondedAt = &now
			eventType := "dispatch.rejected"
			if input.Accepted {
				assignments[idx].Status = model.AssignmentStatusAccepted
				eventType = "dispatch.accepted"
				order := s.orders[orderID]
				prev := order.Status
				order.Status = model.OrderStatusAssigned
				order.DriverID = &driverID
				order.UpdatedAt = now
				s.appendOrderHistoryLocked(order.ID, &prev, order.Status, nil, &input.DriverUserID, now)
				s.broadcastLocked(order)
				s.appendOutboxLocked(order.ID, "order.assigned", map[string]any{
					"order_id":  order.ID,
					"driver_id": driverID,
				})
			} else {
				assignments[idx].Status = model.AssignmentStatusRejected
			}

			s.orderAssignments[orderID] = assignments
			s.appendOutboxLocked(orderID, eventType, map[string]any{
				"order_id":       orderID,
				"assignment_id":  assignments[idx].ID,
				"driver_id":      assignments[idx].DriverID,
				"attempt_number": assignments[idx].AttemptNumber,
			})
			assignmentStatus := "rejected"
			if input.Accepted {
				assignmentStatus = "accepted"
				observability.IncCounter("orders_assigned_total", 1, map[string]string{"type": string(s.orders[orderID].Type)})
			}
			observability.IncCounter("assignment_responses_total", 1, map[string]string{"status": assignmentStatus})

			assignmentCopy := assignments[idx]
			return &assignmentCopy, nil
		}
	}

	return nil, platformerrors.ErrNotFound
}

func (s *Service) CreateMerchant(ctx context.Context, input MerchantInput) (*model.Merchant, error) {
	if input.Name == "" || input.Type == "" {
		return nil, platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	merchantID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate merchant id: %w", err)
	}
	now := time.Now().UTC()
	merchant := &model.Merchant{
		ID:        merchantID.String(),
		Name:      input.Name,
		Type:      input.Type,
		Address:   input.Address,
		Lat:       input.Lat,
		Lng:       input.Lng,
		H3Cell:    deriveH3Cell(input.Lat, input.Lng),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if input.H3Cell != "" {
		merchant.H3Cell = input.H3Cell
	}
	s.merchants[merchant.ID] = merchant

	merchantCopy := *merchant
	return &merchantCopy, nil
}

func (s *Service) GetMerchant(ctx context.Context, merchantID string) (*model.Merchant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	merchant, ok := s.merchants[merchantID]
	if !ok {
		return nil, platformerrors.ErrNotFound
	}
	merchantCopy := *merchant
	merchantCopy.MenuItems = append([]model.MenuItem(nil), s.menuItems[merchantID]...)
	return &merchantCopy, nil
}

func (s *Service) GetMenu(ctx context.Context, merchantID string) ([]model.MenuItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.merchants[merchantID]; !ok {
		return nil, platformerrors.ErrNotFound
	}
	return append([]model.MenuItem(nil), s.menuItems[merchantID]...), nil
}

func (s *Service) ReplaceMenuItems(ctx context.Context, merchantID string, items []MenuItemInput) ([]model.MenuItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.merchants[merchantID]; !ok {
		return nil, platformerrors.ErrNotFound
	}
	newItems := make([]model.MenuItem, 0, len(items))
	for _, item := range items {
		menuItem, err := buildMenuItem(merchantID, item)
		if err != nil {
			return nil, err
		}
		newItems = append(newItems, menuItem)
	}
	s.menuItems[merchantID] = newItems
	return append([]model.MenuItem(nil), newItems...), nil
}

func (s *Service) AddMenuItem(ctx context.Context, merchantID string, input MenuItemInput) (*model.MenuItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.merchants[merchantID]; !ok {
		return nil, platformerrors.ErrNotFound
	}
	menuItem, err := buildMenuItem(merchantID, input)
	if err != nil {
		return nil, err
	}
	s.menuItems[merchantID] = append(s.menuItems[merchantID], menuItem)
	return &menuItem, nil
}

func (s *Service) DeleteMenuItem(ctx context.Context, merchantID, itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.menuItems[merchantID]
	filtered := items[:0]
	found := false
	for _, item := range items {
		if item.ID == itemID {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}
	if !found {
		return platformerrors.ErrNotFound
	}
	s.menuItems[merchantID] = append([]model.MenuItem(nil), filtered...)
	return nil
}

func (s *Service) CreateRating(ctx context.Context, input RatingInput) (*model.Rating, error) {
	if input.OrderID == "" || input.RaterID == "" || input.RateeID == "" || input.Score < 1 || input.Score > 5 {
		return nil, platformerrors.ErrBadRequest
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.orders[input.OrderID]; !ok {
		return nil, platformerrors.ErrNotFound
	}
	ratingID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate rating id: %w", err)
	}
	rating := model.Rating{
		ID:        ratingID.String(),
		OrderID:   input.OrderID,
		RaterID:   input.RaterID,
		RateeID:   input.RateeID,
		RateeType: input.RateeType,
		Score:     input.Score,
		Comment:   input.Comment,
		CreatedAt: time.Now().UTC(),
	}
	s.ratings[input.OrderID] = append(s.ratings[input.OrderID], rating)
	return &rating, nil
}

func validOrderType(value string) bool {
	switch value {
	case string(model.OrderTypeRide), string(model.OrderTypeFood), string(model.OrderTypeParcel):
		return true
	default:
		return false
	}
}

func withDerivedH3(coord model.Coordinate) model.Coordinate {
	if coord.H3Cell == "" {
		coord.H3Cell = deriveH3Cell(coord.Lat, coord.Lng)
	}
	return coord
}

func deriveH3Cell(lat, lng float64) string {
	return fmt.Sprintf("r8:%0.3f:%0.3f", lat, lng)
}

func normalizeItems(items []model.OrderItem) []model.OrderItem {
	normalized := make([]model.OrderItem, 0, len(items))
	for _, item := range items {
		copyItem := item
		if copyItem.ID == "" {
			copyItem.ID = uuid.Must(uuid.NewV7()).String()
		}
		if copyItem.Quantity <= 0 {
			copyItem.Quantity = 1
		}
		if copyItem.Currency == "" {
			copyItem.Currency = "NGN"
		}
		normalized = append(normalized, copyItem)
	}
	return normalized
}

func estimateFare(pickup, dropoff model.Coordinate, items []model.OrderItem) int64 {
	distanceM := haversineMeters(pickup.Lat, pickup.Lng, dropoff.Lat, dropoff.Lng)
	baseFare := int64(1500)
	distanceFare := int64(distanceM / 40)
	itemFare := int64(0)
	for _, item := range items {
		itemFare += int64(item.Quantity) * item.PriceMinor
	}
	if itemFare == 0 {
		itemFare = 500
	}
	return baseFare + distanceFare + itemFare
}

func buildMenuItem(merchantID string, input MenuItemInput) (model.MenuItem, error) {
	if input.Name == "" || input.PriceMinor <= 0 {
		return model.MenuItem{}, platformerrors.ErrBadRequest
	}
	itemID, err := uuid.NewV7()
	if err != nil {
		return model.MenuItem{}, fmt.Errorf("generate menu item id: %w", err)
	}
	currency := input.Currency
	if currency == "" {
		currency = "NGN"
	}
	return model.MenuItem{
		ID:          itemID.String(),
		MerchantID:  merchantID,
		Name:        input.Name,
		Description: input.Description,
		PriceMinor:  input.PriceMinor,
		Currency:    currency,
		IsAvailable: input.IsAvailable,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (s *Service) appendOrderHistoryLocked(orderID string, from *model.OrderStatus, to model.OrderStatus, reason *string, actorID *string, createdAt time.Time) {
	historyID := uuid.Must(uuid.NewV7())
	entry := model.OrderStatusHistory{
		ID:         historyID.String(),
		OrderID:    orderID,
		FromStatus: from,
		ToStatus:   to,
		Reason:     reason,
		ActorID:    actorID,
		CreatedAt:  createdAt,
	}
	s.orderHistory[orderID] = append(s.orderHistory[orderID], entry)
}

func (s *Service) appendOutboxLocked(aggregateID, eventType string, payload map[string]any) {
	eventID := uuid.Must(uuid.NewV7())
	s.outboxEvents = append(s.outboxEvents, model.OutboxEvent{
		ID:          eventID.String(),
		AggregateID: aggregateID,
		EventType:   eventType,
		Payload:     payload,
		CreatedAt:   time.Now().UTC(),
	})
}

func (s *Service) seedAssignmentsLocked(order *model.Order) {
	candidates := make([]dispatch.DriverCandidate, 0)
	driverByID := map[string]*model.Driver{}
	for _, driver := range s.drivers {
		if !driver.Online || !driver.IsActive {
			continue
		}
		distance := haversineMeters(order.Pickup.Lat, order.Pickup.Lng, driver.Lat, driver.Lng)
		candidate := dispatch.DriverCandidate{
			DriverID:   driver.ID,
			DistanceM:  distance,
			Rating:     driver.Rating,
			TotalTrips: driver.TotalTrips,
			H3Cell:     driver.H3Cell,
		}
		candidates = append(candidates, candidate)
		driverByID[driver.ID] = driver
	}
	ranked := dispatch.RankDrivers(candidates, 15000)
	limit := len(ranked)
	if limit > 5 {
		limit = 5
	}
	assignments := make([]model.Assignment, 0, limit)
	for i := 0; i < limit; i++ {
		assignmentID := uuid.Must(uuid.NewV7())
		now := time.Now().UTC()
		assignments = append(assignments, model.Assignment{
			ID:            assignmentID.String(),
			OrderID:       order.ID,
			DriverID:      ranked[i].DriverID,
			AttemptNumber: i + 1,
			Status:        model.AssignmentStatusOffered,
			OfferedAt:     now,
			TimeoutAt:     now.Add(30 * time.Second),
		})
		s.appendOutboxLocked(order.ID, "dispatch.offered", map[string]any{
			"order_id":       order.ID,
			"driver_id":      ranked[i].DriverID,
			"attempt_number": i + 1,
		})
	}
	s.orderAssignments[order.ID] = assignments
}

func (s *Service) broadcastLocked(order *model.Order) {
	for ch := range s.watchers[order.ID] {
		select {
		case ch <- *order:
		default:
		}
	}
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6371000
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

func (s *Service) countDriversOnlineLocked() int {
	total := 0
	for _, driver := range s.drivers {
		if driver.Online {
			total++
		}
	}
	return total
}
