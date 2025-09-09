package models

import (
	"time"
)

// Organization represents a group of users with access to specific labs
type Organization struct {
	ID          string    `json:"id" gorm:"primaryKey;size:8"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Domain      string    `json:"domain" gorm:"index"` // Optional domain for organization
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	ID             string    `json:"id" gorm:"primaryKey;size:8"`
	OrganizationID string    `json:"organization_id" gorm:"not null;size:20;index"`
	UserID         string    `json:"user_id" gorm:"not null;size:8;index"`
	Role           string    `json:"role" gorm:"not null;default:'member'"` // "owner", "admin", "member"
	JoinedAt       time.Time `json:"joined_at"`
}

// Invite represents an invitation to join an organization
type Invite struct {
	ID             string     `json:"id" gorm:"primaryKey;size:8"`
	OrganizationID string     `json:"organization_id" gorm:"not null;size:20;index"`
	Email          string     `json:"email" gorm:"not null;index"`
	InvitedBy      string     `json:"invited_by" gorm:"not null;size:8"`
	Role           string     `json:"role" gorm:"not null;default:'member'"`          // "admin", "member"
	Status         string     `json:"status" gorm:"not null;default:'pending';index"` // "pending", "accepted", "expired"
	ExpiresAt      time.Time  `json:"expires_at" gorm:"not null"`
	CreatedAt      time.Time  `json:"created_at"`
	AcceptedAt     *time.Time `json:"accepted_at"`
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
