package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/telemetry-platform/events"
	"github.com/telemetry-platform/ingestion-service/internal/producer"
)

type telemetryRequest struct {
	DeviceID string                 `json:"device_id"`
	Payload  events.TelemetryPayload `json:"payload"`
}

type Handler struct {
	prod *producer.EventProducer
}

func New(prod *producer.EventProducer) *Handler {
	return &Handler{prod: prod}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/telemetry":
		h.receber(w, r)
	case r.URL.Path == "/health":
		w.WriteHeader(http.StatusOK)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) receber(w http.ResponseWriter, r *http.Request) {
	var req telemetryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if req.DeviceID == "" {
		writeErro(w, http.StatusBadRequest, "device_id é obrigatório")
		return
	}

	if err := h.prod.PublishTelemetry(r.Context(), req.DeviceID, req.Payload); err != nil {
		log.Printf("ERRO ao publicar telemetria device_id=%s: %v", req.DeviceID, err)
		writeErro(w, http.StatusInternalServerError, "falha ao publicar evento")
		return
	}

	log.Printf("Telemetria recebida device_id=%s bateria=%.0f%% temp=%.1f°C vel=%.1fkm/h",
		req.DeviceID, req.Payload.Battery*100, req.Payload.TemperatureC, req.Payload.SpeedKmh)

	w.WriteHeader(http.StatusAccepted)
}

func writeErro(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"erro": msg})
}
