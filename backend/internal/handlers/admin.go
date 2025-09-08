package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wcrum/labby/internal/interfaces"
	"github.com/wcrum/labby/internal/models"
	"github.com/wcrum/labby/internal/services"

	"github.com/gin-gonic/gin"
)

// AdminCleanupRequest represents a request to cleanup any service
type AdminCleanupRequest struct {
	ServiceType     string            `json:"service_type" binding:"required"` // Required: service type (e.g., "palette_project")
	ServiceConfigID string            `json:"service_config_id,omitempty"`     // Optional: specific service config ID
	LabID           string            `json:"lab_id,omitempty"`                // Optional: lab ID for context
	Parameters      map[string]string `json:"parameters,omitempty"`            // Optional: custom parameters for cleanup
}

// AdminCleanupServiceByIDRequest represents a request to cleanup a specific service config by ID and lab UUID
type AdminCleanupServiceByIDRequest struct {
	ServiceConfigID string `json:"service_config_id" binding:"required"` // Required: service config ID (e.g., "palette-project")
	LabID           string `json:"lab_id" binding:"required"`            // Required: lab UUID
}

// AdminCleanupByLabRequest represents a simplified request to cleanup by lab UUID only
type AdminCleanupByLabRequest struct {
	LabID string `json:"lab_id" binding:"required"` // Required: lab UUID
}

// ServiceInfo represents information about a service configuration
type ServiceInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// ServiceTypeInfo represents information about a service type and its configurations
type ServiceTypeInfo struct {
	Type       string          `json:"type"`
	Configs    []ServiceInfo   `json:"configs"`
	Parameters []ParameterInfo `json:"parameters"`
}

// ParameterInfo represents information about a cleanup parameter
type ParameterInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Example     string `json:"example,omitempty"`
}

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
// @Description Get all users in the system with organization information (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.UserWithOrganization
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/users [get]
func (h *Handler) GetUsers(c *gin.Context) {
	users := h.authService.GetAllUsers()

	// Convert users to UserWithOrganization format
	usersWithOrg := make([]*models.UserWithOrganization, len(users))
	orgService := services.NewOrganizationService()

	for i, user := range users {
		userWithOrg := &models.UserWithOrganization{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		// If user has an organization, fetch organization details
		if user.OrganizationID != nil {
			org, err := orgService.GetOrganization(*user.OrganizationID)
			if err == nil {
				userWithOrg.Organization = org
			}
		}

		usersWithOrg[i] = userWithOrg
	}

	c.JSON(http.StatusOK, usersWithOrg)
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

// AdminCleanupService handles flexible cleanup for any service (admin only)
// @Summary Cleanup any service with custom parameters (admin)
// @Description Clean up resources for any service type with custom input parameters (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AdminCleanupRequest true "Service cleanup request"
// @Success 200 {object} map[string]interface{} "Cleanup successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/cleanup/service [post]
func (h *Handler) AdminCleanupService(c *gin.Context) {
	var req AdminCleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if req.ServiceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_type is required"})
		return
	}

	// Get service manager
	serviceConfigManager := h.labService.GetServiceConfigManager()
	serviceManager := services.NewServiceManager(serviceConfigManager)

	// Get the service by type
	service, exists := serviceManager.GetServiceByType(req.ServiceType)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Service type '%s' not found. Available types: palette_project, palette_tenant, proxmox_user, terraform_cloud, guacamole", req.ServiceType)})
		return
	}

	// Configure service with service config if provided
	if req.ServiceConfigID != "" {
		serviceConfig, exists := serviceConfigManager.GetServiceConfig(req.ServiceConfigID)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Service config '%s' not found", req.ServiceConfigID)})
			return
		}

		// Configure the service with the service config
		if configurableService, ok := service.(interface {
			ConfigureFromServiceConfig(*models.ServiceConfig)
		}); ok {
			configurableService.ConfigureFromServiceConfig(serviceConfig)
		}
	}

	// Create cleanup context
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   req.LabID,
		Context: c.Request.Context(),
		Lab:     nil, // No lab instance for admin cleanup
	}

	// Add custom parameters to context
	for key, value := range req.Parameters {
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, key, value)
	}

	// Execute cleanup
	err := service.ExecuteCleanup(cleanupCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to cleanup %s service: %v", req.ServiceType, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      fmt.Sprintf("%s cleanup completed successfully", req.ServiceType),
		"service_type": req.ServiceType,
		"lab_id":       req.LabID,
		"parameters":   req.Parameters,
	})
}

