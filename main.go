package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web_backend_v2/config"
	"web_backend_v2/db"
	"web_backend_v2/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	log.Println("Starting camera event processor service...")

	// Initialize database connection
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}()

	// Initialize RabbitMQ connection
	// if err := rabbit.InitRabbitMQ(cfg); err != nil {
	// 	log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	// }
	// defer func() {
	// 	if err := rabbit.CloseRabbitMQ(); err != nil {
	// 		log.Printf("Error closing RabbitMQ: %v", err)
	// 	} else {
	// 		log.Println("RabbitMQ connection closed")
	// 	}
	// }()

	// // Start consuming messages
	// if err := rabbit.ConsumeMessages(cfg.QueueName, handlers.ProcessCameraEvent); err != nil {
	// 	log.Fatalf("Failed to start consuming messages: %v", err)
	// }

	// Setup HTTP router and API endpoints
	router := setupRouter()

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: router,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("HTTP server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	log.Println("Service is running. Press Ctrl+C to stop.")

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down service...")

	// Shutdown HTTP server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	} else {
		log.Println("HTTP server shut down gracefully")
	}

	// Give some time for ongoing operations to complete
	select {
	case <-ctx.Done():
		log.Println("Shutdown timeout reached, forcing exit")
	case <-time.After(2 * time.Second):
		log.Println("Graceful shutdown completed")
	}
}

// setupRouter configures all HTTP routes
func setupRouter() *gin.Engine {
	router := gin.Default()

	// API v1 routes
	v1 := router.Group("/v1")
	{
		city := new(handlers.CityController)
		// Cities endpoints
		v1.GET("/cities", city.GetCities)

		// // Buildings endpoints
		// v1.GET("/cities/:city_id/buildings", handlers.GetBuildingsByCity)

		// // Auditoriums endpoints
		// v1.GET("/buildings/:building_id/auditoriums", handlers.GetAuditoriumsByBuilding)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	return router
}
