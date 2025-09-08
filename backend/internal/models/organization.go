package models

import (
	"time"
)

// Organization represents a group of users with access to specific labs
type Organization struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Domain      string    `json:"domain" db:"domain"` // Optional domain for organization
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Role           string    `json:"role" db:"role"` // "owner", "admin", "member"
	JoinedAt       time.Time `json:"joined_at" db:"joined_at"`
}

// Invite represents an invitation to join an organization
type Invite struct {
	ID             string     `json:"id" db:"id"`
	OrganizationID string     `json:"organization_id" db:"organization_id"`
	Email          string     `json:"email" db:"email"`
	InvitedBy      string     `json:"invited_by" db:"invited_by"`
	Role           string     `json:"role" db:"role"`     // "admin", "member"
	Status         string     `json:"status" db:"status"` // "pending", "accepted", "expired"
	ExpiresAt      time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	AcceptedAt     *time.Time `json:"accepted_at" db:"accepted_at"`
}

// CreateInviteRequest represents a request to create an invitation
type CreateInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=admin member"`
}

// AcceptInviteRequest represents a request to accept an invitation
type AcceptInviteRequest struct {
	InviteID string `json:"invite_id" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
}

// OrganizationWithMembers represents an organization with its member list
type OrganizationWithMembers struct {
	Organization *Organization        `json:"organization"`
	Members      []OrganizationMember `json:"members"`
	Invites      []Invite             `json:"invites"`
}
