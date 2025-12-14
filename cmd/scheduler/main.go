package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azizbahloul/gpu-scheduler/pkg/api/rest"
	"github.com/azizbahloul/gpu-scheduler/pkg/scheduler/core"
	"github.com/azizbahloul/gpu-scheduler/pkg/storage/postgres"
	"github.com/azizbahloul/gpu-scheduler/pkg/utils"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	if err := utils.InitLogger("development"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer utils.Sync()

	utils.Info("Starting GPU Scheduler")

	// Load configuration
	config, err := utils.LoadConfig("")
	if err != nil {
		utils.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize storage
	storage, err := postgres.NewPostgresRepository(&config.Database)
	if err != nil {
		utils.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer storage.Close()

	utils.Info("Connected to database")

	// Create scheduler
	scheduler := core.NewScheduler(&config.Scheduler, storage)

	// Start scheduler in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := scheduler.Start(ctx); err != nil {
			utils.Error("Scheduler error", zap.Error(err))
		}
	}()

	utils.Info("Scheduler started")

	// Create HTTP server
	handlers := rest.NewHandlers(scheduler, storage)
	router := rest.NewRouter(handlers)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.API.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server
	go func() {
		utils.Info("Starting HTTP server", zap.Int("port", config.API.HTTPPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	utils.Info("Shutting down scheduler...")

	// Graceful shutdown
	scheduler.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		utils.Error("Server shutdown error", zap.Error(err))
	}

	utils.Info("Scheduler stopped gracefully")
}
