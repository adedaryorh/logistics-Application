package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func TestValidateTokenWithOptions(t *testing.T) {
	privateKey, publicKey, manager := testJWTManager(t)
	_ = privateKey
	_ = publicKey

	token, err := manager.GenerateAccessToken("user-1", "customer", []string{"identity:read"})
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	t.Setenv("JWT_PRIVATE_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})))
	pubBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey() error = %v", err)
	}
	t.Setenv("JWT_PUBLIC_KEY", string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})))

	if _, err := ValidateTokenWithOptions(token, ValidationOptions{ExpectedIssuer: "logistics-platform", ExpectedAudience: "logistics-platform"}); err != nil {
		t.Fatalf("ValidateTokenWithOptions() error = %v", err)
	}
	if _, err := ValidateTokenWithOptions(token, ValidationOptions{ExpectedIssuer: "wrong"}); err == nil {
		t.Fatal("expected issuer mismatch")
	}
}

func TestVerifyIDTokenWithJWKS(t *testing.T) {
	privateKey, publicKey, _ := testJWTManager(t)
	keyID := "kid-1"
	token := signJWKSBackedToken(t, privateKey, keyID, "accounts.google.com", "client-id")
	jwks := JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Kid: keyID,
				N:   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
				E:   base64.RawURLEncoding.EncodeToString(bigEndian(publicKey.E)),
			},
		},
	}
	jwksJSON, err := json.Marshal(jwks)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if _, err := VerifyIDTokenWithJWKS(token, jwksJSON, ValidationOptions{ExpectedIssuer: "accounts.google.com", ExpectedAudience: "client-id"}); err != nil {
		t.Fatalf("VerifyIDTokenWithJWKS() error = %v", err)
	}
	if _, err := VerifyIDTokenWithJWKS(token, jwksJSON, ValidationOptions{ExpectedAudience: "wrong"}); err == nil {
		t.Fatal("expected audience mismatch")
	}
}

func testJWTManager(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey, *JWTManager) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey() error = %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER})
	manager, err := NewJWTManager(privatePEM, publicPEM, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager() error = %v", err)
	}
	return privateKey, &privateKey.PublicKey, manager
}

func signJWKSBackedToken(t *testing.T, privateKey *rsa.PrivateKey, kid, issuer, audience string) string {
	t.Helper()
	tokenID, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("NewV7() error = %v", err)
	}
	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			Issuer:    issuer,
			Audience:  []string{audience},
			Subject:   "oauth-user",
		},
		Subject: "oauth-user",
		Role:    "customer",
		Scope:   []string{"openid"},
		Issuer:  issuer,
		AuthAud: []string{audience},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}

func bigEndian(value int) []byte {
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
