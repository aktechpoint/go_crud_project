# ---------- Build Stage ----------
FROM golang:1.24.5-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o app .

# ---------- Runtime Stage ----------
FROM alpine:3.20

RUN adduser -D appuser

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/templates ./templates

# Create uploads directory
RUN mkdir -p /app/uploads && \
    chown -R appuser:appuser /app/uploads

# If uploads folder exists in source and contains default images
# COPY --from=builder /app/uploads ./uploads

EXPOSE 8080

USER appuser

CMD ["./app"]
