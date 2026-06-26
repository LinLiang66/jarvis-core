package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"jarvis-core/scheduler/internal/config"
	scheddb "jarvis-core/scheduler/internal/database"
	adminh "jarvis-core/scheduler/internal/handler/admin"
	workerh "jarvis-core/scheduler/internal/handler/worker"
	"jarvis-core/scheduler/internal/middleware"
	redisc "jarvis-core/scheduler/internal/redis"
	"jarvis-core/scheduler/internal/service"
	"jarvis-core/scheduler/internal/store"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	stores, err := store.Open(cfg.MySQLDSN, cfg.MySQL)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}

	if err := scheddb.Seed(context.Background(), stores); err != nil {
		log.Fatalf("seed: %v", err)
	}

	if !cfg.RedisEnable {
		log.Fatal("redis: REDIS_ENABLE=false but scheduler requires Redis")
	}
	rdb, err := redisc.NewFromConfig(cfg.Redis)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	engine := service.NewEngine(cfg, stores, rdb)
	if err := engine.Start(context.Background()); err != nil {
		log.Fatalf("engine: %v", err)
	}
	defer engine.Stop()

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	admin := r.Group("/admin/v1", middleware.TokenAuth(cfg.AdminToken))
	adminh.New(stores, engine).Register(admin)

	worker := r.Group("/worker/v1", middleware.TokenAuth(cfg.WorkerToken))
	workerh.New(engine).Register(worker)

	log.Printf("scheduler-server listening on %s", cfg.ServerAddr)
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatal(err)
	}
}
