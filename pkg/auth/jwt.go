package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	Subject string   `json:"sub,omitempty"`
	Role    string   `json:"role"`
	Scope   []string `json:"scope"`
	Issuer  string   `json:"iss,omitempty"`
	AuthAud []string `json:"aud,omitempty"`
	Issued  int64    `json:"iat,omitempty"`
	Expiry  int64    `json:"exp,omitempty"`
	TokenID string   `json:"jti,omitempty"`
}

// JWTManager handles JWT creation and validation
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(privateKeyPEM, publicKeyPEM []byte, accessTTL, refreshTTl time.Duration) (*JWTManager, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return nil, err
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTl,
	}, nil
}

// GenerateAccessToken creates a new access token
func (j *JWTManager) GenerateAccessToken(userID, role string, scopes []string) (string, error) {
	tokenID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate token id: %w", err)
	}

	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			Issuer:    "logistics-platform",
			Audience:  []string{"logistics-platform"},
			Subject:   userID,
		},
		Subject: userID,
		Role:    role,
		Scope:   scopes,
		Issuer:  "logistics-platform",
		AuthAud: []string{"logistics-platform"},
		Issued:  now.Unix(),
		Expiry:  now.Add(j.accessTTL).Unix(),
		TokenID: tokenID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

// GenerateRefreshToken creates a new refresh token
func (j *JWTManager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// ValidateToken validates and parses a token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// HashRefreshToken hashes a refresh token for storage
func HashRefreshToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// VerifyRefreshToken compares a token with its hash
func VerifyRefreshToken(token, hash string) bool {
	hashed := HashRefreshToken(token)
	return hashed == hash
}

func HashRefreshTokenWithSecret(token, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

func VerifyRefreshTokenWithSecret(token, secret, hash string) bool {
	hashed := HashRefreshTokenWithSecret(token, secret)
	return hmac.Equal([]byte(hashed), []byte(hash))
}

func managerFromEnv() (*JWTManager, error) {
	privateKey := normalizePEM(os.Getenv("JWT_PRIVATE_KEY"))
	if privateKey == "" {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY is required")
	}

	publicKey := normalizePEM(os.Getenv("JWT_PUBLIC_KEY"))
	if publicKey == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY is required")
	}

	manager, err := NewJWTManager([]byte(privateKey), []byte(publicKey), 15*time.Minute, 7*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("new JWT manager: %w", err)
	}

	return manager, nil
}

// normalizePEM removes indentation commonly introduced by multiline .env and
// YAML values. PEM base64 lines cannot contain those leading spaces.
func normalizePEM(value string) string {
	lines := strings.Split(strings.TrimSpace(value), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n") + "\n"
}

func GenerateAccessToken(userID, role string, scopes []string) (string, error) {
	manager, err := managerFromEnv()
	if err != nil {
		return "", err
	}

	token, err := manager.GenerateAccessToken(userID, role, scopes)
	if err != nil {
		return "", fmt.Errorf("generate access token: %w", err)
	}

	return token, nil
}

func GenerateRefreshToken() (string, error) {
	manager, err := managerFromEnv()
	if err != nil {
		return "", err
	}

	token, err := manager.GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	return token, nil
}

func ValidateToken(token string) (*Claims, error) {
	manager, err := managerFromEnv()
	if err != nil {
		return nil, err
	}

	claims, err := manager.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	return claims, nil
}
