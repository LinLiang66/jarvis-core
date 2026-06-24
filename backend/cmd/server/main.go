package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"jarvis/backend/internal/config"
	"jarvis/backend/internal/database"
	"jarvis/backend/internal/pkg/logx"
	"jarvis/backend/internal/router"
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
