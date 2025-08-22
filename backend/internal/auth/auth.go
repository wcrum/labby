package auth

import (
	"errors"
	"fmt"
	"time"

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
	users     map[string]*models.User // In-memory user store for demo
}

// NewService creates a new auth service
func NewService(jwtSecret string) *Service {
	return &Service{
		jwtSecret: []byte(jwtSecret),
		users:     make(map[string]*models.User),
	}
}

// CreateUser creates a new user (dummy implementation)
func (s *Service) CreateUser(email, name string, role models.UserRole) (*models.User, error) {
	// Check if user already exists
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}

	// Create new user
	now := time.Now()
	user := &models.User{
		ID:        models.GenerateID(),
		Email:     email,
		Name:      name,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.users[user.ID] = user
	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID string) (*models.User, error) {
	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(email string) (*models.User, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

// Login performs dummy authentication (always succeeds)
func (s *Service) Login(email string) (*models.User, error) {
	// Try to find existing user
	user, err := s.GetUserByEmail(email)
	if err != nil {
		// Create new user if not found (dummy auth)
		// Default to user role for new users
		user, err = s.CreateUser(email, "Demo User", models.UserRoleUser)
		if err != nil {
			return nil, err
		}
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
func (s *Service) GetAllUsers() []*models.User {
	users := make([]*models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

// CreateAdminUser creates an admin user
func (s *Service) CreateAdminUser(email, name string) (*models.User, error) {
	return s.CreateUser(email, name, models.UserRoleAdmin)
}

// UpdateUserRole updates a user's role
func (s *Service) UpdateUserRole(userID string, role models.UserRole) error {
	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}
	user.Role = role
	user.UpdatedAt = time.Now()
	return nil
}

// IsAdmin checks if a user is an admin
func (s *Service) IsAdmin(user *models.User) bool {
	return user.Role == models.UserRoleAdmin
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(userID string) error {
	if _, exists := s.users[userID]; !exists {
		return errors.New("user not found")
	}
	delete(s.users, userID)
	return nil
}
