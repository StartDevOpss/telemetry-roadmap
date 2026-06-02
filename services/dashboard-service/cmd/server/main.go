package main

import (
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("dashboard-service iniciando...")
	// TODO Fase 1: consumer device.state.updated + alert.triggered → WebSocket/SSE
	select {}
}
