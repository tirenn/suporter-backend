# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Cache go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build statically linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o server ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.20

WORKDIR /app

# Install runtime dependencies (certificates and timezone data)
RUN apk add --no-cache ca-certificates tzdata

# Copy binary, static assets, and SQL migrations from builder
COPY --from=builder /app/server .
COPY static/ ./static/
COPY migrations/ ./migrations/

# Environment defaults
ENV PORT=8080
EXPOSE 8080

CMD ["./server"]
