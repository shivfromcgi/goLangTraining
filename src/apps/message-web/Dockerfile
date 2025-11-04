# Dockerfile for message-web service
# Single service per container using scratch base

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.work go.work.sum ./
COPY src/pkg/storage/go.mod src/pkg/storage/go.sum src/pkg/storage/
COPY src/pkg/types/go.mod src/pkg/types/go.sum src/pkg/types/
COPY src/apps/message-web/go.mod src/apps/message-web/go.sum src/apps/message-web/

# Download dependencies (cached layer)
RUN go work sync && \
    cd src/apps/message-web && go mod download

# Copy source code
COPY src/ src/

# Build the binary with CGO disabled for scratch compatibility
RUN cd src/apps/message-web && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o message-web .

# Production stage using scratch
FROM scratch

# Copy ca-certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/src/apps/message-web/message-web /message-web

# Expose port
EXPOSE 8090

# Run the binary
ENTRYPOINT ["/message-web"]