package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"jarvis-core/scheduler/internal/pkg/response"
)

func TokenAuth(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("X-Scheduler-Token"))
		if token == "" {
			token = strings.TrimPrefix(strings.TrimSpace(c.GetHeader("Authorization")), "Bearer ")
		}
		if token == "" || token != expected {
			response.Unauthorized(c, "无效的调度令牌")
			c.Abort()
			return
		}
		c.Next()
	}
}
