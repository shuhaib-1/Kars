# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Final stage
FROM alpine:latest

# Install necessary dependencies
RUN apk --no-cache add postgresql-client redis

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/.env .

# Expose port
EXPOSE 3000


# Command to run the application
CMD ["./main"]