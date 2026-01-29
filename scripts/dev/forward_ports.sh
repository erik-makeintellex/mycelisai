#!/bin/bash
set -e

# Define Ports and Services
# Format: "PORT:SERVICE_NAME:K8S_SERVICE:K8S_PORT"
SERVICES=(
    "4222:NATS:nats:4222"
    "5432:Postgres:postgres:5432"
)

NAMESPACE="mycelis"

check_port() {
    local port=$1
    if nc -z localhost $port 2>/dev/null; then
        return 0 # Port is listening
    else
        return 1 # Port is free
    fi
}

check_service_health() {
    local port=$1
    local name=$2
    echo "  > Verifying connectivity to $name on port $port..."
    if nc -zv localhost $port 2>/dev/null; then
        echo "    [OK] $name is reachable."
        return 0
    else
        echo "    [FAIL] Port $port is open but connection failed."
        return 1
    fi
}

start_forward() {
    local port=$1
    local name=$2
    local svc=$3
    local svc_port=$4

    echo "  > Starting port-forward for $name ($port -> $svc:$svc_port)..."
    kubectl port-forward --address 0.0.0.0 svc/$svc $port:$svc_port -n $NAMESPACE > /dev/null 2>&1 &
    local pid=$!
    echo "    Started with PID $pid"
    
    # Wait for it to come up
    local text_red="\033[31m"
    local text_ids="\033[0m"
    for i in {1..10}; do
        if check_port $port; then
             return 0
        fi
        sleep 0.5
    done
    
    echo "    ${text_red}[ERROR] Failed to start port-forward for $name${text_ids}"
    return 1
}

echo "=== Robust Port Forwarding ==="

for entry in "${SERVICES[@]}"; do
    IFS=":" read -r port name svc svc_port <<< "$entry"
    
    echo "Checking $name (Port $port)..."
    
    if check_port $port; then
        echo "  Port $port is already in use."
        if check_service_health $port $name; then
             echo "  Service is healthy. Skipping forward."
        else
             echo "  [WARNING] Port is in use but service is unresponsive. It might be a stale process."
             echo "  Run 'make k8s-stop-forward' to clean up."
             # We do not confirm fail here to allow partial success, but user is warned.
        fi
    else
        start_forward $port $name $svc $svc_port
        
        # Verify post-start
        sleep 1
        if check_service_health $port $name; then
             echo "  [SUCCESS] $name is now forwarded."
        else
             echo "  [ERROR] $name forwarded but unreachable."
        fi
    fi
    echo ""
done

echo "Forwarding complete. Processes running in background."
# Keep script running if arguments say so, otherwise exit (processes are backgrounded)
# If invoked by Make, we usually want to wait or just exit. 
# Kubectl port-forward background jobs will die if parent shell exits usually? 
# No, mostly if we use & they attach to init or stay if we don't disown. 
# But usually Make waits.
# The user's original Makefile used `wait` at the end.
# If we exit here, background jobs might be killed.
# So we should wait if we started anything? 
# But we might have skipped some.
# Better strategy: logic above ensures they are running. 
# If we want to keep them alive attached to this terminal, we need to wait on PIDs.
# But since we might verify existing ones, we don't own those PIDs.

# If we are in "setup mode", we just want them running.
# The user's workflow: `make k8s-forward` -> keeps terminal open.
# So we should block here. 

echo "Monitoring ports (Ctrl+C to stop)..."
while true; do
   sleep 5 
   # Optional: Periodic health check or just sleep
done
