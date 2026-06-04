package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/telemetry-platform/events"
	"github.com/telemetry-platform/ingestion-service/internal/producer"
	"github.com/telemetry-platform/telemetryobs"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type telemetryRequest struct {
	DeviceID string                  `json:"device_id"`
	Payload  events.TelemetryPayload `json:"payload"`
}

type Handler struct {
	prod   *producer.EventProducer
	tracer trace.Tracer
}

func New(prod *producer.EventProducer, tracer trace.Tracer) *Handler {
	return &Handler{prod: prod, tracer: tracer}
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
	ctx, span := h.tracer.Start(r.Context(), "ingestion.receive_telemetry")
	defer span.End()

	inicio := time.Now()

	var req telemetryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		telemetryobs.EventsReceived.WithLabelValues("erro_parse").Inc()
		writeErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if req.DeviceID == "" {
		telemetryobs.EventsReceived.WithLabelValues("erro_validacao").Inc()
		writeErro(w, http.StatusBadRequest, "device_id é obrigatório")
		return
	}

	span.SetAttributes(
		attribute.String("device.id", req.DeviceID),
		attribute.Float64("device.battery", req.Payload.Battery),
		attribute.Float64("device.temperature_c", req.Payload.TemperatureC),
		attribute.Float64("device.speed_kmh", req.Payload.SpeedKmh),
	)

	if err := h.prod.PublishTelemetry(ctx, req.DeviceID, req.Payload); err != nil {
		telemetryobs.EventsReceived.WithLabelValues("erro_publish").Inc()
		log.Printf("ERRO ao publicar telemetria device_id=%s: %v", req.DeviceID, err)
		writeErro(w, http.StatusInternalServerError, "falha ao publicar evento")
		return
	}

	telemetryobs.EventsReceived.WithLabelValues("ok").Inc()
	telemetryobs.ProcessingDuration.WithLabelValues("ingestion-service").Observe(time.Since(inicio).Seconds())

	log.Printf("Telemetria recebida device_id=%s bateria=%.0f%% temp=%.1f°C vel=%.1fkm/h",
		req.DeviceID, req.Payload.Battery*100, req.Payload.TemperatureC, req.Payload.SpeedKmh)

	w.WriteHeader(http.StatusAccepted)
}

func writeErro(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"erro": msg})
}
