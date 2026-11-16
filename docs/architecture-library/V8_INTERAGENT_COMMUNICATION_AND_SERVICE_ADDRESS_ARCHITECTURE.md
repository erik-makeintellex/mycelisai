# V8 Inter-Agent Communication And Service-Address Architecture
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-11-15
> Purpose: Document why distributed inter-agent communication and backend service addressing fail in conventional multi-agent systems and define the Mycelis V8 architecture that prevents those failure modes across Rancher Desktop, k3d, Compose, and Enterprise Kubernetes.

## Industry Research: Why Inter-Agent & Backend Service Addressing Fails in Other Frameworks

Conventional multi-agent frameworks treat agent-to-agent (A2A) coordination and backend infrastructure connections as ad-hoc HTTP calls or developer-local socket configurations. When promoted from single-process prototypes to distributed containers or Kubernetes clusters, five systemic failure modes emerge:

### 1. Point-to-Point HTTP/gRPC Mesh Sprawl & Ephemeral Pod IP Drift
Frameworks that dial peer agents via direct HTTP or gRPC endpoints (`http://agent-b:8080`) couple caller lifecycle to callee availability and network location:
- **Cascading Timeouts on Long LLM Turns:** Synchronous HTTP chains block upstream caller threads across multi-step reasoning or tool-use turns, causing gateway `504 Gateway Timeout` cascades.
- **Broken Callbacks Across Pod Restarts & Autoscaling:** When Kubernetes reschedules or autoscales a specialist pod, in-flight callbacks and static peer URLs break immediately.
- **`O(N^2)` Connection Configuration:** Every new specialist or team requires pairwise endpoint discovery, TLS provisioning, and retry policies across `N` agents.

### 2. Container & Pod Loopback Traps (`localhost` / `127.0.0.1` Split-Horizon)
Configurations that hardcode `localhost:11434`, `127.0.0.1:4222`, or `localhost:5432` succeed during bare-metal host execution but fail immediately when promoted into Kubernetes pods (Rancher Desktop `rancher-desktop`, `k3d`, or enterprise clusters):
- Inside a Kubernetes pod, `127.0.0.1` and `localhost` resolve strictly to the pod's isolated network namespace rather than the developer workstation or cluster service mesh.
- Backend dials fail with `ECONNREFUSED` unless service endpoints are explicitly promoted to cluster DNS names (`mycelis-core-nats`, `mycelis-core-postgresql`) or host-bridge addresses (`host.rancher-desktop.internal`, `host.k3d.internal`, or host LAN IP).

### 3. Zero-Trust Kubernetes `NetworkPolicy` Egress Blackholes
Production-grade Kubernetes charts enforce default-deny egress `NetworkPolicy` rules (`policyTypes: ["Ingress", "Egress"]`). Without explicit egress port allowances and CIDR rules:
- Core pods pass `/healthz` (`200 OK`) because liveness probes only inspect the local HTTP listener, while outbound dials to host-bridged AI engines (`:11434`, `:11435`), vLLM/media servers (`:8000`, `:8001`), external databases, or HTTP/SSE MCP gateways are silently dropped by the CNI firewall.

### 4. Unbounded Payload Bloat ("Telephone Game" & Bus Max-Payload Failures)
Passing raw multi-megabyte files, generated code bundles, HTML snapshots, or vector embeddings inline inside inter-agent messages causes two compounding failures:
- **Message Bus Drops (`ErrMaxPayload`):** High-speed brokers enforce strict frame limits (typically 1 MB to 8 MB in NATS), dropping oversized agent responses mid-run.
- **Context Window Bleed:** Re-serializing full artifact bodies across every hop in a Soma -> Council -> Team Lead -> Specialist chain rapidly exhausts LLM token budgets and corrupts intermediate formatting.

### 5. Uncoordinated Multi-Agent Shared-State Write Collisions
When parallel specialist agents write directly to shared workspace paths or vector memory without causal correlation (`run_id`) or file-ownership arbitration, concurrent workers silently overwrite each other's outputs, producing non-deterministic deliverables with no audit trail.

---

## Mycelis V8 Three-Plane Separation: A2A Bus vs. MCP Tools vs. Claim-Check Storage

Mycelis V8 eliminates these five failure modes by strictly separating runtime communication into three orthogonal planes:

