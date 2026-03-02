# Backend Challenge: User API & Lottery System

This repository contains a high-performance User Management API and a concurrent Lottery Ticket Search system implemented in Go, following clean architecture principles.

## 🚀 Quick Start

### Prerequisites
- [Go 1.26](https://go.dev/doc/install)
- [Docker](https://www.docker.com/products/docker-desktop/) & Docker Compose
- [Make](https://www.gnu.org/software/make/) (optional, but recommended)

### Setup & Execution
1. **Clone the repository**:
   ```bash
   git clone https://github.com/iballbar/backend-challenge.git
   cd backend-challenge
   ```

2. **Configure Environment Variables**:
   Copy the example configuration to `.env`:
   ```bash
   cp .env.example .env
   ```

3. **Start Infrastructure**:
   - **Run everything (API + DB)** in Docker:
     ```bash
     make compose-up
     ```
   - **Start only Database** (for local development):
     ```bash
     make compose-db-up
     ```
   *This starts MongoDB on port 27017 and initializes indexes via `scripts/mongo-init/init.js`.*

3. **Run the Application**:
   - Standard run:
     ```bash
     make run
     ```
   - **Live Reload (Development mode)**:
     ```bash
     make air
     ```
   *The API will be available at `http://localhost:8080`.*

4. **API Documentation**:
   Access the Swagger UI at: `http://localhost:8080/swagger/index.html`



---

## 🧪 Testing

The project includes unit tests with mocks and integration tests using `testcontainers-go`.

- **Run Unit Tests**:
  ```bash
  make test
  ```
- **Run Integration Tests** (requires Docker):
  ```bash
  make test-integration
  ```
- **Run All Tests**:
  ```bash
  make test-all
  ```

---

## 🔐 JWT Guide

The API uses JWT (JSON Web Token) for protecting sensitive user endpoints.

1. **Register**: `POST /register` with `name`, `email`, and `password`.
2. **Login**: `POST /login` with `email` and `password`.
3. **Obtain Token**: The login response includes a `token` field.
4. **Use Token**: Include the token in the `Authorization` header for protected routes:
   `Authorization: Bearer <your-token>`

**Protected Routes**:
- `GET /users` (List)
- `POST /users` (Create)
- `GET /users/:id` (Get)
- `PATCH /users/:id` (Update)
- `DELETE /users/:id` (Delete)

---

## 📡 gRPC Server

The project includes a gRPC server for programmatic user management.

- **Port**: `50051`
- **Proto Definition**: `internal/adapters/grpc/proto/user.proto`
- **Services**: `UserService`
  - `CreateUser`: Creates a new user.
  - `GetUser`: Retrieves a user by ID.

### Generating gRPC Code
If you modify the `.proto` file, regenerate the Go code using:
```bash
make protoc
```

---

## 🎟️ Lottery System Design

### Problem Statement
Search for 6-digit lottery tickets using patterns (e.g., `123***`, `*4*6*8`) and ensure unique distribution among concurrent users.

### Implementation Approach
- **Data Structure**: Tickets are stored in MongoDB with individual digit fields (`d1` to `d6`) for high-performance indexing.
- **Algorithm**:
  1. The pattern (e.g., `123***`) is converted into a MongoDB filter matching specific digit fields.
  2. To ensure **Result Distribution**, we use the `rand` field (a random float) and the `$near` or sort-by-random approach. In this implementation, we use a random `rand` field with an index and a `$gte` filter to select a random starting point in the index, ensuring different users see different results.
  3. **Concurrency Control**: We use an atomic `findOneAndUpdate` with a `reservedUntil` timestamp (TTL) to "lock" tickets for a user, preventing double-selection.

### Performance Analysis
- **Indexing**: Composite indexes on `(d1, d2, d3, d4, d5, d6, rand)` and individual digit indexes ensure that even with 1M+ documents, queries stay in the millisecond range.
- **Scalability**: The use of TTL-based reservations allows the system to scale horizontally without complex distributed locking.

---

## 📝 Assumptions & Design Decisions

- **Password Hashing**: Bcrypt is used for secure password storage.
- **Validation**: Strict email format validation and pattern length (exactly 6) validation are implemented.
- **Architecture**: Follows Hexagonal / Ports & Adapters architecture for clear separation of concerns, making it easy to swap MongoDB for another database or Gin for another web framework.
- **Integration Tests**: `testcontainers-go` was chosen to provide a "real" environment for repository and API tests, ensuring database compatibility and index correctness.
