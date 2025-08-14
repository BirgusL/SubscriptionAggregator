FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /subscription-aggregator ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /subscription-aggregator .
COPY config/docker.yaml ./config/docker.yaml
COPY .env .
COPY --from=builder /app/migrations ./migrations

CMD ["./subscription-aggregator"]