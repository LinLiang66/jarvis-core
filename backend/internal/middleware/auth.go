package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/pkg/response"
	"jarvis-core/backend/internal/store"
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	IsRefresh bool   `json:"is_refresh,omitempty"`
	jwt.RegisteredClaims
}

func Auth(cfg *config.Config, sessions *store.SessionCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid || claims.IsRefresh {
			response.Unauthorized(c, "登录已失效")
			c.Abort()
			return
		}
		if sessions != nil && sessions.Enabled() {
			ok, err := sessions.ExistsToken(c.Request.Context(), tokenStr)
			if err != nil || !ok {
				response.Unauthorized(c, "登录已失效")
				c.Abort()
				return
			}
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token", tokenStr)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	if t := strings.TrimSpace(c.Query("token")); t != "" {
		if strings.HasPrefix(strings.ToLower(t), "bearer ") {
			return strings.TrimSpace(t[7:])
		}
		return t
	}
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return strings.TrimSpace(h)
}
