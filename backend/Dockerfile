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
CMD ["up"]

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/server .
EXPOSE 8082
CMD ["./server"]
