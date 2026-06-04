package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/telemetry-platform/events"
	"github.com/telemetry-platform/telemetryobs"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Consumer struct {
	client *kgo.Client
	tracer trace.Tracer
}

func New(brokers []string, tracer trace.Tracer) (*Consumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup("dashboard-service-cg"),
		kgo.ConsumeTopics(
			events.TopicDeviceStateUpdated,
			events.TopicAlertTriggered,
		),
	)
	if err != nil {
		return nil, err
	}
	return &Consumer{client: cl, tracer: tracer}, nil
}

func (c *Consumer) Close() { c.client.Close() }

func (c *Consumer) Run(ctx context.Context) {
	log.Println("dashboard-service consumidor iniciado...")
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(t string, p int32, err error) {
			log.Printf("ERRO ao ler tópico=%s partição=%d: %v", t, p, err)
		})
		fetches.EachRecord(func(r *kgo.Record) {
			c.handle(r)
		})
	}
}

func (c *Consumer) handle(r *kgo.Record) {
	ctx := context.Background()
	_, span := c.tracer.Start(ctx, "dashboard.process_event")
	defer span.End()

	var env events.Envelope
	if err := json.Unmarshal(r.Value, &env); err != nil {
		log.Printf("ERRO ao deserializar tópico=%s: %v", r.Topic, err)
		return
	}

	span.SetAttributes(
		attribute.String("event.type", env.EventType),
		attribute.String("device.id", env.DeviceID),
	)

	switch env.EventType {
	case events.TypeDeviceStateUpdated:
		var p events.DeviceStatePayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			log.Printf("ERRO ao decodificar DeviceStateUpdated: %v", err)
			return
		}
		telemetryobs.ProcessingDuration.WithLabelValues("dashboard-service").Observe(0)
		log.Printf("[DASHBOARD] ESTADO ATUALIZADO device_id=%s bateria=%.0f%% temp=%.1f°C vel=%.1fkm/h pos=(%.4f,%.4f)",
			env.DeviceID, p.LastBattery*100, p.LastTemperatureC, p.LastSpeedKmh, p.LastLat, p.LastLon)

	case events.TypeAlertTriggered:
		var p events.AlertPayload
		if err := json.Unmarshal(env.Payload, &p); err != nil {
			log.Printf("ERRO ao decodificar AlertTriggered: %v", err)
			return
		}
		log.Printf("[DASHBOARD] ⚠ ALERTA [%s/%s] device_id=%s valor=%.2f limite=%.2f",
			p.Rule, p.Severity, env.DeviceID, p.Value, p.Limit)

	default:
		log.Printf("[DASHBOARD] EVENTO DESCONHECIDO tipo=%s device_id=%s", env.EventType, env.DeviceID)
	}
}
