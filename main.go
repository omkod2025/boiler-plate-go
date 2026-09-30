package main

import (
	"context"
	"github.com/omkod2025/boiler-plate-go/app"
	"github.com/omkod2025/boiler-plate-go/configs"
	"github.com/omkod2025/boiler-plate-go/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var timeout = 5 * time.Second

func main() {
	// สร้าง context ที่สามารถ cancel ได้
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// โหลด config
	cfg := configs.LoadConfig(ctx)
	defer cfg.DB.Close()

	// Initialize app และรับ router
	router := app.InitApp(ctx, cfg)
	if router == nil {
		logger.Error("Failed to initialize application")
		return
	}

	// Create a channel to listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Create error channels for both servers
	httpErrors := make(chan error, 1)
	rpcErrors := make(chan error, 1)

	// Create HTTP server instance
	httpServer := &http.Server{
		Addr:        ":" + cfg.Env.APP_PORT,
		Handler:     router,
		ReadTimeout: timeout,
	}

	// Start HTTP server
	go func() {
		logger.Info("HTTP server is running on port " + cfg.Env.APP_PORT)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			httpErrors <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-shutdown:
		logger.Info("Shutdown signal received")
	case err := <-httpErrors:
		logger.Error("HTTP server error", "error", err)
		return
	case err := <-rpcErrors:
		logger.Error("RPC server error", "error", err)
		return
	}

	// Create a context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Gracefully shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	logger.Info("Server is downed.")
}
