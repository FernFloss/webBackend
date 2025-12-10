package main

import (
	"context"
	"errors"
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
	"web_backend_v2/rabbit"

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

	// Initialize RabbitMQ connection and start consumer
	if err := rabbit.InitRabbitMQ(cfg); err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	rabbitCtx, rabbitCancel := context.WithCancel(context.Background())
	consumerErrCh := make(chan error, 1)
	go func() {
		consumerErrCh <- rabbit.StartConsumer(rabbitCtx, cfg.QueueName, handlers.ProcessCameraEvent)
	}()
	defer func() {
		rabbitCancel()
		if err := <-consumerErrCh; err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("RabbitMQ consumer stopped with error: %v", err)
		}
		if err := rabbit.CloseRabbitMQ(); err != nil {
			log.Printf("Error closing RabbitMQ: %v", err)
		} else {
			log.Println("RabbitMQ connection closed")
		}
	}()

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

	// Allow cross-origin requests (useful for remote frontend testing).
	router.Use(handlers.CORSMiddleware())

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Cities endpoints
		cities := v1.Group("/cities")
		{
			city := new(handlers.CityController)
			cities.GET("/", city.GetCities)
			// Buildings endpoints
			building := new(handlers.BuildingController)
			cities.GET("/:city_id/buildings", building.GetBuildingsByCity)
			auditorium := new(handlers.AuditoriumController)
			cities.GET("/:city_id/buildings/:building_id/auditories", auditorium.GetAuditoriumsByBuilding)
			cities.GET("/:city_id/buildings/:building_id/auditories/occupancy", auditorium.GetOccupancyByBuilding)
			cities.GET("/:city_id/buildings/:building_id/auditories/:auditorium_id/occupancy", auditorium.GetOccupancyByAuditorium)
			camera := new(handlers.CameraController)
			cities.GET("/:city_id/buildings/:building_id/auditories/:auditorium_id/cameras", camera.GetCamerasByAuditorium)
			cities.POST("/:city_id/buildings/:building_id/auditories/:auditorium_id/cameras", camera.AttachCamera)

		}
		// Cameras endpoints 
		cameras := v1.Group("/cameras")
		{
			camera := new(handlers.CameraController)
			cameras.GET("/", camera.GetFreeCameras)
			cameras.GET("/attached", camera.GetAttachedCameras)
			cameras.GET("/:camera_id", camera.GetCamera)
			cameras.POST("/", camera.CreateCamera)
			cameras.DELETE("/:camera_id", camera.DeleteCamera)
			cameras.DELETE("/:camera_id/attachment", camera.DetachCamera)
		}
		// Cameras endpoints (global)


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
