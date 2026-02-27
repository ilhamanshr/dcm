# MrScraper Distributed Config System

A distributed configuration management system composed of three Go microservices: **Controller**, **Agent**, and **Worker**. The Controller acts as the central configuration authority; the Agent polls the Controller for config changes and propagates them to the Worker; the Worker executes the configured workload (HTTP scraping).

---

## Table of Contents

- [Architecture](#architecture)
- [Services Overview](#services-overview)
- [Prerequisites](#prerequisites)
- [Setup & Compile Instructions](#setup--compile-instructions)
  - [Option A – Docker Compose (Recommended)](#option-a--docker-compose-recommended)
  - [Option B – Run Locally](#option-b--run-locally)
- [Configuration Reference](#configuration-reference)
- [API Documentation](#api-documentation)
  - [Controller Service API](#controller-service-api)
  - [Worker Service API](#worker-service-api)
- [Authentication](#authentication)
- [Project Structure](#project-structure)

---

## Architecture
![Architecture](assets/arc.png)

### How It Works

1. **An administrator** calls `POST /config` on the Controller to set a target URL and poll interval.
2. **The Agent** registered itself on startup via `POST /register`, receiving a config with the target URL and poll interval
3. **On each poll cycle**, the Agent calls `GET /config` on the Controller. If the `Version` response header differs from the cached version in Redis, the Agent pushes the new config to the Worker via `POST /config`.
4. **The Worker** stores the URL in memory. When `GET /hit` is called, it performs an HTTP GET to the configured URL and returns the response body.
5. **Back-off and retry**: the Agent uses exponential back-off (capped at 30 s) on errors.

---

## Services Overview

| Service | Port | Language | Storage | Role |
|---|---|---|---|---|
| `controller-service` | `8080` | Go 1.24 | PostgreSQL | Config authority, agent registry |
| `worker-service` | `8081` | Go 1.24 | In-memory | Executes HTTP scrape requests |
| `agent-service` | — (daemon) | Go 1.24 | Redis | Bridges controller ↔ worker |

---

## Prerequisites

| Tool | Minimum Version | Notes |
|---|---|---|
| Go | 1.24 | All three modules |
| Docker & Docker Compose | v2 | For containerised setup |
| PostgreSQL | 16 | Controller storage |
| Redis | 7 | Agent config cache |
| `swag` CLI | latest | Only for re-generating Swagger docs |

Install `swag` (if regenerating docs):
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## Setup & Compile Instructions

### Option A – Docker Compose (Recommended)

The project ships with two compose files. Run them in order:

#### Step 1: Start the Controller (+ its PostgreSQL database)

```bash
docker compose -f docker/controller-service-compose.yaml up --build
```

This starts:
- **PostgreSQL 16** on host port `5434`
- **controller-service** on host port `8080` with auto DB migration on startup

#### Step 2: Start the Agent + Worker (+ Redis)

```bash
docker compose -f docker/agent-worker-service-compose.yaml up --build
```

This starts:
- **Redis 7** on host port `6377`
- **worker-service** on host port `8081`
- **agent-service** (daemon; registers with controller then begins polling)

> **Note:** The Agent compose file references `http://controller:8080` — make sure both compose stacks share a Docker network, or update `CONTROLLER_URL` to the host-accessible URL (e.g. `http://host.docker.internal:8080`).

---

### Option B – Run Locally

#### 1. Start infrastructure

```bash
# PostgreSQL
docker run -d --name pg -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=controller -p 5432:5432 postgres:16

# Redis
docker run -d --name redis -p 6379:6379 redis:7
```

#### 2. Controller Service

```bash
cd controller-service
cp .env.example .env
# Edit .env (see Configuration Reference below)
go run ./cmd/main.go
```

Database migrations run automatically on startup.

#### 3. Worker Service

```bash
cd worker-service
cp .env.example .env
go run ./cmd/main.go
```

#### 4. Agent Service

```bash
cd agent-service
cp .env.example .env
# Ensure CONTROLLER_URL and WORKER_URL point to running services
go run ./cmd/main.go
```

#### Building Binaries

```bash
# From each service directory:
go build -o bin/controller ./cmd/main.go
go build -o bin/worker     ./cmd/main.go
go build -o bin/agent      ./cmd/main.go
```

#### Running Tests (Controller)

```bash
cd controller-service
go test ./...
```

#### Regenerating Swagger Docs

```bash
cd controller-service && swag init -g cmd/main.go
cd worker-service     && swag init -g cmd/main.go
```

---

## Configuration Reference

All services are configured via environment variables. Each directory contains a `.env.example` file.

### Controller Service (`controller-service/.env.example`)

| Variable | Required | Example | Description |
|---|---|---|---|
| `APP_PORT` | ✅ | `8080` | HTTP server port |
| `DB_URL` | ✅ | `postgres://postgres:postgres@localhost:5432/controller?sslmode=disable` | PostgreSQL connection string |
| `API_KEY` | ✅ | `supersecret` | Shared secret for `X-API-Key` authentication |

**`.env` example:**
```env
APP_PORT=8080
DB_URL=postgres://postgres:postgres@localhost:5432/controller?sslmode=disable
API_KEY=supersecret
```

---

### Worker Service (`worker-service/.env.example`)

| Variable | Required | Example | Description |
|---|---|---|---|
| `APP_PORT` | ✅ | `8081` | HTTP server port |
| `API_KEY` | ✅ | `supersecret` | Shared secret for `X-API-Key` authentication |

**`.env` example:**
```env
APP_PORT=8081
API_KEY=supersecret
```

---

### Agent Service (`agent-service/.env.example`)

| Variable | Required | Example | Description |
|---|---|---|---|
| `CONTROLLER_URL` | ✅ | `http://localhost:8080` | Base URL of the Controller Service |
| `WORKER_URL` | ✅ | `http://localhost:8081` | Base URL of the Worker Service |
| `API_KEY` | ✅ | `supersecret` | Shared secret (must match Controller + Worker) |
| `REDIS_ADDR` | ✅ | `localhost:6379` | Redis host and port |
| `REDIS_PASSWORD` | ❌ | _(empty)_ | Redis password (leave blank if none) |
| `REDIS_DB` | ❌ | `0` | Redis logical database index |

**`.env` example:**
```env
CONTROLLER_URL=http://localhost:8080
WORKER_URL=http://localhost:8081
API_KEY=supersecret
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

> ⚠️ **Security Note:** The same `API_KEY` value must be set in all three services. All service-to-service requests carry this key in the `X-API-Key` HTTP header.

---

## API Documentation

Interactive Swagger UI is available at runtime:

| Service | URL |
|---|---|
| Controller | http://localhost:8080/docs/ |
| Worker | http://localhost:8081/docs/ |

---

### Controller Service API

**Base URL:** `http://localhost:8080`  
**Authentication:** All endpoints require `X-API-Key: <your-api-key>` header.

---

#### `POST /register` — Register Agent

Registers a new agent and returns its initial configuration.

**Request Body:**
```json
{
  "name": "agent-abc123"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | ✅ | Human-readable agent name |

**Response `200 OK`:**
```json
{
  "agent_id": "550e8400-e29b-41d4-a716-446655440000",
  "poll_url": "https://example.com/data",
  "poll_interval": 10
}
```

| Field | Type | Description |
|---|---|---|
| `agent_id` | string (UUID) | Unique identifier for this agent |
| `poll_url` | string | Target URL the worker should hit |
| `poll_interval` | int | How often (seconds) the agent should poll |

**Error Responses:**

| Status | Description |
|---|---|
| `400` | Invalid request body |
| `500` | Internal server error |

---

#### `GET /config` — Get Current Config

Returns the current global config. Includes a `Version` response header for change detection.

**Response `200 OK`:**
```json
{
  "poll_url": "https://example.com/data",
  "poll_interval": 10
}
```

**Response Headers:**

| Header | Type | Description |
|---|---|---|
| `Version` | integer | Monotonically increasing config version |

**Error Responses:**

| Status | Description |
|---|---|
| `500` | Failed to retrieve config |

---

#### `POST /config` — Update Config

Updates the global configuration. All registered agents will detect this change on their next poll cycle.

**Request Body:**
```json
{
  "url": "https://example.com/data",
  "poll_interval": 15
}
```

| Field | Type | Required | Validation | Description |
|---|---|---|---|---|
| `url` | string | ✅ | Non-empty | Target URL for workers to scrape |
| `poll_interval` | int | ✅ | > 0 | Agent poll frequency in seconds |

**Response `200 OK`:** Empty body on success.

**Error Responses:**

| Status | Description |
|---|---|
| `400` | Invalid or missing fields |
| `500` | Failed to update config |

---

### Worker Service API

**Base URL:** `http://localhost:8081`  
**Authentication:** All endpoints require `X-API-Key: <your-api-key>` header.

---

#### `POST /config` — Set Worker Target URL

Configures the URL that the worker will hit when `/hit` is called.

**Request Body:**
```json
{
  "url": "https://example.com/data"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `url` | string | ✅ | Target URL to scrape |

**Response `200 OK`:** Empty body on success.

**Error Responses:**

| Status | Description |
|---|---|
| `400` | `url` is empty |
| `500` | Failed to parse request |

---

#### `GET /hit` — Hit Configured URL

Performs an HTTP GET to the configured URL and returns the raw response body.

**Response `200 OK`:** Raw response body from the configured URL.

**Error Responses:**

| Status | Description |
|---|---|
| `400` | No URL configured yet |
| `500` | HTTP request or read failure |

---

## Authentication

All inter-service and external HTTP calls are authenticated via a shared API key passed as a request header:

```
X-API-Key: <your-api-key>
```

The middleware rejects any request without a matching key with `401 Unauthorized`. The same key value must be configured in the `API_KEY` environment variable for all three services.

**Example with cURL:**
```bash
# Register an agent
curl -X POST http://localhost:8080/register \
  -H "X-API-Key: supersecret" \
  -H "Content-Type: application/json" \
  -d '{"name": "my-agent"}'

# Update the global config
curl -X POST http://localhost:8080/config \
  -H "X-API-Key: supersecret" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "poll_interval": 10}'

# Trigger a scrape manually
curl -X GET http://localhost:8081/hit \
  -H "X-API-Key: supersecret"
```

---

## Project Structure

```
mrscraper-test/
├── controller-service/          # Central config management service
│   ├── cmd/main.go              # Entry point; wires deps, registers routes, runs migrations
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/         # HTTP handlers (Register, GetConfig, UpdateConfig)
│   │   │   ├── middleware/      # API key auth middleware
│   │   │   ├── request/         # Request structs + validation
│   │   │   └── response/        # Response structs (ConfigResponse)
│   │   ├── config/              # Env loading (APP_PORT, DB_URL, API_KEY)
│   │   ├── database/            # DB connection + golang-migrate auto-migrations
│   │   ├── repository/          # sqlc-generated DB queries
│   │   └── service/             # Business logic (RegisterAgent, GetConfig, UpdateConfig)
│   ├── docs/                    # Swagger-generated docs (swag init output)
│   ├── Dockerfile
│   ├── sqlc.yaml                # sqlc code generation config
│   └── .env.example
│
├── worker-service/              # HTTP scraping worker
│   ├── cmd/main.go              # Entry point; registers routes
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler/         # WorkerHandler (UpdateConfig, Hit); thread-safe via sync.RWMutex
│   │   │   └── middleware/      # API key auth middleware
│   │   └── config/              # Env loading (APP_PORT, API_KEY)
│   ├── docs/                    # Swagger-generated docs
│   ├── Dockerfile
│   └── .env.example
│
├── agent-service/               # Config propagation daemon
│   ├── cmd/main.go              # Entry point; connects Redis, starts AgentService
│   ├── internal/
│   │   ├── config/              # Env loading (CONTROLLER_URL, WORKER_URL, API_KEY, Redis*)
│   │   ├── repository/
│   │   │   └── redis/           # Redis cache helper (SetKey, GetKey, Ping)
│   │   └── service/
│   │       └── agent.go         # RegisterAgent, polling loop, configCheck, sendConfig
│   └── .env.example
│
└── docker/
    ├── controller-service-compose.yaml    # Controller + PostgreSQL
    └── agent-worker-service-compose.yaml  # Agent + Worker + Redis
```
