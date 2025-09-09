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
	// Create Palette Project service instance
	paletteService := services.NewPaletteProjectService()

	// Configure the service with credentials from service config
	paletteService.ConfigureFromServiceConfig(serviceConfig)

	// Get lab for context
	lab, err := s.repo.GetLabByID(labID)
	if err != nil {
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
			// Convert to models.Credential and save to database
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

			// Save credential to database
			if err := s.repo.CreateCredential(&cred); err != nil {
				return fmt.Errorf("failed to save credential to database: %w", err)
			}

			// Also add to lab object in memory for consistency
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
	err = paletteService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Palette Project setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Palette Project setup failed: %v", err))

		// Set lab status to error
		lab.Status = models.LabStatusError
		lab.UpdatedAt = time.Now()
		s.repo.UpdateLab(lab)
		return
	}

	// Save the lab with updated ServiceData to database
	lab.UpdatedAt = time.Now()
	fmt.Printf("Saving lab %s with ServiceData keys: %v\n", labID, func() []string {
		if lab.ServiceData == nil {
			return []string{"nil"}
		}
		keys := make([]string, 0, len(lab.ServiceData))
		for k := range lab.ServiceData {
			keys = append(keys, k)
		}
		return keys
	}())
	if err := s.repo.UpdateLab(lab); err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Failed to save lab ServiceData: %v", err))
	} else {
		fmt.Printf("Successfully saved lab %s ServiceData to database\n", labID)
	}

	s.progressTracker.AddLog(labID, "Palette Project service setup completed successfully")
}

// provisionProxmoxUserService provisions a Proxmox user service using the real Proxmox User service
func (s *Service) provisionProxmoxUserService(labID string, serviceConfig *models.ServiceConfig) {
	// Create Proxmox User service instance
	proxmoxUserService := services.NewProxmoxUserService()

	// Configure the service from the service configuration
	proxmoxUserService.ConfigureFromServiceConfig(serviceConfig.Config)

	// Get lab for context
	lab, err := s.repo.GetLabByID(labID)
	if err != nil {
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
			// Convert to models.Credential and save to database
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

			// Save credential to database
			if err := s.repo.CreateCredential(&cred); err != nil {
				return fmt.Errorf("failed to save credential to database: %w", err)
			}

			// Also add to lab object in memory for consistency
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
	err = proxmoxUserService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Proxmox user setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Proxmox user setup failed: %v", err))

		// Set lab status to error
		lab.Status = models.LabStatusError
		lab.UpdatedAt = time.Now()
		s.repo.UpdateLab(lab)
		return
	}

	// Save the lab with updated ServiceData to database
	lab.UpdatedAt = time.Now()
	fmt.Printf("Saving lab %s with ServiceData keys: %v\n", labID, func() []string {
		if lab.ServiceData == nil {
			return []string{"nil"}
		}
		keys := make([]string, 0, len(lab.ServiceData))
		for k := range lab.ServiceData {
			keys = append(keys, k)
		}
		return keys
	}())
	if err := s.repo.UpdateLab(lab); err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Failed to save lab ServiceData: %v", err))
	} else {
		fmt.Printf("Successfully saved lab %s ServiceData to database\n", labID)
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

	// Configure the service with credentials from service config
	paletteTenantService.ConfigureFromServiceConfig(serviceConfig)

	// Log the service config values that the service will use
	if host, ok := serviceConfig.Config["palette_host"]; ok {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Service will use palette_host: %s", host))
	}
	if systemUsername, ok := serviceConfig.Config["palette_system_username"]; ok {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Service will use palette_system_username: %s", systemUsername))
	}
	if _, ok := serviceConfig.Config["palette_system_password"]; ok {
		s.progressTracker.AddLog(labID, "Service will use palette_system_password: [REDACTED]")
	}

	// Get lab for context
	lab, err := s.repo.GetLabByID(labID)
	if err != nil {
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
			// Convert to models.Credential and save to database
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

			// Save credential to database
			if err := s.repo.CreateCredential(&cred); err != nil {
				return fmt.Errorf("failed to save credential to database: %w", err)
			}

			// Also add to lab object in memory for consistency
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
	err = paletteTenantService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Palette Tenant setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Palette Tenant setup failed: %v", err))

		// Set lab status to error
		lab.Status = models.LabStatusError
		lab.UpdatedAt = time.Now()
		s.repo.UpdateLab(lab)
		return
	}

	// Save the lab with updated ServiceData to database
	lab.UpdatedAt = time.Now()
	fmt.Printf("Saving lab %s with ServiceData keys: %v\n", labID, func() []string {
		if lab.ServiceData == nil {
			return []string{"nil"}
		}
		keys := make([]string, 0, len(lab.ServiceData))
		for k := range lab.ServiceData {
			keys = append(keys, k)
		}
		return keys
	}())
	if err := s.repo.UpdateLab(lab); err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Failed to save lab ServiceData: %v", err))
	} else {
		fmt.Printf("Successfully saved lab %s ServiceData to database\n", labID)
	}

	s.progressTracker.AddLog(labID, "Palette Tenant service setup completed successfully")
}

