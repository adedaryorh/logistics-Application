package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type ValidationOptions struct {
	ExpectedIssuer   string
	ExpectedAudience string
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func ValidateTokenWithOptions(tokenString string, opts ValidationOptions) (*Claims, error) {
	manager, err := managerFromEnv()
	if err != nil {
		return nil, err
	}

	claims, err := manager.ValidateTokenWithOptions(tokenString, opts)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	return claims, nil
}

func (j *JWTManager) ValidateTokenWithOptions(tokenString string, opts ValidationOptions) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if opts.ExpectedIssuer != "" && claims.Issuer != opts.ExpectedIssuer && claims.RegisteredClaims.Issuer != opts.ExpectedIssuer {
		return nil, fmt.Errorf("invalid issuer")
	}
	if opts.ExpectedAudience != "" {
		if !containsString(claims.AuthAud, opts.ExpectedAudience) && !claims.RegisteredClaims.VerifyAudience(opts.ExpectedAudience, true) {
			return nil, fmt.Errorf("invalid audience")
		}
	}
	return claims, nil
}

func VerifyIDTokenWithJWKS(tokenString string, jwksJSON []byte, opts ValidationOptions) (*Claims, error) {
	var jwks JWKS
	if err := json.Unmarshal(jwksJSON, &jwks); err != nil {
		return nil, fmt.Errorf("unmarshal jwks: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		for _, key := range jwks.Keys {
			if key.Kid == kid {
				return jwkToRSAPublicKey(key)
			}
		}
		return nil, fmt.Errorf("kid %q not found", kid)
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if opts.ExpectedIssuer != "" && claims.Issuer != opts.ExpectedIssuer && claims.RegisteredClaims.Issuer != opts.ExpectedIssuer {
		return nil, fmt.Errorf("invalid issuer")
	}
	if opts.ExpectedAudience != "" {
		if !containsString(claims.AuthAud, opts.ExpectedAudience) && !claims.RegisteredClaims.VerifyAudience(opts.ExpectedAudience, true) {
			return nil, fmt.Errorf("invalid audience")
		}
	}
	return claims, nil
}

func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported kty %q", jwk.Kty)
	}
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if strings.TrimSpace(item) == want {
			return true
		}
	}
	return false
}

func decodeJWTPart(value string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode jwt part: %w", err)
	}
	return decoded, nil
}
