package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"spectro-lab-backend/internal/auth"
	"spectro-lab-backend/internal/handlers"
	"spectro-lab-backend/internal/lab"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Get configuration from environment
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	port := getEnv("PORT", "8080")

	// Initialize services
	authService := auth.NewService(jwtSecret)
	labService := lab.NewService()
	handler := handlers.NewHandler(authService, labService)

	// Create a default admin user for testing
	adminUser, err := authService.CreateAdminUser("admin@spectrocloud.com", "Admin User")
	if err != nil {
		log.Printf("Failed to create admin user: %v", err)
	} else {
		log.Printf("Created admin user: %s (%s)", adminUser.Email, adminUser.Role)
	}

	// Start cleanup scheduler
	labService.StartCleanupScheduler()

	// Set up Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Add CORS middleware for development
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000", "https://tunnel.wcrum.dev"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	})

	router.Use(func(c *gin.Context) {
		corsMiddleware.HandlerFunc(c.Writer, c.Request)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Serve static files from the frontend build
	router.Static("/_next", "./static/_next")
	router.Static("/static", "./static/static")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")

	// Serve the main application for all non-API routes
	router.NoRoute(func(c *gin.Context) {
		// Don't serve static files for API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}

		// Try to serve the requested file first
		requestedPath := c.Request.URL.Path
		if requestedPath == "/" {
			requestedPath = "/index.html"
		}

		filePath := filepath.Join("./static", requestedPath)
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
			return
		}

		// If file doesn't exist, serve index.html for SPA routing
		c.File("./static/index.html")
	})

	// Health check endpoint
	router.GET("/health", handler.HealthCheck)

	// Public routes
	router.POST("/api/auth/login", handler.Login)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(handler.AuthMiddleware())
	{
		// Lab routes
		protected.POST("/labs", handler.CreateLab)
		protected.GET("/labs", handler.GetUserLabs)
		protected.GET("/labs/:id", handler.GetLab)
		protected.DELETE("/labs/:id", handler.DeleteLab)
		protected.POST("/labs/:id/stop", handler.StopLab)
	}

	// Admin routes (require both auth and admin privileges)
	admin := router.Group("/api/admin")
	admin.Use(handler.AuthMiddleware(), handler.AdminMiddleware())
	{
		admin.GET("/labs", handler.GetAllLabs)
		admin.POST("/labs/:id/stop", handler.AdminStopLab)
		admin.DELETE("/labs/:id", handler.AdminDeleteLab)
		admin.POST("/labs/:id/cleanup", handler.CleanupLab)
		admin.POST("/palette-project/cleanup", handler.CleanupPaletteProject)
		admin.GET("/users", handler.GetUsers)
		admin.POST("/users", handler.CreateUser)
		admin.PUT("/users/:id/role", handler.UpdateUserRole)
		admin.DELETE("/users/:id", handler.DeleteUser)
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
