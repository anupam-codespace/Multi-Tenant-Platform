# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the executable with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o /app/ecommerce-platform .

# Final minimal runtime image
FROM alpine:3.20

WORKDIR /app

# Install runtime certificates and timezones
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root unprivileged user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy compiled binary from builder
COPY --from=builder /app/ecommerce-platform /app/ecommerce-platform
COPY --from=builder /app/db_migrations /app/db_migrations
COPY --from=builder /app/.env.example /app/.env.dev

# Set permissions
RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/ecommerce-platform"]
CMD ["api"]
