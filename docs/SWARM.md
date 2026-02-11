# Swarm Intelligence Architecture (Phase 6)

## Overview

The **Swarm Intelligence** layer transforms Mycelis from a passive tool into an active, intent-driven system. It introduces a biological hierarchy for managing agents, optimizing communication, and ensuring security.

## Core Components

### 1. Soma (The Executive)

* **Role**: The "Conscious" center of the system.
* **Responsibility**:
  * Acts as the User Proxy.
  * Receives high-level intents from the **Global Bus**.
  * Maintains the "User Context".
* **Security**: All ingress to Soma acts through the **Guard**, which validates subject allowlists and payload sizes.

### 2. Axon (The Messenger)

* **Role**: Soma's "Subconscious" Assistant.
* **Responsibility**:
  * Optimizes and routes signals.
  * Translates "Intent" into specific "Team Commands".
  * Manages the **Tri-Level Bus Topology**.

### 3. Teams (Neural Clusters)

Teams are the functional units of the swarm.

* **Action Cores**:
  * Focus: Logic, Computation, Content Generation.
  * Examples: *Genesis Core* (System Ops), *Research Core*.
* **Expression Cores**:
  * Focus: Output, Visualization, Hardware Actuation.
  * Examples: *Telemetry Core* (UI), *Voice Core* (TTS).

## Bus Topology

Communication happens on NATS (Internal Network Only):

1. `swarm.global.*`: Public, Guarded, High-Level.
    * `swarm.global.input.gui`: User commands from Frontend.
    * `swarm.global.input.sensor`: Hardware telemetry.
2. `swarm.team.<id>.internal.*`: Private, Team-Only chatter.
3. `swarm.team.<id>.signal.*`: Public outputs from a team.

## Swarm Architect (UI)

The **Swarm Architect** (`/architect`) is a visual workspace for designing these clusters.

* **Team Builder**: Drag-and-drop agents into teams.
* **Knowledge Binding**: Attach RAG contexts or File Watchers to specific teams.

## Configuration

Teams are defined in `core/config/teams/*.yaml`:

```yaml
id: genesis-core
type: action
members:
  - id: architect
inputs:
  - swarm.global.input.cli.command
deliveries:
  - swarm.team.genesis.signal.status
```
