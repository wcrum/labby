package database

import (
	"errors"
	"time"

	"github.com/wcrum/labby/internal/models"
	"gorm.io/gorm"
)

// Repository provides database access methods using GORM
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// AutoMigrate runs database migrations for all models
func (r *Repository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.OrganizationMember{},
		&models.Invite{},
		&models.Lab{},
		&models.Credential{},
		&models.ServiceConfig{},
		&models.ServiceLimit{},
	)
}

// UserRepository methods

// CreateUser creates a new user in the database
func (r *Repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user in the database
func (r *Repository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// GetAllUsers retrieves all users
func (r *Repository) GetAllUsers() ([]*models.User, error) {
	var users []*models.User
	err := r.db.Order("created_at DESC").Find(&users).Error
	return users, err
}

// DeleteUser deletes a user by ID
func (r *Repository) DeleteUser(id string) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// LabRepository methods

// CreateLab creates a new lab in the database
func (r *Repository) CreateLab(lab *models.Lab) error {
	return r.db.Create(lab).Error
}

// GetLabByID retrieves a lab by ID with credentials
func (r *Repository) GetLabByID(id string) (*models.Lab, error) {
	var lab models.Lab
	err := r.db.Preload("Credentials").First(&lab, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &lab, nil
}

// GetLabsByOwnerID retrieves all labs for a specific owner
func (r *Repository) GetLabsByOwnerID(ownerID string) ([]*models.Lab, error) {
	var labs []*models.Lab
	err := r.db.Preload("Credentials").Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&labs).Error
	return labs, err
}

// GetAllLabs retrieves all labs
func (r *Repository) GetAllLabs() ([]*models.Lab, error) {
	var labs []*models.Lab
	err := r.db.Preload("Credentials").Order("created_at DESC").Find(&labs).Error
	return labs, err
}

// UpdateLab updates a lab in the database
func (r *Repository) UpdateLab(lab *models.Lab) error {
	return r.db.Save(lab).Error
}

// DeleteLab deletes a lab by ID
func (r *Repository) DeleteLab(id string) error {
	return r.db.Delete(&models.Lab{}, "id = ?", id).Error
}

// GetExpiredLabs retrieves labs that have expired
func (r *Repository) GetExpiredLabs() ([]*models.Lab, error) {
	var labs []*models.Lab
	err := r.db.Where("ends_at < ? AND status != ?", time.Now(), models.LabStatusExpired).Order("ends_at ASC").Find(&labs).Error
	return labs, err
}

// CredentialRepository methods

// CreateCredential creates a new credential in the database
func (r *Repository) CreateCredential(credential *models.Credential) error {
	return r.db.Create(credential).Error
}

// GetCredentialsByLabID retrieves all credentials for a specific lab
func (r *Repository) GetCredentialsByLabID(labID string) ([]models.Credential, error) {
	var credentials []models.Credential
	err := r.db.Where("lab_id = ?", labID).Order("created_at ASC").Find(&credentials).Error
	return credentials, err
}

// DeleteCredential deletes a credential by ID
func (r *Repository) DeleteCredential(id string) error {
	return r.db.Delete(&models.Credential{}, "id = ?", id).Error
}

// DeleteCredentialsByLabID deletes all credentials for a specific lab
func (r *Repository) DeleteCredentialsByLabID(labID string) error {
	return r.db.Delete(&models.Credential{}, "lab_id = ?", labID).Error
}

// OrganizationRepository methods

// CreateOrganization creates a new organization in the database
func (r *Repository) CreateOrganization(org *models.Organization) error {
	return r.db.Create(org).Error
}

// GetOrganizationByID retrieves an organization by ID
func (r *Repository) GetOrganizationByID(id string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.First(&org, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// GetAllOrganizations retrieves all organizations
func (r *Repository) GetAllOrganizations() ([]*models.Organization, error) {
	var orgs []*models.Organization
	err := r.db.Order("created_at DESC").Find(&orgs).Error
	return orgs, err
}

// UpdateOrganization updates an organization in the database
func (r *Repository) UpdateOrganization(org *models.Organization) error {
	return r.db.Save(org).Error
}

// DeleteOrganization deletes an organization by ID
func (r *Repository) DeleteOrganization(id string) error {
	return r.db.Delete(&models.Organization{}, "id = ?", id).Error
}

// InviteRepository methods

// CreateInvite creates a new invite in the database
func (r *Repository) CreateInvite(invite *models.Invite) error {
	return r.db.Create(invite).Error
}

// GetInviteByID retrieves an invite by ID
func (r *Repository) GetInviteByID(id string) (*models.Invite, error) {
	var invite models.Invite
	err := r.db.First(&invite, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

// GetInvitesByOrganizationID retrieves all invites for a specific organization
func (r *Repository) GetInvitesByOrganizationID(orgID string) ([]*models.Invite, error) {
	var invites []*models.Invite
	err := r.db.Where("organization_id = ?", orgID).Order("created_at DESC").Find(&invites).Error
	return invites, err
}

// GetInvitesByOrganization retrieves all invites for a specific organization (alias for GetInvitesByOrganizationID)
func (r *Repository) GetInvitesByOrganization(orgID string) ([]models.Invite, error) {
	var invites []models.Invite
	err := r.db.Where("organization_id = ?", orgID).Order("created_at DESC").Find(&invites).Error
	return invites, err
}

// GetInvitesByEmail retrieves all invites for a specific email
func (r *Repository) GetInvitesByEmail(email string) ([]*models.Invite, error) {
	var invites []*models.Invite
	err := r.db.Where("email = ?", email).Order("created_at DESC").Find(&invites).Error
	return invites, err
}

// GetAllInvites retrieves all invites across all organizations (admin function)
func (r *Repository) GetAllInvites() ([]*models.Invite, error) {
	var invites []*models.Invite
	err := r.db.Order("created_at DESC").Find(&invites).Error
	return invites, err
}

// GetInviteUsageStats retrieves invite usage statistics with user details
func (r *Repository) GetInviteUsageStats() ([]*models.InviteUsageStats, error) {
	var stats []*models.InviteUsageStats

	// This is a complex query that joins invites with organizations and users
	// For now, let's implement a simpler version and enhance it later
	rows, err := r.db.Table("invites").
		Select(`
			invites.id as invite_id,
			invites.email,
			invites.usage_count,
			invites.usage_limit,
			invites.last_used_at,
			invites.status,
			invites.created_at,
			invites.expires_at,
			organizations.name as organization_name
		`).
		Joins("LEFT JOIN organizations ON invites.organization_id = organizations.id").
		Order("invites.created_at DESC").
		Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stat models.InviteUsageStats
		var orgName string

		err := rows.Scan(
			&stat.InviteID,
			&stat.Email,
			&stat.UsageCount,
			&stat.UsageLimit,
			&stat.LastUsedAt,
			&stat.Status,
			&stat.CreatedAt,
			&stat.ExpiresAt,
			&orgName,
		)
		if err != nil {
			return nil, err
		}

		stat.OrganizationName = orgName
		stats = append(stats, &stat)
	}

	return stats, nil
}

// CreateOrganizationMember creates a new organization member in the database
func (r *Repository) CreateOrganizationMember(member *models.OrganizationMember) error {
	return r.db.Create(member).Error
}

// GetOrganizationMembers retrieves all members of a specific organization
func (r *Repository) GetOrganizationMembers(orgID string) ([]models.OrganizationMember, error) {
	var members []models.OrganizationMember
	err := r.db.Where("organization_id = ?", orgID).Order("joined_at ASC").Find(&members).Error
	return members, err
}

// UpdateInvite updates an invite in the database
func (r *Repository) UpdateInvite(invite *models.Invite) error {
	return r.db.Save(invite).Error
}

// DeleteInvite deletes an invite by ID
func (r *Repository) DeleteInvite(id string) error {
	return r.db.Delete(&models.Invite{}, "id = ?", id).Error
}

// ServiceConfigRepository methods

// CreateServiceConfig creates a new service config in the database
func (r *Repository) CreateServiceConfig(config *models.ServiceConfig) error {
	return r.db.Create(config).Error
}

// GetServiceConfigByID retrieves a service config by ID
func (r *Repository) GetServiceConfigByID(id string) (*models.ServiceConfig, error) {
	var config models.ServiceConfig
	err := r.db.First(&config, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetAllServiceConfigs retrieves all service configs
func (r *Repository) GetAllServiceConfigs() ([]*models.ServiceConfig, error) {
	var configs []*models.ServiceConfig
	err := r.db.Order("created_at DESC").Find(&configs).Error
	return configs, err
}

// UpdateServiceConfig updates a service config in the database
func (r *Repository) UpdateServiceConfig(config *models.ServiceConfig) error {
	return r.db.Save(config).Error
}

// DeleteServiceConfig deletes a service config by ID
func (r *Repository) DeleteServiceConfig(id string) error {
	return r.db.Delete(&models.ServiceConfig{}, "id = ?", id).Error
}

// UpsertServiceConfig creates or updates a service config in the database
func (r *Repository) UpsertServiceConfig(config *models.ServiceConfig) error {
	// Try to find existing config
	var existing models.ServiceConfig
	err := r.db.First(&existing, "id = ?", config.ID).Error

	if err != nil {
		// If not found, create new one
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.Create(config).Error
		}
		// If other error, return it
		return err
	}

	// If found, update it
	config.CreatedAt = existing.CreatedAt // Preserve original creation time
	return r.db.Save(config).Error
}

// ServiceLimitRepository methods

// CreateServiceLimit creates a new service limit in the database
func (r *Repository) CreateServiceLimit(limit *models.ServiceLimit) error {
	return r.db.Create(limit).Error
}

// GetServiceLimitByID retrieves a service limit by ID
func (r *Repository) GetServiceLimitByID(id string) (*models.ServiceLimit, error) {
	var limit models.ServiceLimit
	err := r.db.First(&limit, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// GetServiceLimitByServiceID retrieves a service limit by service ID
func (r *Repository) GetServiceLimitByServiceID(serviceID string) (*models.ServiceLimit, error) {
	var limit models.ServiceLimit
	err := r.db.First(&limit, "service_id = ?", serviceID).Error
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// GetAllServiceLimits retrieves all service limits
func (r *Repository) GetAllServiceLimits() ([]*models.ServiceLimit, error) {
	var limits []*models.ServiceLimit
	err := r.db.Order("created_at DESC").Find(&limits).Error
	return limits, err
}

// UpdateServiceLimit updates a service limit in the database
func (r *Repository) UpdateServiceLimit(limit *models.ServiceLimit) error {
	return r.db.Save(limit).Error
}

// DeleteServiceLimit deletes a service limit by ID
func (r *Repository) DeleteServiceLimit(id string) error {
	return r.db.Delete(&models.ServiceLimit{}, "id = ?", id).Error
}

// UpsertServiceLimit creates or updates a service limit in the database
func (r *Repository) UpsertServiceLimit(limit *models.ServiceLimit) error {
	// Try to find existing limit
	var existing models.ServiceLimit
	err := r.db.First(&existing, "id = ?", limit.ID).Error

	if err != nil {
		// If not found, create new one
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.Create(limit).Error
		}
		// If other error, return it
		return err
	}

	// If found, update it
	limit.CreatedAt = existing.CreatedAt // Preserve original creation time
	return r.db.Save(limit).Error
}
