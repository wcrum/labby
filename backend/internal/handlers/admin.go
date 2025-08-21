package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"spectro-lab-backend/internal/interfaces"
	"spectro-lab-backend/internal/models"
	"spectro-lab-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// GetAllLabs handles getting all labs (admin only)
// @Summary Get all labs (admin)
// @Description Get all labs in the system (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.LabResponse
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/labs [get]
func (h *Handler) GetAllLabs(c *gin.Context) {
	fmt.Printf("GetAllLabs: Admin request received\n")

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		fmt.Printf("GetAllLabs: No user found in context\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj := user.(*models.User)
	fmt.Printf("GetAllLabs: User %s (role: %s) requesting all labs\n", userObj.Email, userObj.Role)

	labs := h.labService.GetAllLabs()
	fmt.Printf("GetAllLabs: Found %d labs\n", len(labs))

	// Convert Labs to LabResponses
	labResponses := make([]*models.LabResponse, len(labs))
	for i, lab := range labs {
		labResponses[i] = h.labService.ConvertLabToResponse(lab, h.authService)
	}

	c.JSON(http.StatusOK, labResponses)
}

// LoadTemplates handles loading lab templates from a directory (admin only)
// @Summary Load lab templates (admin)
// @Description Load lab templates from a directory (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Directory path"
// @Success 200 {object} map[string]interface{} "Templates loaded"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/templates/load [post]
func (h *Handler) LoadTemplates(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dirPath, exists := req["directory"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Directory path is required"})
		return
	}

	err := h.labService.LoadTemplates(dirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Templates loaded successfully"})
}

// GetUsers handles getting all users (admin only)
// @Summary Get all users (admin)
// @Description Get all users in the system (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	users := h.authService.GetAllUsers()
	c.JSON(http.StatusOK, users)
}

// CreateUser handles creating a new user (admin only)
// @Summary Create user (admin)
// @Description Create a new user (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateUserRequest true "User creation request"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
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

// UpdateUserRole handles updating a user's role (admin only)
// @Summary Update user role (admin)
// @Description Update a user's role (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body map[string]string true "Role update request"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/users/{id}/role [put]
func (h *Handler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roleStr, exists := req["role"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}

	// Convert string to UserRole
	var role models.UserRole
	switch roleStr {
	case "admin":
		role = models.UserRoleAdmin
	case "user":
		role = models.UserRoleUser
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'admin' or 'user'"})
		return
	}

	err := h.authService.UpdateUserRole(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// DeleteUser handles deleting a user (admin only)
// @Summary Delete user (admin)
// @Description Delete a user by ID (admin only)
// @Tags admin
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 204 "No content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	err := h.authService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.Status(http.StatusNoContent)
}

// AdminCleanupPaletteProject handles cleaning up a Palette Project by name (admin only)
// @Summary Cleanup Palette Project by name (admin)
// @Description Clean up a Palette Project by project name (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Project cleanup request"
// @Success 200 {object} map[string]interface{} "Cleanup successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/palette-project/cleanup [post]
func (h *Handler) AdminCleanupPaletteProject(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectName, exists := req["project_name"]
	if !exists || projectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name is required"})
		return
	}

	// Execute palette project cleanup directly
	serviceManager := services.NewServiceManager()
	paletteService, exists := serviceManager.GetRegistry().GetService("palette")

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Palette Project service not available"})
		return
	}

	// Determine the actual project name and short ID
	var shortID, actualProjectName string
	if strings.HasPrefix(projectName, "lab-") {
		// If project name already has "lab-" prefix, extract the short ID
		shortID = strings.TrimPrefix(projectName, "lab-")
		actualProjectName = projectName
	} else {
		// If project name is just the short ID, construct the full project name
		shortID = projectName
		actualProjectName = fmt.Sprintf("lab-%s", projectName)
	}

	// Create cleanup context with the short ID as LabID
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   shortID, // Use short ID as LabID for this cleanup
		Context: c.Request.Context(),
		Lab:     nil, // No lab instance for admin cleanup
	}

	// Add project name to context for the service to use
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_project_name", actualProjectName)

	// Execute cleanup
	err := paletteService.ExecuteCleanup(cleanupCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to cleanup Palette Project: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Palette Project cleanup completed successfully"})
}
