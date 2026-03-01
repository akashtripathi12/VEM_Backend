# 🏨 TBO Events Planner

> **A highly scalable, production-grade Go REST API** designed for travel agents and enterprise clients. It powers end-to-end corporate event bookings, providing a unified ecosystem for managing guests, intelligent margin-optimized negotiations, dynamic room allocations, and comprehensive cart management encompassing hotels, banquets, catering, flights, and transfers.

---

## 📋 Table of Contents

- [Project Overview](#project-overview)
- [Key Features](#key-features)
- [Tech Stack](#tech-stack)
- [System Architecture](#system-architecture)
- [Data Models](#data-models)
- [Authentication & Roles](#authentication--roles)
- [Installation & Setup](#installation--setup)
- [Environment Variables](#environment-variables)
- [API Routes Reference](#api-routes-reference)
- [Folder Structure](#folder-structure)
- [Response Format](#response-format)
- [Future Enhancements](#future-enhancements)
- [Contribution Guide](#contribution-guide)
- [License](#license)

---

## 🚀 Project Overview

TBO Events Planner is a **backend-first** platform architected to serve as the scalable transaction and intelligence core for enterprise travel management. It unifies a fragmented industry by providing a **Unified Booking Ecosystem** (Hotels, Banquets, Catering, Flights, and Transfers) in a single intelligent cart.

**Core workflow:**

```
Agent Signup → Create Event → Add Guests → Browse Global Inventory
    → Unified Cart Management → Margin-Optimized Negotiation → Lock Deal → Auto-Allocate Rooms
```

**[View High-Level Architecture Diagram](https://drive.google.com/file/d/17XF_viA8BTCe5bdF_dS0RtrFA92-_d5Z/view?usp=sharing)** | **[View SQL Schema Visualizer](https://mermaid.live/edit#pako:eNq1WNtu3DYQ_RVBQN4cw4kv6923JG2aokgbuM1LYUDgirMSa4lUSGqdje1_75C6rERSXtlJDAO2ODPkzJkreRengkK8ikH-wkgmSXnNI_z5rEBGd83_5qeuGY3w99Mf-zWlJeNZxEkJ3iKUhBXRZ5-7IkrdCkmTnKjcp-aC-5tJUbSLD9e8-edNBlx_kmLDCvD0rFH5JKwsQbl0lwR1bmkGEI9WiJRoJrhHWNeKcVAqGajea_nrFrU8BKNdM2dro_N7l5IDoUlWg_LIrQa50FBM0IJ2-rb8pwRfI8yiVAnjRmkhd56c0kTXar-sWQlmUeqEEg0OATh1limkrCQFQkYz0FPriarwfAfI34z9z4pHhg5AcD0uvatgZvjZWHaOhu2UvzakZMUOSQ4eREq2JUUIKgoVglhLGBDHpr8pOqfNiqZp7YKBZCnG-4nYbJrUGZI7_6AKN0CTSrIUDgZHl1NKsYyjVNlnVW_ZO1FzLXdDi1opk4KzS4112iBr9_szHdo8HDVpo4zd53AmrQWWJMIjhnkvqrog0jn7g0nK-YejprMy2ESzybhEYjDwzIebUonFyM3sDUlZwTQDj4KOzSCpZaHGh4g0rSuC5dCHW4oKJOo7zqBNIYhuKq-rm9lPwpbBbWJRdnWoRMHSgG5dyCeaZMrB9wqj9S8TrLMxfnKZRB_f4F-nHxhjSvI1SUlFjNv8NNFCY5pviPS3TGspYQTqIJAkbGpOyboAF4gUHYG6-zgZZYKQoj184O8etreEf6lBfyDFKDhrs9H3wmaVmUTFVo0EIwdr3G5OFHbno6rBUCuAZzp3V28Z9RdzYFmu3VV0EPESBYipwy5s77AsG20-Aq9_OG7BhuTDVhWjxtHoSxlgNdgNMmQa1IE1Uv-uoXx-K3msh2IYj5pfP_dJs1EHi9s0C-OhpEu5UH_SknBlutME00QfMh76UhOug1EZbml9JmOeExxjNgB-YHK0RLlTHKW423o3Usy2-BSDTSONaIdQV3RE6L30J2RCM1sA_8aS_l2tfwIbK6By9EyixQ3w0bxu09nWK51IrDH0R1p0ZTacZY9qbHctsj3FbJLwulyD9GzFeYNtmPWG12-wgwmFNVpxUimMyEAMl0TeqMA6wR2cjhBEozf6vY3smaNrmwYTJhEmC7xrOJcXZ4I0nxOT55jUbrqXDN57Olm_CTZtTqHZTraRLY7Lpo25xH7OJwpmTpHPjLQG9LdNoZgVZi3yocITzCg7iBnzbDkCOlW5mwIzLxObnX5Q-finLZczIy8l6yRYzlMcNs3sXoS878weY-87RN_7tq3hwT8rCjoInhIHfZd5UiSgEc8MhIqlN3WVTD4vUKxVAao1vJUNpvXPD7AXL6IrKKxeKmeVGrwb3d-_fCnux280q-g6zokyxdd8X8cu_137WmIY7WtIJKR9_GjurCGBfpCxMpSqiOGXCrGOK4Lhb8CIsD0E2N3Q8QSG7zutTPNGYThTwTVhvFfEZxvc541ASTiaHOYfGfnIzvehicHI2Bea8N4-KuiiIGcIkJZ3-ELTahKykamINCuYBrF7i3sEHAUFpCgUbYTcn9c9HnQombt-CCFLaJmaS7nLNbywt4x7vQwzhuyW0b17RqzDG9VB5tE94iD3Ph9yofRA1z4k7oWF23cjDuFgbpr7nV2ZkEPHUpNjaKueN851KW7nssEW7Qh0KBMZ73Tte9fhbGyE4qM4k4zGKy1rOIpLkCUxn7Et-dexzgFrZGzkKA52RuQBZSrC_0VPd2Kod5bHqw0pFH41ha99FO9ZgFOQNvbi1avzhd0jXt3FX-PVy9PLk-PFyfny7OLsfPH61dKQd7h-sTg-O1meXi4ul-eXr08Wpw9H8Td77qvj5XK5WJyen72-uFguT5ZHMVCmhfzYPMrbt_mH_wHrqGSE)**

---

## ✨ Key Features

- **🏢 Scalable Unified Booking Ecosystem:** Unlike fragmented platforms, TBO Backend seamlessly consolidates **Hotels (Rooms, Banquets, Catering)**, **Flights**, and **Transfers** into a single transaction boundary and polymorphic cart. This enables a unified, one-click enterprise checkout.
- **🧠 Negotiation Intelligence & Margin Optimization:** A state-of-the-art multi-round price negotiation engine. `NegotiationSession` tracks historical snapshots, competitor pricing data, and budget constraints. This ensures agents maximize margin optimization while retaining enterprise clients. TBO Events Planner agents review structured JSONB `ProposalSnapshots` before executing atomic `LockDeal` transactions.
- **⚙️ Auto-Allocation Engine:** A deterministic, family-based algorithmic engine securely auto-allocates hundreds of corporate guests into optimal room configurations, effortlessly tracking room inventory capacities.
- **🛡️ Clean API Design & Security Best Practices:** Fully RESTful, strictly validated inputs, parameterized database execution via GORM, robust recovery middleware, request tracing, and hierarchical RBAC (Role-Based Access Control) isolating `agent`, `head_guest`, and `tbo_agent` scopes.
- **⚡ Async Reliability:** Built on top of Asynq + Redis. Crucial processes like firing invitation links or bulk email delivery to hundreds of corporate guests happen in the background, ensuring the main thread remains unblocked and highly responsive.

---

## 🛠 Tech Stack

| Component | Technology |
|---|---|
| **Language** | Go 1.25 |
| **Web Framework** | [Fiber v2](https://github.com/gofiber/fiber) |
| **ORM** | [GORM](https://gorm.io) with `gorm.io/datatypes` for JSONB |
| **Database** | PostgreSQL (via `pgx` driver) |
| **Caching** | Redis (`go-redis/v9`) |
| **Task Queue** | [Asynq](https://github.com/hibiken/asynq) (backed by Redis) |
| **Authentication** | JWT (`golang-jwt/jwt/v5`) |
| **Password Hashing** | bcrypt (`golang.org/x/crypto`) |
| **ID Generation** | UUID v4 (`google/uuid`) |
| **Config** | `godotenv` + `os.Getenv` |
| **Testing** | `testify` |

---

## 📐 System Architecture

The application is built using a horizontal, stateless server design backed by robust queuing and data layers.

```
┌─────────────────────────────────────────────────────────────┐
│                        Client (Frontend)                     │
└─────────────────────┬───────────────────────────────────────┘
                      │  HTTP/REST
┌─────────────────────▼───────────────────────────────────────┐
│                    Fiber HTTP Server                         │
│  ┌─────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  CORS MW    │  │  Auth Middleware │  │  Logger/Recover │  │
│  └─────────────┘  └─────────────────┘  └─────────────────┘  │
│                                                              │
│  ┌───────────────────────────────────────────────────────┐   │
│  │                      Handlers                         │   │
│  │  Auth │ Events │ Guests │ Hotels │ Cart │ Negotiation │   │
│  │  Flights │ Transfers │ Allocations │ Locations │ ...  │   │
│  └───────────────────────────────────────────────────────┘   │
└─────────────┬─────────────────────────┬────────────────────-─┘
              │                         │
   ┌──────────▼──────────┐   ┌──────────▼──────────┐
   │      PostgreSQL      │   │        Redis         │
   │  (Primary Data Store)│   │ (Cache + Task Queue) │
   └──────────────────────┘   └──────────────────────┘
                                        │
                              ┌──────────▼──────────┐
                              │   Asynq Worker       │
                              │  (Email Delivery)    │
                              └──────────────────────┘
```

### Key Architectural Decisions

- **Scalability by Design (Repository Pattern)** — Handlers sit on a stateless `Repository` interface managing database pools, Redis connections, and async task clients, easily scalable across instances.
- **Security Best Practices** — JWT-based authentication via robust middleware, CORS enforcement, parameterized SQL queries (preventing SQLi), and role-based access control (RBAC).
- **Enterprise Async Job Processing** — Non-blocking work (e.g., transactional emails) is dispatched to distributed Redis-backed Asynq workers to keep API response times < 50ms.
- **High-Performance Redis Caching** — Sub-millisecond read latency for event catalogs and inventory metrics via `go-redis/v9`.
- **Strict Transaction Boundaries** — Complex mutations (like multi-round negotiation locks or cascading event deletions) execute within serialized Postgres transactions, eliminating race conditions.
- **Unified Polymorphic Cart** — A highly abstracted `CartItem` table leverages a `Type + RefID` matrix, consolidating Rooms, Banquets, Catering, Flights, and Transfers without schema sprawl.
- **Hybrid Schema Flexibility** — `RoomsInventory`, `Facilities`, `Policies`, and `ProposalSnapshot` are stored as PostgreSQL `jsonb` columns for NoSQL-like flexibility where needed.

---

## 🗄️ Data Models

### Core Entities

```
Users ──────── AgentProfile (1:1)
  │
  ├── Events ─── Guests ──── GuestAllocations ─── RoomOffers
  │       │                                            │
  │       ├── CartItems ────── FlightBookings      Hotels ─── BanquetHalls
  │       │       │            TransferBookings        │
  │       │       └── (rooms, banquets,               └── CateringMenus
  │       │             catering, flights,
  │       └── NegotiationSessions ── NegotiationRounds
  │                  (per CartItem ProposalSnapshot)
  │
Countries ─── Cities ─── Hotels
                  │
              Flights / Transfers (global catalog)
```

### Model Summary

| Model | Key Fields |
|---|---|
| `User` | `id`, `name`, `email`, `role` (`agent`/`head_guest`/`tbo_agent`) |
| `AgentProfile` | `agency_name`, `agency_code`, `location`, `business_phone` |
| `Event` | `name`, `hotel_id`, `location`, `start_date`, `end_date`, `budget`, `status`, `rooms_inventory` (JSONB) |
| `Guest` | `guest_name`, `age`, `type` (`adult`/`child`), `email`, `phone`, `family_id`, `arrival_date`, `departure_date` |
| `GuestAllocation` | `event_id`, `guest_id`, `room_offer_id`, `locked_price`, `status`, `assigned_mode` |
| `Hotel` | `hotel_code`, `name`, `star_rating`, `facilities` (JSONB), `image_urls` (JSONB), `policies` (JSONB) |
| `RoomOffer` | `name`, `booking_code`, `max_capacity`, `total_fare`, `is_refundable`, `cancel_policies` (JSONB) |
| `BanquetHall` | `name`, `capacity`, `price_per_day`, `hall_type`, dimensions, `features` (JSONB) |
| `CateringMenu` | `name`, `type` (`veg`/`non-veg`), `price_per_plate`, `dietary_tags` (JSONB) |
| `CartItem` | `type`, `ref_id`, `status` (`wishlist`/`cart`/`approved`/`booked`), `locked_price`, `quantity` |
| `NegotiationSession` | `event_id`, `status`, `share_token` (UUID), `current_round` |
| `NegotiationRound` | `session_id`, `round_number`, `modified_by`, `proposal_snapshot` (JSONB), `remarks` |

---

## 🔒 Authentication & Roles

### Auth Mechanism

- **JWT Bearer Tokens** — All protected routes require `Authorization: Bearer <token>` header.
- Tokens are signed with a secret and carry `user_id`, `email`, and `role` as claims.
- The `Protected` middleware validates the token and injects claims into `c.Locals`.

### Roles

| Role | Description | Key Capabilities |
|---|---|---|
| `agent` | Travel agent (signs up publicly) | Create events, manage guests, browse hotels, manage cart, start negotiations |
| `head_guest` | The main guest of an event (created by agent) | View their event, register family members via invite link |
| `tbo_agent` | TBO Events Planner platform administrator | View all negotiations, submit counter-offers, lock deals |

### Auth Flow

```
[Agent]
  POST /api/v1/auth/signup  →  Creates User + AgentProfile  →  Returns JWT
  POST /api/v1/auth/login   →  Verifies bcrypt hash         →  Returns JWT + User info

[Head Guest]
  POST /api/v1/auth/login   →  Same endpoint                →  Returns JWT + eventId

[TBO Events Planner Agent]
  Created via seeder script  →  role: "tbo_agent"
```

> **Head Guest creation:** When an agent assigns a head guest (`POST /events/:id/head-guest`), a new `User` with role `head_guest` is created and temporary credentials are emailed via the Asynq queue.

---

## ⚙️ Installation & Setup

### Prerequisites

- Go 1.25+
- PostgreSQL
- Redis

### 1. Clone the repository

```bash
git clone https://github.com/akashtripathi12/TBO_Backend.git
cd TBO_Backend
```

### 2. Copy and configure environment

```bash
cp .env.example .env
# Edit .env with your DB and Redis credentials
```

### 3. Install dependencies

```bash
go mod tidy
```

### 4. Run database migrations (if using migration scripts)

```bash
go run cmd/migrate_lifecycle/main.go
```

### 5. Seed initial data (optional)

```bash
go run cmd/seed/main.go
```

### 6. Start the server

```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`.

### Run Tests

```bash
go test ./internal/handlers/... -v
```

---

## 🔐 Environment Variables

Create a `.env` file in the project root:

```env
# Server
PORT=8080
ENV=development

# Database (PostgreSQL)
DATABASE_URL=postgres://user:password@localhost:5432/tbo_backend

# Redis
REDIS_URL=redis://127.0.0.1:6379

# CORS
ALLOWED_ORIGINS=http://localhost:3000,https://your-frontend.vercel.app
FRONTEND_URL=http://localhost:3000

# JWT
JWT_SECRET=your-super-secret-key

# Email
SMTP_EMAIL=tboemailservice@gmail.com
SMTP_PASS=qnsy jmcx zzhe hzgq
```

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server port |
| `ENV` | `development` | Environment name |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `REDIS_URL` | `redis://127.0.0.1:6379` | Redis connection URI |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated CORS origins |
| `FRONTEND_URL` | `http://localhost:3000` | Frontend base URL (used in email links) |
| `JWT_SECRET` | — | Secret for signing JWT tokens |
| `SMTP_EMAIL` | `tboemailservice@gmail.com` | SMTP email address for sending emails |
| `SMTP_PASS` | `qnsy jmcx zzhe hzgq` | SMTP application password |

---

## 🛣 API Routes Reference

*Base URL: `/api/v1`*

### 🔓 Public Routes

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `POST` | `/auth/signup` | Register a new agent account |
| `POST` | `/auth/login` | Login (any role) → returns JWT |
| `POST` | `/agents/onboarding` | Legacy alias for signup |
| `GET` | `/locations/countries` | List all countries |
| `GET` | `/locations/cities` | List all cities (with filtering) |

### 🔐 Protected Routes (JWT Required)

#### 👤 User

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/me` | Get current authenticated user profile |
| `GET` | `/dashboard/metrics` | Get dashboard metrics |

#### 📅 Events

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/events` | List agent's events (with metrics: guest count, budget spent, days until) |
| `POST` | `/events` | Create a new event |
| `GET` | `/events/:id` | Get single event with full metrics |
| `PUT` | `/events/:id` | Update event details |
| `DELETE` | `/events/:id` | Delete event (cascading: guests, allocations, cart, negotiations) |
| `GET` | `/events/:id/guests` | List all guests for an event |
| `POST` | `/events/:id/guests` | Add a guest to an event |
| `GET` | `/events/:id/venues` | Get hotels added to this event's cart |
| `GET` | `/events/:id/allocations` | Get room allocations for the event |
| `POST` | `/events/:id/head-guest` | Assign/create head guest (sends email with credentials) |
| `POST` | `/events/:id/send-invites` | Send invitation emails to guests |
| `POST` | `/events/:id/auto-allocate` | Automatically allocate guests to rooms by family groups |
| `POST` | `/events/:id/finalize` | Finalize room allocations (lock rooms) |
| `POST` | `/events/:id/reopen` | Reopen a finalized allocation |

#### 🛒 Cart

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/events/:id/cart` | Get cart (grouped by hotel → rooms, banquets, catering) |
| `POST` | `/events/:id/cart` | Add a single item to cart |
| `POST` | `/events/:id/cart/bulk` | Bulk add multiple items |
| `PATCH` | `/events/:id/cart/:cartItemId` | Update a cart item (quantity, notes, price) |
| `DELETE` | `/events/:id/cart/:cartItemId` | Remove a cart item |
| `DELETE` | `/events/:id/cart/hotel/:hotelId` | Remove all items for a hotel from cart |
| `POST` | `/events/:id/cart/approve` | Update cart status (wishlist → cart → approved) |

#### 👥 Guests

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/guests/:id` | Get single guest |
| `PATCH` | `/guests/:id` | Update guest details (name, age, email, phone, dates) |
| `DELETE` | `/guests/:id` | Delete a guest |
| `POST` | `/guests/:id/subguests` | Add a family member (sub-guest) to existing guest |

#### 🏠 Allocations

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/allocations` | Manually create a room allocation |
| `PUT` | `/allocations/:id` | Update an allocation |

#### 🏨 Hotels

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/hotels` | List hotels by city (with filters: star rating, property type, user rating) |
| `GET` | `/hotels/:id` | Get hotel details |
| `GET` | `/hotels/:hotelCode/rooms` | Get room offers for a hotel |
| `GET` | `/hotels/:hotelCode/banquets` | Get banquet halls for a hotel |
| `GET` | `/hotels/:hotelCode/catering` | Get catering menus for a hotel |

#### ✈️ Flights

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/flights` | List all global flights |
| `GET` | `/flights/:id` | Get a single flight |
| `POST` | `/flights` | Create a flight (admin) |
| `PUT` | `/flights/:id` | Update a flight |
| `DELETE` | `/flights/:id` | Delete a flight |
| `GET` | `/flights/locations` | Get unique booking locations |
| `GET` | `/events/:id/flights` | Get flight bookings for an event |
| `POST` | `/events/:id/flights/book` | Book a flight for an event |
| `DELETE` | `/events/:id/flights/:booking_id` | Cancel a flight booking |

#### 🚌 Transfers

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/transfers` | List all global transfers |
| `GET` | `/transfers/:id` | Get a single transfer |
| `POST` | `/transfers` | Create a transfer (admin) |
| `PUT` | `/transfers/:id` | Update a transfer |
| `DELETE` | `/transfers/:id` | Delete a transfer |
| `GET` | `/events/:id/transfers` | Get transfer bookings for an event |
| `POST` | `/events/:id/transfers/book` | Book a transfer for an event |
| `DELETE` | `/events/:id/transfers/:booking_id` | Cancel a transfer booking |

#### 🤝 Negotiation

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `POST` | `/negotiation/init` | Protected | Start negotiation from event cart |
| `POST` | `/negotiation/counter` | Public (share token) | Submit counter offer |
| `GET` | `/negotiation/:id/diff` | Public | Get price diff between rounds |
| `POST` | `/negotiation/lock` | Protected | Lock deal and apply final prices to cart |
| `GET` | `/negotiation/token/:token` | Public | Resolve share token to session |

#### 🛡️ Admin (TBO Events Planner Agent Only)

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/admin/negotiations` | List all active (non-locked) negotiation sessions |

---

### 📮 Postman Collection

Explore and test the complete API suite using our detailed Postman Workspace:
[**View TBO Backend Postman Collection**](https://www.postman.com/winter-star-664762/workspace/tbo-backend/folder/37428593-3d6aca16-83c4-4361-a584-211488943b6a?action=share&creator=37428593&ctx=documentation)

---

## 📁 Folder Structure

```text
TBO_Backend/
├── cmd/
│   ├── api/                    # Main application entry point
│   ├── seed/                   # Data seeder scripts
│   ├── migrate_lifecycle/      # Migration runners
│   ├── migrate_negotiation/    # Negotiation schema migrations
│   └── ...                     # Various dev/debug utility scripts
│
├── internal/
│   ├── config/                 # Config loading from env vars
│   ├── handlers/               # HTTP handler functions (business logic)
│   ├── middleware/             # Fiber middleware (Auth, CORS, Logger, Recovery)
│   ├── models/                 # GORM models & DTOs
│   ├── queue/                  # Asynq task definitions & background workers
│   ├── routes/                 # All API route registrations
│   ├── store/                  # DB & Redis initialization
│   ├── utils/                  # JWT, response helpers, cache invalidation
│   └── scripts/                # Internal utility scripts
│
├── migrations/                 # SQL migration files
├── docs/                       # Postman collections
├── go.mod                      # Go module dependencies
└── .env                        # Environment configurations
```

---

## 📦 Response Format

All endpoints return a consistent JSON structure:

```json
// Success
{
  "message": "Events Fetched Successfully",
  "events": [...]
}

// Error
{
  "error": "Description of what went wrong"
}
```

HTTP status codes follow REST conventions: `200 OK`, `201 Created`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `409 Conflict`, `500 Internal Server Error`.

---

## 🚀 Phase-2 Future Enhancements

**Core Idea:** Transforming high-touch, manual MICE bookings into a zero-touch, error-free digital workflow through decentralized data, automated negotiations, and algorithmic margin protection.

### 💹 Inventory & Rate Optimization
- **Intelligent Policy Arbitrage:** Automatically swaps guests between Non-Refundable and Flexible rooms during cancellations to eliminate penalty fees by optimizing the group's internal mix.
- **Variable Cost Stripping:** Automatically executes API modifications to downgrade rate plans (e.g., stripping Meal Plans or Extra Bed charges) during total cancellations to recover 30-40% variable costs.

### 💰 Financial Conversion & Audit
- **Asset Conversion Workflow:** Automates the conversion of 'Sunk Inventory' into 'Banquet Credit', generating a barter request to apply lost room revenue toward the group's master F&B bill instead of a zero-refund cancellation.
- **Automated Audit 'Shadow Folio':** Maintains a real-time 'Shadow Bill' of all authorized charges and auto-reconciles at checkout to flag unauthorized 'Extras' (minibar, laundry) before final payment.

### 🤖 AI Sourcing & Anchor-Satellite Routing
- **Predictive B2B Venue Matching:** AI bypasses basic B2C filters to instantly shortlist hotels with the highest historical probability of clearing massive 500+ bulk room negotiations.
- **Anchor-Satellite Distribution:** Intelligently splits oversized groups across geofenced properties—routing VIPs to a flagship "Anchor" hotel and standard attendees to nearby "Satellite" hotels.
- **Logistics-Aware Aggregation:** Auto-calculates transit times between properties and seamlessly bundles multi-hotel rooms, transfers, and banquets into one unified booking cart.

### 🌍 Strategic Ecosystem Expansion
- **Railway Integration:** Mass Transit API Sourcing to seamlessly bundle bulk railway charters (e.g., IRCTC) directly into the master event cart for cost-effective, large-scale domestic delegate movement.
- **Umrah Packages:** Specialized Religious Routing with automated end-to-end workflows including visa processing, Haram-near hotel filtering, and structured group logistics.
- **Dynamic ML Bid Predictor:** Uses regression models on historical occupancy, seasonality, and event trends to predict hotel acceptance probability and instantly suggest the optimal "sweet spot" group rate.

---

## 🤝 Contribution Guide

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes strictly following Conventional Commits (`git commit -m 'feat: Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request for review

---

## 🏆 Acknowledgements

- **Akash Tripathi** - [akashtripathi12](https://github.com/akashtripathi12)
- **Kavya Rathod** - [kavyarathod05](https://github.com/kavyarathod05)
- **Naman Jain** - [belnoinja](https://github.com/belnoinja)
- **Tushar Jawale** - [Tushar-Jawale](https://github.com/Tushar-Jawale)
- **Tushar Tiwari** - [tushar330](https://github.com/tushar330)

- Designed for scale, modern travel workflows, and enterprise event management using the robust Go ecosystem.

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

*Built with ❤️ using Go + Fiber + GORM + PostgreSQL + Redis*