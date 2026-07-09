# Nexus — Production-Grade Microservices E-Commerce Platform

<div align="center">

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-16.2-000000?style=for-the-badge&logo=next.js&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)
![Elasticsearch](https://img.shields.io/badge/Elasticsearch-8.17-005571?style=for-the-badge&logo=elasticsearch&logoColor=white)
![Stripe](https://img.shields.io/badge/Stripe-Integrated-635BFF?style=for-the-badge&logo=stripe&logoColor=white)

**A fully functional, production-grade e-commerce backend built from scratch using Go microservices, event-driven architecture, and a modern Next.js frontend.**

</div>

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Microservices](#microservices)
- [Inter-Service Communication](#inter-service-communication)
- [Key Design Patterns](#key-design-patterns)
- [Complete Checkout Flow](#complete-checkout-flow)
- [Frontend — Buyer App](#frontend--buyer-app)
- [Local Setup Guide](#local-setup-guide)
- [Environment Variables Reference](#environment-variables-reference)
- [API Reference](#api-reference)
- [Infrastructure](#infrastructure)

---

## Overview

Nexus is a **monorepo** containing 7 Go microservices and a Next.js buyer-facing storefront. It was built to demonstrate real-world backend engineering patterns including:

- **Clean Architecture** across every service
- **Event-driven communication** with RabbitMQ and the Outbox pattern
- **gRPC** for synchronous inter-service calls
- **OAuth 2.0 + JWT (RS256)** for authentication
- **Full-text product search** with Elasticsearch
- **Stripe payment processing** with webhook-driven order fulfillment
- **PDF invoice generation** stored on MinIO (S3-compatible)
- **Dual ID strategy** — UUID v7 internally, NanoID publicly

> ⚠️ **Note on Deployment:** This project is not publicly hosted. The infrastructure cost (Elasticsearch, MinIO, PostgreSQL, RabbitMQ, Redis, Stripe live keys) makes cloud deployment impractical for a personal project. Everything is designed to run fully locally via Docker Compose.

---

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          BUYER FRONTEND                                  │
│               Next.js 16 + Tailwind v4 + shadcn/ui                      │
│         (Auth, Product Browse, Cart, Checkout, Orders, Profile)          │
└───────────────────┬──────────────┬────────────────┬─────────────────────┘
                    │ REST          │ REST            │ REST
        ┌───────────▼──┐  ┌────────▼────────┐  ┌────▼──────────────────┐
        │  Auth Service │  │ Catalog Service │  │   Order Service       │
        │  :8080        │  │ :8082           │  │   :8084               │
        │               │  │                 │  │                       │
        │ JWT (RS256)   │  │ Products        │  │ Cart (Redis)          │
        │ Google OAuth  │  │ Variants        │  │ Checkout              │
        │ OTP via Email │  │ Categories      │  │ Orders                │
        │ Refresh Token │  │ Sellers         │  │ Invoice (MinIO/PDF)   │
        └───────┬───────┘  └────────┬────────┘  └──────────┬────────────┘
                │ HTTP               │ gRPC :50051            │ gRPC
                ▼                   │                        ▼
        ┌───────────────┐           │               ┌────────────────────┐
        │ Email Service │           │               │  Payment Service   │
        │ :8081         │           │               │  :8085 / :50052    │
        │ SMTP + OTP    │           │               │                    │
        └───────────────┘           │               │  Stripe Checkout   │
                                    │               │  Webhook listener  │
        ┌───────────────┐           │               │  Outbox Pattern    │
        │ Media Service │           │               └────────┬───────────┘
        │ :8083         │           │                        │ RabbitMQ
        │ MinIO uploads │           │               ┌────────▼───────────┐
        └───────────────┘           │               │  Search Service    │
                                    │               │  :8086             │
        ┌───────────────────────────▼──────────┐   │  Elasticsearch     │
        │           INFRASTRUCTURE              │   │  Full-text search  │
        │  PostgreSQL · Redis · RabbitMQ        │   └────────────────────┘
        │  MinIO · Elasticsearch                │
        └───────────────────────────────────────┘
```

### Event-Driven Data Flow

```
Catalog Service ──[product.created/updated/deleted]──► RabbitMQ ──► Search Service
                                                       (catalog_events exchange)
                                                       Indexes/updates Elasticsearch

Payment Service ──[payment.OrderPaid]──► RabbitMQ ──► Order Service
                 (Outbox pattern)        (payment_events exchange)
                                         Updates order, decreases inventory, clears cart
```

---

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Language | Go 1.26 | All backend microservices |
| HTTP Framework | Gin | REST API handlers |
| ORM | GORM (generics `gorm.G[T]`) | Type-safe database access |
| Primary DB | PostgreSQL 15 | Per-service relational storage |
| Cache / Cart | Redis 7 | Cart state, OTP tokens, session data |
| Message Broker | RabbitMQ 3 (AMQP 0.9.1) | Async inter-service events |
| File Storage | MinIO (S3-compatible) | Product images + PDF invoices |
| Search Engine | Elasticsearch 8.17 | Full-text product search |
| Payment Gateway | Stripe | Checkout sessions + webhooks |
| RPC Protocol | gRPC + Protocol Buffers | Synchronous inter-service calls |
| Auth | JWT RS256 + Google OAuth 2.0 | Stateless authentication |
| Logging | Uber Zap | Structured, leveled logging |
| ID Generation | NanoID + UUID v7 | Public/internal dual ID strategy |
| PDF Generation | gofpdf | Invoice generation |
| Frontend | Next.js 16.2.7 (App Router) | Buyer-facing storefront |
| UI Components | shadcn/ui | Accessible component library |
| CSS | Tailwind CSS v4 (oklch) | Design system |
| State Management | Zustand (persisted) | Client-side auth + cart state |
| HTTP Client | Axios (with interceptors) | Frontend API calls with JWT refresh |

---

## Project Structure

```
ecommerce-project/
├── go.work                     # Go workspace — links all 7 modules
│
├── pkg/                        # Shared libraries (consumed by all services)
│   ├── broker/                 # RabbitMQ producer/consumer helpers
│   ├── database/               # GORM connection setup + generic helpers
│   ├── logger/                 # Uber Zap logger initializer
│   ├── protobufs/              # Generated gRPC Go code
│   │   ├── catalog/            # CatalogService (CheckPrices, DecreaseInventory)
│   │   └── payment/            # PaymentService (CreatePaymentSession)
│   └── utils/                  # Shared utilities
│
├── services/
│   ├── auth/                   # ✅ JWT + Google OAuth + OTP email verification
│   │   ├── cmd/server/         # Entry point
│   │   ├── internal/
│   │   │   ├── domain/         # User, RefreshToken, OTP models
│   │   │   ├── handler/        # HTTP routes + middleware + JWT verification
│   │   │   ├── service/        # Business logic (register, login, OAuth, OTP)
│   │   │   ├── repository/     # GORM data access
│   │   │   ├── workers/        # Background jobs
│   │   │   ├── client/         # HTTP client for Email service
│   │   │   └── utils/          # RSA key loading, JWT helpers
│   │   └── migrations/         # SQL migration files
│   │
│   ├── catalog/                # ✅ Products, Variants, Categories, Sellers
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── domain/         # Product, Variant, Category, Seller models
│   │   │   ├── handler/        # HTTP + gRPC server (port 50051)
│   │   │   ├── service/        # Catalog business logic + event publishing
│   │   │   └── repository/     # GORM + RabbitMQ publishing
│   │   └── migrations/
│   │
│   ├── order/                  # ✅ Cart, Checkout, Orders, Invoice PDF
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── domain/         # Order, OrderItem, Invoice models
│   │   │   ├── handler/        # HTTP handlers + routes
│   │   │   ├── service/        # Order + Invoice business logic
│   │   │   ├── repository/     # GORM + Redis (cart) + MinIO (invoices)
│   │   │   ├── client/         # gRPC clients (Catalog, Payment)
│   │   │   └── workers/        # PaymentConsumer (RabbitMQ listener)
│   │   └── migrations/
│   │
│   ├── payment/                # ✅ Stripe integration + Outbox pattern
│   │   ├── cmd/server/
│   │   ├── internal/
│   │   │   ├── domain/         # Payment, OutboxEvent models
│   │   │   ├── handler/        # HTTP handlers (checkout, webhook)
│   │   │   ├── service/        # Stripe session creation, webhook processing
│   │   │   ├── repository/     # GORM outbox persistence
│   │   │   └── workers/        # OutboxWorker (polls + publishes to RabbitMQ)
│   │   └── migrations/
│   │
│   ├── search/                 # ✅ Elasticsearch full-text + filter search
│   │   ├── cmd/server/
│   │   └── internal/
│   │       ├── domain/         # SearchProduct ES document model
│   │       ├── handler/        # Search HTTP endpoint
│   │       ├── repository/     # Elasticsearch client queries
│   │       ├── service/        # Search logic + filter composition
│   │       └── workers/        # CatalogConsumer (indexes products from RabbitMQ)
│   │
│   ├── email/                  # ✅ SMTP email sending + templates
│   └── media/                  # ✅ MinIO file upload (product images)
│
├── web/
│   └── buyer-app/              # ✅ Next.js 16 buyer storefront (15 routes)
│       └── src/
│           ├── app/            # App Router pages
│           │   ├── page.tsx            # Home (hero, categories, products)
│           │   ├── products/           # Product listing + detail page
│           │   ├── cart/               # Cart with quantity controls
│           │   ├── checkout/           # Checkout + Stripe redirect + success/cancel
│           │   ├── orders/             # Order list + order detail + invoice
│           │   ├── profile/            # Profile + addresses
│           │   └── auth/               # Login, register, OTP verify, OAuth callback
│           ├── components/     # Navbar, Footer, ProductCard, CategorySidebar
│           └── lib/            # api.ts, store.ts (Zustand), types.ts
│
└── deploy/
    ├── docker-compose.yml      # All 5 infrastructure services
    └── init.sql                # Database initialization (7 DBs)
```

---

## Microservices

### Auth Service — Port 8080

The central authentication service. All other services validate JWTs against Auth's public key.

**Capabilities:**
- Email/password registration with bcrypt hashing
- **Email OTP verification** — sends a 6-digit code via Email service, stored in Redis with TTL
- Login → issues **Access Token** (short-lived, RS256 JWT) + **Refresh Token** (HttpOnly cookie)
- Token refresh endpoint — validates stored refresh token, issues new pair
- **Google OAuth 2.0** — full authorization code flow, callback handler, JWT issuance
- Public key endpoint (`GET /api/v1/auth/public-key`) — other services fetch RSA public key at startup

**JWT Strategy:** Signed with RSA private key (RS256). Payload contains `user_id` and `role`. Middleware in each service verifies the signature using the cached public key.

---

### Catalog Service — Port 8082 / gRPC :50051

Product catalog management with event publishing to Search.

**Capabilities:**
- Full CRUD for Products, Variants, Categories, Sellers
- **Variant pricing** — each product has multiple variants (size/color) with independent pricing and inventory
- **gRPC server** on port 50051 — exposes `CheckPrices` and `DecreaseInventory` RPCs for Order service
- **RabbitMQ publishing** — emits `product.created`, `product.updated`, `product.deleted` events to `catalog_events` exchange, consumed by Search service

**Dual ID pattern:** Each entity has a UUID v7 primary key (internal) and a NanoID with prefix (`itm_`, `var_`, `cat_`, `sel_`) exposed in API responses.

---

### Order Service — Port 8084

The most complex service. Orchestrates the entire checkout flow.

**Capabilities:**
- **Cart** stored in Redis (`cart:{user_id}` JSON blob) — add, update, remove, clear
- **Checkout** — fetches cart → gRPC `CheckPrices` → creates Order record → gRPC `CreatePaymentSession` → returns Stripe URL
- Order management — list orders, get order detail, cancel pending orders
- **Invoice PDF generation** with `gofpdf` — A4 format, itemized bill, uploaded to MinIO on payment success
- **Invoice download** — presigned MinIO URL (15-minute TTL)
- **PaymentConsumer** — listens on RabbitMQ for `payment.OrderPaid` events, updates order status, calls `DecreaseInventory`, clears cart

---

### Payment Service — Port 8085 / gRPC :50052

Handles Stripe integration with guaranteed event delivery via the Outbox pattern.

**Capabilities:**
- gRPC `CreatePaymentSession` — creates Stripe Checkout session, returns hosted payment URL
- **Stripe webhook** — receives `checkout.session.completed`, verifies signature, updates payment record
- **Outbox Pattern** — payment status update + OutboxEvent written in a single DB transaction
- **OutboxWorker** — polls the outbox table every 5 seconds, publishes undelivered events to RabbitMQ, marks as delivered

**Why Outbox?** Prevents the dual-write problem — if the service crashes after updating payment but before publishing to RabbitMQ, the outbox guarantees eventual consistency.

---

### Search Service — Port 8086

Stateless service — no SQL database. Uses Elasticsearch exclusively.

**Capabilities:**
- Full-text search across product name, description, category
- Filters: category, price range (`min_price`/`max_price`), in-stock only
- **Custom Elasticsearch analyzer** for better tokenization
- **CatalogConsumer** — listens to RabbitMQ `catalog_events`, indexes/updates/deletes Elasticsearch documents in real-time

---

### Email Service — Port 8081

Internal-only service for SMTP email delivery.

- OTP emails triggered by Auth service
- HTML email templates

---

### Media Service — Port 8083

Internal file upload service backed by MinIO.

- Accepts `multipart/form-data` uploads
- Stores in MinIO `ecommerce-images` bucket
- Returns publicly accessible URL

---

## Inter-Service Communication

```
                    ┌─── HTTP ───► Email Service (OTP dispatch)
Auth Service ───────┤
                    └─── RSA public key endpoint (consumed by all services at startup)

Order Service ──────┬─── gRPC ──► Catalog Service:50051 (CheckPrices, DecreaseInventory)
                    └─── gRPC ──► Payment Service:50052 (CreatePaymentSession)

Payment Service ────── Outbox ──► RabbitMQ ──► Order Service (payment.OrderPaid)
                                  payment_events exchange

Catalog Service ────── Events ──► RabbitMQ ──► Search Service (product.created/updated/deleted)
                                  catalog_events exchange
```

### Protocol Buffers

**`pkg/protobufs/catalog/catalog.proto`**
```protobuf
service CatalogService {
  rpc CheckPrices(CheckPricesRequest) returns (CheckPricesResponse) {}
  rpc DecreaseInventory(DecreaseInventoryRequest) returns (DecreaseInventoryResponse) {}
}
```

**`pkg/protobufs/payment/payment.proto`**
```protobuf
service PaymentService {
  rpc CreatePaymentSession(CreatePaymentRequest) returns (CreatePaymentResponse) {}
}
```

---

## Key Design Patterns

### 1. Clean Architecture (Every Service)
Each service enforces strict layering with no dependency inversion violations:
```
cmd/server/main.go          ← Dependency injection root (wires everything)
internal/
  domain/                   ← Pure Go structs; zero framework imports
  handler/                  ← HTTP/gRPC; calls service interfaces
  service/                  ← Business logic; calls repository interfaces
  repository/               ← Data access; GORM, Redis, MinIO, Elasticsearch
  workers/                  ← Background goroutines (RabbitMQ consumers)
  client/                   ← gRPC/HTTP clients to other services
```

### 2. Dual ID Strategy
- **UUID v7** — internal primary key in PostgreSQL. Time-sortable, avoids index fragmentation
- **NanoID with prefix** — `ORD-xxxx`, `itm_xxxx`, `var_xxxx` — safe to expose in URLs and API responses, prevents enumeration attacks

### 3. Outbox Pattern (Payment → Order)
Solves the distributed transaction problem without a 2PC coordinator:
1. Payment webhook handler writes `Payment(status=paid)` + `OutboxEvent` in one DB transaction
2. `OutboxWorker` goroutine polls `outbox_events WHERE delivered=false` every 5s
3. Publishes to RabbitMQ, marks `delivered=true`
4. Order service's `PaymentConsumer` processes the event idempotently

### 4. JWT Public Key Federation
Auth service holds the RSA private key. All other services call `GET /api/v1/auth/public-key` **once at startup** and cache the public key in memory. Subsequent JWT verifications are local — no network hop per request.

### 5. Cart as Redis Ephemeral State
Cart is not a database entity. It lives in Redis as a JSON blob under `cart:{user_id}`. This choice enables:
- Sub-millisecond reads
- Automatic TTL expiry for abandoned carts
- Simple atomic updates with Redis SET

### 6. Event-Driven Search Indexing
The Search service never talks to the Catalog database. It maintains its own read model in Elasticsearch, updated in real-time via RabbitMQ events from Catalog. This is the **CQRS** pattern applied to search.

---

## Complete Checkout Flow

```
1. User adds items
   POST /api/v1/cart/add  →  Order Service stores {variant_id, quantity} in Redis

2. User views cart
   GET /api/v1/cart  →  Order Service reads Redis, returns enriched cart

3. User initiates checkout
   POST /api/v1/checkout  →  Order Service:
     a. Fetches cart from Redis
     b. gRPC → Catalog.CheckPrices   (validates current prices + stock)
     c. Creates Order (status: "pending") in PostgreSQL
     d. gRPC → Payment.CreateCheckoutSession  (Stripe hosted page URL)
     e. Returns { order_id, payment_url } to frontend

4. User pays on Stripe hosted page

5. Stripe webhook fires
   POST /api/v1/payment/webhook  →  Payment Service:
     a. Verifies Stripe signature
     b. Marks payment "success"
     c. Writes OutboxEvent — atomically with payment update

6. OutboxWorker (5s poll)
   Publishes { order_id, user_id, status: "paid" } to RabbitMQ

7. Order.PaymentConsumer receives event
     a. Updates order status → "paid"
     b. gRPC → Catalog.DecreaseInventory  (deducts stock per variant)
     c. Generates PDF invoice (gofpdf) → uploads to MinIO
     d. Clears user's Redis cart

8. User downloads invoice
   GET /api/v1/orders/:id/invoice  →  15-minute presigned MinIO URL
```

---

## Frontend — Buyer App

Built with Next.js 16 App Router, TypeScript strict mode, Tailwind CSS v4 (oklch color space), and shadcn/ui components.

**15 Routes:**

| Route | Type | Description |
|-------|------|-------------|
| `/` | Server Component | Hero, featured categories, product grid, CTA |
| `/products` | Client Component | Live search (debounced), category/price/stock filters |
| `/products/[id]` | Client Component | Image gallery, variant selector, add to cart |
| `/cart` | Client Component | Cart items, quantity controls, order summary |
| `/checkout` | Client Component | Shipping form, → Stripe redirect |
| `/checkout/success` | Client Component | Clears local cart, shows order ID |
| `/checkout/cancel` | Client Component | Back-to-cart reassurance |
| `/orders` | Client Component | Order list with status badges |
| `/orders/[id]` | Client Component | Order detail, items, invoice download, cancel |
| `/profile` | Client Component | Edit name/phone, manage saved addresses |
| `/auth/login` | Client Component | Email/password form + Google OAuth button |
| `/auth/register` | Client Component | Registration form with password strength meter |
| `/auth/verify` | Client Component | 6-digit OTP with auto-focus, paste support, 60s resend |
| `/auth/callback` | Client Component | OAuth landing — extracts JWT from query params |

**Key Frontend Patterns:**
- **Zustand stores** (`useAuthStore`, `useCartStore`) persisted to `localStorage`
- **Axios interceptors** — 401 response auto-triggers `refreshAccessToken()` and retries
- **Search autocomplete** — 200ms debounced suggestions from Elasticsearch
- **Dark mode only** — oklch color palette, glassmorphism, Inter + Outfit fonts

---

## Local Setup Guide

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | ≥ 1.26 | Backend services |
| Node.js | ≥ 20 | Frontend |
| Docker & Docker Compose | Latest | Infrastructure |
| Stripe CLI | Latest | Webhook forwarding |
| OpenSSL | Any | RSA key generation |

---

### Step 1 — Start Infrastructure

```bash
# Clone the repository
git clone https://github.com/<your-username>/ecommerce-project.git
cd ecommerce-project

# Spin up PostgreSQL, Redis, RabbitMQ, MinIO, Elasticsearch
make infra-up

# Wait ~30 seconds for Elasticsearch to be healthy
docker compose -f deploy/docker-compose.yml ps
```

Infrastructure UIs available at:
- **RabbitMQ Management**: http://localhost:15672 (admin/password)
- **MinIO Console**: http://localhost:9001 (admin/password)
- **Elasticsearch**: http://localhost:9200

---

### Step 2 — Create MinIO Buckets

Open the MinIO Console at http://localhost:9001 and create two buckets:
- `ecommerce-images` — for product images
- `ecommerce-invoices` — for PDF invoices

Or via CLI:
```bash
# Install mc (MinIO client) then:
mc alias set local http://localhost:9000 admin password
mc mb local/ecommerce-images
mc mb local/ecommerce-invoices
```

---

### Step 3 — Generate RSA Keys for Auth Service

```bash
cd services/auth/cmd/secrets

# Generate RSA private key
openssl genrsa -out private.pem 2048

# Extract public key
openssl rsa -in private.pem -pubout -out public.pem

cd ../../../../
```

---

### Step 4 — Configure Environment Variables

Each service has an `.env.example` file. Copy and fill in each one:

```bash
# Auth Service
cp services/auth/.env.example services/auth/.env

# Catalog Service
cp services/catalog/.env.example services/catalog/.env

# Order Service
cp services/order/.env.example services/order/.env

# Payment Service
cp services/payment/.env.example services/payment/.env

# Search Service
cp services/search/.env.example services/search/.env

# Email Service
cp services/email/.env.example services/email/.env

# Media Service
cp services/media/.env.example services/media/.env

# Frontend
cp web/buyer-app/.env.example web/buyer-app/.env.local
```

> **Google OAuth**: Create a project at [console.cloud.google.com](https://console.cloud.google.com), enable Google+ API, create OAuth 2.0 credentials, and add `http://localhost:8080/api/v1/auth/google/callback` as an authorized redirect URI.

> **Stripe**: Create a [Stripe](https://stripe.com) account, get your test secret key from the dashboard.

---

### Step 5 — Install Frontend Dependencies

```bash
cd web/buyer-app
npm install
cd ../../
```

---

### Step 6 — Run All Services

Open **8 terminal tabs** and run each service (order matters — Auth must start first so other services can fetch its public key):

```bash
# Tab 1 — Auth Service (start first)
make auth

# Tab 2 — Email Service
make email

# Tab 3 — Catalog Service
make catalog

# Tab 4 — Media Service
make media

# Tab 5 — Order Service
make order

# Tab 6 — Payment Service
make payment

# Tab 7 — Search Service
make search

# Tab 8 — Frontend
make frontend
```

Frontend is available at: **http://localhost:3000**

---

### Step 7 — Stripe Webhook (for payment testing)

In a separate terminal:
```bash
# Install Stripe CLI: https://stripe.com/docs/stripe-cli
make stripe
# or: stripe listen --forward-to localhost:8085/api/v1/payment/webhook
```

Use card `4242 4242 4242 4242` (any future date, any CVC) on the Stripe test checkout page.

---

### Optional — Build Verification

```bash
# Build all Go services from workspace root
make build

# Vet all Go code
make vet
```

---

### Quick Reset

```bash
# Wipe all data and restart infrastructure fresh
make infra-fresh
```

---

## Environment Variables Reference

### Auth Service (`services/auth/.env`)

| Variable | Description |
|----------|-------------|
| `POSTGRES_DSN` | PostgreSQL connection string for `auth_db` |
| `PRIVATE_PEM_PATH` | Path to RSA private key (e.g., `./cmd/secrets/private.pem`) |
| `PUBLIC_PEM_PATH` | Path to RSA public key |
| `EMAIL_SERVICE_BASE_URL` | Internal URL for Email service |
| `GOOGLE_CLIENT_ID` | Google OAuth 2.0 client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth 2.0 client secret |
| `REDIRECT_URL` | OAuth callback URL (must match Google Console) |
| `FRONTEND_URL` | Frontend base URL for post-OAuth redirect |
| `RABBITMQ_URL` | RabbitMQ AMQP connection string |

### Order Service (`services/order/.env`)

| Variable | Description |
|----------|-------------|
| `AUTH_SERVICE_URL` | Auth service base URL (for public key fetch) |
| `DATABASE_DSN` | PostgreSQL connection string for `order_db` |
| `REDIS_DSN` | Redis connection string |
| `RABBIT_MQ_URL` | RabbitMQ AMQP connection string |
| `CATALOG_GRPC_URL` | Catalog gRPC address (`localhost:50051`) |
| `S3_ENDPOINT` | MinIO endpoint |
| `S3_ACCESS_KEY` | MinIO access key |
| `S3_SECRET_KEY` | MinIO secret key |
| `S3_BUCKET_NAME` | Bucket name for invoices |

### Frontend (`web/buyer-app/.env.local`)

| Variable | Description |
|----------|-------------|
| `NEXT_PUBLIC_AUTH_API_URL` | Auth service URL (`http://localhost:8080`) |
| `NEXT_PUBLIC_API_URL` | Catalog service URL (`http://localhost:8082`) |
| `NEXT_PUBLIC_ORDER_API_URL` | Order service URL (`http://localhost:8084`) |
| `NEXT_PUBLIC_SEARCH_API_URL` | Search service URL (`http://localhost:8086`) |

---

## API Reference

### Auth Service `:8080`
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | Register new user |
| `POST` | `/api/v1/auth/verify-otp` | Verify email OTP |
| `POST` | `/api/v1/auth/login` | Login → returns JWT + sets refresh cookie |
| `POST` | `/api/v1/auth/refresh` | Refresh access token |
| `POST` | `/api/v1/auth/logout` | Revoke refresh token |
| `GET`  | `/api/v1/auth/google/login` | Initiate Google OAuth |
| `GET`  | `/api/v1/auth/google/callback` | OAuth callback handler |
| `GET`  | `/api/v1/auth/public-key` | RSA public key (PEM) |
| `GET`  | `/api/v1/auth/profile` | Get current user profile |
| `PUT`  | `/api/v1/auth/profile` | Update profile |
| `POST` | `/api/v1/auth/address` | Add saved address |
| `GET`  | `/api/v1/auth/address` | List saved addresses |

### Catalog Service `:8082`
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET`  | `/api/v1/catalog/products` | List products (paginated) |
| `GET`  | `/api/v1/catalog/products/:id` | Get product detail |
| `POST` | `/api/v1/catalog/products` | Create product (seller) |
| `PUT`  | `/api/v1/catalog/products/:id` | Update product |
| `DELETE` | `/api/v1/catalog/products/:id` | Delete product |
| `GET`  | `/api/v1/catalog/categories` | List categories |
| `GET`  | `/api/v1/catalog/sellers` | List sellers |

### Order Service `:8084`
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/cart/add` | Add item to cart |
| `GET`  | `/api/v1/cart` | Get current cart |
| `PUT`  | `/api/v1/cart/update` | Update item quantity |
| `DELETE` | `/api/v1/cart/remove` | Remove item from cart |
| `DELETE` | `/api/v1/cart/clear` | Clear entire cart |
| `POST` | `/api/v1/checkout` | Initiate checkout → Stripe URL |
| `GET`  | `/api/v1/orders` | List user's orders |
| `GET`  | `/api/v1/orders/:id` | Get order detail |
| `PATCH` | `/api/v1/orders/:id/cancel` | Cancel pending order |
| `GET`  | `/api/v1/orders/:id/invoice` | Download invoice (presigned URL) |

### Search Service `:8086`
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET`  | `/api/v1/search/products` | Full-text + filter search |

Query params: `q`, `category`, `min_price`, `max_price`, `in_stock`, `page`, `size`

---

## Infrastructure

All infrastructure runs via Docker Compose (`deploy/docker-compose.yml`).

| Container | Image | Port(s) | Data Volume |
|-----------|-------|---------|------------|
| `nexus_postgres` | postgres:15-alpine | 5432 | `postgres_data` |
| `nexus_redis` | redis:7-alpine | 6379 | `redis_data` |
| `nexus_rabbitmq` | rabbitmq:3-management-alpine | 5672, 15672 | `rabbitmq_data` |
| `nexus_minio` | minio/minio | 9000, 9001 | `minio_data` |
| `nexus_elasticsearch` | elasticsearch:8.17.0 | 9200 | `elasticsearch_data` |

### PostgreSQL Databases (all in `nexus_postgres`)

| Database | Service |
|----------|---------|
| `auth_db` | Auth Service |
| `catalog_db` | Catalog Service |
| `order_db` | Order Service |
| `payment_db` | Payment Service |

Each service owns its own database — no cross-service DB queries. All cross-service data access goes through gRPC or events.

---

## Architectural Decisions & Trade-offs

| Decision | Rationale |
|----------|-----------|
| **Go monorepo with `go.work`** | Shared `pkg/` code without publishing packages; workspace-level builds and linting |
| **Per-service PostgreSQL DB** | True data isolation; enforces service boundaries; prevents shared schema coupling |
| **Cart in Redis, not PostgreSQL** | Ephemeral data; sub-ms reads; TTL-based expiry; avoids table-per-user row churn |
| **gRPC for CheckPrices/DecreaseInventory** | Synchronous operations that must not fail silently during checkout |
| **RabbitMQ + Outbox for Payment→Order** | Async, durable, at-least-once delivery without coupling payment to order availability |
| **Search as separate service** | Decouples ES from Catalog; Catalog stays fast even if ES is slow; independent scaling |
| **RS256 JWT** | Asymmetric — services can verify without the private key; public key distribution is safe |
| **NanoID public IDs** | Short, URL-safe, non-enumerable; UUID v7 for DB (time-ordered, efficient B-tree indexing) |
| **Outbox polling at 5s** | Simple, no saga orchestrator needed; acceptable latency for order confirmation emails |

---

<div align="center">

Built with Go 1.26 · Next.js 16 · PostgreSQL · Redis · RabbitMQ · Elasticsearch · Stripe · MinIO

</div>