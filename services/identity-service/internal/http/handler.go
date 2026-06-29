package http

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	platformauth "github.com/adedaryorh/logistics-platform/pkg/auth"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	platformkafka "github.com/adedaryorh/logistics-platform/pkg/kafka"
	platformmiddleware "github.com/adedaryorh/logistics-platform/pkg/middleware"
	platformvalidation "github.com/adedaryorh/logistics-platform/pkg/validation"
	"github.com/adedaryorh/logistics-platform/services/identity-service/internal/repository"
	identityservice "github.com/adedaryorh/logistics-platform/services/identity-service/internal/service"
)

type Handler struct {
	authService *identityservice.AuthService
}

type App struct {
	cfg     *platformconfig.Config
	store   *repository.Store
	handler *Handler
}

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginRequest struct {
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8"`
	DeviceID *string `json:"device_id"`
}

type refreshRequest struct {
	RefreshToken string  `json:"refresh_token" validate:"required"`
	DeviceID     *string `json:"device_id"`
}

type updateMeRequest struct {
	Phone *string `json:"phone"`
}

type tokenOnlyRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type magicLinkVerifyRequest struct {
	Token string `json:"token" validate:"required"`
}

type emailOnlyRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type oauthCodeRequest struct {
	Code         string  `json:"code" validate:"required"`
	RedirectURI  *string `json:"redirect_uri"`
	SessionID    *string `json:"session_id"`
	State        *string `json:"state"`
	CodeVerifier *string `json:"code_verifier"`
}

type oauthSessionRequest struct {
	RedirectURI         string  `json:"redirect_uri" validate:"required"`
	CodeChallenge       *string `json:"code_challenge"`
	CodeChallengeMethod *string `json:"code_challenge_method"`
}

type passwordResetRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

func RegisterRoutes(router *gin.Engine, cfg *platformconfig.Config) {
	NewApp(cfg).RegisterRoutes(router)
}

func NewApp(cfg *platformconfig.Config) *App {
	store, err := repository.NewStore(databaseDSN(cfg))
	if err != nil {
		panic(fmt.Errorf("initialize identity store: %w", err))
	}

	return &App{
		cfg:   cfg,
		store: store,
		handler: &Handler{
			authService: identityservice.NewAuthServiceWithProviders(store, cfg, identityservice.NewOAuthProvidersFromEnv()),
		},
	}
}

func (a *App) RegisterRoutes(router *gin.Engine) {
	handler := a.handler
	router.GET("/readyz", handler.readyz)

	v1 := router.Group("/api/v1")
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", handler.register)
	authGroup.POST("/login", handler.login)
	authGroup.POST("/refresh", handler.refresh)
	authGroup.POST("/logout", handler.logout)
	authGroup.POST("/logout-all", handler.logoutAll)
	authGroup.POST("/magic-link/send", handler.magicLinkSend)
	authGroup.POST("/magic-link/verify", handler.magicLinkVerify)
	authGroup.POST("/oauth/google/session", handler.startGoogleOAuthSession)
	authGroup.POST("/oauth/apple/session", handler.startAppleOAuthSession)
	authGroup.POST("/oauth/google", handler.googleOAuth)
	authGroup.POST("/oauth/apple", handler.appleOAuth)
	authGroup.POST("/password/reset-request", handler.passwordResetRequest)
	authGroup.POST("/password/reset", handler.passwordReset)

	usersGroup := v1.Group("/users")
	usersGroup.GET("/me", handler.me)
	usersGroup.PATCH("/me", handler.updateMe)
	usersGroup.DELETE("/me", handler.deleteMe)

	internal := router.Group("/internal")
	internal.Use(platformmiddleware.InternalOnly(a.cfg, "ops", "ops.admin"))
	internal.GET("/oauth/metrics", handler.oauthMetrics)
}

