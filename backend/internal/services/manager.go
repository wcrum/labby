package services

import (
	"fmt"

	"github.com/wcrum/labby/internal/interfaces"
)

// ServiceManager manages all available services
type ServiceManager struct {
	registry *interfaces.ServiceRegistry
	// Map Service Config IDs to service instances
	serviceConfigMap map[string]interfaces.Service
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	registry := interfaces.NewServiceRegistry()

	// Create service instances
	paletteProjectService := NewPaletteProjectService()
	proxmoxUserService := NewProxmoxUserService()
	paletteTenantService := NewPaletteTenantService()

	// Register services with their GetName() for backward compatibility
	registry.RegisterService(paletteProjectService)
	registry.RegisterService(proxmoxUserService)
	registry.RegisterService(paletteTenantService)

	// Create mapping from Service Config IDs to service instances
	serviceConfigMap := make(map[string]interfaces.Service)

	// Map known Service Config IDs to their corresponding services
	// This mapping should be updated when new service configs are added
	serviceConfigMap["palette-project-1"] = paletteProjectService
	serviceConfigMap["proxmox-user-1"] = proxmoxUserService
	serviceConfigMap["palette-tenant-1"] = paletteTenantService

	return &ServiceManager{
		registry:         registry,
		serviceConfigMap: serviceConfigMap,
	}
}

// GetRegistry returns the service registry
func (sm *ServiceManager) GetRegistry() *interfaces.ServiceRegistry {
	return sm.registry
}

// GetServiceByConfigID returns a service by its Service Config ID
func (sm *ServiceManager) GetServiceByConfigID(configID string) (interfaces.Service, bool) {
	service, exists := sm.serviceConfigMap[configID]
	return service, exists
}

// SetupLabServices sets up all services for a lab
func (sm *ServiceManager) SetupLabServices(ctx *interfaces.SetupContext) error {
	services := sm.registry.GetAllServices()

	for _, service := range services {
		if err := service.ExecuteSetup(ctx); err != nil {
			return err
		}
	}

	return nil
}

// CleanupLabServices cleans up only the services that were used for a lab
func (sm *ServiceManager) CleanupLabServices(ctx *interfaces.CleanupContext) error {
	// Get the lab to check which services were used
	if ctx.Lab == nil {
		return fmt.Errorf("lab is required for cleanup")
	}

	fmt.Printf("Starting cleanup for lab %s (ID: %s)\n", ctx.Lab.Name, ctx.LabID)
	fmt.Printf("Lab used services: %v\n", ctx.Lab.UsedServices)

	// If no used services are tracked, clean up all services (backward compatibility)
	if len(ctx.Lab.UsedServices) == 0 {
		fmt.Printf("No used services tracked, cleaning up all registered services (backward compatibility)\n")
		services := sm.registry.GetAllServices()
		for serviceName, service := range services {
			fmt.Printf("Cleaning up service: %s\n", serviceName)
			if err := service.ExecuteCleanup(ctx); err != nil {
				fmt.Printf("Error cleaning up service %s: %v\n", serviceName, err)
				return err
			}
		}
		return nil
	}

	// Only clean up services that were actually used for this lab
	fmt.Printf("Cleaning up only services that were used for this lab\n")
	for _, serviceConfigID := range ctx.Lab.UsedServices {
		// First try to get service by Service Config ID
		service, exists := sm.GetServiceByConfigID(serviceConfigID)
		if !exists {
			// Fallback to registry lookup (for backward compatibility)
			service, exists = sm.registry.GetService(serviceConfigID)
			if !exists {
				// Log warning but continue with other services
				fmt.Printf("Warning: Service %s not found in registry during cleanup for lab %s\n", serviceConfigID, ctx.LabID)
				continue
			}
		}

		fmt.Printf("Cleaning up service: %s (config ID: %s)\n", service.GetName(), serviceConfigID)
		if err := service.ExecuteCleanup(ctx); err != nil {
			fmt.Printf("Error cleaning up service %s: %v\n", serviceConfigID, err)
			return err
		}
	}

	fmt.Printf("Cleanup completed successfully for lab %s\n", ctx.LabID)
	return nil
}
