# Build stage for frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy TypeScript source files
COPY tsconfig.json ./
COPY vitest.config.ts ./
COPY src ./src

# Build frontend
RUN npm run build

# Build stage for Go backend
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd ./cmd
COPY internal ./internal

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/id-100 ./cmd/id-100

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=backend-builder /app/bin/id-100 /app/id-100

# Copy web templates and static files
COPY web /app/web

# Copy built frontend from frontend-builder
COPY --from=frontend-builder /app/web/static/main.js /app/web/static/main.js
COPY --from=frontend-builder /app/web/static/main.js.map /app/web/static/main.js.map

# Expose port
EXPOSE 8080

# Run the application
CMD ["/app/id-100"]
