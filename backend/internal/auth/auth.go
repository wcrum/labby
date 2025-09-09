package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/wcrum/labby/internal/database"
	"github.com/wcrum/labby/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// Service handles authentication
type Service struct {
	jwtSecret []byte
	repo      *database.Repository
}

// NewService creates a new auth service
func NewService(jwtSecret string, repo *database.Repository) *Service {
	return &Service{
		jwtSecret: []byte(jwtSecret),
		repo:      repo,
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(email, name string, role models.UserRole) (*models.User, error) {
	return s.CreateUserWithOrganization(email, name, role, nil)
}

// CreateUserWithOrganization creates a new user with optional organization
func (s *Service) CreateUserWithOrganization(email, name string, role models.UserRole, organizationID *string) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.repo.GetUserByEmail(email)
	if err == nil {
		// If user exists but has no organization and we're providing one, update it
		if existingUser.OrganizationID == nil && organizationID != nil {
			existingUser.OrganizationID = organizationID
			existingUser.UpdatedAt = time.Now()
			if err := s.repo.UpdateUser(existingUser); err != nil {
				return nil, fmt.Errorf("failed to update user organization: %w", err)
			}
			fmt.Printf("DEBUG: Updated existing user %s with organization %s\n", existingUser.Email, *organizationID)
		}
		return existingUser, nil
	}

	// Create new user
	now := time.Now()
	user := &models.User{
		ID:             models.GenerateID(),
		Email:          email,
		Name:           name,
		Role:           role,
		OrganizationID: organizationID, // Assign to provided organization or nil
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if organizationID != nil {
		fmt.Printf("DEBUG: Created new user: %s (ID: %s) with organization %s\n", user.Email, user.ID, *organizationID)
	} else {
		fmt.Printf("DEBUG: Created new user: %s (ID: %s) with no organization\n", user.Email, user.ID)
	}
	return user, nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID string) (*models.User, error) {
	return s.repo.GetUserByID(userID)
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(email string) (*models.User, error) {
	return s.repo.GetUserByEmail(email)
}

// Login performs authentication
func (s *Service) Login(email string) (*models.User, error) {
	return s.LoginWithOrganization(email, nil)
}

// LoginWithOrganization performs authentication with optional organization assignment
func (s *Service) LoginWithOrganization(email string, organizationID *string) (*models.User, error) {
	// Try to find existing user
	user, err := s.GetUserByEmail(email)
	if err != nil {
		// Create new user if not found
		// Default to user role for new users
		user, err = s.CreateUserWithOrganization(email, "User", models.UserRoleUser, organizationID)
		if err != nil {
			return nil, err
		}
	} else if organizationID != nil && user.OrganizationID == nil {
		// If user exists but has no organization and we're providing one, update it
		user.OrganizationID = organizationID
		user.UpdatedAt = time.Now()
		fmt.Printf("DEBUG: Updated existing user %s with organization %s during login\n", user.Email, *organizationID)
	}

	return user, nil
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(user *models.User) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken validates a JWT token and returns the user
func (s *Service) ValidateToken(tokenString string) (*models.User, error) {
	fmt.Printf("ValidateToken: Validating token: %s...\n", tokenString[:10]+"...")

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		fmt.Printf("ValidateToken: JWT parsing failed: %v\n", err)
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		fmt.Printf("ValidateToken: Token claims - UserID: %s, Email: %s, Role: %s\n", claims.UserID, claims.Email, claims.Role)
		user, err := s.GetUserByID(claims.UserID)
		if err != nil {
			fmt.Printf("ValidateToken: User not found by ID %s: %v\n", claims.UserID, err)
			return nil, ErrInvalidToken
		}
		fmt.Printf("ValidateToken: User found: %s\n", user.Email)
		return user, nil
	}

	fmt.Printf("ValidateToken: Token claims invalid or token not valid\n")
	return nil, ErrInvalidToken
}

// GetAllUsers returns all users (for admin purposes)
func (s *Service) GetAllUsers() ([]*models.User, error) {
	return s.repo.GetAllUsers()
}

// CreateAdminUser creates an admin user
func (s *Service) CreateAdminUser(email, name string) (*models.User, error) {
	return s.CreateUser(email, name, models.UserRoleAdmin)
}

// UpdateUserRole updates a user's role
func (s *Service) UpdateUserRole(userID string, role models.UserRole) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	user.Role = role
	user.UpdatedAt = time.Now()
	return s.repo.UpdateUser(user)
}

// UpdateUserOrganization updates a user's organization
func (s *Service) UpdateUserOrganization(userID string, organizationID *string) error {
	fmt.Printf("DEBUG: UpdateUserOrganization called for userID: %s, organizationID: %v\n", userID, organizationID)

	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		fmt.Printf("ERROR: User not found with ID: %s\n", userID)
		return errors.New("user not found")
	}

	fmt.Printf("DEBUG: Found user: %s (email: %s), current org: %v\n", user.Name, user.Email, user.OrganizationID)

	user.OrganizationID = organizationID
	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user organization: %w", err)
	}

	fmt.Printf("DEBUG: Updated user organization to: %v\n", user.OrganizationID)
	return nil
}

// AssignUserToDefaultOrganization assigns a user to the default organization if they don't have one
func (s *Service) AssignUserToDefaultOrganization(userID string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Only assign to default organization if user has no organization
	if user.OrganizationID == nil {
		user.OrganizationID = stringPtr("org-default")
		user.UpdatedAt = time.Now()
		if err := s.repo.UpdateUser(user); err != nil {
			return fmt.Errorf("failed to update user organization: %w", err)
		}
		fmt.Printf("DEBUG: Assigned user %s to default organization\n", user.Email)
	}

	return nil
}

// IsAdmin checks if a user is an admin
func (s *Service) IsAdmin(user *models.User) bool {
	return user.Role == models.UserRoleAdmin
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(userID string) error {
	return s.repo.DeleteUser(userID)
}
