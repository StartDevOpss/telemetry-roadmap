FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY pkg/events ./pkg/events
COPY pkg/telemetryobs ./pkg/telemetryobs
COPY services/dashboard-service/go.mod ./services/dashboard-service/
WORKDIR /workspace/services/dashboard-service
ENV GOFLAGS=-mod=mod
RUN go mod download
COPY services/dashboard-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
