package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nerzhul/home-bt-broker/internal/database"
	"github.com/nerzhul/home-bt-broker/internal/handlers"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Health check endpoints
	e.GET("/readyz", h.Readiness)
	e.GET("/livez", h.Liveness)

	// API routes
	api := e.Group("/api/v1")
	api.POST("/tokens", h.CreateToken)
	api.GET("/tokens", h.GetTokens)
	api.GET("/tokens/:username", h.GetToken)
	api.DELETE("/tokens/:username", h.DeleteToken)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}