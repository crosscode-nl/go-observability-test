package otel

import (
	"context"
	"github.com/agoda-com/opentelemetry-go/otelslog"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"go.opentelemetry.io/otel/sdk/resource"
	"log"
	"log/slog"
	"time"
)

func InitLogger(ctx context.Context) func() {
	// configure opentelemetry logger provider
	logExporter, err := otlplogs.NewExporter(ctx)
	if err != nil {
		log.Fatalf("failed to create log exporter: %v", err)
	}

	res, err := resource.New(ctx,
		// The service name is now picked up from the OTEL_SERVICE_NAME environment variable.
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),   // This option configures a set of Detectors that discover process information
		resource.WithOS(),        // This option configures a set of Detectors that discover OS information
		resource.WithContainer(), // This option configures a set of Detectors that discover container information
		resource.WithHost(),      // This option configures a set of Detectors that discover host information
	)

	if err != nil {
		log.Fatalf("failed to create log resource: %v", err)
	}

	loggerProvider := sdk.NewLoggerProvider(
		sdk.WithBatcher(logExporter),
		sdk.WithResource(res),
	)

	otelLogger := slog.New(otelslog.NewOtelHandler(loggerProvider, &otelslog.HandlerOptions{}))

	//configure default logger
	slog.SetDefault(otelLogger)

	return func() {
		ctx, cancelDeadline := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
		defer cancelDeadline()
		if err := loggerProvider.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down log provider: %v", err)
		}
	}
}
