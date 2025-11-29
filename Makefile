.DEFAULT_GOAL := help

.PHONY: help setup start stop clean dev ui api infra

help:
	@echo "Mycelis Service Network Makefile"
	@echo "Usage: make [target]"
	@echo ""
	@echo "General Targets:"
	@echo "  setup       Install dependencies (uv, npm)"
	@echo "  dev         Start infra and print instructions for API/UI"
	@echo "  stop-apps   Stop all application services (API & UI)"
	@echo "  shutdown    Full shutdown (Infra + Apps)"
	@echo ""
	@echo "Backend Targets:"
	@echo "  api         Run API server (dev mode)"
	@echo "  stop-api    Stop API service only"
	@echo "  test-api    Run backend tests (pytest)"
	@echo ""
	@echo "Frontend Targets:"
	@echo "  ui          Run UI server (dev mode)"
	@echo "  stop-ui     Stop UI service only"
	@echo ""
	@echo "Infrastructure Targets:"
	@echo "  infra       Start infrastructure (NATS, Postgres) in Docker"
	@echo "  stop        Stop infrastructure"
	@echo "  logs        View infrastructure logs"
	@echo ""
	@echo "Maintenance Targets:"
	@echo "  clean       Remove artifacts and dependencies"
	@echo "  help        Show this help message"


# Setup dependencies
setup:
	@echo "Setting up project..."
	@uv sync
	@cd ui && npm install
	@echo "Setup complete."

# Start infrastructure (NATS)
infra:
	@echo "Starting infrastructure..."
	@docker compose up -d

# Aliases for starting infrastructure
services: infra
up: infra

# View infrastructure logs
logs:
	@docker compose logs -f

# Run API in development mode
api:
	@echo "Starting API..."
	@uv run uvicorn api.main:app --reload --port 8000

# Run backend tests
test-api:
	@echo "Running backend tests..."
	@uv run pytest

# Run UI in development mode
ui:
	@echo "Starting UI..."
	@cd ui && npm run dev

# Run MCP Bridge locally
run-bridge:
	@echo "Starting MCP Bridge..."
	@uv run services/mcp_bridge.py

# Run a specific agent locally
# Usage: make run-agent NAME=a1
run-agent:
	@echo "Starting Agent $(NAME)..."
	@uv run agents/runner.py $(NAME) --api http://localhost:8000 --nats nats://localhost:4222

# Local Dev Help
dev-help:
	@echo "Local Development Commands:"
	@echo "  make infra        - Start NATS & Postgres (Docker)"
	@echo "  make api          - Start API (Local)"
	@echo "  make ui           - Start UI (Local)"
	@echo "  make run-bridge   - Start MCP Bridge (Local)"
	@echo "  make run-agent NAME=<name> - Start Agent (Local)"


# Start everything (requires multiple terminals or backgrounding)
dev: infra
	@echo "Infrastructure started. Please run 'make api' and 'make ui' in separate terminals."

# Stop infrastructure
stop:
	@echo "Stopping infrastructure..."
	@docker compose down

# Alias for stop
down: stop

# Stop API
stop-api:
	@echo "Stopping API..."
	@-pkill -f "uvicorn api.main:app" || true

# Stop UI
stop-ui:
	@echo "Stopping UI..."
	@-pkill -f "next" || true

# Stop all application services
stop-apps: stop-api stop-ui
	@echo "Application services stopped."

# Kill running application processes (Alias for stop-apps)
kill: stop-apps

# Full shutdown (Infrastructure + Apps)
shutdown: stop kill
	@echo "Project shut down completely."

# Check if infrastructure is running
check-infra:
	@echo "Checking infrastructure..."
	@if ! docker compose ps | grep "Up"; then \
		echo "Error: Infrastructure is not running. Please run 'make infra' first."; \
		exit 1; \
	fi

# Forward K8s ports to localhost (Hybrid Dev)
k8s-forward:
	@echo "Forwarding K8s services to localhost..."
	@echo "NATS: 4222, Postgres: 5432"
	@trap 'kill %1 %2' SIGINT; \
	kubectl port-forward svc/nats 4222:4222 -n mycelis & \
	kubectl port-forward svc/postgres 5432:5432 -n mycelis & \
	wait

# Clean artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf .venv
	@rm -rf ui/node_modules
	@rm -rf ui/.next
	@find . -type d -name "__pycache__" -exec rm -rf {} +
	@echo "Clean complete."

# -----------------------------------------------------------------------------
# Kubernetes Targets
# -----------------------------------------------------------------------------

KIND_CLUSTER ?= kind
SERVICE ?= api

.PHONY: k8s-build k8s-load k8s-apply k8s-dev k8s-up k8s-down k8s-reset

# Create Kind cluster and install Ingress Controller
k8s-up:
	@echo "Creating Kind cluster '$(KIND_CLUSTER)'..."
	@kind create cluster --name $(KIND_CLUSTER) --config kind-config.yaml
	@echo "Installing Nginx Ingress Controller..."
	@kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	@echo "Waiting for Ingress Controller to be ready..."
	@kubectl wait --namespace ingress-nginx \
	  --for=condition=ready pod \
	  --selector=app.kubernetes.io/component=controller \
	  --timeout=90s

# Delete Kind cluster
k8s-down:
	@echo "Deleting Kind cluster '$(KIND_CLUSTER)'..."
	@kind delete cluster --name $(KIND_CLUSTER)

# Reset cluster (Down + Up)
k8s-reset: k8s-down k8s-up

# Build Docker images
k8s-build:
	@echo "Building Docker images..."
	@docker build -t mycelis-api:latest -f api/Dockerfile .
	@docker build -t mycelis-ui:latest -f ui/Dockerfile ui/

# Load images into Kind
k8s-load:
	@echo "Loading images into Kind cluster '$(KIND_CLUSTER)'..."
	@kind load docker-image mycelis-api:latest --name $(KIND_CLUSTER)
	@kind load docker-image mycelis-ui:latest --name $(KIND_CLUSTER)

# Apply manifests
k8s-apply:
	@echo "Applying Kubernetes manifests..."
	@kubectl apply -f k8s/

# Standardized dev loop for a specific service
# Usage: make k8s-dev service=api
k8s-dev:
	@echo "Deploying update for service: $(SERVICE)"
	@if [ "$(SERVICE)" = "api" ]; then \
		docker build -t mycelis-api:latest -f api/Dockerfile .; \
		kind load docker-image mycelis-api:latest --name $(KIND_CLUSTER); \
		kubectl rollout restart deployment/api -n mycelis; \
		kubectl rollout restart deployment/mcp-bridge -n mycelis; \
	elif [ "$(SERVICE)" = "ui" ]; then \
		docker build -t mycelis-ui:latest -f ui/Dockerfile ui/; \
		kind load docker-image mycelis-ui:latest --name $(KIND_CLUSTER); \
		kubectl rollout restart deployment/ui -n mycelis; \
	else \
		echo "Unknown service: $(SERVICE). Available: api, ui"; \
	fi

