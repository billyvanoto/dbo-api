# DBO-API — RESTful API Technical Documentation

| Document Metadata | Details |
| :--- | :--- |
| **Project Name** | `dbo-api` (Customer & Order Management REST API) |
| **API Version** | `v1.0.0` |
| **Base URL** | `http://localhost:8080/api/v1` |
| **Authentication** | Bearer Token (JWT - HS256) |
| **Content-Type** | `application/json` |
| **Tech Stack** | Go 1.26 (Gin-Gonic), GORM, PostgreSQL 16, Docker |
| **Date** | November 21, 2026 |

---

## 1. Overview & Architecture

`dbo-api` is a containerized RESTful backend service designed around the **Model-View-Controller (MVC)** and **Data Transfer Object (DTO)** patterns. It provides modular management for User Authentication, Customers, and Master-Detail Orders.

### Key Architectural Highlights
- **Authentication & Authorization:** Stateless JSON Web Tokens (JWT) passed via the `Authorization: Bearer <token>` HTTP header. User identity (`user_id`, `email`) is extracted via middleware and injected into the request context.
- **Audit Trail Logging:** Every entity automatically tracks `created_at`, `created_by`, `updated_at`, and `updated_by` using the authenticated user's email from the JWT payload.
- **Soft Deletion:** Records in `customers` and `orders` are never permanently removed via `DELETE` endpoints; instead, deletion timestamps and user emails are recorded (`deleted_at`/`deleted_by`), and active queries automatically filter them out.
- **Server-Side Integrity:** Order reference numbers (`orderNo`) are atomically generated via a PostgreSQL sequence (`ORD-YYYYMMDD-00001`), and all line-item and grand totals are computed server-side inside ACID database transactions.

---

## 2. Authentication & Security Headers

All endpoints under `/api/v1/customers`, `/api/v1/orders`, and `/api/v1/auth/me` require a valid JWT token obtained from `POST /api/v1/auth/login`.

### Required Request Headers
```http
Content-Type: application/json
Authorization: Bearer <your_jwt_token>
```

---

## 3. Standard HTTP Status Codes & Error Format

### Status Codes
| HTTP Status | Meaning | Common Trigger |
| :--- | :--- | :--- |
| `200 OK` | Success | Successful `GET`, `PUT`, `DELETE`, or `POST /auth/login` request |
| `201 Created` | Resource Created | Successful `POST /auth/register`, `POST /customers`, or `POST /orders` |
| `400 Bad Request` | Validation / Input Error | Missing required fields, invalid Indonesian phone format, invalid ID URI, or malformed JSON |
| `401 Unauthorized` | Authentication Failure | Missing `Authorization` header, malformed Bearer token, expired JWT, or wrong login credentials |
| `404 Not Found` | Resource Not Found | Target record ID does not exist or has already been soft-deleted |
| `409 Conflict` | Resource Conflict | Duplicate unique constraint violation (e.g., email already registered) |
| `500 Internal Server Error` | Server Error | Unexpected database or sequence generation error |
| `503 Service Unavailable` | Unhealthy Dependency | Database ping failed on `GET /health` |

### Standard Error Response Structure
```json
{
  "error": "Descriptive error message explaining what went wrong"
}
```

---

## 4. Endpoint Summary Matrix

| Module | Method | Endpoint | Auth Required | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Health** | `GET` | `/health` | No | Check HTTP server and PostgreSQL connection health |
| **Auth** | `POST` | `/api/v1/auth/register` | No | Register a new system user |
| **Auth** | `POST` | `/api/v1/auth/login` | No | Authenticate user and issue JWT token |
| **Auth** | `GET` | `/api/v1/auth/me` | Yes | Retrieve current logged-in user details |
| **Customers** | `GET` | `/api/v1/customers` | Yes | List active customers (paginated & searchable) |
| **Customers** | `GET` | `/api/v1/customers/options` | Yes | Lightweight customer list for UI dropdowns |
| **Customers** | `GET` | `/api/v1/customers/:id` | Yes | Get single customer details with order history |
| **Customers** | `POST` | `/api/v1/customers` | Yes | Create a new customer |
| **Customers** | `PUT` | `/api/v1/customers/:id` | Yes | Update an existing active customer |
| **Customers** | `DELETE` | `/api/v1/customers/:id` | Yes | Soft-delete a customer |
| **Orders** | `GET` | `/api/v1/orders` | Yes | List active orders (paginated, searchable & filterable) |
| **Orders** | `GET` | `/api/v1/orders/:id` | Yes | Get single order details with customer & line items |
| **Orders** | `POST` | `/api/v1/orders` | Yes | Create a new order with line items |
| **Orders** | `PUT` | `/api/v1/orders/:id` | Yes | Update an order and replace line items atomically |
| **Orders** | `DELETE` | `/api/v1/orders/:id` | Yes | Soft-delete an order |

