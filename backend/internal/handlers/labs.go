package handlers

import (
	"net/http"

	"github.com/wcrum/labby/internal/interfaces"
	"github.com/wcrum/labby/internal/lab"
	"github.com/wcrum/labby/internal/models"
	"github.com/wcrum/labby/internal/services"

	"github.com/gin-gonic/gin"
)

// GetLabs handles getting all labs for a user
// @Summary Get user labs
// @Description Get all labs for the authenticated user
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.LabResponse
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs [get]
func (h *Handler) GetLabs(c *gin.Context) {
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

	// Convert Labs to LabResponses
	labResponses := make([]*models.LabResponse, len(labs))
	for i, lab := range labs {
		labResponses[i] = h.labService.ConvertLabToResponse(lab, h.authService)
	}

	c.JSON(http.StatusOK, labResponses)
}

// GetLab handles getting a specific lab
// @Summary Get lab
// @Description Get a specific lab by ID
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} models.LabResponse
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs/{id} [get]
func (h *Handler) GetLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID is required"})
		return
	}

	// Check lab access authorization
	labInstance, ok := h.CheckLabAccess(c, labID)
	if !ok {
		return
	}

	// Convert Lab to LabResponse
	labResponse := h.labService.ConvertLabToResponse(labInstance, h.authService)
	c.JSON(http.StatusOK, labResponse)
}

// CreateLab handles creating a new lab
// @Summary Create lab
// @Description Create a new lab session
// @Tags labs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateLabRequest true "Lab creation request"
// @Success 201 {object} models.Lab
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs [post]
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
	labInstance, err := h.labService.CreateLab(req.Name, userObj.ID, req.Duration)
	if err != nil {
		if err == lab.ErrInvalidDuration {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create lab"})
		}
		return
	}

	c.JSON(http.StatusCreated, labInstance)
}

// DeleteLab handles deleting a lab
// @Summary Delete lab
// @Description Delete a lab by ID
// @Tags labs
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 204 "No content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs/{id} [delete]
func (h *Handler) DeleteLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID is required"})
		return
	}

	// Check lab access authorization
	_, ok := h.CheckLabAccess(c, labID)
	if !ok {
		return
	}

	err := h.labService.DeleteLab(labID)
	if err != nil {
		if err == lab.ErrLabNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete lab"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// StopLab handles stopping a lab
// @Summary Stop lab
// @Description Stop a lab by ID (marks it as expired)
// @Tags labs
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} models.Lab
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs/{id}/stop [post]
func (h *Handler) StopLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID is required"})
		return
	}

	// Check lab access authorization
	_, ok := h.CheckLabAccess(c, labID)
	if !ok {
		return
	}

	err := h.labService.StopLab(labID)
	if err != nil {
		if err == lab.ErrLabNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Lab not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop lab"})
		}
		return
	}

	// Get the updated lab
	labInstance, err := h.labService.GetLab(labID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated lab"})
		return
	}

	c.JSON(http.StatusOK, labInstance)
}

// GetLabProgress handles getting lab progress
// @Summary Get lab progress
// @Description Get the progress of a lab's provisioning
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} lab.LabProgress
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs/{id}/progress [get]
func (h *Handler) GetLabProgress(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID is required"})
		return
	}

	// Check lab access authorization
	_, ok := h.CheckLabAccess(c, labID)
	if !ok {
		return
	}

	progress := h.labService.GetProgress(labID)
	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lab progress not found"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// CleanupPaletteProject handles cleaning up a specific Palette Project
// @Summary Cleanup Palette Project
// @Description Clean up a specific Palette Project service
// @Tags labs
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Cleanup successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs/{id}/cleanup/palette-project [post]
func (h *Handler) CleanupPaletteProject(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID is required"})
		return
	}

	// Check lab access authorization
	labInstance, ok := h.CheckLabAccess(c, labID)
	if !ok {
		return
	}

	// Execute palette project cleanup directly
	serviceManager := services.NewServiceManager(h.repo)

	// Get the service by type (palette_project)
	paletteService, exists := serviceManager.GetServiceByType("palette_project")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Palette Project service not available"})
		return
	}

	// Create cleanup context
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   labID,
		Context: c.Request.Context(),
		Lab:     labInstance,
	}

	// Execute cleanup
	err := paletteService.ExecuteCleanup(cleanupCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup Palette Project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Palette Project cleanup completed successfully"})
}

// GetUserLabs handles getting all labs for a user (alias for GetLabs)
// @Summary Get user labs
// @Description Get all labs for the authenticated user
// @Tags labs
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Lab
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs [get]
func (h *Handler) GetUserLabs(c *gin.Context) {
	h.GetLabs(c)
}

// CleanupFailedLab handles cleaning up a failed lab
// @Summary Cleanup failed lab
// @Description Clean up a lab that has failed to provision
// @Tags labs
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Cleanup successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /labs/{id}/cleanup [post]
func (h *Handler) CleanupFailedLab(c *gin.Context) {
	labID := c.Param("id")
	if labID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lab ID is required"})
		return
	}

	// Check lab access authorization
	labInstance, ok := h.CheckLabAccess(c, labID)
	if !ok {
		return
	}

	// Create cleanup context
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   labID,
		Context: c.Request.Context(),
		Lab:     labInstance,
	}

	// Execute cleanup
	err := h.labService.CleanupLabServices(cleanupCtx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup lab"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lab cleanup completed successfully"})
}

// AdminStopLab handles stopping a lab (admin only)
// @Summary Stop lab (admin)
// @Description Stop a lab by ID (admin only)
// @Tags admin
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} models.Lab
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/labs/{id}/stop [post]
func (h *Handler) AdminStopLab(c *gin.Context) {
	h.StopLab(c)
}

// AdminDeleteLab handles deleting a lab (admin only)
// @Summary Delete lab (admin)
// @Description Delete a lab by ID (admin only)
// @Tags admin
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 204 "No content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/labs/{id} [delete]
func (h *Handler) AdminDeleteLab(c *gin.Context) {
	h.DeleteLab(c)
}

// CleanupLab handles cleaning up a lab (admin only)
// @Summary Cleanup lab (admin)
// @Description Clean up a lab by ID (admin only)
// @Tags admin
// @Security BearerAuth
// @Param id path string true "Lab ID"
// @Success 200 {object} map[string]interface{} "Cleanup successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Lab not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/labs/{id}/cleanup [post]
func (h *Handler) CleanupLab(c *gin.Context) {
	h.CleanupFailedLab(c)
}
