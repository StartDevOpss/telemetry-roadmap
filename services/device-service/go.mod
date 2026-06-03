module github.com/telemetry-platform/device-service

go 1.25

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/redis/go-redis/v9 v9.7.3
	github.com/telemetry-platform/events v0.0.0-00010101000000-000000000000
	github.com/twmb/franz-go v1.17.1
)

replace github.com/telemetry-platform/events => ../../pkg/events
