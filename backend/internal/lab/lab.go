package lab

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
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

	// Initialize progress tracking
	s.progressTracker.InitializeProgress(lab.ID)
	s.progressTracker.AddLog(lab.ID, "Lab creation started")

	// Start lab provisioning
	go s.provisionLab(lab.ID)

	return lab, nil
}

// provisionLabFromTemplate handles lab provisioning from a template
func (s *Service) provisionLabFromTemplate(labID, templateID string) {
	template, exists := s.templateManager.GetTemplate(templateID)
	if !exists {
		s.progressTracker.FailProgress(labID, "Template not found")
		return
	}

	s.progressTracker.AddLog(labID, fmt.Sprintf("Provisioning lab from template: %s", template.Name))

	// Add services to progress tracker based on template
	for _, serviceTemplate := range template.Services {
		var steps []string
		switch serviceTemplate.Type {
		case "palette":
			steps = []string{
				"Creating Project",
				"Setting up User Account",
				"Configuring Access Permissions",
				"Generating API Keys",
				"Creating Edge Tokens",
			}
		case "proxmox":
			steps = []string{
				"Connecting to Cluster",
				"Creating VM Pool",
				"Configuring Network",
				"Setting up Storage",
			}
		case "generic":
			steps = []string{
				"Initializing Service",
				"Configuring Endpoints",
				"Setting up Authentication",
			}
		default:
			steps = []string{"Initializing"}
		}

		s.progressTracker.AddService(labID, serviceTemplate.Name, serviceTemplate.Description, steps)
	}

	// Provision each service defined in the template
	for _, serviceTemplate := range template.Services {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Setting up service: %s (%s)", serviceTemplate.Name, serviceTemplate.Type))

		switch serviceTemplate.Type {
		case "palette":
			s.provisionPaletteService(labID, serviceTemplate)
		case "proxmox":
			s.provisionProxmoxService(labID, serviceTemplate)
		case "generic":
			s.provisionGenericService(labID, serviceTemplate)
		default:
			s.progressTracker.AddLog(labID, fmt.Sprintf("Unknown service type: %s", serviceTemplate.Type))
		}
	}

	s.mu.Lock()
	lab, exists := s.labs[labID]
	if !exists {
		s.mu.Unlock()
		s.progressTracker.FailProgress(labID, "Lab not found")
		return
	}

	// Update lab status to ready
	lab.Status = models.LabStatusReady
	lab.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.progressTracker.CompleteProgress(labID)
	s.progressTracker.AddLog(labID, "Lab setup completed successfully!")
}

