package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	mcpservice "github.com/adedaryorh/logistics-platform/services/mcp-server/internal/service"
)

type Handler struct {
	service *mcpservice.Service
}

type toolCallRequest struct {
	Input map[string]any `json:"input"`
}

func RegisterRoutes(router *gin.Engine, cfg *platformconfig.Config) {
	handler := &Handler{service: mcpservice.New(cfg)}
	router.GET("/readyz", func(c *gin.Context) {
		writeSuccess(c, http.StatusOK, gin.H{"status": "ready"})
	})

	v1 := router.Group("/api/v1/mcp")
	v1.Use(handler.apiKeyAuth())
	v1.GET("/tools", handler.tools)
	v1.POST("/tools/:name/call", handler.callTool)
	v1.GET("/audit", handler.audit)
}

func (h *Handler) apiKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "ApiKey" {
			writeError(c, platformerrors.ErrUnauthorized)
			c.Abort()
			return
		}
		record, err := h.service.Authenticate(parts[1])
		if err != nil {
			writeError(c, platformerrors.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set("mcp_actor", record)
		c.Next()
	}
}

func (h *Handler) tools(c *gin.Context) {
	record := c.MustGet("mcp_actor").(mcpservice.APIKeyRecord)
	writeSuccess(c, http.StatusOK, gin.H{"tools": h.service.ToolsForRole(record.Role)})
}

func (h *Handler) callTool(c *gin.Context) {
	record := c.MustGet("mcp_actor").(mcpservice.APIKeyRecord)
	var req toolCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	output, err := h.service.CallTool(c.Request.Context(), record, c.Param("name"), req.Input)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, http.StatusOK, gin.H{"output": output})
}

func (h *Handler) audit(c *gin.Context) {
	writeSuccess(c, http.StatusOK, gin.H{"entries": h.service.AuditLog()})
}

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"success": true, "data": data, "error": nil, "meta": gin.H{"timestamp": time.Now().UTC(), "version": "v1"}})
}

func writeError(c *gin.Context, apiErr *platformerrors.APIError) {
	c.JSON(apiErr.HTTPStatus, gin.H{"success": false, "data": nil, "error": gin.H{"code": apiErr.Code, "message": apiErr.Message}, "meta": gin.H{"timestamp": time.Now().UTC(), "version": "v1"}})
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, platformerrors.ErrBadRequest):
		writeError(c, platformerrors.ErrBadRequest)
	case errors.Is(err, platformerrors.ErrUnauthorized):
		writeError(c, platformerrors.ErrUnauthorized)
	case errors.Is(err, platformerrors.ErrNotFound):
		writeError(c, platformerrors.ErrNotFound)
	default:
		writeError(c, platformerrors.ErrInternal)
	}
}
