# Multi-stage Docker build for Go Training Services
# Following coding guidelines for standardized build

# Stage 1: Build environment
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go modules files
COPY go.work go.work.sum ./
COPY src/ src/
COPY proto/ proto/

# Build services
RUN cd src/apps/message-cli && go mod tidy && go build -o /app/build/message-cli .
RUN cd src/apps/message-api && go mod tidy && go build -o /app/build/message-api .
RUN cd src/apps/message-grpc && go mod tidy && go build -o /app/build/message-grpc .
RUN cd src/apps/message-web && go mod tidy && go build -o /app/build/message-web .
RUN cd src/apps/message-grpc/cmd/client && go mod tidy && go build -o /app/build/grpc-client .

# Stage 2: Production image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/build/ ./

# Create non-root user following security best practices
RUN adduser -D -s /bin/sh appuser
USER appuser

# Default to API service
EXPOSE 8080

# Default command
CMD ["./message-api"]