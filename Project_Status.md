# Ecommerce Project — Complete Overview

## 🏗️ Architecture: Go Microservices Monorepo

This is a **production-grade, microservices-based e-commerce backend** written in Go, organized as a single Go workspace (`go.work`). The architecture follows **Clean Architecture / DDD principles** with a strict layered structure inside every service.

---

## 📁 Top-Level Structure

```
ecommerce-project/
├── services/          # All backend microservices
├── pkg/               # Shared libraries (used by all services)
├── api/               # API contracts (proto + openapi) — currently empty
├── web/               # Frontend apps (buyer, seller, delivery)
├── deploy/            # Docker Compose infrastructure
├── go.work            # Go workspace file linking all modules
```

---

## 🐳 Infrastructure (`deploy/docker-compose.yml`)

Four services are containerized and ready to run:

| Container | Image | Port(s) | Purpose |
|---|---|---|---|
| `nexus_postgres` | postgres:15-alpine | 5432 | Primary relational DB |
| `nexus_redis` | redis:7-alpine | 6379 | OTP storage / caching |
| `nexus_rabbitmq` | rabbitmq:3-management | 5672, 15672 | Async messaging / events |
| `nexus_minio` | minio/minio | 9000, 9001 | File storage (fake S3) |

> All containers use named Docker volumes for data persistence.

---

## 📦 Shared Packages (`pkg/`)

| Package | Files | Description |
|---|---|---|
| `pkg/database` | `postgres.go`, `redis.go`, `database.go`, `postgres_test.go` | GORM Postgres + Redis connection wrappers |
| `pkg/broker` | `rabbitmq.go` | RabbitMQ client (declare exchange, queue, bind, publish/consume) |
| `pkg/logger` | `logger.go` | Zap-based structured logger with env-aware init |
| `pkg/utils` | *(empty)* | Placeholder |
| `pkg/protobufs/catalog` | `catalog.proto`, generated `.pb.go` files | gRPC contract for Catalog service |
| `pkg/protobufs/payment` | `payment.proto`, generated `.pb.go` files | gRPC contract for Payment service |

---

## 🚀 Microservices (`services/`)

### ✅ 1. Auth Service — **FULLY IMPLEMENTED**
**Port:** `8080` | **DB:** PostgreSQL (`auth_db`) + Redis

**Architecture layers:**
- `domain/` — `User` (nanoid PK, email/password/role/provider, isVerified, isOnboarded), `Token` (refresh token)
- `handler/` — Gin HTTP handlers with full Swagger godoc annotations
- `service/` — Business logic (AuthService interface)
- `repository/` — UserRepository, TokenRepository, OTPRepository
- `workers/` — `user_consumer.go` (RabbitMQ consumer listening for `seller.onboarded` events)
- `client/` — `email_client.go` (HTTP client to Email service)
- `utils/` — RSA key loading, JWT generation, SHA256 hashing, OTP generator, OAuth config

**Auth Endpoints (`/api/v1/auth/`):**
| Method | Path | Description |
|---|---|---|
| POST | `/register` | Register buyer/seller/logistic + async OTP email |
| POST | `/login` | Validate credentials → JWT + HttpOnly refresh cookie |
| POST | `/refresh` | Rotate refresh token (family-based theft detection) |
| POST | `/logout` | Revoke token family + clear cookie |
| POST | `/verify` | Verify email OTP → issue tokens |
| POST | `/resend-otp` | Resend OTP if unverified |
| GET | `/google/login` | Initiate Google OAuth2 redirect |
| GET | `/google/callback` | Handle OAuth callback → issue tokens |
| GET | `/public-key` | Expose RSA public key for downstream JWT verification |
| GET | `/swagger/*` | Swagger UI |

**Key Design Decisions:**
- Refresh tokens are **hashed (SHA256) before storage** — never stored in plain text
- **Token family rotation** — using/re-using an old token invalidates the entire family (reuse detection)
- JWT signed with **RSA private key**; other services verify with the public key endpoint
- Email sending is done **asynchronously** in a goroutine
- Google OAuth uses CSRF state stored in a cookie

---

### ✅ 2. Catalog Service — **FULLY IMPLEMENTED**
**Port:** *(not set in env)* | **DB:** PostgreSQL (`catalog_db`)

