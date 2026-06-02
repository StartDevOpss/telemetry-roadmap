FROM golang:1.23-alpine AS builder
WORKDIR /workspace
COPY pkg/events ./pkg/events
COPY services/device-service/go.mod ./services/device-service/
WORKDIR /workspace/services/device-service
ENV GOFLAGS=-mod=mod
RUN go mod download
COPY services/device-service ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
