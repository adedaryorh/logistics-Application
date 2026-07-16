package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net/url"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	platformauth "github.com/adedaryorh/logistics-platform/pkg/auth"
	platformconfig "github.com/adedaryorh/logistics-platform/pkg/config"
	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/adedaryorh/logistics-platform/services/identity-service/internal/model"
)

type mockStore struct {
	pingFn                       func(ctx context.Context) error
	getUserByEmailFn             func(ctx context.Context, email string) (*model.User, error)
	createUserFn                 func(ctx context.Context, fullName, email, passwordHash, role string, outboxPayload any) (*model.User, error)
	getUserByIDFn                func(ctx context.Context, id string) (*model.User, error)
	updateUserProfileFn          func(ctx context.Context, id string, phone *string) (*model.User, error)
	softDeleteUserFn             func(ctx context.Context, id string) error
	createRefreshTokenFn         func(ctx context.Context, token model.RefreshToken) error
	getRefreshTokenByHashFn      func(ctx context.Context, hash string) (*model.RefreshToken, error)
	revokeRefreshTokenFn         func(ctx context.Context, id string) error
	revokeRefreshTokenFamilyFn   func(ctx context.Context, familyID string) error
	revokeAllRefreshTokensFn     func(ctx context.Context, userID string) error
	createMagicLinkFn            func(ctx context.Context, link model.MagicLink) error
	countRecentMagicLinksFn      func(ctx context.Context, userID string, since time.Time) (int, error)
	getMagicLinkByHashFn         func(ctx context.Context, hash string) (*model.MagicLink, error)
	useMagicLinkFn               func(ctx context.Context, id string) error
	markUserEmailVerifiedFn      func(ctx context.Context, userID string) error
	createPasswordResetFn        func(ctx context.Context, reset model.PasswordReset) error
	getPasswordResetByHashFn     func(ctx context.Context, hash string) (*model.PasswordReset, error)
	usePasswordResetFn           func(ctx context.Context, id string) error
	updatePasswordHashFn         func(ctx context.Context, userID, passwordHash string) error
	upsertOAuthUserFn            func(ctx context.Context, provider, providerUserID, email string, emailVerified bool) (*model.User, error)
	createOAuthSessionFn         func(ctx context.Context, session model.OAuthSession) error
	getOAuthSessionByIDFn        func(ctx context.Context, id string) (*model.OAuthSession, error)
	useOAuthSessionFn            func(ctx context.Context, id string) error
	deleteExpiredOAuthSessionsFn func(ctx context.Context, before time.Time) (int64, error)
	createAuditLogFn             func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error
}

func (m *mockStore) Ping(ctx context.Context) error {
	if m.pingFn != nil {
		return m.pingFn(ctx)
	}
	return nil
}

func (m *mockStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *mockStore) CreateUser(ctx context.Context, fullName, email, passwordHash, role string, outboxPayload any) (*model.User, error) {
	if m.createUserFn != nil {
		return m.createUserFn(ctx, fullName, email, passwordHash, role, outboxPayload)
	}
	return nil, nil
}

func (m *mockStore) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockStore) UpdateUserProfile(ctx context.Context, id string, phone *string) (*model.User, error) {
	if m.updateUserProfileFn != nil {
		return m.updateUserProfileFn(ctx, id, phone)
	}
	return nil, nil
}

func (m *mockStore) SoftDeleteUser(ctx context.Context, id string) error {
	if m.softDeleteUserFn != nil {
		return m.softDeleteUserFn(ctx, id)
	}
	return nil
}

func (m *mockStore) CreateRefreshToken(ctx context.Context, token model.RefreshToken) error {
	if m.createRefreshTokenFn != nil {
		return m.createRefreshTokenFn(ctx, token)
	}
	return nil
}

func (m *mockStore) GetRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	if m.getRefreshTokenByHashFn != nil {
		return m.getRefreshTokenByHashFn(ctx, hash)
	}
	return nil, nil
}

func (m *mockStore) RevokeRefreshToken(ctx context.Context, id string) error {
	if m.revokeRefreshTokenFn != nil {
		return m.revokeRefreshTokenFn(ctx, id)
	}
	return nil
}