**Architecture layers:**
- `domain/` — `Product`, `Variant`, `Category`, `Seller`, `Image`
- `handler/` — Handlers for products, variants, categories, sellers + gRPC handler
- `service/` — `ProductService`, `VariantService`, `CategoryService`, `SellerService`
- `repository/` — Repos for all 4 entities

**Catalog Endpoints (`/api/v1/catalog/`):**
| Access | Method | Path | Description |
|---|---|---|---|
| Public | GET | `/categories` | List all categories |
| Public | GET | `/categories/:id/breadcrumbs` | Get category breadcrumbs |
| Public | GET | `/products` | List products |
| Public | GET | `/products/:id` | Get product by ID |
| Public | GET | `/variants/:sku` | Get variant by SKU |
| Authenticated | POST | `/sellers` | Register as seller |
| Authenticated | GET | `/sellers/me` | Get seller profile |
| Seller-only | POST | `/seller/products` | Create product |
| Seller-only | PUT | `/seller/products/:id` | Update product |
| Seller-only | DELETE | `/seller/products/:id` | Delete product |
| Seller-only | POST | `/seller/products/:id/variants` | Create variant |
| Seller-only | PUT | `/seller/variants/:id` | Update variant |
| Seller-only | DELETE | `/seller/variants/:id` | Delete variant |

**Key Design Decisions:**
- Products have **dual IDs**: internal UUID (v7) + public NanoID with prefix `itm_`
- Sellers have prefix `sel_`, Categories have `cat_`, Variants have `var_`
- Category supports **hierarchical paths** using PostgreSQL `ltree` extension
- Product images stored as **JSONB array** in Postgres
- Variant specifications stored as **JSONB with GIN index** for flexible filtering
- Middleware verifies JWT via RSA public key; seller routes require `RequireSeller()` middleware
- **gRPC handler** exists for internal service-to-service calls

---

### ✅ 3. Order Service — **MOSTLY IMPLEMENTED**
**Port:** *(not set)* | **DB:** PostgreSQL + Redis (cart)

**Architecture layers:**
- `domain/` — `Order`, `OrderItem`, `Customer`, `Cart`
- `handler/` — `CartHandler`, `CustomerHandler`, `OrderHandler`
- `service/` — `CartService`, `CustomerService`, `OrderService`
- `repository/` — matching repos
- `workers/` — `payment_consumer.go` (listens for payment status updates)
- `client/` — `payment_client.go` (HTTP/gRPC client to Payment service)

**Order Endpoints (`/api/v1/`) — all require JWT:**
| Method | Path | Description |
|---|---|---|
| GET | `/cart` | Get current cart |
| POST | `/cart/add` | Add item to cart |
| DELETE | `/cart/remove/:product_id` | Remove item from cart |
| DELETE | `/cart` | Clear cart |
| GET | `/profile` | Get customer profile |
| POST | `/profile` | Create customer profile |
| POST | `/profile/addresses` | Add address |
| POST | `/checkout` | Create order → triggers payment |
| GET | `/orders/:public_id` | Get single order |
| GET | `/orders` | Get all user orders |

**Key Design Decisions:**
- Cart is stored in **Redis** (ephemeral, session-based)
- Checkout creates an `Order` and calls Payment service to get a Stripe checkout URL
- Orders use **Outbox pattern** for reliable event publishing (payment status updates via RabbitMQ)

---

### ✅ 4. Payment Service — **CORE IMPLEMENTED**
**DB:** PostgreSQL | **Gateway:** Stripe

**Architecture layers:**
- `domain/` — `Payment` (UUID v7 + NanoID public ID), `OutboxEvent`
- `handler/` — `WebhookHandler` (Stripe webhook), `gRPC handler`
- `service/` — `PaymentService` (create Stripe checkout session, mark success)
- `repository/` — PaymentRepository
- `workers/` — `outbox_worker.go` (polls outbox table, publishes to RabbitMQ)

**Endpoints:**
| Method | Path | Description |
|---|---|---|
| POST | `/webhook` | Stripe webhook receiver |
| gRPC | internal | `CreateCheckoutSession` |

**Key Design Decisions:**
- Uses **Stripe Checkout** for payment (hosted payment page)
- Webhook validates `Stripe-Signature` header for security
- **Outbox pattern** ensures payment events are reliably published even if RabbitMQ is temporarily down
- Amount stored in **paise** (1/100 of INR)

---

### ✅ 5. Email Service — **IMPLEMENTED (Simple)**
**Port:** `8081`