```
+-----------------------------------------------------------------------------------+
| Plane 1: Location-Transparent Subject-Addressed A2A Bus (NATS JetStream)          |
|   swarm.team.{team_id}.internal.command | .signal.status | .signal.result         |
|   swarm.council.{agent_id}.request      | swarm.mission.events.{run_id}           |
+-----------------------------------------+-----------------------------------------+
                                          | Carries compact OutputRef + Attribution
                                          v
+-----------------------------------------------------------------------------------+
| Plane 2: Claim-Check Artifact & Proof Plane (OutputRef + OutputProofEnvelope)     |
|   storage_ref + SHA-256 checksum + path_boundary_status + readback_status         |
|   Durable bytes in PostgreSQL & /data/workspace PVC; collision arbitration        |
+-----------------------------------------+-----------------------------------------+
                                          | Executes governed external capabilities
                                          v
+-----------------------------------------------------------------------------------+
| Plane 3: Governed Capability & Backend Service Plane (ServiceAddressBook + MCP)   |
|   Explicit service address book (nats, postgres, text_ai, media_ai, search,       |
|   mcp_gateway, cors_origin) + governed MCP/internal tool invocation               |
+-----------------------------------------------------------------------------------+
```

### Plane 1 — Location-Transparent Subject-Addressed A2A Bus (NATS JetStream)
Agents never dial peer IP addresses or HTTP ports. All inter-agent coordination routes through canonical NATS subjects defined in `core/pkg/protocol/topics.go`:
- `swarm.team.{team_id}.internal.command` — directed team input and task dispatch
- `swarm.team.{team_id}.signal.status` — concise operator-readable status transitions
- `swarm.team.{team_id}.signal.result` — bounded execution outcomes and `OutputRef` envelopes
- `swarm.council.{agent_id}.request` — request-reply specialist consultations
- `swarm.mission.events.{run_id}` — run-linked persistent mission event fanout

Every governed product signal carries mandatory attribution metadata (`run_id`, `team_id`, `agent_id`, `source_kind`, `source_channel`, `payload_kind`, `timestamp`), enabling location-transparent queue groups, durable replay, and zero peer-IP coupling.

### Plane 2 — Claim-Check Artifact & Proof Plane (`OutputRef` + `OutputProofEnvelope`)
Inter-agent messages never carry raw multi-megabyte artifacts inline. Instead, NATS payloads carry compact `OutputRef` pointers (`storage_ref`, SHA-256 `checksum`, `path_boundary_status`, `readback_status`), while durable bytes reside in PostgreSQL and `/data/workspace` PVCs.
- **Write-Collision Governance:** In `core/internal/server/team_delivery_package_collision.go`, `DetectTeamPackageCollisions` inspects every multi-agent `DeliveryPackage`. If two or more agents write to the same normalized `FilePath` (or overwrite with differing content), verification posture automatically downgrades to `pending_review`, marks `OutputRef.Proof.PathBoundaryStatus = "collision_detected"` with `ReadbackStatus = "degraded"`, and emits a `V8.2` `OperationalLogContext` (`review_scope: "central_review"`, `audiences: ["soma", "meta_agentry", "team_lead"]`).

### Plane 3 — Governed Capability & Backend Service Plane (`ServiceAddressBook` + MCP)
External and backend services (`nats`, `postgres`, `text_ai`, `media_ai`, `search`, `mcp_gateway`, `cors_origin`) are resolved through an explicit, environment-promoted `ServiceAddressBook` (`core/internal/server/service_addresses.go`) rather than scattered hardcoded URLs.

---

## Canonical Backend Service-Address Contract (`ServiceAddressBook`)

`ResolveServiceAddressBook()` in `core/internal/server/service_addresses.go` resolves every backend dependency through a deterministic precedence chain and exposes live resolution posture via `GET /api/v1/services/status`.

### Service-Address Precedence Table

| Service Key | 1. Runtime Env Override (Highest) | 2. Legacy / Component Runtime Env | 3. Helm Chart Values (`charts/mycelis-core/values.yaml`) | 4. Invoke K8s Overrides (`ops/k8s.py`) | 5. Local Host Fallback |
| --- | --- | --- | --- | --- | --- |
| `nats` | `MYCELIS_SERVICE_NATS_ADDRESS` | `NATS_URL` | `.Values.serviceAddresses.natsUrl` | `MYCELIS_K8S_NATS_URL` | `nats://localhost:4222` |
| `postgres` | `MYCELIS_SERVICE_DB_ADDRESS` | `DB_HOST` + `DB_PORT` | `.Values.serviceAddresses.dbHost` + `.Values.serviceAddresses.dbPort` | `MYCELIS_K8S_DB_HOST` + `MYCELIS_K8S_DB_PORT` | `127.0.0.1:5432` |
| `text_ai` | `MYCELIS_SERVICE_TEXT_AI_ADDRESS` | `MYCELIS_PROVIDER_OLLAMA_ENDPOINT` | `.Values.ai.textEndpoint` | `MYCELIS_K8S_TEXT_ENDPOINT` | `http://localhost:11434/v1` |
| `media_ai` | `MYCELIS_SERVICE_MEDIA_AI_ADDRESS` | `MYCELIS_MEDIA_ENDPOINT` | `.Values.ai.mediaEndpoint` | `MYCELIS_K8S_MEDIA_ENDPOINT` | *(optional / disabled)* |
| `search` | `MYCELIS_SERVICE_SEARCH_ADDRESS` | `MYCELIS_SEARCH_LOCAL_API_ENDPOINT` / `MYCELIS_SEARCH_SEARXNG_ENDPOINT` | `.Values.search.localApiEndpoint` / `.Values.search.searxngEndpoint` | `MYCELIS_K8S_SEARCH_ENDPOINT` | `workspace_local` |
| `mcp_gateway` | `MYCELIS_SERVICE_MCP_GATEWAY_ADDRESS` | `MYCELIS_MCP_GATEWAY_ADDRESS` | `.Values.serviceAddresses.mcpGateway` | `MYCELIS_K8S_MCP_GATEWAY` | `in-process-local-first` |
| `cors_origin` | `MYCELIS_SERVICE_CORS_ORIGIN` | `MYCELIS_CORS_ORIGIN` | `.Values.serviceAddresses.corsOrigin` | `MYCELIS_K8S_CORS_ORIGIN` | `http://localhost:3000` |

