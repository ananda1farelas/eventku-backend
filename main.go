package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eventku/config"
	"eventku/database"
	"eventku/route"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err) // logger belum ada, jadi pakai panic di titik paling awal ini
	}

	logger := config.NewLogger(cfg)

	pool, err := database.NewPool(cfg)
	if err != nil {
		logger.Fatalf("gagal konek database: %v", err)
	}
	defer pool.Close()

	// TODO: ganti nil dengan helper.ErrorHandler begitu Orang 2 selesai bikinnya
	app := config.NewFiberApp(nil)

	route.RegisterRoutes(app, pool)

	go func() {
		logger.Printf("server jalan di :%s (env: %s)", cfg.AppPort, cfg.AppEnv)
		if err := app.Listen(":" + cfg.AppPort); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown: tunggu sinyal interrupt/terminate, lalu kasih waktu
	// 10 detik buat request yang masih berjalan sebelum dimatikan paksa.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Printf("forced shutdown: %v", err)
	}

	logger.Println("server berhenti dengan bersih")
}