func (a *App) StartBackground(ctx context.Context, cfg *platformconfig.Config) error {
	errCh := make(chan error, 2)

	if cfg.Kafka.Enabled {
		producer := platformkafka.NewProducer(cfg.Kafka.Brokers).WithSchemaRegistryURL(cfg.Kafka.SchemaRegistryURL)
		defer producer.Close()

		go func() {
			relay := platformkafka.NewOutboxRelay(
				a.store.DB(),
				producer,
				time.Duration(cfg.Kafka.OutboxPollIntervalMillis)*time.Millisecond,
				50,
			).WithTableName("identity_.outbox").WithTopicPrefix(cfg.Kafka.TopicPrefix).WithProducerName("identity-service")
			errCh <- relay.Start(ctx)
		}()
	}

	ticker := time.NewTicker(identityCleanupInterval())
	defer ticker.Stop()

	for {
		select {
		case err := <-errCh:
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, err := a.handler.authService.CleanupExpiredOAuthSessions(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}
		}
	}
}

func (h *Handler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), identityservice.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeSuccess(c, nethttp.StatusCreated, gin.H{
		"user": user,
	})
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}

	pair, user, err := h.authService.Login(c.Request.Context(), identityservice.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		DeviceID:  req.DeviceID,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeSuccess(c, nethttp.StatusOK, gin.H{
		"user":   user,
		"tokens": pair,
	})
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}

	pair, err := h.authService.Refresh(c.Request.Context(), identityservice.RefreshInput{
		RefreshToken: req.RefreshToken,
		DeviceID:     req.DeviceID,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeSuccess(c, nethttp.StatusOK, gin.H{
		"tokens": pair,
	})
}

func (h *Handler) me(c *gin.Context) {
	claims, err := bearerClaims(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		writeError(c, platformerrors.ErrUnauthorized)
		return
	}

	user, err := h.authService.Me(c.Request.Context(), claims.Subject)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	writeSuccess(c, nethttp.StatusOK, gin.H{
		"user": user,
	})
}

func (h *Handler) updateMe(c *gin.Context) {
	claims, err := bearerClaims(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		writeError(c, platformerrors.ErrUnauthorized)
		return
	}
	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	user, err := h.authService.UpdateMe(c.Request.Context(), claims.Subject, identityservice.UpdateMeInput{Phone: req.Phone})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"user": user})
}

func (h *Handler) deleteMe(c *gin.Context) {
	claims, err := bearerClaims(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		writeError(c, platformerrors.ErrUnauthorized)
		return
	}
	if err := h.authService.DeleteMe(c.Request.Context(), claims.Subject, c.ClientIP(), c.Request.UserAgent()); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) logout(c *gin.Context) {
	var req tokenOnlyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken, c.ClientIP(), c.Request.UserAgent()); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"logged_out": true})
}

func (h *Handler) logoutAll(c *gin.Context) {
	claims, err := bearerClaims(c.Request.Context(), c.GetHeader("Authorization"))
	if err != nil {
		writeError(c, platformerrors.ErrUnauthorized)
		return
	}
	if err := h.authService.LogoutAll(c.Request.Context(), claims.Subject, c.ClientIP(), c.Request.UserAgent()); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"logged_out_all": true})
}

func (h *Handler) magicLinkSend(c *gin.Context) {
	var req emailOnlyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	if err := h.authService.SendMagicLink(c.Request.Context(), identityservice.MagicLinkSendInput{Email: req.Email}); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusAccepted, gin.H{"sent": true})
}

func (h *Handler) magicLinkVerify(c *gin.Context) {
	var req magicLinkVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	pair, err := h.authService.VerifyMagicLink(c.Request.Context(), identityservice.MagicLinkVerifyInput{Token: req.Token})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"tokens": pair})
}

func (h *Handler) passwordResetRequest(c *gin.Context) {
	var req emailOnlyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}
	if err := h.authService.RequestPasswordReset(c.Request.Context(), identityservice.PasswordResetRequestInput{Email: req.Email}); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusAccepted, gin.H{"requested": true})
}

func (h *Handler) passwordReset(c *gin.Context) {
	var req passwordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if err := h.authService.ResetPassword(c.Request.Context(), identityservice.PasswordResetInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{"reset": true})
}

func (h *Handler) googleOAuth(c *gin.Context) {
	var req oauthCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	pair, user, err := h.authService.GoogleOAuth(c.Request.Context(), identityservice.OAuthExchangeRequest{
		Code:         req.Code,
		RedirectURI:  stringValue(req.RedirectURI),
		SessionID:    stringValue(req.SessionID),
		State:        stringValue(req.State),
		CodeVerifier: stringValue(req.CodeVerifier),
	}, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{
		"user":   user,
		"tokens": pair,
	})
}

