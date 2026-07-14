package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	platformmiddleware "github.com/adedaryorh/logistics-platform/pkg/middleware"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/evidence"
	"github.com/adedaryorh/logistics-platform/services/logistics-service/internal/model"
	logisticsservice "github.com/adedaryorh/logistics-platform/services/logistics-service/internal/service"
)

type Handler struct {
	service  *logisticsservice.Service
	evidence evidence.Presigner
}

type App struct {
	handler *Handler
	cfg     *platformconfig.Config
}

type createOrderRequest struct {
	Type           string            `json:"type"`
	Pickup         model.Coordinate  `json:"pickup"`
	Dropoff        model.Coordinate  `json:"dropoff"`
	Items          []model.OrderItem `json:"items"`
	IdempotencyKey string            `json:"idempotency_key"`
	MerchantID     *string           `json:"merchant_id"`
}

type cancelOrderRequest struct {
	Reason string `json:"reason"`
}

type driverRequest struct {
	UserID   string  `json:"user_id"`
	FullName string  `json:"full_name"`
	Phone    string  `json:"phone"`
	Type     string  `json:"type"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	H3Cell   string  `json:"h3_cell"`
}

type availabilityRequest struct {
	Online bool `json:"online"`
}

type merchantRequest struct {
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	H3Cell  string  `json:"h3_cell"`
}

type menuItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceMinor  int64  `json:"price_minor"`
	Currency    string `json:"currency"`
	IsAvailable bool   `json:"is_available"`
}

type menuPatchRequest struct {
	Items []menuItemRequest `json:"items"`
}

type assignmentResponseRequest struct {
	UserID string `json:"user_id"`
}

type ratingRequest struct {
	RaterID   string `json:"rater_id"`
	RateeID   string `json:"ratee_id"`
	RateeType string `json:"ratee_type"`
	Score     int    `json:"score"`
	Comment   string `json:"comment"`
}

type agriculturalQuoteRequest struct {
	PlatformUserID       string                     `json:"platform_user_id"`
	FarmSenseRequestID   string                     `json:"farmsense_request_id"`
	MarketplaceRequestID string                     `json:"marketplace_request_id"`
	IdempotencyKey       string                     `json:"idempotency_key"`
	Pickup               model.Coordinate           `json:"pickup"`
	Dropoff              model.Coordinate           `json:"dropoff"`
	Shipment             model.AgriculturalShipment `json:"shipment"`
}
type agriculturalBookingRequest struct {
	QuoteID        string `json:"quote_id"`
	IdempotencyKey string `json:"idempotency_key"`
}
type proofRequest struct {
	EvidenceURL   string           `json:"evidence_url"`
	Notes         string           `json:"notes"`
	RecipientName string           `json:"recipient_name"`
	Coordinate    model.Coordinate `json:"coordinate"`
	CapturedAt    time.Time        `json:"captured_at"`
}

func RegisterRoutes(router *gin.Engine, cfg *platformconfig.Config) {
	NewApp(cfg).RegisterRoutes(router)
}

func NewApp(cfg *platformconfig.Config) *App {
	return &App{
		cfg: cfg,
		handler: &Handler{
			service:  logisticsservice.New(cfg),
			evidence: evidence.HMACPresigner{UploadBaseURL: cfg.Security.EvidenceUploadBaseURL, PublicBaseURL: cfg.Security.EvidencePublicBaseURL, Secret: cfg.Security.EvidenceSigningSecret},
		},
	}
}

func (a *App) RegisterRoutes(router *gin.Engine) {
	handler := a.handler
	router.GET("/readyz", handler.readyz)

	v1 := router.Group("/api/v1")
	v1.POST("/orders", handler.createOrder)
	v1.GET("/orders/:id", handler.getOrder)
	v1.GET("/orders", handler.listOrders)
	v1.POST("/orders/:id/cancel", handler.cancelOrder)
	v1.GET("/orders/:id/track", handler.trackOrder)
	v1.POST("/orders/:id/ratings", handler.createRating)
	v1.POST("/orders/:id/proofs/pickup", handler.recordPickupProof)
	v1.POST("/orders/:id/proofs/delivery", handler.recordDeliveryProof)
	v1.POST("/orders/:id/proofs/uploads", handler.createEvidenceUpload)

	v1.POST("/drivers", handler.createDriver)
	v1.GET("/drivers/me", handler.getDriverMe)
	v1.PATCH("/drivers/me/availability", handler.setDriverAvailability)
	v1.GET("/drivers/nearby", handler.nearbyDrivers)
	v1.POST("/drivers/assignments/:id/accept", handler.acceptAssignment)
	v1.POST("/drivers/assignments/:id/reject", handler.rejectAssignment)

	v1.POST("/merchants", handler.createMerchant)
	v1.GET("/merchants/:id", handler.getMerchant)
	v1.GET("/merchants/:id/menu", handler.getMenu)
	v1.PATCH("/merchants/:id/menu/items", handler.replaceMenuItems)
	v1.POST("/merchants/:id/menu/items", handler.addMenuItem)
	v1.DELETE("/merchants/:id/menu/items/:item_id", handler.deleteMenuItem)

	secrets := map[string]string{"farmsense": a.cfg.Security.FarmSenseServiceSecret, "taskam": a.cfg.Security.TaskAmServiceSecret}
	internal := router.Group("/internal/v1/agricultural", platformmiddleware.PlatformServiceAuth(secrets, 5*time.Minute))
	internal.POST("/quotes", handler.createAgriculturalQuote)
	internal.POST("/bookings", handler.createAgriculturalBooking)
	internal.GET("/bookings/:id", handler.getAgriculturalBooking)
	canonical := router.Group("/api/v1", platformmiddleware.PlatformServiceAuth(secrets, 5*time.Minute))
	canonical.POST("/deliveries/quotes", handler.createCanonicalQuote)
	canonical.POST("/deliveries", handler.createCanonicalBooking)
	canonical.GET("/deliveries/:id", handler.getCanonicalDelivery)
}

func (h *Handler) createEvidenceUpload(c *gin.Context) {
	var req struct {
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
	}
	if c.ShouldBindJSON(&req) != nil || req.FileName == "" || !strings.HasPrefix(req.ContentType, "image/") {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	upload, err := h.evidence.Presign(c.Param("id"), req.FileName, req.ContentType, 15*time.Minute)
	if err != nil {
		writeError(c, platformerrors.ErrInternal)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"upload": upload})
}

func (h *Handler) createAgriculturalQuote(c *gin.Context) {
	var req agriculturalQuoteRequest
	if c.ShouldBindJSON(&req) != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	quote, err := h.service.CreateAgriculturalQuote(c.Request.Context(), logisticsservice.CreateAgriculturalQuoteInput{PlatformService: c.GetString("platform_service"), PlatformUserID: req.PlatformUserID, FarmSenseRequestID: req.FarmSenseRequestID, MarketplaceRequestID: req.MarketplaceRequestID, IdempotencyKey: req.IdempotencyKey, Pickup: req.Pickup, Dropoff: req.Dropoff, Shipment: req.Shipment})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"quote": quote})
}

func (h *Handler) createCanonicalQuote(c *gin.Context) {
	var req agriculturalQuoteRequest
	if c.ShouldBindJSON(&req) != nil || c.GetHeader("Idempotency-Key") == "" {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	req.IdempotencyKey = c.GetHeader("Idempotency-Key")
	quote, err := h.service.CreateAgriculturalQuote(c.Request.Context(), logisticsservice.CreateAgriculturalQuoteInput{PlatformService: c.GetString("platform_service"), PlatformUserID: req.PlatformUserID, FarmSenseRequestID: req.FarmSenseRequestID, MarketplaceRequestID: req.MarketplaceRequestID, IdempotencyKey: req.IdempotencyKey, Pickup: req.Pickup, Dropoff: req.Dropoff, Shipment: req.Shipment})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(nethttp.StatusCreated, gin.H{"quotes": []gin.H{{"quote_id": quote.ID, "platform_user_id": quote.PlatformUserID, "farmsense_request_id": quote.FarmSenseRequestID, "marketplace_request_id": quote.MarketplaceRequestID, "price_minor": quote.PriceMinor, "currency": quote.Currency, "expires_at": quote.ExpiresAt, "shipment": quote.Shipment}}})
}

func (h *Handler) createCanonicalBooking(c *gin.Context) {
	var req agriculturalBookingRequest
	if c.ShouldBindJSON(&req) != nil || c.GetHeader("Idempotency-Key") == "" {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	order, err := h.service.CreateAgriculturalBooking(c.Request.Context(), logisticsservice.CreateAgriculturalBookingInput{PlatformService: c.GetString("platform_service"), QuoteID: req.QuoteID, IdempotencyKey: c.GetHeader("Idempotency-Key")})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(nethttp.StatusCreated, gin.H{"platform_user_id": order.PlatformUserID, "farmsense_request_id": order.FarmSenseRequestID, "marketplace_request_id": order.MarketplaceRequestID, "logistics_delivery_id": order.ID, "status": logisticsservice.NormalizedStatus(order.Status)})
}

func (h *Handler) getCanonicalDelivery(c *gin.Context) {
	order, err := h.service.GetOrder(c.Request.Context(), c.Param("id"))
	if err != nil || order.SourcePlatformService != c.GetString("platform_service") {
		writeError(c, platformerrors.ErrNotFound)
		return
	}
	c.JSON(nethttp.StatusOK, gin.H{"platform_user_id": order.PlatformUserID, "farmsense_request_id": order.FarmSenseRequestID, "marketplace_request_id": order.MarketplaceRequestID, "logistics_delivery_id": order.ID, "status": logisticsservice.NormalizedStatus(order.Status)})
}
func (h *Handler) createAgriculturalBooking(c *gin.Context) {
	var req agriculturalBookingRequest
	if c.ShouldBindJSON(&req) != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	order, err := h.service.CreateAgriculturalBooking(c.Request.Context(), logisticsservice.CreateAgriculturalBookingInput{PlatformService: c.GetString("platform_service"), QuoteID: req.QuoteID, IdempotencyKey: req.IdempotencyKey})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"booking": order})
}
func (h *Handler) getAgriculturalBooking(c *gin.Context) {
	order, err := h.service.GetOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	if order.AgriculturalShipment == nil {
		writeError(c, platformerrors.ErrNotFound)
		return
	}
	if order.SourcePlatformService != c.GetString("platform_service") {
		writeError(c, platformerrors.ErrNotFound)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"booking": order})
}
func (h *Handler) recordPickupProof(c *gin.Context)   { h.recordProof(c, "pickup") }
func (h *Handler) recordDeliveryProof(c *gin.Context) { h.recordProof(c, "delivery") }
func (h *Handler) recordProof(c *gin.Context, kind string) {
	var req proofRequest
	if c.ShouldBindJSON(&req) != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	order, err := h.service.RecordDeliveryProof(c.Request.Context(), logisticsservice.RecordProofInput{OrderID: c.Param("id"), DriverUserID: c.GetHeader("X-User-ID"), Type: kind, EvidenceURL: req.EvidenceURL, Notes: req.Notes, RecipientName: req.RecipientName, Coordinate: req.Coordinate, CapturedAt: req.CapturedAt})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"order": order, "proof": order.Proofs[len(order.Proofs)-1]})
}

func (a *App) StartBackground(ctx context.Context, cfg *platformconfig.Config) error {
	go a.handler.service.StartWebhookWorker(ctx)
	if !cfg.Kafka.Enabled {
		<-ctx.Done()
		return ctx.Err()
	}

	producer := platformkafka.NewProducer(cfg.Kafka.Brokers).WithSchemaRegistryURL(cfg.Kafka.SchemaRegistryURL)
	defer producer.Close()
	consumer := platformkafka.NewConsumer(cfg.Kafka.Brokers).WithSchemaRegistryURL(cfg.Kafka.SchemaRegistryURL)

	errCh := make(chan error, 2)
	go func() {
		errCh <- a.handler.service.StartOutboxWorker(
			ctx,
			producer,
			time.Duration(cfg.Kafka.OutboxPollIntervalMillis)*time.Millisecond,
			cfg.Kafka.TopicPrefix,
		)
	}()
	go func() {
		errCh <- consumer.SubscribeAll(ctx,
			platformkafka.Subscription{
				Topic:   cfg.Kafka.TopicPrefix + "payment.completed",
				GroupID: cfg.Kafka.GroupID + "-logistics",
				Handler: func(ctx context.Context, msg kafka.Message) error {
					event, err := platformkafka.DecodeEvent(msg.Value)
					if err != nil {
						return err
					}
					var payload struct {
						OrderID string `json:"order_id"`
					}
					if err := json.Unmarshal(event.Payload, &payload); err != nil {
						return fmt.Errorf("decode payment completed payload: %w", err)
					}
					return a.handler.service.HandlePaymentCompleted(ctx, payload.OrderID)
				},
			},
			platformkafka.Subscription{
				Topic:   cfg.Kafka.TopicPrefix + "payment.failed",
				GroupID: cfg.Kafka.GroupID + "-logistics",
				Handler: func(ctx context.Context, msg kafka.Message) error {
					event, err := platformkafka.DecodeEvent(msg.Value)
					if err != nil {
						return err
					}
					var payload struct {
						OrderID string `json:"order_id"`
						Reason  string `json:"reason"`
					}
					if err := json.Unmarshal(event.Payload, &payload); err != nil {
						return fmt.Errorf("decode payment failed payload: %w", err)
					}
					return a.handler.service.HandlePaymentFailed(ctx, payload.OrderID, payload.Reason)
				},
			},
			platformkafka.Subscription{
				Topic:   cfg.Kafka.TopicPrefix + "payment.refunded",
				GroupID: cfg.Kafka.GroupID + "-logistics",
				Handler: func(ctx context.Context, msg kafka.Message) error {
					event, err := platformkafka.DecodeEvent(msg.Value)
					if err != nil {
						return err
					}
					var payload struct {
						OrderID string `json:"order_id"`
					}
					if err := json.Unmarshal(event.Payload, &payload); err != nil {
						return fmt.Errorf("decode payment refunded payload: %w", err)
					}
					return a.handler.service.HandlePaymentRefunded(ctx, payload.OrderID)
				},
			},
		)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (h *Handler) createOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}

	customerID := c.GetHeader("X-User-ID")
	order, err := h.service.CreateOrder(c.Request.Context(), logisticsservice.CreateOrderInput{
		CustomerID:     customerID,
		Type:           req.Type,
		Pickup:         req.Pickup,
		Dropoff:        req.Dropoff,
		Items:          req.Items,
		IdempotencyKey: req.IdempotencyKey,
		MerchantID:     req.MerchantID,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"order": order})
}

func (h *Handler) getOrder(c *gin.Context) {
	order, err := h.service.GetOrder(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"order": order})
}

func (h *Handler) listOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	orders, next, err := h.service.ListOrders(c.Request.Context(), logisticsservice.ListOrdersInput{
		Cursor: c.Query("cursor"),
		Limit:  limit,
		Status: c.Query("status"),
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{
		"orders": orders,
		"next":   next,
	})
}

func (h *Handler) cancelOrder(c *gin.Context) {
	var req cancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	order, err := h.service.CancelOrder(c.Request.Context(), logisticsservice.CancelOrderInput{
		OrderID: c.Param("id"),
		ActorID: c.GetHeader("X-User-ID"),
		Reason:  req.Reason,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"order": order})
}

func (h *Handler) trackOrder(c *gin.Context) {
	updates, cancel, err := h.service.SubscribeOrder(c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	defer cancel()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case order, ok := <-updates:
			if !ok {
				return false
			}
			c.SSEvent("order.status", gin.H{
				"id":         order.ID,
				"status":     order.Status,
				"driver_id":  order.DriverID,
				"updated_at": order.UpdatedAt,
			})
			return true
		case <-time.After(15 * time.Second):
			c.SSEvent("ping", gin.H{"timestamp": time.Now().UTC()})
			return true
		}
	})
}

func (h *Handler) createDriver(c *gin.Context) {
	var req driverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	driver, err := h.service.CreateDriver(c.Request.Context(), logisticsservice.DriverOnboardingInput{
		UserID:   req.UserID,
		FullName: req.FullName,
		Phone:    req.Phone,
		Type:     req.Type,
		Lat:      req.Lat,
		Lng:      req.Lng,
		H3Cell:   req.H3Cell,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"driver": driver})
}

func (h *Handler) getDriverMe(c *gin.Context) {
	driver, err := h.service.GetDriverByUserID(c.Request.Context(), c.GetHeader("X-User-ID"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"driver": driver})
}

func (h *Handler) setDriverAvailability(c *gin.Context) {
	var req availabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	driver, err := h.service.SetDriverAvailability(c.Request.Context(), logisticsservice.DriverAvailabilityInput{
		UserID: c.GetHeader("X-User-ID"),
		Online: req.Online,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"driver": driver})
}

func (h *Handler) nearbyDrivers(c *gin.Context) {
	lat, errLat := strconv.ParseFloat(c.Query("lat"), 64)
	lng, errLng := strconv.ParseFloat(c.Query("lng"), 64)
	if errLat != nil || errLng != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	drivers, err := h.service.NearbyDrivers(c.Request.Context(), logisticsservice.NearbyDriversInput{
		Lat:  lat,
		Lng:  lng,
		Type: c.Query("type"),
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"drivers": drivers})
}

func (h *Handler) acceptAssignment(c *gin.Context) {
	h.respondAssignment(c, true)
}

func (h *Handler) rejectAssignment(c *gin.Context) {
	h.respondAssignment(c, false)
}

func (h *Handler) respondAssignment(c *gin.Context, accepted bool) {
	var req assignmentResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	assignment, err := h.service.RespondToAssignment(c.Request.Context(), logisticsservice.AssignmentResponseInput{
		AssignmentID: c.Param("id"),
		DriverUserID: req.UserID,
		Accepted:     accepted,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"assignment": assignment})
}

func (h *Handler) createMerchant(c *gin.Context) {
	var req merchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	merchant, err := h.service.CreateMerchant(c.Request.Context(), logisticsservice.MerchantInput{
		Name:    req.Name,
		Type:    req.Type,
		Address: req.Address,
		Lat:     req.Lat,
		Lng:     req.Lng,
		H3Cell:  req.H3Cell,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"merchant": merchant})
}

func (h *Handler) getMerchant(c *gin.Context) {
	merchant, err := h.service.GetMerchant(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"merchant": merchant})
}

func (h *Handler) getMenu(c *gin.Context) {
	items, err := h.service.GetMenu(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"items": items})
}

func (h *Handler) replaceMenuItems(c *gin.Context) {
	var req menuPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	items, err := h.service.ReplaceMenuItems(c.Request.Context(), c.Param("id"), mapMenuInputs(req.Items))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"items": items})
}

func (h *Handler) addMenuItem(c *gin.Context) {
	var req menuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	item, err := h.service.AddMenuItem(c.Request.Context(), c.Param("id"), logisticsservice.MenuItemInput{
		Name:        req.Name,
		Description: req.Description,
		PriceMinor:  req.PriceMinor,
		Currency:    req.Currency,
		IsAvailable: req.IsAvailable,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"item": item})
}

func (h *Handler) deleteMenuItem(c *gin.Context) {
	if err := h.service.DeleteMenuItem(c.Request.Context(), c.Param("id"), c.Param("item_id")); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) createRating(c *gin.Context) {
	var req ratingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	rating, err := h.service.CreateRating(c.Request.Context(), logisticsservice.RatingInput{
		OrderID:   c.Param("id"),
		RaterID:   req.RaterID,
		RateeID:   req.RateeID,
		RateeType: req.RateeType,
		Score:     req.Score,
		Comment:   req.Comment,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"rating": rating})
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

func mapMenuInputs(items []menuItemRequest) []logisticsservice.MenuItemInput {
	result := make([]logisticsservice.MenuItemInput, 0, len(items))
	for _, item := range items {
		result = append(result, logisticsservice.MenuItemInput{
			Name:        item.Name,
			Description: item.Description,
			PriceMinor:  item.PriceMinor,
			Currency:    item.Currency,
			IsAvailable: item.IsAvailable,
		})
	}
	return result
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
	case errors.Is(err, platformerrors.ErrUnauthorized):
		writeError(c, platformerrors.ErrUnauthorized)
	case errors.Is(err, platformerrors.ErrConflict):
		writeError(c, platformerrors.ErrConflict)
	case errors.Is(err, platformerrors.ErrNotFound):
		writeError(c, platformerrors.ErrNotFound)
	default:
		writeError(c, platformerrors.ErrInternal)
	}
}

func databaseDSN(cfg *platformconfig.Config) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)
}

func bearerUserID(value string) string {
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return value
}
