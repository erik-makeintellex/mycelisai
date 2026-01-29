.DEFAULT_GOAL := help

KIND_CLUSTER ?= kind

.PHONY: help setup clean test-api runner

help:
	@echo "Mycelis Service Network - Workflow"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Local Development (Recommended):"
	@echo "  k8s-services      Deploy infra (NATS, Postgres) to Kind"
	@echo "  k8s-forward       Port-forward infra (Idempotent + Verify)"
	@echo "  k8s-stop-forward  Stop port forwarding"
	@echo "  dev-api           Run API locally (hot-reload)"
	@echo "  dev-ui            Run UI locally"
	@echo ""
	@echo "Full Kubernetes Deployment:"
	@echo "  k8s-up        Create Kind cluster"
	@echo "  k8s-init      Build & deploy EVERYTHING to Kind"
	@echo "  k8s-down      Delete Kind cluster"
	@echo ""
	@echo "Utilities:"
	@echo "  setup         Install dependencies"
	@echo "  k8s-status    Check service status"
	@echo "  k8s-logs      Tail logs (SERVICE=nats|postgres)"
	@echo "  clean         Remove artifacts"

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

# -----------------------------------------------------------------------------
# Local Development Targets
# -----------------------------------------------------------------------------

# Run API locally (requires k8s-services + k8s-forward)
dev-api:
	@echo "Starting API locally..."
	@export DATABASE_URL="postgresql+asyncpg://mycelis:password@localhost/mycelis" && \
	export NATS_URL="nats://localhost:4222" && \
	export OLLAMA_BASE_URL="http://192.168.50.156:11434" && \
	uv run uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload

# Run UI locally (Clean Start, Network Accessible)
dev-ui:
	@echo "Cleaning UI cache..."
	@rm -rf ui/.next
	@echo "Starting UI on 0.0.0.0:3000..."
	@export API_INTERNAL_URL="http://127.0.0.1:8000" && cd ui && npm run dev -- -H 0.0.0.0 -p 3000

# Run Agent Runner locally
dev-runner:
	@echo "Starting Agent Runner..."
	@export DATABASE_URL="postgresql+asyncpg://mycelis:password@localhost/mycelis" && \
	export NATS_URL="nats://localhost:4222" && \
	export OLLAMA_BASE_URL="http://192.168.50.156:11434" && \
	uv run python runner/main.py

# Clean artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf .venv ui/node_modules ui/.next
	@find . -type d -name "__pycache__" -exec rm -rf {} +
	@echo "Clean complete."

# -----------------------------------------------------------------------------
# Kubernetes Targets
# -----------------------------------------------------------------------------

SERVICE ?= api

.PHONY: k8s-up k8s-down k8s-reset k8s-build k8s-load k8s-apply k8s-init
.PHONY: k8s-dev k8s-status k8s-logs k8s-forward k8s-restart

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
	@echo "Cluster ready!"

# Delete Kind cluster
k8s-down:
	@echo "Deleting Kind cluster '$(KIND_CLUSTER)'..."
	@kind delete cluster --name $(KIND_CLUSTER)

# Reset cluster (Down + Up)
k8s-reset: k8s-down k8s-up k8s-init

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

# Apply all k8s manifests
k8s-apply:
	@echo "Applying Kubernetes manifests..."
	@kubectl apply -f k8s/

# Full initialization (build + load + deploy)
k8s-init: k8s-build k8s-load k8s-apply
	@echo "Waiting for deployments to be ready..."
	@kubectl wait --for=condition=available --timeout=120s deployment/nats -n mycelis
	@kubectl wait --for=condition=available --timeout=120s deployment/postgres -n mycelis
	@kubectl wait --for=condition=available --timeout=120s deployment/api -n mycelis
	@kubectl wait --for=condition=available --timeout=120s deployment/ui -n mycelis
	@echo "All services deployed!"

