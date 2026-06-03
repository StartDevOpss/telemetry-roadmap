module github.com/telemetry-platform/alert-service

go 1.24

require (
	github.com/google/uuid v1.6.0
	github.com/telemetry-platform/events v0.0.0-00010101000000-000000000000
	github.com/twmb/franz-go v1.17.1
)

replace github.com/telemetry-platform/events => ../../pkg/events
