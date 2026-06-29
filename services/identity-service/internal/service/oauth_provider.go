package service

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"

	platformauth "github.com/adedaryorh/logistics-platform/pkg/auth"
)

type OAuthExchangeRequest struct {
	Code          string
	RedirectURI   string
	SessionID     string
	State         string
	CodeVerifier  string
	ExpectedNonce string
}

type OAuthCodeExchanger interface {
	ExchangeCode(ctx context.Context, req OAuthExchangeRequest) (*OAuthTokenResponse, error)
}

type OAuthCodeExchangerFunc func(ctx context.Context, req OAuthExchangeRequest) (*OAuthTokenResponse, error)

func (fn OAuthCodeExchangerFunc) ExchangeCode(ctx context.Context, req OAuthExchangeRequest) (*OAuthTokenResponse, error) {
	return fn(ctx, req)
}

type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type OAuthAuthorizationRequest struct {
	Provider            string
	State               string
	Nonce               string
	RedirectURI         string
	CodeChallenge       *string
	CodeChallengeMethod *string
}

type OAuthUserInfo struct {
	Email         string
	EmailVerified bool
}

type OAuthUserInfoFetcher interface {
	FetchUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error)
}

type JWKSProvider interface {
	JWKS(ctx context.Context) ([]byte, error)
}

type OAuthHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type StaticJWKSProvider struct {
	jwks []byte
}

func NewStaticJWKSProvider(jwks []byte) *StaticJWKSProvider {
	return &StaticJWKSProvider{jwks: append([]byte(nil), jwks...)}
}

func (p *StaticJWKSProvider) JWKS(context.Context) ([]byte, error) {
	if len(p.jwks) == 0 {
		return nil, fmt.Errorf("jwks is empty")
	}
	return append([]byte(nil), p.jwks...), nil
}

type RemoteJWKSProvider struct {
	url       string
	client    OAuthHTTPClient
	cacheTTL  time.Duration
	mu        sync.RWMutex
	cached    []byte
	expiresAt time.Time
}

func NewRemoteJWKSProvider(url string, client OAuthHTTPClient, cacheTTL time.Duration) *RemoteJWKSProvider {
	if client == nil {
		client = http.DefaultClient
	}
	if cacheTTL <= 0 {
		cacheTTL = 15 * time.Minute
	}
	return &RemoteJWKSProvider{
		url:      url,
		client:   client,
		cacheTTL: cacheTTL,
	}
}

