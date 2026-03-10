# Build stage
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

# Migrations stage
FROM migrate/migrate:latest AS migrations
COPY migrations /migrations
ENTRYPOINT ["migrate", "-path", "/migrations", "-database"]
CMD ["postgres://postgres:postgres@postgres:5432/voice_ai?sslmode=disable", "up"]

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/server .
COPY .env.example .env
EXPOSE 8080
CMD ["./server"]
