package lab

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"spectro-lab-backend/internal/interfaces"
	"spectro-lab-backend/internal/models"
	"spectro-lab-backend/internal/services"
)

var (
	ErrLabNotFound     = errors.New("lab not found")
	ErrLabExpired      = errors.New("lab expired")
	ErrLabNotReady     = errors.New("lab not ready")
	ErrInvalidDuration = errors.New("invalid duration")
)

// Service handles lab lifecycle management
type Service struct {
	labs           map[string]*models.Lab
	mu             sync.RWMutex
	serviceManager *services.ServiceManager
}

// NewService creates a new lab service
func NewService() *Service {
	return &Service{
		labs:           make(map[string]*models.Lab),
		serviceManager: services.NewServiceManager(),
	}
}

// CreateLab creates a new lab session
func (s *Service) CreateLab(name, ownerID string, durationMinutes int) (*models.Lab, error) {
	if durationMinutes < 15 || durationMinutes > 480 {
		return nil, ErrInvalidDuration
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	lab := &models.Lab{
		ID:          models.GenerateID(),
		Name:        name,
		Status:      models.LabStatusProvisioning,
		OwnerID:     ownerID,
		StartedAt:   now,
		EndsAt:      now.Add(time.Duration(durationMinutes) * time.Minute),
		CreatedAt:   now,
		UpdatedAt:   now,
		Credentials: []models.Credential{},
	}

	s.labs[lab.ID] = lab

	// Simulate lab provisioning (in real implementation, this would be async)
	go s.provisionLab(lab.ID)

	return lab, nil
}

// provisionLab simulates lab provisioning
func (s *Service) provisionLab(labID string) {
	// Simulate provisioning time
	time.Sleep(2 * time.Second)

	s.mu.Lock()
	lab, exists := s.labs[labID]
	if !exists {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	// Set up lab services
	setupCtx := &interfaces.SetupContext{
		LabID:    labID,
		LabName:  lab.Name,
		Duration: int(lab.EndsAt.Sub(lab.StartedAt).Minutes()),
		OwnerID:  lab.OwnerID,
		Context:  context.Background(),
		Lab:      lab, // Pass lab reference for persistent data storage
		AddCredential: func(credential *interfaces.Credential) error {
			s.mu.Lock()
			defer s.mu.Unlock()

			// Convert interface credential to model credential
			modelCred := models.Credential{
				ID:        credential.ID,
				LabID:     credential.LabID,
				Label:     credential.Label,
				Username:  credential.Username,
				Password:  credential.Password,
				URL:       credential.URL,
				ExpiresAt: credential.ExpiresAt,
				Notes:     credential.Notes,
				CreatedAt: credential.CreatedAt,
				UpdatedAt: credential.UpdatedAt,
			}

			// Add to lab credentials
			if lab, exists := s.labs[labID]; exists {
				lab.Credentials = append(lab.Credentials, modelCred)
			}

			return nil
		},
	}

	// Execute service setup
	if err := s.serviceManager.SetupLabServices(setupCtx); err != nil {
		fmt.Printf("Warning: Failed to setup services for lab %s: %v\n", labID, err)
		// Continue with mock credentials as fallback
		s.mu.Lock()
		lab.Credentials = s.generateCredentials(labID, lab.EndsAt)
		s.mu.Unlock()
	}

	s.mu.Lock()
	// Update lab status to ready
	lab.Status = models.LabStatusReady
	lab.UpdatedAt = time.Now()
	s.mu.Unlock()
}

// generateCredentials generates fallback credentials for lab services
func (s *Service) generateCredentials(labID string, expiresAt time.Time) []models.Credential {
	// Only generate fallback credentials for Palette Project service
	// This is used when the real Palette Project service setup fails
	services := []struct {
		label    string
		username string
		url      string
		notes    string
	}{
		{
			label:    "Palette Project",
			username: "lab-user",
			url:      "https://training.spectrocloud.com/login",
			notes:    "Spectro Cloud control plane access (fallback credentials).",
		},
	}

	credentials := make([]models.Credential, 0, len(services))
	now := time.Now()

	for _, service := range services {
		cred := models.Credential{
			ID:        models.GenerateID(),
			LabID:     labID,
			Label:     service.label,
			Username:  service.username,
			Password:  s.generatePassword(),
			URL:       service.url,
			ExpiresAt: expiresAt,
			Notes:     service.notes,
			CreatedAt: now,
			UpdatedAt: now,
		}
		credentials = append(credentials, cred)
	}

	return credentials
}

// generatePassword generates a random password
func (s *Service) generatePassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, 12)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

// GetLab retrieves a lab by ID
func (s *Service) GetLab(labID string) (*models.Lab, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lab, exists := s.labs[labID]
	if !exists {
		return nil, ErrLabNotFound
	}

	// Check if lab is expired
	if models.IsExpired(lab.EndsAt) {
		lab.Status = models.LabStatusExpired
	}

	return lab, nil
}

// GetLabsByOwner retrieves all labs for a specific owner
func (s *Service) GetLabsByOwner(ownerID string) ([]*models.Lab, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var labs []*models.Lab
	for _, lab := range s.labs {
		if lab.OwnerID == ownerID {
			// Check if lab is expired
			if models.IsExpired(lab.EndsAt) {
				lab.Status = models.LabStatusExpired
			}
			labs = append(labs, lab)
		}
	}

	return labs, nil
}

// GetAllLabs retrieves all labs (for admin purposes)
func (s *Service) GetAllLabs() []*models.Lab {
	s.mu.RLock()
	defer s.mu.RUnlock()

	labs := make([]*models.Lab, 0, len(s.labs))
	for _, lab := range s.labs {
		// Check if lab is expired
		if models.IsExpired(lab.EndsAt) {
			lab.Status = models.LabStatusExpired
		}
		labs = append(labs, lab)
	}

	return labs
}

// DeleteLab deletes a lab
func (s *Service) DeleteLab(labID string) error {
	s.mu.Lock()
	lab, exists := s.labs[labID]
	if !exists {
		s.mu.Unlock()
		return ErrLabNotFound
	}
	s.mu.Unlock()

	// Cleanup services before deleting the lab
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   labID,
		Context: context.Background(),
		Lab:     lab, // Pass lab reference for accessing stored service data
	}

	// Store lab data in context for cleanup
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_name", lab.Name)
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_duration", int(lab.EndsAt.Sub(lab.StartedAt).Minutes()))
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_owner_id", lab.OwnerID)

	if err := s.serviceManager.CleanupLabServices(cleanupCtx); err != nil {
		// Log error but continue with deletion
		fmt.Printf("Warning: Failed to cleanup services for lab %s: %v\n", labID, err)
	}

	s.mu.Lock()
	delete(s.labs, labID)
	s.mu.Unlock()

	return nil
}

