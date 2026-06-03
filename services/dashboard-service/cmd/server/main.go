package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/telemetry-platform/dashboard-service/internal/consumer"
)

func main() {
	log.SetOutput(os.Stdout)

	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:19092"), ",")

	c, err := consumer.New(brokers)
	if err != nil {
		log.Fatalf("consumer init: %v", err)
	}
	defer c.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go servirHealth(":8081")
	log.Println("dashboard-service iniciado")
	c.Run(ctx)
	log.Println("dashboard-service encerrado")
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