func (m *mockStore) RevokeRefreshTokenFamily(ctx context.Context, familyID string) error {
	if m.revokeRefreshTokenFamilyFn != nil {
		return m.revokeRefreshTokenFamilyFn(ctx, familyID)
	}
	return nil
}

func (m *mockStore) RevokeAllRefreshTokensForUser(ctx context.Context, userID string) error {
	if m.revokeAllRefreshTokensFn != nil {
		return m.revokeAllRefreshTokensFn(ctx, userID)
	}
	return nil
}

func (m *mockStore) CreateMagicLink(ctx context.Context, link model.MagicLink) error {
	if m.createMagicLinkFn != nil {
		return m.createMagicLinkFn(ctx, link)
	}
	return nil
}

func (m *mockStore) CountRecentMagicLinksByUser(ctx context.Context, userID string, since time.Time) (int, error) {
	if m.countRecentMagicLinksFn != nil {
		return m.countRecentMagicLinksFn(ctx, userID, since)
	}
	return 0, nil
}

func (m *mockStore) GetMagicLinkByHash(ctx context.Context, hash string) (*model.MagicLink, error) {
	if m.getMagicLinkByHashFn != nil {
		return m.getMagicLinkByHashFn(ctx, hash)
	}
	return nil, nil
}

func (m *mockStore) UseMagicLink(ctx context.Context, id string) error {
	if m.useMagicLinkFn != nil {
		return m.useMagicLinkFn(ctx, id)
	}
	return nil
}

func (m *mockStore) MarkUserEmailVerified(ctx context.Context, userID string) error {
	if m.markUserEmailVerifiedFn != nil {
		return m.markUserEmailVerifiedFn(ctx, userID)
	}
	return nil
}

func (m *mockStore) CreatePasswordReset(ctx context.Context, reset model.PasswordReset) error {
	if m.createPasswordResetFn != nil {
		return m.createPasswordResetFn(ctx, reset)
	}
	return nil
}

func (m *mockStore) GetPasswordResetByHash(ctx context.Context, hash string) (*model.PasswordReset, error) {
	if m.getPasswordResetByHashFn != nil {
		return m.getPasswordResetByHashFn(ctx, hash)
	}
	return nil, nil
}

func (m *mockStore) UsePasswordReset(ctx context.Context, id string) error {
	if m.usePasswordResetFn != nil {
		return m.usePasswordResetFn(ctx, id)
	}
	return nil
}

func (m *mockStore) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	if m.updatePasswordHashFn != nil {
		return m.updatePasswordHashFn(ctx, userID, passwordHash)
	}
	return nil
}

func (m *mockStore) UpsertOAuthUser(ctx context.Context, provider, providerUserID, email string, emailVerified bool) (*model.User, error) {
	if m.upsertOAuthUserFn != nil {
		return m.upsertOAuthUserFn(ctx, provider, providerUserID, email, emailVerified)
	}
	return nil, nil
}

func (m *mockStore) CreateOAuthSession(ctx context.Context, session model.OAuthSession) error {
	if m.createOAuthSessionFn != nil {
		return m.createOAuthSessionFn(ctx, session)
	}
	return nil
}

