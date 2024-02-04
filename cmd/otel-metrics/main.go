package main

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	metricApi "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const meterName = "https://github.com/crosscode-nl/go-observability-test/cmd/otel-metrics"

func initMeter(ctx context.Context) (metricApi.Meter, func()) {

	exp, err := otlpmetrichttp.New(ctx)
	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))

	otel.SetMeterProvider(meterProvider)

	meter := meterProvider.Meter(meterName)
	return meter, func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
}

func main() {
	ctx := context.Background()
	meter, cancelMeterProvider := initMeter(ctx)
	defer cancelMeterProvider()
	// Context to handle cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up channel to catch system signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Ticker for logging every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	requestCounter, err := meter.Int64Counter("request")
	if err != nil {
		panic(err)
	}

	var tempMutex sync.Mutex
	var temperature float64

	_, err = meter.Float64ObservableGauge("temperature", metricApi.WithFloat64Callback(func(ctx context.Context, observer metricApi.Float64Observer) error {
		tempMutex.Lock()
		temp := temperature
		tempMutex.Unlock()
		observer.Observe(temp)
		return nil
	}))

	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				// If context is cancelled, stop the goroutine
				return
			case <-ticker.C:
				// Generate a random temperature
				tempMutex.Lock()
				temperature = rand.Float64() * 35.0 // Random temperature up to 35 degrees Celsius
				tempMutex.Unlock()
				requestCounter.Add(ctx, 1)
			}
		}
	}()

	// Wait for stop signal
	<-signals
}
