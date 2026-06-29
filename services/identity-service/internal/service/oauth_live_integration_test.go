package service

import (
	"context"
	"os"
	"testing"
)

func TestGoogleOAuthLiveIntegration(t *testing.T) {
	if os.Getenv("RUN_LIVE_OAUTH_TESTS") != "1" {
		t.Skip("set RUN_LIVE_OAUTH_TESTS=1 with real provider credentials and a fresh authorization code")
	}

	clientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
	authCode := os.Getenv("GOOGLE_OAUTH_TEST_CODE")
	if clientID == "" || clientSecret == "" || redirectURI == "" || authCode == "" {
		t.Fatal("missing live Google OAuth env vars")
	}

	provider := newGoogleOAuthProviderFromEnv(nil)
	if provider == nil {
		t.Fatal("expected Google OAuth provider from env")
	}

	identity, err := provider.ExchangeCode(context.Background(), OAuthExchangeRequest{
		Code:        authCode,
		RedirectURI: redirectURI,
	})
	if err != nil {
		t.Fatalf("ExchangeCode() live error = %v", err)
	}
	if identity == nil || identity.ProviderUserID == "" || identity.Email == "" {
		t.Fatalf("expected live identity details, got %+v", identity)
	}
}
