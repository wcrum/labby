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
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Status      LabStatus         `json:"status"`
	OwnerID     string            `json:"owner_id"`
	StartedAt   time.Time         `json:"started_at"`
	EndsAt      time.Time         `json:"ends_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Credentials []Credential      `json:"credentials"`
	ServiceData map[string]string `json:"service_data,omitempty"` // Store service-specific data for cleanup
	TemplateID  string            `json:"template_id,omitempty"`  // Reference to the template used
}

// Credential represents access credentials for a lab service
type Credential struct {
	ID        string    `json:"id"`
	LabID     string    `json:"lab_id"`
	Label     string    `json:"label"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	URL       string    `json:"url,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
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
	Email string `json:"email" binding:"required,email"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// LabResponse represents a lab response with owner information
type LabResponse struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Status      LabStatus    `json:"status"`
	Owner       User         `json:"owner"`
	StartedAt   time.Time    `json:"started_at"`
	EndsAt      time.Time    `json:"ends_at"`
	Credentials []Credential `json:"credentials"`
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