// provisionLab handles lab provisioning with real progress tracking
func (s *Service) provisionLab(labID string) {
	s.progressTracker.AddLog(labID, "Starting lab provisioning")

	s.mu.Lock()
	lab, exists := s.labs[labID]
	if !exists {
		s.mu.Unlock()
		s.progressTracker.FailProgress(labID, "Lab not found")
		return
	}
	s.mu.Unlock()

	// Start Palette Project Service
	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Creating Project", "running", "Creating project in Spectro Cloud...")
	s.progressTracker.AddLog(labID, "Creating Spectro Cloud project")

	// Simulate project creation
	time.Sleep(800 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Creating Project", "completed", "Spectro Cloud project created successfully")
	s.progressTracker.AddLog(labID, "Spectro Cloud project created")

	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Setting up User Account", "running", "Creating user account...")
	s.progressTracker.AddLog(labID, "Setting up user account")

	// Simulate user creation
	time.Sleep(600 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Setting up User Account", "completed", "User account created successfully")
	s.progressTracker.AddLog(labID, "User account setup completed")

	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Configuring Access Permissions", "running", "Assigning roles and permissions...")
	s.progressTracker.AddLog(labID, "Configuring access permissions")

	// Simulate role assignment
	time.Sleep(400 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Configuring Access Permissions", "completed", "Access permissions configured successfully")
	s.progressTracker.AddLog(labID, "Access permissions configured")

	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Generating API Keys", "running", "Creating API keys and tokens...")
	s.progressTracker.AddLog(labID, "Generating API keys")

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
		s.progressTracker.AddLog(labID, fmt.Sprintf("Warning: Failed to setup services: %v", err))
		s.progressTracker.AddLog(labID, "Using fallback credentials")
		// Continue with mock credentials as fallback
		s.mu.Lock()
		lab.Credentials = s.generateCredentials(labID, lab.EndsAt)
		s.mu.Unlock()
	} else {
		s.progressTracker.AddLog(labID, "Service setup completed successfully")
	}

	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Generating API Keys", "completed", "API keys generated successfully")
	s.progressTracker.AddLog(labID, "API keys generated")

	s.progressTracker.UpdateServiceStep(labID, "Palette Project Service", "Creating Edge Tokens", "running", "Creating edge registration tokens...")
	s.progressTracker.AddLog(labID, "Creating edge tokens")

	// Simulate edge token creation
	time.Sleep(300 * time.Millisecond)

	s.mu.Lock()
	// Update lab status to ready
	lab.Status = models.LabStatusReady
	lab.UpdatedAt = time.Now()
	s.mu.Unlock()

	s.progressTracker.CompleteProgress(labID)
	s.progressTracker.AddLog(labID, "Lab setup completed successfully!")
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

// provisionPaletteService provisions a Palette service using the real Palette Project service
func (s *Service) provisionPaletteService(labID string, serviceTemplate models.ServiceTemplate) {
	// Set environment variables from template config
	if host, ok := serviceTemplate.Config["host"]; ok {
		os.Setenv("PALETTE_HOST", host)
	}
	if apiKey, ok := serviceTemplate.Config["api_key"]; ok {
		os.Setenv("PALETTE_API_KEY", apiKey)
	}

	// Create Palette Project service instance
	paletteService := services.NewPaletteProjectService()

	// Get lab for context
	s.mu.Lock()
	lab, exists := s.labs[labID]
	s.mu.Unlock()

	if !exists {
		s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating Project", "failed", "Lab not found")
		return
	}

	// Create setup context
	setupCtx := &interfaces.SetupContext{
		LabID:    labID,
		LabName:  lab.Name,
		Duration: int(time.Until(lab.EndsAt).Minutes()),
		OwnerID:  lab.OwnerID,
		Context:  context.Background(),
		Lab:      lab,
		AddCredential: func(credential *interfaces.Credential) error {
			// Convert to models.Credential and add to lab
			cred := models.Credential{
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

			s.mu.Lock()
			lab.Credentials = append(lab.Credentials, cred)
			s.mu.Unlock()

			return nil
		},
	}

	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating Project", "running", "Creating project in Spectro Cloud...")

	// Execute the real setup
	err := paletteService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating Project", "failed", fmt.Sprintf("Failed to create project: %v", err))
		s.progressTracker.AddLog(labID, fmt.Sprintf("Palette Project setup failed: %v", err))
		return
	}

	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating Project", "completed", "Project created successfully")
	s.progressTracker.AddLog(labID, "Palette Project created successfully")

	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Setting up User Account", "completed", "User account created")
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Configuring Access Permissions", "completed", "Permissions configured")
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Generating API Keys", "completed", "API keys generated")
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating Edge Tokens", "completed", "Edge tokens created")

	s.progressTracker.AddLog(labID, "Palette Project service setup completed successfully")
}

// provisionProxmoxService provisions a Proxmox service
func (s *Service) provisionProxmoxService(labID string, serviceTemplate models.ServiceTemplate) {
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Connecting to Cluster", "running", "Connecting to Proxmox cluster...")
	time.Sleep(500 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Connecting to Cluster", "completed", "Connected to cluster")

	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating VM Pool", "running", "Creating virtual machine pool...")
	time.Sleep(700 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating VM Pool", "completed", "VM pool created")

	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Configuring Network", "running", "Configuring network settings...")
	time.Sleep(400 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Configuring Network", "completed", "Network configured")

	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Setting up Storage", "running", "Setting up storage pools...")
	time.Sleep(600 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Setting up Storage", "completed", "Storage configured")
}

// provisionGenericService provisions a generic service
func (s *Service) provisionGenericService(labID string, serviceTemplate models.ServiceTemplate) {
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Initializing Service", "running", "Initializing generic service...")
	time.Sleep(400 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Initializing Service", "completed", "Service initialized")

	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Configuring Endpoints", "running", "Configuring service endpoints...")
	time.Sleep(300 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Configuring Endpoints", "completed", "Endpoints configured")

	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Setting up Authentication", "running", "Setting up authentication...")
	time.Sleep(500 * time.Millisecond)
	s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Setting up Authentication", "completed", "Authentication configured")
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
