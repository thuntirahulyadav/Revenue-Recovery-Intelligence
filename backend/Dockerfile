 FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build tools
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

# Minimal Runtime Image
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata wget

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8082

CMD ["/app/api"]
