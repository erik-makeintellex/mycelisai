#!/bin/bash

# Mycelis Development Startup Script

echo "Starting Mycelis Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
  echo "Error: Docker is not running."
  exit 1
fi

# Start Infrastructure (Backend in Kind)
echo "Ensuring backend infrastructure is up..."
make k8s-backend

# Check for tmux
if command -v tmux &> /dev/null; then
    echo "Tmux detected. Starting session 'mycelis'..."
    
    # Create new session
    tmux new-session -d -s mycelis -n 'infra'
    
    # Window 0: Infra Logs (NATS)
    tmux send-keys -t mycelis:0 'make k8s-logs SERVICE=nats' C-m
    
    # Window 1: API (Logs from K8s)
    tmux new-window -t mycelis -n 'api'
    tmux send-keys -t mycelis:1 'make k8s-logs SERVICE=api' C-m
    
    # Window 2: UI (Local)
    tmux new-window -t mycelis -n 'ui'
    tmux send-keys -t mycelis:2 'cd ui && npm run dev' C-m
    
    # Window 3: Bridge (Logs from K8s)
    tmux new-window -t mycelis -n 'bridge'
    tmux send-keys -t mycelis:3 'make k8s-logs SERVICE=mcp-bridge' C-m
    
    # Window 4: Runner (Local)
    tmux new-window -t mycelis -n 'runner'
    tmux send-keys -t mycelis:4 'export OLLAMA_BASE_URL=http://192.168.50.156:11434 && uv run python runner/main.py' C-m

    # Window 5: Shell (for running agents/tools)
    tmux new-window -t mycelis -n 'shell'
    tmux send-keys -t mycelis:5 'echo "Environment: Backend in Kind, UI/Agents Local"' C-m
    
    # Attach
    tmux attach-session -t mycelis
else
    echo "Tmux not found. Please install tmux for the best experience."
    echo "Hybrid Setup Ready:"
    echo "1. Backend (API, NATS, Postgres) running in Kind."
    echo "2. Run UI locally: 'cd ui && npm run dev'"
    echo "3. Run Agents locally: 'uv run python runner/main.py'"
fi
