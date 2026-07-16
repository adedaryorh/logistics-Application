package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	platformauth "github.com/adedaryorh/logistics-platform/pkg/auth"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/adedaryorh/logistics-platform/pkg/observability"
	platformredis "github.com/adedaryorh/logistics-platform/pkg/redis"
	"github.com/adedaryorh/logistics-platform/services/identity-service/internal/model"
)

var upperAndDigitRegex = regexp.MustCompile(`[A-Z].*[0-9]|[0-9].*[A-Z]`)

type AuthService struct {
	store          identityStore
	cfg            *platformconfig.Config
	oauthProviders map[string]OAuthProvider
	redis          platformredis.Store
	oauthMetrics   *OAuthRuntimeMetrics
}

type identityStore interface {
	Ping(ctx context.Context) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	CreateUser(ctx context.Context, fullName, email, passwordHash, role string, outboxPayload any) (*model.User, error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateUserProfile(ctx context.Context, id string, phone *string) (*model.User, error)
	SoftDeleteUser(ctx context.Context, id string) error
	CreateRefreshToken(ctx context.Context, token model.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string) error
	RevokeRefreshTokenFamily(ctx context.Context, familyID string) error
	RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error
	CreateMagicLink(ctx context.Context, link model.MagicLink) error
	CountRecentMagicLinksByUser(ctx context.Context, userID string, since time.Time) (int, error)
	GetMagicLinkByHash(ctx context.Context, hash string) (*model.MagicLink, error)
	UseMagicLink(ctx context.Context, id string) error
	MarkUserEmailVerified(ctx context.Context, userID string) error
	CreatePasswordReset(ctx context.Context, reset model.PasswordReset) error
	GetPasswordResetByHash(ctx context.Context, hash string) (*model.PasswordReset, error)
	UsePasswordReset(ctx context.Context, id string) error
	UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error
	UpsertOAuthUser(ctx context.Context, provider, providerUserID, email string, emailVerified bool) (*model.User, error)
	CreateOAuthSession(ctx context.Context, session model.OAuthSession) error
	GetOAuthSessionByID(ctx context.Context, id string) (*model.OAuthSession, error)
	UseOAuthSession(ctx context.Context, id string) error
	DeleteExpiredOAuthSessions(ctx context.Context, before time.Time) (int64, error)
	CreateAuditLog(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error
}

type OAuthProvider interface {
	ExchangeCode(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error)
}

type OAuthProviderFunc func(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error)

func (fn OAuthProviderFunc) ExchangeCode(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error) {
	return fn(ctx, req)
}

type OAuthIdentity struct {
	Provider       string
	ProviderUserID string
	Email          string
	EmailVerified  bool
}

type OAuthSessionStartInput struct {
	RedirectURI         string
	CodeChallenge       *string
	CodeChallengeMethod *string
}

type OAuthSessionStart struct {
	SessionID        string    `json:"session_id"`
	Provider         string    `json:"provider"`
	State            string    `json:"state"`
	Nonce            string    `json:"nonce"`
	ExpiresAt        time.Time `json:"expires_at"`
	AuthorizationURL string    `json:"authorization_url"`
}

type RegisterInput struct {
	FullName string
	Email    string
	Password string
}

type LoginInput struct {
	Email     string
	Password  string
	DeviceID  *string
	IP        string
	UserAgent string
}

type RefreshInput struct {
	RefreshToken string
	DeviceID     *string
	IP           string
	UserAgent    string
}

type UpdateMeInput struct {
	Phone *string
}

type MagicLinkSendInput struct {
	Email string
}

type MagicLinkVerifyInput struct {
	Token string
}

type PasswordResetRequestInput struct {
	Email string
}

type PasswordResetInput struct {
	Token       string
	NewPassword string
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewAuthService(store identityStore, cfg *platformconfig.Config) *AuthService {
	return NewAuthServiceWithProviders(store, cfg, nil)
}

func NewAuthServiceWithProviders(store identityStore, cfg *platformconfig.Config, oauthProviders map[string]OAuthProvider) *AuthService {
	metrics := NewOAuthRuntimeMetrics()
	if oauthProviders == nil {
		oauthProviders = map[string]OAuthProvider{}
	}
	for _, provider := range oauthProviders {
		if jwtProvider, ok := provider.(*JWTSourcedOAuthProvider); ok {
			jwtProvider.WithMetrics(metrics)
		}
	}
	return &AuthService{
		store:          store,
		cfg:            cfg,
		oauthProviders: oauthProviders,
		redis:          platformredis.NewStoreFromConfig(cfg),
		oauthMetrics:   metrics,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*model.User, error) {
	fullName := strings.Join(strings.Fields(input.FullName), " ")
	if len(fullName) < 2 || len(fullName) > 120 {
		return nil, platformerrors.ErrBadRequest
	}
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	existingUser, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("lookup user by email: %w", err)
	}
	if existingUser != nil {
		return nil, platformerrors.ErrConflict
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.store.CreateUser(ctx, fullName, email, string(passwordHash), "customer", map[string]any{
		"full_name": fullName,
		"email":     email,
		"role":      "customer",
	})
	if err != nil {
		return nil, fmt.Errorf("create identity user: %w", err)
	}
	if s.cfg.Security.AutoVerifyEmail {
		if err := s.store.MarkUserEmailVerified(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("auto-verify local identity user: %w", err)
		}
		user.EmailVerified = true
	}
	observability.IncCounter("identity_registrations_total", 1, map[string]string{"role": "customer"})

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, *model.User, error) {
	user, err := s.store.GetUserByEmail(ctx, strings.TrimSpace(strings.ToLower(input.Email)))
	if err != nil {
		return nil, nil, fmt.Errorf("lookup user during login: %w", err)
	}
	if user == nil {
		return nil, nil, platformerrors.ErrUnauthorized
	}
	if !user.EmailVerified {
		return nil, nil, platformerrors.ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, nil, platformerrors.ErrUnauthorized
	}

	pair, err := s.issueTokenPair(ctx, user, input.DeviceID)
	if err != nil {
		return nil, nil, fmt.Errorf("issue login token pair: %w", err)
	}

	if err := s.store.CreateAuditLog(ctx, &user.ID, "login", input.IP, input.UserAgent, map[string]any{
		"email": user.Email,
	}); err != nil {
		return nil, nil, fmt.Errorf("write login audit log: %w", err)
	}
	observability.IncCounter("identity_logins_total", 1, map[string]string{"method": "password"})

	return pair, user, nil
}

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (*TokenPair, error) {
	if s.cfg.JWT.RefreshSecret == "" {
		return nil, fmt.Errorf("refresh secret is required")
	}

	tokenHash := platformauth.HashRefreshTokenWithSecret(input.RefreshToken, s.cfg.JWT.RefreshSecret)
	storedToken, err := s.store.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("lookup refresh token: %w", err)
	}
	if storedToken == nil {
		return nil, platformerrors.ErrUnauthorized
	}
	if storedToken.RevokedAt != nil {
		if err := s.store.RevokeRefreshTokenFamily(ctx, storedToken.FamilyID); err != nil {
			return nil, fmt.Errorf("revoke reused refresh token family: %w", err)
		}
		return nil, platformerrors.ErrUnauthorized
	}
	if storedToken.ExpiresAt.Before(time.Now().UTC()) {
		if err := s.store.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
			return nil, fmt.Errorf("revoke expired refresh token: %w", err)
		}
		return nil, platformerrors.ErrUnauthorized
	}

	user, err := s.store.GetUserByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("lookup refresh token user: %w", err)
	}
	if user == nil {
		return nil, platformerrors.ErrUnauthorized
	}

	if err := s.store.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		return nil, fmt.Errorf("revoke previous refresh token: %w", err)
	}

	pair, err := s.issueTokenPairWithFamily(ctx, user, input.DeviceID, storedToken.FamilyID)
	if err != nil {
		return nil, fmt.Errorf("issue refreshed token pair: %w", err)
	}

	if err := s.store.CreateAuditLog(ctx, &user.ID, "refresh", input.IP, input.UserAgent, map[string]any{
		"family_id": storedToken.FamilyID,
	}); err != nil {
		return nil, fmt.Errorf("write refresh audit log: %w", err)
	}
	observability.IncCounter("identity_token_refresh_total", 1, nil)

	return pair, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (*model.User, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lookup current user: %w", err)
	}
	if user == nil {
		return nil, platformerrors.ErrNotFound
	}
	return user, nil
}

