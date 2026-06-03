package consumer

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/telemetry-platform/alert-service/internal/rules"
	"github.com/telemetry-platform/events"
	"github.com/twmb/franz-go/pkg/kgo"
)

// seen rastreia event_ids processados em memória.
// TODO Fase 2+: substituir por tabela PostgreSQL para resiliência a restarts.
var seen sync.Map

type Consumer struct {
	client *kgo.Client
	prod   *kgo.Client
}

func New(brokers []string) (*Consumer, error) {
	reader, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup("alert-service-cg"),
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
	return &Consumer{client: reader, prod: writer}, nil
}

func (c *Consumer) Close() {
	c.client.Close()
	c.prod.Close()
}

func (c *Consumer) Run(ctx context.Context) {
	log.Println("alert-service consumidor iniciado...")
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
	var env events.Envelope
	if err := json.Unmarshal(r.Value, &env); err != nil {
		return err
	}

	if _, loaded := seen.LoadOrStore(env.EventID, struct{}{}); loaded {
		log.Printf("IGNORANDO evento duplicado event_id=%s", env.EventID)
		return nil
	}

	var p events.TelemetryPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return err
	}

	violations := rules.Evaluate(p)
	if len(violations) == 0 {
		return nil
	}

	for _, v := range violations {
		log.Printf("ALERTA [%s/%s] device_id=%s valor=%.2f limite=%.2f",
			v.Rule, v.Severity, env.DeviceID, v.Value, v.Limit)
		if err := c.publishAlert(ctx, env.DeviceID, v); err != nil {
			log.Printf("ERRO ao publicar alerta %s device_id=%s: %v", v.Rule, env.DeviceID, err)
		}
	}
	return nil
}

func (c *Consumer) publishAlert(ctx context.Context, deviceID string, v rules.Violation) error {
	alertPayload := events.AlertPayload{
		Rule:     v.Rule,
		Severity: v.Severity,
		Value:    v.Value,
		Limit:    v.Limit,
	}
	rawPayload, err := json.Marshal(alertPayload)
	if err != nil {
		return err
	}
	env := events.Envelope{
		EventID:    uuid.New().String(),
		EventType:  events.TypeAlertTriggered,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		DeviceID:   deviceID,
		Payload:    rawPayload,
	}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return c.prod.ProduceSync(ctx, &kgo.Record{
		Topic: events.TopicAlertTriggered,
		Key:   []byte(deviceID),
		Value: data,
	}).FirstErr()
}
