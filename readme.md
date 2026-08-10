# Bookie gRPC - Learning Project

A production-ready gRPC microservices application demonstrating gRPC, BFF (Backend-for-Frontend) pattern, containerization, and cloud-native best practices.

## 🏗️ Architecture

```
┌─────────┐      HTTP      ┌──────────────┐     gRPC      ┌─────────────┐
│ Browser │ ◄─────────────► │ HTTP Client  │ ◄────────────► │ gRPC Server │
│         │   :8080         │    (BFF)     │    :8020      │   (Bookie)  │
└─────────┘                 └──────────────┘               └─────────────┘
```

### Components

- **gRPC Server** (Port 8020): Backend service managing books using gRPC protocol
- **HTTP Client/BFF** (Port 8080): REST API gateway that translates HTTP to gRPC calls

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose (recommended)
- Or Go 1.26+ for local development

### Run with Docker (Easiest Way)

```bash
# 1. Build the images
make docker-build

# 2. Start the services
make docker-up

# 3. Test it works
curl http://localhost:8080/books

# 4. View logs (optional)
make docker-logs

# 5. Stop when done
make docker-down
```

**That's it!** Services will be available at:

- 🌐 HTTP API: http://localhost:8080
- 🔌 gRPC Server: localhost:8020

### Alternative: Helper Script

```bash
./scripts/docker-dev.sh build   # Build images
./scripts/docker-dev.sh up      # Start services
./scripts/docker-dev.sh logs    # View logs
./scripts/docker-dev.sh down    # Stop services
./scripts/docker-dev.sh help    # See all commands
```

### Run Locally Without Docker

**One-time setup:**

```bash
# Install protoc first (Protocol Buffers compiler)
# macOS:
brew install protobuf

# Linux:
sudo apt-get install -y protobuf-compiler

# Then run setup (installs all Go dependencies and plugins)
make setup
```

**Run:**

```bash
make serve     # Terminal 1: Start gRPC server
make client    # Terminal 2: Start HTTP client
```

## 📚 API Examples

### Get All Books

```bash
curl http://localhost:8080/books
```

### Get Book by ID

```bash
curl http://localhost:8080/books/1234
```

## 🐳 Docker

### Common Commands

```bash
# Development
make docker-build       # Build images
make docker-up          # Start services in background
make docker-down        # Stop services
make docker-logs        # View logs (all services)
make docker-logs-server # View gRPC server logs
make docker-logs-client # View HTTP client logs
make docker-restart     # Restart services
make docker-clean       # Remove everything

# Production
make docker-prod-up     # Start with production config
make docker-scan        # Security scan (needs trivy)
```

### Features

- ✅ **Minimal images**: ~15-20MB (distroless base)
- ✅ **Secure**: Non-root user, read-only filesystem
- ✅ **Fast**: Multi-stage builds with caching
- ✅ **Production-ready**: Resource limits, graceful shutdown

See [DOCKER.md](DOCKER.md) for detailed documentation.

## ☸️ Kubernetes

