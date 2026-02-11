# Docker Compose for Local Development

This document explains the difference between `docker-compose.yml` and `docker-compose.dev.yml`.

## Overview

- **`docker-compose.yml`**: Production setup with nginx reverse proxy
- **`docker-compose.dev.yml`**: Local development setup without nginx

## Key Differences

### 1. Network Configuration

**Production (`docker-compose.yml`)**:
- Uses two networks: `id100-network` (internal) and `nginx-proxy` (external)
- Services like minio, meilisearch, and webapp are connected to both networks
- Assumes nginx reverse proxy is running externally

**Development (`docker-compose.dev.yml`)**:
- Uses only `id100-network` (internal bridge network)
- No external nginx-proxy dependency
- Direct access to all services

### 2. Webapp Port Configuration

**Production (`docker-compose.yml`)**:
```yaml
webapp:
  expose:
    - "8080"  # Only accessible within docker network
  networks:
    - id100-network
    - nginx-proxy  # Accessible via nginx
```

**Development (`docker-compose.dev.yml`)**:
```yaml
webapp:
  ports:
    - "8080:8080"  # Directly accessible on host
  networks:
    - id100-network  # Only internal network needed
```

### 3. Environment Variables

**Development** includes default values for all environment variables using the `${VAR:-default}` syntax:
- `POSTGRES_USER: ${POSTGRES_USER:-dev}`
- `MINIO_ROOT_USER: ${MINIO_ROOT_USER:-minioadmin}`
- etc.

This allows running without a `.env` file, though `.env.dev` is provided for convenience.

## Usage

### Local Development

```bash
# Start all services for local development
docker compose -f docker-compose.dev.yml --env-file .env.dev up -d

# Access the application directly
open http://localhost:8080
```

### Production

```bash
# Start services (requires nginx reverse proxy)
docker compose up -d

# Application accessible via nginx on configured domain
```

## Access Points (Development Mode)

- **Webapp**: http://localhost:8080
- **MinIO Console**: http://localhost:9001
- **MinIO API**: http://localhost:9000
- **PostgreSQL**: localhost:5432
- **Meilisearch**: http://localhost:8081
