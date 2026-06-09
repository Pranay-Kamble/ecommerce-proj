# ============================================================
# Nexus Ecommerce — Makefile
# ============================================================

# Infrastructure
infra-up:
	cd deploy && docker compose up -d

infra-down:
	cd deploy && docker compose down

infra-fresh:
	cd deploy && docker compose down -v && docker compose up -d

# ────────────────────────────────────────────
# Run each Go service (from its own directory)
# ────────────────────────────────────────────
auth:
	cd services/auth && go run cmd/server/main.go

catalog:
	cd services/catalog && go run cmd/server/main.go

order:
	cd services/order && go run cmd/server/main.go

payment:
	cd services/payment && go run cmd/server/main.go

email:
	cd services/email && go run cmd/server/main.go

media:
	cd services/media && go run cmd/server/main.go

search:
	cd services/search && go run cmd/server/main.go

# ────────────────────────────────────────────
# Frontend
# ────────────────────────────────────────────
frontend:
	cd web/buyer-app && npm run dev

# ────────────────────────────────────────────
# Build all Go services (workspace-level)
# ────────────────────────────────────────────
build:
	go build ./...

vet:
	go vet ./...

# ────────────────────────────────────────────
# Stripe webhook forwarding (for local payment testing)
# ────────────────────────────────────────────
stripe:
	stripe listen --forward-to localhost:8085/api/v1/payment/webhook

.PHONY: infra-up infra-down infra-fresh auth catalog order payment email media search frontend build vet stripe
