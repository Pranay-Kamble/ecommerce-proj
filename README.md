<br># Nexus — Production-Grade E-Commerce Platform

<div align="center">

![Nexus Banner](./All%20Services.excalidraw.png)

**A full-stack, microservices-based e-commerce platform.**  
Go backend · Next.js 16 frontend · Stripe payments · Elasticsearch · gRPC · RabbitMQ

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-16.2-black?logo=next.js)](https://nextjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?logo=postgresql)](https://postgresql.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

</div>

---

## ✨ What Is This?

**Nexus** is a production-grade e-commerce backend. It demonstrates advanced backend architecture patterns — clean architecture, the outbox pattern, gRPC inter-service communication, event-driven messaging, and JWT + OAuth2 authentication — all tied together with a polished Next.js buyer frontend.

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    Next.js Buyer App                     │
│           (App Router · Tailwind v4 · Zustand)           │
└─────────────────────┬────────────────────────────────────┘
                      │ REST (Axios + JWT Bearer)
        ┌─────────────┼──────────────────────────────┐
        │             │                              │
  ┌─────▼─────┐ ┌─────▼──────┐ ┌───────────┐ ┌─────▼──────┐
  │   Auth    │ │  Catalog   │ │   Order   │ │   Search   │
  │  :8080    │ │  :8082     │ │  :8084    │ │  :8086     │
  └─────┬─────┘ └─────┬──────┘ └─────┬─────┘ └────────────┘
        │       gRPC  │        gRPC  │  RabbitMQ
        │       :50051│        :50052│
  ┌─────▼─────┐ ┌─────▼──────┐ ┌─────▼─────┐
  │  Email    │ │  Payment   │ │  Media    │
  │  :8081    │ │  :8085     │ │  :8083    │
  └───────────┘ └─────┬──────┘ └───────────┘
                      │ Outbox → RabbitMQ
                ┌─────▼──────────────┐
                │   Infrastructure   │
                │  Postgres · Redis  │
                │  RabbitMQ · MinIO  │
                │  Elasticsearch     │
                └────────────────────┘
```

### Key Flows
| Flow | Path |
|---|---|
| **Google OAuth** | Frontend → Auth (redirect) → Google → Auth Callback → Frontend `/auth/callback?jwt=...` |
| **Checkout** | Cart (Redis) → Catalog gRPC price check → Order created → Payment gRPC → Stripe URL |
| **Payment Confirmed** | Stripe webhook → Payment → OutboxWorker → RabbitMQ → Order consumer → DecreaseInventory gRPC → Clear cart |
| **Product Search** | Catalog publishes events → RabbitMQ → Search consumer → Elasticsearch index |
| **Invoice** | Payment confirmed → PDF generated via gofpdf → Uploaded to MinIO → Presigned URL returned |

---

## 🔧 Tech Stack

| Concern | Technology |
|---|---|
| **Language** | Go 1.26 |
| **HTTP Framework** | Gin |
| **ORM** | GORM with Go generics (`gorm.G[T]`) |
| **Primary DB** | PostgreSQL 15 |
| **Cache / Cart** | Redis 7 |
| **Message Broker** | RabbitMQ 3 (AMQP 0.9.1) |
| **File Storage** | MinIO (S3-compatible) |
| **Search Engine** | Elasticsearch 8.17 |
| **Payment Gateway** | Stripe (Checkout Sessions + Webhooks) |
| **Inter-service RPC** | gRPC + Protocol Buffers |
| **Auth** | RS256 JWT + Google OAuth 2.0 |
| **ID Generation** | NanoID (public) + UUID v7 (internal) |
| **Logging** | Uber Zap |
| **PDF Generation** | gofpdf |
| **Frontend** | Next.js 16.2 (App Router) |
| **UI Components** | shadcn/ui |
| **CSS** | Tailwind CSS v4 (oklch color space) |
| **State Management** | Zustand (localStorage-persisted) |
| **HTTP Client** | Axios with JWT interceptors |

---

## ✅ Features

### Backend
- **Auth Service** — Email/password registration with 6-digit OTP verification, RS256 JWT, refresh token rotation, Google OAuth 2.0
- **Catalog Service** — Products, variants, categories, sellers; gRPC server for price checking and inventory management
- **Order Service** — Redis cart management, full checkout flow, order lifecycle (pending → paid → shipped → delivered), PDF invoice generation via MinIO
- **Payment Service** — Stripe Checkout Sessions, webhook handling, Transactional Outbox Pattern for guaranteed message delivery
- **Search Service** — Elasticsearch full-text search with price/category/stock filters; auto-indexed from catalog events via RabbitMQ
- **Email Service** — SMTP email sending via Resend API for OTP delivery
- **Media Service** — MinIO file upload for product images

### Frontend (Buyer App)
- Dark mode design with glassmorphism and oklch color palette
- Homepage with hero, featured products, and categories
- Product listing with Elasticsearch-powered search + autocomplete suggestions
- Product detail with variant selector and add-to-cart
- Cart with quantity management and backend sync
- Checkout form with Stripe redirect
- Order history with status badges
- Order detail with invoice download
- Profile management with saved addresses
- Google OAuth flow end-to-end

---

## 📁 Project Structure

```
ecommerce-project/
├── pkg/                    # Shared Go libraries
│   ├── database/           # Postgres + Redis connection helpers
│   ├── broker/             # RabbitMQ connection helper
│   ├── logger/             # Uber Zap logger
│   └── protobufs/          # Generated gRPC code (catalog, payment)
├── services/
│   ├── auth/               # Port 8080 — JWT, OAuth, OTP
│   ├── catalog/            # Port 8082 / gRPC 50051
│   ├── order/              # Port 8084 — Cart, Checkout, Orders, Invoices
│   ├── payment/            # Port 8085 / gRPC 50052 — Stripe
│   ├── email/              # Port 8081 — SMTP via Resend
│   ├── media/              # Port 8083 — MinIO uploads
│   └── search/             # Port 8086 — Elasticsearch
├── web/
│   └── buyer-app/          # Next.js 16 frontend
├── deploy/
│   ├── docker-compose.yml  # All infrastructure containers
│   └── init.sql            # Database initialization
├── Makefile                # Convenience commands
└── go.work                 # Go workspace
```

---

## 🚀 How to Run Locally

### Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | 1.26+ | [golang.org/dl](https://golang.org/dl/) |
| Node.js | 18+ | [nodejs.org](https://nodejs.org/) |
| Docker + Docker Compose | Latest | [docker.com](https://docker.com/) |
| Stripe CLI | Latest | For local webhook testing |

---

### Step 1 — Start Infrastructure

```bash
# From the project root
make infra-up

# This starts:
#   nexus_postgres      (5432)
#   nexus_redis         (6379)
#   nexus_rabbitmq      (5672, UI: 15672)
#   nexus_minio         (9000, UI: 9001)
#   nexus_elasticsearch (9200)
```

Databases are auto-created by `deploy/init.sql` on first start.

---

### Step 2 — Configure Environment Variables

Each service has a `.env` file. The defaults in the repo work with the Docker Compose setup above (password: `password`). The only secrets you need to provide yourself are:

| Service | File | Variable | Where to Get |
|---|---|---|---|
| Auth | `services/auth/.env` | `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET` | [console.cloud.google.com](https://console.cloud.google.com) |
| Auth | `services/auth/.env` | `REDIRECT_URL` | Must match Google OAuth authorized redirect URI |
| Auth | `services/auth/.env` | `FRONTEND_URL` | `http://localhost:3000` |
| Payment | `services/payment/.env` | `PAYMENT_GATEWAY_SECRET_KEY` | [dashboard.stripe.com](https://dashboard.stripe.com) |
| Payment | `services/payment/.env` | `WEBHOOK_SECRET_KEY` | From `stripe listen` (see Step 5) |
| Email | `services/email/.env` | `RESEND_API_KEY` | [resend.com](https://resend.com) |

**Auth service RSA keys** — generate once and place in `services/auth/cmd/secrets/`:
```bash
openssl genrsa -out services/auth/cmd/secrets/private.pem 2048
openssl rsa -in services/auth/cmd/secrets/private.pem -pubout -out services/auth/cmd/secrets/public.pem
```

---

### Step 3 — Run the Backend Services

> **Important:** The Auth service must start first — all other services fetch the RSA public key from it at startup.

Open a terminal for each service (or use a process manager like `tmux`):

```bash
# Terminal 1 — Auth (MUST start first)
make auth

# Terminal 2 — Email
make email

# Terminal 3 — Catalog
make catalog

# Terminal 4 — Media
make media

# Terminal 5 — Order
make order

# Terminal 6 — Payment
make payment

# Terminal 7 — Search
make search
```

| Service | Port | Health Check |
|---|---|---|
| Auth | 8080 | `curl http://localhost:8080/api/v1/auth/ping` |
| Email | 8081 | — |
| Catalog | 8082 | `curl http://localhost:8082/api/v1/catalog/ping` |
| Media | 8083 | — |
| Order | 8084 | — |
| Payment | 8085 | — |
| Search | 8086 | — |

---

### Step 4 — Run the Frontend

```bash
# Create the env file
cat > web/buyer-app/.env.local << 'EOF'
NEXT_PUBLIC_AUTH_API_URL=http://localhost:8080
NEXT_PUBLIC_API_URL=http://localhost:8082
NEXT_PUBLIC_ORDER_API_URL=http://localhost:8084
NEXT_PUBLIC_SEARCH_API_URL=http://localhost:8086
EOF

# Install and start
cd web/buyer-app
npm install
npm run dev
```

Frontend runs at **http://localhost:3000**

---

### Step 5 — Enable Stripe Webhooks (for payment testing)

```bash
# In a new terminal
make stripe
# Copy the webhook signing secret printed by the CLI
# Paste it as WEBHOOK_SECRET_KEY in services/payment/.env
# Then restart the payment service
```

**Test card:** `4242 4242 4242 4242` · Any future expiry · Any CVC

---

## 🔑 Architecture Patterns

### Clean Architecture (every service)
```
cmd/server/main.go      → Wire everything (DI)
internal/domain/        → Pure domain models (no framework deps)
internal/handler/       → HTTP + gRPC handlers, routes, middleware
internal/service/       → Business logic (interface-based)
internal/repository/    → Data access via GORM / Redis
internal/workers/       → Background consumers (RabbitMQ)
internal/client/        → gRPC / HTTP clients to other services
```

### Transactional Outbox Pattern
Payment service writes `OutboxEvent` atomically in the same DB transaction as the payment status update. A background `OutboxWorker` polls every 5s and publishes to RabbitMQ. This guarantees zero message loss even if the broker is temporarily unavailable.

### Dual ID Pattern
- **Internal**: UUID v7 (used in DB foreign keys, never exposed in API)
- **Public**: NanoID with prefix (`itm_`, `var_`, `sel_`, `cat_`, `ORD-`)

### JWT Auth Flow
Auth service signs JWTs with an RSA private key. All other services fetch the RSA public key from `GET /api/v1/auth/public-key` at startup and verify tokens locally — no auth service round-trip per request.

---

## 🧩 gRPC Contracts

### Catalog Service (`pkg/protobufs/catalog/catalog.proto`)
```protobuf
service CatalogService {
  rpc CheckPrices(CheckPricesRequest) returns (CheckPricesResponse);
  rpc DecreaseInventory(DecreaseInventoryRequest) returns (DecreaseInventoryResponse);
}
```

### Payment Service (`pkg/protobufs/payment/payment.proto`)
```protobuf
service PaymentService {
  rpc CreatePaymentSession(CreatePaymentRequest) returns (CreatePaymentResponse);
}
```

---

## 📡 API Reference

All REST endpoints are under `/api/v1/`.

| Service | Base URL | Notable Endpoints |
|---|---|---|
| Auth | `:8080/api/v1/auth` | `POST /register`, `POST /login`, `GET /google/login`, `POST /verify`, `POST /refresh` |
| Catalog | `:8082/api/v1/catalog` | `GET /products`, `GET /products/:id`, `POST /products`, `GET /categories` |
| Order | `:8084/api/v1` | `POST /cart/add`, `GET /cart`, `POST /checkout`, `GET /orders`, `GET /orders/:id/invoice` |
| Payment | `:8085/api/v1/payment` | `POST /webhook` (Stripe) |
| Search | `:8086/api/v1/search` | `GET /products?q=&category=&min_price=&max_price=&in_stock=` |
| Media | `:8083/api/v1/media` | `POST /upload` |

---

## 🎬 Demo

> 📹 **[Watch the demo video](#)** — *link to be added after recording*

### Full User Flow
1. Register with email → verify 6-digit OTP → login  
2. Browse products, search with autocomplete, filter by category/price  
3. Add items to cart → proceed to checkout → fill shipping form  
4. Complete Stripe test payment (`4242 4242 4242 4242`)  
5. View order history → download PDF invoice  
6. Edit profile and saved addresses  

---

## 🛠️ Development Notes

- **Go Workspace**: All services share `go.work` — run `go build ./...` from root to verify all services compile.
- **Elasticsearch**: Products are indexed automatically when the Catalog service publishes events to RabbitMQ. Start the Search service after Catalog to auto-index.
- **MinIO Setup**: Create a bucket named `ecommerce-images` in the MinIO console at `http://localhost:9001` (admin/password) before uploading images or generating invoices.
- **init.sql**: Only runs on the **first** Docker volume creation. If you add databases later, run them manually: `docker exec nexus_postgres psql -U admin -c "CREATE DATABASE newdb;"`.

---

## 📄 License

MIT © 2026 Pranay Kamble