// AdminGetAvailableServices returns all available service types and their cleanup parameters
// @Summary Get available services for cleanup (admin)
// @Description Get all available service types and their cleanup parameters (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Available services"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/cleanup/services [get]
func (h *Handler) AdminGetAvailableServices(c *gin.Context) {
	serviceConfigManager := h.labService.GetServiceConfigManager()

	// Get all service configs
	serviceConfigs := serviceConfigManager.GetAllServiceConfigs()

	// Group by service type
	servicesByType := make(map[string][]ServiceInfo)
	for _, config := range serviceConfigs {
		serviceType := config.Type
		if servicesByType[serviceType] == nil {
			servicesByType[serviceType] = []ServiceInfo{}
		}
		servicesByType[serviceType] = append(servicesByType[serviceType], ServiceInfo{
			ID:          config.ID,
			Name:        config.Name,
			Description: config.Description,
			IsActive:    config.IsActive,
		})
	}

	// Get cleanup parameters for each service type
	serviceTypes := []ServiceTypeInfo{}
	for serviceType, configs := range servicesByType {
		serviceInfo := ServiceTypeInfo{
			Type:       serviceType,
			Configs:    configs,
			Parameters: getCleanupParametersForServiceType(serviceType),
		}
		serviceTypes = append(serviceTypes, serviceInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"available_services": serviceTypes,
		"usage": gin.H{
			"endpoint":        "/admin/cleanup/service",
			"method":          "POST",
			"required_fields": []string{"service_type"},
			"optional_fields": []string{"service_config_id", "lab_id", "parameters"},
			"example": gin.H{
				"service_type":      "palette_project",
				"service_config_id": "palette-project",
				"lab_id":            "abc123",
				"parameters": gin.H{
					"project_name": "lab-abc123",
					"user_email":   "lab+abc123@spectrocloud.com",
				},
			},
		},
	})
}

// AdminCleanupServiceByID handles cleanup for a specific service config by ID and lab UUID (admin only)
// @Summary Cleanup a specific service config for a lab by UUID (admin)
// @Description Clean up resources for a specific service config using the service config ID and lab UUID - automatically constructs all resource names (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AdminCleanupServiceByIDRequest true "Service cleanup request"
// @Success 200 {object} map[string]interface{} "Cleanup successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/cleanup/service-by-id [post]
func (h *Handler) AdminCleanupServiceByID(c *gin.Context) {
	var req AdminCleanupServiceByIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if req.ServiceConfigID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_config_id is required"})
		return
	}
	if req.LabID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lab_id is required"})
		return
	}

	// Get service manager
	serviceConfigManager := h.labService.GetServiceConfigManager()
	serviceManager := services.NewServiceManager(serviceConfigManager)

	// Get the specific service config
	serviceConfig, exists := serviceConfigManager.GetServiceConfig(req.ServiceConfigID)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Service config '%s' not found", req.ServiceConfigID)})
		return
	}

	// Get the service by type from the service config
	service, exists := serviceManager.GetServiceByType(serviceConfig.Type)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Service type '%s' not found for config '%s'", serviceConfig.Type, req.ServiceConfigID)})
		return
	}

	// Configure service with the specific service config
	if configurableService, ok := service.(interface {
		ConfigureFromServiceConfig(*models.ServiceConfig)
	}); ok {
		configurableService.ConfigureFromServiceConfig(serviceConfig)
	}

	// Create cleanup context with auto-constructed parameters
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   req.LabID,
		Context: c.Request.Context(),
		Lab:     nil, // No lab instance for admin cleanup
	}

	// Auto-construct parameters based on service type
	switch serviceConfig.Type {
	case "palette_project":
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_project_name", fmt.Sprintf("lab-%s", req.LabID))
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_user_email", fmt.Sprintf("lab+%s@spectrocloud.com", req.LabID))
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_api_key_name", fmt.Sprintf("lab-%s-api-key", req.LabID))
	case "palette_tenant":
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_tenant_id", fmt.Sprintf("tenant-%s", req.LabID))
	case "proxmox_user":
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "proxmox_user_username", fmt.Sprintf("lab-%s@pve", req.LabID))
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "proxmox_pool_name", fmt.Sprintf("lab-%s-pool", req.LabID))
	case "terraform_cloud":
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "terraform_workspace_name", fmt.Sprintf("lab-%s-workspace", req.LabID))
	case "guacamole":
		cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "guacamole_username", fmt.Sprintf("lab-%s", req.LabID))
	}

	// Execute cleanup
	err := service.ExecuteCleanup(cleanupCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to cleanup service config '%s': %v", req.ServiceConfigID, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":                    fmt.Sprintf("Service config '%s' cleanup completed successfully", req.ServiceConfigID),
		"service_config_id":          req.ServiceConfigID,
		"service_type":               serviceConfig.Type,
		"lab_id":                     req.LabID,
		"auto_constructed_resources": getAutoConstructedResources(serviceConfig.Type, req.LabID),
	})
}

