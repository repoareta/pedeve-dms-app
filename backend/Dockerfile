# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git (needed for some Go dependencies)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Build from cmd/api directory (Clean Architecture structure)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Build seed-companies binary for production use
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o seed-companies ./cmd/seed-companies

# Stage 2: Production
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy seed-companies binary for seeder functionality
COPY --from=builder /app/seed-companies .
RUN chmod +x seed-companies

# Expose port (adjust based on your application)
EXPOSE 8080

# Run the binary
CMD ["./main"]

