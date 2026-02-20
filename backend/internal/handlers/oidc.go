package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wcrum/labby/internal/models"
)

// OIDCLogin initiates OIDC login flow following Dex documentation
// @Summary OIDC Login
// @Description Initiate OIDC login flow with Dex
// @Tags auth
// @Produce json
// @Param invite_code query string false "Optional invite code for organization assignment"
// @Success 302 {string} string "Redirect to OIDC provider"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/oidc/login [get]
func (h *Handler) OIDCLogin(c *gin.Context) {
	// Get invite code from query parameter if provided
	inviteCode := c.Query("invite_code")

	// Generate state parameter to prevent CSRF attacks (following Dex docs)
	state := uuid.New().String()

	// Store invite code in session/state for later use
	// For now, we'll encode it in the state parameter
	if inviteCode != "" {
		state = fmt.Sprintf("%s:%s", state, inviteCode)
	}

	// Get OIDC authorization URL following Dex documentation pattern
	authURL := h.oidcService.GetAuthURL(state)

	// Redirect to OIDC provider (following Dex docs: "Client app redirects user to dex with an OAuth2 request")
	c.Redirect(http.StatusFound, authURL)
}

// OIDCCallback handles OIDC callback following Dex documentation
// @Summary OIDC Callback
// @Description Handle OIDC callback from Dex
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code from OIDC provider"
// @Param state query string true "State parameter for CSRF protection"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/oidc/callback [get]
func (h *Handler) OIDCCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code or state parameter"})
		return
	}

	// Extract invite code from state if present
	var inviteCode string
	if parts := splitState(state); len(parts) == 2 {
		state = parts[0]
		inviteCode = parts[1]
	}

	ctx := context.Background()

	// Exchange authorization code for tokens (following Dex docs: "Client exchanges code with dex for an id_token")
	oauth2Token, err := h.oidcService.ExchangeCodeForToken(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	// Validate ID token and get user info (following Dex docs pattern)
	fmt.Printf("DEBUG: Validating ID token from OIDC callback\n")
	userInfo, err := h.oidcService.ValidateIDToken(ctx, oauth2Token)
	if err != nil {
		fmt.Printf("DEBUG: Failed to validate ID token: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to validate ID token"})
		return
	}
	fmt.Printf("DEBUG: ID token validated successfully for user: %s\n", userInfo.Email)

	// Authenticate user and create/update user in database
	fmt.Printf("DEBUG: Authenticating user via OIDC\n")
	user, err := h.oidcService.AuthenticateUser(ctx, userInfo)
	if err != nil {
		fmt.Printf("DEBUG: Authentication failed: %v\n", err)
		// Check if this is an organization authorization error
		if strings.Contains(err.Error(), "not authorized") {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": fmt.Sprintf("Your account is not authorized to access this application. Please contact your administrator to ensure your group membership is properly configured."),
				"details": err.Error(),
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Authentication failed: %v", err)})
		}
		return
	}
	fmt.Printf("DEBUG: User authenticated successfully: %s\n", user.Email)

	// If invite code is provided, handle organization assignment
	if inviteCode != "" {
		invite, err := h.repo.GetInviteByID(inviteCode)
		if err == nil {
			// Update user's organization to match invite
			user.OrganizationID = &invite.OrganizationID
			user.UpdatedAt = time.Now()
			if err := h.repo.UpdateUser(user); err != nil {
				fmt.Printf("Warning: Failed to update user organization from invite: %v\n", err)
			}
		}
	}

	// Generate JWT token for the application
	jwtToken, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate application token"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, models.LoginResponse{
		Token: jwtToken,
		User:  *user,
	})
}

// OIDCLogout handles OIDC logout
// @Summary OIDC Logout
// @Description Logout from OIDC provider
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Router /auth/oidc/logout [post]
func (h *Handler) OIDCLogout(c *gin.Context) {
	// For now, just return success
	// In a full implementation, you might want to call the OIDC provider's logout endpoint
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// splitState splits state parameter to extract invite code
func splitState(state string) []string {
	// Simple split on first colon
	for i, char := range state {
		if char == ':' {
			return []string{state[:i], state[i+1:]}
		}
	}
	return []string{state}
}
