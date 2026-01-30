.DEFAULT_GOAL := help

KIND_CLUSTER ?= mycelis-cluster

.PHONY: help setup clean test-api runner proto build-core
.PHONY: docker-build kind-load deploy restart up logs

help:
	@echo "Mycelis Service Network - Workflow"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Gen-2 (Go Migration):"
	@echo "  proto         Generate Go code from Protobuf definitions"
	@echo "  build-core    Build the Go Core binary"
	@echo ""
	@echo "Gen-2 Deployment (Kind):"
	@echo "  up            Build -> Load -> Deploy"
	@echo "  docker-build  Build Core Docker image"
	@echo "  kind-load     Load image to Kind"
	@echo "  restart       Restart Neural Core"
	@echo ""
	@echo "Local Development (Recommended):"
	@echo "  dev-up        Start infrastructure (NATS/Postgres) via Docker"
	@echo "  dev-down      Stop infrastructure"
	@echo "  dev-api       Run API locally (hot-reload)"
	@echo "  dev-ui        Run UI locally"
	@echo ""
	@echo "Legacy Kubernetes Deployment:"
	@echo "  k8s-up        Create Kind cluster"
	@echo "  k8s-init      Build & deploy EVERYTHING to Kind"
	@echo "  k8s-down      Delete Kind cluster"
	@echo ""
	@echo "Utilities:"
	@echo "  setup         Install dependencies (uv)"
	@echo "  clean         Remove artifacts"

# -----------------------------------------------------------------------------
# Gen-2 Targets (Go)
# -----------------------------------------------------------------------------

proto:
	@uv run scripts/dev.py proto

proto-go: proto

proto-py:
	@uv run scripts/dev.py proto-py

build-core:
	@uv run scripts/dev.py build-go

test-hybrid:
	@echo "--- Testing Go Core ---"
	@cd core && go test ./internal/state/...
	@echo "--- Testing Python Relay ---"
	@uv run scripts/dev.py test-python


# -----------------------------------------------------------------------------
# Local Development Targets
# -----------------------------------------------------------------------------

# Start Infra using Universal Runner
dev-up:
	@uv run scripts/dev.py up

dev-down:
	@uv run scripts/dev.py down

# Setup dependencies
setup:
	@echo "Setting up project..."
	@uv sync
	@cd ui && npm install
	@echo "Setup complete."

# Run backend tests
test-api:
	@echo "Running backend tests..."
	@uv run pytest

# Run Agent Runner locally
runner:
	@echo "Starting Agent Runner..."
	@uv run python runner/main.py

# Run API locally
dev-api:
	@echo "Starting API locally..."
	@export DATABASE_URL="postgresql+asyncpg://mycelis:password@localhost/mycelis" && \
	export NATS_URL="nats://localhost:4222" && \
	export OLLAMA_BASE_URL="http://192.168.50.156:11434" && \
	uv run uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload

# Run UI locally
dev-ui:
	@echo "Cleaning UI cache..."
	@rm -rf ui/.next
	@echo "Starting UI on 0.0.0.0:3000..."
	@export API_INTERNAL_URL="http://127.0.0.1:8000" && cd ui && npm run dev -- -H 0.0.0.0 -p 3000

# -----------------------------------------------------------------------------
# Clean
# -----------------------------------------------------------------------------
clean:
	@echo "Cleaning up..."
	@rm -rf .venv ui/node_modules ui/.next
	@find . -type d -name "__pycache__" -exec rm -rf {} +
	@echo "Clean complete."

# -----------------------------------------------------------------------------
# Legacy Kubernetes Targets (Preserved for Reference)
# -----------------------------------------------------------------------------
# (Imports or older commands can stay if needed, simplified for clarity here)
k8s-up:
	@kind create cluster --name $(KIND_CLUSTER) --config kind-config.yaml

k8s-down:
	@kind delete cluster --name $(KIND_CLUSTER)

# -----------------------------------------------------------------------------
# Deployment Targets (Gen-2)
# -----------------------------------------------------------------------------

docker-build:
	docker build -t mycelis/core:latest -f core/Dockerfile .

kind-load:
	kind load docker-image mycelis/core:latest --name $(KIND_CLUSTER)

deploy:
	kubectl apply -f k8s/

restart:
	kubectl rollout restart deployment/neural-core -n mycelis

# Composite
up: docker-build kind-load deploy
	@echo "Deployed to Kind! Waiting for rollout..."
	@kubectl wait --for=condition=available --timeout=480s deployment/neural-core -n mycelis

# Logs
logs:
	kubectl logs -l app=neural-core -n mycelis -f

