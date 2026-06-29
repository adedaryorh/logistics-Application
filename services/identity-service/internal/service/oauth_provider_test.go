package service

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	platformauth "github.com/adedaryorh/logistics-platform/pkg/auth"
)

func TestJWTSourcedOAuthProvider_ExchangeCode(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	jwksJSON, token := oauthJWKSAndToken(t, privateKey, "kid-google", "https://accounts.google.com", "google-client-id", "google-user-1", "user@example.com", true)

	provider := NewJWTSourcedOAuthProvider(
		"google",
		[]string{"accounts.google.com", "https://accounts.google.com"},
		"google-client-id",
		NewStaticJWKSProvider(jwksJSON),
		OAuthCodeExchangerFunc(func(ctx context.Context, req OAuthExchangeRequest) (*OAuthTokenResponse, error) {
			if req.Code != "auth-code" {
				t.Fatalf("unexpected auth code %q", req.Code)
			}
			return &OAuthTokenResponse{IDToken: token}, nil
		}),
	)

	identity, err := provider.ExchangeCode(context.Background(), OAuthExchangeRequest{Code: "auth-code"})
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if identity.ProviderUserID != "google-user-1" {
		t.Fatalf("expected provider user id google-user-1, got %q", identity.ProviderUserID)
	}
	if identity.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %q", identity.Email)
	}
	if !identity.EmailVerified {
		t.Fatal("expected email to be verified")
	}
}

