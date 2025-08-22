package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/wcrum/labby/internal/interfaces"
	"github.com/wcrum/labby/internal/models"
	"github.com/wcrum/labby/internal/services"

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

// GetServiceConfigs returns all service configurations
// @Summary Get service configurations
// @Description Get all service configurations (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ServiceConfig
// @Router /admin/service-configs [get]
func (h *Handler) GetServiceConfigs(c *gin.Context) {
	configs := h.labService.GetServiceConfigManager().GetAllServiceConfigs()
	c.JSON(http.StatusOK, configs)
}

// GetServiceLimits returns all service limits
// @Summary Get service limits
// @Description Get all service limits (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ServiceLimit
// @Router /admin/service-limits [get]
func (h *Handler) GetServiceLimits(c *gin.Context) {
	limits := h.labService.GetServiceConfigManager().GetAllServiceLimits()
	c.JSON(http.StatusOK, limits)
}

// GetServiceUsage returns service usage information
// @Summary Get service usage
// @Description Get current usage information for all services (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ServiceUsage
// @Router /admin/service-usage [get]
func (h *Handler) GetServiceUsage(c *gin.Context) {
	usage := h.labService.GetServiceUsage()
	c.JSON(http.StatusOK, usage)
}

// CreateServiceConfig creates a new service configuration
// @Summary Create service configuration
// @Description Create a new service configuration (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body models.ServiceConfig true "Service configuration"
// @Success 201 {object} models.ServiceConfig
// @Router /admin/service-configs [post]
func (h *Handler) CreateServiceConfig(c *gin.Context) {
	var config models.ServiceConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamps
	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	h.labService.GetServiceConfigManager().AddServiceConfig(&config)
	c.JSON(http.StatusCreated, config)
}

// CreateServiceLimit creates a new service limit
// @Summary Create service limit
// @Description Create a new service limit (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit body models.ServiceLimit true "Service limit"
// @Success 201 {object} models.ServiceLimit
// @Router /admin/service-limits [post]
func (h *Handler) CreateServiceLimit(c *gin.Context) {
	var limit models.ServiceLimit
	if err := c.ShouldBindJSON(&limit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set timestamps
	now := time.Now()
	limit.CreatedAt = now
	limit.UpdatedAt = now

	h.labService.GetServiceConfigManager().AddServiceLimit(&limit)
	c.JSON(http.StatusCreated, limit)
}

// UpdateServiceConfig updates a service configuration
// @Summary Update service configuration
// @Description Update an existing service configuration (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Service configuration ID"
// @Param config body models.ServiceConfig true "Service configuration"
// @Success 200 {object} models.ServiceConfig
// @Router /admin/service-configs/{id} [put]
func (h *Handler) UpdateServiceConfig(c *gin.Context) {
	id := c.Param("id")

	var config models.ServiceConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.ID = id
	config.UpdatedAt = time.Now()

	h.labService.GetServiceConfigManager().AddServiceConfig(&config)
	c.JSON(http.StatusOK, config)
}

// UpdateServiceLimit updates a service limit
// @Summary Update service limit
// @Description Update an existing service limit (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Service limit ID"
// @Param limit body models.ServiceLimit true "Service limit"
// @Success 200 {object} models.ServiceLimit
// @Router /admin/service-limits/{id} [put]
func (h *Handler) UpdateServiceLimit(c *gin.Context) {
	id := c.Param("id")

	var limit models.ServiceLimit
	if err := c.ShouldBindJSON(&limit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit.ID = id
	limit.UpdatedAt = time.Now()

	h.labService.GetServiceConfigManager().AddServiceLimit(&limit)
	c.JSON(http.StatusOK, limit)
}

// DeleteServiceConfig deletes a service configuration
// @Summary Delete service configuration
// @Description Delete a service configuration (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Service configuration ID"
// @Success 204 "No Content"
// @Router /admin/service-configs/{id} [delete]
func (h *Handler) DeleteServiceConfig(c *gin.Context) {
	id := c.Param("id")
	h.labService.GetServiceConfigManager().RemoveServiceConfig(id)
	c.Status(http.StatusNoContent)
}

// DeleteServiceLimit deletes a service limit
// @Summary Delete service limit
// @Description Delete a service limit (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Service limit ID"
// @Success 204 "No Content"
// @Router /admin/service-limits/{id} [delete]
func (h *Handler) DeleteServiceLimit(c *gin.Context) {
	id := c.Param("id")
	h.labService.GetServiceConfigManager().RemoveServiceLimit(id)
	c.Status(http.StatusNoContent)
}