- Sends **verification emails** with OTP via SMTP
- Has HTML email `templates/`
- Exposes `POST /api/v1/email/verify`
- Used by Auth service as an internal HTTP client

---

### ✅ 6. Media Service — **IMPLEMENTED**
**Storage:** MinIO (S3-compatible)

**Architecture layers:**
- `handler/` — Upload handler with JWT middleware
- `service/` — MinIO interaction
- `storage/` — MinIO client wrapper
- `utils/` — Helpers

Used for product image uploads (referenced by Catalog service).

---

### 🔲 7. Cart Service — **SCAFFOLDED**
Has directory structure (`domain/`, `handler/`, `repository/`, `service/`) but `go.mod` shows minimal/empty dependencies. Functionality is embedded inside the Order service.

### 🔲 8. Search Service — **SCAFFOLDED**
Directory structure exists, no implementation yet.

### 🔲 9. Logistics Service — **SCAFFOLDED**
Directory structure exists, no implementation yet.

### 🔲 10. ML-Core Service — **DIRECTORY ONLY**
Listed in `/services/` but not in `go.work`. Not started.

---

## 🌐 Frontend (`web/`)

Three frontend apps are scaffolded:
- `web/buyer-app/` — Customer shopping app
- `web/seller-dashboard/` — Seller management portal
- `web/delivery-app/` — Delivery partner app

**Status:** Directory-level scaffolding only — no actual frontend code confirmed.

---

## 🔗 Inter-Service Communication

```
Auth ──────────────────► Email (HTTP REST)
         sends OTP via HTTP client

Order ─────────────────► Payment (gRPC)
         creates checkout session

Payment ────────────────► Order (RabbitMQ: payment events via Outbox)
         publishes payment.success events

Auth ────────────────────► Auth (RabbitMQ: listens for seller.onboarded)
         updates isOnboarded flag on user

Catalog ─────────────────► other services (gRPC server exposed)
         internal product data queries
```

---

## 🛠️ Tech Stack Summary

| Concern | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP Framework | Gin |
| ORM | GORM |
| Primary DB | PostgreSQL |
| Cache/OTP | Redis |
| Message Broker | RabbitMQ (AMQP) |
| File Storage | MinIO (S3-compatible) |
| Payment Gateway | Stripe |
| Inter-service RPC | gRPC + Protocol Buffers |
| JWT | RS256 (RSA key pair) |
| OAuth2 | Google OAuth 2.0 |
| Logging | Uber Zap |
| ID Generation | NanoID + UUID v7 |
| API Docs | Swagger (swaggo) |
| Containerization | Docker Compose |

---

## 📋 What's Done vs. What's Pending

### ✅ Done
- Full auth flow (register, email OTP, login, Google OAuth, refresh token rotation with reuse detection, logout)
- RSA JWT-based authentication shared across services
- Full catalog service (products, variants, categories, seller profiles)
- Order flow (cart → checkout → order creation → payment session)
- Stripe payment integration with webhook handling
- Outbox pattern for reliable payment→order event delivery
- Email service for OTP delivery
- Media service for file uploads (MinIO)
- RabbitMQ event bus (seller onboarded, payment events)
- Swagger docs on Auth + Catalog services
- Docker Compose infrastructure

### 🔲 Pending / Next Up
- **Search Service** — product search (likely ElasticSearch/Typesense)
- **Logistics Service** — delivery partner assignment, tracking
- **ML-Core Service** — recommendations, personalization
- **Frontend Apps** — buyer app, seller dashboard, delivery app
- **API Gateway** — centralized auth enforcement (noted in `Future Refactors`)
- **Structured error types** — replace string-based error checking (noted in `Future Refactors`)
- `api/proto` and `api/openapi` directories are currently empty
- Cart service as standalone (currently embedded in Order service)
- Admin panel

---

## 🔑 Key Patterns You're Using

1. **Clean Architecture** — Handler → Service → Repository, domain is framework-free
2. **Token Family Rotation** — Refresh token reuse detection via family IDs
3. **Outbox Pattern** — Reliable event publishing from Payment service
4. **Dual IDs** — Internal UUIDs for DB relations, public NanoIDs for API exposure
5. **Async Email Sending** — Goroutine-based, non-blocking
6. **Shared RSA Key** — Auth generates JWT with private key; other services verify with public key endpoint
