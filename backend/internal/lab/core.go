package lab

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wcrum/labby/internal/interfaces"
	"github.com/wcrum/labby/internal/models"
	"github.com/wcrum/labby/internal/services"
)

var (
	ErrLabNotFound     = errors.New("lab not found")
	ErrLabExpired      = errors.New("lab expired")
	ErrLabNotReady     = errors.New("lab not ready")
	ErrInvalidDuration = errors.New("invalid duration")
)

// Service handles lab lifecycle management
type Service struct {
	labs                 map[string]*models.Lab
	mu                   sync.RWMutex
	serviceManager       *services.ServiceManager
	progressTracker      *ProgressTracker
	templateManager      *models.LabTemplateManager
	templateLoader       *TemplateLoader
	serviceConfigManager *models.ServiceConfigManager
}

// NewService creates a new lab service
func NewService() *Service {
	templateManager := models.NewLabTemplateManager()
	templateLoader := NewTemplateLoader(templateManager)
	serviceConfigManager := models.NewServiceConfigManager()

	return &Service{
		labs:                 make(map[string]*models.Lab),
		serviceManager:       services.NewServiceManager(),
		progressTracker:      NewProgressTracker(),
		templateManager:      templateManager,
		templateLoader:       templateLoader,
		serviceConfigManager: serviceConfigManager,
	}
}

