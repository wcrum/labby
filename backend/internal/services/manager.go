package services

import (
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

// CleanupLabServices cleans up all services for a lab
func (sm *ServiceManager) CleanupLabServices(ctx *interfaces.CleanupContext) error {
	services := sm.registry.GetAllServices()

	for _, service := range services {
		if err := service.ExecuteCleanup(ctx); err != nil {
			return err
		}
	}

	return nil
}
