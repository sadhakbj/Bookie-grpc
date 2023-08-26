# Makefile

# Variables
PROTO_DIR := protos
OUTPUT_DIR := protos/bookie

# Targets
all: generate

generate:
	protoc \
	--proto_path=$(PROTO_DIR) \
	--go_out=$(OUTPUT_DIR) \
	--go_opt=paths=source_relative \
	--go-grpc_out=$(OUTPUT_DIR) \
	--go-grpc_opt=paths=source_relative \
	$(PROTO_DIR)/*.proto

.PHONY: lint
lint:
	golangci-lint run ${args} ./ ...

.PHONY: lint-fix
lint-fix:
	@make lint args=' --fix -v' cons_args='-v'

clean:
	rm -rf $(OUTPUT_DIR)/*

serve:
	go run src/cmd/server/main.go

client:
	go run src/cmd/client/main.go

.PHONY: all generate clean