func TestJWTSourcedOAuthProvider_RealCodeExchangeAndRemoteJWKS(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	jwksJSON, token := oauthJWKSAndToken(t, privateKey, "kid-google", "https://accounts.google.com", "google-client-id", "google-user-remote", "remote@example.com", true)

	var tokenPosts int
	var jwksGets int
	var gotForm url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenPosts++
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm() error = %v", err)
			}
			gotForm = r.PostForm
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(OAuthTokenResponse{
				IDToken:     token,
				AccessToken: "access-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			})
		case "/jwks":
			jwksGets++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(jwksJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := NewJWTSourcedOAuthProvider(
		"google",
		[]string{"accounts.google.com", "https://accounts.google.com"},
		"google-client-id",
		NewRemoteJWKSProvider(server.URL+"/jwks", server.Client(), time.Minute),
		NewHTTPTokenExchanger(
			"google",
			server.URL+"/token",
			"google-client-id",
			"google-client-secret",
			"https://default.example.com/callback",
			nil,
			server.Client(),
		),
	)

	identity, err := provider.ExchangeCode(context.Background(), OAuthExchangeRequest{
		Code:          "real-auth-code",
		RedirectURI:   "https://app.example.com/oauth/callback",
		CodeVerifier:  "pkce-verifier",
		ExpectedNonce: "expected-nonce",
	})
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if identity.ProviderUserID != "google-user-remote" {
		t.Fatalf("expected remote provider user id, got %q", identity.ProviderUserID)
	}
	if identity.Email != "remote@example.com" {
		t.Fatalf("expected remote email, got %q", identity.Email)
	}
	if tokenPosts != 1 {
		t.Fatalf("expected one token exchange request, got %d", tokenPosts)
	}
	if jwksGets != 1 {
		t.Fatalf("expected one jwks request, got %d", jwksGets)
	}
	if gotForm.Get("code") != "real-auth-code" {
		t.Fatalf("expected exchanged code real-auth-code, got %q", gotForm.Get("code"))
	}
	if gotForm.Get("redirect_uri") != "https://app.example.com/oauth/callback" {
		t.Fatalf("expected redirect uri to be forwarded, got %q", gotForm.Get("redirect_uri"))
	}
	if gotForm.Get("code_verifier") != "pkce-verifier" {
		t.Fatalf("expected code verifier to be forwarded, got %q", gotForm.Get("code_verifier"))
	}
	if gotForm.Get("client_secret") != "google-client-secret" {
		t.Fatalf("expected client secret to be forwarded, got %q", gotForm.Get("client_secret"))
	}
}

func TestJWTSourcedOAuthProvider_UserInfoFallback(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	jwksJSON, token := oauthJWKSAndToken(t, privateKey, "kid-google", "https://accounts.google.com", "google-client-id", "google-user-fallback", "", false)

	var tokenPosts int
	var userInfoGets int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenPosts++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(OAuthTokenResponse{
				IDToken:     token,
				AccessToken: "access-token",
				TokenType:   "Bearer",
				ExpiresIn:   3600,
			})
		case "/userinfo":
			userInfoGets++
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"email":"fallback@example.com","email_verified":true}`))
		case "/jwks":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(jwksJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	metrics := NewOAuthRuntimeMetrics()
	provider := NewJWTSourcedOAuthProvider(
		"google",
		[]string{"accounts.google.com", "https://accounts.google.com"},
		"google-client-id",
		NewRemoteJWKSProvider(server.URL+"/jwks", server.Client(), time.Minute),
		NewHTTPTokenExchanger("google", server.URL+"/token", "google-client-id", "secret", "https://app.example.com/oauth/callback", nil, server.Client()),
	).WithUserInfoFetcher(NewHTTPUserInfoFetcher("google", server.URL+"/userinfo", server.Client())).WithMetrics(metrics)

	identity, err := provider.ExchangeCode(context.Background(), OAuthExchangeRequest{
		Code:          "real-auth-code",
		RedirectURI:   "https://app.example.com/oauth/callback",
		ExpectedNonce: "expected-nonce",
	})
	if err != nil {
		t.Fatalf("ExchangeCode() error = %v", err)
	}
	if identity.Email != "fallback@example.com" {
		t.Fatalf("expected fallback email, got %q", identity.Email)
	}
	snapshot := metrics.Snapshot()
	if snapshot.Providers["google"].ProfileFallbacks != 1 {
		t.Fatalf("expected one profile fallback, got %+v", snapshot.Providers["google"])
	}
	if tokenPosts != 1 || userInfoGets != 1 {
		t.Fatalf("expected one token exchange and one userinfo call, got token=%d userinfo=%d", tokenPosts, userInfoGets)
	}
}

func TestSignAppleClientSecret(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	signed, err := signAppleClientSecret(privateKey, "team-id", "key-id", "client-id", time.Unix(1_700_000_000, 0).UTC())
	if err != nil {
		t.Fatalf("signAppleClientSecret() error = %v", err)
	}

	token, err := jwt.ParseWithClaims(signed, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !token.Valid {
		t.Fatal("expected signed apple client secret to be valid")
	}
	if token.Header["kid"] != "key-id" {
		t.Fatalf("expected kid header key-id, got %v", token.Header["kid"])
	}
}

func TestExtractIdentityClaims_EmailVerifiedString(t *testing.T) {
	claimsJSON := `{"email":"user@example.com","email_verified":"true"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))
	token := strings.Join([]string{"header", payload, "signature"}, ".")

	claims, err := extractIdentityClaims(token)
	if err != nil {
		t.Fatalf("extractIdentityClaims() error = %v", err)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %q", claims.Email)
	}
	if !claims.EmailVerified {
		t.Fatal("expected email_verified to coerce to true")
	}
}

func oauthJWKSAndToken(t *testing.T, privateKey *rsa.PrivateKey, kid, issuer, audience, subject, email string, emailVerified bool) ([]byte, string) {
	t.Helper()
	tokenID := mustUUID(t)
	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"jti":            tokenID,
		"iss":            issuer,
		"aud":            []string{audience},
		"sub":            subject,
		"email":          email,
		"email_verified": emailVerified,
		"nonce":          "expected-nonce",
		"exp":            expiresAt.Unix(),
		"iat":            issuedAt.Unix(),
	})
	token.Header["kid"] = kid
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	jwks := platformauth.JWKS{
		Keys: []platformauth.JWK{
			{
				Kty: "RSA",
				Kid: kid,
				N:   base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(localBigEndian(privateKey.PublicKey.E)),
			},
		},
	}
	jwksJSON, err := json.Marshal(jwks)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return jwksJSON, signed
}

func mustUUID(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("NewV7() error = %v", err)
	}
	return id.String()
}

func localBigEndian(value int) []byte {
	if value == 0 {
		return []byte{0}
	}
	out := make([]byte, 0, 4)
	for value > 0 {
		out = append([]byte{byte(value & 0xff)}, out...)
		value >>= 8
	}
	return out
}
