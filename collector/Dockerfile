FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/collector ./cmd/collector

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/collector /app/collector
RUN adduser -D appuser && mkdir -p /app/data && chown -R appuser /app
USER appuser
CMD ["/app/collector", "-out", "/app/data", "-serve", "-addr", ":8080"]
