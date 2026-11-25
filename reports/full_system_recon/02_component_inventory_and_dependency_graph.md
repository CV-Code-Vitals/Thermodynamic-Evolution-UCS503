# Phase 0.2 — Component Inventory & Dependency Graph

**Date:** 2026-11-24  
**Auditor:** Principal Systems Architect & Independent Release Qualification Authority  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Commit:** `4b992caf951aef397ec9ef043bb9f1e3755b0cae`

---

## 1. Complete Component Inventory

### 1.1 Core Executable Components & Services

| Component ID | Subsystem Name | Language / Runtime | Primary Entry Point | Build Command | Default Ports / Interfaces | Primary Dependencies | Storage & State Boundaries |
|---|---|---|---|---|---|---|---|
| **COMP-01** | `go-backend` | Go 1.21+ | `go-backend/main.go` | `go build ./...` | HTTP `8080` (or `PORT`) | `thermodynamic-ast-engine`, `ucs503-sandbox-runner` | Local filesystem `uploads/`, ephemeral sandbox directories |
| **COMP-02** | `api` | Go 1.21+ | `api/main.go` | `go build ./...` | HTTP `8081` (local) / Lambda handler | AWS SDK v2, DynamoDB / `local_store.go`, S3 | SQLite/JSON local store, S3 bucket `archives/`, DynamoDB |
| **COMP-03** | `thermodynamic-ast-engine` | Rust (stable) | `thermodynamic-ast-engine/src/main.rs` | `cargo build --release` | CLI binary (`thermodynamic-ast-engine.exe`) | `tree-sitter` (Go, Python, JS, TS, Rust, C, Java, etc.) | Pure function CLI; input path via CLI -> JSON stdout |
| **COMP-04** | `admin-portal` | JavaScript / React 18 / Vite | `admin-portal/src/main.jsx` | `npm run build` | HTTP `5173` (dev) / static bundle in `dist/` | React, Vite, Lucide React, Axios | Browser localStorage/sessionStorage, memory state |
| **COMP-05** | `graph-visualizer` | JavaScript / Three.js / D3 | `graph-visualizer/graph.mjs` | Static vanilla JS | Embedded iframe / static files in `graph-visualizer/` | Three.js, OrbitControls, D3-Force-3D | Read-only JSON payload from DOM / URL query |
| **COMP-06** | `stream-proxy` | Node.js (v22+) | `infra/analysis/stream-proxy/index.mjs` | `npm test` | HTTP `3001` (local) / Lambda function URL | Native `node:http`, `node:https` | Stateless proxy, streaming SSE buffer |
| **COMP-07** | `analysis-authorizer` | Go 1.21+ | `analysis-authorizer/main.go` | `go build ./...` | Lambda API Gateway Custom Authorizer | AWS Lambda Go SDK | Stateless HMAC / token validation |
| **COMP-08** | `source-stager` | Go 1.21+ | `source-stager/main.go` | `go build ./...` | Lambda handler / CLI | AWS SDK v2 (S3) | S3 staging bucket -> Lambda `/tmp` workspace |
| **COMP-09** | `sandbox-runner` | Docker Container | `docker/sandbox-runner/Dockerfile` | `docker build -t ucs503-sandbox-runner` | Isolated container CLI (`go vet`, `go test`) | Debian slim, Go 1.21 toolchain | Ephemeral container RAM (`tmpfs /tmp`), read-only root |
| **COMP-10** | `serve.js` | Node.js | `serve.js` | `node serve.js` | HTTP `8000` / `3000` | Native `node:http`, `node:fs` | Reverse proxy `/api/*` -> Backend, static files |
| **COMP-11** | `parser` (prototype) | Rust | `parser/src/main.rs` | `cargo check` | Standalone CLI prototype | Early phase AST prototype (inactive in production) | None |

---

## 2. Component Dependency & Data Flow Graph

```mermaid
flowchart TD
    subgraph ClientTier ["Frontend & Presentation Tier"]
        U["End User Browser"]
        AP["admin-portal (React 18 / Vite)"]
        GV["graph-visualizer (Three.js / D3)"]
        U -->|Interacts| AP
        AP -->|Renders 3D View| GV
    end

    subgraph GatewayTier ["Gateway & Proxy Tier"]
        SJ["serve.js (Static Dev Proxy :8000)"]
        SP["stream-proxy (Node.js SSE Proxy :3001)"]
        AUTH["analysis-authorizer (Lambda / Go)"]
        NGX["nginx (Optional Reverse Proxy :80/:443)"]
        AP -->|HTTP Requests| SJ
        AP -->|SSE Events| SP
        AP -->|Auth Tokens| AUTH
    end

    subgraph ServiceTier ["Core Backend Services"]
        API["api (Archive & Accounts Service :8081)"]
        GB["go-backend (Analysis & Repair Engine :8080)"]
        SS["source-stager (S3 Unpacker Lambda)"]
        SJ -->|Proxy /api/v1/*| API
        SJ -->|Proxy /api/scan, /api/remediation| GB
        SP -->|SSE Pipe /api/evolve| GB
        API -->|Proxy Analysis Requests| GB
        SS -->|Prepares Staged Repo| GB
    end

    subgraph EngineTier ["Engines & Isolation Sandbox"]
        TAE["thermodynamic-ast-engine (Rust Binary)"]
        DOCK["ucs503-sandbox-runner (Docker Container)"]
        GB -->|Executes CLI AST Scan| TAE
        GB -->|Evaluates Oracles (--network none)| DOCK
    end

    subgraph StorageTier ["Storage & State Tier"]
        LOCAL_WS["Local Filesystem Workspaces (uploads/)"]
        S3["AWS S3 (Archives & Deliverables)"]
        DDB["Amazon DynamoDB / Local JSON Store"]
        GB -->|Workspace Staging| LOCAL_WS
        API -->|Presigned Forms & Uploads| S3
        API -->|Account Sessions & Ledger| DDB
        DOCK -->|Mounts :ro minimal cache & ws :rw| LOCAL_WS
    end
```

---

## 3. Communication Protocols & Trust Boundaries

### 3.1 Network Protocols
- **Frontend to API / Gateway:** HTTP/1.1 and HTTP/2 REST APIs with JSON payloads (`application/json`).
- **Evolution Streaming:** Server-Sent Events (`text/event-stream`) via `stream-proxy` or direct backend streaming.
- **Backend to AST Engine:** Local subprocess execution via `os/exec` with arguments and structured JSON output over stdout.
- **Backend to Sandbox Runner:** Docker CLI / daemon communication via Unix socket / Windows pipe executing hardened container commands (`run --rm --network none --read-only`).

### 3.2 Security Boundaries
1. **Public Ingress Boundary:**
   - Handled by `analysis-authorizer` or `api/auth.go` session bearer tokens.
   - Public repository URLs are validated against safe URL schemes (`https://github.com/...`) and safe archive extensions (`.zip`).
2. **Analysis Staging Boundary:**
   - Handled by `source-stager` and `workspace.go`.
   - Protects against Zip Slip, path traversal (`../`), symlink escapes, and oversized archives.
3. **Execution Sandbox Boundary:**
   - Dynamic tests from untrusted repositories are strictly forbidden from running on the host.
   - Container executes under non-root (`10001:10001`), with all Linux capabilities dropped (`--cap-drop ALL`), disabled networking (`--network none`), read-only root, bounded CPU/RAM/PIDs, and read-only minimal Go module cache mounting.
4. **Approval Integrity Boundary:**
   - HMAC-SHA256 signed `ApprovalToken` required for patch application and verification execution. Single-use and bound to exact workspace file hashes.