func (s *AuthService) UpdateMe(ctx context.Context, userID string, input UpdateMeInput) (*model.User, error) {
	user, err := s.store.UpdateUserProfile(ctx, userID, input.Phone)
	if err != nil {
		return nil, fmt.Errorf("update current user: %w", err)
	}
	if user == nil {
		return nil, platformerrors.ErrNotFound
	}
	return user, nil
}

func (s *AuthService) DeleteMe(ctx context.Context, userID string, ip, userAgent string) error {
	if err := s.store.SoftDeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("delete current user: %w", err)
	}
	if err := s.store.RevokeAllRefreshTokensForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke user refresh tokens during delete: %w", err)
	}
	if err := s.store.CreateAuditLog(ctx, &userID, "delete_me", ip, userAgent, map[string]any{}); err != nil {
		return fmt.Errorf("write delete audit log: %w", err)
	}
	return nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken, ip, userAgent string) error {
	if s.cfg.JWT.RefreshSecret == "" {
		return fmt.Errorf("refresh secret is required")
	}
	tokenHash := platformauth.HashRefreshTokenWithSecret(refreshToken, s.cfg.JWT.RefreshSecret)
	storedToken, err := s.store.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("lookup logout refresh token: %w", err)
	}
	if storedToken == nil {
		return platformerrors.ErrUnauthorized
	}
	if err := s.store.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		return fmt.Errorf("revoke logout refresh token: %w", err)
	}
	if err := s.store.CreateAuditLog(ctx, &storedToken.UserID, "logout", ip, userAgent, map[string]any{
		"family_id": storedToken.FamilyID,
	}); err != nil {
		return fmt.Errorf("write logout audit log: %w", err)
	}
	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID, ip, userAgent string) error {
	if err := s.store.RevokeAllRefreshTokensForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	if err := s.store.CreateAuditLog(ctx, &userID, "logout_all", ip, userAgent, map[string]any{}); err != nil {
		return fmt.Errorf("write logout all audit log: %w", err)
	}
	return nil
}

