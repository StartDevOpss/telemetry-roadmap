FROM golang:1.25-alpine AS builder
WORKDIR /workspace
COPY pkg/events ./pkg/events
COPY services/ingestion-service/go.mod ./services/ingestion-service/
WORKDIR /workspace/services/ingestion-service
ENV GOFLAGS=-mod=mod
RUN go mod download
COPY services/ingestion-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /server /server
EXPOSE 8081
ENTRYPOINT ["/server"]
