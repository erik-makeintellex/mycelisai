# Implementation Plan - Mycelis Service Network

# Goal Description
Design and implement the Mycelis Service Network, a platform for orchestrating AI agent teams with diverse messaging needs (event streams, binary, text). The system will support custom agent development, flexible configuration via UI, and integration with external services.

## User Review Required
> [!IMPORTANT]
> **Messaging Infrastructure**: NATS (JetStream) is proposed as the core messaging backbone due to its support for streams and request-reply patterns. Confirm if this aligns with infrastructure preferences.

> [!NOTE]
> **UI Stack**: Next.js is proposed for the frontend to provide a modern, responsive interface for agent configuration and monitoring.

## Proposed Changes

### Project Structure
- Initialize a new Python project using `uv`.
- Structure:
  - `agents/`: Agent implementations and base classes.
  - `api/`: FastAPI application for internal/external APIs.
  - `ui/`: Next.js frontend application.
  - `shared/`: Shared schemas and utilities.
- **Project Management**:
  - `Makefile`: For managing project state (start, stop, clean, install).
  - **New**: `docker-compose.yml`: For orchestrating NATS and PostgreSQL.

### Data Layer
- **Technology**: PostgreSQL.
- **Purpose**: Persistent storage for agents, teams, models, and credentials.
- **Schemas**:
  - `agents`: Stores agent configuration and state.
  - `teams`: Stores team composition and channels.
  - `models`: Stores available AI models.
  - `credentials`: Encrypted storage for API keys and secrets (linked to agents/teams).

### Messaging Layer
- **Technology**: NATS JetStream.
- **Schemas**:
  - `EventMessage`: For sensor/video streams.
  - `BinaryMessage`: For raw binary data.
  - `TextMessage`: For standard text communication.
  - **New**: `AgentMessage`: For structured inter-agent communication (sender, recipient, content, intent).
- **Implementation**:
  - `MessageBroker` class to handle connection and pub/sub.

### Agent Framework
- **BaseAgent**: Abstract base class for all agents.
- **Configuration**:
  - JSON/YAML based configuration.
  - Attributes: `languages` (list), `prompt_config` (dict), `capabilities` (list).
  - **New**: `messaging_channels` (input/output topics).
  - **New**: `deployment_config` (replicas, host constraints).
  - **New**: `backend` (enum: openai, anthropic, ollama, etc).
  - **New**: `inter_agent_comm` (methods to publish/subscribe to team chat).
- **Discovery**: Mechanism for agents to register their presence and capabilities.
- **Model Registry**:
  - Schema for available models (name, provider, context_window, etc).
  - API to manage models.
- **Team Management**:
  - Schema for Teams (name, agents list, shared_context).
  - **New**: `channels` (list of associated channels).
  - **New**: `inter_comm_channel` (dedicated channel for team chat).
  - **New**: `ui/src/app/teams/page.tsx` - Master-detail view for teams.
    - Left sidebar: Scrollable list of teams.
    - Main area: Team details (Agents, Channels) + Live Log Console (SSE from `team.{id}.>`).
  - **Modify**: `ui/src/components/Navbar.tsx` to link to `/teams` instead of just `/teams/create`.
  - **New**: `resource_access` (policy for accessing other resources).
  - API to create/manage teams.
  - **New**: Auto-provision NATS streams for team channels on creation.
  - **New**: Auto-assign "admin" channel to new teams.
- **Logic Refinement**:
  - Implement `AgentManager` to handle registration and health checks.
  - Support "Cluster" mode (multiple instances of same agent).
  - **New**: Integrate `SQLAlchemy` or `Prisma` for DB access.
  - **New**: Implement credential management API (store/retrieve secure keys).
    - `POST /credentials`: Store a key (encrypted).
    - `GET /credentials/{owner_id}`: Retrieve keys for an agent/team.
    - `DELETE /credentials/{id}`: Revoke a key.

### API Layer
- **Internal API**:
  - `POST /agents/register`: Register a new agent/team.
  - `GET /agents`: List active agents.
  - `POST /config`: Update agent configuration.
- **External API**:
  - `POST /ingest/{channel}`: Ingest data into a specific channel.
  - `GET /stream/{channel}`: Subscribe to a data stream (WebSocket/SSE).

### User Interface
- **Tech Stack**: Next.js, Tailwind CSS.
- **Features**:
  - **Dashboard**: 
    - Real-time view of active agent teams.
    - **New**: Live event stream visualization using SSE.
    - **New**: Real-time "Events / Sec" metric.
  - **Configuration**: 
    - Form-based agent configuration with dropdowns for standard attributes (e.g., "Python", "JavaScript", "Go" for languages).
    - **New**: Dynamic form for `prompt_config`.
    - **New**: Channel configuration (inputs/outputs).
    - **New**: Backend selection dropdown.
  - **Model Management**:
    - Page to add/edit available models.
    - "Add Model" form.
  - **Team Management**:
    - Dedicated workflow for Team creation.
    - Dedicated workflow for Team creation.
    - Interface to select existing agents and add them to a team.
    - **New**: Input fields for `channels`, `inter_comm_channel`, and `resource_access`.
  - **Integration**: Views for managing API keys and external service webhooks.

  - **New**: **Service & Channel Management UI**:
    - **Service Registry**:
      - Page to register external services (IoT devices, APIs, Databases).
      - Schema: `Service` (name, type, config, status).
      - UI: Wizard to register a new service (e.g., "Add IoT Sensor" -> Topic -> Data Format).
    - **Channel Manager**:
      - Page to view active NATS streams and subjects.
      - UI: Create/Delete channels, view message rates/size.
      - Visual graph of Channel -> Agent connections.
    - **Visual Flow Builder** (Future):
      - Drag-and-drop interface to connect Services -> Channels -> Agents.

  - **New**: **MCP Actuator Integration** (Priority):
    - **Objective**: Enable agents to execute commands on external devices via Model Context Protocol.
    - **Default Implementation**: **Generic MQTT MCP Server**.
      - **Why**: MQTT is the standard for IoT. A generic MCP server can expose `publish(topic, payload)` as a tool to agents.
      - **Workflow**: Agent -> `call_tool("mqtt_publish", ...)` -> MCP Server -> MQTT Broker -> Device.
    - **Alternative**: **HTTP Request MCP Server** for API-based actuators.

## Verification Plan

### Automated Tests
- **Unit Tests**:
  - Test `BaseAgent` initialization and configuration loading.
  - Test message schema validation.
- **Integration Tests**:
  - Spin up local NATS container.
  - Verify message publication and consumption by a mock agent.
  - Test API endpoints using `pytest` and `TestClient`.

### Manual Verification
- **UI Walkthrough**:
  - Create a new agent team via the UI.
  - Verify that "Languages" dropdown is populated and selectable.
  - Send a test message via the External API and verify it appears in the agent's log/dashboard.
