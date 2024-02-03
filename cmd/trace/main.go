package main

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracer() *sdktrace.TracerProvider {
	ctx := context.Background()

	// Initialize the OTLP exporter using environment variables for configuration.
	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		log.Fatalf("failed to create exporter: %v", err)
	}

	res, err := resource.New(ctx,
		// The service name is now picked up from the OTEL_SERVICE_NAME environment variable.
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp
}

func simulateWork(ctx context.Context, tracer trace.Tracer, workName string) {
	_, span := tracer.Start(ctx, workName)
	defer span.End()

	workDuration := time.Duration(rand.Intn(1000)) * time.Millisecond
	time.Sleep(workDuration)
	span.SetAttributes(attribute.Int64("work.duration", workDuration.Milliseconds()))
}

func main() {
	tp := initTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := otel.Tracer("ExampleTracer")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-signals
		log.Println("Received termination signal, shutting down...")
		cancel()
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ctx, span := tracer.Start(ctx, "MainOperation")
			simulateWork(ctx, tracer, "WorkPart1")
			simulateWork(ctx, tracer, "WorkPart2")
			simulateWork(ctx, tracer, "WorkPart3")
			span.End()
		}
	}

}
