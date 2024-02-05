package otel

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	metricApi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"log"
	"time"
)

var Meter metricApi.Meter

func InitMeter(ctx context.Context, name string, options ...metricApi.MeterOption) (cancel func()) {

	exp, err := otlpmetrichttp.New(ctx)
	if err != nil {
		log.Fatalf("failed to create metric exporter: %v", err)
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
		log.Fatalf("failed to create metrics resource: %v", err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(15*time.Second))),
		metric.WithResource(res))

	otel.SetMeterProvider(meterProvider)

	Meter = meterProvider.Meter(name, options...)
	return func() {
		ctx, cancelDeadline := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
		defer cancelDeadline()
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down metrics provider: %v", err)
		}
	}
}
