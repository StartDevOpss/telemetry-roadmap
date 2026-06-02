package main

import (
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("alert-service iniciando...")
	// TODO Fase 1: consumer telemetry.received → avalia regras → publica alert.triggered
	select {}
}
