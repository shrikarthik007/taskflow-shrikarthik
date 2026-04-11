package models

import (
	"time"
)

// User represents a registered user in the system.
type User struct {
	ID        string    `json:"id"         db:"id"`
	Name      string    `json:"name"       db:"name"`
	Email     string    `json:"email"      db:"email"`
	Password  string    `json:"-"          db:"password"` // never serialised to JSON
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Project belongs to an owner (User) and contains Tasks.
type Project struct {
	ID          string    `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	Description *string   `json:"description" db:"description"`
	OwnerID     string    `json:"owner_id"    db:"owner_id"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
}

// Task belongs to a Project and can be assigned to a User.
type Task struct {
	ID          string    `json:"id"           db:"id"`
	Title       string    `json:"title"        db:"title"`
	Description *string   `json:"description"  db:"description"`
	Status      string    `json:"status"       db:"status"`
	Priority    string    `json:"priority"     db:"priority"`
	ProjectID   string    `json:"project_id"   db:"project_id"`
	AssigneeID  *string   `json:"assignee_id"  db:"assignee_id"`
	DueDate     *string   `json:"due_date"     db:"due_date"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"   db:"updated_at"`
}

// ---- Request bodies ----

type RegisterRequest struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description *string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateTaskRequest struct {
	Title       string  `json:"title"       binding:"required"`
	Description *string `json:"description"`
	Priority    string  `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

// ---- Response bodies ----

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}