### Cluster Loopback Detection & Live Status Visibility
- **`IsClusterRuntime()`:** Detects Kubernetes execution via `KUBERNETES_SERVICE_HOST` or `MYCELIS_RUNTIME_TOPOLOGY=kubernetes`.
- **`IsLoopbackServiceAddress(addr)`:** Parses URLs and `host:port` pairs to detect `localhost`, `127.0.0.1`, or `::1`.
- **Proactive Misconfiguration Warning:** When `IsClusterRuntime()` is true and any configured service resolves to a loopback address, `GET /api/v1/services/status` populates the service entry's `address` field alongside an explicit `warning` explaining that `localhost`/`127.0.0.1` resolves inside the pod network namespace rather than the host or cluster service mesh.

---

## Topology Reference: Rancher Desktop (`rancher-desktop`), `k3d`, Compose & Enterprise Kubernetes

| Aspect | Local Docker Compose | Local Rancher Desktop (`rancher-desktop` / `mother-brain` k3s) | Local `k3d` Validation | Enterprise Self-Hosted Kubernetes |
| --- | --- | --- | --- | --- |
| **Container / Image Daemon** | Docker / Rancher Moby | Shared Rancher Desktop Moby (`docker build` immediately visible to k3s) | `k3d image import` into k3d containerd | Private OCI Registry (`imagePullSecrets`) |
| **NATS Address (`nats`)** | `nats://nats:4222` (or host `:4222`) | `nats://mycelis-core-nats:4222` (or `.Values.serviceAddresses.natsUrl`) | `nats://mycelis-core-nats:4222` | External HA JetStream or cluster service via `.Values.serviceAddresses.natsUrl` |
| **PostgreSQL (`postgres`)** | `postgres:5432` (or host `:5432`) | `mycelis-core-postgresql:5432` (or `.Values.serviceAddresses.dbHost`) | `mycelis-core-postgresql:5432` | Managed pgvector DB or cluster service via `.Values.serviceAddresses.dbHost` |
| **Text AI (`text_ai`)** | `http://host.docker.internal:11434/v1` | `http://host.rancher-desktop.internal:11434/v1` or host LAN IP | `MYCELIS_K8S_TEXT_ENDPOINT` (`http://host.k3d.internal:11434/v1`) | Cluster vLLM/Ollama service or enterprise LLM gateway URL |
| **Media AI (`media_ai`)** | `http://host.docker.internal:8001/v1` | `http://host.rancher-desktop.internal:8001/v1` or host LAN IP | `MYCELIS_K8S_MEDIA_ENDPOINT` | Dedicated GPU inference service URL |
| **NetworkPolicy Egress** | Bridge network | `.Values.networkPolicy.allowedEgressPorts` (`4222, 5432, 8000, 8001, 8080, 11434, 11435`) | `.Values.networkPolicy.allowedEgressPorts` | Strict CIDR + port egress rules + `.Values.oidc` federation |

### Platform-Neutral Rancher Deployment Phase (`ops/k8s_rancher.py`)
- **Default Localhost Rancher Target:** `rancher_target_url()` defaults to `https://127.0.0.1:6443` (`localhost` API server, overridable via `MYCELIS_RANCHER_URL` or `MYCELIS_K8S_SERVER`) and `rancher_target_context()` defaults to `rancher-desktop` (overridable via `MYCELIS_K8S_CONTEXT` or `MYCELIS_RANCHER_CONTEXT`).
- **Cross-Platform Neutrality (Windows, WSL, Linux, macOS):** `_k8s_backend()` in `ops/k8s.py` requires only standard `kubectl` + `helm` talking to the target Rancher context (`rancher-desktop` at `https://127.0.0.1:6443`), without requiring OS-specific wrappers. When invoked inside WSL without a local `~/.kube/config`, `_ensure_shared_kubeconfig()` automatically binds `KUBECONFIG` to the shared host config so `uv run inv k8s.*` deploys to the same local `localhost:6443` Rancher cluster from any shell.
