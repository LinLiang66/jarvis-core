package router

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/database"
	"jarvis-core/backend/internal/handler/auth"
	"jarvis-core/backend/internal/handler/health"
	openh "jarvis-core/backend/internal/handler/openplatform"
	"jarvis-core/backend/internal/handler/system"
	"jarvis-core/backend/internal/middleware"
	"jarvis-core/backend/internal/model"
	storagesvc "jarvis-core/backend/internal/service/storage"
	opsvc "jarvis-core/backend/internal/service/openplatform"
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
	system.Register(secured, app, cfg)

	registerLocalStatic(r, cfg, app)

	return r
}

func registerLocalStatic(r *gin.Engine, cfg *config.Config, app *database.App) {
	list, err := app.Stores.SysStorage.List(context.Background(), model.StorageTypeLocal)
	if err != nil {
		return
	}
	for _, st := range list {
		if st.Status != "0" {
			continue
		}
		root := strings.TrimSpace(st.BucketName)
		if root == "" {
			root = cfg.UploadDir
		}
		_ = storagesvc.EnsureLocalDir(root)
		prefix := strings.TrimRight(cfg.StaticURLPrefix, "/") + "/" + st.Code
		r.Static(prefix, root)
	}
}
