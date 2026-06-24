package openplatform

import (
	"time"

	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/pkg/logx"
)

func accessLogMiddleware(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		query := c.Request.URL.RawQuery
		if query == "" {
			logx.Infof("[openplatform][%s] >>> %s %s ip=%s",
				module, c.Request.Method, path, c.ClientIP())
		} else {
			logx.Infof("[openplatform][%s] >>> %s %s ip=%s query=%s",
				module, c.Request.Method, path, c.ClientIP(), query)
		}
		c.Next()
		logx.Infof("[openplatform][%s] <<< %s %s status=%d duration=%dms",
			module, c.Request.Method, path, c.Writer.Status(), time.Since(start).Milliseconds())
	}
}
