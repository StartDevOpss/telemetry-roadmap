package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/telemetry-platform/device-service/internal/consumer"
	"github.com/telemetry-platform/device-service/internal/repository"
	"github.com/telemetry-platform/telemetryobs"
)

func main() {
	log.SetOutput(os.Stdout)

	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:19092"), ",")
	dbURL := env("DATABASE_URL", "postgres://app:app_pass@localhost:5432/telemetry?sslmode=disable")
	redisURL := env("REDIS_URL", "redis://localhost:6379")
	otlpEndpoint := env("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	tracer, shutdown, err := telemetryobs.InitTracer(ctx, "device-service", otlpEndpoint)
	if err != nil {
		log.Printf("AVISO: tracer OTel não iniciado: %v", err)
	} else {
		defer shutdown()
	}

	telemetryobs.ServeMetrics(":9090")

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("db ping: %v", err)
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("redis url: %v", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping: %v", err)
	}

	repo := repository.New(db)
	if err := repo.Migrate(ctx); err != nil {
		log.Fatalf("migration: %v", err)
	}

	c, err := consumer.New(brokers, repo, rdb, tracer)
	if err != nil {
		log.Fatalf("consumer init: %v", err)
	}
	defer c.Close()

	go servirHealth(":8081")
	log.Println("device-service iniciado | métricas em :9090/metrics")
	c.Run(ctx)
	log.Println("device-service encerrado")
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
