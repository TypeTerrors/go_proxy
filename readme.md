# prx — Reverse Proxy Management Toolkit

A single binary toolkit to manage HTTP and gRPC reverse proxy records on Kubernetes. It provides:

- **Server**: runs in-cluster, listens on HTTP `:80` and gRPC `:50051`, routes incoming requests via dynamic `Ingress` and `Secret` resources.
- **HTTP API**: REST endpoints (`/api/prx`) to create, update, delete, list redirects using JWT authentication.
- **gRPC API**: `Reverse` service with `Add`, `Update`, `Delete`, `List` RPCs, secured by the same JWT.
- **CLI client**: one binary `prx` exposes subcommands:
  - `prx secret` (Bubble Tea UI): generates a new HMAC key for JWT signing.
  - `prx auth`   (Bubble Tea UI): generates a JWT for client authentication.
  - `prx add/update/delete/list`: gRPC client to configure proxies.

---

## Table of Contents

1. [Project Overview](#project-overview)  
2. [Architecture](#architecture)  
3. [Prerequisites](#prerequisites)  
4. [Setup & Configuration](#setup--configuration)  
5. [Running Locally](#running-locally)  
6. [Deployment](#deployment)  
7. [Usage](#usage)  
   - [HTTP API Examples](#http-api-examples)  
   - [gRPC CLI Examples](#grpc-cli-examples)  
8. [GitHub Workflow](#github-workflow)  
9. [License](#license)

---

## Project Overview

`prx` is designed to let operators dynamically manage reverse proxy rules in a Kubernetes cluster. It:

- Automates creation of TLS `Secret` and `Ingress` per hostname.
- Stores redirect-to URLs in a `ConfigMap` for fallback and listing.
- Secures all configuration operations behind JWT-based auth.
- Offers both REST and gRPC interfaces.
- Ships a user-friendly CLI, including TUI modes for secret & token generation.

Use cases:

- Multi-tenant ingress management without manual YAML.
- Dynamic redirects for legacy hostnames.
- Automated certificate & ingress lifecycle.

---

## Architecture

```text
        ┌────────────┐
        │  HTTP API  │<─── JWT bearer token
        |            |--------------┐
        └────────────┘              │
             │  HandleAddNewProxy   │
┌────────┐   │  HandleDeleteProxy   │      ┌───────────┐
│  CLI   │───┼  HandleGetRecords    │      │ Kubernetes│
│ prx    │   │  HandlePatchProxy    │      │  cluster  │
└────────┘   └────────┬─────────────┘      └───────────┘
             Uses Kube Client          Creates Secrets, Ingress, ConfigMap
             and JWT Service
                 │
                 ▼
        ┌──────────┐
        │ gRPC API │
        └──────────┘
```  

- **Server** (`cmd/server/main.go`): starts HTTP server on port 80 and gRPC on 50051.  
- **Kubernetes Client** (`internal/services/kubectl.go`): applies/deletes `Secret`, `Ingress`, `ConfigMap`.  
- **Persistence**: in-memory map + `ConfigMap` fallback.  
- **Auth**: `JWTService` signs tokens, validated by middleware and interceptor.

---

## Prerequisites

- Go ≥1.20  
- `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc`  
- Kubernetes cluster (v1.24+) with `IngressController`  
- Docker (for building image)  
- Helm v3 (for deployment chart)  

---

## Setup & Configuration

1. **Clone** the repo:
   ```bash
   git clone https://github.com/TypeTerrors/go_proxy.git
   cd go_proxy
   ```

2. **Generate gRPC code**:
   ```bash
   cd proto
   protoc --go_out=. --go-grpc_out=. reverse.proto
   ```  
3. **Build binaries**:
   ```bash
   go build -o prx-server ./cmd/server
   go build -o prx       ./cmd/client
   ```

4. **Environment Variables** for server:
   - `NAMESPACE`   – Kubernetes namespace to manage.  
   - `JWT_SECRET`  – base64 HMAC key (use `prx secret`).  
   - `PRX_KUBE_CONFIG` – optional base64 kubeconfig override.

---

## Running Locally

1. **Generate secret**:
   ```bash
   prx secret
   ```
2. **Get a JWT**:
   ```bash
   prx auth --secret YOUR_GENERATED_SECRET
   ```
3. **Run server**:
   ```bash
   export NAMESPACE=default
   export JWT_SECRET=...  
   kubectl port-forward svc/my-ingress-controller 80:80 &
   ./prx-server
   ```

4. **Test HTTP API**:
   ```bash
   curl -H "Authorization: Bearer $JWT" \
     -d '{"from":"example.com","to":"http://1.2.3.4","cert":"...","key":"..."}' \
     http://localhost/api/prx
   ```

5. **Test gRPC**:
   ```bash
   prx add --addr localhost:50051 \
       --token $JWT --from example.com --to http://1.2.3.4 \
       --cert tls.crt --key tls.key
   prx list --addr localhost:50051 --token $JWT
   ```

---

## Deployment

We use GitHub Actions to build, tag, push Docker image and deploy via Helm.

1. **Docker image** built and tagged `ghcr.io/typeterrors/go_proxy:<tag>`, latest.  
2. **Kubernetes setup**: ServiceAccount `prx-user` with non-expiring token, Role/RoleBinding for secrets, ingresses, configmaps.  
3. **Generate kubeconfig** for CLI users and set `PRX_KUBE_CONFIG` in Helm values.  
4. **Helm Chart** in `./charts/go-proxy`:
   ```bash
   helm upgrade --install go-proxy ./charts/go-proxy \
     --namespace $NAMESPACE \
     --set application.image.repository=ghcr.io/typeterrors/go_proxy \
     --set application.image.tag=<tag> \
     --set application.JWT_SECRET=$JWT_SECRET \
     --set global.PRX_KUBE_CONFIG=$KUBECONFIG_B64
   ```

---

## Usage

### HTTP API Examples

- **List** records:
  ```bash
  curl -H "Authorization: Bearer $JWT" http://<host>/api/prx
  ```
- **Add** record:
  ```bash
  curl -X POST -H "Authorization: Bearer $JWT" \
    -H "Content-Type: application/json" \
    -d '{"from":"foo.com","to":"http://1.2.3.4","cert":"$(base64 tls.crt)","key":"$(base64 tls.key)"}' \
    http://<host>/api/prx
  ```

### gRPC CLI Examples

- **Add**:
  ```bash
  prx add \
    --addr proxy.mydomain:50051 \
    --token $JWT \
    --from example.com \
    --to http://1.2.3.4 \
    --cert tls.crt \
    --key tls.key
  ```
- **List**:
  ```bash
  prx list --addr proxy:50051 --token $JWT
  ```

---

## GitHub Workflow

The `BuildPushDeploy` action triggers on PR to `main`/`dev`:

1. Checkout, get short SHA + PR number ⇒ `<tag>`.
2. Build `/go_proxy.dockerfile` with `VERSION=<tag>`, push to GHCR.
3. Create GitHub release.
4. Setup `kubectl` & `helm`.
5. Configure cluster with `prx-user` SA, Role, RoleBinding.
6. Generate and encode kubeconfig for CLI users.
7. Create imagePullSecret & TLS secret for ingress.
8. Run `helm upgrade --install go-proxy ...`.

---

