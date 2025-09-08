package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wcrum/labby/internal/models"
)

// OrganizationService handles organization-related operations
type OrganizationService struct {
	organizations map[string]*models.Organization
	members       map[string]*models.OrganizationMember
	invites       map[string]*models.Invite
}

var (
	organizationServiceInstance *OrganizationService
	organizationServiceOnce     sync.Once
)

// NewOrganizationService creates a new organization service instance (singleton)
func NewOrganizationService() *OrganizationService {
	organizationServiceOnce.Do(func() {
		fmt.Printf("DEBUG: Initializing new OrganizationService (singleton)\n")

		organizationServiceInstance = &OrganizationService{
			organizations: make(map[string]*models.Organization),
			members:       make(map[string]*models.OrganizationMember),
			invites:       make(map[string]*models.Invite),
		}

		// Create a default organization for demo purposes
		defaultOrg := &models.Organization{
			ID:          "org-default",
			Name:        "SpectroCloud",
			Description: "Default organization for SpectroCloud labs",
			Domain:      "spectrocloud.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		organizationServiceInstance.organizations[defaultOrg.ID] = defaultOrg

		fmt.Printf("DEBUG: Service initialized with %d organizations, %d members, %d invites\n",
			len(organizationServiceInstance.organizations), len(organizationServiceInstance.members), len(organizationServiceInstance.invites))
	})

	fmt.Printf("DEBUG: Returning existing OrganizationService instance\n")
	return organizationServiceInstance
}

// CreateOrganization creates a new organization
func (s *OrganizationService) CreateOrganization(name, description, domain string) (*models.Organization, error) {
	org := &models.Organization{
		ID:          fmt.Sprintf("org-%s", uuid.New().String()[:8]),
		Name:        name,
		Description: description,
		Domain:      domain,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	s.organizations[org.ID] = org
	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *OrganizationService) GetOrganization(id string) (*models.Organization, error) {
	org, exists := s.organizations[id]
	if !exists {
		return nil, fmt.Errorf("organization not found")
	}
	return org, nil
}

// GetAllOrganizations returns all organizations
func (s *OrganizationService) GetAllOrganizations() []*models.Organization {
	orgs := make([]*models.Organization, 0, len(s.organizations))
	for _, org := range s.organizations {
		orgs = append(orgs, org)
	}
	return orgs
}

// AddMember adds a user to an organization
func (s *OrganizationService) AddMember(organizationID, userID, role string) (*models.OrganizationMember, error) {
	// Check if organization exists
	if _, exists := s.organizations[organizationID]; !exists {
		return nil, fmt.Errorf("organization not found")
	}

	// Check if user is already a member
	for _, member := range s.members {
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

	s.members[member.ID] = member
	return member, nil
}

// GetOrganizationMembers returns all members of an organization
func (s *OrganizationService) GetOrganizationMembers(organizationID string) []models.OrganizationMember {
	var members []models.OrganizationMember
	for _, member := range s.members {
		if member.OrganizationID == organizationID {
			members = append(members, *member)
		}
	}
	return members
}

// CreateInvite creates an invitation to join an organization
func (s *OrganizationService) CreateInvite(organizationID, email, role, invitedBy string) (*models.Invite, error) {
	fmt.Printf("DEBUG: Creating invite for org: %s, email: %s, role: %s\n", organizationID, email, role)

	// Check if organization exists
	if _, exists := s.organizations[organizationID]; !exists {
		return nil, fmt.Errorf("organization not found")
	}

	// Check if user is already a member
	for _, member := range s.members {
		if member.OrganizationID == organizationID {
			// This is a simplified check - in a real app you'd check by email
			// For now, we'll just create the invite
		}
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
	}

	s.invites[invite.ID] = invite
	fmt.Printf("DEBUG: Created invite with ID: %s, total invites now: %d\n", invite.ID, len(s.invites))
	fmt.Printf("DEBUG: Invite details: %+v\n", invite)
	return invite, nil
}

// GetInvite retrieves an invite by ID
func (s *OrganizationService) GetInvite(id string) (*models.Invite, error) {
	// Debug: log the search attempt
	fmt.Printf("DEBUG: Searching for invite with ID: %s\n", id)
	fmt.Printf("DEBUG: Total invites in system: %d\n", len(s.invites))

	// List all available invite IDs for debugging
	if len(s.invites) > 0 {
		fmt.Printf("DEBUG: Available invite IDs: ")
		for inviteID := range s.invites {
			fmt.Printf("%s ", inviteID)
		}
		fmt.Printf("\n")
	} else {
		fmt.Printf("DEBUG: No invites exist in the system\n")
	}

	invite, exists := s.invites[id]
	if !exists {
		return nil, fmt.Errorf("invite not found: ID '%s' does not exist in the system (total invites: %d)", id, len(s.invites))
	}

	// Check if invite has expired
	if time.Now().After(invite.ExpiresAt) {
		invite.Status = "expired"
	}

	fmt.Printf("DEBUG: Found invite: %+v\n", invite)
	return invite, nil
}

// AcceptInvite accepts an invitation and adds the user to the organization
func (s *OrganizationService) AcceptInvite(inviteID, userID string) error {
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

	// Add user to organization
	_, err = s.AddMember(invite.OrganizationID, userID, invite.Role)
	if err != nil {
		return err
	}

	// Mark invite as accepted
	invite.Status = "accepted"
	now := time.Now()
	invite.AcceptedAt = &now

	return nil
}

// GetInvitesByEmail returns all invites for a specific email
func (s *OrganizationService) GetInvitesByEmail(email string) []*models.Invite {
	var invites []*models.Invite
	for _, invite := range s.invites {
		if invite.Email == email && invite.Status == "pending" {
			// Check if invite has expired
			if time.Now().After(invite.ExpiresAt) {
				invite.Status = "expired"
			} else {
				invites = append(invites, invite)
			}
		}
	}
	return invites
}

// GetOrganizationWithMembers returns an organization with its members and invites
func (s *OrganizationService) GetOrganizationWithMembers(organizationID string) (*models.OrganizationWithMembers, error) {
	org, err := s.GetOrganization(organizationID)
	if err != nil {
		return nil, err
	}

	members := s.GetOrganizationMembers(organizationID)

	var invites []models.Invite
	for _, invite := range s.invites {
		if invite.OrganizationID == organizationID {
			invites = append(invites, *invite)
		}
	}

	return &models.OrganizationWithMembers{
		Organization: org,
		Members:      members,
		Invites:      invites,
	}, nil
}
