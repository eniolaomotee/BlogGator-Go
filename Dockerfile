# -------------------------
# Build stage
# -------------------------
# Use the official Go Alpine image to build the binary
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Install git (needed for go modules that come from VCS)
RUN apk add --no-cache git

# Copy only go.mod and go.sum first
# This allows Docker to cache dependency downloads
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build a statically-linked Linux binary
# - CGO_ENABLED=0: avoids libc dependencies
# - GOOS=linux: ensures Linux compatibility
# - Output binary named "gator"
RUN CGO_ENABLED=0 GOOS=linux go build -o gator ./cmd/gator


# -------------------------
# Runtime stage
# -------------------------
# Use a minimal Alpine image for runtime
FROM alpine:3.18

# Install CA certificates (required for HTTPS calls)
# curl is optional but useful for debugging/logging
RUN apk add --no-cache ca-certificates curl

# Set the working directory for the runtime container
WORKDIR /app

# Copy the compiled binary from the build stage
COPY --from=builder /app/gator .

# Copy SQL files needed at runtime
COPY sql ./sql

# Create a non-root user and group for better security
RUN addgroup -S app && adduser -S -G app app

# Ensure the binary is executable
RUN chmod +x /app/gator

# Default port (Cloud Run injects PORT automatically)
ENV PORT=8080

# Informational only — Cloud Run ignores EXPOSE
EXPOSE 8080

# Switch to the non-root user
USER app

# Start the application
# Shell form is required so ${PORT} is expanded at runtime
CMD ["sh", "-c", "./gator serve ${PORT}"]
