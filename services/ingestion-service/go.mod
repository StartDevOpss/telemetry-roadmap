module github.com/telemetry-platform/ingestion-service

go 1.25

require (
	github.com/google/uuid v1.6.0
	github.com/telemetry-platform/events v0.0.0-00010101000000-000000000000
	github.com/twmb/franz-go v1.17.1
)

require (
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
)

replace github.com/telemetry-platform/events => ../../pkg/events
