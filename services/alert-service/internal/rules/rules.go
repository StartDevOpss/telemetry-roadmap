package rules

import "github.com/telemetry-platform/events"

type Violation struct {
	Rule     string
	Severity string
	Value    float64
	Limit    float64
}

// Evaluate aplica todas as regras e retorna as violações encontradas.
func Evaluate(p events.TelemetryPayload) []Violation {
	var v []Violation
	if p.Battery < 0.20 {
		v = append(v, Violation{"battery_low", "critical", p.Battery, 0.20})
	}
	if p.TemperatureC > 40.0 {
		v = append(v, Violation{"high_temperature", "warning", p.TemperatureC, 40.0})
	}
	if p.SpeedKmh > 80.0 {
		v = append(v, Violation{"speeding", "warning", p.SpeedKmh, 80.0})
	}
	return v
}
