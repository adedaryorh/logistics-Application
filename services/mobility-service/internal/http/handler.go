package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	"github.com/adedaryorh/logistics-platform/services/mobility-service/internal/service"
	"github.com/adedaryorh/logistics-platform/services/mobility-service/internal/tracking"
)

type Handler struct {
	service *service.Service
}

type App struct {
	handler *Handler
}

type routeRequest struct {
	FromLat float64 `json:"from_lat"`
	FromLng float64 `json:"from_lng"`
	ToLat   float64 `json:"to_lat"`
	ToLng   float64 `json:"to_lng"`
}

type trackingSessionRequest struct {
	OrderID string `json:"order_id"`
}

type locationUpdateRequest struct {
	DriverID   string  `json:"driver_id"`
	OrderID    string  `json:"order_id"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	SpeedKMH   float64 `json:"speed_kmh"`
	HeadingDeg float64 `json:"heading_deg"`
}

func RegisterRoutes(router *gin.Engine, cfg *platformconfig.Config) {
	NewApp(cfg).RegisterRoutes(router)
}

func NewApp(cfg *platformconfig.Config) *App {
	return &App{
		handler: &Handler{service: service.New(cfg)},
	}
}

func (a *App) RegisterRoutes(router *gin.Engine) {
	handler := a.handler
	router.GET("/readyz", handler.readyz)

	v1 := router.Group("/api/v1/mobility")
	v1.GET("/drivers/nearby", handler.findNearbyDrivers)
	v1.GET("/h3/cell", handler.getH3Cell)
	v1.GET("/geofences/check", handler.checkGeofence)
	v1.GET("/surge", handler.getSurgeMultiplier)
	v1.POST("/route", handler.getRoute)
	v1.POST("/tracking/sessions", handler.startTrackingSession)
	v1.POST("/tracking/sessions/end", handler.endTrackingSession)
	v1.POST("/tracking/location", handler.updateDriverLocation)
	v1.GET("/orders/:id/track/ws", handler.trackOrder)
}

func (a *App) StartBackground(ctx context.Context, cfg *platformconfig.Config) error {
	if !cfg.Kafka.Enabled {
		<-ctx.Done()
		return ctx.Err()
	}

	consumer := platformkafka.NewConsumer(cfg.Kafka.Brokers).WithSchemaRegistryURL(cfg.Kafka.SchemaRegistryURL)
	group := cfg.Kafka.GroupID + "-mobility"

	return consumer.SubscribeAll(ctx,
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "order.created",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				event, err := platformkafka.DecodeEvent(msg.Value)
				if err != nil {
					return err
				}
				var payload struct {
					OrderID    string   `json:"order_id"`
					PickupLat  *float64 `json:"pickup_lat"`
					PickupLng  *float64 `json:"pickup_lng"`
					DropoffLat *float64 `json:"dropoff_lat"`
					DropoffLng *float64 `json:"dropoff_lng"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					return fmt.Errorf("decode order created payload: %w", err)
				}
				return a.handler.service.HandleOrderCreated(ctx, payload.OrderID, payload.PickupLat, payload.PickupLng, payload.DropoffLat, payload.DropoffLng)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "order.assigned",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				event, err := platformkafka.DecodeEvent(msg.Value)
				if err != nil {
					return err
				}
				var payload struct {
					OrderID string `json:"order_id"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					return fmt.Errorf("decode order assigned payload: %w", err)
				}
				return a.handler.service.HandleOrderAssigned(ctx, payload.OrderID)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "order.cancelled",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				event, err := platformkafka.DecodeEvent(msg.Value)
				if err != nil {
					return err
				}
				var payload struct {
					OrderID string `json:"order_id"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					return fmt.Errorf("decode order cancelled payload: %w", err)
				}
				return a.handler.service.HandleOrderCancelled(ctx, payload.OrderID)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "order.payment_failed",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				event, err := platformkafka.DecodeEvent(msg.Value)
				if err != nil {
					return err
				}
				var payload struct {
					OrderID string `json:"order_id"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					return fmt.Errorf("decode order payment failed payload: %w", err)
				}
				return a.handler.service.HandlePaymentFailed(ctx, payload.OrderID)
			},
		},
	)
}

func (h *Handler) findNearbyDrivers(c *gin.Context) {
	lat, errLat := strconv.ParseFloat(c.Query("lat"), 64)
	lng, errLng := strconv.ParseFloat(c.Query("lng"), 64)
	k, errK := strconv.Atoi(defaultString(c.Query("k_ring"), "1"))
	if errLat != nil || errLng != nil || errK != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	drivers, err := h.service.FindNearbyDrivers(c.Request.Context(), lat, lng, k, c.Query("vehicle_type"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"drivers": drivers})
}

func (h *Handler) getH3Cell(c *gin.Context) {
	lat, errLat := strconv.ParseFloat(c.Query("lat"), 64)
	lng, errLng := strconv.ParseFloat(c.Query("lng"), 64)
	resolution, errResolution := strconv.Atoi(defaultString(c.Query("resolution"), "8"))
	if errLat != nil || errLng != nil || errResolution != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"cell": h.service.GetH3Cell(c.Request.Context(), lat, lng, resolution)})
}

func (h *Handler) checkGeofence(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{
		"in_geofence": h.service.IsInGeofence(c.Request.Context(), c.Query("cell"), c.Query("name")),
	})
}

func (h *Handler) getSurgeMultiplier(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{
		"multiplier": h.service.GetSurgeMultiplier(c.Request.Context(), c.Query("cell")),
	})
}

func (h *Handler) getRoute(c *gin.Context) {
	var req routeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	route, err := h.service.GetRoute(c.Request.Context(), req.FromLat, req.FromLng, req.ToLat, req.ToLng)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"route": route})
}

func (h *Handler) startTrackingSession(c *gin.Context) {
	var req trackingSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	sessionID, err := h.service.StartTrackingSession(c.Request.Context(), req.OrderID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"session_id": sessionID})
}

func (h *Handler) endTrackingSession(c *gin.Context) {
	var req trackingSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if err := h.service.EndTrackingSession(c.Request.Context(), req.OrderID); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"ended": true})
}

func (h *Handler) updateDriverLocation(c *gin.Context) {
	var req locationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if err := h.service.UpdateDriverLocation(c.Request.Context(), tracking.DriverLocationMsg{
		DriverID:   req.DriverID,
		OrderID:    req.OrderID,
		Lat:        req.Lat,
		Lng:        req.Lng,
		SpeedKMH:   req.SpeedKMH,
		HeadingDeg: req.HeadingDeg,
	}); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusAccepted, gin.H{"updated": true})
}

func (h *Handler) trackOrder(c *gin.Context) {
	updates, cancel := h.service.Subscribe(c.Param("id"))
	defer cancel()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case update, ok := <-updates:
			if !ok {
				return false
			}
			c.SSEvent("driver.location", update)
			return true
		case <-time.After(15 * time.Second):
			c.SSEvent("ping", gin.H{"timestamp": time.Now().UTC()})
			return true
		}
	})
}

func (h *Handler) readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
	defer cancel()
	if err := h.service.Ready(ctx); err != nil {
		writeError(c, platformerrors.ErrInternal)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"status": "ready"})
}

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
		"error":   nil,
		"meta": gin.H{
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now().UTC(),
			"version":    "v1",
		},
	})
}

func writeError(c *gin.Context, apiErr *platformerrors.APIError) {
	c.JSON(apiErr.HTTPStatus, gin.H{
		"success": false,
		"data":    nil,
		"error": gin.H{
			"code":    apiErr.Code,
			"message": apiErr.Message,
		},
		"meta": gin.H{
			"request_id": c.GetString("request_id"),
			"timestamp":  time.Now().UTC(),
			"version":    "v1",
		},
	})
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, platformerrors.ErrBadRequest):
		writeError(c, platformerrors.ErrBadRequest)
	case errors.Is(err, platformerrors.ErrNotFound):
		writeError(c, platformerrors.ErrNotFound)
	default:
		writeError(c, platformerrors.ErrInternal)
	}
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
