# E-Commerce Microservices Platform

> **Note:** This project is currently in active development. The architecture and features are continuously evolving. Please note that this repository focuses strictly on the backend infrastructure; a frontend client is not yet developed.

A backend infrastructure for an e-commerce platform built with Go. This project uses a microservices architecture, relying on gRPC for internal service-to-service communication and RabbitMQ for asynchronous, event-driven workflows.

## Tech Stack

  * **Language:** Go (Golang)
  * **API Framework:** Gin (REST), gRPC (RPC)
  * **Message Broker:** RabbitMQ
  * **Database:** PostgreSQL (with GORM)
  * **Payments:** Stripe API
  * **Authentication:** JWT (JSON Web Tokens)

-----

## Services Overview

The platform is divided into independent microservices, each managing its own domain and database:

  * **Auth Service:** Handles user registration, JWT generation (with user\_id claims), and triggers email verification workflows.
  * **Catalog Service:** Manages product inventory, variant availability, and price verification during checkout.
  * **Order Service:** Manages the user's shopping cart and order lifecycle. Communicates with the Payment service via gRPC to initiate checkout sessions.
  * **Payment Service:** Integrates with Stripe for processing payments. Listens for Stripe webhooks and securely records transactions.
  * **Email Service:** Consumes events to send out asynchronous notifications (like OTPs and order confirmations).

-----

## Key Architecture Patterns

  * **Event-Driven Communication:** Services communicate state changes asynchronously via RabbitMQ exchanges and queues (e.g., publishing an OrderPaid event).
  * **Transactional Outbox Pattern:** To ensure zero data loss during network failures, the Payment service uses the Outbox pattern. Database state updates (marking an order paid) and event publishing are handled atomically using a local outbox table and a dedicated background worker.
  * **Database per Service:** Each microservice maintains its own isolated PostgreSQL database (e.g., order\_db, payment\_db, auth\_db) to prevent tight coupling.

-----

## Roadmap & Future Features

This platform is actively expanding. Currently targeted features for upcoming releases include:

  * **Seller Order Approvals:** A dedicated workflow for sellers to review, approve, and manage fulfillment for incoming orders.
  * **Logistics & Delivery Assignment:** A system for delivery personnel to take up assigned orders, track routes, and update real-time delivery statuses.
  * **Automated PDF Invoicing:** Auto-generating PDF bills upon successful payment and dispatching them to customers via the Email service.

-----

## Getting Started

### Prerequisites

  * [Go 1.21+](https://www.google.com/search?q=https://golang.org/dl/)
  * [Docker & Docker Compose](https://www.google.com/search?q=https://www.docker.com/) (for running Postgres and RabbitMQ)
  * [Stripe CLI](https://www.google.com/search?q=https://stripe.com/docs/stripe-cli) (for testing webhooks locally)

### Local Setup

**1. Clone the repository:**

```bash
git clone https://github.com/Pranay-Kamble/ecommerce-proj
cd ecommerce-project
```

**2. Start the infrastructure:**
Use Docker to spin up PostgreSQL and RabbitMQ.

```bash
docker-compose up -d
```

**3. Initialize the Databases:**
Run the provided `deploy/init.sql` script in your local Postgres instance to create the necessary databases (`buyer_db`, `seller_db`, `payment_db`, `order_db`, etc.).

**4. Configure Environment Variables:**
Create a `.env` file in the root (or within each service) with your configuration. You will need:

  * Database connection strings
  * RabbitMQ URL (e.g., `amqp://admin:password@localhost:5672/`)
  * Stripe Secret Key and Webhook Signing Secret (`WEBHOOK_SECRET_KEY`)
  * JWT Secret

**5. Run the Stripe Webhook Listener:**
To test payments locally, forward Stripe events to your local Payment service:

```bash
stripe listen --forward-to localhost:8085/api/v1/payment/webhook
```

**6. Start the Services:**
Run the core services from their respective directories:

```bash
go run services/auth/cmd/server/main.go
go run services/payment/cmd/server/main.go
go run services/order/cmd/server/main.go
```
