FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o quote-service ./cmd/quotes/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/quote-service .

COPY config.docker.yaml ./config.yaml

EXPOSE 8080

CMD ["./quote-service"]