func (m *mockStore) GetOAuthSessionByID(ctx context.Context, id string) (*model.OAuthSession, error) {
	if m.getOAuthSessionByIDFn != nil {
		return m.getOAuthSessionByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockStore) UseOAuthSession(ctx context.Context, id string) error {
	if m.useOAuthSessionFn != nil {
		return m.useOAuthSessionFn(ctx, id)
	}
	return nil
}

func (m *mockStore) DeleteExpiredOAuthSessions(ctx context.Context, before time.Time) (int64, error) {
	if m.deleteExpiredOAuthSessionsFn != nil {
		return m.deleteExpiredOAuthSessionsFn(ctx, before)
	}
	return 0, nil
}

func (m *mockStore) CreateAuditLog(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
	if m.createAuditLogFn != nil {
		return m.createAuditLogFn(ctx, userID, action, ipAddress, userAgent, metadata)
	}
	return nil
}

func TestRegister_Success(t *testing.T) {
	cfg := testConfig()
	store := &mockStore{}
	service := NewAuthService(store, cfg)

	var capturedEmail string
	var capturedRole string
	var capturedPasswordHash string
	var capturedPayload map[string]any

	store.createUserFn = func(ctx context.Context, fullName, email, passwordHash, role string, outboxPayload any) (*model.User, error) {
		capturedEmail = email
		capturedRole = role
		capturedPasswordHash = passwordHash
		capturedPayload = outboxPayload.(map[string]any)
		return &model.User{
			ID:            "user-1",
			FullName:      fullName,
			Email:         email,
			Role:          role,
			EmailVerified: false,
			Status:        "active",
		}, nil
	}

	user, err := service.Register(context.Background(), RegisterInput{
		FullName: "Test Driver",
		Email:    "Test@Example.com",
		Password: "Password1",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if user.Email != "test@example.com" {
		t.Fatalf("expected normalized email, got %q", user.Email)
	}
	if capturedEmail != "test@example.com" {
		t.Fatalf("expected store email to be normalized, got %q", capturedEmail)
	}
	if capturedRole != "customer" {
		t.Fatalf("expected role customer, got %q", capturedRole)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(capturedPasswordHash), []byte("Password1")); err != nil {
		t.Fatalf("expected password to be bcrypt hashed: %v", err)
	}
	if capturedPayload["email"] != "test@example.com" {
		t.Fatalf("expected outbox email payload, got %#v", capturedPayload["email"])
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	cfg := testConfig()
	store := &mockStore{
		getUserByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "existing", Email: email}, nil
		},
	}
	service := NewAuthService(store, cfg)

	_, err := service.Register(context.Background(), RegisterInput{
		FullName: "Test Driver",
		Email:    "test@example.com",
		Password: "Password1",
	})
	if !errors.Is(err, platformerrors.ErrConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestRegister_AutoVerifiesEmailWhenEnabled(t *testing.T) {
	cfg := testConfig()
	cfg.Security.AutoVerifyEmail = true
	verifiedUserID := ""
	store := &mockStore{
		createUserFn: func(ctx context.Context, fullName, email, passwordHash, role string, outboxPayload any) (*model.User, error) {
			return &model.User{ID: "user-1", FullName: fullName, Email: email, Role: role, Status: "active"}, nil
		},
		markUserEmailVerifiedFn: func(ctx context.Context, userID string) error {
			verifiedUserID = userID
			return nil
		},
	}

	user, err := NewAuthService(store, cfg).Register(context.Background(), RegisterInput{
		FullName: "Test Driver", Email: "driver@example.com", Password: "Password1",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if verifiedUserID != user.ID || !user.EmailVerified {
		t.Fatalf("expected auto-verified user, got id=%q verified=%v", verifiedUserID, user.EmailVerified)
	}
}

func TestLogin_Success(t *testing.T) {
	setJWTEnv(t)
	cfg := testConfig()
	passwordHash := mustHashPassword(t, "Password1")
	user := &model.User{
		ID:            "user-1",
		Email:         "test@example.com",
		PasswordHash:  passwordHash,
		EmailVerified: true,
		Role:          "customer",
		Status:        "active",
	}

	var auditCalled bool
	var refreshStored model.RefreshToken
	store := &mockStore{
		getUserByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return user, nil
		},
		createRefreshTokenFn: func(ctx context.Context, token model.RefreshToken) error {
			refreshStored = token
			return nil
		},
		createAuditLogFn: func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
			auditCalled = true
			if action != "login" {
				t.Fatalf("expected login action, got %q", action)
			}
			return nil
		},
	}

	service := NewAuthService(store, cfg)
	pair, gotUser, err := service.Login(context.Background(), LoginInput{
		Email:     "test@example.com",
		Password:  "Password1",
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected token pair, got %+v", pair)
	}
	if gotUser.ID != user.ID {
		t.Fatalf("expected returned user %q, got %q", user.ID, gotUser.ID)
	}
	if refreshStored.UserID != user.ID {
		t.Fatalf("expected refresh token for user %q, got %q", user.ID, refreshStored.UserID)
	}
	if refreshStored.TokenHash == pair.RefreshToken {
		t.Fatalf("refresh token should be hashed before storage")
	}
	if !auditCalled {
		t.Fatalf("expected audit log to be written")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	cfg := testConfig()
	store := &mockStore{
		getUserByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:            "user-1",
				Email:         email,
				PasswordHash:  mustHashPassword(t, "Password1"),
				EmailVerified: true,
			}, nil
		},
	}
	service := NewAuthService(store, cfg)

	_, _, err := service.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "WrongPassword1",
	})
	if !errors.Is(err, platformerrors.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestLogin_UnverifiedEmail(t *testing.T) {
	cfg := testConfig()
	store := &mockStore{
		getUserByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:            "user-1",
				Email:         email,
				PasswordHash:  mustHashPassword(t, "Password1"),
				EmailVerified: false,
			}, nil
		},
	}
	service := NewAuthService(store, cfg)

	_, _, err := service.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "Password1",
	})
	if !errors.Is(err, platformerrors.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestRefreshToken_Rotation(t *testing.T) {
	setJWTEnv(t)
	cfg := testConfig()
	user := &model.User{
		ID:            "user-1",
		Email:         "test@example.com",
		EmailVerified: true,
		Role:          "customer",
		Status:        "active",
	}
	refreshToken := "refresh-token"
	expectedHash := platformauth.HashRefreshTokenWithSecret(refreshToken, cfg.JWT.RefreshSecret)

	var revokedTokenID string
	var newRefresh model.RefreshToken
	var auditCalled bool
	store := &mockStore{
		getRefreshTokenByHashFn: func(ctx context.Context, hash string) (*model.RefreshToken, error) {
			if hash != expectedHash {
				t.Fatalf("expected hashed refresh token %q, got %q", expectedHash, hash)
			}
			return &model.RefreshToken{
				ID:        "rt-1",
				UserID:    user.ID,
				FamilyID:  "family-1",
				TokenHash: expectedHash,
				ExpiresAt: time.Now().UTC().Add(time.Hour),
				CreatedAt: time.Now().UTC(),
			}, nil
		},
		getUserByIDFn: func(ctx context.Context, id string) (*model.User, error) {
			return user, nil
		},
		revokeRefreshTokenFn: func(ctx context.Context, id string) error {
			revokedTokenID = id
			return nil
		},
		createRefreshTokenFn: func(ctx context.Context, token model.RefreshToken) error {
			newRefresh = token
			return nil
		},
		createAuditLogFn: func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
			auditCalled = true
			if action != "refresh" {
				t.Fatalf("expected refresh action, got %q", action)
			}
			auditMetadata, ok := metadata.(map[string]any)
			if !ok {
				t.Fatalf("expected audit metadata map, got %T", metadata)
			}
			if auditMetadata["family_id"] != "family-1" {
				t.Fatalf("expected family id metadata, got %#v", auditMetadata["family_id"])
			}
			return nil
		},
	}
	service := NewAuthService(store, cfg)

	pair, err := service.Refresh(context.Background(), RefreshInput{
		RefreshToken: refreshToken,
		IP:           "127.0.0.1",
		UserAgent:    "test-agent",
	})
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if revokedTokenID != "rt-1" {
		t.Fatalf("expected previous token to be revoked, got %q", revokedTokenID)
	}
	if newRefresh.FamilyID != "family-1" {
		t.Fatalf("expected token family to be preserved, got %q", newRefresh.FamilyID)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected new token pair, got %+v", pair)
	}
	if !auditCalled {
		t.Fatalf("expected refresh audit log to be written")
	}
}

func TestRefreshToken_ReuseDetection(t *testing.T) {
	cfg := testConfig()
	refreshToken := "refresh-token"
	expectedHash := platformauth.HashRefreshTokenWithSecret(refreshToken, cfg.JWT.RefreshSecret)
	now := time.Now().UTC()

	var revokedFamily string
	store := &mockStore{
		getRefreshTokenByHashFn: func(ctx context.Context, hash string) (*model.RefreshToken, error) {
			if hash != expectedHash {
				t.Fatalf("expected hashed refresh token %q, got %q", expectedHash, hash)
			}
			return &model.RefreshToken{
				ID:        "rt-1",
				UserID:    "user-1",
				FamilyID:  "family-1",
				TokenHash: expectedHash,
				ExpiresAt: now.Add(time.Hour),
				RevokedAt: &now,
				CreatedAt: now.Add(-time.Hour),
			}, nil
		},
		revokeRefreshTokenFamilyFn: func(ctx context.Context, familyID string) error {
			revokedFamily = familyID
			return nil
		},
	}
	service := NewAuthService(store, cfg)

	_, err := service.Refresh(context.Background(), RefreshInput{RefreshToken: refreshToken})
	if !errors.Is(err, platformerrors.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
	if revokedFamily != "family-1" {
		t.Fatalf("expected family-1 to be revoked, got %q", revokedFamily)
	}
}

func TestMagicLink_SendAndVerify(t *testing.T) {
	setJWTEnv(t)
	cfg := testConfig()
	user := &model.User{
		ID:            "user-1",
		Email:         "test@example.com",
		EmailVerified: false,
		Role:          "customer",
		Status:        "active",
	}

	var createdLink model.MagicLink
	var usedLinkID string
	var verifiedUserID string
	var refreshStored model.RefreshToken

	store := &mockStore{
		getUserByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return user, nil
		},
		countRecentMagicLinksFn: func(ctx context.Context, userID string, since time.Time) (int, error) {
			return 0, nil
		},
		createMagicLinkFn: func(ctx context.Context, link model.MagicLink) error {
			createdLink = link
			return nil
		},
		createAuditLogFn: func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
			return nil
		},
		getMagicLinkByHashFn: func(ctx context.Context, hash string) (*model.MagicLink, error) {
			if hash != sha256Hash("magic-token") {
				t.Fatalf("expected sha256 hash for verify token, got %q", hash)
			}
			return &model.MagicLink{
				ID:        "ml-1",
				UserID:    user.ID,
				TokenHash: hash,
				ExpiresAt: time.Now().UTC().Add(10 * time.Minute),
				CreatedAt: time.Now().UTC(),
			}, nil
		},
		useMagicLinkFn: func(ctx context.Context, id string) error {
			usedLinkID = id
			return nil
		},
		markUserEmailVerifiedFn: func(ctx context.Context, userID string) error {
			verifiedUserID = userID
			return nil
		},
		getUserByIDFn: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID:            user.ID,
				Email:         user.Email,
				EmailVerified: true,
				Role:          user.Role,
				Status:        user.Status,
			}, nil
		},
		createRefreshTokenFn: func(ctx context.Context, token model.RefreshToken) error {
			refreshStored = token
			return nil
		},
	}
	service := NewAuthService(store, cfg)

	if err := service.SendMagicLink(context.Background(), MagicLinkSendInput{Email: user.Email}); err != nil {
		t.Fatalf("SendMagicLink() error = %v", err)
	}
	if createdLink.UserID != user.ID {
		t.Fatalf("expected magic link for user %q, got %q", user.ID, createdLink.UserID)
	}
	if createdLink.TokenHash == "" {
		t.Fatalf("expected stored magic link hash")
	}

	pair, err := service.VerifyMagicLink(context.Background(), MagicLinkVerifyInput{Token: "magic-token"})
	if err != nil {
		t.Fatalf("VerifyMagicLink() error = %v", err)
	}
	if usedLinkID != "ml-1" {
		t.Fatalf("expected magic link ml-1 to be consumed, got %q", usedLinkID)
	}
	if verifiedUserID != user.ID {
		t.Fatalf("expected user %q to be verified, got %q", user.ID, verifiedUserID)
	}
	if refreshStored.UserID != user.ID {
		t.Fatalf("expected refresh token for user %q, got %q", user.ID, refreshStored.UserID)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected token pair, got %+v", pair)
	}
}

