package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/igorfazlyev/dental-marketplace/internal/database"
	"github.com/igorfazlyev/dental-marketplace/internal/routes"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate database schema
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed sample data
	if err := database.SeedData(); err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Gin router
	r := gin.Default()

	// Load templates and static files
	r.LoadHTMLGlob("templates/**/*")
	r.Static("/static", "./static")

	uploadPath := os.Getenv("UPLOAD_PATH")
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	r.Static("/uploads", uploadPath)

	// Setup routes
	routes.SetupRoutes(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	log.Printf("Visit http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