// CreateLab creates a new lab session
func (s *Service) CreateLab(name, ownerID string, durationMinutes int) (*models.Lab, error) {
	if durationMinutes < MinLabDurationMinutes || durationMinutes > MaxLabDurationMinutes {
		return nil, ErrInvalidDuration
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	labID := models.GenerateID()
	lab := &models.Lab{
		ID:           labID,
		Name:         fmt.Sprintf("lab-%s", labID), // Use consistent lab name format
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
	fmt.Printf("Service.LoadTemplates: Loading from %s\n", dirPath)
	templateLoader := NewTemplateLoader(s.templateManager)
	err := templateLoader.LoadTemplatesFromDirectory(dirPath)
	if err != nil {
		fmt.Printf("Service.LoadTemplates: Failed to load: %v\n", err)
		return err
	}

	// Enrich templates with service type information
	s.templateManager.EnrichTemplatesWithServiceTypes(s.serviceConfigManager)

	// Log what was loaded
	templates := s.templateManager.GetAllTemplates()
	fmt.Printf("Service.LoadTemplates: Loaded %d templates:\n", len(templates))
	for _, template := range templates {
		fmt.Printf("  - %s (ID: %s, Services: %d)\n", template.Name, template.ID, len(template.Services))
		for _, service := range template.Services {
			fmt.Printf("    * %s (ID: %s, Type: %s)\n", service.Name, service.ServiceID, service.Type)
		}
	}
	return nil
}

// LoadServiceConfigs loads service configurations from a directory
func (s *Service) LoadServiceConfigs(dirPath string) error {
	fmt.Printf("Service.LoadServiceConfigs: Loading from %s\n", dirPath)
	serviceConfigLoader := NewServiceConfigLoader(s.serviceConfigManager)
	err := serviceConfigLoader.LoadServiceConfigsFromDirectory(dirPath)
	if err != nil {
		fmt.Printf("Service.LoadServiceConfigs: Failed to load: %v\n", err)
		return err
	}

	// Log what was loaded
	configs := s.serviceConfigManager.GetAllServiceConfigs()
	fmt.Printf("Service.LoadServiceConfigs: Loaded %d service configurations:\n", len(configs))
	for _, config := range configs {
		fmt.Printf("  - %s (ID: %s, Type: %s, Active: %v)\n", config.Name, config.ID, config.Type, config.IsActive)
	}
	return nil
}

// LoadServiceLimits loads service limits from a directory
func (s *Service) LoadServiceLimits(dirPath string) error {
	fmt.Printf("Service.LoadServiceLimits: Loading from %s\n", dirPath)
	serviceConfigLoader := NewServiceConfigLoader(s.serviceConfigManager)
	err := serviceConfigLoader.LoadServiceLimitsFromDirectory(dirPath)
	if err != nil {
		fmt.Printf("Service.LoadServiceLimits: Failed to load: %v\n", err)
		return err
	}

	// Log what was loaded
	limits := s.serviceConfigManager.GetAllServiceLimits()
	fmt.Printf("Service.LoadServiceLimits: Loaded %d service limits:\n", len(limits))
	for _, limit := range limits {
		fmt.Printf("  - %s (ServiceID: %s, MaxLabs: %d, Active: %v)\n", limit.ID, limit.ServiceID, limit.MaxLabs, limit.IsActive)
	}
	return nil
}

// GetTemplates returns all available lab templates
func (s *Service) GetTemplates() []*models.LabTemplate {
	return s.templateManager.GetAllTemplates()
}

// GetTemplate returns a specific lab template
func (s *Service) GetTemplate(templateID string) (*models.LabTemplate, bool) {
	return s.templateManager.GetTemplate(templateID)
}

// EnrichTemplatesWithServiceTypes enriches all templates with service type information
func (s *Service) EnrichTemplatesWithServiceTypes() {
	s.templateManager.EnrichTemplatesWithServiceTypes(s.serviceConfigManager)
}

// CreateLabFromTemplate creates a lab from a template
func (s *Service) CreateLabFromTemplate(templateID, ownerID string) (*models.Lab, error) {
	fmt.Printf("CreateLabFromTemplate: Starting lab creation for template %s, owner %s\n", templateID, ownerID)

	// Get the template
	template, exists := s.templateManager.GetTemplate(templateID)
	if !exists {
		fmt.Printf("CreateLabFromTemplate: Template %s not found\n", templateID)
		return nil, errors.New("template not found")
	}
	fmt.Printf("CreateLabFromTemplate: Found template %s with %d services\n", templateID, len(template.Services))

	// Check service availability and limits for all services in the template
	for _, serviceRef := range template.Services {
		fmt.Printf("CreateLabFromTemplate: Checking service %s (ID: %s)\n", serviceRef.Name, serviceRef.ServiceID)

		// Get current usage for this service
		currentUsage := s.getServiceUsage(serviceRef.ServiceID)
		fmt.Printf("CreateLabFromTemplate: Service %s current usage: %d\n", serviceRef.ServiceID, currentUsage)

		// Check if service is available and within limits
		if err := s.serviceConfigManager.CheckServiceAvailability(serviceRef.ServiceID, currentUsage); err != nil {
			fmt.Printf("CreateLabFromTemplate: Service %s availability check failed: %v\n", serviceRef.ServiceID, err)
			return nil, fmt.Errorf("service %s (%s) not available: %w", serviceRef.Name, serviceRef.ServiceID, err)
		}
		fmt.Printf("CreateLabFromTemplate: Service %s availability check passed\n", serviceRef.ServiceID)
	}

	fmt.Printf("CreateLabFromTemplate: All service checks passed, creating lab from template\n")
	lab, err := s.templateLoader.CreateLabFromTemplate(templateID, ownerID)
	if err != nil {
		fmt.Printf("CreateLabFromTemplate: Failed to create lab from template: %v\n", err)
		return nil, err
	}
	fmt.Printf("CreateLabFromTemplate: Lab created with ID %s\n", lab.ID)

	s.mu.Lock()
	s.labs[lab.ID] = lab
	s.mu.Unlock()

	// Initialize progress tracking
	s.progressTracker.InitializeProgress(lab.ID)
	s.progressTracker.AddLog(lab.ID, "Lab creation started from template")

	// Start lab provisioning
	fmt.Printf("CreateLabFromTemplate: Starting lab provisioning for lab %s\n", lab.ID)
	go s.provisionLabFromTemplate(lab.ID, templateID)

	fmt.Printf("CreateLabFromTemplate: Lab creation completed successfully for lab %s\n", lab.ID)
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

// getServiceUsage returns the current number of active labs using a specific service
func (s *Service) getServiceUsage(serviceID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, lab := range s.labs {
		if lab.Status == models.LabStatusReady || lab.Status == models.LabStatusProvisioning {
			// Check if this lab uses the specified service
			for _, usedService := range lab.UsedServices {
				if usedService == serviceID {
					count++
					break
				}
			}
		}
	}
	return count
}

// GetServiceUsage returns usage information for all services
func (s *Service) GetServiceUsage() []*models.ServiceUsage {
	usage := make([]*models.ServiceUsage, 0)

	// Get all service configs
	configs := s.serviceConfigManager.GetAllServiceConfigs()

	for _, config := range configs {
		currentUsage := s.getServiceUsage(config.ID)
		usageInfo := s.serviceConfigManager.GetServiceUsage(config.ID, currentUsage)
		usage = append(usage, usageInfo)
	}

	return usage
}

// GetServiceConfigManager returns the service configuration manager
func (s *Service) GetServiceConfigManager() *models.ServiceConfigManager {
	return s.serviceConfigManager
}
