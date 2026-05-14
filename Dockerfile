# ── Stage 1: Build ──────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and compile
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# ── Stage 2: Run ────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

# Install CA certificates (needed for HTTPS calls to Supabase/Resend)
RUN apk add --no-cache ca-certificates

# Copy the compiled binary from builder
COPY --from=builder /app/server .

# Render sets PORT automatically; default to 8080 locally
ENV PORT=8080
EXPOSE 8080

CMD ["./server"]
