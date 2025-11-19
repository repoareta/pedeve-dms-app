# Stage 1: Build
FROM golang:1.24-alpine AS builder

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
# Adjust the build command based on your main.go location
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 2: Production
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port (adjust based on your application)
EXPOSE 8080

# Run the binary
CMD ["./main"]

