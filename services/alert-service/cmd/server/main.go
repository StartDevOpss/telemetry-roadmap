package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/telemetry-platform/alert-service/internal/consumer"
	"github.com/telemetry-platform/telemetryobs"
)

func main() {
	log.SetOutput(os.Stdout)

	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:19092"), ",")
	otlpEndpoint := env("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	tracer, shutdown, err := telemetryobs.InitTracer(ctx, "alert-service", otlpEndpoint)
	if err != nil {
		log.Printf("AVISO: tracer OTel não iniciado: %v", err)
	} else {
		defer shutdown()
	}

	telemetryobs.ServeMetrics(":9090")

	c, err := consumer.New(brokers, tracer)
	if err != nil {
		log.Fatalf("consumer init: %v", err)
	}
	defer c.Close()

	go servirHealth(":8081")
	log.Println("alert-service iniciado | métricas em :9090/metrics")
	c.Run(ctx)
	log.Println("alert-service encerrado")
}

func servirHealth(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("health server: %v", err)
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
