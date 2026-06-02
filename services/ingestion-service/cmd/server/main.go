package main

import (
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("ingestion-service iniciando...")
	// TODO Fase 1: HTTP server + Kafka producer (telemetry.received)
	select {}
}
