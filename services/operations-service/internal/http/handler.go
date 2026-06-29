package http

import (
	"context"
	"errors"
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	platformmiddleware "github.com/adedaryorh/logistics-platform/pkg/middleware"
	operationsservice "github.com/adedaryorh/logistics-platform/services/operations-service/internal/service"
)

type Handler struct {
	service *operationsservice.Service
}

type App struct {
	handler *Handler
}

func RegisterRoutes(router *gin.Engine, cfg *platformconfig.Config) {
	NewApp(cfg).RegisterRoutes(router)
}

func NewApp(cfg *platformconfig.Config) *App {
	return &App{
		handler: &Handler{service: operationsservice.New(cfg)},
	}
}

func (a *App) RegisterRoutes(router *gin.Engine) {
	handler := a.handler
	router.GET("/readyz", handler.readyz)

	admin := router.Group("/api/v1/admin")
	admin.Use(platformmiddleware.RequireBearer("admin"))
	admin.GET("/health", handler.health)
	admin.GET("/feature-flags", handler.featureFlags)
	admin.POST("/feature-flags/:name/toggle", handler.toggleFeatureFlag)
	admin.GET("/jobs", handler.jobs)
	admin.POST("/jobs/:id/retry", handler.retryJob)
	admin.GET("/notifications", handler.notifications)
	admin.GET("/incidents", handler.incidents)
	admin.POST("/incidents/:id/ack", handler.ackIncident)
	admin.POST("/incidents/:id/resolve", handler.resolveIncident)
	admin.GET("/workflows", handler.workflows)
	admin.POST("/workflows/:id/retry", handler.retryWorkflow)
	admin.GET("/orders/:id/trace", handler.orderTrace)
}

func (a *App) StartBackground(ctx context.Context, cfg *platformconfig.Config) error {
	if !cfg.Kafka.Enabled {
		<-ctx.Done()
		return ctx.Err()
	}

	consumer := platformkafka.NewConsumer(cfg.Kafka.Brokers).WithSchemaRegistryURL(cfg.Kafka.SchemaRegistryURL)
	group := cfg.Kafka.GroupID + "-operations"

	return consumer.SubscribeAll(ctx,
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "user.registered",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				return handleUserRegisteredEvent(ctx, a.handler.service, msg.Value)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "order.created",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				return handleOrderCreatedEvent(ctx, a.handler.service, msg.Value)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "order.assigned",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				return handleOrderAssignedEvent(ctx, a.handler.service, msg.Value)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "dispatch.offered",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				return handleDispatchOfferedEvent(ctx, a.handler.service, msg.Value)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "payment.failed",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				return handlePaymentFailedEvent(ctx, a.handler.service, msg.Value)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "payment.completed",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				return handlePaymentCompletedEvent(ctx, a.handler.service, msg.Value)
			},
		},
		platformkafka.Subscription{
			Topic:   cfg.Kafka.TopicPrefix + "order.cancelled",
			GroupID: group,
			Handler: func(ctx context.Context, msg kafka.Message) error {
				return handleOrderCancelledEvent(ctx, a.handler.service, msg.Value)
			},
		},
	)
}

func (h *Handler) health(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{"services": h.service.Health(c.Request.Context())})
}

func (h *Handler) featureFlags(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{"feature_flags": h.service.ListFeatureFlags(c.Request.Context())})
}

func (h *Handler) toggleFeatureFlag(c *gin.Context) {
	flag, err := h.service.ToggleFeatureFlag(c.Request.Context(), c.Param("name"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"feature_flag": flag})
}

func (h *Handler) jobs(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{"jobs": h.service.ListJobs(c.Request.Context())})
}

func (h *Handler) retryJob(c *gin.Context) {
	job, err := h.service.RetryJob(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"job": job})
}

func (h *Handler) notifications(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{"notifications": h.service.ListNotifications(c.Request.Context())})
}

func (h *Handler) incidents(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{"incidents": h.service.ListIncidents(c.Request.Context())})
}

func (h *Handler) ackIncident(c *gin.Context) {
	incident, err := h.service.AcknowledgeIncident(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"incident": incident})
}

func (h *Handler) resolveIncident(c *gin.Context) {
	incident, err := h.service.ResolveIncident(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"incident": incident})
}

func (h *Handler) workflows(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{"workflows": h.service.ListWorkflows(c.Request.Context())})
}

func (h *Handler) retryWorkflow(c *gin.Context) {
	workflow, err := h.service.RetryWorkflow(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"workflow": workflow})
}

func (h *Handler) orderTrace(c *gin.Context) {
	trace, err := h.service.GetOrderTrace(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"trace": trace})
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
	case errors.Is(err, platformerrors.ErrNotFound):
		writeError(c, platformerrors.ErrNotFound)
	default:
		writeError(c, platformerrors.ErrInternal)
	}
}
