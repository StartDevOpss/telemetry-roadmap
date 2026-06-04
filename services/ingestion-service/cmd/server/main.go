package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/telemetry-platform/ingestion-service/internal/handler"
	"github.com/telemetry-platform/ingestion-service/internal/producer"
	"github.com/telemetry-platform/telemetryobs"
)

func main() {
	log.SetOutput(os.Stdout)

	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:19092"), ",")
	addr := env("HTTP_ADDR", ":8081")
	otlpEndpoint := env("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")

	ctx := context.Background()

	tracer, shutdown, err := telemetryobs.InitTracer(ctx, "ingestion-service", otlpEndpoint)
	if err != nil {
		log.Printf("AVISO: tracer OTel não iniciado: %v", err)
	} else {
		defer shutdown()
	}

	telemetryobs.ServeMetrics(":9090")

	prod, err := producer.New(brokers)
	if err != nil {
		log.Fatalf("kafka producer: %v", err)
	}
	defer prod.Close()

	h := handler.New(prod, tracer)

	log.Printf("ingestion-service escutando em %s | métricas em :9090/metrics", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
