package handlers

import (
	"fmt"
	"net/http"

	"github.com/wcrum/labby/internal/models"

	"github.com/gin-gonic/gin"
)

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

	fmt.Printf("DEBUG: Login request for email: %s, invite_code: %v\n", req.Email, req.InviteCode)

	// If invite code is provided, get the organization from the invite
	var organizationID *string
	if req.InviteCode != nil && *req.InviteCode != "" {
		invite, err := h.repo.GetInviteByID(*req.InviteCode)
		if err != nil {
			fmt.Printf("DEBUG: Failed to get invite %s: %v\n", *req.InviteCode, err)
			// Continue with login even if invite is invalid
		} else {
			organizationID = &invite.OrganizationID
			fmt.Printf("DEBUG: Found invite organization: %s\n", invite.OrganizationID)
		}
	}

	user, err := h.authService.LoginWithOrganization(req.Email, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	token, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  *user,
	})
}

// GetCurrentUser handles getting the current authenticated user
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

	// If user has no organization, assign them to the default organization
	// This handles users who access the system directly (not through invites)
	if userObj.OrganizationID == nil {
		err := h.authService.AssignUserToDefaultOrganization(userObj.ID)
		if err != nil {
			fmt.Printf("Warning: Failed to assign user to default organization: %v\n", err)
		} else {
			// Update the user object in context
			userObj.OrganizationID = stringPtr("org-default")
		}
	}

	c.JSON(http.StatusOK, userObj)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
