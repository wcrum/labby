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
	// Usage tracking fields
	UsageLimit *int       `json:"usage_limit,omitempty" gorm:"default:1"` // Maximum number of times this invite can be used (nil = unlimited)
	UsageCount int        `json:"usage_count" gorm:"default:0"`           // Number of times this invite has been used
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`                 // When the invite was last used
	UsedBy     []string   `json:"used_by,omitempty" gorm:"type:text[]"`   // List of user IDs who have used this invite
}

// CreateInviteRequest represents a request to create an invitation
type CreateInviteRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Role       string `json:"role" binding:"required,oneof=admin member"`
	UsageLimit *int   `json:"usage_limit,omitempty"` // Optional usage limit (nil = unlimited)
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

// InviteWithDetails represents an invite with additional details for admin views
type InviteWithDetails struct {
	Invite
	OrganizationName string `json:"organization_name"`
	InvitedByName    string `json:"invited_by_name"`
	InvitedByEmail   string `json:"invited_by_email"`
}

// InviteUsageStats represents usage statistics for an invite
type InviteUsageStats struct {
	InviteID         string     `json:"invite_id"`
	Email            string     `json:"email"`
	OrganizationName string     `json:"organization_name"`
	UsageCount       int        `json:"usage_count"`
	UsageLimit       *int       `json:"usage_limit,omitempty"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	UsedBy           []UserInfo `json:"used_by,omitempty"`
}

// UserInfo represents basic user information for invite usage tracking
type UserInfo struct {
	ID     string    `json:"id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
	UsedAt time.Time `json:"used_at"`
}
