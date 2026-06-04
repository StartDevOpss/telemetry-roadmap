// Package telemetryobs configura Prometheus e OpenTelemetry para todos os serviços.
package telemetryobs

import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Métricas compartilhadas — registradas uma vez, usadas por todos os serviços.
var (
	EventsReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "telemetry_events_received_total",
		Help: "Total de eventos de telemetria recebidos pelo ingestion-service.",
	}, []string{"status"})

	AlertsTriggered = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "telemetry_alerts_triggered_total",
		Help: "Total de alertas disparados pelo alert-service.",
	}, []string{"rule", "severity"})

	ProcessingDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "telemetry_processing_duration_seconds",
		Help:    "Latência de processamento de evento por serviço.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service"})

	ActiveDevices = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "telemetry_active_devices",
		Help: "Número de dispositivos únicos com estado atualizado.",
	})
)

func init() {
	prometheus.MustRegister(EventsReceived, AlertsTriggered, ProcessingDuration, ActiveDevices)
}

// ServeMetrics expõe /metrics no addr fornecido (ex: ":9090").
func ServeMetrics(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("metrics server: %v", err)
		}
	}()
}

// InitTracer configura o tracer OTel e envia spans para o OTel Collector via gRPC.
// otlpEndpoint ex: "otel-collector:4317"
func InitTracer(ctx context.Context, serviceName, otlpEndpoint string) (trace.Tracer, func(), error) {
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, nil, err
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	shutdown := func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("tracer shutdown: %v", err)
		}
	}

	return tp.Tracer(serviceName), shutdown, nil
}
