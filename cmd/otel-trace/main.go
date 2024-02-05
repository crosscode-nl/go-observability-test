package main

import (
	"context"
	otel "github.com/crosscode-nl/go-observability-test/pkg/otel"
	"go.opentelemetry.io/otel/attribute"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func simulateWork(ctx context.Context, workName string) {
	_, span := otel.Tracer.Start(ctx, workName)
	defer span.End()

	workDuration := time.Duration(rand.Intn(1000)) * time.Millisecond
	time.Sleep(workDuration)
	span.SetAttributes(attribute.Int64("work.duration", workDuration.Milliseconds()))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelTraceProvider := otel.InitTracer(ctx, "https://github.com/crosscode-nl/go-observability-test/cmd/otel-trace")
	defer cancelTraceProvider()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

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
			ctx, span := otel.Tracer.Start(ctx, "MainOperation")
			simulateWork(ctx, "WorkPart1")
			simulateWork(ctx, "WorkPart2")
			simulateWork(ctx, "WorkPart3")
			span.End()
		}
	}

}
