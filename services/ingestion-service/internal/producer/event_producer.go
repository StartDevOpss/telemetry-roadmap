package producer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/telemetry-platform/events"
	"github.com/twmb/franz-go/pkg/kgo"
)

type EventProducer struct {
	client *kgo.Client
}

func New(brokers []string) (*EventProducer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, err
	}
	return &EventProducer{client: cl}, nil
}

func (p *EventProducer) Close() { p.client.Close() }

func (p *EventProducer) PublishTelemetry(ctx context.Context, deviceID string, payload events.TelemetryPayload) error {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	env := events.Envelope{
		EventID:    uuid.New().String(),
		EventType:  events.TypeTelemetryReceived,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		DeviceID:   deviceID,
		Payload:    rawPayload,
	}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return p.client.ProduceSync(ctx, &kgo.Record{
		Topic: events.TopicTelemetryReceived,
		Key:   []byte(deviceID),
		Value: data,
	}).FirstErr()
}