# Backend-only initialization (No UI, No Agents)
k8s-backend: k8s-build-backend k8s-load-backend k8s-apply-backend
	@echo "Waiting for backend deployments..."
	@kubectl wait --for=condition=available --timeout=120s deployment/nats -n mycelis
	@kubectl wait --for=condition=available --timeout=120s deployment/postgres -n mycelis
	@kubectl wait --for=condition=available --timeout=120s deployment/api -n mycelis
	@echo "Backend services deployed!"

k8s-build-backend:
	@echo "Building Backend Docker images..."
	@docker build -t mycelis-api:latest -f api/Dockerfile .

k8s-load-backend:
	@echo "Loading Backend images into Kind..."
	@kind load docker-image mycelis-api:latest --name $(KIND_CLUSTER)

k8s-apply-backend:
	@echo "Applying Backend Kubernetes manifests..."
	@kubectl apply -f k8s/00-namespace.yaml
	@kubectl apply -f k8s/nats.yaml
	@kubectl apply -f k8s/postgres.yaml
	@kubectl apply -f k8s/api.yaml
	@kubectl apply -f k8s/mcp-bridge.yaml
	@kubectl apply -f k8s/ingress.yaml

# Deploy only infrastructure services (NATS, Postgres)
k8s-services:
	@echo "Deploying infrastructure services..."
	@kubectl apply -f k8s/00-namespace.yaml
	@kubectl apply -f k8s/nats.yaml
	@kubectl apply -f k8s/postgres.yaml
	@echo "Waiting for infrastructure..."
	@kubectl wait --for=condition=available --timeout=120s deployment/nats -n mycelis
	@kubectl wait --for=condition=available --timeout=120s deployment/postgres -n mycelis
	@echo "Infrastructure ready!"

# Quick dev loop for specific service
# Usage: make k8s-dev SERVICE=api
k8s-dev:
	@echo "Rebuilding and restarting: $(SERVICE)"
	@if [ "$(SERVICE)" = "api" ]; then \
		docker build -t mycelis-api:latest -f api/Dockerfile .; \
		kind load docker-image mycelis-api:latest --name $(KIND_CLUSTER); \
		kubectl rollout restart deployment/api -n mycelis; \
	elif [ "$(SERVICE)" = "ui" ]; then \
		docker build -t mycelis-ui:latest -f ui/Dockerfile ui/; \
		kind load docker-image mycelis-ui:latest --name $(KIND_CLUSTER); \
		kubectl rollout restart deployment/ui -n mycelis; \
	else \
		echo "Unknown service: $(SERVICE). Available: api, ui"; \
		exit 1; \
	fi
	@echo "Waiting for rollout..."
	@kubectl rollout status deployment/$(SERVICE) -n mycelis

# Check status of all services
k8s-status:
	@echo "=== Mycelis Services Status ==="
	@kubectl get all -n mycelis
	@echo ""
	@echo "=== Ingress Status ==="
	@kubectl get ingress -n mycelis
	@echo ""
	@echo "Access URLs:"
	@echo "  UI:  http://localhost/"
	@echo "  API: http://localhost/api/"

# Tail logs for specific service
# Usage: make k8s-logs SERVICE=api
k8s-logs:
	@echo "Tailing logs for: $(SERVICE)"
	@kubectl logs -f -l app=$(SERVICE) -n mycelis --tail=50

# Port-forward services to localhost
# Port-forward services to localhost (Robust)
k8s-forward:
	@./scripts/dev/forward_ports.sh

# Stop port forwarding
k8s-stop-forward:
	@./scripts/dev/stop_ports.sh

# Restart all deployments
k8s-restart:
	@echo "Restarting all deployments..."
	@kubectl rollout restart deployment -n mycelis
	@echo "Waiting for rollouts..."
	@kubectl rollout status deployment/nats -n mycelis
	@kubectl rollout status deployment/postgres -n mycelis
	@kubectl rollout status deployment/api -n mycelis
	@kubectl rollout status deployment/ui -n mycelis

