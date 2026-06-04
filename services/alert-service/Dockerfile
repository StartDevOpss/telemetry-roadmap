FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY pkg/events ./pkg/events
COPY pkg/telemetryobs ./pkg/telemetryobs
COPY services/alert-service/go.mod ./services/alert-service/
WORKDIR /workspace/services/alert-service
ENV GOFLAGS=-mod=mod
RUN go mod download
COPY services/alert-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