func (s *AuthService) SendMagicLink(ctx context.Context, input MagicLinkSendInput) error {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if err := validateEmail(email); err != nil {
		return err
	}
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("lookup magic link user: %w", err)
	}
	if user == nil {
		return platformerrors.ErrNotFound
	}

	count, err := s.store.CountRecentMagicLinksByUser(ctx, user.ID, time.Now().UTC().Add(-10*time.Minute))
	if err != nil {
		return fmt.Errorf("count magic links for user: %w", err)
	}
	cacheCount, _, err := s.redis.IncrWithTTL(ctx, "magic-link:"+email, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("increment magic link redis counter: %w", err)
	}
	if count >= 3 || cacheCount > 3 {
		return platformerrors.ErrRateLimit
	}

	token, err := platformauth.GenerateRefreshToken()
	if err != nil {
		return fmt.Errorf("generate magic link token: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate magic link id: %w", err)
	}

	tokenHash := sha256Hash(token)
	if err := s.store.CreateMagicLink(ctx, model.MagicLink{
		ID:        id.String(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(15 * time.Minute),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("create magic link: %w", err)
	}
	if err := s.store.CreateAuditLog(ctx, &user.ID, "magic_link_send", "", "", map[string]any{
		"email": email,
	}); err != nil {
		return fmt.Errorf("write magic link audit log: %w", err)
	}
	return nil
}

func (s *AuthService) VerifyMagicLink(ctx context.Context, input MagicLinkVerifyInput) (*TokenPair, error) {
	link, err := s.store.GetMagicLinkByHash(ctx, sha256Hash(input.Token))
	if err != nil {
		return nil, fmt.Errorf("lookup magic link: %w", err)
	}
	if link == nil || link.UsedAt != nil || link.ExpiresAt.Before(time.Now().UTC()) {
		return nil, platformerrors.ErrUnauthorized
	}
	if err := s.store.UseMagicLink(ctx, link.ID); err != nil {
		return nil, fmt.Errorf("use magic link: %w", err)
	}
	if err := s.store.MarkUserEmailVerified(ctx, link.UserID); err != nil {
		return nil, fmt.Errorf("mark email verified from magic link: %w", err)
	}
	user, err := s.store.GetUserByID(ctx, link.UserID)
	if err != nil {
		return nil, fmt.Errorf("lookup magic link user: %w", err)
	}
	if user == nil {
		return nil, platformerrors.ErrNotFound
	}
	pair, err := s.issueTokenPair(ctx, user, nil)
	if err != nil {
		return nil, fmt.Errorf("issue token pair from magic link: %w", err)
	}
	return pair, nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, input PasswordResetRequestInput) error {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if err := validateEmail(email); err != nil {
		return err
	}
	user, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("lookup password reset user: %w", err)
	}
	if user == nil {
		return platformerrors.ErrNotFound
	}

	token, err := platformauth.GenerateRefreshToken()
	if err != nil {
		return fmt.Errorf("generate password reset token: %w", err)
	}
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate password reset id: %w", err)
	}

	if err := s.store.CreatePasswordReset(ctx, model.PasswordReset{
		ID:        id.String(),
		UserID:    user.ID,
		TokenHash: sha256Hash(token),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("create password reset: %w", err)
	}
	if err := s.store.CreateAuditLog(ctx, &user.ID, "password_reset_request", "", "", map[string]any{
		"email": email,
	}); err != nil {
		return fmt.Errorf("write password reset request audit log: %w", err)
	}
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, input PasswordResetInput) error {
	if err := validatePassword(input.NewPassword); err != nil {
		return err
	}
	reset, err := s.store.GetPasswordResetByHash(ctx, sha256Hash(input.Token))
	if err != nil {
		return fmt.Errorf("lookup password reset: %w", err)
	}
	if reset == nil || reset.UsedAt != nil || reset.ExpiresAt.Before(time.Now().UTC()) {
		return platformerrors.ErrUnauthorized
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		return fmt.Errorf("hash reset password: %w", err)
	}
	if err := s.store.UpdatePasswordHash(ctx, reset.UserID, string(passwordHash)); err != nil {
		return fmt.Errorf("update reset password hash: %w", err)
	}
	if err := s.store.UsePasswordReset(ctx, reset.ID); err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}
	if err := s.store.RevokeAllRefreshTokensForUser(ctx, reset.UserID); err != nil {
		return fmt.Errorf("revoke refresh tokens after password reset: %w", err)
	}
	return nil
}

