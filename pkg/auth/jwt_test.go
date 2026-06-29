package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"
)

func TestJWTManagerGenerateAndValidateToken(t *testing.T) {
	t.Parallel()

	privatePEM, publicPEM := mustRSAKeys(t)
	manager, err := NewJWTManager(privatePEM, publicPEM, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager returned error: %v", err)
	}

	token, err := manager.GenerateAccessToken("user-1", "customer", []string{"read"})
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if claims.Subject != "user-1" {
		t.Fatalf("expected subject user-1, got %s", claims.Subject)
	}
	if claims.Role != "customer" {
		t.Fatalf("expected role customer, got %s", claims.Role)
	}
	if len(claims.Scope) != 1 || claims.Scope[0] != "read" {
		t.Fatalf("unexpected scopes: %#v", claims.Scope)
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	t.Parallel()

	privatePEM, publicPEM := mustRSAKeys(t)
	manager, err := NewJWTManager(privatePEM, publicPEM, 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("NewJWTManager returned error: %v", err)
	}

	token, err := manager.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty refresh token")
	}
}

func TestHashRefreshToken(t *testing.T) {
	t.Parallel()

	hash := HashRefreshToken("token-123")
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if !VerifyRefreshToken("token-123", hash) {
		t.Fatal("expected refresh token verification to pass")
	}
}

func mustRSAKeys(t *testing.T) ([]byte, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey returned error: %v", err)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey returned error: %v", err)
	}

	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBytes,
	})

	return privatePEM, publicPEM
}
