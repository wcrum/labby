package handlers

import (
	"net/http"

	"github.com/wcrum/labby/internal/auth"
	"github.com/wcrum/labby/internal/database"
	"github.com/wcrum/labby/internal/lab"
	"github.com/wcrum/labby/internal/models"

	"github.com/gin-gonic/gin"
)

// Handler contains all the handlers
type Handler struct {
	authService *auth.Service
	oidcService *auth.OIDCService
	labService  *lab.Service
	repo        *database.Repository
}

// NewHandler creates a new handler
func NewHandler(authService *auth.Service, oidcService *auth.OIDCService, labService *lab.Service, repo *database.Repository) *Handler {
	return &Handler{
		authService: authService,
		oidcService: oidcService,
		labService:  labService,
		repo:        repo,
	}
}

// AuthMiddleware validates JWT tokens
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		user, err := h.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func (h *Handler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		userObj := user.(*models.User)

		if !h.authService.IsAdmin(userObj) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HealthCheck handles health check endpoint
// @Summary Health check
// @Description Check if the API is running
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{} "API status"
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Spectro Lab Backend is running",
	})
}
