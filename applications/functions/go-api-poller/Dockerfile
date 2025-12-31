FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY handler.go .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-api-poller handler.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /go-api-poller .

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["./go-api-poller"]
