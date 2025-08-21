package handlers

import (
	"fmt"
	"net/http"

	"spectro-lab-backend/internal/auth"
	"spectro-lab-backend/internal/lab"
	"spectro-lab-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Handler contains all the handlers
type Handler struct {
	authService *auth.Service
	labService  *lab.Service
}

// NewHandler creates a new handler
func NewHandler(authService *auth.Service, labService *lab.Service) *Handler {
	return &Handler{
		authService: authService,
		labService:  labService,
	}
}

// AuthMiddleware validates JWT tokens
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		fmt.Printf("AuthMiddleware: Request to %s\n", c.Request.URL.Path)
		fmt.Printf("AuthMiddleware: Authorization header: %s\n", token)

		if token == "" {
			fmt.Printf("AuthMiddleware: No Authorization header found\n")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
			fmt.Printf("AuthMiddleware: Removed Bearer prefix, token length: %d\n", len(token))
		}

		user, err := h.authService.ValidateToken(token)
		if err != nil {
			fmt.Printf("AuthMiddleware: Token validation failed: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		fmt.Printf("AuthMiddleware: Token validated successfully for user: %s\n", user.Email)
		c.Set("user", user)
		c.Next()
	}
}

// AdminMiddleware ensures the user is an admin
func (h *Handler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("AdminMiddleware: Checking admin access for %s\n", c.Request.URL.Path)

		user, exists := c.Get("user")
		if !exists {
			fmt.Printf("AdminMiddleware: No user found in context\n")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		userObj := user.(*models.User)
		fmt.Printf("AdminMiddleware: User %s has role %s\n", userObj.Email, userObj.Role)

		if !h.authService.IsAdmin(userObj) {
			fmt.Printf("AdminMiddleware: User %s is not an admin\n", userObj.Email)
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		fmt.Printf("AdminMiddleware: User %s is admin, proceeding\n", userObj.Email)
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
