package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/database"
	"jarvis-core/backend/internal/pkg/logx"
	"jarvis-core/backend/internal/router"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	logx.Init(cfg.LogLevel)
	app, err := database.Open(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	r := router.Setup(cfg, app)
	log.Printf("Go backend listening on %s (db=%s)", cfg.Addr, app.Driver)
	if err := r.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
