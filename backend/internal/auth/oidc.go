package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/wcrum/labby/internal/database"
	"github.com/wcrum/labby/internal/models"
	"golang.org/x/oauth2"
)

// OIDCService handles OIDC authentication with Dex
type OIDCService struct {
	provider    *oidc.Provider
	config      oauth2.Config
	verifier    *oidc.IDTokenVerifier
	repo        *database.Repository
	authService *Service
}

// OIDCUserInfo represents user information from OIDC
type OIDCUserInfo struct {
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}

// NewOIDCService creates a new OIDC service following Dex documentation pattern
func NewOIDCService(issuer, clientID, clientSecret, redirectURL string, repo *database.Repository, authService *Service) (*OIDCService, error) {
	ctx := context.Background()

	fmt.Printf("DEBUG: Creating OIDC provider with issuer: %s\n", issuer)
	// Initialize a provider by specifying dex's issuer URL
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		fmt.Printf("DEBUG: Failed to create OIDC provider: %v\n", err)
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}
	fmt.Printf("DEBUG: OIDC provider created successfully\n")

	// Configure the OAuth2 config with the client values following Dex docs
	config := oauth2.Config{
		// client_id and client_secret of the client
		ClientID:     clientID,
		ClientSecret: clientSecret,

		// The redirectURL
		RedirectURL: redirectURL,

		// Discovery returns the OAuth2 endpoints
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows
		// Other scopes, such as "groups" can be requested
		Scopes: []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}

	// Create an ID token parser following Dex docs
	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCService{
		provider:    provider,
		config:      config,
		verifier:    verifier,
		repo:        repo,
		authService: authService,
	}, nil
}

// GetAuthURL returns the OIDC authorization URL
func (s *OIDCService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state)
}

// ExchangeCodeForToken exchanges authorization code for tokens
func (s *OIDCService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.config.Exchange(ctx, code)
}

// ValidateIDToken validates the ID token and returns user info following Dex docs
func (s *OIDCService) ValidateIDToken(ctx context.Context, token *oauth2.Token) (*OIDCUserInfo, error) {
	// Extract the ID Token from OAuth2 token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token response")
	}

	// Parse and verify ID Token payload
	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Extract custom claims following Dex documentation pattern
	var claims struct {
		Email    string   `json:"email"`
		Name     string   `json:"name"`
		Verified bool     `json:"email_verified"`
		Groups   []string `json:"groups"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse ID token claims: %w", err)
	}

	// Verify email is verified (following Dex docs recommendation)
	if !claims.Verified {
		return nil, fmt.Errorf("email (%q) in returned claims was not verified", claims.Email)
	}

	return &OIDCUserInfo{
		Email:  claims.Email,
		Name:   claims.Name,
		Groups: claims.Groups,
	}, nil
}

// AuthenticateUser authenticates a user via OIDC and creates/updates user in database
// Users are only created if their group claim matches an existing organization
func (s *OIDCService) AuthenticateUser(ctx context.Context, userInfo *OIDCUserInfo) (*models.User, error) {
	fmt.Printf("DEBUG: Authenticating OIDC user: %s with groups: %v\n", userInfo.Email, userInfo.Groups)

	// First, validate that user has a group that matches an existing organization
	var organizationID *string
	var matchedGroup string

	// Check if user is admin (admins can be created without organization match)
	isAdmin := contains(userInfo.Groups, "admin")

	if !isAdmin {
		// For non-admin users, find organization based on groups
		for _, group := range userInfo.Groups {
			// Try to find organization by name or create mapping
			org, err := s.findOrganizationByGroup(group)
			if err == nil {
				organizationID = &org.ID
				matchedGroup = group
				fmt.Printf("DEBUG: Found organization %s for group %s\n", org.Name, group)
				break
			} else {
				fmt.Printf("DEBUG: No organization found for group %s: %v\n", group, err)
			}
		}

		// If no organization found for non-admin user, reject authentication
		if organizationID == nil {
			fmt.Printf("DEBUG: Rejecting authentication for user %s - no matching organization found for groups: %v\n", userInfo.Email, userInfo.Groups)
			return nil, fmt.Errorf("user %s is not authorized - no matching organization found for groups: %v", userInfo.Email, userInfo.Groups)
		}
	} else {
		// For admin users, try to find organization but don't require it
		for _, group := range userInfo.Groups {
			if group != "admin" {
				org, err := s.findOrganizationByGroup(group)
				if err == nil {
					organizationID = &org.ID
					matchedGroup = group
					fmt.Printf("DEBUG: Found organization %s for admin user group %s\n", org.Name, group)
					break
				}
			}
		}
	}

	// Check if user exists
	user, err := s.authService.GetUserByEmail(userInfo.Email)
	if err != nil {
		fmt.Printf("DEBUG: User %s doesn't exist, creating new user with organization: %v\n", userInfo.Email, organizationID)
		// User doesn't exist, create new user
		// Determine role based on groups
		role := models.UserRoleUser
		if isAdmin {
			role = models.UserRoleAdmin
		}

		user, err = s.authService.CreateUserWithOrganization(userInfo.Email, userInfo.Name, role, organizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		fmt.Printf("DEBUG: Created new user: %s with organization: %v (matched group: %s)\n", user.Email, user.OrganizationID, matchedGroup)
	} else {
		fmt.Printf("DEBUG: User %s exists, updating if needed\n", userInfo.Email)
		// User exists, update role and organization if needed
		updated := false

		// Update role if user is in admin group
		if isAdmin && user.Role != models.UserRoleAdmin {
			user.Role = models.UserRoleAdmin
			updated = true
		}

		// Update organization based on matched group
		if organizationID != nil && (user.OrganizationID == nil || *user.OrganizationID != *organizationID) {
			user.OrganizationID = organizationID
			updated = true
			fmt.Printf("DEBUG: Updated user organization to %s for group %s\n", *organizationID, matchedGroup)
		}

		if updated {
			user.UpdatedAt = s.authService.getCurrentTime()
			if err := s.repo.UpdateUser(user); err != nil {
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
		}
	}

	fmt.Printf("DEBUG: User %s authenticated successfully with organization: %v\n", user.Email, user.OrganizationID)
	return user, nil
}

// findOrganizationByGroup finds an organization by group name
func (s *OIDCService) findOrganizationByGroup(group string) (*models.Organization, error) {
	// Try to find organization by name (case-insensitive)
	orgs, err := s.repo.GetAllOrganizations()
	if err != nil {
		return nil, err
	}

	groupLower := strings.ToLower(group)
	for _, org := range orgs {
		if strings.ToLower(org.Name) == groupLower {
			return org, nil
		}
	}

	// If not found, try to find by domain
	for _, org := range orgs {
		if org.Domain != "" && strings.ToLower(org.Domain) == groupLower {
			return org, nil
		}
	}

	return nil, fmt.Errorf("organization not found for group: %s", group)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getCurrentTime returns current time (helper method)
func (s *Service) getCurrentTime() time.Time {
	return time.Now()
}