func TestMagicLink_ExpiredToken(t *testing.T) {
	cfg := testConfig()
	store := &mockStore{
		getMagicLinkByHashFn: func(ctx context.Context, hash string) (*model.MagicLink, error) {
			return &model.MagicLink{
				ID:        "ml-1",
				UserID:    "user-1",
				TokenHash: hash,
				ExpiresAt: time.Now().UTC().Add(-time.Minute),
				CreatedAt: time.Now().UTC().Add(-2 * time.Minute),
			}, nil
		},
	}
	service := NewAuthService(store, cfg)

	_, err := service.VerifyMagicLink(context.Background(), MagicLinkVerifyInput{Token: "expired-token"})
	if !errors.Is(err, platformerrors.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestOAuth_Google_NewUser(t *testing.T) {
	setJWTEnv(t)
	cfg := testConfig()
	var capturedProvider string
	var capturedProviderUserID string
	var capturedEmail string
	var capturedEmailVerified bool
	var refreshStored model.RefreshToken

	store := &mockStore{
		upsertOAuthUserFn: func(ctx context.Context, provider, providerUserID, email string, emailVerified bool) (*model.User, error) {
			capturedProvider = provider
			capturedProviderUserID = providerUserID
			capturedEmail = email
			capturedEmailVerified = emailVerified
			return &model.User{
				ID:            "user-1",
				Email:         email,
				EmailVerified: emailVerified,
				Role:          "customer",
				Status:        "active",
			}, nil
		},
		createRefreshTokenFn: func(ctx context.Context, token model.RefreshToken) error {
			refreshStored = token
			return nil
		},
		createAuditLogFn: func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
			if action != "google_oauth_login" {
				t.Fatalf("expected google oauth audit action, got %q", action)
			}
			return nil
		},
	}

	service := NewAuthServiceWithProviders(store, cfg, map[string]OAuthProvider{
		"google": OAuthProviderFunc(func(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error) {
			if req.Code != "google-code" {
				t.Fatalf("expected google code, got %q", req.Code)
			}
			if req.RedirectURI != "https://app.example.com/oauth/callback" {
				t.Fatalf("expected redirect uri to flow through, got %q", req.RedirectURI)
			}
			return &OAuthIdentity{
				Provider:       "google",
				ProviderUserID: "google-user-1",
				Email:          "new@example.com",
				EmailVerified:  true,
			}, nil
		}),
	})

	pair, user, err := service.GoogleOAuth(context.Background(), OAuthExchangeRequest{
		Code:        "google-code",
		RedirectURI: "https://app.example.com/oauth/callback",
	}, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("GoogleOAuth() error = %v", err)
	}
	if capturedProvider != "google" || capturedProviderUserID != "google-user-1" {
		t.Fatalf("unexpected oauth provider data: provider=%q provider_user_id=%q", capturedProvider, capturedProviderUserID)
	}
	if capturedEmail != "new@example.com" || !capturedEmailVerified {
		t.Fatalf("unexpected oauth email data: email=%q verified=%v", capturedEmail, capturedEmailVerified)
	}
	if user.Email != "new@example.com" {
		t.Fatalf("expected oauth user email, got %q", user.Email)
	}
	if refreshStored.UserID != "user-1" {
		t.Fatalf("expected refresh token for oauth user, got %q", refreshStored.UserID)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected oauth token pair, got %+v", pair)
	}
}

func TestStartOAuthSession_Google(t *testing.T) {
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "google-client-id")
	t.Setenv("GOOGLE_OAUTH_AUTHORIZE_URL", "https://accounts.google.com/o/oauth2/v2/auth")
	cfg := testConfig()
	var created model.OAuthSession
	var auditAction string

	store := &mockStore{
		createOAuthSessionFn: func(ctx context.Context, session model.OAuthSession) error {
			created = session
			return nil
		},
		createAuditLogFn: func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
			auditAction = action
			return nil
		},
	}

	service := NewAuthServiceWithProviders(store, cfg, map[string]OAuthProvider{
		"google": OAuthProviderFunc(func(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error) {
			return nil, nil
		}),
	})

	challenge := "challenge-value"
	method := "S256"
	session, err := service.StartGoogleOAuthSession(context.Background(), OAuthSessionStartInput{
		RedirectURI:         "myapp://auth/google",
		CodeChallenge:       &challenge,
		CodeChallengeMethod: &method,
	}, "127.0.0.1", "agent")
	if err != nil {
		t.Fatalf("StartGoogleOAuthSession() error = %v", err)
	}
	if session.SessionID == "" || session.State == "" || session.Nonce == "" {
		t.Fatalf("expected generated oauth session fields, got %+v", session)
	}
	if session.AuthorizationURL == "" {
		t.Fatalf("expected authorization url, got %+v", session)
	}
	if created.Provider != "google" || created.RedirectURI != "myapp://auth/google" {
		t.Fatalf("unexpected created session %+v", created)
	}
	if created.CodeChallenge == nil || *created.CodeChallenge != challenge {
		t.Fatalf("expected code challenge to be stored, got %+v", created.CodeChallenge)
	}
	if auditAction != "google_oauth_session_started" {
		t.Fatalf("expected audit action google_oauth_session_started, got %q", auditAction)
	}
	parsed, err := url.Parse(session.AuthorizationURL)
	if err != nil {
		t.Fatalf("Parse() authorization url error = %v", err)
	}
	values := parsed.Query()
	if values.Get("state") != session.State {
		t.Fatalf("expected state in authorization url, got %q", values.Get("state"))
	}
	if values.Get("nonce") != session.Nonce {
		t.Fatalf("expected nonce in authorization url, got %q", values.Get("nonce"))
	}
	if values.Get("redirect_uri") != "myapp://auth/google" {
		t.Fatalf("expected redirect_uri in authorization url, got %q", values.Get("redirect_uri"))
	}
	if values.Get("code_challenge") != challenge || values.Get("code_challenge_method") != method {
		t.Fatalf("expected PKCE params in authorization url, got %q / %q", values.Get("code_challenge"), values.Get("code_challenge_method"))
	}
}

