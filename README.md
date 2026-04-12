# TaskFlow ⚡

TaskFlow is a modern, full-stack Kanban board and project management application built for the Zomato assignment. 

## Live Demo
* **Frontend (Vercel)**: [https://taskflow-shrikarthik-gi93.vercel.app](https://taskflow-shrikarthik-gi93.vercel.app)
* **Backend (Railway)**: [https://taskflow-shrikarthik-production.up.railway.app](https://taskflow-shrikarthik-production.up.railway.app)

---

## 1. Description
TaskFlow allows users to create projects, add tasks with priorities and due dates, and seamlessly drag-and-drop tasks across custom Kanban columns. It features a completely custom, premium glassmorphism UI built in React and a robust PostgreSQL-backed Go API.

---

## 2. Running Locally

The easiest way to run the entire application is via Docker Compose.

```bash
# 1. Clone the repository
git clone https://github.com/your-username/taskflow.git
cd taskflow

# 2. Set up environment variables
cp .env.example .env

# 3. Start the application
docker compose up -d
```

The application will be instantly available at `http://localhost:3000` and the API at `http://localhost:4000`.

---

## 3. Running Migrations

Database tables and schema migrations are handled **automatically on container startup**. 
The `go.mod` backend image includes an initialization `seed.sql` script that automatically seeds the PostgreSQL database to ensure that tables for `users`, `projects`, and `tasks` are available as soon as Docker reports the database is healthy.

---

## 4. Test Credentials

If you'd like to test the application without registering a new account, you can use the pre-seeded test credentials:

* **Email:** `test@example.com`
* **Password:** `password123`

---

## 5. API Reference

All requests and responses use `application/json`. Authenticated routes require a `Bearer <token>` in the Authorization header.

### Authentication

**`POST /api/register`**
* **Request:** `{"name": "Test", "email": "test@example.com", "password": "password123"}`
* **Response (201):** `{"message": "User registered successfully"}`

**`POST /api/login`**
* **Request:** `{"email": "test@example.com", "password": "password123"}`
* **Response (200):** `{"token": "eyJhbG...", "user": {"id": "uuid", "name": "Test"}}`

### Projects (Authenticated)

**`GET /api/projects`**
* **Response (200):** 
  ```json
  { "projects": [ { "id": "uuid", "name": "My Project", "description": "...", "created_at": "..." } ] }
  ```

**`POST /api/projects`**
* **Request:** `{"name": "New Project", "description": "Optional"}`
* **Response (201):** `{"project": { "id": "uuid", ... }}`

**`GET /api/projects/:id`**
* **Response (200):** `{"project": { "id": "uuid", "name": "...", ... }}`

### Tasks (Authenticated)

**`GET /api/projects/:id/tasks`**
* **Response (200):** 
  ```json
  { "tasks": [ { "id": "uuid", "project_id": "uuid", "title": "Task 1", "status": "todo", "priority": "high", "assignee_id": "uuid" } ] }
  ```

**`POST /api/projects/:id/tasks`**
* **Request:** `{"title": "Fix bug", "description": "...", "priority": "high", "status": "todo", "dueDate": "2026-04-15"}`
* **Response (201):** `{"task": { "id": "uuid", ... }}`

**`PATCH /api/tasks/:id`**
* **Request:** `{"status": "in_progress"}`
* **Response (200):** `{"task": { "status": "in_progress", ... }}`

**`DELETE /api/tasks/:id`**
* **Response (204 No Content)**

---

## 6. What I'd Do With More Time

If given more time past this evaluation, I would implement:
1. **WebSockets:** Replace standard HTTP polling/mutations with real-time WebSockets so that multiple users on the same Kanban board see tasks move instantaneously.
2. **Rate Limiting:** Protect the authentication endpoints against brute-force attacks and DDOS using a Redis-backed rate limiter in Go.
3. **Comprehensive Testing:** Add isolated unit tests for Go components (using `testify`) and rigorous E2E browser tests for the drag-and-drop interface (using `Cypress` or `Playwright`).
4. **Automated CI/CD Pipeline:** Configure strict GitHub Actions workflows to block merges if type definitions, linting rules, or tests fail, ensuring production code is pristine.
