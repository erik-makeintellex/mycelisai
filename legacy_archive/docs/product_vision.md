# Mycelis Service Network - Product Vision & Target Usage

## 1. The Core Philosophy
Mycelis is designed for **Autonomous Agent Orchestration** in high-throughput, real-time environments. Unlike traditional workflow automation tools (e.g., n8n, Zapier) that execute deterministic, linear sequences ("If X happens, do Y"), Mycelis enables **probabilistic, context-aware decision making** by teams of AI agents.

**Mycelis is the "Brain"; n8n is the "Nervous System".**
*   **n8n/Zapier**: Excellent for connecting APIs and moving data (e.g., "When a row is added to Sheets, send an email").
*   **Mycelis**: Excellent for understanding complex events and deciding *what* to do (e.g., "Sensor temp is rising, but it's a hot day and the machine is idle. Should I alert a human or just log it?").

## 2. Target User
*   **AI Engineers**: Who need to deploy multi-agent systems that share memory and context.
*   **IoT Architects**: Who deal with high-velocity sensor data (MQTT/NATS) that overwhelms standard HTTP-based workflow tools.
*   **Enterprise Developers**: Building "Digital Workers" that need to collaborate (e.g., a Coder agent working with a Reviewer agent).

## 3. Target Usage Scenarios

### A. Intelligent IoT Response (The "Smart Factory")
*   **Scenario**: A vibration sensor on a manufacturing arm reports a 5% anomaly.
*   **Mycelis Approach (Swarm Reaction)**:
    1.  **Ingest Agent** detects the anomaly stream.
    2.  **Diagnostic Agent** analyzes the waveform.
    3.  **Safety Agent** immediately triggers a "Slow Down" command to the motor (Direct Actuation via MQTT).
    4.  **Maintenance Agent** schedules a checkup.
    *   *Key Differentiator*: Multiple agents react simultaneously (Swarm) to a single event, coordinating immediate safety actions *and* long-term resolution.

### B. Multi-Agent Research & Development
*   **Scenario**: "Research the latest battery tech and draft a spec sheet."
*   **Mycelis Approach**:
    1.  **Manager Agent** breaks down the task.
    2.  **Researcher Agent** browses the web (using MCP tools).
    3.  **Writer Agent** drafts the content.
    4.  **Reviewer Agent** critiques it against internal standards.
    5.  Agents communicate via the `team.chat` channel until the task is done.

## 4. Integration with Existing Services (The "Hybrid" Model)
Mycelis handles both **Direct Actuation** and **Delegated Workflows**.

*   **Direct Actuation (Real-time)**: For low-latency or complex control loops (e.g., robotics, safety systems), Mycelis Agents use **MCP** to talk directly to hardware (MQTT, Serial, API).
*   **Delegated Workflows (Business Logic)**: For standard business processes (e.g., "Send an Invoice"), Mycelis can trigger external tools like n8n.

## 5. Summary of Differentiation
| Feature | n8n / Workflow Tools | Mycelis Service Network |
| :--- | :--- | :--- |
| **Core Logic** | Deterministic (If/Else) | Probabilistic (LLM Reasoning) |
| **Data Flow** | HTTP / Webhooks (Request/Response) | NATS JetStream (Streaming / PubSub) |
| **Actuation** | API Calls / Webhooks | **Direct Swarm Control** (MCP) & API Calls |
| **State** | Stateless execution of a flow | Stateful Agents with Memory & Context |
| **Complexity** | Linear or DAG flows | Cyclic, conversational, multi-agent loops |
| **Primary Use** | Automating repetitive tasks | **Orchestrating intelligent, reactive swarms** |
