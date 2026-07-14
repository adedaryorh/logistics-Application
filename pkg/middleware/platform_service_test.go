package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPlatformServiceAuthAcceptsSignatureAndRejectsReplay(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := `{"quote_id":"q1"}`
	timestamp := fmt.Sprint(time.Now().Unix())
	nonce := "nonce-1"
	hash := sha256.Sum256([]byte(body))
	canonical := fmt.Sprintf("POST\n/internal/v1/agricultural/bookings\n%s\n%s\n%s", timestamp, nonce, hex.EncodeToString(hash[:]))
	mac := hmac.New(sha256.New, []byte("farm-secret"))
	mac.Write([]byte(canonical))
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	router := gin.New()
	router.POST("/internal/v1/agricultural/bookings", PlatformServiceAuth(map[string]string{"farmsense": "farm-secret"}, time.Minute), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	call := func() int {
		req := httptest.NewRequest(http.MethodPost, "/internal/v1/agricultural/bookings", strings.NewReader(body))
		req.Header.Set("X-Platform-Service", "farmsense")
		req.Header.Set("X-Platform-Timestamp", timestamp)
		req.Header.Set("X-Platform-Nonce", nonce)
		req.Header.Set("X-Platform-Signature", sig)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		return res.Code
	}
	if got := call(); got != http.StatusNoContent {
		t.Fatalf("first status %d", got)
	}
	if got := call(); got != http.StatusUnauthorized {
		t.Fatalf("replay status %d", got)
	}
}
