# ---------------------------------------------------------
# Stage 1: Builder
# ---------------------------------------------------------
FROM golang:1.25-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Copy dependency files first (for better caching)
COPY go.mod go.sum ./

# Download modules
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
# CGO_ENABLED=0: Disables C bindings (ensures a static binary)
# GOOS=linux: Compiles explicitly for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/main ./cmd/api/

# ---------------------------------------------------------
# Stage 2: Runner
# ---------------------------------------------------------
FROM alpine:latest

# Install certificates (useful if your app makes HTTPS calls to external APIs)
# RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/bin/main .

# Expose the port your app runs on (e.g., 8080 or 3000)
EXPOSE 80

# Command to run the executable
CMD ["./main"]