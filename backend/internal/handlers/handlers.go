package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"spectro-lab-backend/internal/auth"
	"spectro-lab-backend/internal/interfaces"
	"spectro-lab-backend/internal/lab"
	"spectro-lab-backend/internal/models"
	"spectro-lab-backend/internal/services"

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

// Login handles user login
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Login(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  *user,
	})
}

// CreateLab handles lab creation
func (h *Handler) CreateLab(c *gin.Context) {
	var req models.CreateLabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj := user.(*models.User)

	// Generate UUID if no name is provided
	labName := req.Name
	if labName == "" {
		labName = models.GenerateID()
	}

	lab, err := h.labService.CreateLab(labName, userObj.ID, req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lab)
}

// GetLab handles retrieving a specific lab
func (h *Handler) GetLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID required"})
		return
	}

	lab, err := h.labService.GetLab(labID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		return
	}

	// Get owner information
	owner, err := h.authService.GetUserByID(lab.OwnerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get owner information"})
		return
	}

	response := models.LabResponse{
		ID:          lab.ID,
		Name:        lab.Name,
		Status:      lab.Status,
		Owner:       *owner,
		StartedAt:   lab.StartedAt,
		EndsAt:      lab.EndsAt,
		Credentials: lab.Credentials,
	}

	c.JSON(http.StatusOK, response)
}

// GetUserLabs handles retrieving all labs for the authenticated user
func (h *Handler) GetUserLabs(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj := user.(*models.User)
	labs, err := h.labService.GetLabsByOwner(userObj.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get labs"})
		return
	}

	// Convert to response format with owner information
	responses := make([]models.LabResponse, 0, len(labs))
	for _, lab := range labs {
		response := models.LabResponse{
			ID:          lab.ID,
			Name:        lab.Name,
			Status:      lab.Status,
			Owner:       *userObj,
			StartedAt:   lab.StartedAt,
			EndsAt:      lab.EndsAt,
			Credentials: lab.Credentials,
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, responses)
}

// GetAllLabs handles retrieving all labs (admin endpoint)
func (h *Handler) GetAllLabs(c *gin.Context) {
	labs := h.labService.GetAllLabs()

	// Convert to response format with owner information
	responses := make([]models.LabResponse, 0, len(labs))
	for _, lab := range labs {
		owner, err := h.authService.GetUserByID(lab.OwnerID)
		if err != nil {
			continue // Skip labs with invalid owners
		}

		response := models.LabResponse{
			ID:          lab.ID,
			Name:        lab.Name,
			Status:      lab.Status,
			Owner:       *owner,
			StartedAt:   lab.StartedAt,
			EndsAt:      lab.EndsAt,
			Credentials: lab.Credentials,
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, responses)
}

// DeleteLab handles lab deletion
func (h *Handler) DeleteLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID required"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj := user.(*models.User)

	// Check if user owns the lab
	lab, err := h.labService.GetLab(labID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		return
	}

	if lab.OwnerID != userObj.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this lab"})
		return
	}

	err = h.labService.DeleteLab(labID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lab"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lab deleted successfully"})
}

// StopLab handles lab stopping
func (h *Handler) StopLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID required"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj := user.(*models.User)

	// Check if user owns the lab
	lab, err := h.labService.GetLab(labID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		return
	}

	if lab.OwnerID != userObj.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to stop this lab"})
		return
	}

	err = h.labService.StopLab(labID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop lab"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lab stopped successfully"})
}

// AdminStopLab handles lab stopping (admin only)
func (h *Handler) AdminStopLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID required"})
		return
	}

	// Check if lab exists
	_, err := h.labService.GetLab(labID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		return
	}

	err = h.labService.StopLab(labID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop lab"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lab stopped successfully"})
}

// AdminDeleteLab handles lab deletion (admin only)
func (h *Handler) AdminDeleteLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID required"})
		return
	}

	// Check if lab exists
	_, err := h.labService.GetLab(labID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		return
	}

	err = h.labService.DeleteLab(labID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lab"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lab deleted successfully"})
}

// CleanupLab handles lab cleanup (admin only)
func (h *Handler) CleanupLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID required"})
		return
	}

	// Check if lab exists
	lab, err := h.labService.GetLab(labID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		return
	}

	// Create cleanup context with lab data
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   labID,
		Context: context.Background(),
		Lab:     lab, // Pass lab reference for accessing stored service data
	}

	// Store lab data in context for cleanup
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_name", lab.Name)
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_duration", int(lab.EndsAt.Sub(lab.StartedAt).Minutes()))
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_owner_id", lab.OwnerID)

	// Execute cleanup
	if err := h.labService.CleanupLabServices(cleanupCtx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to cleanup lab services: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lab cleanup completed successfully"})
}

// CreateUser handles user creation (admin endpoint)
func (h *Handler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.CreateUser(req.Email, req.Name, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUsers handles retrieving all users (admin endpoint)
func (h *Handler) GetUsers(c *gin.Context) {
	users := h.authService.GetAllUsers()
	c.JSON(http.StatusOK, users)
}

// UpdateUserRole handles updating a user's role (admin endpoint)
func (h *Handler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	var req struct {
		Role models.UserRole `json:"role" binding:"required,oneof=user admin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.UpdateUserRole(userID, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// DeleteUser handles user deletion (admin endpoint)
func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	err := h.authService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// HealthCheck handles health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Spectro Lab Backend is running",
	})
}

// CleanupPaletteProject handles direct palette project cleanup (admin only)
func (h *Handler) CleanupPaletteProject(c *gin.Context) {
	var req struct {
		ProjectName string `json:"project_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name required"})
		return
	}

	// Extract sandbox ID from project name (format: lab-{sandboxID})
	projectName := req.ProjectName
	var sandboxID string

	if strings.HasPrefix(projectName, "lab-") {
		sandboxID = strings.TrimPrefix(projectName, "lab-")
	} else {
		// If not in lab- format, use the project name as sandbox ID
		sandboxID = projectName
	}

	// Create a minimal cleanup context with just the sandbox ID
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   sandboxID, // Use sandbox ID as lab ID for consistency
		Context: context.Background(),
		Lab:     nil, // No lab context needed for direct cleanup
	}

	// Execute palette project cleanup directly
	serviceManager := services.NewServiceManager()
	paletteService, exists := serviceManager.GetRegistry().GetService("palette_project")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Palette Project service not available"})
		return
	}

	if err := paletteService.ExecuteCleanup(cleanupCtx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to cleanup palette project: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Palette Project cleanup completed for %s", projectName)})
}
