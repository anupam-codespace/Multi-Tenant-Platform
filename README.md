# Multi-Tenant Quick-Commerce Platform & Agentic AI Engine

[![CI](https://github.com/anupam-codespace/Multi-Tenant-Platform/actions/workflows/ci.yml/badge.svg)](https://github.com/anupam-codespace/Multi-Tenant-Platform/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance, multi-tenant backend engine engineered in **Go** for **Quick-Commerce (10-15 minute delivery) and Multi-Store E-Commerce ecosystems**. Designed to solve the demanding challenges of hyper-local dark stores, high-concurrency flash ordering, atomic inventory reservations, and autonomous **Agentic AI shopping workflows**.

---

## Architectural Highlights

- **Hyper-Local Dark Store Architecture**: Multi-tenant isolation for micro-warehouses (pods) and merchant partners, isolating catalog visibility and stock per geo-cluster.
- **Atomic Inventory Reservation with TTL**: High-throughput distributed inventory manager that prevents race conditions and overselling during flash traffic spikes, featuring background sweepers for automated TTL expiration.
- **Idempotent Order Pipeline**: Custom middleware enforcing `X-Idempotency-Key` semantics to prevent double-charging and duplicate checkouts caused by mobile network drops.
- **Autonomous Agentic AI Shopping Assistant**: Built-in Go Agent implementing an autonomous tool-calling loop (`FindNearestDarkStore`, `SearchCatalog`, `CheckStock`, `GetSubstitutions`, `AddToCart`) for natural language grocery and meal ordering with automatic out-of-stock substitutions.
- **Enterprise Middleware Chain**: Token bucket rate limiting (DDoS resilience), JWT-based Role-Based Access Control (RBAC), CORS negotiation, structured logging, and panic recovery.
- **Resilient Operations**: Graceful SIGINT/SIGTERM shutdown with active connection draining, database connection pooling, and multi-stage Docker containerization.

---

## System Architecture

```text
                               +----------------------------------------+
                               |     Mobile Apps / Web Clients          |
                               +-------------------+--------------------+
                                                   |
                                                   v
                         +----------------------------------------------------+
                         |             HTTP Gateway / Middleware              |
                         |  - Token Bucket Rate Limiter (DDoS Mitigation)     |
                         |  - X-Idempotency-Key Deduplication Engine          |
                         |  - JWT Authentication & RBAC Engine                |
                         +-------------------------+--------------------------+
                                                   |
                     +-----------------------------+-----------------------------+
                     |                                                           |
                     v                                                           v
       +---------------------------+                               +---------------------------+
       |   Quick-Commerce Engine   |                               |     Agentic AI Engine     |
       |  - Atomic Stock Hold (TTL)|                               |  - Natural Language Parser|
       |  - Order State Machine    |                               |  - Multi-Step Tool Loop   |
       |  - Pod Fulfillment Router |                               |  - Semantic Substitutions |
       +-------------+-------------+                               +-------------+-------------+
                     |                                                           |
                     +-----------------------------+-----------------------------+
                                                   |
                                                   v
                               +----------------------------------------+
                               |            Storage Layer               |
                               |  - PostgreSQL 16 (Relational Schema)   |
                               |  - Redis 7 (In-Memory Stock & Cache)   |
                               +----------------------------------------+
```

---

## Tech Stack & Design Rationale

| Component | Technology | Engineering Rationale |
| :--- | :--- | :--- |
| **Language** | **Go 1.22+** | Statically typed, zero garbage collection overhead on critical paths, first-class concurrency via goroutines and channels. |
| **Web Framework** | **Gin Gonic** | Low-latency HTTP routing with minimal allocations per request. |
| **Database** | **PostgreSQL 16** | ACID compliance, transactional isolation for payments, and relational integrity across stores and items. |
| **In-Memory Cache**| **Redis 7** | Sub-millisecond atomic inventory decrementing and idempotency token tracking. |
| **AI Integration** | **Agentic Go Engine** | Multi-tool autonomous agent coordinating catalog discovery, availability checks, and automatic item substitutions. |
| **CLI & Ops** | **Cobra & Viper** | 12-factor configuration management and extensible operational command-line interfaces. |
| **Containerization**| **Docker & Compose** | Reproducible multi-stage Alpine builds with non-root execution profiles. |

---

## Agentic AI Shopping Workflow

The built-in Instamart Agent allows customers to express high-level culinary or grocery intents (e.g., *"Prepare groceries for Paneer Butter Masala dinner for 2 people in Whitefield"*):

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant Agent as Instamart AI Agent
    participant Catalog as Dark Store Catalog
    participant Inv as Inventory Engine

    User->>Agent: "Ingredients for Paneer Butter Masala in Whitefield"
    Agent->>Catalog: FindNearestDarkStore("560066")
    Catalog-->>Agent: Returns Pod 101 (ETA: 9 mins)
    Agent->>Catalog: SearchCatalog(101, ["paneer", "butter", "rice", "masala"])
    Catalog-->>Agent: Matches found (Daawat Rice: Out of Stock)
    Agent->>Catalog: GetSubstitutions(Daawat Rice)
    Catalog-->>Agent: Recommends India Gate Super Basmati (In Stock)
    Agent->>Inv: ReserveStock(ReservationID, ItemIDs)
    Inv-->>Agent: Hold acquired (TTL: 5 mins)
    Agent-->>User: Structured Cart + ETA (9m) + Execution Audit Trail
```

---

## Quick-Commerce Endpoints

### 1. Atomic Stock Reservation
- **`POST /quickcommerce/reserve`**
```json
{
  "reservation_id": "res_89312847",
  "store_id": 101,
  "item_id": 3,
  "user_id": 42,
  "quantity": 2
}
```
*Returns HTTP 200 with reservation TTL details or HTTP 409 Conflict if stock is insufficient.*

### 2. Commit Reservation (Payment Succeeded)
- **`POST /quickcommerce/commit/:reservation_id`**
*Permanently updates dark store inventory upon payment authorization.*

### 3. Agentic AI Grocery Assistant
- **`POST /quickcommerce/ai/assist`**
```json
{
  "prompt": "Groceries for butter chicken and basmati rice tonight",
  "pincode": "560066"
}
```
*Returns execution plan, nearest dark store routing, substituted items, and estimated delivery time.*

---

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local builds)
- PostgreSQL 16+ (if running without Docker)

### 1. Clone & Configure
```bash
git clone https://github.com/anupam-codespace/Multi-Tenant-Platform.git
cd Multi-Tenant-Platform
cp .env.example .env.dev
```

### 2. Run with Docker Compose (Recommended)
Boot PostgreSQL, Redis, database migrations, and the Go API with one command:
```bash
docker compose up --build -d
```

Verify service readiness:
```bash
curl http://localhost:8080/public/health-check
```

### 3. Run Locally with Make
```bash
make build
make api
```

### 4. Run Concurrency & Agentic Tests
```bash
go test -v -race ./...
```

---

## Concurrency & Performance Benchmarks

- **Atomic Stock Reservation**: Benchmarked with **1,000 concurrent goroutines** competing for limited inventory. Demonstrates zero over-allocation and deterministic stock exhaustion under race conditions.
- **Idempotency Deduplication**: Average cached replay latency `< 1.2ms`.
- **Memory Footprint**: Alpine container consumes `< 35MB` resident set size (RSS) under baseline load.

---

## Project Structure

```text
.
├── agentic/               # Instamart Autonomous Agent, tools, and substitution engine
│   ├── agent.go           # Autonomous reasoning and tool-calling loop
│   ├── tools.go           # Catalog and dark store discovery interfaces
│   └── agent_test.go      # Agent loop execution and substitution tests
├── quickcommerce/         # Quick-commerce real-time inventory and idempotency engine
│   ├── inventory_service.go      # Atomic reservation, locks, and TTL expiration
│   ├── inventory_service_test.go # High-concurrency race condition test suite
│   └── idempotency_middleware.go # X-Idempotency-Key replay protection
├── cmd/                   # Application commands (Cobra CLI bootstrap)
│   ├── api_server.go      # Server lifecycle and dependency wiring
│   └── root.go            # Root CLI commands
├── configs/               # Viper environment configuration
├── controllers/           # HTTP controllers for admin, seller, and user domains
├── Dao/                   # Database data access objects
├── db_migrations/         # Version-controlled raw SQL migrations
├── middleware/            # Rate limiting, RBAC, CORS, and auth middleware
├── models/                # Core domain and DTO models
├── router/                # Route registration and HTTP handlers
├── Dockerfile             # Multi-stage production container build
├── docker-compose.yml     # Local orchestration (PostgreSQL + Redis + API)
└── Makefile               # Development automation targets
```

---

## Author & Contact

**Anupam Saha**  
- **GitHub**: [@anupam-codespace](https://github.com/anupam-codespace)  
- **Specialization**: Backend Engineering (Go), Distributed Systems, Quick-Commerce Platforms, Agentic AI.
