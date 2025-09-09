package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

// User represents a lab user
type User struct {
	ID             string    `json:"id" gorm:"primaryKey;size:8"`
	Email          string    `json:"email" gorm:"uniqueIndex;not null"`
	Name           string    `json:"name" gorm:"not null"`
	Role           UserRole  `json:"role" gorm:"not null;default:'user'"`
	OrganizationID *string   `json:"organization_id,omitempty" gorm:"size:20"` // Optional organization membership
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LabStatus represents the status of a lab
type LabStatus string

const (
	LabStatusProvisioning LabStatus = "provisioning"
	LabStatusReady        LabStatus = "ready"
	LabStatusError        LabStatus = "error"
	LabStatusExpired      LabStatus = "expired"
)

// Lab represents a lab session
type Lab struct {
	ID           string            `json:"id" gorm:"primaryKey;size:8"`
	Name         string            `json:"name" gorm:"not null"`
	Status       LabStatus         `json:"status" gorm:"not null;default:'provisioning'"`
	OwnerID      string            `json:"owner_id" gorm:"not null;size:8"`
	StartedAt    time.Time         `json:"started_at" gorm:"not null"`
	EndsAt       time.Time         `json:"ends_at" gorm:"not null"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Credentials  []Credential      `json:"credentials" gorm:"foreignKey:LabID;constraint:OnDelete:CASCADE"`
	ServiceData  map[string]string `json:"service_data,omitempty" gorm:"type:jsonb"`   // Store service-specific data for cleanup
	TemplateID   string            `json:"template_id,omitempty" gorm:"size:255"`      // Reference to the template used
	UsedServices []string          `json:"used_services,omitempty" gorm:"type:text[]"` // Track which services were used for this lab
}

// Credential represents access credentials for a lab service
type Credential struct {
	ID        string    `json:"id" gorm:"primaryKey;size:8"`
	LabID     string    `json:"lab_id" gorm:"not null;size:8;index"`
	Label     string    `json:"label" gorm:"not null"`
	Username  string    `json:"username" gorm:"not null"`
	Password  string    `json:"password" gorm:"not null"`
	URL       string    `json:"url,omitempty"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null;index"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateLabRequest represents a request to create a new lab
type CreateLabRequest struct {
	Name     string `json:"name"` // Optional, will generate UUID if not provided
	OwnerID  string `json:"owner_id" binding:"required"`
	Duration int    `json:"duration" binding:"required,min=15,max=480"` // Duration in minutes
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Email string   `json:"email" binding:"required,email"`
	Name  string   `json:"name" binding:"required"`
	Role  UserRole `json:"role" binding:"required,oneof=user admin"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email      string  `json:"email" binding:"required,email"`
	InviteCode *string `json:"invite_code,omitempty"` // Optional invite code for organization assignment
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// LabResponse represents a lab response with owner information
type LabResponse struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Status       LabStatus          `json:"status"`
	Owner        User               `json:"owner"`
	StartedAt    time.Time          `json:"started_at"`
	EndsAt       time.Time          `json:"ends_at"`
	Credentials  []Credential       `json:"credentials"`
	UsedServices []ServiceReference `json:"used_services,omitempty"` // Track which services were used for this lab
}

// GenerateID generates a new short ID (8 characters)
func GenerateID() string {
	uuid := uuid.New()
	// Take first 8 characters of UUID for shorter, readable IDs
	return uuid.String()[:8]
}

// GenerateLabName generates a lab name with short ID format
func GenerateLabName() string {
	shortID := GenerateID()
	return fmt.Sprintf("lab-%s", shortID)
}

// IsExpired checks if a lab or credential is expired
func IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// GetRemainingTime returns the remaining time until expiration
func GetRemainingTime(expiresAt time.Time) time.Duration {
	if IsExpired(expiresAt) {
		return 0
	}
	return expiresAt.Sub(time.Now())
}

// ServiceLimit represents a limit for a specific service
type ServiceLimit struct {
	ID          string    `json:"id" yaml:"id" gorm:"primaryKey"`
	ServiceID   string    `json:"service_id" yaml:"service_id" gorm:"not null;size:20;index"`   // Reference to ServiceConfig
	MaxLabs     int       `json:"max_labs" yaml:"max_labs" gorm:"not null;default:10"`          // Maximum number of concurrent labs using this service
	MaxDuration int       `json:"max_duration" yaml:"max_duration" gorm:"not null;default:480"` // Maximum duration in minutes
	IsActive    bool      `json:"is_active" yaml:"is_active" gorm:"not null;default:true"`      // Whether this limit is currently active
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
}

// ServiceConfig represents a preconfigured service configuration
type ServiceConfig struct {
	ID          string            `json:"id" yaml:"id" gorm:"primaryKey"`
	Name        string            `json:"name" yaml:"name" gorm:"not null"`
	Type        string            `json:"type" yaml:"type" gorm:"not null;index"` // palette_project, palette_tenant, proxmox_user
	Description string            `json:"description" yaml:"description"`
	Logo        string            `json:"logo" yaml:"logo"`                                              // Path to logo file (SVG/PNG)
	Config      map[string]string `json:"config" yaml:"config" gorm:"type:jsonb;not null;default:'{}'"`  // Service-specific configuration
	IsActive    bool              `json:"is_active" yaml:"is_active" gorm:"not null;default:true;index"` // Whether this service config is available
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}

// ServiceUsage represents current usage of a service
type ServiceUsage struct {
	ServiceID  string `json:"service_id"`
	ActiveLabs int    `json:"active_labs"`
	Limit      int    `json:"limit"`
}

// UserWithOrganization represents a user with organization information
type UserWithOrganization struct {
	ID           string        `json:"id"`
	Email        string        `json:"email"`
	Name         string        `json:"name"`
	Role         UserRole      `json:"role"`
	Organization *Organization `json:"organization,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