// getAutoConstructedResources returns the list of auto-constructed resource names for a service type
func getAutoConstructedResources(serviceType, labID string) map[string]string {
	resources := make(map[string]string)

	switch serviceType {
	case "palette_project":
		resources["project_name"] = fmt.Sprintf("lab-%s", labID)
		resources["user_email"] = fmt.Sprintf("lab+%s@spectrocloud.com", labID)
		resources["api_key_name"] = fmt.Sprintf("lab-%s-api-key", labID)
	case "palette_tenant":
		resources["tenant_id"] = fmt.Sprintf("tenant-%s", labID)
	case "proxmox_user":
		resources["username"] = fmt.Sprintf("lab-%s@pve", labID)
		resources["pool_name"] = fmt.Sprintf("lab-%s-pool", labID)
	case "terraform_cloud":
		resources["workspace_name"] = fmt.Sprintf("lab-%s-workspace", labID)
	case "guacamole":
		resources["username"] = fmt.Sprintf("lab-%s", labID)
	}

	return resources
}

// AdminCleanupByLab handles simplified cleanup by lab UUID only (admin only)
// @Summary Cleanup all services for a lab by UUID (admin)
// @Description Clean up all resources for a lab using just the lab UUID - automatically constructs all resource names (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AdminCleanupByLabRequest true "Lab cleanup request"
// @Success 200 {object} map[string]interface{} "Cleanup successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/cleanup/lab [post]
func (h *Handler) AdminCleanupByLab(c *gin.Context) {
	var req AdminCleanupByLabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate required fields
	if req.LabID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lab_id is required"})
		return
	}

	// Get service manager
	serviceConfigManager := h.labService.GetServiceConfigManager()
	serviceManager := services.NewServiceManager(serviceConfigManager)

	// Get all available service types
	serviceConfigs := serviceConfigManager.GetAllServiceConfigs()
	serviceTypes := make(map[string]bool)
	for _, config := range serviceConfigs {
		serviceTypes[config.Type] = true
	}

	// Track cleanup results
	results := make(map[string]interface{})
	errors := make(map[string]string)

	// Cleanup each service type
	for serviceType := range serviceTypes {
		service, exists := serviceManager.GetServiceByType(serviceType)
		if !exists {
			errors[serviceType] = "Service not available"
			continue
		}

		// Configure service with service config if available
		for _, config := range serviceConfigs {
			if config.Type == serviceType && config.IsActive {
				if configurableService, ok := service.(interface {
					ConfigureFromServiceConfig(*models.ServiceConfig)
				}); ok {
					configurableService.ConfigureFromServiceConfig(config)
				}
				break
			}
		}

		// Create cleanup context with auto-constructed parameters
		cleanupCtx := &interfaces.CleanupContext{
			LabID:   req.LabID,
			Context: c.Request.Context(),
			Lab:     nil, // No lab instance for admin cleanup
		}

		// Auto-construct common parameters based on service type
		switch serviceType {
		case "palette_project":
			cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_project_name", fmt.Sprintf("lab-%s", req.LabID))
			cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_user_email", fmt.Sprintf("lab+%s@spectrocloud.com", req.LabID))
			cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_api_key_name", fmt.Sprintf("lab-%s-api-key", req.LabID))
		case "palette_tenant":
			cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "palette_tenant_id", fmt.Sprintf("tenant-%s", req.LabID))
		case "proxmox_user":
			cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "proxmox_user_username", fmt.Sprintf("lab-%s@pve", req.LabID))
			cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "proxmox_pool_name", fmt.Sprintf("lab-%s-pool", req.LabID))
		case "terraform_cloud":
			cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "terraform_workspace_name", fmt.Sprintf("lab-%s-workspace", req.LabID))
		case "guacamole":
			cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "guacamole_username", fmt.Sprintf("lab-%s", req.LabID))
		}

		// Execute cleanup
		err := service.ExecuteCleanup(cleanupCtx)
		if err != nil {
			errors[serviceType] = err.Error()
		} else {
			results[serviceType] = "Cleanup completed successfully"
		}
	}

	// Prepare response
	response := gin.H{
		"message":    "Lab cleanup completed",
		"lab_id":     req.LabID,
		"results":    results,
		"errors":     errors,
		"successful": len(results),
		"failed":     len(errors),
	}

	// Determine HTTP status
	if len(errors) > 0 && len(results) == 0 {
		c.JSON(http.StatusInternalServerError, response)
	} else if len(errors) > 0 {
		c.JSON(http.StatusPartialContent, response)
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// getCleanupParametersForServiceType returns the cleanup parameters for a specific service type
func getCleanupParametersForServiceType(serviceType string) []ParameterInfo {
	switch serviceType {
	case "palette_project":
		return []ParameterInfo{
			{
				Name:        "project_name",
				Description: "Name of the Palette project to cleanup (e.g., 'lab-abc123')",
				Required:    false,
				Example:     "lab-abc123",
			},
			{
				Name:        "user_email",
				Description: "Email of the user to cleanup (e.g., 'lab+abc123@spectrocloud.com')",
				Required:    false,
				Example:     "lab+abc123@spectrocloud.com",
			},
			{
				Name:        "api_key_name",
				Description: "Name of the API key to cleanup (e.g., 'lab-abc123-api-key')",
				Required:    false,
				Example:     "lab-abc123-api-key",
			},
		}
	case "palette_tenant":
		return []ParameterInfo{
			{
				Name:        "tenant_id",
				Description: "ID of the tenant to cleanup (e.g., 'tenant-abc123')",
				Required:    false,
				Example:     "tenant-abc123",
			},
		}
	case "proxmox_user":
		return []ParameterInfo{
			{
				Name:        "username",
				Description: "Proxmox username to cleanup (e.g., 'lab-abc123@pve')",
				Required:    false,
				Example:     "lab-abc123@pve",
			},
			{
				Name:        "pool_name",
				Description: "Resource pool name to cleanup (e.g., 'lab-abc123-pool')",
				Required:    false,
				Example:     "lab-abc123-pool",
			},
		}
	case "terraform_cloud":
		return []ParameterInfo{
			{
				Name:        "workspace_name",
				Description: "Terraform Cloud workspace name to cleanup (e.g., 'lab-abc123-workspace')",
				Required:    false,
				Example:     "lab-abc123-workspace",
			},
		}
	case "guacamole":
		return []ParameterInfo{
			{
				Name:        "username",
				Description: "Guacamole username to cleanup (e.g., 'lab-abc123')",
				Required:    false,
				Example:     "lab-abc123",
			},
		}
	default:
		return []ParameterInfo{
			{
				Name:        "lab_id",
				Description: "Lab ID for context (used to construct resource names)",
				Required:    false,
				Example:     "abc123",
			},
		}
	}
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
