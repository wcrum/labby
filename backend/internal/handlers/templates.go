package handlers

import (
	"net/http"
	"spectro-lab-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetTemplates handles getting all lab templates
// @Summary Get lab templates
// @Description Get all available lab templates
// @Tags templates
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.LabTemplate
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /templates [get]
func (h *Handler) GetTemplates(c *gin.Context) {
	templates := h.labService.GetTemplates()
	c.JSON(http.StatusOK, templates)
}

// GetTemplate handles getting a specific lab template
// @Summary Get lab template
// @Description Get a specific lab template by ID
// @Tags templates
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 200 {object} models.LabTemplate
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Template not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /templates/{id} [get]
func (h *Handler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
		return
	}

	template, exists := h.labService.GetTemplate(templateID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// CreateLabFromTemplate handles creating a lab from a template
// @Summary Create lab from template
// @Description Create a new lab from a specific template
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 201 {object} models.Lab
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Template not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /templates/{id}/create-lab [post]
func (h *Handler) CreateLabFromTemplate(c *gin.Context) {
	templateID := c.Param("id")
	if templateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID is required"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lab from template"})
		return
	}

	c.JSON(http.StatusCreated, lab)
}

// GetLabTemplates handles getting all lab templates (alias for GetTemplates)
// @Summary Get lab templates
// @Description Get all available lab templates
// @Tags templates
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.LabTemplate
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /templates [get]
func (h *Handler) GetLabTemplates(c *gin.Context) {
	h.GetTemplates(c)
}

// GetLabTemplate handles getting a specific lab template (alias for GetTemplate)
// @Summary Get lab template
// @Description Get a specific lab template by ID
// @Tags templates
// @Produce json
// @Security BearerAuth
// @Param id path string true "Template ID"
// @Success 200 {object} models.LabTemplate
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Template not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /templates/{id} [get]
func (h *Handler) GetLabTemplate(c *gin.Context) {
	h.GetTemplate(c)
}
