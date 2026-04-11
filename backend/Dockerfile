# ---- Stage 1: Build ----
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# ---- Stage 2: Run ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy compiled binary and required assets
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/seed.sql .

EXPOSE 4000

CMD ["./server"]
