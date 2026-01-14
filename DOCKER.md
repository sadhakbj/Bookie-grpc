# Docker Deployment Guide

This guide covers how to build, run, and deploy the Bookie gRPC application using Docker.

## Table of Contents
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Building Images](#building-images)
- [Running with Docker Compose](#running-with-docker-compose)
- [Production Deployment](#production-deployment)
- [Security Considerations](#security-considerations)
- [Monitoring & Health Checks](#monitoring--health-checks)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites
- Docker 20.10+ and Docker Compose v2.0+
- Make (optional, for convenience commands)

### Development Mode

```bash
# Build and start all services
docker-compose up --build

# Or use make commands
make docker-build
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

Access the services:
- HTTP API (BFF): http://localhost:8080
- gRPC Server: localhost:8020

### Test the API

```bash
# Get all books
curl http://localhost:8080/books

# Get book by ID
curl http://localhost:8080/books/1234
```

## Architecture

The application consists of two services:

1. **gRPC Server** (Port 8020)
   - Backend service managing books
   - Implements Bookie gRPC service
   - Stateless design

2. **HTTP Client/BFF** (Port 8080)
   - Backend-for-Frontend pattern
   - Exposes REST API
   - Communicates with gRPC server
   - Translates HTTP requests to gRPC calls

```
┌─────────┐      HTTP      ┌──────────────┐     gRPC      ┌─────────────┐
│ Browser │ ◄─────────────► │ HTTP Client  │ ◄────────────► │ gRPC Server │
│         │   :8080         │    (BFF)     │    :8020      │   (Bookie)  │
└─────────┘                 └──────────────┘               └─────────────┘
```

## Building Images

### Build Both Services

```bash
make docker-build
```

Or manually:

```bash
docker build -t bookie-grpc-server:latest -f Dockerfile.server .
docker build -t bookie-http-client:latest -f Dockerfile.client .
```

### Build Individual Services

```bash
# Server only
make docker-build-server

# Client only
make docker-build-client
```

### Image Details

Both images use **multi-stage builds** for optimal security and size:

- **Build Stage**: golang:1.24-alpine
  - Downloads dependencies
  - Compiles Go binary with optimizations

- **Runtime Stage**: gcr.io/distroless/static:nonroot
  - Minimal attack surface (no shell, package manager, etc.)
  - Non-root user (UID 65532)
  - Only contains the application binary and essential files
  - Typical image size: ~15-20MB

## Running with Docker Compose

### Development Environment

```bash
# Start in foreground (see logs)
docker-compose up

# Start in background (detached)
docker-compose up -d

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f grpc-server
docker-compose logs -f http-client

# Restart services
docker-compose restart

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Environment Variables

Create a `.env` file (see `config.example`):

```env
# gRPC Server
PORT=8020

# HTTP Client
HTTP_PORT=8080
GRPC_SERVER_ADDR=grpc-server:8020

# Timezone
TZ=UTC
```

## Production Deployment

### Using Production Compose

```bash
# Build images with version tag
VERSION=1.0.0 make docker-build

# Tag for registry
docker tag bookie-grpc-server:latest your-registry/bookie-grpc-server:1.0.0
docker tag bookie-http-client:latest your-registry/bookie-http-client:1.0.0

# Push to registry
docker push your-registry/bookie-grpc-server:1.0.0
docker push your-registry/bookie-http-client:1.0.0

# Deploy with production config
VERSION=1.0.0 docker-compose -f docker-compose.prod.yml up -d
```

### Production Features

The `docker-compose.prod.yml` includes:

1. **Resource Limits**
   - CPU and memory constraints
   - Prevents resource exhaustion

2. **Restart Policies**
   - Automatic recovery from failures
   - Backoff and retry limits

3. **Logging Configuration**
   - Size-limited log files
   - Rotation policies
   - Structured labels

4. **Security Hardening**
   - Read-only root filesystem
   - Dropped capabilities
   - No new privileges

## Security Considerations

### Image Security

1. **Distroless Base Images**
   - Minimal attack surface
   - No shell access
   - Reduced vulnerability exposure

2. **Non-Root User**
   - Runs as UID 65532 (nonroot)
   - Prevents privilege escalation

3. **Static Binary**
   - No external dependencies
   - No dynamic linking vulnerabilities

### Runtime Security

1. **Read-Only Filesystem**
   ```yaml
   read_only: true
   tmpfs:
     - /tmp:noexec,nosuid,size=64M
   ```

2. **Dropped Capabilities**
   ```yaml
   cap_drop:
     - ALL
   ```

3. **No New Privileges**
   ```yaml
   security_opt:
     - no-new-privileges:true
   ```

4. **Resource Limits**
   - Prevents DoS attacks
   - Ensures fair resource distribution

### Scanning for Vulnerabilities

```bash
# Install Trivy (macOS)
brew install trivy

# Scan images
make docker-scan

# Or manually
trivy image bookie-grpc-server:latest
trivy image bookie-http-client:latest
```

### Network Security

- Services communicate over dedicated Docker network
- gRPC server not exposed externally by default
- Only HTTP client needs public exposure

## Monitoring & Health Checks

### Current Implementation

The distroless images don't include health check binaries. For production, consider:

1. **External Health Checks**
   - Load balancer health probes
   - Kubernetes liveness/readiness probes

2. **Implement Health Endpoints**
   ```go
   // Add to server
   func (s *bookieService) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
       return &health.HealthCheckResponse{
           Status: health.HealthCheckResponse_SERVING,
       }, nil
   }
   ```

3. **gRPC Health Probe** (Recommended)
   ```dockerfile
   # Add to Dockerfile
   COPY --from=grpc-ecosystem/grpc-health-probe:latest /ko-app/grpc-health-probe /
   ```

### Logs

```bash
# View all logs
docker-compose logs -f

# Filter by service
docker-compose logs -f grpc-server

# View last 100 lines
docker-compose logs --tail=100 http-client

# Follow logs with timestamps
docker-compose logs -f -t
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

```bash
# Find process using port
lsof -i :8080
lsof -i :8020

# Kill the process or change ports in docker-compose.yml
```

#### 2. Connection Refused (Client → Server)

- Ensure both services are on the same network
- Check `GRPC_SERVER_ADDR` environment variable
- Verify server is running: `docker-compose ps`

#### 3. Permission Denied

```bash
# Check file permissions
ls -la Dockerfile.*

# Fix if needed
chmod 644 Dockerfile.*
```

#### 4. Build Failures

```bash
# Clean and rebuild
make docker-clean
make docker-build

# Check Go modules
go mod verify
go mod tidy
```

#### 5. Image Size Issues

```bash
# Check image sizes
docker images | grep bookie

# Expected: ~15-20MB per image
# If larger, check .dockerignore
```

### Debugging

```bash
# Exec into running container (won't work with distroless)
# Use alpine-based debug image for troubleshooting

# View container details
docker inspect bookie-grpc-server

# Check resource usage
docker stats

# View network configuration
docker network inspect bookie-grpc_bookie-network
```

## Next Steps

### Kubernetes Deployment

The current Docker setup is K8s-ready. Key considerations:

1. **Convert to K8s Manifests**
   - Deployments for server and client
   - Services for networking
   - ConfigMaps for configuration
   - Secrets for sensitive data

2. **Use kompose for quick conversion**
   ```bash
   kompose convert -f docker-compose.yml
   ```

3. **Add Kubernetes-specific features**
   - Horizontal Pod Autoscaling
   - Ingress for external access
   - PersistentVolumes (if needed)
   - NetworkPolicies for security

### Observability

1. **Metrics**
   - Prometheus instrumentation
   - Custom metrics for gRPC and HTTP

2. **Tracing**
   - OpenTelemetry integration
   - Distributed tracing across services

3. **Logging**
   - Structured logging (JSON)
   - Log aggregation (ELK, Loki)

## Best Practices Implemented

✅ Multi-stage builds (minimal image size)  
✅ Distroless images (enhanced security)  
✅ Non-root user (security)  
✅ Read-only filesystem (immutability)  
✅ Resource limits (stability)  
✅ Dropped capabilities (least privilege)  
✅ Graceful shutdown (signal handling)  
✅ Environment-based configuration  
✅ Proper .dockerignore (faster builds)  
✅ Network isolation  
✅ Logging configuration  
✅ Version tagging strategy  

## Resources

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Distroless Images](https://github.com/GoogleContainerTools/distroless)
- [gRPC Health Checking](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)
- [Container Security](https://cheatsheetseries.owasp.org/cheatsheets/Docker_Security_Cheat_Sheet.html)