func (p *RemoteJWKSProvider) JWKS(ctx context.Context) ([]byte, error) {
	p.mu.RLock()
	if len(p.cached) > 0 && time.Now().UTC().Before(p.expiresAt) {
		cached := append([]byte(nil), p.cached...)
		p.mu.RUnlock()
		return cached, nil
	}
	p.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return nil, fmt.Errorf("build jwks request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read jwks response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jwks request returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	p.mu.Lock()
	p.cached = append([]byte(nil), body...)
	p.expiresAt = time.Now().UTC().Add(p.cacheTTL)
	p.mu.Unlock()

	return append([]byte(nil), body...), nil
}

type HTTPTokenExchanger struct {
	provider           string
	tokenURL           string
	clientID           string
	clientSecret       string
	defaultRedirectURI string
	extraFormValues    map[string]string
	client             OAuthHTTPClient
}

type HTTPUserInfoFetcher struct {
	provider string
	url      string
	client   OAuthHTTPClient
}

func NewHTTPUserInfoFetcher(provider, url string, client OAuthHTTPClient) *HTTPUserInfoFetcher {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPUserInfoFetcher{provider: provider, url: url, client: client}
}

func (f *HTTPUserInfoFetcher) FetchUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s oauth build userinfo request: %w", f.provider, err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s oauth fetch userinfo: %w", f.provider, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%s oauth read userinfo response: %w", f.provider, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s oauth userinfo returned %d: %s", f.provider, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Email         string `json:"email"`
		EmailVerified any    `json:"email_verified"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("%s oauth decode userinfo response: %w", f.provider, err)
	}
	return &OAuthUserInfo{
		Email:         strings.TrimSpace(payload.Email),
		EmailVerified: coerceEmailVerified(payload.EmailVerified),
	}, nil
}

func NewHTTPTokenExchanger(provider, tokenURL, clientID, clientSecret, defaultRedirectURI string, extraFormValues map[string]string, client OAuthHTTPClient) *HTTPTokenExchanger {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPTokenExchanger{
		provider:           provider,
		tokenURL:           tokenURL,
		clientID:           clientID,
		clientSecret:       clientSecret,
		defaultRedirectURI: defaultRedirectURI,
		extraFormValues:    cloneStringMap(extraFormValues),
		client:             client,
	}
}

func (e *HTTPTokenExchanger) ExchangeCode(ctx context.Context, req OAuthExchangeRequest) (*OAuthTokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", strings.TrimSpace(req.Code))
	form.Set("client_id", e.clientID)
	if e.clientSecret != "" {
		form.Set("client_secret", e.clientSecret)
	}
	redirectURI := strings.TrimSpace(req.RedirectURI)
	if redirectURI == "" {
		redirectURI = strings.TrimSpace(e.defaultRedirectURI)
	}
	if redirectURI != "" {
		form.Set("redirect_uri", redirectURI)
	}
	if strings.TrimSpace(req.CodeVerifier) != "" {
		form.Set("code_verifier", strings.TrimSpace(req.CodeVerifier))
	}
	for key, value := range e.extraFormValues {
		if strings.TrimSpace(value) != "" {
			form.Set(key, value)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, e.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%s oauth build token request: %w", e.provider, err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s oauth exchange code: %w", e.provider, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%s oauth read token response: %w", e.provider, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s oauth token endpoint returned %d: %s", e.provider, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%s oauth decode token response: %w", e.provider, err)
	}
	if strings.TrimSpace(tokenResp.IDToken) == "" {
		return nil, fmt.Errorf("%s oauth token response missing id_token", e.provider)
	}
	return &tokenResp, nil
}

type JWTSourcedOAuthProvider struct {
	provider           string
	issuers            []string
	audience           string
	jwksProvider       JWKSProvider
	exchanger          OAuthCodeExchanger
	userInfoFetcher    OAuthUserInfoFetcher
	metrics            *OAuthRuntimeMetrics
	allowDirectIDToken bool
}

func NewJWTSourcedOAuthProvider(provider string, issuers []string, audience string, jwksProvider JWKSProvider, exchanger OAuthCodeExchanger) *JWTSourcedOAuthProvider {
	return &JWTSourcedOAuthProvider{
		provider:           provider,
		issuers:            append([]string(nil), issuers...),
		audience:           audience,
		jwksProvider:       jwksProvider,
		exchanger:          exchanger,
		userInfoFetcher:    nil,
		metrics:            nil,
		allowDirectIDToken: true,
	}
}

func (p *JWTSourcedOAuthProvider) WithUserInfoFetcher(fetcher OAuthUserInfoFetcher) *JWTSourcedOAuthProvider {
	p.userInfoFetcher = fetcher
	return p
}

func (p *JWTSourcedOAuthProvider) WithMetrics(metrics *OAuthRuntimeMetrics) *JWTSourcedOAuthProvider {
	p.metrics = metrics
	return p
}

func (p *JWTSourcedOAuthProvider) ExchangeCode(ctx context.Context, req OAuthExchangeRequest) (*OAuthIdentity, error) {
	tokenResp, err := p.exchange(ctx, req)
	if err != nil {
		return nil, err
	}
	tokenString := strings.TrimSpace(tokenResp.IDToken)
	jwksJSON, err := p.jwksProvider.JWKS(ctx)
	if err != nil {
		if p.metrics != nil {
			p.metrics.IncFailure(p.provider)
		}
		return nil, fmt.Errorf("%s oauth load jwks: %w", p.provider, err)
	}
	claims, err := platformauth.VerifyIDTokenWithJWKS(tokenString, jwksJSON, platformauth.ValidationOptions{
		ExpectedAudience: p.audience,
	})
	if err != nil {
		if p.metrics != nil {
			p.metrics.IncFailure(p.provider)
		}
		return nil, fmt.Errorf("%s oauth verify id token: %w", p.provider, err)
	}
	if !matchesAnyIssuer(claims, p.issuers) {
		if p.metrics != nil {
			p.metrics.IncFailure(p.provider)
		}
		return nil, fmt.Errorf("%s oauth verify id token: invalid issuer", p.provider)
	}

	extraClaims, err := extractIdentityClaims(tokenString)
	if err != nil {
		if p.metrics != nil {
			p.metrics.IncFailure(p.provider)
		}
		return nil, fmt.Errorf("%s oauth decode id token claims: %w", p.provider, err)
	}

	email := strings.TrimSpace(extraClaims.Email)
	emailVerified := extraClaims.EmailVerified
	if email == "" {
		userInfo, fetchErr := p.fetchUserInfo(ctx, tokenResp.AccessToken)
		if fetchErr == nil && userInfo != nil {
			email = strings.TrimSpace(userInfo.Email)
			emailVerified = emailVerified || userInfo.EmailVerified
			if p.metrics != nil {
				p.metrics.IncProfileFallback(p.provider)
			}
		}
	}
	if email == "" {
		if p.metrics != nil {
			p.metrics.IncFailure(p.provider)
		}
		return nil, fmt.Errorf("%s oauth id token missing email", p.provider)
	}
	if req.ExpectedNonce != "" && strings.TrimSpace(extraClaims.Nonce) != strings.TrimSpace(req.ExpectedNonce) {
		if p.metrics != nil {
			p.metrics.IncFailure(p.provider)
		}
		return nil, fmt.Errorf("%s oauth id token nonce mismatch", p.provider)
	}
	if p.metrics != nil {
		p.metrics.IncSuccess(p.provider)
	}

	return &OAuthIdentity{
		Provider:       p.provider,
		ProviderUserID: claims.Subject,
		Email:          strings.ToLower(email),
		EmailVerified:  emailVerified,
	}, nil
}

func (p *JWTSourcedOAuthProvider) exchange(ctx context.Context, req OAuthExchangeRequest) (*OAuthTokenResponse, error) {
	if p.exchanger != nil {
		return p.exchanger.ExchangeCode(ctx, req)
	}
	if p.allowDirectIDToken && strings.Count(strings.TrimSpace(req.Code), ".") == 2 {
		return &OAuthTokenResponse{IDToken: strings.TrimSpace(req.Code)}, nil
	}
	return nil, fmt.Errorf("%s oauth code exchange not configured", p.provider)
}

func (p *JWTSourcedOAuthProvider) fetchUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	if p.userInfoFetcher == nil {
		return nil, fmt.Errorf("userinfo fallback unavailable")
	}
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("access token missing for userinfo fallback")
	}
	return p.userInfoFetcher.FetchUserInfo(ctx, accessToken)
}

func NewOAuthProvidersFromEnv() map[string]OAuthProvider {
	providers := map[string]OAuthProvider{}

	if provider := newGoogleOAuthProviderFromEnv(http.DefaultClient); provider != nil {
		providers["google"] = provider
	}
	if provider := newAppleOAuthProviderFromEnv(http.DefaultClient); provider != nil {
		providers["apple"] = provider
	}

	return providers
}

func newGoogleOAuthProviderFromEnv(client OAuthHTTPClient) OAuthProvider {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
	if clientID == "" {
		return nil
	}

	tokenURL := firstNonEmptyEnv("GOOGLE_OAUTH_TOKEN_URL", "https://oauth2.googleapis.com/token")
	redirectURI := os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")
	clientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")

	return NewJWTSourcedOAuthProvider(
		"google",
		[]string{"accounts.google.com", "https://accounts.google.com"},
		clientID,
		jwksProviderFromEnv("GOOGLE_OAUTH_JWKS_JSON", "GOOGLE_OAUTH_JWKS_URL", "https://www.googleapis.com/oauth2/v3/certs", client),
		NewHTTPTokenExchanger("google", tokenURL, clientID, clientSecret, redirectURI, nil, client),
	).WithUserInfoFetcher(NewHTTPUserInfoFetcher("google", firstNonEmptyEnv("GOOGLE_OAUTH_USERINFO_URL", "https://openidconnect.googleapis.com/v1/userinfo"), client))
}

func newAppleOAuthProviderFromEnv(client OAuthHTTPClient) OAuthProvider {
	clientID := strings.TrimSpace(os.Getenv("APPLE_OAUTH_CLIENT_ID"))
	if clientID == "" {
		return nil
	}

	clientSecret := strings.TrimSpace(os.Getenv("APPLE_OAUTH_CLIENT_SECRET"))
	if clientSecret == "" {
		generated, err := appleClientSecretFromEnv(clientID)
		if err == nil {
			clientSecret = generated
		}
	}
	if clientSecret == "" {
		return nil
	}

	tokenURL := firstNonEmptyEnv("APPLE_OAUTH_TOKEN_URL", "https://appleid.apple.com/auth/token")
	redirectURI := os.Getenv("APPLE_OAUTH_REDIRECT_URI")
	extra := map[string]string{}
	if scope := strings.TrimSpace(os.Getenv("APPLE_OAUTH_SCOPE")); scope != "" {
		extra["scope"] = scope
	}

	return NewJWTSourcedOAuthProvider(
		"apple",
		[]string{"https://appleid.apple.com"},
		clientID,
		jwksProviderFromEnv("APPLE_OAUTH_JWKS_JSON", "APPLE_OAUTH_JWKS_URL", "https://appleid.apple.com/auth/keys", client),
		NewHTTPTokenExchanger("apple", tokenURL, clientID, clientSecret, redirectURI, extra, client),
	)
}

func jwksProviderFromEnv(staticEnvKey, urlEnvKey, defaultURL string, client OAuthHTTPClient) JWKSProvider {
	if jwks := strings.TrimSpace(os.Getenv(staticEnvKey)); jwks != "" {
		return NewStaticJWKSProvider([]byte(jwks))
	}
	url := firstNonEmptyEnv(urlEnvKey, defaultURL)
	return NewRemoteJWKSProvider(url, client, 15*time.Minute)
}

func appleClientSecretFromEnv(clientID string) (string, error) {
	teamID := strings.TrimSpace(os.Getenv("APPLE_OAUTH_TEAM_ID"))
	keyID := strings.TrimSpace(os.Getenv("APPLE_OAUTH_KEY_ID"))
	privateKeyPEM := strings.TrimSpace(os.Getenv("APPLE_OAUTH_PRIVATE_KEY"))
	if teamID == "" || keyID == "" || privateKeyPEM == "" {
		return "", fmt.Errorf("apple oauth client secret env is incomplete")
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("parse apple oauth private key: %w", err)
	}
	return signAppleClientSecret(privateKey, teamID, keyID, clientID, time.Now().UTC())
}

func signAppleClientSecret(privateKey *ecdsa.PrivateKey, teamID, keyID, clientID string, now time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    teamID,
		Subject:   clientID,
		Audience:  []string{"https://appleid.apple.com"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * 24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("sign apple oauth client secret: %w", err)
	}
	return signed, nil
}

type rawIdentityTokenClaims struct {
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Nonce         string `json:"nonce"`
}

type identityTokenClaims struct {
	Email         string
	EmailVerified bool
	Nonce         string
}

func extractIdentityClaims(tokenString string) (*identityTokenClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt parts")
	}
	payload, err := decodeJWTSection(parts[1])
	if err != nil {
		return nil, err
	}
	var claims rawIdentityTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal identity claims: %w", err)
	}
	return &identityTokenClaims{
		Email:         claims.Email,
		EmailVerified: coerceEmailVerified(claims.EmailVerified),
		Nonce:         claims.Nonce,
	}, nil
}

func coerceEmailVerified(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func matchesAnyIssuer(claims *platformauth.Claims, issuers []string) bool {
	actual := strings.TrimSpace(firstNonEmpty(claims.Issuer, claims.RegisteredClaims.Issuer))
	for _, issuer := range issuers {
		if strings.TrimSpace(issuer) == actual {
			return true
		}
	}
	return false
}

func decodeJWTSection(value string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode jwt section: %w", err)
	}
	return decoded, nil
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonEmptyEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func BuildAuthorizationURL(req OAuthAuthorizationRequest) (string, error) {
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("state", req.State)
	values.Set("nonce", req.Nonce)
	values.Set("redirect_uri", req.RedirectURI)

	switch strings.TrimSpace(req.Provider) {
	case "google":
		clientID := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_CLIENT_ID"))
		if clientID == "" {
			return "", fmt.Errorf("google oauth client id is required")
		}
		values.Set("client_id", clientID)
		values.Set("scope", firstNonEmptyEnv("GOOGLE_OAUTH_SCOPE", "openid email profile"))
		values.Set("access_type", firstNonEmptyEnv("GOOGLE_OAUTH_ACCESS_TYPE", "offline"))
		values.Set("prompt", firstNonEmptyEnv("GOOGLE_OAUTH_PROMPT", "consent"))
		if req.CodeChallenge != nil && strings.TrimSpace(*req.CodeChallenge) != "" {
			values.Set("code_challenge", strings.TrimSpace(*req.CodeChallenge))
			values.Set("code_challenge_method", firstNonEmpty(stringPtrValue(req.CodeChallengeMethod), "S256"))
		}
		return firstNonEmptyEnv("GOOGLE_OAUTH_AUTHORIZE_URL", "https://accounts.google.com/o/oauth2/v2/auth") + "?" + values.Encode(), nil
	case "apple":
		clientID := strings.TrimSpace(os.Getenv("APPLE_OAUTH_CLIENT_ID"))
		if clientID == "" {
			return "", fmt.Errorf("apple oauth client id is required")
		}
		values.Set("client_id", clientID)
		values.Set("scope", firstNonEmptyEnv("APPLE_OAUTH_SCOPE", "name email"))
		values.Set("response_mode", firstNonEmptyEnv("APPLE_OAUTH_RESPONSE_MODE", "form_post"))
		if req.CodeChallenge != nil && strings.TrimSpace(*req.CodeChallenge) != "" {
			values.Set("code_challenge", strings.TrimSpace(*req.CodeChallenge))
			values.Set("code_challenge_method", firstNonEmpty(stringPtrValue(req.CodeChallengeMethod), "S256"))
		}
		return firstNonEmptyEnv("APPLE_OAUTH_AUTHORIZE_URL", "https://appleid.apple.com/auth/authorize") + "?" + values.Encode(), nil
	default:
		return "", fmt.Errorf("unsupported oauth provider %q", req.Provider)
	}
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
