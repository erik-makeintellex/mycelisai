#!/bin/bash

# Mycelis Development Startup Script

echo "Starting Mycelis Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
  echo "Error: Docker is not running."
  exit 1
fi

# Start Infrastructure
echo "Ensuring infrastructure is up..."
make infra

# Check for tmux
if command -v tmux &> /dev/null; then
    echo "Tmux detected. Starting session 'mycelis'..."
    
    # Create new session
    tmux new-session -d -s mycelis -n 'infra'
    
    # Window 0: Infra Logs
    tmux send-keys -t mycelis:0 'make logs' C-m
    
    # Window 1: API
    tmux new-window -t mycelis -n 'api'
    tmux send-keys -t mycelis:1 'export OLLAMA_BASE_URL=http://192.168.50.156:11434 && make api' C-m
    
    # Window 2: UI
    tmux new-window -t mycelis -n 'ui'
    tmux send-keys -t mycelis:2 'make ui' C-m
    
    # Window 3: Bridge
    tmux new-window -t mycelis -n 'bridge'
    tmux send-keys -t mycelis:3 'make run-bridge' C-m
    
    # Window 4: Runner
    tmux new-window -t mycelis -n 'runner'
    tmux send-keys -t mycelis:4 'export OLLAMA_BASE_URL=http://192.168.50.156:11434 && make runner' C-m

    # Window 5: Shell (for running agents)
    tmux new-window -t mycelis -n 'shell'
    tmux send-keys -t mycelis:5 'export OLLAMA_BASE_URL=http://192.168.50.156:11434 && echo "Run agents here: make run-agent NAME=a1"' C-m
    
    # Attach
    tmux attach-session -t mycelis
else
    echo "Tmux not found. Please install tmux for the best experience."
    echo "Alternatively, run the following in separate terminals:"
    echo "1. make logs"
    echo "2. make api"
    echo "3. make ui"
    echo "4. make run-bridge"
    echo "5. make runner"
fi
