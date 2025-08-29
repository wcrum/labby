package lab

import (
	"fmt"
	"time"

	"github.com/wcrum/labby/internal/models"
)

// provisionLabFromTemplate handles lab provisioning from a template
func (s *Service) provisionLabFromTemplate(labID, templateID string) {
	template, exists := s.templateManager.GetTemplate(templateID)
	if !exists {
		s.progressTracker.FailProgress(labID, "Template not found")
		// Set lab status to error
		s.mu.Lock()
		if lab, exists := s.labs[labID]; exists {
			lab.Status = models.LabStatusError
			lab.UpdatedAt = time.Now()
		}
		s.mu.Unlock()
		return
	}

	s.progressTracker.AddLog(labID, fmt.Sprintf("Provisioning lab from template: %s", template.Name))

	// Add services to progress tracker based on template
	for _, serviceRef := range template.Services {
		// Get the service configuration
		serviceConfig, exists := s.serviceConfigManager.GetServiceConfig(serviceRef.ServiceID)
		if !exists {
			s.progressTracker.AddLog(labID, fmt.Sprintf("Service configuration not found: %s", serviceRef.ServiceID))
			continue
		}

		var steps []string
		switch serviceConfig.Type {
		case "palette_project":
			steps = []string{
				"Creating Project",
				"Setting up User Account",
				"Configuring Access Permissions",
				"Generating API Keys",
				"Creating Edge Tokens",
			}
		case "proxmox_user":
			steps = []string{
				"Connecting to Proxmox",
				"Creating User Account",
				"Setting Password",
			}
		case "palette_tenant":
			steps = []string{
				"Connecting to Palette",
				"Creating User Account",
				"Setting Password",
			}
		default:
			steps = []string{"Initializing"}
		}

		s.progressTracker.AddService(labID, serviceConfig.Name, serviceRef.Description, steps)
	}

	// Provision each service defined in the template
	hasFailures := false
	for _, serviceRef := range template.Services {
		// Get the service configuration
		serviceConfig, exists := s.serviceConfigManager.GetServiceConfig(serviceRef.ServiceID)
		if !exists {
			s.progressTracker.AddLog(labID, fmt.Sprintf("Service configuration not found: %s", serviceRef.ServiceID))
			continue
		}

		s.progressTracker.AddLog(labID, fmt.Sprintf("Setting up service: %s (%s)", serviceRef.Name, serviceConfig.Type))

		switch serviceConfig.Type {
		case "palette_project":
			s.provisionPaletteService(labID, serviceConfig)
		case "proxmox_user":
			s.provisionProxmoxUserService(labID, serviceConfig)
		case "palette_tenant":
			s.provisionPaletteTenantService(labID, serviceConfig)
		case "terraform_cloud":
			s.provisionTerraformCloudService(labID, serviceConfig)
		default:
			s.progressTracker.AddLog(labID, fmt.Sprintf("Unknown service type: %s", serviceConfig.Type))
		}

		// Check if the lab status is now error (indicating a failure)
		s.mu.RLock()
		if lab, exists := s.labs[labID]; exists && lab.Status == models.LabStatusError {
			hasFailures = true
		}
		s.mu.RUnlock()

		// If we have failures, stop provisioning
		if hasFailures {
			break
		}
	}

	s.mu.Lock()
	lab, exists := s.labs[labID]
	if !exists {
		s.mu.Unlock()
		s.progressTracker.FailProgress(labID, "Lab not found")
		return
	}

	// Only set status to ready if no failures occurred
	if !hasFailures {
		lab.Status = models.LabStatusReady
		lab.UpdatedAt = time.Now()
		s.progressTracker.CompleteProgress(labID)
		s.progressTracker.AddLog(labID, "Lab setup completed successfully!")
	} else {
		// Ensure lab status is set to error if not already set
		if lab.Status != models.LabStatusError {
			lab.Status = models.LabStatusError
			lab.UpdatedAt = time.Now()
		}
		s.progressTracker.AddLog(labID, "Lab setup failed due to service errors")
	}
	s.mu.Unlock()
}