// StopLab stops a lab (sets status to expired)
func (s *Service) StopLab(labID string) error {
	s.mu.Lock()
	lab, exists := s.labs[labID]
	if !exists {
		s.mu.Unlock()
		return ErrLabNotFound
	}
	s.mu.Unlock()

	// Cleanup services before stopping the lab
	cleanupCtx := &interfaces.CleanupContext{
		LabID:   labID,
		Context: context.Background(),
		Lab:     lab, // Pass lab reference for accessing stored service data
	}

	// Store lab data in context for cleanup
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_name", lab.Name)
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_duration", int(lab.EndsAt.Sub(lab.StartedAt).Minutes()))
	cleanupCtx.Context = context.WithValue(cleanupCtx.Context, "lab_owner_id", lab.OwnerID)

	if err := s.serviceManager.CleanupLabServices(cleanupCtx); err != nil {
		// Log error but continue with stopping
		fmt.Printf("Warning: Failed to cleanup services for lab %s: %v\n", labID, err)
	}

	s.mu.Lock()
	// Set lab status to expired and update the end time
	lab.Status = models.LabStatusExpired
	lab.EndsAt = time.Now()
	lab.UpdatedAt = time.Now()
	s.mu.Unlock()

	return nil
}

// CleanupExpiredLabs removes expired labs (should be called periodically)
func (s *Service) CleanupExpiredLabs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for labID, lab := range s.labs {
		if now.After(lab.EndsAt) {
			delete(s.labs, labID)
		}
	}
}

// StartCleanupScheduler starts a background task to clean up expired labs
func (s *Service) StartCleanupScheduler() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			s.CleanupExpiredLabs()
		}
	}()
}

// CleanupLabServices executes cleanup for a specific lab
func (s *Service) CleanupLabServices(cleanupCtx *interfaces.CleanupContext) error {
	return s.serviceManager.CleanupLabServices(cleanupCtx)
}
