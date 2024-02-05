package main

import (
	"context"
	"github.com/crosscode-nl/go-observability-test/pkg/otel"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// Context to handle cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelLogProvider := otel.InitLogger(ctx)
	defer cancelLogProvider()

	// Initialize the logger
	//logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

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
				slog.Info("Current temperature", "temperature", temperature)
			}
		}
	}()

	// Wait for stop signal
	<-signals
	slog.Info("Shutdown signal received, exiting...")
}
