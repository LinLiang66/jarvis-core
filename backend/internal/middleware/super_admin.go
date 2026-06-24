package middleware

import (
	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/database"
	"jarvis-core/backend/internal/pkg/response"
	"jarvis-core/backend/internal/service/rbac"
)

func RequireSuperAdmin(app *database.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		user, err := app.Stores.SysUser.GetByID(c.Request.Context(), uid)
		if err != nil || !rbac.IsSuperAdmin(*user) {
			response.Fail(c, 403, "需要超级管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
