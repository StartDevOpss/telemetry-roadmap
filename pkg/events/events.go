package events

import "encoding/json"

// Tópicos — fonte única da verdade para todos os serviços.
const (
	TopicTelemetryReceived  = "telemetry.received"
	TopicDeviceStateUpdated = "device.state.updated"
	TopicAlertTriggered     = "alert.triggered"
)

// Tipos de evento.
const (
	TypeTelemetryReceived  = "telemetry.received"
	TypeDeviceStateUpdated = "device.state.updated"
	TypeAlertTriggered     = "alert.triggered"
)

// Envelope é o wrapper comum de todos os eventos no barramento.
// Chave de partição do Kafka = DeviceID (garante ordem por dispositivo).
type Envelope struct {
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt string          `json:"occurred_at"` // RFC3339 UTC
	DeviceID   string          `json:"device_id"`
	Payload    json.RawMessage `json:"payload"`
}

// ── Payloads ──────────────────────────────────────────────────────────────────

// TelemetryPayload é o dado bruto enviado pelo dispositivo.
type TelemetryPayload struct {
	Lat           float64 `json:"lat"`
	Lon           float64 `json:"lon"`
	Battery       float64 `json:"battery"`        // 0.0 – 1.0
	TemperatureC  float64 `json:"temperature_c"`
	SpeedKmh      float64 `json:"speed_kmh"`
}

// DeviceStatePayload representa o estado persistido do dispositivo.
type DeviceStatePayload struct {
	LastLat          float64 `json:"last_lat"`
	LastLon          float64 `json:"last_lon"`
	LastBattery      float64 `json:"last_battery"`
	LastTemperatureC float64 `json:"last_temperature_c"`
	LastSpeedKmh     float64 `json:"last_speed_kmh"`
	UpdatedAt        string  `json:"updated_at"`
}

// AlertPayload descreve o alerta disparado por uma regra.
type AlertPayload struct {
	Rule     string  `json:"rule"`     // ex: "battery_low", "high_temperature", "speeding"
	Severity string  `json:"severity"` // "warning" | "critical"
	Value    float64 `json:"value"`    // valor que violou o limite
	Limit    float64 `json:"limit"`    // limite configurado
}
