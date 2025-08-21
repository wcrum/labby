package lab

import (
	"errors"
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
	labs            map[string]*models.Lab
	mu              sync.RWMutex
	serviceManager  *services.ServiceManager
	progressTracker *ProgressTracker
	templateManager *models.LabTemplateManager
	templateLoader  *TemplateLoader
}

// NewService creates a new lab service
func NewService() *Service {
	templateManager := models.NewLabTemplateManager()
	templateLoader := NewTemplateLoader(templateManager)

	return &Service{
		labs:            make(map[string]*models.Lab),
		serviceManager:  services.NewServiceManager(),
		progressTracker: NewProgressTracker(),
		templateManager: templateManager,
		templateLoader:  templateLoader,
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
		ID:           models.GenerateID(),
		Name:         models.GenerateLabName(), // Use short lab name format
		Status:       models.LabStatusProvisioning,
		OwnerID:      ownerID,
		StartedAt:    now,
		EndsAt:       now.Add(time.Duration(durationMinutes) * time.Minute),
		CreatedAt:    now,
		UpdatedAt:    now,
		Credentials:  []models.Credential{},
		UsedServices: []string{}, // Empty for labs created without templates
	}

	s.labs[lab.ID] = lab

	// Initialize progress tracking
	s.progressTracker.InitializeProgress(lab.ID)
	s.progressTracker.AddLog(lab.ID, "Lab creation started")

	// Start lab provisioning
	go s.provisionLabFromTemplate(lab.ID, "")

	return lab, nil
}

// GetProgress returns the progress for a lab
func (s *Service) GetProgress(labID string) *LabProgress {
	return s.progressTracker.GetProgress(labID)
}

// LoadTemplates loads lab templates from a directory
func (s *Service) LoadTemplates(dirPath string) error {
	return s.templateLoader.LoadTemplatesFromDirectory(dirPath)
}

// GetTemplates returns all available lab templates
func (s *Service) GetTemplates() []*models.LabTemplate {
	return s.templateManager.GetAllTemplates()
}

// GetTemplate returns a specific lab template
func (s *Service) GetTemplate(templateID string) (*models.LabTemplate, bool) {
	return s.templateManager.GetTemplate(templateID)
}

// CreateLabFromTemplate creates a lab from a template
func (s *Service) CreateLabFromTemplate(templateID, ownerID string) (*models.Lab, error) {
	lab, err := s.templateLoader.CreateLabFromTemplate(templateID, ownerID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.labs[lab.ID] = lab
	s.mu.Unlock()

	// Initialize progress tracking
	s.progressTracker.InitializeProgress(lab.ID)
	s.progressTracker.AddLog(lab.ID, "Lab creation started from template")

	// Start lab provisioning
	go s.provisionLabFromTemplate(lab.ID, templateID)

	return lab, nil
}

// CleanupLabServices executes cleanup for a specific lab
func (s *Service) CleanupLabServices(cleanupCtx *interfaces.CleanupContext) error {
	return s.serviceManager.CleanupLabServices(cleanupCtx)
}

// ConvertLabToResponse converts a Lab to LabResponse by looking up the owner
func (s *Service) ConvertLabToResponse(lab *models.Lab, authService interface{}) *models.LabResponse {
	// Try to get the owner user
	var owner models.User
	if authSvc, ok := authService.(interface {
		GetUserByID(userID string) (*models.User, error)
	}); ok {
		if user, err := authSvc.GetUserByID(lab.OwnerID); err == nil {
			owner = *user
		} else {
			// If user lookup fails, create a fallback user
			owner = models.User{
				ID:    lab.OwnerID,
				Email: "unknown@example.com",
				Name:  "Unknown User",
			}
		}
	} else {
		// If auth service doesn't have GetUserByID, create a fallback user
		owner = models.User{
			ID:    lab.OwnerID,
			Email: "unknown@example.com",
			Name:  "Unknown User",
		}
	}

	return &models.LabResponse{
		ID:          lab.ID,
		Name:        lab.Name,
		Status:      lab.Status,
		Owner:       owner,
		StartedAt:   lab.StartedAt,
		EndsAt:      lab.EndsAt,
		Credentials: lab.Credentials,
	}
}
