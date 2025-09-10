package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wcrum/labby/internal/models"
	"github.com/wcrum/labby/internal/services"
)

// CreateOrganization handles creating a new organization (admin only)
// @Summary Create organization (admin)
// @Description Create a new organization (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Organization creation request"
// @Success 201 {object} models.Organization
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/organizations [post]
func (h *Handler) CreateOrganization(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name, exists := req["name"]
	if !exists || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization name is required"})
		return
	}

	description := req["description"]
	domain := req["domain"]

	// Create organization using database repository
	org := &models.Organization{
		ID:          uuid.New().String()[:8], // Generate 8-character ID
		Name:        name,
		Description: description,
		Domain:      domain,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.repo.CreateOrganization(org); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	c.JSON(http.StatusCreated, org)
}

// GetOrganizations handles getting all organizations (admin only)
// @Summary Get all organizations (admin)
// @Description Get all organizations in the system (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Organization
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/organizations [get]
func (h *Handler) GetOrganizations(c *gin.Context) {
	orgs, err := h.repo.GetAllOrganizations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve organizations"})
		return
	}
	c.JSON(http.StatusOK, orgs)
}

// GetOrganization handles getting a specific organization (admin only)
// @Summary Get organization (admin)
// @Description Get a specific organization by ID (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 200 {object} models.Organization
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Router /admin/organizations/{id} [get]
func (h *Handler) GetOrganization(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID is required"})
		return
	}

	org, err := h.repo.GetOrganizationByID(orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

// CreateInvite handles creating an invitation to join an organization (admin only)
// @Summary Create invite (admin)
// @Description Create an invitation to join an organization (admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateInviteRequest true "Invite creation request"
// @Success 201 {object} models.Invite
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/organizations/{id}/invites [post]
func (h *Handler) CreateInvite(c *gin.Context) {
	orgID := c.Param("id")
	fmt.Printf("DEBUG: CreateInvite handler called with orgID: %s\n", orgID)

	if orgID == "" {
		fmt.Printf("DEBUG: Organization ID is empty\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID is required"})
		return
	}

	var req models.CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("DEBUG: Failed to bind JSON request: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("DEBUG: CreateInvite request: %+v\n", req)

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		fmt.Printf("DEBUG: User not found in context\n")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userObj := user.(*models.User)
	fmt.Printf("DEBUG: User from context: %+v\n", userObj)

	// Validate that the organization exists
	org, err := h.repo.GetOrganizationByID(orgID)
	if err != nil {
		fmt.Printf("DEBUG: Organization %s not found: %v\n", orgID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization not found"})
		return
	}
	fmt.Printf("DEBUG: Found organization: %s (ID: %s)\n", org.Name, org.ID)

	// Create invite using organization service
	orgService := services.NewOrganizationService()
	orgService.SetRepository(h.repo)

	invite, err := orgService.CreateInvite(orgID, req.Email, req.Role, userObj.ID, req.UsageLimit)
	if err != nil {
		fmt.Printf("DEBUG: CreateInvite service error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invite"})
		return
	}

	fmt.Printf("DEBUG: Successfully created invite: %+v\n", invite)
	c.JSON(http.StatusCreated, invite)
}

// GetInvite handles getting an invite by ID (public endpoint for accepting invites)
// @Summary Get invite
// @Description Get an invitation by ID (public endpoint)
// @Tags public
// @Produce json
// @Param id path string true "Invite ID"
// @Success 200 {object} models.Invite
// @Failure 404 {object} map[string]interface{} "Invite not found"
// @Router /invites/{id} [get]
func (h *Handler) GetInvite(c *gin.Context) {
	inviteID := c.Param("id")
	fmt.Printf("DEBUG: GetInvite handler called with inviteID: %s\n", inviteID)

	if inviteID == "" {
		fmt.Printf("DEBUG: Invite ID is empty\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invite ID is required"})
		return
	}

	invite, err := h.repo.GetInviteByID(inviteID)
	if err != nil {
		fmt.Printf("DEBUG: GetInvite database error: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Invite not found"})
		return
	}

	// Check if invite has expired
	if time.Now().After(invite.ExpiresAt) {
		invite.Status = "expired"
		// Update the invite status in the database
		h.repo.UpdateInvite(invite)
	}

	fmt.Printf("DEBUG: Successfully retrieved invite: %+v\n", invite)
	c.JSON(http.StatusOK, invite)
}

