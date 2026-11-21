# dbo-api

A containerized RESTful API built with **Go 1.26 (Gin-Gonic)**, **PostgreSQL 16**, **GORM**, and **JWT Authentication** following the **MVC** architectural pattern. It provides complete Customer Management, Master-Detail Order Management, and User Authentication modules with audit trails and soft deletion.

---

## Tech Stack

- **Language:** Go 1.26
- **HTTP Framework:** Gin-Gonic (`github.com/gin-gonic/gin`)
- **ORM & Database:** GORM (`gorm.io/gorm`) + PostgreSQL 16
- **Authentication & Security:** JWT (`github.com/golang-jwt/jwt/v5`) + Bcrypt password hashing
- **Validation:** `go-playground/validator/v10` (custom Indonesian phone number & unique email rules)
- **Containerization:** Docker & Docker Compose (multi-stage Alpine build with graceful shutdown & healthchecks)

---

## Project Structure (MVC)

```text
dbo-api/
├── config/
│   └── database.go              # PostgreSQL connection setup
├── controllers/
│   ├── auth.go                  # Register, Login, Get Current User
│   ├── customer.go              # Customer CRUD, Pagination, Search, Options, Soft Delete
│   └── order.go                 # Order + Items CRUD, Pagination, Search, Auto OrderNo, Soft Delete
├── middlewares/
│   └── jwt-middleware.go        # Bearer JWT validation & user context injection
├── migrations/
│   └── init.sql                 # Database DDL, sequences, and indexes
├── models/
│   ├── customer.go              # Customer entity & Create/Update/Option DTOs
│   ├── order.go                 # Order & OrderItem entities + Request DTOs
│   └── user.go                  # User entity & Auth Request/Response DTOs
├── routes/
│   └── routes.go                # API v1 route definitions & middleware grouping
├── utils/
│   ├── jwt.go                   # JWT generation & claims helper
│   └── validation/
│       └── validator.go         # Custom validator registration (phone_id, unique_customer_email)
├── .env.example                 # Environment variable template
├── Dockerfile                   # Multi-stage Go 1.26 build
├── docker-compose.yml           # App + PostgreSQL service orchestration
├── go.mod
└── main.go                      # Entry point with /health check & graceful shutdown
```
---

## Getting Started

- Clone this repository
- Copy `.env.example` to `.env`
- Run docker
  ```bash
  docker compose up -d --build
  ```
- Verify both container running by
  ```bash
  docker compose ps
  curl http://localhost:8080/health
  ```
  the response should be
  ```json
    {
        "db": "connected",
        "status": "ok"
    }
  ```

  ---
  ## Documentation

  API Collection and ERD design available in `docs` folder.