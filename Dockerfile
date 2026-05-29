# Build stage for frontend
FROM node:24-alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy TypeScript source files
COPY tsconfig.json ./
COPY vitest.config.ts ./
COPY postcss.config.js ./
COPY src ./src
COPY scripts ./scripts
COPY web/static/style.css ./web/static/style.css
COPY web/static/admin.styles.css ./web/static/admin.styles.css

# Build frontend
RUN npm run build

# Build stage for Go backend
FROM golang:1.26-alpine AS backend-builder

# Version: APP_VERSION overrides; otherwise resolved from the latest GitHub
# release at build time.
ARG APP_VERSION=dev
ARG GITHUB_REPO=dmnktoe/id-100
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git build-base libwebp-dev curl ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd ./cmd
COPY internal ./internal

# Build the application with CGO enabled
RUN set -e; \
    if [ -z "${APP_VERSION}" ] || [ "${APP_VERSION}" = "dev" ]; then \
      RESP="$(curl -fsSL -H 'Accept: application/vnd.github+json' "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" || true)"; \
      APP_VERSION="$(printf '%s' "$RESP" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/' || true)"; \
    fi; \
    [ -z "${APP_VERSION}" ] && APP_VERSION=dev; \
    echo "Building id-100 version: ${APP_VERSION}"; \
    CGO_ENABLED=1 GOOS=linux go build -ldflags "-X 'id-100/internal/version.Version=${APP_VERSION}'" -o /app/bin/id-100 ./cmd/id-100

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (wget für healthcheck im compose)
RUN apk add --no-cache ca-certificates libwebp wget

# Copy the binary from builder
COPY --from=backend-builder /app/bin/id-100 /app/id-100

# Copy web templates and static files
COPY web /app/web

# Copy built frontend from frontend-builder
COPY --from=frontend-builder /app/web/static/dist /app/web/static/dist
COPY --from=frontend-builder /app/web/static/manifest.json /app/web/static/manifest.json

# Copy startup script
COPY scripts/startup.sh /app/scripts/startup.sh
RUN chmod +x /app/scripts/startup.sh

# Expose port
EXPOSE 8080

# Use startup script as entrypoint
ENTRYPOINT ["/app/scripts/startup.sh"]
CMD ["/app/id-100"]