func (h *Handler) appleOAuth(c *gin.Context) {
	var req oauthCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	pair, user, err := h.authService.AppleOAuth(c.Request.Context(), identityservice.OAuthExchangeRequest{
		Code:         req.Code,
		RedirectURI:  stringValue(req.RedirectURI),
		SessionID:    stringValue(req.SessionID),
		State:        stringValue(req.State),
		CodeVerifier: stringValue(req.CodeVerifier),
	}, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusOK, gin.H{
		"user":   user,
		"tokens": pair,
	})
}

func (h *Handler) startGoogleOAuthSession(c *gin.Context) {
	h.startOAuthSession(c, "google")
}

func (h *Handler) startAppleOAuthSession(c *gin.Context) {
	h.startOAuthSession(c, "apple")
}

func (h *Handler) startOAuthSession(c *gin.Context, provider string) {
	var req oauthSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if errs := platformvalidation.ValidateStruct(req); len(errs) > 0 {
		writeError(c, platformerrors.WithDetails(platformerrors.ErrBadRequest, errs))
		return
	}

	var (
		session *identityservice.OAuthSessionStart
		err     error
	)
	input := identityservice.OAuthSessionStartInput{
		RedirectURI:         req.RedirectURI,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
	}
	switch provider {
	case "google":
		session, err = h.authService.StartGoogleOAuthSession(c.Request.Context(), input, c.ClientIP(), c.Request.UserAgent())
	case "apple":
		session, err = h.authService.StartAppleOAuthSession(c.Request.Context(), input, c.ClientIP(), c.Request.UserAgent())
	default:
		writeError(c, platformerrors.ErrBadRequest)
		return
	}
	if err != nil {
		writeServiceError(c, err)
		return
	}
	writeSuccess(c, nethttp.StatusCreated, gin.H{"oauth_session": session})
}

func (h *Handler) readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.authService.Ready(ctx); err != nil {
		writeError(c, platformerrors.ErrInternal)
		return
	}

	writeSuccess(c, nethttp.StatusOK, gin.H{
		"status": "ready",
	})
}

func (h *Handler) oauthMetrics(c *gin.Context) {
	writeSuccess(c, nethttp.StatusOK, gin.H{
		"oauth_metrics": h.authService.OAuthMetricsSnapshot(),
	})
}

func bearerClaims(ctx context.Context, authHeader string) (*platformauth.Claims, error) {
	if authHeader == "" {
		return nil, platformerrors.ErrUnauthorized
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, platformerrors.ErrUnauthorized
	}

	claims, err := platformauth.ValidateTokenWithOptions(parts[1], platformauth.ValidationOptions{
		ExpectedIssuer:   "logistics-platform",
		ExpectedAudience: "logistics-platform",
	})
	if err != nil {
		return nil, fmt.Errorf("validate bearer token: %w", err)
	}
	return claims, nil
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

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func identityCleanupInterval() time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(os.Getenv("OAUTH_SESSION_CLEANUP_INTERVAL_SECONDS")))
	if err != nil || seconds <= 0 {
		return 10 * time.Minute
	}
	return time.Duration(seconds) * time.Second
}

func writeSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
		"error":   nil,
		"meta": gin.H{
			"request_id": c.GetString("request_id"),
			"version":    "v1",
			"timestamp":  time.Now().UTC(),
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
			"version":    "v1",
			"timestamp":  time.Now().UTC(),
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
	case errors.Is(err, platformerrors.ErrRateLimit):
		writeError(c, platformerrors.ErrRateLimit)
	case errors.Is(err, platformerrors.ErrConflict):
		writeError(c, platformerrors.ErrConflict)
	case errors.Is(err, platformerrors.ErrNotFound):
		writeError(c, platformerrors.ErrNotFound)
	default:
		if errors.Is(err, sql.ErrNoRows) {
			writeError(c, platformerrors.ErrNotFound)
			return
		}
		writeError(c, platformerrors.ErrInternal)
	}
}
