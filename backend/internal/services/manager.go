package services

import (
	"fmt"
	"spectro-lab-backend/internal/interfaces"
)

// ServiceManager manages all available services
type ServiceManager struct {
	registry *interfaces.ServiceRegistry
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	registry := interfaces.NewServiceRegistry()

	// Register all available services
	registry.RegisterService(NewPaletteProjectService())
	registry.RegisterService(NewProxmoxUserService())
	registry.RegisterService(NewPaletteTenantService())

	return &ServiceManager{
		registry: registry,
	}
}

// GetRegistry returns the service registry
func (sm *ServiceManager) GetRegistry() *interfaces.ServiceRegistry {
	return sm.registry
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
	for _, serviceName := range ctx.Lab.UsedServices {
		service, exists := sm.registry.GetService(serviceName)
		if !exists {
			// Log warning but continue with other services
			fmt.Printf("Warning: Service %s not found in registry during cleanup for lab %s\n", serviceName, ctx.LabID)
			continue
		}

		fmt.Printf("Cleaning up service: %s\n", serviceName)
		if err := service.ExecuteCleanup(ctx); err != nil {
			fmt.Printf("Error cleaning up service %s: %v\n", serviceName, err)
			return err
		}
	}

	fmt.Printf("Cleanup completed successfully for lab %s\n", ctx.LabID)
	return nil
}