---

## 5. Detailed API Reference

### 5.1 System Health

#### Check Service & Database Health
- **Endpoint:** `GET /health`
- **Authentication:** None (Public)

**Response (`200 OK`):**
```json
{
  "db": "connected",
  "status": "ok"
}
```

**Response (`503 Service Unavailable`):**
```json
{
  "db": "disconnected",
  "status": "unhealthy"
}
```

---

### 5.2 Authentication Module

#### 1. Register User
Creates a new user account with a bcrypt-hashed password.

- **Endpoint:** `POST /api/v1/auth/register`
- **Authentication:** None (Public)

**Request Body Parameters:**
| Field | Type | Required | Validation / Constraints |
| :--- | :--- | :--- | :--- |
| `name` | `string` | Yes | Unique display name |
| `email` | `string` | Yes | Valid email format, unique |
| `password` | `string` | Yes | Minimum 6 characters |

**Sample Request:**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "password": "abcd1234"
}
```

**Sample Response (`201 Created`):**
```json
{
  "message": "User created",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@example.com",
    "createdAt": "2026-11-21T11:30:00Z",
    "updatedAt": "2026-11-21T11:30:00Z"
  }
}
```

---

#### 2. Login User
Authenticates credentials and returns a Bearer JWT token valid for 24 hours (`86400` seconds).

- **Endpoint:** `POST /api/v1/auth/login`
- **Authentication:** None (Public)

**Request Body Parameters:**
| Field | Type | Required | Validation / Constraints |
| :--- | :--- | :--- | :--- |
| `email` | `string` | Yes | Registered user email |
| `password` | `string` | Yes | Plaintext password |

**Sample Request:**
```json
{
  "email": "john.doe@example.com",
  "password": "abcd1234"
}
```

**Sample Response (`200 OK`):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@example.com",
    "createdAt": "2026-11-21T11:30:00Z",
    "updatedAt": "2026-11-21T11:30:00Z"
  }
}
```

---

#### 3. Validate Current Login State
Retrieves the profile of the currently authenticated user based on their JWT token.

- **Endpoint:** `GET /api/v1/auth/me`
- **Authentication:** Bearer Token Required

**Sample Response (`200 OK`):**
```json
{
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@example.com",
    "createdAt": "2026-11-21T11:30:00Z",
    "updatedAt": "2026-11-21T11:30:00Z"
  }
}
```

---

### 5.3 Customer Management Module

#### 1. List Customers (Paginated & Searchable)
Retrieves a paginated list of active (`deleted_at IS NULL`) customers ordered by newest first.

- **Endpoint:** `GET /api/v1/customers`
- **Authentication:** Bearer Token Required

**Query Parameters:**
| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `page` | `integer` | `1` | Page number (1-indexed) |
| `limit` | `integer` | `10` | Number of records per page |
| `search` | `string` | `""` | Case-insensitive partial match against `name`, `email`, or `phone` |

**Sample Request:**
`GET /api/v1/customers?page=1&limit=10&search=budi`

**Sample Response (`200 OK`):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Budi Santoso",
      "email": "budi@example.com",
      "phone": "081234567890",
      "address": "Jl. Sudirman No. 10, Jakarta",
      "createdAt": "2026-11-21T11:32:00Z",
      "createdBy": "john.doe@example.com",
      "updatedAt": "2026-11-21T11:32:00Z",
      "updatedBy": "john.doe@example.com"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

---

#### 2. Get Customer Lookup Options (Dropdown)
Returns a lightweight, bounded list of active customers sorted alphabetically by name for frontend select/autocomplete components.

- **Endpoint:** `GET /api/v1/customers/options`
- **Authentication:** Bearer Token Required

**Query Parameters:**
| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `search` | `string` | `""` | Case-insensitive partial match against `name` or `email` |
| `limit` | `integer` | `20` | Maximum items returned (capped at `100`) |

**Sample Request:**
`GET /api/v1/customers/options?search=budi&limit=20`

