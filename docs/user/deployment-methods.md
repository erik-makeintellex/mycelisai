# Deployment Method Selection
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Use this page to choose the deployment path that matches the environment you actually plan to run.

## TOC

- [Start With The Target Environment](#start-with-the-target-environment)
- [Docker Compose Single Host](#docker-compose-single-host)
- [Local Kubernetes With k3d](#local-kubernetes-with-k3d)
- [Enterprise Self-Hosted Kubernetes](#enterprise-self-hosted-kubernetes)
- [Edge Or Small Node Deployments](#edge-or-small-node-deployments)
- [Developer Source Mode](#developer-source-mode)
- [AI Endpoint Rules](#ai-endpoint-rules)
- [Validation Before User Testing](#validation-before-user-testing)

## Start With The Target Environment

Choose the runtime by deployment target, not by whichever tool is already open.

| Target environment | Recommended path | Best fit |
| :-- | :-- | :-- |
| One machine, rapid iteration, founder demo, personal proof | Docker Compose | Fastest local full-stack development/proof runtime |
| Local Helm and Kubernetes validation in WSL/Linux | `k3d` | Best local proof for chart behavior before a real cluster |
| Customer-managed, enterprise-managed, or release target cluster | Enterprise self-hosted Kubernetes | Best fit for real cluster policy, ingress, registry, storage, and secret management |
| Edge node, Raspberry Pi style control node, small Linux box | Packaged binary or node-attached service | Lightweight host footprint with remote AI service support |
| Active code changes and UI/backend iteration | WSL worktree + Docker Compose first, source-mode fallback when needed | Best fit for implementation work, not for production-style deployment |

Quick rule:
- if the question is "how do I iterate quickly or prove the stack on one host?" use Docker Compose
- if the question is "how do I prove the Helm chart or cluster behavior locally?" use `k3d`
- if the question is "how will this be deployed for release or a customer-owned environment?" use the Helm chart for enterprise self-hosted Kubernetes
- if the question is "how do I place a lightweight node on small hardware?" use the packaged binary path and keep AI remote

Supported user access lanes:
- Windows Docker Desktop: run the rapid local Compose stack on Windows and open `http://localhost:3000` from the Windows browser on that same machine
- Windows + WSL Docker: run the rapid local Compose stack from WSL, then use the Windows browser as the first operator proof path through `http://localhost:3000`
- Kubernetes / Helm clustered deployment: run the verified chart on the target cluster and open the UI through the same ingress, hostname, or IP operators will really use remotely

## Docker Compose Single Host

Choose Docker Compose when you want rapid local development, same-machine proof, or a quick demo loop.

Typical fit:
- home-lab or personal owner proof
- demos and partner evaluation
- one machine running the full stack
- Windows Docker Desktop single-machine runtime with the Windows browser on the same host
- Windows + WSL Docker with the Windows browser as the first operator proof path
- Linux server bring-up where remote clients should use the published host/IP or ingress address
- WSL2, Linux, or macOS bring-up where local Kubernetes would add unnecessary overhead

Recommended path:

Windows Docker Desktop:

```powershell
Copy-Item .env.compose.example .env.compose
uv run inv auth.dev-key
uv run inv auth.posture
uv run inv install
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
```

WSL2/Linux/macOS:

```bash
cp .env.compose.example .env.compose
uv run inv auth.dev-key
uv run inv auth.posture
uv run inv install
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
```

Linux server/self-hosted release:

```bash
cp .env.compose.example .env.compose
uv run inv auth.dev-key
uv run inv auth.posture
uv run inv install
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
```

Use this path first for rapid iteration, not as the target clustered deployment contract.
Use the same operator-facing address you will really support after delivery:
- Windows Docker Desktop on the same machine: `http://localhost:3000`
- Windows browser against a WSL-hosted stack: start with `http://localhost:3000`, then prove the host/IP path for second-machine access
- clustered release: use the published Kubernetes ingress/hostname/IP from the operator machine

## Local Kubernetes With k3d

Choose `k3d` when you need local Kubernetes behavior, Helm validation, or cluster-shaped testing in Docker.

Typical fit:
- validating Helm values or readiness behavior locally
- proving Kubernetes networking, PVC, ingress, or rollout expectations before a remote cluster
- development in WSL/Linux where Docker is available and you want chart parity

Recommended path:

```bash
cp .env.example .env
# set MYCELIS_K8S_TEXT_ENDPOINT to a reachable AI service when needed
export MYCELIS_K8S_VALUES_FILE=charts/mycelis-core/values-k3d.yaml
uv run inv k8s.up
uv run inv k8s.status
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
```

Notes:
- `k3d` is the preferred local Kubernetes backend
- use `MYCELIS_K8S_BACKEND=kind` only when you intentionally need the older local fallback
- this is a local validation lane for the clustered deployment contract, not the recommended default just to get a single machine running

## Enterprise Self-Hosted Kubernetes

Choose enterprise self-hosted Kubernetes when the target environment is a customer or enterprise cluster with its own registry, secrets, storage classes, ingress, and policy controls.

Typical fit:
- enterprise cluster operations
- customer-managed Kubernetes
- production-style internal cluster rollout
- environments where platform teams manage ingress, certificates, storage, and image pull policy

Use the Helm chart as the canonical deployment contract.
Treat `k3d` as the local preflight lane for that chart, not as the production target.

Promoted preset files:
- `charts/mycelis-core/values-enterprise.yaml`
- `charts/mycelis-core/values-enterprise-windows-ai.yaml`

Expect to provide:
- image registry and pull credentials
- secret management for API keys and runtime credentials
- ingress, DNS, and TLS decisions
- storage class and PVC sizing
- scheduling policy such as node selectors, tolerations, affinity, and spread constraints
- explicit AI endpoint wiring for text and optional media services

Example preflight command:

```powershell
$env:MYCELIS_K8S_VALUES_FILE="charts/mycelis-core/values-enterprise-windows-ai.yaml"
$env:MYCELIS_K8S_TEXT_ENDPOINT="http://<windows-ai-host>:11434/v1"
uv run inv k8s.up
```

The Windows-AI preset is intentionally not self-filling.
If `MYCELIS_K8S_TEXT_ENDPOINT` is missing, `k8s.deploy` and `k8s.up` fail closed instead of pretending the placeholder endpoint is deployable.

Package-verification shortcut:

```bash
uv run inv k8s.deploy --verify-package --values-file=charts/mycelis-core/values-enterprise.yaml --release-label=enterprise
```

Use that verification path when you need lint/render/package proof and artifact output under `dist/helm/` before touching a live cluster.

Open-standard chart gate:

```bash
uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml
```

That gate keeps the deployment contract on standard Kubernetes/Helm surfaces: Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy, startup/readiness/liveness probes, non-root security context, image pull/digest posture, and cluster-managed output storage.

## Edge Or Small Node Deployments

Choose a packaged binary or node-attached service when the target host is a small Linux node, Raspberry Pi style device, or lightweight cluster member.

Typical fit:
- local control node on small hardware
- branch office or lab node
- helper service attached to a larger customer cluster

Guidance:
- keep the runtime lightweight
- prefer remote Ollama or another reachable AI service instead of assuming the edge node should host heavyweight inference
- build for the target architecture instead of copying desktop-generated artifacts blindly

## Developer Source Mode

Use the source-run lifecycle path when you are changing code, debugging, or iterating on UI/backend behavior and specifically need repo-local source or Kubernetes validation.

Recommended path:

```bash
cp .env.example .env
uv run inv install
uv run inv k8s.up
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
```

Developer source mode is not a deployment method. It is the implementation lane for changing the product when Compose-backed development proof is not the right slice.

Windows host note:
- if the machine is Windows, prefer a WSL worktree plus Docker Compose for day-to-day code changes
- use the Windows browser against `http://localhost:3000` as the first operator-facing check for that WSL-hosted stack
- keep the Windows-native source path for explicit host-local Kubernetes validation or host-specific troubleshooting

## AI Endpoint Rules

Keep AI endpoints explicit in every runtime.

Use these patterns:
- Docker Compose: `MYCELIS_COMPOSE_OLLAMA_HOST=http://<windows-ai-host>:11434`
- Kubernetes: `MYCELIS_K8S_TEXT_ENDPOINT=http://<windows-ai-host>:11434/v1`
- optional media endpoint: `MYCELIS_K8S_MEDIA_ENDPOINT=http://<media-host>:8001/v1`

Rules:
- do not use `localhost` for container or cluster deployments when the model service is on another host
- when Ollama or another AI service runs on a Windows GPU host, point Compose or Kubernetes at the reachable Windows IP or hostname
- treat the Windows GPU host as the AI service endpoint, not as the place the whole runtime must live

## Validation Before User Testing

Before user testing or UI testing, prove the lane you picked:

- Docker Compose rapid development/proof: `uv run inv compose.status` and `uv run inv compose.health`
- local Kubernetes: `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-k3d.yaml`, `uv run inv k8s.status`, and `uv run inv lifecycle.health`
- enterprise self-hosted Kubernetes: run `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`, render/package the Helm chart with the real values, then prove ingress, readiness, storage, and the explicit AI endpoint on the target cluster
- Windows Docker Desktop: open the UI from the Windows browser on the same machine through `http://localhost:3000`, then confirm the runtime reaches the explicit Windows AI host instead of a loopback-only deployment assumption
- Windows browser against a WSL-hosted stack: start with `http://localhost:3000` on that same Windows machine, then use the host/IP path for second-machine or LAN proof
- Kubernetes clustered release: open the UI from the operator machine through the same ingress/host/IP/hostname that will be used remotely, not `localhost` on the server
- release gate: `uv run inv ci.release-preflight --lane=release` (recommended preset for runtime-posture + service-health + live-backend proof; the legacy flags remain available when you need a narrower/manual combination)