The app ships with production-ready Kubernetes manifests in [`k8s/`](k8s/), managed with [Kustomize](https://kustomize.io/). They work on any cluster — OrbStack, minikube, kind, GKE, EKS, AKS.

### Run locally with OrbStack (macOS)

OrbStack includes a built-in Kubernetes cluster:

```bash
# 1. Enable Kubernetes (once): OrbStack app → Settings → Kubernetes, or:
orb start k8s

# 2. Confirm kubectl points at it
kubectl config current-context   # should print: orbstack

# 3. Deploy the app
kubectl apply -k k8s/

# 4. Wait for pods to be ready
kubectl get pods -n bookie -w

# 5. Test the API
kubectl port-forward -n bookie svc/http-client 8080:8080
curl http://localhost:8080/books
curl http://localhost:8080/books/1234
```

> 💡 **Offline development:** the manifests use `imagePullPolicy: IfNotPresent`, so locally built images are used when present — otherwise pods pull the latest images published to ghcr.io by CI:
> ```bash
> docker build -t ghcr.io/sadhakbj/bookie-grpc-server:latest -f Dockerfile.server .
> docker build -t ghcr.io/sadhakbj/bookie-http-client:latest -f Dockerfile.client .
> ```

### Going to the cloud (GKE, EKS, AKS)

The same manifests work as-is:

```bash
gcloud container clusters get-credentials my-cluster --region us-central1   # GKE example
kubectl apply -k k8s/
```

Production tweaks to consider:

- **Expose the API**: change `k8s/http-client-service.yaml` to `type: LoadBalancer` (cloud allocates an external IP) or add an Ingress.
- **Pin versions**: override `images` in `k8s/kustomization.yaml` to a specific `sha-<hash>` tag instead of `latest` for reproducible rollouts.
- **Roll out a new version** (images are published by CI on every merge to main):
  ```bash
  kubectl -n bookie set image deployment/grpc-server grpc-server=ghcr.io/sadhakbj/bookie-grpc-server:sha-abc1234
  kubectl -n bookie set image deployment/http-client http-client=ghcr.io/sadhakbj/bookie-http-client:sha-abc1234
  kubectl -n bookie rollout status deployment/grpc-server
  ```
- **Scale out**: `kubectl -n bookie scale deployment/grpc-server --replicas=3`
- **Clean up**: `kubectl delete -k k8s/`

## 📊 Observability (Coming)

- Prometheus metrics
- Jaeger tracing
- Grafana dashboards

## 🛠️ Development

### Project Structure

```
.
├── protos/              # Protocol Buffer definitions
│   ├── book.proto
│   └── bookie/          # Generated code
├── src/
│   ├── cmd/
│   │   ├── server/      # gRPC server
│   │   └── client/      # HTTP client (BFF)
│   └── internal/
│       ├── client/      # HTTP controllers
│       ├── services/    # gRPC client service
│       └── utils/       # Shared utilities
├── scripts/
│   ├── build.sh
│   └── docker-dev.sh    # Docker helper
├── Dockerfile.server    # gRPC server image
├── Dockerfile.client    # HTTP client image
├── docker-compose.yml   # Development environment
├── docker-compose.prod.yml  # Production environment
└── k8s/                 # Kubernetes manifests (Kustomize)
│   ├── namespace.yaml
│   ├── grpc-server-deployment.yaml
│   ├── grpc-server-service.yaml
│   ├── http-client-deployment.yaml
│   ├── http-client-service.yaml
│   └── kustomization.yaml
```

### Makefile Commands

```bash
make setup         # One-time setup (install all dependencies)
make generate      # Generate protobuf code
make build         # Build binaries
make serve         # Run gRPC server
make client        # Run HTTP client
make lint          # Run linter
make lint-fix      # Fix linting issues
```

### Generating Protocol Buffers

```bash
make generate
```

This generates:

- `protos/bookie/book.pb.go` - Protocol Buffer messages
- `protos/bookie/book_grpc.pb.go` - gRPC service code

### Code Quality

```bash
# Run linter
make lint

# Auto-fix issues
make lint-fix
```

## 🔒 Security

This project implements comprehensive security best practices:

- **Container Security**: Distroless images, non-root user, read-only filesystem
- **Runtime Security**: Dropped capabilities, resource limits, no new privileges
- **Supply Chain**: Multi-stage builds, dependency verification, image scanning
- **CI/CD**: Automated security scanning, SARIF reports, image signing

See [SECURITY.md](SECURITY.md) for detailed security information.

### Security Scanning

```bash
# Scan images for vulnerabilities
make docker-scan

# Or manually with Trivy
trivy image bookie-grpc-server:latest
trivy image bookie-http-client:latest
```

## 🔄 CI/CD

GitHub Actions workflows included for:

- Automated testing
- Docker image publishing to GitHub Container Registry (ghcr.io)
- Security scanning (Trivy)
- Image signing (Cosign)
- Multi-platform builds (amd64, arm64)

See `.github/workflows/docker-publish.yml` — builds both images multi-arch and pushes `latest`, a git-tag version, and a unique `sha-<hash>` tag to ghcr.io on every merge to main.

## 🧪 Testing

```bash
# Run tests (when implemented)
go test -v ./src/...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./src/...

# View coverage
go tool cover -html=coverage.out
```

## 🎯 Learning Goals

This project demonstrates:

- ✅ gRPC service implementation
- ✅ BFF (Backend-for-Frontend) pattern
- ✅ Protocol Buffers
- ✅ Docker containerization
- ✅ Multi-stage builds
- ✅ Security hardening
- ✅ Production-ready configuration
- 🔄 Kubernetes deployment (Next)
- 🔄 Observability (Metrics, Tracing, Logging) (Next)
- 🔄 Service mesh (Istio/Linkerd) (Future)

## 🤝 Contributing

This is a learning project, but contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## 📝 License

This is a learning project. Feel free to use it for educational purposes.

## 🙏 Acknowledgments

Built with:

- [gRPC](https://grpc.io/)
- [Protocol Buffers](https://protobuf.dev/)
- [Go](https://golang.org/)
- [Docker](https://www.docker.com/)
- [Distroless Images](https://github.com/GoogleContainerTools/distroless)

## 📬 Contact

For questions or feedback about this learning project, please open an issue.

---

**Happy Learning! 🚀**