// AcceptInvite handles accepting an invitation (public endpoint)
// @Summary Accept invite
// @Description Accept an invitation to join an organization (public endpoint)
// @Tags public
// @Accept json
// @Produce json
// @Param id path string true "Invite ID"
// @Param request body models.AcceptInviteRequest true "Accept invite request"
// @Success 200 {object} map[string]interface{} "Invite accepted"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Invite not found"
// @Router /invites/{id}/accept [post]
func (h *Handler) AcceptInvite(c *gin.Context) {
	inviteID := c.Param("id")
	fmt.Printf("DEBUG: AcceptInvite handler called with inviteID: %s\n", inviteID)

	if inviteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invite ID is required"})
		return
	}

	var req models.AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("DEBUG: Failed to bind JSON request: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("DEBUG: AcceptInvite request: %+v\n", req)

	orgService := services.NewOrganizationService()
	orgService.SetRepository(h.repo)

	// First, get the invite to get the organization ID
	invite, err := orgService.GetInvite(inviteID)
	if err != nil {
		fmt.Printf("DEBUG: Failed to get invite: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("DEBUG: Found invite: %+v\n", invite)

	// Accept the invite (adds user to organization members)
	err = orgService.AcceptInvite(inviteID, req.UserID)
	if err != nil {
		fmt.Printf("DEBUG: Failed to accept invite: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("DEBUG: Successfully accepted invite, now updating user organization\n")
	fmt.Printf("DEBUG: User ID: %s, Organization ID: %s\n", req.UserID, invite.OrganizationID)

	// Update the user's organization ID in the auth service
	fmt.Printf("DEBUG: About to update user organization - UserID: %s, OrgID: %s\n", req.UserID, invite.OrganizationID)
	err = h.authService.UpdateUserOrganization(req.UserID, &invite.OrganizationID)
	if err != nil {
		// Log the error but don't fail the request since the invite was already accepted
		fmt.Printf("ERROR: Failed to update user organization: %v\n", err)
	} else {
		fmt.Printf("DEBUG: Successfully updated user organization\n")

		// Verify the update worked by getting the user again
		updatedUser, err := h.authService.GetUserByID(req.UserID)
		if err != nil {
			fmt.Printf("ERROR: Failed to get updated user: %v\n", err)
		} else {
			fmt.Printf("DEBUG: Verified user organization update - User: %s, OrgID: %v\n", updatedUser.Email, updatedUser.OrganizationID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite accepted successfully"})
}

// GetUserOrganization handles getting the current user's organization
// @Summary Get user organization
// @Description Get the current user's organization
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.Organization
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not in organization"
// @Router /user/organization [get]
func (h *Handler) GetUserOrganization(c *gin.Context) {
	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	userObj := user.(*models.User)

	fmt.Printf("DEBUG: GetUserOrganization called for user: %s (ID: %s)\n", userObj.Email, userObj.ID)
	fmt.Printf("DEBUG: User OrganizationID: %v\n", userObj.OrganizationID)

	// Check if user has an organization
	if userObj.OrganizationID == nil {
		fmt.Printf("DEBUG: User has no organization assigned\n")
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not a member of any organization"})
		return
	}

	fmt.Printf("DEBUG: User has organization ID: %s\n", *userObj.OrganizationID)

	orgService := services.NewOrganizationService()
	orgService.SetRepository(h.repo)
	organization, err := orgService.GetOrganization(*userObj.OrganizationID)
	if err != nil {
		fmt.Printf("DEBUG: Failed to get organization %s: %v\n", *userObj.OrganizationID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	fmt.Printf("DEBUG: Found organization: %+v\n", organization)
	c.JSON(http.StatusOK, organization)
}

// GetAllInvites handles getting all invites across all organizations (admin only)
// @Summary Get all invites (admin)
// @Description Get all invites across all organizations (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Invite
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/invites [get]
func (h *Handler) GetAllInvites(c *gin.Context) {
	orgService := services.NewOrganizationService()
	orgService.SetRepository(h.repo)

	invites, err := orgService.GetAllInvites()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invites"})
		return
	}

	c.JSON(http.StatusOK, invites)
}

// GetInviteUsageStats handles getting invite usage statistics (admin only)
// @Summary Get invite usage statistics (admin)
// @Description Get usage statistics for all invites (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.InviteUsageStats
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /admin/invites/usage [get]
func (h *Handler) GetInviteUsageStats(c *gin.Context) {
	orgService := services.NewOrganizationService()
	orgService.SetRepository(h.repo)

	stats, err := orgService.GetInviteUsageStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invite usage statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetOrganizationInvites handles getting all invites for a specific organization (admin only)
// @Summary Get organization invites (admin)
// @Description Get all invites for a specific organization (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 200 {array} models.Invite
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Organization not found"
// @Router /admin/organizations/{id}/invites [get]
func (h *Handler) GetOrganizationInvites(c *gin.Context) {
	orgID := c.Param("id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Organization ID is required"})
		return
	}

	// Verify organization exists
	_, err := h.repo.GetOrganizationByID(orgID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	orgService := services.NewOrganizationService()
	orgService.SetRepository(h.repo)

	invites, err := orgService.GetInvitesByOrganizationID(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve organization invites"})
		return
	}

	c.JSON(http.StatusOK, invites)
}
