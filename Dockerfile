# -------- Stage 1: Build --------
FROM golang:1.22-alpine AS builder

# Install git (needed for go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files first (cache optimization)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app .

# -------- Stage 2: Run --------
FROM alpine:3.20

# Create non-root user (security best practice)
RUN adduser -D appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/app .COPY --from=builder /app/templates ./templates
#COPY --from=builder /app/static ./static


# Expose application port
EXPOSE 8080

# Use non-root user
USER appuser

# Run the application
CMD ["./app"]
