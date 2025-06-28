# Dockerfile Optimization Rules (Dockerfile\* files ONLY)

## Multi-Stage Build Optimization (Mandatory)

### Zero-Waste Image Pattern

```dockerfile
# ✅ ALWAYS: Multi-stage builds for minimal final image
FROM golang:1.21-alpine AS builder

# Build stage optimizations
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Production stage - minimal image
FROM alpine:3.18
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Distroless Pattern for Maximum Security

```dockerfile
# ✅ ALWAYS: Use distroless for production Go binaries
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

# Distroless final stage
FROM gcr.io/distroless/static-debian11
COPY --from=builder /app/main /
ENTRYPOINT ["/main"]
```

## Layer Optimization (Memory → Disk → CPU)

### Combine RUN Commands (Disk Optimization)

```dockerfile
# ✅ ALWAYS: Combine RUN commands to reduce layers
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        package1 \
        package2 \
        package3 && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    rm -rf /tmp/* && \
    rm -rf /var/tmp/*
```

### Strategic Layer Ordering (Cache Optimization)

```dockerfile
# ✅ ALWAYS: Order by change frequency (least → most frequent)
FROM golang:1.21-alpine

# 1. System dependencies (change rarely)
RUN apk add --no-cache git ca-certificates tzdata

# 2. Go dependencies (change occasionally)
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# 3. Source code (changes frequently) - last
COPY . .
RUN go build -o main .
```

## .dockerignore Optimization

### Essential Exclusions

```dockerfile
# ✅ ALWAYS: Create comprehensive .dockerignore
# Version control
.git
.gitignore
.gitattributes

# Documentation
*.md
README*
CHANGELOG*
LICENSE*

# Development files
.env*
.vscode/
.idea/
*.log
tmp/
temp/

# Build artifacts
target/
dist/
build/
*.exe
*.dll
*.so

# Test files
*_test.go
testdata/
coverage.out

# Docker files
Dockerfile*
docker-compose*
.dockerignore
```

## Security Hardening

### Non-Root User Pattern

```dockerfile
# ✅ ALWAYS: Create and use non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of working directory
WORKDIR /app
COPY --chown=appuser:appgroup . .

# Switch to non-root user
USER appuser

# Verify user
RUN id && whoami
```

### Specific Version Pinning

```dockerfile
# ✅ ALWAYS: Pin exact versions for security
FROM golang:1.21.5-alpine3.18

# Pin package versions
RUN apk add --no-cache \
    ca-certificates=20230506-r0 \
    tzdata=2023c-r1
```

## Health Check Implementation

### Comprehensive Health Monitoring

```dockerfile
# ✅ ALWAYS: Implement health checks
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# For Go applications without curl
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./main -health-check || exit 1
```

## Build Arguments and Environment

### Flexible Build Configuration

```dockerfile
# ✅ ALWAYS: Use build args for flexibility
ARG GO_VERSION=1.21
ARG ALPINE_VERSION=3.18
ARG BUILD_DATE
ARG VERSION
ARG COMMIT_SHA

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Build-time labels
LABEL org.opencontainers.image.title="My Application"
LABEL org.opencontainers.image.description="High-performance Go application"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
LABEL org.opencontainers.image.source="https://github.com/user/repo"
```

## Resource Optimization

### Memory-Efficient Base Images

```dockerfile
# ✅ ALWAYS: Choose minimal base images
# Alpine for small size (5MB)
FROM alpine:3.18

# Distroless for security (static binary)
FROM gcr.io/distroless/static-debian11

# Scratch for minimal possible size (static binary only)
FROM scratch
```

### CPU-Optimized Builds

```dockerfile
# ✅ ALWAYS: Optimize Go builds for production
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -X main.version=${VERSION} -X main.commit=${COMMIT_SHA}" \
    -o main .
```

## Volume and Data Management

### Proper Volume Handling

```dockerfile
# ✅ ALWAYS: Declare volumes for persistent data
VOLUME ["/data", "/logs"]

# Create directories with proper permissions
RUN mkdir -p /data /logs && \
    chown -R appuser:appgroup /data /logs && \
    chmod 755 /data /logs
```

## Network Configuration

### Port and Protocol Declaration

```dockerfile
# ✅ ALWAYS: Document exposed ports
EXPOSE 8080/tcp
EXPOSE 8081/tcp

# Document port purpose in comments
# 8080: HTTP API
# 8081: Metrics endpoint
```

## Build Cache Optimization

### Dependency Caching Strategy

```dockerfile
# ✅ ALWAYS: Leverage build cache for dependencies
FROM golang:1.21-alpine AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

FROM deps AS builder
COPY . .
RUN go build -o main .
```

## Production Optimization

### Runtime Environment

```dockerfile
# ✅ ALWAYS: Set production environment variables
ENV GO_ENV=production
ENV GIN_MODE=release
ENV CGO_ENABLED=0

# Timezone configuration
ENV TZ=UTC
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
```

## Docker Compose Integration

### Service Definition Best Practices

```yaml
# docker-compose.yml
version: "3.8"

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        - BUILD_DATE=${BUILD_DATE}
        - VERSION=${VERSION}
        - COMMIT_SHA=${COMMIT_SHA}

    # ✅ ALWAYS: Resource limits
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
        reservations:
          cpus: "0.25"
          memory: 256M

    # ✅ ALWAYS: Restart policy
    restart: unless-stopped

    # ✅ ALWAYS: Health check
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

    # ✅ ALWAYS: Logging configuration
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

    # Environment variables
    environment:
      - GO_ENV=production
      - LOG_LEVEL=info

    # Volume mounts
    volumes:
      - app_data:/data
      - app_logs:/logs

volumes:
  app_data:
  app_logs:
```

## Security Scanning Integration

### Vulnerability Assessment

```dockerfile
# ✅ ALWAYS: Include security scan comments
# Security scanning commands:
# docker scout cve <image>
# docker scout recommendations <image>
# trivy image <image>

# Build with security context
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder
```

## Metadata and Labels

### Comprehensive Image Labeling

```dockerfile
# ✅ ALWAYS: Complete OCI labels
LABEL maintainer="team@company.com"
LABEL org.opencontainers.image.title="Application Name"
LABEL org.opencontainers.image.description="Application description"
LABEL org.opencontainers.image.url="https://company.com"
LABEL org.opencontainers.image.source="https://github.com/company/repo"
LABEL org.opencontainers.image.documentation="https://docs.company.com"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
LABEL org.opencontainers.image.vendor="Company Name"
LABEL org.opencontainers.image.licenses="MIT"
```

## Build Validation Commands

### Quality Assurance

```bash
# ✅ ALWAYS: Validate Dockerfile
hadolint Dockerfile

# ✅ ALWAYS: Security scanning
docker scout cve <image>
trivy image <image>

# ✅ ALWAYS: Size optimization check
docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

# ✅ ALWAYS: Layer analysis
docker history <image>
```

## Error Handling Pattern

```dockerfile
# ✅ ALWAYS: Handle command failures
RUN set -e && \
    command1 && \
    command2 && \
    command3

# ✅ ALWAYS: Verify critical operations
RUN go mod download && \
    go mod verify && \
    test -f go.sum
```

## Documentation Requirements

### Inline Documentation

```dockerfile
# Application: My Go Application
# Description: High-performance microservice
# Version: 1.0.0
# Build: docker build -t myapp:latest .
# Run: docker run -p 8080:8080 myapp:latest

# Build stage
FROM golang:1.21-alpine AS builder
# ... rest of Dockerfile
```
