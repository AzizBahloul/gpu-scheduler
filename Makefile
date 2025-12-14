.PHONY: help build test lint proto docker-build deploy clean

help: ## Display this help message
	@echo "GPU Scheduler - Make targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build all binaries
	@echo "Building scheduler..."
	go build -o bin/scheduler ./cmd/scheduler
	@echo "Building agent..."
	go build -o bin/agent ./cmd/agent
	@echo "Building CLI..."
	go build -o bin/gpu-cli ./cmd/cli

test: ## Run unit tests
	go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests
	go test -v -tags=integration ./tests/integration/...

bench: ## Run benchmarks
	go test -bench=. -benchmem ./tests/benchmarks/...

lint: ## Run linter
	golangci-lint run ./...

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	./scripts/generate-proto.sh

docker-build: ## Build Docker images
	docker build -f deploy/docker/Dockerfile.scheduler -t gpu-scheduler:latest .
	docker build -f deploy/docker/Dockerfile.agent -t gpu-agent:latest .
	docker build -f deploy/docker/Dockerfile.cli -t gpu-cli:latest .

deploy: ## Deploy to Kubernetes using Helm
	helm upgrade --install gpu-scheduler ./deploy/helm --namespace gpu-system --create-namespace

deploy-local: ## Deploy to local kind cluster
	./scripts/setup-cluster.sh
	make deploy

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

frontend-install: ## Install frontend dependencies
	cd web/frontend && npm install

frontend-dev: ## Run frontend in dev mode
	cd web/frontend && npm run dev

frontend-build: ## Build frontend for production
	cd web/frontend && npm run build

migrate-up: ## Run database migrations up
	./scripts/run-migrations.sh up

migrate-down: ## Run database migrations down
	./scripts/run-migrations.sh down

run-scheduler: build ## Run scheduler locally
	./bin/scheduler --config config/scheduler-config.yaml

run-agent: build ## Run agent locally
	./bin/agent --config config/agent-config.yaml

install-tools: ## Install development tools
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

tidy: ## Tidy go modules
	go mod tidy

vendor: ## Vendor dependencies
	go mod vendor