// provisionTerraformCloudService provisions a Terraform Cloud service
func (s *Service) provisionTerraformCloudService(labID string, serviceConfig *models.ServiceConfig) {
	s.progressTracker.AddLog(labID, fmt.Sprintf("Service will use tf_cloud_host: %s", serviceConfig.Config["host"]))
	s.progressTracker.AddLog(labID, fmt.Sprintf("Service will use tf_cloud_organization: %s", serviceConfig.Config["organization"]))

	// Get lab for context
	lab, err := s.repo.GetLabByID(labID)
	if err != nil {
		s.progressTracker.UpdateServiceStep(labID, serviceConfig.Name, "Creating Workspace", "failed", "Lab not found")
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
			// Convert to models.Credential and save to database
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

			// Save credential to database
			if err := s.repo.CreateCredential(&cred); err != nil {
				return fmt.Errorf("failed to save credential to database: %w", err)
			}

			// Also add to lab object in memory for consistency
			s.mu.Lock()
			lab.Credentials = append(lab.Credentials, cred)
			s.mu.Unlock()

			return nil
		},
		UpdateProgress: func(stepName, status, message string) {
			s.progressTracker.UpdateServiceStep(labID, serviceConfig.Name, stepName, status, message)
		},
	}

	// Create Terraform Cloud service instance
	terraformCloudService := services.NewTerraformCloudService()

	// Configure the service from the service configuration
	terraformCloudService.ConfigureFromServiceConfig(serviceConfig.Config, labID)

	// Execute the real setup - services will update their own progress
	err = terraformCloudService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Terraform Cloud setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Terraform Cloud setup failed: %v", err))

		// Set lab status to error
		lab.Status = models.LabStatusError
		lab.UpdatedAt = time.Now()
		s.repo.UpdateLab(lab)
		return
	}

	// Save the lab with updated ServiceData to database
	lab.UpdatedAt = time.Now()
	fmt.Printf("Saving lab %s with ServiceData keys: %v\n", labID, func() []string {
		if lab.ServiceData == nil {
			return []string{"nil"}
		}
		keys := make([]string, 0, len(lab.ServiceData))
		for k := range lab.ServiceData {
			keys = append(keys, k)
		}
		return keys
	}())
	if err := s.repo.UpdateLab(lab); err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Failed to save lab ServiceData: %v", err))
	} else {
		fmt.Printf("Successfully saved lab %s ServiceData to database\n", labID)
	}

	s.progressTracker.AddLog(labID, fmt.Sprintf("Terraform Cloud setup completed for lab %s", lab.Name))
}

// provisionGuacamoleService provisions a Guacamole service using the real Guacamole service
func (s *Service) provisionGuacamoleService(labID string, serviceConfig *models.ServiceConfig) {
	// Create Guacamole service instance
	guacamoleService := services.NewGuacamoleService()

	// Configure the service from the service configuration
	guacamoleService.ConfigureFromServiceConfig(serviceConfig.Config)

	// Get lab for context
	lab, err := s.repo.GetLabByID(labID)
	if err != nil {
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
			// Convert to models.Credential and save to database
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

			// Save credential to database
			if err := s.repo.CreateCredential(&cred); err != nil {
				return fmt.Errorf("failed to save credential to database: %w", err)
			}

			// Also add to lab object in memory for consistency
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
	err = guacamoleService.ExecuteSetup(setupCtx)
	if err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Guacamole setup failed: %v", err))
		s.progressTracker.FailProgress(labID, fmt.Sprintf("Guacamole setup failed: %v", err))

		// Set lab status to error
		lab.Status = models.LabStatusError
		lab.UpdatedAt = time.Now()
		s.repo.UpdateLab(lab)
		return
	}

	// Save the lab with updated ServiceData to database
	lab.UpdatedAt = time.Now()
	fmt.Printf("Saving lab %s with ServiceData keys: %v\n", labID, func() []string {
		if lab.ServiceData == nil {
			return []string{"nil"}
		}
		keys := make([]string, 0, len(lab.ServiceData))
		for k := range lab.ServiceData {
			keys = append(keys, k)
		}
		return keys
	}())
	if err := s.repo.UpdateLab(lab); err != nil {
		s.progressTracker.AddLog(labID, fmt.Sprintf("Failed to save lab ServiceData: %v", err))
	} else {
		fmt.Printf("Successfully saved lab %s ServiceData to database\n", labID)
	}

	s.progressTracker.AddLog(labID, "Guacamole user created successfully")
}
