package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	platformerrors "github.com/adedaryorh/logistics-platform/pkg/errors"
	"github.com/gin-gonic/gin"
)

// PlatformServiceAuth authenticates FarmSense and TaskAm requests with a
// body-bound HMAC signature and rejects stale or replayed requests.
func PlatformServiceAuth(secrets map[string]string, maxSkew time.Duration) gin.HandlerFunc {
	if maxSkew <= 0 {
		maxSkew = 5 * time.Minute
	}
	var mu sync.Mutex
	seen := map[string]time.Time{}
	return func(c *gin.Context) {
		service := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Platform-Service")))
		secret := secrets[service]
		timestamp := c.GetHeader("X-Platform-Timestamp")
		nonce := strings.TrimSpace(c.GetHeader("X-Platform-Nonce"))
		provided := strings.TrimPrefix(strings.TrimSpace(c.GetHeader("X-Platform-Signature")), "sha256=")
		unix, err := strconv.ParseInt(timestamp, 10, 64)
		if secret == "" || nonce == "" || err != nil || time.Since(time.Unix(unix, 0)).Abs() > maxSkew {
			c.AbortWithStatusJSON(platformerrors.ErrUnauthorized.HTTPStatus, errorBody(c, platformerrors.ErrUnauthorized))
			return
		}
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(platformerrors.ErrBadRequest.HTTPStatus, errorBody(c, platformerrors.ErrBadRequest))
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		digest := sha256.Sum256(body)
		canonical := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", c.Request.Method, c.Request.URL.EscapedPath(), timestamp, nonce, hex.EncodeToString(digest[:]))
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write([]byte(canonical))
		expected := hex.EncodeToString(mac.Sum(nil))
		decodedExpected, e1 := hex.DecodeString(expected)
		decodedProvided, e2 := hex.DecodeString(provided)
		if e1 != nil || e2 != nil || !hmac.Equal(decodedExpected, decodedProvided) {
			c.AbortWithStatusJSON(platformerrors.ErrUnauthorized.HTTPStatus, errorBody(c, platformerrors.ErrUnauthorized))
			return
		}
		now := time.Now().UTC()
		replayKey := service + ":" + nonce
		mu.Lock()
		for key, expiry := range seen {
			if now.After(expiry) {
				delete(seen, key)
			}
		}
		_, replayed := seen[replayKey]
		if !replayed {
			seen[replayKey] = now.Add(maxSkew)
		}
		mu.Unlock()
		if replayed {
			c.AbortWithStatusJSON(platformerrors.ErrUnauthorized.HTTPStatus, errorBody(c, platformerrors.ErrUnauthorized))
			return
		}
		c.Set("platform_service", service)
		c.Next()
	}
}
