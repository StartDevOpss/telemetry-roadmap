package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/telemetry-platform/device-service/internal/repository"
	"github.com/telemetry-platform/events"
	"github.com/telemetry-platform/telemetryobs"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Consumer struct {
	repo   *repository.DeviceRepository
	rdb    *redis.Client
	client *kgo.Client
	prod   *kgo.Client
	tracer trace.Tracer
}

func New(brokers []string, repo *repository.DeviceRepository, rdb *redis.Client, tracer trace.Tracer) (*Consumer, error) {
	reader, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup("device-service-cg"),
		kgo.ConsumeTopics(events.TopicTelemetryReceived),
	)
	if err != nil {
		return nil, err
	}
	writer, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		reader.Close()
		return nil, err
	}
	return &Consumer{repo: repo, rdb: rdb, client: reader, prod: writer, tracer: tracer}, nil
}

func (c *Consumer) Close() {
	c.client.Close()
	c.prod.Close()
}

func (c *Consumer) Run(ctx context.Context) {
	log.Println("device-service consumidor iniciado...")
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(t string, p int32, err error) {
			log.Printf("ERRO ao ler tópico=%s partição=%d: %v", t, p, err)
		})
		fetches.EachRecord(func(r *kgo.Record) {
			if err := c.handle(ctx, r); err != nil {
				log.Printf("ERRO ao processar evento offset=%d: %v", r.Offset, err)
			}
		})
	}
}

func (c *Consumer) handle(ctx context.Context, r *kgo.Record) error {
	ctx, span := c.tracer.Start(ctx, "device.update_state")
	defer span.End()

	inicio := time.Now()

	var env events.Envelope
	if err := json.Unmarshal(r.Value, &env); err != nil {
		return err
	}

	// Idempotência via PostgreSQL
	dup, err := c.repo.MarkProcessed(ctx, env.EventID)
	if err != nil {
		return err
	}
	if dup {
		log.Printf("IGNORANDO evento duplicado event_id=%s", env.EventID)
		return nil
	}

	var p events.TelemetryPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return err
	}

	span.SetAttributes(
		attribute.String("device.id", env.DeviceID),
		attribute.Float64("device.battery", p.Battery),
	)

	log.Printf("PROCESSANDO telemetria device_id=%s bateria=%.0f%% temp=%.1f°C vel=%.1fkm/h",
		env.DeviceID, p.Battery*100, p.TemperatureC, p.SpeedKmh)

	// Persiste no PostgreSQL
	if err := c.repo.UpsertDevice(ctx, env.DeviceID, p.Lat, p.Lon, p.Battery, p.TemperatureC, p.SpeedKmh); err != nil {
		return fmt.Errorf("upsert device: %w", err)
	}

	// Atualiza estado quente no Redis
	stateJSON, _ := json.Marshal(map[string]any{
		"lat":           p.Lat,
		"lon":           p.Lon,
		"battery":       p.Battery,
		"temperature_c": p.TemperatureC,
		"speed_kmh":     p.SpeedKmh,
		"updated_at":    time.Now().UTC().Format(time.RFC3339),
	})
	if err := c.rdb.Set(ctx, "device:"+env.DeviceID, stateJSON, 0).Err(); err != nil {
		log.Printf("AVISO Redis device_id=%s: %v", env.DeviceID, err)
	}

	if err := c.publishStateUpdated(ctx, env.DeviceID, p); err != nil {
		return err
	}
	telemetryobs.ProcessingDuration.WithLabelValues("device-service").Observe(time.Since(inicio).Seconds())
	return nil
}

func (c *Consumer) publishStateUpdated(ctx context.Context, deviceID string, p events.TelemetryPayload) error {
	now := time.Now().UTC().Format(time.RFC3339)
	statePayload := events.DeviceStatePayload{
		LastLat:          p.Lat,
		LastLon:          p.Lon,
		LastBattery:      p.Battery,
		LastTemperatureC: p.TemperatureC,
		LastSpeedKmh:     p.SpeedKmh,
		UpdatedAt:        now,
	}
	rawPayload, err := json.Marshal(statePayload)
	if err != nil {
		return err
	}
	env := events.Envelope{
		EventID:    uuid.New().String(),
		EventType:  events.TypeDeviceStateUpdated,
		OccurredAt: now,
		DeviceID:   deviceID,
		Payload:    rawPayload,
	}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if err := c.prod.ProduceSync(ctx, &kgo.Record{
		Topic: events.TopicDeviceStateUpdated,
		Key:   []byte(deviceID),
		Value: data,
	}).FirstErr(); err != nil {
		return err
	}
	telemetryobs.ActiveDevices.Inc()
	log.Printf("Evento DeviceStateUpdated publicado device_id=%s", deviceID)
	return nil
}
