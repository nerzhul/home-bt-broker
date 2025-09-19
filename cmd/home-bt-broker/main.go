package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nerzhul/home-bt-broker/internal/database"
	"github.com/nerzhul/home-bt-broker/internal/handlers"
	"github.com/nerzhul/home-bt-broker/internal/wireplumber"
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

	// Initialize WirePlumber configuration manager
	wpConfigManager, err := wireplumber.NewConfigManager()
	if err != nil {
		log.Fatalf("Failed to initialize WirePlumber config manager: %v", err)
	}

	// Ensure WirePlumber configuration exists
	if err := wpConfigManager.EnsureConfig(); err != nil {
		log.Printf("Warning: Failed to setup WirePlumber configuration: %v", err)
	}

	// Initialize Bluetooth handler
	btHandler, err := handlers.NewBluetoothHandler()
	if err != nil {
		log.Fatalf("Failed to initialize Bluetooth handler: %v", err)
	}
	defer btHandler.Close()

	// Log Bluetooth adapters at startup
	adapters, err := btHandler.GetAdaptersRaw()
	if err != nil {
		log.Printf("Could not list Bluetooth adapters: %v", err)
	} else if len(adapters) == 0 {
		log.Printf("No Bluetooth adapters found.")
	} else {
		log.Printf("Bluetooth adapters detected:")
		for _, a := range adapters {
			log.Printf("- Name: %s, Address: %s, Powered: %v, Discoverable: %v, Discovering: %v", a.Name, a.Address, a.Powered, a.Discoverable, a.Discovering)
		}
	}

	// Create Echo instance
	e := echo.New()

	e.File("/", "internal/handlers/static/index.html")

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	h := handlers.NewHandler(db)

	// Health check endpoints
	e.GET("/readyz", h.Readiness)
	e.GET("/livez", h.Liveness)

	// API routes
	api := e.Group("/api/v1")

	tokenGroup := api.Group("/tokens", handlers.AuthMiddleware(db))
	tokenGroup.POST("", h.CreateToken)
	tokenGroup.GET("", h.GetTokens)
	tokenGroup.GET("/:username", h.GetToken)
	tokenGroup.DELETE("/:username", h.DeleteToken)

	bluetoothGroup := api.Group("/bluetooth", handlers.AuthMiddleware(db))
	bluetoothGroup.GET("/adapters", btHandler.GetAdapters)
	bluetoothGroup.PATCH("/adapters/:adapter/discoverable", btHandler.SetDiscoverable)
	bluetoothGroup.PATCH("/adapters/:adapter/discovering", btHandler.SetDiscovering)
	bluetoothGroup.GET("/adapters/:adapter/devices", btHandler.GetDevices)
	bluetoothGroup.GET("/adapters/:adapter/devices/trusted", btHandler.GetTrustedDevices)
	bluetoothGroup.GET("/adapters/:adapter/devices/connected", btHandler.GetConnectedDevices)
	bluetoothGroup.POST("/adapters/:adapter/devices/:mac/pair", btHandler.PairDevice)
	bluetoothGroup.POST("/adapters/:adapter/devices/:mac/connect", btHandler.ConnectDevice)
	bluetoothGroup.POST("/adapters/:adapter/devices/:mac/trust", btHandler.TrustDevice)
	bluetoothGroup.DELETE("/adapters/:adapter/devices/:mac", btHandler.RemoveDevice)

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
