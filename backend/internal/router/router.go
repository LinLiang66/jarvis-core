package router

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"jarvis/backend/internal/config"
	"jarvis/backend/internal/database"
	"jarvis/backend/internal/handler/auth"
	"jarvis/backend/internal/handler/health"
	openh "jarvis/backend/internal/handler/openplatform"
	"jarvis/backend/internal/handler/system"
	"jarvis/backend/internal/middleware"
	opsvc "jarvis/backend/internal/service/openplatform"
)

func Setup(cfg *config.Config, app *database.App) *gin.Engine {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,api-key")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", health.New(app).Health)

	api := r.Group("/api/v1")
	authH := auth.New(app, cfg)
	authH.Register(api.Group("/auth"))

	opSession := opsvc.NewSessionStore(app.Redis)
	opStatStore := opsvc.NewStatStore(app.Redis, app.Stores.OpenAPIStat)
	opQuotaStore := opsvc.NewQuotaStore(app.Redis, app.Stores.OpenApp)
	opStatSync := opsvc.NewStatSync(opStatStore, opQuotaStore)
	opStatSync.Start(context.Background())

	opService := opsvc.NewService(app.Stores.OpenApp, app.Stores.OpenAPIStat, app.Stores.OpenAPIAction, opStatStore, opQuotaStore, opSession)

	if n, err := opService.SyncActionRegistry(context.Background()); err != nil {
		log.Printf("[openplatform] action registry sync failed: %v", err)
	} else {
		log.Printf("[openplatform] action registry synced: %d actions", n)
	}
	openh.NewGatewayHandler(opService).Register(api.Group("/open"))

	secured := api.Group("")
	secured.Use(middleware.Auth(cfg, app.Session))
	openAdmin := openh.NewAdminHandler(app.Stores, opService)
	openAdmin.Register(secured.Group("/open-app"))
	openh.NewActionHandler(opService).Register(secured.Group("/open-app"))
	openh.NewDocHandler(opService).Register(secured.Group("/open-app"))
	system.Register(secured, app)

	return r
}
