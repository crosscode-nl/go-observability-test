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

func main() {
	// Initialize the logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

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
				logger.Info("Current temperature", "temperature", temperature)
			}
		}
	}()

	// Wait for stop signal
	<-signals
	logger.Info("Shutdown signal received, exiting...")
}