func (s *AuthService) GoogleOAuth(ctx context.Context, req OAuthExchangeRequest, ip, userAgent string) (*TokenPair, *model.User, error) {
	return s.oauthLogin(ctx, "google", req, ip, userAgent)
}

func (s *AuthService) AppleOAuth(ctx context.Context, req OAuthExchangeRequest, ip, userAgent string) (*TokenPair, *model.User, error) {
	return s.oauthLogin(ctx, "apple", req, ip, userAgent)
}

func (s *AuthService) StartGoogleOAuthSession(ctx context.Context, input OAuthSessionStartInput, ip, userAgent string) (*OAuthSessionStart, error) {
	return s.startOAuthSession(ctx, "google", input, ip, userAgent)
}

func (s *AuthService) StartAppleOAuthSession(ctx context.Context, input OAuthSessionStartInput, ip, userAgent string) (*OAuthSessionStart, error) {
	return s.startOAuthSession(ctx, "apple", input, ip, userAgent)
}

func (s *AuthService) CleanupExpiredOAuthSessions(ctx context.Context) (int64, error) {
	deleted, err := s.store.DeleteExpiredOAuthSessions(ctx, time.Now().UTC())
	if err != nil {
		if s.oauthMetrics != nil {
			s.oauthMetrics.IncCleanupFailure()
		}
		return 0, fmt.Errorf("cleanup expired oauth sessions: %w", err)
	}
	if s.oauthMetrics != nil {
		s.oauthMetrics.AddCleanupResult(deleted)
	}
	return deleted, nil
}

func (s *AuthService) OAuthMetricsSnapshot() OAuthMetricsSnapshot {
	if s.oauthMetrics == nil {
		return OAuthMetricsSnapshot{}
	}
	return s.oauthMetrics.Snapshot()
}

func (s *AuthService) Ready(ctx context.Context) error {
	if err := s.store.Ping(ctx); err != nil {
		return fmt.Errorf("identity store readiness: %w", err)
	}
	if s.redis != nil {
		if err := s.redis.Ping(ctx); err != nil {
			return fmt.Errorf("identity redis readiness: %w", err)
		}
	}
	return nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *model.User, deviceID *string) (*TokenPair, error) {
	familyID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate refresh family id: %w", err)
	}
	return s.issueTokenPairWithFamily(ctx, user, deviceID, familyID.String())
}

func (s *AuthService) issueTokenPairWithFamily(ctx context.Context, user *model.User, deviceID *string, familyID string) (*TokenPair, error) {
	if s.cfg.JWT.RefreshSecret == "" {
		return nil, fmt.Errorf("refresh secret is required")
	}

	accessToken, err := platformauth.GenerateAccessToken(user.ID, user.Role, []string{"identity:read"})
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := platformauth.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshTokenID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token id: %w", err)
	}

	expiresAt := time.Now().UTC().Add(time.Duration(s.cfg.JWT.RefreshTTL) * time.Second)
	if err := s.store.CreateRefreshToken(ctx, model.RefreshToken{
		ID:        refreshTokenID.String(),
		UserID:    user.ID,
		TokenHash: platformauth.HashRefreshTokenWithSecret(refreshToken, s.cfg.JWT.RefreshSecret),
		FamilyID:  familyID,
		DeviceID:  deviceID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().UTC().Add(time.Duration(s.cfg.JWT.AccessTTL) * time.Second),
	}, nil
}

