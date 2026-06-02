package main

import (
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("device-service iniciando...")
	// TODO Fase 1: consumer telemetry.received → PostgreSQL + Redis → publica device.state.updated
	select {}
}
