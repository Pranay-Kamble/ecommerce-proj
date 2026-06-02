# 🗓️ 7-Day Resume Sprint Plan

> **Goal:** A fully demo-able, impressive ecommerce project for your resume.  
> **Constraint:** Web only. 1 week. Beginner frontend.

---

## 🎯 What Actually Impresses on a Resume

For a backend-heavy project, interviewers care about:
1. ✅ **Working end-to-end flow** — register → browse → cart → checkout → pay → order confirmed
2. ✅ **Microservices architecture** — already built, already impressive
3. ✅ **Search** — Elasticsearch shows real-world DB knowledge
4. ✅ **Payment integration** — Stripe is a real industry tool
5. ✅ **JWT + OAuth** — already done
6. ✅ **A live frontend** — so you can demo/record it
7. ✅ **PDF Invoice** — tiny effort, huge "wow" factor

**What does NOT matter for a resume project:**
- ❌ Logistics service (complex, not visible in demo)
- ❌ Delivery app (nobody will see it)
- ❌ Seller dashboard (mention it as "future scope")
- ❌ ML-Core (same)
- ❌ Admin panel

---

## ✂️ Final Scope: Keep vs. Cut

| Feature | Decision | Reason |
|---|---|---|
| Order Service (complete) | ✅ **MUST** | Core flow, already 75% done |
| Search (Elasticsearch) | ✅ **MUST** | Huge resume talking point |
| PDF Invoice | ✅ **MUST** | Quick win, very impressive in demo |
| Buyer Frontend | ✅ **MUST** | Makes the project demoable |
| Logistics Service | ❌ **SKIP** | Complex, not visible, not worth the week |
| Delivery App | ❌ **SKIP** | Web only rule |
| Seller Dashboard | ⚠️ **OPTIONAL** | Only add after Day 6 if time permits |
| ML-Core | ❌ **SKIP** | Not in scope ever |

---

## 📅 7-Day Daily Plan

---

### 🗓️ Day 1 — Finish Order Service Backend
**Goal:** Full checkout flow working end-to-end locally.

**Morning (3–4 hrs):**
- [ ] Create `services/order/cmd/server/main.go`
  - Wire Postgres + Redis
  - AutoMigrate all domain models
  - Dial Catalog gRPC + Payment gRPC clients
  - Start `PaymentConsumer` goroutine
  - Register routes, start Gin server
- [ ] Test: POST `/api/v1/cart/add` → GET `/api/v1/cart` → POST `/api/v1/checkout`

**Afternoon (2–3 hrs):**
- [ ] Add `DecreaseInventory` RPC to `catalog.proto` + implement in Catalog's `grpc_handler.go`
- [ ] Call it from `payment_consumer.go` after order marked as `paid`
- [ ] Add `GET /api/v1/orders/:id/invoice` endpoint stub (returns 404 for now, implement Day 4)
- [ ] Add Elasticsearch to `docker-compose.yml`

---

### 🗓️ Day 2 — Search Service (Elasticsearch)
**Goal:** `GET /api/v1/search/products?q=shirt&category=men&min_price=200` works.

**Morning (3–4 hrs):**
- [ ] Scaffold `services/search/`:
  - `go.mod` (`module ecommerce/services/search`)
  - Add to `go.work`
  - `internal/domain/document.go` — ES document struct
  - `internal/repository/es_repository.go` — wraps ES client
  - `internal/service/search_service.go`
  - `internal/handler/search_handler.go` + `routes.go`
  - `internal/workers/catalog_consumer.go`
  - `cmd/server/main.go`

**Afternoon (2–3 hrs):**
- [ ] Add product event publishing to Catalog service:
  - `product.created`, `product.updated`, `product.deleted` → `catalog_events` exchange
- [ ] Implement `catalog_consumer.go` to index/update/delete ES documents
- [ ] Implement `SearchHandler`: full-text + filter query DSL
- [ ] Test: add a product in catalog → see it appear in search

**ES Index design (use this):**
```json
title, brand, description, category_name, seller_name, price, in_stock, slug, public_id
```

---

### 🗓️ Day 3 — PDF Invoice + Order Service Polish
**Goal:** Download a real PDF invoice for a completed order.

**Morning (2–3 hrs):**
- [ ] `go get github.com/jung-kurt/gofpdf` in Order service
- [ ] Add `Invoice` domain model + migration
- [ ] Implement `InvoiceService.GenerateAndStore(ctx, order)`:
  - Build PDF: logo placeholder, invoice #, date, line items table, totals with GST
  - Upload to MinIO under `invoices/` bucket
  - Save `Invoice` record in DB
- [ ] Call it inside `payment_consumer.processMessage()` after `UpdateOrderStatus`

**Afternoon (2 hrs):**
- [ ] Implement `GET /orders/:public_id/invoice` → generate MinIO presigned URL → return download link
- [ ] End-to-end test: complete a Stripe test payment → invoice PDF downloads correctly
- [ ] Order service cleanup: add `PATCH /orders/:id/cancel` endpoint
- [ ] Test all order endpoints with Postman/curl

---

### 🗓️ Day 4 — Frontend Setup + Core Pages
**Goal:** Home page + product listing + product detail page live.

**Stack:** Next.js 14 (App Router) + Tailwind CSS + shadcn/ui

```bash
cd web/buyer-app
npx create-next-app@latest ./ --typescript --eslint --tailwind --app --src-dir --no-git
npx shadcn@latest init
npm install axios zustand react-hot-toast
```