func validateEmail(email string) error {
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return platformerrors.ErrBadRequest
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return platformerrors.ErrBadRequest
	}
	if !upperAndDigitRegex.MatchString(password) {
		return platformerrors.ErrBadRequest
	}
	return nil
}

func sha256Hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func IsConflict(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate")
}

func (s *AuthService) oauthLogin(ctx context.Context, providerName string, req OAuthExchangeRequest, ip, userAgent string) (*TokenPair, *model.User, error) {
	req.Code = strings.TrimSpace(req.Code)
	req.RedirectURI = strings.TrimSpace(req.RedirectURI)
	if req.Code == "" {
		return nil, nil, platformerrors.ErrBadRequest
	}

	provider := s.oauthProviders[providerName]
	if provider == nil {
		return nil, nil, platformerrors.ErrBadRequest
	}

	if err := s.validateOAuthExchangeRequest(ctx, providerName, &req); err != nil {
		return nil, nil, err
	}

	identity, err := provider.ExchangeCode(ctx, req)
	if err != nil {
		observability.IncCounter("oauth_exchanges_total", 1, map[string]string{"provider": providerName, "status": "failed"})
		return nil, nil, fmt.Errorf("exchange oauth code for %s: %w", providerName, err)
	}
	if identity == nil || strings.TrimSpace(identity.ProviderUserID) == "" || strings.TrimSpace(identity.Email) == "" {
		return nil, nil, platformerrors.ErrBadRequest
	}

	user, err := s.store.UpsertOAuthUser(
		ctx,
		providerName,
		strings.TrimSpace(identity.ProviderUserID),
		strings.TrimSpace(strings.ToLower(identity.Email)),
		identity.EmailVerified,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("upsert oauth user: %w", err)
	}

	pair, err := s.issueTokenPair(ctx, user, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("issue oauth token pair: %w", err)
	}
	if req.SessionID != "" {
		if err := s.store.UseOAuthSession(ctx, req.SessionID); err != nil {
			return nil, nil, fmt.Errorf("consume oauth session: %w", err)
		}
	}

	if err := s.store.CreateAuditLog(ctx, &user.ID, providerName+"_oauth_login", ip, userAgent, map[string]any{
		"provider":         providerName,
		"provider_user_id": identity.ProviderUserID,
		"email":            user.Email,
		"oauth_session_id": req.SessionID,
	}); err != nil {
		return nil, nil, fmt.Errorf("write oauth audit log: %w", err)
	}
	observability.IncCounter("oauth_exchanges_total", 1, map[string]string{"provider": providerName, "status": "success"})

	return pair, user, nil
}

func (s *AuthService) startOAuthSession(ctx context.Context, provider string, input OAuthSessionStartInput, ip, userAgent string) (*OAuthSessionStart, error) {
	if s.oauthProviders[provider] == nil {
		return nil, platformerrors.ErrBadRequest
	}
	input.RedirectURI = strings.TrimSpace(input.RedirectURI)
	if input.RedirectURI == "" {
		return nil, platformerrors.ErrBadRequest
	}
	if err := validatePKCEChallenge(input.CodeChallenge, input.CodeChallengeMethod); err != nil {
		return nil, err
	}
	if err := s.enforceOAuthSessionRateLimit(ctx, provider, ip); err != nil {
		return nil, err
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate oauth session id: %w", err)
	}
	state, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate oauth state: %w", err)
	}
	nonce, err := generateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("generate oauth nonce: %w", err)
	}

	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	session := model.OAuthSession{
		ID:                  sessionID.String(),
		Provider:            provider,
		State:               state,
		Nonce:               nonce,
		RedirectURI:         input.RedirectURI,
		CodeChallenge:       input.CodeChallenge,
		CodeChallengeMethod: input.CodeChallengeMethod,
		ExpiresAt:           expiresAt,
		CreatedAt:           time.Now().UTC(),
	}
	if err := s.store.CreateOAuthSession(ctx, session); err != nil {
		return nil, fmt.Errorf("create oauth session: %w", err)
	}
	if err := s.store.CreateAuditLog(ctx, nil, provider+"_oauth_session_started", ip, userAgent, map[string]any{
		"oauth_session_id": session.ID,
		"redirect_uri":     session.RedirectURI,
	}); err != nil {
		return nil, fmt.Errorf("write oauth session audit log: %w", err)
	}
	if s.oauthMetrics != nil {
		s.oauthMetrics.IncSessionStart(provider)
	}
	observability.IncCounter("oauth_sessions_started_total", 1, map[string]string{"provider": provider})

	authURL, err := BuildAuthorizationURL(OAuthAuthorizationRequest{
		Provider:            provider,
		State:               session.State,
		Nonce:               session.Nonce,
		RedirectURI:         session.RedirectURI,
		CodeChallenge:       session.CodeChallenge,
		CodeChallengeMethod: session.CodeChallengeMethod,
	})
	if err != nil {
		return nil, fmt.Errorf("build oauth authorization url: %w", err)
	}

	return &OAuthSessionStart{
		SessionID:        session.ID,
		Provider:         provider,
		State:            session.State,
		Nonce:            session.Nonce,
		ExpiresAt:        session.ExpiresAt,
		AuthorizationURL: authURL,
	}, nil
}

