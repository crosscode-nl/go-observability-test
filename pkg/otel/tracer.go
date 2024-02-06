package otel

import (
	"context"
	otelContext "github.com/crosscode-nl/go-observability-test/pkg/otel/context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"
)

func InitTracer(ctx context.Context, name string, opts ...trace.TracerOption) (newCtx context.Context, cancel func()) {

	// Initialize the OTLP exporter using environment variables for configuration.
	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		log.Fatalf("failed to create trace exporter: %v", err)
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
		log.Fatalf("failed to create trace resource: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	ctx = otelContext.WithTracer(ctx, otel.Tracer(name, opts...))

	return ctx, func() {
		ctx, cancelDeadline := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
		defer cancelDeadline()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}
}