func TestStartOAuthSession_RateLimited(t *testing.T) {
	t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "google-client-id")
	t.Setenv("GOOGLE_OAUTH_AUTHORIZE_URL", "https://accounts.google.com/o/oauth2/v2/auth")
	cfg := testConfig()
	store := &mockStore{
		createAuditLogFn: func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
			return nil
		},
	}
	service := NewAuthServiceWithProviders(store, cfg, map[string]OAuthProvider{
		"google": OAuthProviderFunc(func(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error) { return nil, nil }),
	})

	for i := 0; i < 10; i++ {
		if _, err := service.StartGoogleOAuthSession(context.Background(), OAuthSessionStartInput{
			RedirectURI: "myapp://auth/google",
		}, "127.0.0.1", "agent"); err != nil {
			t.Fatalf("unexpected rate limit before threshold at attempt %d: %v", i+1, err)
		}
	}

	_, err := service.StartGoogleOAuthSession(context.Background(), OAuthSessionStartInput{
		RedirectURI: "myapp://auth/google",
	}, "127.0.0.1", "agent")
	if !errors.Is(err, platformerrors.ErrRateLimit) {
		t.Fatalf("expected rate limit error, got %v", err)
	}
}

func TestOAuth_Google_SessionValidation(t *testing.T) {
	setJWTEnv(t)
	cfg := testConfig()
	var usedSessionID string

	verifier := "verifier-123"
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	method := "S256"

	store := &mockStore{
		getOAuthSessionByIDFn: func(ctx context.Context, id string) (*model.OAuthSession, error) {
			return &model.OAuthSession{
				ID:                  "oauth-session-1",
				Provider:            "google",
				State:               "expected-state",
				Nonce:               "expected-nonce",
				RedirectURI:         "myapp://auth/google",
				CodeChallenge:       &challenge,
				CodeChallengeMethod: &method,
				ExpiresAt:           time.Now().UTC().Add(5 * time.Minute),
				CreatedAt:           time.Now().UTC(),
			}, nil
		},
		useOAuthSessionFn: func(ctx context.Context, id string) error {
			usedSessionID = id
			return nil
		},
		upsertOAuthUserFn: func(ctx context.Context, provider, providerUserID, email string, emailVerified bool) (*model.User, error) {
			return &model.User{
				ID:            "user-1",
				Email:         email,
				EmailVerified: emailVerified,
				Role:          "customer",
				Status:        "active",
			}, nil
		},
		createRefreshTokenFn: func(ctx context.Context, token model.RefreshToken) error { return nil },
		createAuditLogFn: func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
			return nil
		},
	}

	service := NewAuthServiceWithProviders(store, cfg, map[string]OAuthProvider{
		"google": OAuthProviderFunc(func(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error) {
			if req.ExpectedNonce != "expected-nonce" {
				t.Fatalf("expected nonce to be forwarded, got %q", req.ExpectedNonce)
			}
			if req.RedirectURI != "myapp://auth/google" {
				t.Fatalf("expected redirect uri from session, got %q", req.RedirectURI)
			}
			if req.CodeVerifier != verifier {
				t.Fatalf("expected code verifier to flow through, got %q", req.CodeVerifier)
			}
			return &OAuthIdentity{
				Provider:       "google",
				ProviderUserID: "google-user-1",
				Email:          "oauth@example.com",
				EmailVerified:  true,
			}, nil
		}),
	})

	pair, user, err := service.GoogleOAuth(context.Background(), OAuthExchangeRequest{
		Code:         "google-code",
		RedirectURI:  "myapp://auth/google",
		SessionID:    "oauth-session-1",
		State:        "expected-state",
		CodeVerifier: verifier,
	}, "", "")
	if err != nil {
		t.Fatalf("GoogleOAuth() with session error = %v", err)
	}
	if usedSessionID != "oauth-session-1" {
		t.Fatalf("expected oauth session to be consumed, got %q", usedSessionID)
	}
	if user.Email != "oauth@example.com" || pair.AccessToken == "" {
		t.Fatalf("expected oauth login to succeed, got user=%+v pair=%+v", user, pair)
	}
}