func (s *AuthService) validateOAuthExchangeRequest(ctx context.Context, provider string, req *OAuthExchangeRequest) error {
	if strings.TrimSpace(req.SessionID) == "" && strings.TrimSpace(req.State) == "" && strings.TrimSpace(req.CodeVerifier) == "" {
		return nil
	}
	if strings.TrimSpace(req.SessionID) == "" || strings.TrimSpace(req.State) == "" {
		return platformerrors.ErrBadRequest
	}
	session, err := s.store.GetOAuthSessionByID(ctx, req.SessionID)
	if err != nil {
		return fmt.Errorf("lookup oauth session: %w", err)
	}
	if session == nil || session.Provider != provider {
		return platformerrors.ErrUnauthorized
	}
	if session.UsedAt != nil || session.ExpiresAt.Before(time.Now().UTC()) {
		return platformerrors.ErrUnauthorized
	}
	if req.State != session.State {
		return platformerrors.ErrUnauthorized
	}
	if req.RedirectURI != "" && req.RedirectURI != session.RedirectURI {
		return platformerrors.ErrUnauthorized
	}
	if err := validatePKCEVerifier(req.CodeVerifier, session.CodeChallenge, session.CodeChallengeMethod); err != nil {
		return err
	}

	req.RedirectURI = session.RedirectURI
	req.ExpectedNonce = session.Nonce

	return nil
}

func validatePKCEChallenge(challenge, method *string) error {
	if challenge == nil && method == nil {
		return nil
	}
	if challenge == nil || strings.TrimSpace(*challenge) == "" {
		return platformerrors.ErrBadRequest
	}
	if method == nil || strings.TrimSpace(*method) == "" {
		return platformerrors.ErrBadRequest
	}
	switch strings.TrimSpace(*method) {
	case "S256", "plain":
		return nil
	default:
		return platformerrors.ErrBadRequest
	}
}

func validatePKCEVerifier(verifier string, challenge, method *string) error {
	if challenge == nil || strings.TrimSpace(*challenge) == "" {
		return nil
	}
	verifier = strings.TrimSpace(verifier)
	if verifier == "" {
		return platformerrors.ErrUnauthorized
	}

	switch strings.TrimSpace(stringValue(method)) {
	case "plain":
		if verifier != strings.TrimSpace(*challenge) {
			return platformerrors.ErrUnauthorized
		}
	case "S256":
		sum := sha256.Sum256([]byte(verifier))
		expected := base64.RawURLEncoding.EncodeToString(sum[:])
		if expected != strings.TrimSpace(*challenge) {
			return platformerrors.ErrUnauthorized
		}
	default:
		return platformerrors.ErrBadRequest
	}
	return nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *AuthService) enforceOAuthSessionRateLimit(ctx context.Context, provider, ip string) error {
	if strings.TrimSpace(ip) == "" {
		ip = "unknown"
	}
	count, _, err := s.redis.IncrWithTTL(ctx, "oauth-session:"+provider+":"+ip, 10*time.Minute)
	if err != nil {
		return fmt.Errorf("increment oauth session redis counter: %w", err)
	}
	if count > 10 {
		if s.oauthMetrics != nil {
			s.oauthMetrics.IncRateLimited(provider)
		}
		return platformerrors.ErrRateLimit
	}
	return nil
}
