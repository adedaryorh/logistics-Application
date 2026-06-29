package health

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CheckFunc represents a health check function
type CheckFunc func(ctx context.Context) error

// Handler returns a gin handler for health checks
func Handler(checks ...CheckFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		results := make(map[string]string)
		allOK := true

		for i, check := range checks {
			name := "check"
			if len(checks) > 1 {
				name = "check_" + string(rune('a'+i))
			}
			if err := check(ctx); err != nil {
				results[name] = err.Error()
				allOK = false
			} else {
				results[name] = "ok"
			}
		}

		if allOK {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"checks": results,
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
				"checks": results,
			})
		}
	}
}
