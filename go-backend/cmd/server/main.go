package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tid/go-backend/internal/api"
	"tid/go-backend/internal/config"
)

func main() {
	cfg := config.Load()
	logger := log.New(os.Stdout, "tid-api ", log.LstdFlags|log.LUTC)

	app := api.NewApp(cfg, logger)

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           app.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	logger.Printf("starting TID API on %s", addr)
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-stop:
		logger.Printf("shutdown signal: %s", sig)
	case err := <-errCh:
		logger.Fatalf("server failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("graceful shutdown failed: %v", err)
	}
	logger.Println("server stopped")
}