func TestOAuth_Google_ExistingUser(t *testing.T) {
	setJWTEnv(t)
	cfg := testConfig()
	var upsertCalls int

	store := &mockStore{
		upsertOAuthUserFn: func(ctx context.Context, provider, providerUserID, email string, emailVerified bool) (*model.User, error) {
			upsertCalls++
			return &model.User{
				ID:            "existing-user",
				Email:         "existing@example.com",
				EmailVerified: true,
				Role:          "customer",
				Status:        "active",
			}, nil
		},
		createRefreshTokenFn: func(ctx context.Context, token model.RefreshToken) error {
			return nil
		},
		createAuditLogFn: func(ctx context.Context, userID *string, action, ipAddress, userAgent string, metadata any) error {
			return nil
		},
	}

	service := NewAuthServiceWithProviders(store, cfg, map[string]OAuthProvider{
		"google": OAuthProviderFunc(func(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error) {
			return &OAuthIdentity{
				Provider:       "google",
				ProviderUserID: "google-user-existing",
				Email:          "existing@example.com",
				EmailVerified:  true,
			}, nil
		}),
	})

	pair, user, err := service.GoogleOAuth(context.Background(), OAuthExchangeRequest{
		Code: "google-code",
	}, "", "")
	if err != nil {
		t.Fatalf("GoogleOAuth() error = %v", err)
	}
	if upsertCalls != 1 {
		t.Fatalf("expected one oauth upsert call, got %d", upsertCalls)
	}
	if user.ID != "existing-user" {
		t.Fatalf("expected existing user, got %q", user.ID)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatalf("expected oauth token pair, got %+v", pair)
	}
}

func TestCleanupExpiredOAuthSessions(t *testing.T) {
	cfg := testConfig()
	store := &mockStore{
		deleteExpiredOAuthSessionsFn: func(ctx context.Context, before time.Time) (int64, error) {
			return 3, nil
		},
	}
	service := NewAuthServiceWithProviders(store, cfg, nil)

	deleted, err := service.CleanupExpiredOAuthSessions(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpiredOAuthSessions() error = %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected 3 deleted sessions, got %d", deleted)
	}
}

func testConfig() *platformconfig.Config {
	return &platformconfig.Config{
		JWT: platformconfig.JWTConfig{
			RefreshSecret: "test-refresh-secret",
			AccessTTL:     900,
			RefreshTTL:    604800,
		},
	}
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	return string(hash)
}

func setJWTEnv(t *testing.T) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: mustMarshalPKIXPublicKey(t, &privateKey.PublicKey),
	})

	t.Setenv("JWT_PRIVATE_KEY", string(privatePEM))
	t.Setenv("JWT_PUBLIC_KEY", string(publicPEM))
}

func mustMarshalPKIXPublicKey(t *testing.T, key *rsa.PublicKey) []byte {
	t.Helper()
	bytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey() error = %v", err)
	}
	return bytes
}