**Sample Response (`200 OK`):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Budi Santoso",
      "email": "budi@example.com"
    }
  ]
}
```

---

#### 3. Get Customer Detail
Retrieves a single active customer by ID, including their associated order history.

- **Endpoint:** `GET /api/v1/customers/:id`
- **Authentication:** Bearer Token Required

**Path Parameters:**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | `uint` | Yes | Positive integer Customer ID (`> 0`) |

**Sample Response (`200 OK`):**
```json
{
  "data": {
    "id": 1,
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "phone": "081234567890",
    "address": "Jl. Sudirman No. 10, Jakarta",
    "orders": [
      {
        "orderId": 1,
        "customerId": 1,
        "orderNo": "ORD-20261121-00001",
        "total": 1850000,
        "status": "PENDING",
        "createdAt": "2026-11-21T11:35:00Z",
        "createdBy": "john.doe@example.com",
        "updatedAt": "2026-11-21T11:35:00Z",
        "updatedBy": "john.doe@example.com"
      }
    ],
    "createdAt": "2026-11-21T11:32:00Z"
  }
}
```

---

#### 4. Create Customer
Creates a new customer record and sets `createdBy` and `updatedBy` to the authenticated user's email.

- **Endpoint:** `POST /api/v1/customers`
- **Authentication:** Bearer Token Required

**Request Body Parameters:**
| Field | Type | Required | Validation / Constraints |
| :--- | :--- | :--- | :--- |
| `name` | `string` | Yes | Customer full name |
| `email` | `string` | Yes | Valid email format; must be unique among active customers |
| `phone` | `string` | Yes | Indonesian mobile number matching `^(\+62\|62\|0)8[1-9][0-9]{6,10}$` |
| `address` | `string` | No | Physical or billing address |

**Sample Request:**
```json
{
  "name": "Budi Santoso",
  "email": "budi@example.com",
  "phone": "081234567890",
  "address": "Jl. Sudirman No. 10, Jakarta"
}
```

**Sample Response (`201 Created`):**
```json
{
  "data": {
    "id": 1,
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "phone": "081234567890",
    "address": "Jl. Sudirman No. 10, Jakarta",
    "createdAt": "2026-11-21T11:32:00Z"
  }
}
```

---

#### 5. Update Customer
Updates an existing active customer's details and sets `updatedBy` to the authenticated user's email while preserving `createdAt` and `createdBy`.

- **Endpoint:** `PUT /api/v1/customers/:id`
- **Authentication:** Bearer Token Required

**Path Parameters:**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | `uint` | Yes | Positive integer Customer ID (`> 0`) |

**Request Body Parameters:**
| Field | Type | Required | Validation / Constraints |
| :--- | :--- | :--- | :--- |
| `name` | `string` | Yes | Customer full name |
| `email` | `string` | Yes | Valid email format; must not belong to a different active customer |
| `phone` | `string` | Yes | Indonesian mobile number matching `^(\+62\|62\|0)8[1-9][0-9]{6,10}$` |
| `address` | `string` | No | Physical or billing address |

**Sample Request:**
```json
{
  "name": "Budi Santoso Wijaya",
  "email": "budi@example.com",
  "phone": "+6281234567890",
  "address": "Jl. Thamrin No. 25, Jakarta"
}
```

**Sample Response (`200 OK`):**
```json
{
  "data": {
    "id": 1,
    "name": "Budi Santoso Wijaya",
    "email": "budi@example.com",
    "phone": "+6281234567890",
    "address": "Jl. Thamrin No. 25, Jakarta",
    "createdAt": "2026-11-21T11:32:00Z"
  }
}
```

---

#### 6. Soft-Delete Customer
Marks an active customer as deleted by populating `deleted_at` with the current timestamp and `deleted_by` with the authenticated user's email.

- **Endpoint:** `DELETE /api/v1/customers/:id`
- **Authentication:** Bearer Token Required

**Path Parameters:**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | `uint` | Yes | Positive integer Customer ID (`> 0`) |

**Sample Response (`200 OK`):**
```json
{
  "message": "Customer deleted successfully"
}
```

---

### 5.4 Order Management Module

#### 1. List Orders (Paginated, Searchable & Filterable)
Retrieves a paginated list of active (`deleted_at IS NULL`) orders with preloaded customer and line-item details.

- **Endpoint:** `GET /api/v1/orders`
- **Authentication:** Bearer Token Required

**Query Parameters:**
| Parameter | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `page` | `integer` | `1` | Page number (1-indexed) |
| `limit` | `integer` | `10` | Number of records per page |
| `search` | `string` | `""` | Case-insensitive partial match against `orderNo` |
| `status` | `string` | `""` | Exact status filter (`PENDING`, `PAID`, `SHIPPED`, `CANCELLED`) |

**Sample Request:**
`GET /api/v1/orders?page=1&limit=10&status=PENDING`

**Sample Response (`200 OK`):**
```json
{
  "data": [
    {
      "orderId": 1,
      "customerId": 1,
      "customer": {
        "id": 1,
        "name": "Budi Santoso",
        "email": "budi@example.com",
        "phone": "081234567890",
        "address": "Jl. Sudirman No. 10, Jakarta",
        "createdAt": "2026-11-21T11:32:00Z"
      },
      "orderNo": "ORD-20261121-00001",
      "total": 1850000,
      "status": "PENDING",
      "items": [
        {
          "orderItemId": 1,
          "orderId": 1,
          "name": "Mechanical Keyboard",
          "qty": 2,
          "price": 750000,
          "total": 1500000,
          "createdAt": "2026-11-21T11:35:00Z",
          "createdBy": "john.doe@example.com",
          "updatedAt": "2026-11-21T11:35:00Z",
          "updatedBy": "john.doe@example.com"
        },
        {
          "orderItemId": 2,
          "orderId": 1,
          "name": "Wireless Mouse",
          "qty": 1,
          "price": 350000,
          "total": 350000,
          "createdAt": "2026-11-21T11:35:00Z",
          "createdBy": "john.doe@example.com",
          "updatedAt": "2026-11-21T11:35:00Z",
          "updatedBy": "john.doe@example.com"
        }
      ],
      "createdAt": "2026-11-21T11:35:00Z",
      "createdBy": "john.doe@example.com",
      "updatedAt": "2026-11-21T11:35:00Z",
      "updatedBy": "john.doe@example.com"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

---

#### 2. Get Order Detail
Retrieves a single active order by ID along with its customer and line items.

- **Endpoint:** `GET /api/v1/orders/:id`
- **Authentication:** Bearer Token Required

**Path Parameters:**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | `uint` | Yes | Positive integer Order ID (`> 0`) |

**Sample Response (`200 OK`):**
```json
{
  "data": {
    "orderId": 1,
    "customerId": 1,
    "customer": {
      "id": 1,
      "name": "Budi Santoso",
      "email": "budi@example.com",
      "phone": "081234567890",
      "address": "Jl. Sudirman No. 10, Jakarta",
      "createdAt": "2026-11-21T11:32:00Z"
    },
    "orderNo": "ORD-20261121-00001",
    "total": 1850000,
    "status": "PENDING",
    "items": [
      {
        "orderItemId": 1,
        "orderId": 1,
        "name": "Paint",
        "qty": 2,
        "price": 750000,
        "total": 1500000,
        "createdAt": "2026-11-21T11:35:00Z",
        "createdBy": "john.doe@example.com",
        "updatedAt": "2026-11-21T11:35:00Z",
        "updatedBy": "john.doe@example.com"
      }
    ],
    "createdAt": "2026-11-21T11:35:00Z",
    "createdBy": "john.doe@example.com",
    "updatedAt": "2026-11-21T11:35:00Z",
    "updatedBy": "john.doe@example.com"
  }
}
```

---

#### 3. Create Order
Creates a new order and its line items inside an ACID transaction. `orderNo` is auto-generated from `order_no_seq`, and all item totals (`qty * price`) and order grand totals are computed automatically by the server.

- **Endpoint:** `POST /api/v1/orders`
- **Authentication:** Bearer Token Required

**Request Body Parameters:**
| Field | Type | Required | Validation / Constraints |
| :--- | :--- | :--- | :--- |
| `customerId` | `uint` | Yes | Must reference an existing active customer (`deleted_at IS NULL`) |
| `status` | `string` | No | Defaults to `PENDING`. Allowed: `PENDING`, `PAID`, `SHIPPED`, `CANCELLED` |
| `items` | `array` | Yes | Minimum 1 item (`min=1`) |
| `items[].name` | `string` | Yes | Item/product name (max 50 characters) |
| `items[].qty` | `integer` | Yes | Quantity (`> 0`) |
| `items[].price` | `number` | Yes | Unit price (`> 0`) |

**Sample Request:**
```json
{
  "customerId": 1,
  "status": "PENDING",
  "items": [
    {
      "name": "Paint",
      "qty": 2,
      "price": 750000
    },
    {
      "name": "Pipe",
      "qty": 1,
      "price": 350000
    }
  ]
}
```

**Sample Response (`201 Created`):**
```json
{
  "data": {
    "orderId": 1,
    "customerId": 1,
    "customer": {
      "id": 1,
      "name": "Budi Santoso",
      "email": "budi@example.com",
      "phone": "081234567890",
      "address": "Jl. Sudirman No. 10, Jakarta"
    },
    "orderNo": "ORD-20261121-00001",
    "total": 1850000,
    "status": "PENDING",
    "items": [
      {
        "orderItemId": 1,
        "orderId": 1,
        "name": "Paint",
        "qty": 2,
        "price": 750000,
        "total": 1500000,
        "createdAt": "2026-11-21T11:35:00Z",
        "createdBy": "john.doe@example.com",
        "updatedAt": "2026-11-21T11:35:00Z",
        "updatedBy": "john.doe@example.com"
      },
      {
        "orderItemId": 2,
        "orderId": 1,
        "name": "Pipe",
        "qty": 1,
        "price": 350000,
        "total": 350000,
        "createdAt": "2026-11-21T11:35:00Z",
        "createdBy": "john.doe@example.com",
        "updatedAt": "2026-11-21T11:35:00Z",
        "updatedBy": "john.doe@example.com"
      }
    ],
    "createdAt": "2026-11-21T11:35:00Z",
    "createdBy": "john.doe@example.com",
    "updatedAt": "2026-11-21T11:35:00Z",
    "updatedBy": "john.doe@example.com"
  }
}
```

---

#### 4. Update Order
Updates an existing active order's customer, status, and line items inside an atomic database transaction. The original `orderNo` remains immutable, and totals are recalculated server-side.

- **Endpoint:** `PUT /api/v1/orders/:id`
- **Authentication:** Bearer Token Required

**Path Parameters:**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | `uint` | Yes | Positive integer Order ID (`> 0`) |

**Request Body Parameters:**
| Field | Type | Required | Validation / Constraints |
| :--- | :--- | :--- | :--- |
| `customerId` | `uint` | Yes | Must reference an existing active customer (`deleted_at IS NULL`) |
| `status` | `string` | Yes | Allowed: `PENDING`, `PAID`, `SHIPPED`, `CANCELLED` |
| `items` | `array` | Yes | Minimum 1 item (`min=1`) |
| `items[].name` | `string` | Yes | Item/product name (max 50 characters) |
| `items[].qty` | `integer` | Yes | Quantity (`> 0`) |
| `items[].price` | `number` | Yes | Unit price (`> 0`) |

**Sample Request:**
```json
{
  "customerId": 1,
  "status": "PAID",
  "items": [
    {
      "name": "Paint",
      "qty": 2,
      "price": 750000
    }
  ]
}
```

**Sample Response (`200 OK`):**
```json
{
  "data": {
    "orderId": 1,
    "customerId": 1,
    "orderNo": "ORD-20261121-00001",
    "total": 1500000,
    "status": "PAID",
    "items": [
      {
        "orderItemId": 3,
        "orderId": 1,
        "name": "Mechanical Keyboard",
        "qty": 2,
        "price": 750000,
        "total": 1500000,
        "createdAt": "2026-11-21T11:45:00Z",
        "createdBy": "john.doe@example.com",
        "updatedAt": "2026-11-21T11:45:00Z",
        "updatedBy": "john.doe@example.com"
      }
    ],
    "createdAt": "2026-11-21T11:35:00Z",
    "createdBy": "john.doe@example.com",
    "updatedAt": "2026-11-21T11:45:00Z",
    "updatedBy": "john.doe@example.com"
  }
}
```

---

#### 5. Soft-Delete Order
Marks an active order as deleted by populating `deleted_at` with the current timestamp and `deleted_by` with the authenticated user's email.

- **Endpoint:** `DELETE /api/v1/orders/:id`
- **Authentication:** Bearer Token Required

**Path Parameters:**
| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | `uint` | Yes | Positive integer Order ID (`> 0`) |

**Sample Response (`200 OK`):**
```json
{
  "message": "Order deleted successfully"
}
```