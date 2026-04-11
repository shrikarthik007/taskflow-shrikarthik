package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/shrikarthik007/taskflow/internal/config"
	"github.com/shrikarthik007/taskflow/internal/db"
	"github.com/shrikarthik007/taskflow/internal/handlers"
	"github.com/shrikarthik007/taskflow/internal/middleware"
)

func main() {
	// --- Structured logging -------------------------------------------------
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// --- Config -------------------------------------------------------------
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// --- Database -----------------------------------------------------------
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(cfg.DatabaseURL, "migrations"); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if err := db.RunSeed(ctx, pool, "seed.sql"); err != nil {
		slog.Warn("seed failed (non-fatal)", "error", err)
	}

	// --- Gin Router ---------------------------------------------------------
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middlewares
	r.Use(gin.Recovery())
	r.Use(requestLogger())
	r.Use(corsMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Handlers
	authH := handlers.NewAuthHandler(pool, cfg.JWTSecret)
	projectH := handlers.NewProjectHandler(pool)
	taskH := handlers.NewTaskHandler(pool)

	// Auth routes (no JWT required)
	auth := r.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
	}

	// Protected routes
	api := r.Group("/")
	api.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		// Projects
		api.GET("/projects", projectH.List)
		api.POST("/projects", projectH.Create)
		api.GET("/projects/:id", projectH.Get)
		api.PATCH("/projects/:id", projectH.Update)
		api.DELETE("/projects/:id", projectH.Delete)

		// Tasks
		api.GET("/projects/:id/tasks", taskH.List)
		api.POST("/projects/:id/tasks", taskH.Create)
		api.PATCH("/tasks/:id", taskH.Update)
		api.DELETE("/tasks/:id", taskH.Delete)
	}

	// --- HTTP Server with graceful shutdown ---------------------------------
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for SIGTERM or SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "error", err)
	}

	slog.Info("server stopped")
}

// requestLogger is a simple slog-based request logger middleware.
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method",   c.Request.Method,
			"path",     c.Request.URL.Path,
			"status",   c.Writer.Status(),
			"duration", time.Since(start).String(),
			"ip",       c.ClientIP(),
		)
	}
}

// corsMiddleware allows requests from the React frontend.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
