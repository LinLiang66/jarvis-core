// bootbench 测量 database.Open 启动耗时（加载 backend/.env）。
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"jarvis-core/backend/internal/config"
	"jarvis-core/backend/internal/database"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	start := time.Now()
	_, err := database.Open(context.Background(), cfg)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Fprintf(os.Stderr, "boot failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("BOOT_MS=%d\n", elapsed.Milliseconds())
	fmt.Printf("BOOT_SECONDS=%.2f\n", elapsed.Seconds())
}
