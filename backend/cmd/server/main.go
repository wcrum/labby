package main

// @title           Spectro Lab API
// @version         1.0
// @description     A REST API for managing Spectro Cloud lab environments
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/wcrum/labby/internal/auth"
	"github.com/wcrum/labby/internal/database"
	"github.com/wcrum/labby/internal/handlers"
	"github.com/wcrum/labby/internal/lab"

	_ "github.com/wcrum/labby/docs" // This will be generated

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Get configuration from environment
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	port := getEnv("PORT", "8080")

	// Initialize database
	log.Println("Initializing database connection...")
	dbConfig := database.NewConfig()
	db, err := database.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run database migrations
	log.Println("Running database migrations...")
	repo := database.NewRepository(db)
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize services
	authService := auth.NewService(jwtSecret, repo)
	labService := lab.NewService(repo)

	// Load lab templates
	if err := labService.LoadTemplates("./templates"); err != nil {
		log.Printf("Warning: Failed to load templates: %v", err)
	}

	// Load service configurations
	log.Printf("Loading service configurations from ./service-configs")
	if err := labService.LoadServiceConfigs("./service-configs"); err != nil {
		log.Printf("Warning: Failed to load service configs: %v", err)
	} else {
		log.Printf("Successfully loaded service configurations")
	}

	// Load service limits
	log.Printf("Loading service limits from ./service-configs")
	if err := labService.LoadServiceLimits("./service-configs"); err != nil {
		log.Printf("Warning: Failed to load service limits: %v", err)
	} else {
		log.Printf("Successfully loaded service limits")
	}

	// Enrich templates with service type information
	log.Printf("Enriching templates with service type information")
	labService.EnrichTemplatesWithServiceTypes()

	handler := handlers.NewHandler(authService, labService, repo)

	// Create a default admin user
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

		// Check if this is a valid Next.js route by looking for the corresponding HTML file
		// Next.js generates HTML files for each route in the out directory
		htmlPath := filepath.Join("./static", requestedPath, "index.html")
		if _, err := os.Stat(htmlPath); err == nil {
			c.File(htmlPath)
			return
		}

		// If no valid route found, serve 404 page
		c.File("./static/404.html")
	})

	// Health check endpoint
	router.GET("/health", handler.HealthCheck)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	router.POST("/api/auth/login", handler.Login)

	// Public invite routes
	router.GET("/api/invites/:id", handler.GetInvite)
	router.POST("/api/invites/:id/accept", handler.AcceptInvite)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(handler.AuthMiddleware())
	{
		// Auth routes
		protected.GET("/auth/me", handler.GetCurrentUser)

		// User routes
		protected.GET("/user/organization", handler.GetUserOrganization)

		// Lab routes
		protected.POST("/labs", handler.CreateLab)
		protected.GET("/labs", handler.GetUserLabs)
		protected.GET("/labs/:id", handler.GetLab)
		protected.GET("/labs/:id/progress", handler.GetLabProgress)
		protected.DELETE("/labs/:id", handler.DeleteLab)
		protected.POST("/labs/:id/stop", handler.StopLab)
		protected.POST("/labs/:id/cleanup", handler.CleanupFailedLab)

		// Template routes
		protected.GET("/templates", handler.GetLabTemplates)
		protected.GET("/templates/:id", handler.GetLabTemplate)
		protected.POST("/templates/:id/labs", handler.CreateLabFromTemplate)
	}

	// Admin routes (require both auth and admin privileges)
	admin := router.Group("/api/admin")
	admin.Use(handler.AuthMiddleware(), handler.AdminMiddleware())
	{
		admin.GET("/labs", handler.GetAllLabs)
		admin.POST("/labs/:id/stop", handler.AdminStopLab)
		admin.DELETE("/labs/:id", handler.AdminDeleteLab)
		admin.POST("/labs/:id/cleanup", handler.CleanupLab)
		// Cleanup endpoints
		admin.POST("/cleanup/service", handler.AdminCleanupService)
		admin.POST("/cleanup/service-by-id", handler.AdminCleanupServiceByID)
		admin.POST("/cleanup/lab", handler.AdminCleanupByLab)
		admin.GET("/cleanup/services", handler.AdminGetAvailableServices)
		admin.GET("/users", handler.GetUsers)
		admin.POST("/users", handler.CreateUser)
		admin.PUT("/users/:id/role", handler.UpdateUserRole)
		admin.DELETE("/users/:id", handler.DeleteUser)

		// Organization management
		admin.GET("/organizations", handler.GetOrganizations)
		admin.POST("/organizations", handler.CreateOrganization)
		admin.GET("/organizations/:id", handler.GetOrganization)
		admin.POST("/organizations/:id/invites", handler.CreateInvite)

		// Service configuration and limit management
		admin.GET("/service-configs", handler.GetServiceConfigs)
		admin.POST("/service-configs", handler.CreateServiceConfig)
		admin.PUT("/service-configs/:id", handler.UpdateServiceConfig)
		admin.DELETE("/service-configs/:id", handler.DeleteServiceConfig)
		admin.GET("/service-limits", handler.GetServiceLimits)
		admin.POST("/service-limits", handler.CreateServiceLimit)
		admin.PUT("/service-limits/:id", handler.UpdateServiceLimit)
		admin.DELETE("/service-limits/:id", handler.DeleteServiceLimit)
		admin.GET("/service-usage", handler.GetServiceUsage)
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
