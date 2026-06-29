package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	platformmiddleware "github.com/adedaryorh/logistics-platform/pkg/middleware"
	platformvalidation "github.com/adedaryorh/logistics-platform/pkg/validation"
	paymentservice "github.com/adedaryorh/logistics-platform/services/payment-service/internal/service"
)

type Handler struct {
	service *paymentservice.Service
}

type App struct {
	cfg     *platformconfig.Config
	handler *Handler
}

type initializeRequest struct {
	OrderID        string `json:"order_id" validate:"required"`
	CustomerID     string `json:"customer_id" validate:"required"`
	AmountMinor    int64  `json:"amount_minor" validate:"required"`
	Provider       string `json:"provider" validate:"required,oneof=flutterwave paystack stripe"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
	CustomerEmail  string `json:"customer_email" validate:"required,email"`
	CustomerName   string `json:"customer_name" validate:"required,min=2"`
}

type refundRequest struct {
	AmountMinor int64  `json:"amount_minor"`
	Reason      string `json:"reason" validate:"required"`
}

type payoutRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	Provider       string `json:"provider" validate:"required,oneof=flutterwave paystack stripe"`
	AmountMinor    int64  `json:"amount_minor" validate:"required"`
	Currency       string `json:"currency"`
	DestinationRef string `json:"destination_ref" validate:"required"`
	IdempotencyKey string `json:"idempotency_key" validate:"required"`
}

type internalReleaseRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

func RegisterRoutes(router *gin.Engine, cfg *platformconfig.Config) {
	NewApp(cfg).RegisterRoutes(router)
}

func NewApp(cfg *platformconfig.Config) *App {
	return &App{
		cfg:     cfg,
		handler: &Handler{service: paymentservice.New(cfg)},
	}
}

func (a *App) RegisterRoutes(router *gin.Engine) {
	handler := a.handler
	router.GET("/readyz", handler.readyz)

	v1 := router.Group("/api/v1")
	v1.POST("/payments/initialize", handler.initializePayment)
	v1.GET("/payments/:id", handler.getPayment)
	v1.POST("/payments/:id/refund", handler.refundPayment)
	v1.POST("/webhooks/flutterwave", handler.webhook("flutterwave"))
	v1.POST("/webhooks/paystack", handler.webhook("paystack"))
	v1.POST("/webhooks/stripe", handler.webhook("stripe"))

	internal := router.Group("/internal")
	internal.Use(platformmiddleware.InternalOnly(a.cfg, "ops", "finance", "ops.admin"))
	internal.POST("/payments/initialize", handler.initializePayment)
	internal.POST("/payments/:id/release", handler.releaseFunds)
	internal.POST("/payments/:id/cancel", handler.cancelPayment)
	internal.POST("/payouts", handler.createPayout)
	internal.POST("/payments/reconcile", handler.reconcilePayments)
	internal.GET("/payments/audit", handler.auditRecords)
}

func (a *App) StartBackground(ctx context.Context, cfg *platformconfig.Config) error {
	if !cfg.Kafka.Enabled {
		<-ctx.Done()
		return ctx.Err()
	}

	producer := platformkafka.NewProducer(cfg.Kafka.Brokers).WithSchemaRegistryURL(cfg.Kafka.SchemaRegistryURL)
	defer producer.Close()
	consumer := platformkafka.NewConsumer(cfg.Kafka.Brokers).WithSchemaRegistryURL(cfg.Kafka.SchemaRegistryURL)

	errCh := make(chan error, 3)
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
				Topic:   cfg.Kafka.TopicPrefix + "order.cancelled",
				GroupID: cfg.Kafka.GroupID + "-payment",
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
		)
	}()
	go func() {
		errCh <- a.handler.service.StartOperationsWorker(
			ctx,
			30*time.Duration(cfg.Kafka.OutboxPollIntervalMillis)*time.Millisecond,
		)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (h *Handler) initializePayment(c *gin.Context) {
	var req initializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	tx, err := h.service.InitializePayment(c.Request.Context(), paymentservice.InitializeInput{
		OrderID:        req.OrderID,
		CustomerID:     req.CustomerID,
		AmountMinor:    req.AmountMinor,
		Currency:       "NGN",
		Provider:       req.Provider,
		IdempotencyKey: req.IdempotencyKey,
		CustomerEmail:  req.CustomerEmail,
		CustomerName:   req.CustomerName,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"payment": tx})
}

func (h *Handler) getPayment(c *gin.Context) {
	tx, err := h.service.GetPayment(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"payment": tx})
}

func (h *Handler) refundPayment(c *gin.Context) {
	var req refundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	refund, err := h.service.RefundPayment(c.Request.Context(), paymentservice.RefundInput{
		TransactionID: c.Param("id"),
		AmountMinor:   req.AmountMinor,
		Reason:        req.Reason,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"refund": refund})
}

func (h *Handler) webhook(provider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, err := c.GetRawData()
		if err != nil {
			writeSuccess(c, nethttp.StatusOK, gin.H{"processed": false})
			return
		}
		processed, _ := h.service.HandleWebhook(c.Request.Context(), paymentservice.WebhookInput{
			Provider:  provider,
			Payload:   payload,
			Signature: c.GetHeader("X-Signature"),
		})
		writeSuccess(c, nethttp.StatusOK, gin.H{"processed": processed})
	}
}

func (h *Handler) releaseFunds(c *gin.Context) {
	var req internalReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	if err := h.service.ReleaseFunds(c.Request.Context(), c.Param("id"), req.UserID); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"released": true})
}

func (h *Handler) cancelPayment(c *gin.Context) {
	if err := h.service.CancelPayment(c.Request.Context(), c.Param("id")); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"cancelled": true})
}

func (h *Handler) createPayout(c *gin.Context) {
	var req payoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	payout, err := h.service.CreatePayout(c.Request.Context(), paymentservice.PayoutInput{
		UserID:         req.UserID,
		Provider:       req.Provider,
		AmountMinor:    req.AmountMinor,
		Currency:       req.Currency,
		DestinationRef: req.DestinationRef,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"payout": payout})
}

func (h *Handler) reconcilePayments(c *gin.Context) {
	result, err := h.service.ReconcileTransactions(c.Request.Context())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"reconciliation": result})
}

func (h *Handler) auditRecords(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{"audits": h.service.ListAuditRecords(c.Request.Context())})
}

func (h *Handler) readyz(c *gin.Context) {
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
			"details": apiErr.Details,
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
	case errors.Is(err, platformerrors.ErrConflict):
		writeError(c, platformerrors.ErrConflict)
	default:
		writeError(c, platformerrors.ErrInternal)
	}
}
