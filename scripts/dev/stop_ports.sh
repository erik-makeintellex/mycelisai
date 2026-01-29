#!/bin/bash

echo "=== Stopping Port Forwards ==="

# Kill by looking for matching command patterns
# Pattern: "kubectl port-forward ... svc/nats ... -n mycelis"

PIDS=$(ps aux | grep "kubectl port-forward" | grep "svc/nats\|svc/postgres" | grep -v grep | awk '{print $2}')

if [ -z "$PIDS" ]; then
    echo "No port-forward processes found."
else
    echo "Killing PIDs: $PIDS"
    kill $PIDS
    sleep 1
    # Force kill if needed
    kill -9 $PIDS 2>/dev/null || true
    echo "Stopped."
fi
