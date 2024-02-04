package main

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func initLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func main() {
	// Initialize the logger
	initLogger()

	// Context to handle cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up channel to catch system signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Ticker for logging every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				// If context is cancelled, stop the goroutine
				return
			case <-ticker.C:
				// Generate a random temperature
				temperature := rand.Float64() * 35.0 // Random temperature up to 35 degrees Celsius
				slog.InfoContext(ctx, "Current temperature", slog.Float64("temperature", temperature))
			}
		}
	}()

	// Wait for stop signal
	<-signals
	slog.InfoContext(ctx, "Shutdown signal received, exiting...")
}
