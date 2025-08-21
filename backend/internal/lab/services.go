package lab

import (
	"context"
	"fmt"
	"os"
	"time"

	"spectro-lab-backend/internal/interfaces"
	"spectro-lab-backend/internal/models"
	"spectro-lab-backend/internal/services"
)

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
		UpdateProgress: func(stepName, status, message string) {
			s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, stepName, status, message)
		},
	}

	// Execute the real setup - services will update their own progress
	err := paletteService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Palette Project setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Palette Project setup failed: %v", err))

		// Set lab status to error
		s.mu.Lock()
		if lab, exists := s.labs[labID]; exists {
			lab.Status = models.LabStatusError
			lab.UpdatedAt = time.Now()
		}
		s.mu.Unlock()
		return
	}

	s.progressTracker.AddLog(labID, "Palette Project service setup completed successfully")
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

// provisionProxmoxUserService provisions a Proxmox user service using the real Proxmox User service
func (s *Service) provisionProxmoxUserService(labID string, serviceTemplate models.ServiceTemplate) {
	// Set environment variables from template config
	if uri, ok := serviceTemplate.Config["uri"]; ok {
		os.Setenv("PROXMOX_URI", uri)
	}
	if adminUser, ok := serviceTemplate.Config["admin_user"]; ok {
		os.Setenv("PROXMOX_ADMIN_USER", adminUser)
	}
	if adminPass, ok := serviceTemplate.Config["admin_pass"]; ok {
		os.Setenv("PROXMOX_ADMIN_PASS", adminPass)
	}
	if skipTLSVerify, ok := serviceTemplate.Config["skip_tls_verify"]; ok {
		os.Setenv("PROXMOX_SKIP_TLS_VERIFY", skipTLSVerify)
	}

	// Create Proxmox User service instance
	proxmoxUserService := services.NewProxmoxUserService()

	// Get lab for context
	s.mu.Lock()
	lab, exists := s.labs[labID]
	s.mu.Unlock()

	if !exists {
		s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating User", "failed", "Lab not found")
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
		UpdateProgress: func(stepName, status, message string) {
			s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, stepName, status, message)
		},
	}

	// Execute the real setup - services will update their own progress
	err := proxmoxUserService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Proxmox user setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Proxmox user setup failed: %v", err))

		// Set lab status to error
		s.mu.Lock()
		if lab, exists := s.labs[labID]; exists {
			lab.Status = models.LabStatusError
			lab.UpdatedAt = time.Now()
		}
		s.mu.Unlock()
		return
	}

	s.progressTracker.AddLog(labID, "Proxmox user created successfully")
}

// provisionPaletteTenantService provisions a Palette Tenant service using the real Palette Tenant service
func (s *Service) provisionPaletteTenantService(labID string, serviceTemplate models.ServiceTemplate) {
	s.progressTracker.AddLog(labID, fmt.Sprintf("Starting Palette Tenant service setup for lab %s", labID))

	// Set environment variables from template config with comprehensive logging
	s.progressTracker.AddLog(labID, "Setting up environment variables from template config...")

	if host, ok := serviceTemplate.Config["palette_host"]; ok {
		os.Setenv("palette_host", host)
		s.progressTracker.AddLog(labID, fmt.Sprintf("Set palette_host: %s", host))
	} else {
		s.progressTracker.AddLog(labID, "Warning: palette_host not found in template config")
	}

	if systemUsername, ok := serviceTemplate.Config["palette_system_username"]; ok {
		os.Setenv("palette_system_username", systemUsername)
		s.progressTracker.AddLog(labID, fmt.Sprintf("Set palette_system_username: %s", systemUsername))
	} else {
		s.progressTracker.AddLog(labID, "Warning: palette_system_username not found in template config")
	}

	if systemPassword, ok := serviceTemplate.Config["palette_system_password"]; ok {
		os.Setenv("palette_system_password", systemPassword)
		s.progressTracker.AddLog(labID, "Set palette_system_password: [REDACTED]")
	} else {
		s.progressTracker.AddLog(labID, "Warning: palette_system_password not found in template config")
	}

	// Log all available config keys for debugging
	configKeys := make([]string, 0, len(serviceTemplate.Config))
	for key := range serviceTemplate.Config {
		configKeys = append(configKeys, key)
	}
	s.progressTracker.AddLog(labID, fmt.Sprintf("Available config keys: %v", configKeys))

	// Create Palette Tenant service instance
	s.progressTracker.AddLog(labID, "Creating Palette Tenant service instance...")
	paletteTenantService := services.NewPaletteTenantService()

	// Log the environment variables that the service will use
	s.progressTracker.AddLog(labID, fmt.Sprintf("Service will use palette_host: %s", os.Getenv("palette_host")))
	s.progressTracker.AddLog(labID, fmt.Sprintf("Service will use palette_system_username: %s", os.Getenv("palette_system_username")))
	s.progressTracker.AddLog(labID, "Service will use palette_system_password: [REDACTED]")

	// Get lab for context
	s.mu.Lock()
	lab, exists := s.labs[labID]
	s.mu.Unlock()

	if !exists {
		s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, "Creating User", "failed", "Lab not found")
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
		UpdateProgress: func(stepName, status, message string) {
			s.progressTracker.UpdateServiceStep(labID, serviceTemplate.Name, stepName, status, message)
		},
	}

	// Execute the real setup - services will update their own progress
	err := paletteTenantService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Palette tenant setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Palette tenant setup failed: %v", err))

		// Set lab status to error
		s.mu.Lock()
		if lab, exists := s.labs[labID]; exists {
			lab.Status = models.LabStatusError
			lab.UpdatedAt = time.Now()
		}
		s.mu.Unlock()
		return
	}

	s.progressTracker.AddLog(labID, "Palette tenant user created successfully")
	s.progressTracker.AddLog(labID, fmt.Sprintf("Palette Tenant service setup completed successfully for lab %s", labID))
}
