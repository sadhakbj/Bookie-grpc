# Makefile

# Variables
PROTO_DIR := protos
OUTPUT_DIR := protos/bookie

# Targets
all: generate

.PHONY: setup
setup:
	@echo "🔧 Setting up development environment..."
	@echo ""
	@echo "1. Installing Go dependencies..."
	@go mod download
	@echo "✅ Go dependencies installed"
	@echo ""
	@echo "2. Installing protobuf Go plugins..."
	@export PATH=$$PATH:$$(go env GOPATH)/bin && \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
		go install go.uber.org/mock/mockgen@latest
	@echo "✅ Protobuf plugins and mockgen installed"
	@echo ""
	@echo "3. Checking protoc..."
	@which protoc > /dev/null 2>&1 && echo "✅ protoc found" || \
		(echo "⚠️  protoc not found. Install with:" && \
		 echo "   macOS: brew install protobuf" && \
		 echo "   Linux: sudo apt-get install -y protobuf-compiler")
	@echo ""
	@echo "📝 Add this to your ~/.zshrc (or ~/.bashrc):"
	@echo "   export PATH=\$$PATH:\$$(go env GOPATH)/bin"
	@echo ""
	@echo "🎉 Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  • Docker: make docker-build && make docker-up"
	@echo "  • Local:  make serve (in one terminal) && make client (in another)"

generate: check-deps
	protoc \
	--proto_path=$(PROTO_DIR) \
	--go_out=$(OUTPUT_DIR) \
	--go_opt=paths=source_relative \
	--go-grpc_out=$(OUTPUT_DIR) \
	--go-grpc_opt=paths=source_relative \
	$(PROTO_DIR)/*.proto

.PHONY: lint
lint:
	golangci-lint run ${args} ./src/...

.PHONY: lint-fix
lint-fix:
	@make lint args=' --fix -v' cons_args='-v'

# Regenerate mocks (usage: make gen/mocks) - you can also click the
# green go:generate gutter icon in GoLand on the directive file
.PHONY: gen/mocks
gen/mocks:
	go generate ./...

# Run all tests (usage: make test [args="-run TestGrpc -v"])
.PHONY: test local-config
test:
	go test ${args} ./src/...

local-config:
	@test -f .env || cp config.example .env

clean:
	rm -rf $(OUTPUT_DIR)


serve: local-config
	go run src/cmd/server/main.go

client: local-config
	go run src/cmd/client/main.go

deps:
	@echo "Downloading Go dependencies..."
	go mod download
	@echo ""
	@echo "✅ Go dependencies installed"
	@echo ""
	@echo "Note: If you need to generate protobuf code, ensure you have:"
	@echo "  1. protoc installed (brew install protobuf)"
	@echo "  2. Go plugins installed (see README.md)"

.PHONY: check-deps
check-deps:
	@echo "Checking dependencies..."
	@which protoc > /dev/null || (echo "❌ protoc not found. Install: brew install protobuf" && exit 1)
	@which protoc-gen-go > /dev/null || (echo "❌ protoc-gen-go not found. Install: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" && exit 1)
	@which protoc-gen-go-grpc > /dev/null || (echo "❌ protoc-gen-go-grpc not found. Install: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" && exit 1)
	@echo "✅ All dependencies found!"

vendor:
	go mod vendor

.PHONY: build
build:
	@./scripts/build.sh

# Docker commands
.PHONY: docker-build
docker-build:
	docker build -t bookie-grpc-server:latest -f Dockerfile.server .
	docker build -t bookie-http-client:latest -f Dockerfile.client .

.PHONY: docker-build-server
docker-build-server:
	docker build -t bookie-grpc-server:latest -f Dockerfile.server .

.PHONY: docker-build-client
docker-build-client:
	docker build -t bookie-http-client:latest -f Dockerfile.client .

.PHONY: docker-up
docker-up:
	docker-compose up -d

.PHONY: docker-down
docker-down:
	docker-compose down

.PHONY: docker-logs
docker-logs:
	docker-compose logs -f

.PHONY: docker-logs-server
docker-logs-server:
	docker-compose logs -f grpc-server

.PHONY: docker-logs-client
docker-logs-client:
	docker-compose logs -f http-client

.PHONY: docker-restart
docker-restart:
	docker-compose restart

.PHONY: docker-clean
docker-clean:
	docker-compose down -v --rmi all

# Observability stack commands
.PHONY: observability-up
observability-up:
	docker-compose up -d jaeger prometheus loki alloy grafana

.PHONY: observability-down
observability-down:
	docker-compose stop jaeger prometheus loki alloy grafana

.PHONY: observability-logs
observability-logs:
	docker-compose logs -f jaeger prometheus loki alloy grafana

.PHONY: obs
obs: docker-build docker-up observability-up
	@echo ""
	@echo "🎉 Full stack running!"
	@echo "  gRPC server: localhost:8020"
	@echo "  HTTP BFF:    localhost:8080"
	@echo "  Grafana:     http://localhost:3000 (admin/admin)"
	@echo "  Jaeger UI:   http://localhost:16686"
	@echo "  Prometheus:  http://localhost:9090"
	@echo "  Loki:        http://localhost:3100"

# Generate demo traffic so the Grafana dashboard has data to show
.PHONY: traffic
traffic:
	@echo "🚗 Generating traffic for 60s..."
	@for i in $$(seq 1 30); do \
		curl -s http://$${BOOKIE_HOST:-localhost}:8080/books/1234 > /dev/null; \
		curl -s http://$${BOOKIE_HOST:-localhost}:8080/books > /dev/null; \
		curl -s "http://$${BOOKIE_HOST:-localhost}:8080/books/$$(( ( $$RANDOM % 9000 ) + 1000 ))" > /dev/null; \
		sleep 1; \
	done
	@echo "✅ Done — check Grafana/Jaeger"

# Security scanning with trivy (install: brew install trivy)
.PHONY: docker-scan
docker-scan:
	trivy image bookie-grpc-server:latest
	trivy image bookie-http-client:latest

.PHONY: all generate clean