**Pages to build today:**
- [ ] `app/page.tsx` — Homepage: hero banner + featured categories + product grid
- [ ] `app/products/page.tsx` — Product list with search bar + category filter sidebar
- [ ] `app/products/[id]/page.tsx` — Product detail: images, title, price, variants, "Add to Cart" button
- [ ] Reusable components: `ProductCard`, `Navbar`, `Footer`, `CategorySidebar`
- [ ] API utility: `lib/api.ts` (axios instance pointing to your backend)

**Design tip:** Use a clean, dark or white modern theme. shadcn/ui + Tailwind makes this fast.

---

### 🗓️ Day 5 — Auth + Cart + Checkout Frontend
**Goal:** Complete user auth flow + working cart + checkout redirects to Stripe.

**Morning (3–4 hrs) — Auth:**
- [ ] `app/auth/login/page.tsx` — email/password form
- [ ] `app/auth/register/page.tsx` — name/email/password/role form
- [ ] `app/auth/verify/page.tsx` — 6-digit OTP input
- [ ] Auth state: Zustand store holding JWT + user info
- [ ] Store JWT in `localStorage`, send as `Authorization: Bearer <token>` header
- [ ] Refresh token is auto-handled via the HttpOnly cookie

**Afternoon (3 hrs) — Cart + Checkout:**
- [ ] Cart state in Zustand (sync with backend on load)
- [ ] `app/cart/page.tsx` — items list, quantities, totals, "Proceed to Checkout" button
- [ ] `app/checkout/page.tsx` — shipping address form → POST `/api/v1/checkout` → redirect to `payment_url` (Stripe hosted page)
- [ ] `app/checkout/success/page.tsx` — "Payment successful! Your order is confirmed."
- [ ] `app/checkout/cancel/page.tsx` — "Payment cancelled. Return to cart."

---

### 🗓️ Day 6 — Orders + Polish + Search UI
**Goal:** Order history, invoice download, search connected to ES, responsive design.

**Morning (3 hrs):**
- [ ] `app/orders/page.tsx` — list all orders with status badges
- [ ] `app/orders/[id]/page.tsx` — order detail: items, status, shipping, "Download Invoice" button
- [ ] Connect search bar on product listing page to Search service
- [ ] Autocomplete suggestions on search input
- [ ] Profile page: view/edit name, phone, saved addresses

**Afternoon (2–3 hrs) — Polish:**
- [ ] Loading skeletons on product cards (shadcn `Skeleton` component)
- [ ] Toast notifications on add-to-cart, login success, errors (`react-hot-toast`)
- [ ] Mobile responsiveness check (use Chrome DevTools)
- [ ] Error states: empty cart, no orders, product not found
- [ ] Navbar: show login/logout based on auth state, cart item count badge

---

### 🗓️ Day 7 — Testing, README, and Demo Recording
**Goal:** Everything works, project is presentable on GitHub.

**Morning (2–3 hrs):**
- [ ] End-to-end test the full flow:
  1. Register new buyer → verify OTP → login
  2. Browse products → search "shirt" → filter by price
  3. Add to cart → checkout → complete Stripe test payment
  4. View order → download PDF invoice
- [ ] Fix any broken flows found during testing
- [ ] Update `docker-compose.yml` with all services

**Afternoon (2–3 hrs):**
- [ ] Write `README.md`:
  - Project description + architecture diagram (screenshot your Excalidraw)
  - Tech stack table
  - Features list
  - "How to run locally" (docker-compose up + env setup)
  - Demo GIF or link to video
- [ ] Record a 2–3 minute demo video (Loom is free)
- [ ] Put the demo video link in README

---

## 🏆 What Your Resume Bullet Points Will Look Like

```
• Built a microservices e-commerce backend in Go with 7 independent services
  (Auth, Catalog, Order, Payment, Search, Email, Media) communicating via
  REST, gRPC, and RabbitMQ event bus.

• Implemented JWT authentication with RSA key pairs, Google OAuth2 login,
  and refresh token rotation with family-based theft detection.

• Integrated Stripe Checkout for payments and Elasticsearch for full-text
  product search with category and price filtering.

• Built an async event pipeline using the Outbox pattern to ensure reliable
  payment→order status updates even during message broker downtime.

• Auto-generates PDF invoices (uploaded to MinIO) on payment completion,
  exposed via presigned download URLs.

• Developed a Next.js buyer storefront with complete shopping flow:
  browse → cart → checkout → order tracking → invoice download.
```

---

## ⚡ Cheat Codes to Move Fast

**On Frontend (you're a beginner — use these):**
1. **v0.dev** — type what UI you want, it gives you shadcn/Tailwind code. Copy-paste. Edit.
2. **shadcn/ui docs** — every component has copy-paste code. Don't write CSS from scratch.
3. **ChatGPT for React doubts** — "how do I make this API call on page load in Next.js App Router?"

**On Backend:**
1. Copy `main.go` structure from Auth service — 80% is identical (Postgres, Redis, Gin, env loading)
2. Copy `middleware.go` from Catalog or Order service for JWT verification in Search service

**For the Demo:**
- Use Stripe's test card: `4242 4242 4242 4242`, any future date, any CVC
- Pre-seed a few products in Catalog before recording

---

## 📊 Revised Timeline Summary

| Day | Focus | Deliverable |
|---|---|---|
| Day 1 | Order Service `main.go` + inventory decrement | Full checkout flow boots and runs |
| Day 2 | Search Service + ES indexing | Search endpoint returns products |
| Day 3 | PDF Invoice generation | PDF downloads after payment |
| Day 4 | Frontend: Home + Products + Detail | Browsable product catalog |
| Day 5 | Frontend: Auth + Cart + Checkout | Full shopping flow to Stripe |
| Day 6 | Frontend: Orders + Search UI + Polish | Complete buyer experience |
| Day 7 | Testing + README + Demo recording | GitHub-ready resume project |
