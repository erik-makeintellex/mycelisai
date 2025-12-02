.DEFAULT_GOAL := help

KIND_CLUSTER ?= kind

.PHONY: help setup clean test-api runner

help:
	@echo "Mycelis Service Network - K8s Workflow"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Setup:"
	@echo "  setup         Install dependencies (uv, npm)"
	@echo "  k8s-up        Create Kind cluster with Ingress"
	@echo "  k8s-init      Build, load images, deploy all services"
	@echo ""
	@echo "Development:"
	@echo "  k8s-build     Build Docker images (API, UI)"
	@echo "  k8s-load      Load images into Kind cluster"
	@echo "  k8s-apply     Apply all k8s manifests"
	@echo "  k8s-dev       Quick rebuild & restart (SERVICE=api|ui)"
	@echo "  k8s-forward   Port-forward services to localhost"
	@echo ""
	@echo "Testing:"
	@echo "  test-api      Run backend tests (pytest)"
	@echo "  k8s-status    Check all services status"
	@echo "  k8s-logs      Tail logs (SERVICE=api|ui|nats|postgres)"
	@echo ""
	@echo "Agent Runtime:"
	@echo "  runner        Run agent runner locally"
	@echo ""
	@echo "Maintenance:"
	@echo "  k8s-restart   Restart all deployments"
	@echo "  k8s-reset     Delete and recreate cluster"
	@echo "  k8s-down      Delete Kind cluster"
	@echo "  clean         Remove build artifacts"

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
k8s-forward:
	@echo "Port-forwarding services..."
	@echo "NATS: 4222, Postgres: 5432, API: 8000, UI: 3000"
	@trap 'kill %1 %2 %3 %4' SIGINT; \
	kubectl port-forward svc/nats 4222:4222 -n mycelis & \
	kubectl port-forward svc/postgres 5432:5432 -n mycelis & \
	kubectl port-forward svc/api 8000:8000 -n mycelis & \
	kubectl port-forward svc/ui 3000:3000 -n mycelis & \
	wait

# Restart all deployments
k8s-restart:
	@echo "Restarting all deployments..."
	@kubectl rollout restart deployment -n mycelis
	@echo "Waiting for rollouts..."
	@kubectl rollout status deployment/nats -n mycelis
	@kubectl rollout status deployment/postgres -n mycelis
	@kubectl rollout status deployment/api -n mycelis
	@kubectl rollout status deployment/ui -n mycelis

