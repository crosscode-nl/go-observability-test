package main

import (
	"context"
	"fmt"
	"github.com/agoda-com/opentelemetry-go/otelslog"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"go.opentelemetry.io/otel/sdk/resource"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func initLogger(ctx context.Context) func() {
	res, err := resource.New(ctx,
		// The service name is now picked up from the OTEL_SERVICE_NAME environment variable.
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)

	if err != nil {
		panic(fmt.Sprintf("failed to create exporter: %v", err))
	}

	// configure opentelemetry logger provider
	logExporter, _ := otlplogs.NewExporter(ctx)
	loggerProvider := sdk.NewLoggerProvider(
		sdk.WithBatcher(logExporter),
		sdk.WithResource(res),
	)

	otelLogger := slog.New(otelslog.NewOtelHandler(loggerProvider, &otelslog.HandlerOptions{}))

	//configure default logger
	slog.SetDefault(otelLogger)

	return func() {
		_ = loggerProvider.Shutdown(ctx)
	}
}

func main() {

	// Context to handle cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelLogProvider := initLogger(ctx)
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
