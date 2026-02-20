package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wcrum/labby/internal/database"
	"github.com/wcrum/labby/internal/models"
)

// OrganizationService handles organization-related operations
type OrganizationService struct {
	repo *database.Repository
}

var (
	organizationServiceInstance *OrganizationService
	organizationServiceOnce     sync.Once
)

// NewOrganizationService creates a new organization service instance (singleton)
func NewOrganizationService() *OrganizationService {
	organizationServiceOnce.Do(func() {
		fmt.Printf("DEBUG: Initializing new OrganizationService (singleton)\n")

		// Get the database repository instance
		// Note: This assumes the repository is already initialized
		// In a real application, you might want to pass the repository as a parameter
		organizationServiceInstance = &OrganizationService{
			repo: nil, // Will be set when repository is available
		}

		fmt.Printf("DEBUG: Service initialized\n")
	})

	fmt.Printf("DEBUG: Returning existing OrganizationService instance\n")
	return organizationServiceInstance
}

// SetRepository sets the database repository for the organization service
func (s *OrganizationService) SetRepository(repo *database.Repository) {
	s.repo = repo
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(name, description, domain string) (*models.Organization, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	org := &models.Organization{
		ID:          fmt.Sprintf("org-%s", uuid.New().String()[:8]),
		Name:        name,
		Description: description,
		Domain:      domain,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateOrganization(org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(id string) (*models.Organization, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	org, err := s.repo.GetOrganizationByID(id)
	if err != nil {
		return nil, fmt.Errorf("organization not found")
	}
	return org, nil
}

// GetAllOrganizations returns all organizations
func (s *OrganizationService) GetAllOrganizations() ([]*models.Organization, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	orgs, err := s.repo.GetAllOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	return orgs, nil
}

// AddMember adds a user to an organization
func (s *OrganizationService) AddMember(organizationID, userID, role string) (*models.OrganizationMember, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	// Check if organization exists
	_, err := s.repo.GetOrganizationByID(organizationID)
	if err != nil {
		return nil, fmt.Errorf("organization not found")
	}

	// Check if user is already a member
	existingMembers, err := s.repo.GetOrganizationMembers(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing members: %w", err)
	}

	for _, member := range existingMembers {
		if member.OrganizationID == organizationID && member.UserID == userID {
			return nil, fmt.Errorf("user is already a member of this organization")
		}
	}

	member := &models.OrganizationMember{
		ID:             fmt.Sprintf("member-%s", uuid.New().String()[:8]),
		OrganizationID: organizationID,
		UserID:         userID,
		Role:           role,
		JoinedAt:       time.Now(),
	}

	if err := s.repo.CreateOrganizationMember(member); err != nil {
		return nil, fmt.Errorf("failed to create organization member: %w", err)
	}

	return member, nil
}

// GetOrganizationMembers returns all members of an organization
func (s *OrganizationService) GetOrganizationMembers(organizationID string) ([]models.OrganizationMember, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	members, err := s.repo.GetOrganizationMembers(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization members: %w", err)
	}
	return members, nil
}

// CreateInvite creates an invitation to join an organization
func (s *OrganizationService) CreateInvite(organizationID, email, role, invitedBy string, usageLimit *int) (*models.Invite, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	fmt.Printf("DEBUG: Creating invite for org: %s, email: %s, role: %s, usageLimit: %v\n", organizationID, email, role, usageLimit)

	// Check if organization exists
	_, err := s.repo.GetOrganizationByID(organizationID)
	if err != nil {
		return nil, fmt.Errorf("organization not found")
	}

	// Set default usage limit if not provided
	if usageLimit == nil {
		defaultLimit := 1
		usageLimit = &defaultLimit
	}

	invite := &models.Invite{
		ID:             uuid.New().String()[:8],
		OrganizationID: organizationID,
		Email:          email,
		InvitedBy:      invitedBy,
		Role:           role,
		Status:         "pending",
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:      time.Now(),
		UsageLimit:     usageLimit,
		UsageCount:     0,
		UsedBy:         []string{},
	}

	if err := s.repo.CreateInvite(invite); err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	fmt.Printf("DEBUG: Created invite with ID: %s\n", invite.ID)
	fmt.Printf("DEBUG: Invite details: %+v\n", invite)
	return invite, nil
}

// GetInvite retrieves an invite by ID
func (s *OrganizationService) GetInvite(id string) (*models.Invite, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	// Debug: log the search attempt
	fmt.Printf("DEBUG: Searching for invite with ID: %s\n", id)

	invite, err := s.repo.GetInviteByID(id)
	if err != nil {
		return nil, fmt.Errorf("invite not found: ID '%s' does not exist in the system", id)
	}

	// Check if invite has expired
	if time.Now().After(invite.ExpiresAt) {
		invite.Status = "expired"
		// Update the invite status in the database
		s.repo.UpdateInvite(invite)
	}

	fmt.Printf("DEBUG: Found invite: %+v\n", invite)
	return invite, nil
}

// AcceptInvite accepts an invitation and adds the user to the organization
func (s *OrganizationService) AcceptInvite(inviteID, userID string) error {
	if s.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	invite, err := s.GetInvite(inviteID)
	if err != nil {
		return err
	}

	if invite.Status != "pending" {
		return fmt.Errorf("invite is not pending")
	}

	if time.Now().After(invite.ExpiresAt) {
		return fmt.Errorf("invite has expired")
	}

	// Check usage limit
	if invite.UsageLimit != nil && invite.UsageCount >= *invite.UsageLimit {
		return fmt.Errorf("invite usage limit exceeded")
	}

	// Add user to organization
	_, err = s.AddMember(invite.OrganizationID, userID, invite.Role)
	if err != nil {
		return err
	}

	// Update usage tracking
	invite.UsageCount++
	now := time.Now()
	invite.LastUsedAt = &now
	invite.UsedBy = append(invite.UsedBy, userID)

	// If this was a single-use invite, mark it as accepted
	if invite.UsageLimit != nil && invite.UsageCount >= *invite.UsageLimit {
		invite.Status = "accepted"
		invite.AcceptedAt = &now
	}

	// Update the invite in the database
	if err := s.repo.UpdateInvite(invite); err != nil {
		return fmt.Errorf("failed to update invite status: %w", err)
	}

	return nil
}

// GetInvitesByEmail returns all invites for a specific email
func (s *OrganizationService) GetInvitesByEmail(email string) ([]*models.Invite, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	invites, err := s.repo.GetInvitesByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get invites by email: %w", err)
	}

	// Filter for pending invites and check expiration
	var pendingInvites []*models.Invite
	for _, invite := range invites {
		if invite.Status == "pending" {
			// Check if invite has expired
			if time.Now().After(invite.ExpiresAt) {
				invite.Status = "expired"
				// Update the invite status in the database
				s.repo.UpdateInvite(invite)
			} else {
				pendingInvites = append(pendingInvites, invite)
			}
		}
	}
	return pendingInvites, nil
}

// GetOrganizationWithMembers returns an organization with its members and invites
func (s *OrganizationService) GetOrganizationWithMembers(organizationID string) (*models.OrganizationWithMembers, error) {
	org, err := s.GetOrganization(organizationID)
	if err != nil {
		return nil, err
	}

	members, err := s.GetOrganizationMembers(organizationID)
	if err != nil {
		return nil, err
	}

	// Get invites for this organization
	invites, err := s.repo.GetInvitesByOrganization(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invites for organization: %w", err)
	}

	return &models.OrganizationWithMembers{
		Organization: org,
		Members:      members,
		Invites:      invites,
	}, nil
}

// GetAllInvites returns all invites across all organizations (admin function)
func (s *OrganizationService) GetAllInvites() ([]*models.Invite, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	invites, err := s.repo.GetAllInvites()
	if err != nil {
		return nil, fmt.Errorf("failed to get all invites: %w", err)
	}
	return invites, nil
}

// GetInviteUsageStats returns usage statistics for all invites (admin function)
func (s *OrganizationService) GetInviteUsageStats() ([]*models.InviteUsageStats, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	stats, err := s.repo.GetInviteUsageStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get invite usage stats: %w", err)
	}
	return stats, nil
}

// GetInvitesByOrganizationID returns all invites for a specific organization
func (s *OrganizationService) GetInvitesByOrganizationID(orgID string) ([]models.Invite, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}

	invites, err := s.repo.GetInvitesByOrganization(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invites for organization: %w", err)
	}
	return invites, nil
}
