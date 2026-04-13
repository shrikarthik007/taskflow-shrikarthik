
# TaskFlow ⚡

TaskFlow is a modern, full-stack Kanban board and project management application built for the Zomato assignment.

## Live Demo
* **Frontend (Vercel)**: [https://taskflow-shrikarthik-gi93.vercel.app](https://taskflow-shrikarthik-gi93.vercel.app)
* **Backend (Railway)**: [https://taskflow-shrikarthik-production.up.railway.app](https://taskflow-shrikarthik-production.up.railway.app)

---

## Tech Stack
- **Backend:** Go, Gin, PostgreSQL, golang-migrate, bcrypt, JWT
- **Frontend:** React, TypeScript, Vite, React Router, Axios, shadcn/ui
- **Infrastructure:** Docker, Docker Compose, Railway (backend), Vercel (frontend)

---

## Architecture Decisions

**Go + Gin for the backend:** Gin was chosen for its performance, minimal overhead, and excellent middleware support. The project structure follows a clean separation — `config`, `db`, `handlers`, `middleware` — keeping each layer independently testable and easy to reason about.

**Raw SQL with sqlx over an ORM:** Using raw SQL gives full control over queries and makes the data model explicit. ORMs like GORM can hide complexity that becomes a problem at scale. For a project of this scope, raw SQL is simpler and more transparent.

**golang-migrate for migrations:** Migrations run automatically on container startup via `db.RunMigrations()`. Both up and down migrations are provided for every change, making rollbacks safe and explicit.

**JWT stored in localStorage:** A pragmatic choice for this scope. In production, HttpOnly cookies would be preferable to mitigate XSS risk — noted as a known tradeoff.

**Monorepo structure:** Frontend and backend live in the same repo for easier review and a single `docker compose up` that spins up the full stack.

**What was intentionally left out:** Pagination, rate limiting, and WebSocket support were deprioritized to ship a complete, working core product within the 72-hour window.

---

## 1. Description
TaskFlow allows users to create projects, add tasks with priorities and due dates, and seamlessly drag-and-drop tasks across custom Kanban columns. It features a completely custom, premium glassmorphism UI built in React and a robust PostgreSQL-backed Go API.

---

## 2. Running Locally

The easiest way to run the entire application is via Docker Compose.

```bash
# 1. Clone the repository
git clone https://github.com/shrikarthik007/taskflow-shrikarthik.git
cd taskflow-shrikarthik

# 2. Set up environment variables
cp .env.example .env

# 3. Start the application
docker compose up -d
```

The application will be instantly available at `http://localhost:3000` and the API at `http://localhost:4000`.

---

## 3. Running Migrations

Database tables and schema migrations are handled **automatically on container startup**.
The backend image runs `db.RunMigrations()` on startup, which applies all pending migrations and seeds the database with test data automatically.

---

## 4. Test Credentials

If you'd like to test the application without registering a new account, you can use the pre-seeded test credentials:

* **Email:** `test@example.com`
* **Password:** `password123`

---

## 5. API Reference

All requests and responses use `application/json`. Authenticated routes require a `Bearer <token>` in the Authorization header.

### Authentication

**`POST /auth/register`**
* **Request:** `{"name": "Test", "email": "test@example.com", "password": "password123"}`
* **Response (201):** `{"message": "User registered successfully"}`

**`POST /auth/login`**
* **Request:** `{"email": "test@example.com", "password": "password123"}`
* **Response (200):** `{"token": "eyJhbG...", "user": {"id": "uuid", "name": "Test"}}`

### Projects (Authenticated)

**`GET /projects`**
* **Response (200):**
  ```json
  { "projects": [ { "id": "uuid", "name": "My Project", "description": "...", "created_at": "..." } ] }
  ```

**`POST /projects`**
* **Request:** `{"name": "New Project", "description": "Optional"}`
* **Response (201):** `{"project": { "id": "uuid", ... }}`

**`GET /projects/:id`**
* **Response (200):** `{"project": { "id": "uuid", "name": "...", ... }}`

**`PATCH /projects/:id`**
* **Request:** `{"name": "Updated Name"}`
* **Response (200):** `{"project": { "id": "uuid", ... }}`

**`DELETE /projects/:id`**
* **Response (204 No Content)**

### Tasks (Authenticated)

**`GET /projects/:id/tasks`**
* **Response (200):**
  ```json
  { "tasks": [ { "id": "uuid", "title": "Task 1", "status": "todo", "priority": "high" } ] }
  ```

**`POST /projects/:id/tasks`**
* **Request:** `{"title": "Fix bug", "description": "...", "priority": "high", "status": "todo", "due_date": "2026-04-15"}`
* **Response (201):** `{"task": { "id": "uuid", ... }}`

**`PATCH /tasks/:id`**
* **Request:** `{"status": "in_progress"}`
* **Response (200):** `{"task": { "status": "in_progress", ... }}`

**`DELETE /tasks/:id`**
* **Response (204 No Content)**

---

## 6. What I'd Do With More Time

1. **WebSockets:** Real-time task updates so multiple users see changes instantaneously.
2. **Rate Limiting:** Redis-backed rate limiter on auth endpoints to prevent brute-force attacks.
3. **Comprehensive Testing:** Unit tests for Go handlers using `testify` and E2E tests using `Playwright`.
4. **CI/CD Pipeline:** GitHub Actions workflows to enforce linting, type checks, and tests on every PR.
5. **Pagination:** Add `?page=&limit=` support on list endpoints for scalability.
6. **HttpOnly Cookies:** Replace localStorage JWT storage with HttpOnly cookies to eliminate XSS risk.

```
