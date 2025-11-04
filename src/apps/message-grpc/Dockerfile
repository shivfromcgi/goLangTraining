# Dockerfile for message-grpc service
# Single service per container using scratch base

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.work go.work.sum ./
COPY src/pkg/storage/go.mod src/pkg/storage/go.sum src/pkg/storage/
COPY src/pkg/types/go.mod src/pkg/types/go.sum src/pkg/types/
COPY src/apps/message-grpc/go.mod src/apps/message-grpc/go.sum src/apps/message-grpc/
COPY proto/message_service/go.mod proto/message_service/go.sum proto/message_service/

# Download dependencies (cached layer)
RUN go work sync && \
    cd src/apps/message-grpc && go mod download

# Copy source code
COPY src/ src/
COPY proto/ proto/

# Build the binary with CGO disabled for scratch compatibility
RUN cd src/apps/message-grpc && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o message-grpc .

# Production stage using scratch
FROM scratch

# Copy ca-certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/src/apps/message-grpc/message-grpc /message-grpc

# Expose port
EXPOSE 50051

# Run the binary
ENTRYPOINT ["/message-grpc"]