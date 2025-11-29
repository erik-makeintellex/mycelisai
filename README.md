# Mycelis Service Network

Mycelis is a decentralized AI agent orchestration platform designed for high-speed data ingestion, hybrid execution, and visual orchestration.

## Prerequisites

- **Python**: 3.10+
- **Node.js**: 18+
- **Docker**: For running infrastructure (NATS, PostgreSQL)
- **uv**: Python package manager (`pip install uv`)

## Getting Started

The project uses a `Makefile` to manage common tasks.

### 1. Setup Dependencies
Install Python and Node.js dependencies:
```bash
make setup
```

### 2. Start Infrastructure
Start NATS and PostgreSQL containers:
```bash
make infra
```

### 3. Run Application
You can run the API and UI in separate terminals:

**Terminal 1 (API):**
```bash
make api
```

**Terminal 2 (UI):**
```bash
make ui
```

Access the UI at [http://localhost:3000](http://localhost:3000).
Access the API docs at [http://localhost:8000/docs](http://localhost:8000/docs).

## Management Commands

- `make logs`: View infrastructure logs.
- `make stop`: Stop infrastructure containers.
- `make stop-apps`: Stop running API and UI processes.
- `make shutdown`: Stop everything (infra + apps).
- `make clean`: Remove artifacts and dependencies.

Run `make help` for a full list of commands.

## User Guide

### 1. Agents
Agents are the core workers of the network. Each agent has:
- **Name**: Unique identifier.
- **Backend Service**: The LLM provider (e.g., OpenAI, Ollama) it uses.
- **Capabilities**: Tags describing what the agent can do (e.g., "coding", "research").
- **Prompt Config**: Instructions that define the agent's persona and behavior.

**To Create an Agent:**
1. Navigate to the **Agents** page.
2. Click **Create Agent**.
3. Fill in the details and select a backend model.

### 2. Teams
Teams are groups of agents working together.
- **Channels**: NATS subjects the team listens to or publishes to (e.g., `data.ingest`, `alerts`).
- **Resource Access**: Key-value pairs defining what external resources the team can access (e.g., `db:read`, `s3:write`).
- **Inter-Comm Channel**: A dedicated channel (`team.{id}.chat`) for agents to talk to each other.

**To Create a Team:**
1. Navigate to the **Teams** page.
2. Click **Create Team**.
3. Select agents and define channels/resources.

**To Manage a Team:**
- **Add Agent**: Click "+ Add Agent" in the team view.
- **Remove Agent**: Hover over an agent and click "×".
- **Edit Team**: Click "Edit Team" to update name, description, channels, or resources.

### 3. Models
Register LLM backends (local or cloud) for agents to use.
- **Provider**: The service provider (e.g., `openai`, `anthropic`, `ollama`).
- **Context Window**: Maximum tokens the model can handle.

## Documentation

- [Product Vision](docs/product_vision.md): High-level vision and target usage scenarios.
- [Roadmap](docs/roadmap.md): Planned features and expansion phases.
- [Implementation Plan](docs/implementation_plan.md): Technical details and architecture.

