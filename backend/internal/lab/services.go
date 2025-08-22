package lab

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wcrum/labby/internal/interfaces"
	"github.com/wcrum/labby/internal/models"
	"github.com/wcrum/labby/internal/services"
)

// provisionPaletteService provisions a Palette service using the real Palette Project service
func (s *Service) provisionPaletteService(labID string, serviceConfig *models.ServiceConfig) {
	// Set environment variables from service config
	if host, ok := serviceConfig.Config["host"]; ok {
		os.Setenv("PALETTE_HOST", host)
	}
	if apiKey, ok := serviceConfig.Config["api_key"]; ok {
		os.Setenv("PALETTE_API_KEY", apiKey)
	}

	// Create Palette Project service instance
	paletteService := services.NewPaletteProjectService()

	// Get lab for context
	s.mu.Lock()
	lab, exists := s.labs[labID]
	s.mu.Unlock()

	if !exists {
		s.progressTracker.UpdateServiceStep(labID, serviceConfig.Name, "Creating Project", "failed", "Lab not found")
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
			s.progressTracker.UpdateServiceStep(labID, serviceConfig.Name, stepName, status, message)
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

// provisionProxmoxUserService provisions a Proxmox user service using the real Proxmox User service
func (s *Service) provisionProxmoxUserService(labID string, serviceConfig *models.ServiceConfig) {
	// Set environment variables from service config
	if uri, ok := serviceConfig.Config["uri"]; ok {
		os.Setenv("PROXMOX_URI", uri)
	}
	if adminUser, ok := serviceConfig.Config["admin_user"]; ok {
		os.Setenv("PROXMOX_ADMIN_USER", adminUser)
	}
	if adminPass, ok := serviceConfig.Config["admin_pass"]; ok {
		os.Setenv("PROXMOX_ADMIN_PASS", adminPass)
	}
	if skipTLSVerify, ok := serviceConfig.Config["skip_tls_verify"]; ok {
		os.Setenv("PROXMOX_SKIP_TLS_VERIFY", skipTLSVerify)
	}

	// Create Proxmox User service instance
	proxmoxUserService := services.NewProxmoxUserService()

	// Get lab for context
	s.mu.Lock()
	lab, exists := s.labs[labID]
	s.mu.Unlock()

	if !exists {
		s.progressTracker.UpdateServiceStep(labID, serviceConfig.Name, "Creating User", "failed", "Lab not found")
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
			s.progressTracker.UpdateServiceStep(labID, serviceConfig.Name, stepName, status, message)
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
func (s *Service) provisionPaletteTenantService(labID string, serviceConfig *models.ServiceConfig) {
	s.progressTracker.AddLog(labID, fmt.Sprintf("Starting Palette Tenant service setup for lab %s", labID))

	// Set environment variables from service config with comprehensive logging
	s.progressTracker.AddLog(labID, "Setting up environment variables from service config...")

	if host, ok := serviceConfig.Config["palette_host"]; ok {
		os.Setenv("palette_host", host)
		s.progressTracker.AddLog(labID, fmt.Sprintf("Set palette_host: %s", host))
	} else {
		s.progressTracker.AddLog(labID, "Warning: palette_host not found in service config")
	}

	if systemUsername, ok := serviceConfig.Config["palette_system_username"]; ok {
		os.Setenv("palette_system_username", systemUsername)
		s.progressTracker.AddLog(labID, fmt.Sprintf("Set palette_system_username: %s", systemUsername))
	} else {
		s.progressTracker.AddLog(labID, "Warning: palette_system_username not found in service config")
	}

	if systemPassword, ok := serviceConfig.Config["palette_system_password"]; ok {
		os.Setenv("palette_system_password", systemPassword)
		s.progressTracker.AddLog(labID, "Set palette_system_password: [REDACTED]")
	} else {
		s.progressTracker.AddLog(labID, "Warning: palette_system_password not found in service config")
	}

	// Log all available config keys for debugging
	configKeys := make([]string, 0, len(serviceConfig.Config))
	for key := range serviceConfig.Config {
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
		s.progressTracker.UpdateServiceStep(labID, serviceConfig.Name, "Creating User", "failed", "Lab not found")
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
			s.progressTracker.UpdateServiceStep(labID, serviceConfig.Name, stepName, status, message)
		},
	}

	// Execute the real setup - services will update their own progress
	err := paletteTenantService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Palette Tenant setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Palette Tenant setup failed: %v", err))

		// Set lab status to error
		s.mu.Lock()
		if lab, exists := s.labs[labID]; exists {
			lab.Status = models.LabStatusError
			lab.UpdatedAt = time.Now()
		}
		s.mu.Unlock()
		return
	}

	s.progressTracker.AddLog(labID, "Palette Tenant service setup completed successfully")
}
