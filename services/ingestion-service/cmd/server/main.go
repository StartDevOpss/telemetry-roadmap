package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/telemetry-platform/ingestion-service/internal/handler"
	"github.com/telemetry-platform/ingestion-service/internal/producer"
)

func main() {
	log.SetOutput(os.Stdout)

	brokers := strings.Split(env("KAFKA_BROKERS", "localhost:19092"), ",")
	addr := env("HTTP_ADDR", ":8081")

	prod, err := producer.New(brokers)
	if err != nil {
		log.Fatalf("kafka producer: %v", err)
	}
	defer prod.Close()

	h := handler.New(prod)

	log.Printf("ingestion-service escutando em %s", addr)
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
