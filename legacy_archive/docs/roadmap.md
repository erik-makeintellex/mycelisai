# Mycelis Service Network - Expansion Roadmap

This document outlines the strategic roadmap for the Mycelis Service Network, moving from core feature completion to advanced orchestration and enterprise-grade capabilities.

## Phase 1: Core Feature Completion (Immediate)
**Goal**: Make the system fully functional for end-to-end agent orchestration.

- [ ] **Dynamic Channel Configuration**
    - **Objective**: Allow users to select specific NATS channels for agent inputs/outputs via the UI.
    - **Impact**: Enables flexible data flow wiring between agents without code changes.
- [ ] **Credential Management System**
    - **Objective**: Securely store and retrieve API keys (OpenAI, Anthropic, etc.) in PostgreSQL (encrypted).
    - **Impact**: Removes hardcoded keys and enables multi-tenant usage.
- [ ] **Automated Infrastructure Provisioning**
    - **Objective**: Automatically create NATS streams and consumers when teams or channels are defined in the UI.
    - **Impact**: Reduces manual setup and ensures infrastructure matches configuration.
- [ ] **Agent "Brain" Activation**
    - **Objective**: Connect `BaseAgent` to actual LLM backends using configured prompts and credentials.
    - **Impact**: Agents become operational and can process messages.

## Phase 2: Enhanced Usability & Observability (Short-term)
**Goal**: Improve the developer experience (DX) and system visibility.

- [ ] **Visual Team Builder**
    - **Objective**: Drag-and-drop interface for composing teams and visually connecting agent inputs/outputs (node-based editor).
    - **Impact**: Makes complex architectures easier to design and understand.
- [ ] **Interactive Chat Console**
    - **Objective**: A direct chat interface within the Team View to interact with agents or the team broadcast channel.
    - **Impact**: Facilitates debugging and human-in-the-loop workflows.
- [ ] **Advanced Observability**
    - **Objective**: Structured logging with filtering, and real-time metrics (tokens/sec, latency, cost) per agent.
    - **Impact**: Critical for performance tuning and cost management.
- [ ] **Agent Template Library**
    - **Objective**: Pre-configured agent templates (e.g., "Researcher", "Coder", "Reviewer") for quick start.
    - **Impact**: Accelerates user onboarding.

## Phase 3: Security & Scalability (Medium-term)
**Goal**: Prepare the system for production and multi-user environments.

- [ ] **Authentication & Authorization (AuthN/AuthZ)**
    - **Objective**: Implement user login (OAuth/OIDC) and Role-Based Access Control (RBAC) for teams and resources.
    - **Impact**: Secures the platform for multi-user organizations.
- [ ] **Containerization & Orchestration**
    - **Objective**: Dockerfiles for dynamic agent generation and Helm charts for Kubernetes deployment.
    - **Impact**: Enables scalable deployment across clusters.
- [ ] **Secure Communication**
    - **Objective**: mTLS for inter-agent communication and payload encryption.
    - **Impact**: Ensures data privacy and integrity.

## Phase 4: Advanced Intelligence (Long-term)
**Goal**: Enable complex, autonomous, and stateful agent behaviors.

- [ ] **Workflow Engine (DAGs)**
    - **Objective**: Define complex, multi-step workflows with conditional logic and parallel execution.
    - **Impact**: Supports sophisticated business processes.
- [ ] **Long-term Memory (RAG)**
    - **Objective**: Integration with vector databases (e.g., pgvector, Qdrant) for persistent semantic memory.
    - **Impact**: Agents retain context across sessions and vast amounts of data.
- [ ] **Tool Registry (MCP)**
    - **Objective**: Full implementation of the Model Context Protocol for dynamic tool discovery and usage.
    - **Impact**: Agents can dynamically discover and use external tools and APIs.
