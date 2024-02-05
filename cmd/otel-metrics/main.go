package main

import (
	"context"
	"github.com/crosscode-nl/go-observability-test/pkg/otel"
	metricApi "go.opentelemetry.io/otel/metric"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const meterName = "https://github.com/crosscode-nl/go-observability-test/cmd/otel-metrics"

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelMeterProvider := otel.InitMeter(ctx, meterName)
	defer cancelMeterProvider()
	// Context to handle cancellation

	// Set up channel to catch system signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Ticker for logging every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	requestCounter, err := otel.Meter.Int64Counter("request")
	if err != nil {
		panic(err)
	}

	var tempMutex sync.Mutex
	var temperature float64

	_, err = otel.Meter.Float64ObservableGauge("temperature", metricApi.WithFloat64Callback(func(ctx context.Context, observer metricApi.Float64Observer) error {
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
