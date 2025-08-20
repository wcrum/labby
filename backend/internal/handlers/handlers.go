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
// @Summary Login user
// @Description Authenticate a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
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

// GetCurrentUser returns the current authenticated user
// @Summary Get current user
// @Description Get information about the currently authenticated user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /auth/me [get]
func (h *Handler) GetCurrentUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj := user.(*models.User)
	c.JSON(http.StatusOK, *userObj)
}

// GetLabProgress returns the progress for a lab
// @Summary Get lab progress
// @Description Get the progress information for a specific lab
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} lab.LabProgress
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Progress not found"
// @Router /labs/{id}/progress [get]
func (h *Handler) GetLabProgress(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID required"})
		return
	}

	progress := h.labService.GetProgress(labID)
	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Progress not found for lab"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetLabTemplates returns all available lab templates
// @Summary Get lab templates
// @Description Get all available lab templates
// @Tags templates
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.LabTemplate
// @Router /templates [get]
func (h *Handler) GetLabTemplates(c *gin.Context) {
	templates := h.labService.GetTemplates()
	c.JSON(http.StatusOK, templates)
}

// GetLabTemplate returns a specific lab template
// @Summary Get lab template
// @Description Get a specific lab template by ID
// @Tags templates
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 200 {object} models.LabTemplate
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Template not found"
// @Router /templates/{id} [get]
func (h *Handler) GetLabTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID required"})
		return
	}

	template, exists := h.labService.GetTemplate(templateID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// CreateLabFromTemplate creates a lab from a template
// @Summary Create lab from template
// @Description Create a new lab instance from a template
// @Tags templates
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 201 {object} models.Lab
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /templates/{id}/labs [post]
func (h *Handler) CreateLabFromTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID required"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj := user.(*models.User)

	lab, err := h.labService.CreateLabFromTemplate(templateID, userObj.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, lab)
}

// CreateLab handles lab creation (deprecated - use templates instead)
func (h *Handler) CreateLab(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "Direct lab creation is not allowed. Please use a lab template instead."})
}

// GetLab handles retrieving a specific lab
// @Summary Get lab
// @Description Get a specific lab by ID
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} models.LabResponse
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Router /labs/{id} [get]
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
// @Summary Get user labs
// @Description Get all labs owned by the authenticated user
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.LabResponse
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs [get]
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
// @Summary Get all labs (Admin)
// @Description Get all labs in the system (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.LabResponse
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/labs [get]
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
// @Summary Delete lab
// @Description Delete a lab owned by the authenticated user
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Lab deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Router /labs/{id} [delete]
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
// @Summary Stop lab
// @Description Stop a lab owned by the authenticated user
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Lab stopped successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Router /labs/{id}/stop [post]
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
// @Summary Stop lab (Admin)
// @Description Stop any lab in the system (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Lab stopped successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Router /admin/labs/{id}/stop [post]
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
// @Summary Delete lab (Admin)
// @Description Delete any lab in the system (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Lab deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Router /admin/labs/{id} [delete]
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
// @Summary Cleanup lab (Admin)
// @Description Cleanup resources for a specific lab (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Lab cleanup completed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Router /admin/labs/{id}/cleanup [post]
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

// CleanupFailedLab handles cleanup of a failed lab (user can cleanup their own failed labs)
// @Summary Cleanup failed lab
// @Description Cleanup resources for a failed lab (user can cleanup their own failed labs)
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Lab cleanup completed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Router /labs/{id}/cleanup [post]
func (h *Handler) CleanupFailedLab(c *gin.Context) {
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

	// Check if lab exists and user owns it
	lab, err := h.labService.GetLab(labID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		return
	}

	if lab.OwnerID != userObj.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to cleanup this lab"})
		return
	}

	// Only allow cleanup of failed labs
	if lab.Status != models.LabStatusError {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can only cleanup failed labs"})
		return
	}

	// Create cleanup context with lab data
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   labID,
		Context: context.Background(),
		Lab:     lab,
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

	// Delete the lab from the service
	if err := h.labService.DeleteLab(labID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lab"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Failed lab cleanup completed successfully"})
}

// CreateUser handles user creation (admin endpoint)
// @Summary Create user (Admin)
// @Description Create a new user (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateUserRequest true "User creation data"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/users [post]
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
// @Summary Get all users (Admin)
// @Description Get all users in the system (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	users := h.authService.GetAllUsers()
	c.JSON(http.StatusOK, users)
}

// UpdateUserRole handles updating a user's role (admin endpoint)
// @Summary Update user role (Admin)
// @Description Update a user's role (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body object true "Role update data" schema="{role: string}"
// @Success 200 {object} map[string]interface{} "User role updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/users/{id}/role [put]
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
// @Summary Delete user (Admin)
// @Description Delete a user (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/users/{id} [delete]
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

// CleanupPaletteProject handles direct palette project cleanup (admin only)
// @Summary Cleanup palette project (Admin)
// @Description Cleanup a palette project directly (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object true "Project cleanup data" schema="{project_name: string}"
// @Success 200 {object} map[string]interface{} "Palette Project cleanup completed"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/palette-project/cleanup [post]
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
