package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shrikarthik007/taskflow/internal/models"
)

type ProjectHandler struct {
	db *pgxpool.Pool
}

func NewProjectHandler(db *pgxpool.Pool) *ProjectHandler {
	return &ProjectHandler{db: db}
}

// List handles GET /projects
func (h *ProjectHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.db.Query(c.Request.Context(),
		`SELECT id, name, description, owner_id, created_at
		 FROM projects WHERE owner_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	defer rows.Close()

	projects := []models.Project{}
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		projects = append(projects, p)
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

// Create handles POST /projects
func (h *ProjectHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "validation failed",
			"fields": parseValidationError(err),
		})
		return
	}

	var p models.Project
	err := h.db.QueryRow(c.Request.Context(),
		`INSERT INTO projects (name, description, owner_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, description, owner_id, created_at`,
		req.Name, req.Description, userID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, p)
}

// Get handles GET /projects/:id
func (h *ProjectHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var p models.Project
	err := h.db.QueryRow(c.Request.Context(),
		`SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if p.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// Update handles PATCH /projects/:id
func (h *ProjectHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	// Check ownership first
	var ownerID string
	err := h.db.QueryRow(c.Request.Context(),
		`SELECT owner_id FROM projects WHERE id = $1`, id,
	).Scan(&ownerID)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	var p models.Project
	err = h.db.QueryRow(c.Request.Context(),
		`UPDATE projects
		 SET name        = COALESCE($1, name),
		     description = COALESCE($2, description)
		 WHERE id = $3
		 RETURNING id, name, description, owner_id, created_at`,
		req.Name, req.Description, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// Delete handles DELETE /projects/:id
func (h *ProjectHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	var ownerID string
	err := h.db.QueryRow(c.Request.Context(),
		`SELECT owner_id FROM projects WHERE id = $1`, id,
	).Scan(&ownerID)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if _, err := h.db.Exec(c.Request.Context(), `DELETE FROM projects WHERE id = $1`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
