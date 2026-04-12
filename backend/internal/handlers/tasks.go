package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shrikarthik007/taskflow/internal/models"
)

type TaskHandler struct {
	db *pgxpool.Pool
}

func NewTaskHandler(db *pgxpool.Pool) *TaskHandler {
	return &TaskHandler{db: db}
}

// List handles GET /projects/:id/tasks
func (h *TaskHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("id")

	// Verify project ownership
	if err := h.verifyProjectAccess(c, projectID, userID); err != nil {
		return
	}

	// Build query with optional filters
	query := `SELECT id, title, description, status, priority, project_id,
	           assignee_id, due_date::text, created_at, updated_at
	           FROM tasks WHERE project_id = $1`
	args := []any{projectID}
	argIdx := 2

	if status := c.Query("status"); status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if assignee := c.Query("assignee"); assignee != "" {
		query += fmt.Sprintf(" AND assignee_id = $%d", argIdx)
		args = append(args, assignee)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		tasks = append(tasks, t)
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// Create handles POST /projects/:id/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")
	projectID := c.Param("id")

	if err := h.verifyProjectAccess(c, projectID, userID); err != nil {
		return
	}

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "validation failed",
			"fields": parseValidationError(err),
		})
		return
	}

	if req.Priority == "" {
		req.Priority = "medium"
	}

	var assigneeID *string
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		val := *req.AssigneeID
		var id string
		err := h.db.QueryRow(c.Request.Context(), `SELECT id FROM users WHERE id::text = $1 OR name ILIKE $1 LIMIT 1`, val).Scan(&id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "validation failed",
				"fields": map[string]string{"assignee_id": "User not found by name or ID"},
			})
			return
		}
		assigneeID = &id
	}

	var dueDate *string
	if req.DueDate != nil && *req.DueDate != "" {
		val := *req.DueDate
		dueDate = &val
	}

	var t models.Task
	err := h.db.QueryRow(c.Request.Context(),
		`INSERT INTO tasks (title, description, priority, project_id, assignee_id, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6::date)
		 RETURNING id, title, description, status, priority, project_id,
		           assignee_id, due_date::text, created_at, updated_at`,
		req.Title, req.Description, req.Priority, projectID, assigneeID, dueDate,
	).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, t)
}

// Update handles PATCH /tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("id")

	// Get task's project to check ownership
	var projectID string
	err := h.db.QueryRow(c.Request.Context(),
		`SELECT project_id FROM tasks WHERE id = $1`, taskID,
	).Scan(&projectID)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := h.verifyProjectAccess(c, projectID, userID); err != nil {
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	// Validate status if provided
	if req.Status != nil {
		valid := map[string]bool{"todo": true, "in_progress": true, "done": true}
		if !valid[*req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "validation failed",
				"fields": map[string]string{"status": "must be todo, in_progress, or done"},
			})
			return
		}
	}

	var t models.Task
	err = h.db.QueryRow(c.Request.Context(),
		`UPDATE tasks
		 SET title       = COALESCE($1, title),
		     description = COALESCE($2, description),
		     status      = COALESCE($3, status),
		     priority    = COALESCE($4, priority),
		     assignee_id = COALESCE($5, assignee_id),
		     due_date    = COALESCE($6::date, due_date)
		 WHERE id = $7
		 RETURNING id, title, description, status, priority, project_id,
		           assignee_id, due_date::text, created_at, updated_at`,
		req.Title, req.Description, req.Status, req.Priority, req.AssigneeID, req.DueDate, taskID,
	).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, t)
}

// Delete handles DELETE /tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	taskID := c.Param("id")

	var projectID string
	err := h.db.QueryRow(c.Request.Context(),
		`SELECT project_id FROM tasks WHERE id = $1`, taskID,
	).Scan(&projectID)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := h.verifyProjectAccess(c, projectID, userID); err != nil {
		return
	}

	if _, err := h.db.Exec(c.Request.Context(), `DELETE FROM tasks WHERE id = $1`, taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// verifyProjectAccess checks that the project exists and the user owns it.
// Writes the appropriate error response and returns a non-nil error if access is denied.
func (h *TaskHandler) verifyProjectAccess(c *gin.Context, projectID, userID string) error {
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(),
		`SELECT owner_id FROM projects WHERE id = $1`, projectID,
	).Scan(&ownerID)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return err
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return err
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return fmt.Errorf("forbidden")
	}
	return nil
}

// containsStr is a simple substring check (avoids importing strings).
func init() {
	_ = strings.Contains // keep import used
}
