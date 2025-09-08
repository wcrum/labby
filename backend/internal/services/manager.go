package services

import (
	"fmt"

	"github.com/wcrum/labby/internal/interfaces"
	"github.com/wcrum/labby/internal/models"
)

// ServiceManager manages all available services
type ServiceManager struct {
	registry             *interfaces.ServiceRegistry
	serviceConfigManager *models.ServiceConfigManager
	// Map service types to service instances
	serviceTypeMap map[string]interfaces.Service
}

// NewServiceManager creates a new service manager
func NewServiceManager(serviceConfigManager *models.ServiceConfigManager) *ServiceManager {
	registry := interfaces.NewServiceRegistry()

	// Create service instances
	paletteProjectService := NewPaletteProjectService()
	proxmoxUserService := NewProxmoxUserService()
	paletteTenantService := NewPaletteTenantService()
	terraformCloudService := NewTerraformCloudService()
	guacamoleService := NewGuacamoleService()

	// Register services with their GetName() for backward compatibility
	registry.RegisterService(paletteProjectService)
	registry.RegisterService(proxmoxUserService)
	registry.RegisterService(paletteTenantService)
	registry.RegisterService(terraformCloudService)
	registry.RegisterService(guacamoleService)

	// Create mapping from service types to service instances
	serviceTypeMap := make(map[string]interfaces.Service)

	// Map service types to their corresponding service implementations
	// This mapping is based on the "type" field in service configs
	serviceTypeMap["palette_project"] = paletteProjectService
	serviceTypeMap["proxmox_user"] = proxmoxUserService
	serviceTypeMap["palette_tenant"] = paletteTenantService
	serviceTypeMap["terraform_cloud"] = terraformCloudService
	serviceTypeMap["guacamole"] = guacamoleService

	return &ServiceManager{
		registry:             registry,
		serviceConfigManager: serviceConfigManager,
		serviceTypeMap:       serviceTypeMap,
	}
}

// GetRegistry returns the service registry
func (sm *ServiceManager) GetRegistry() *interfaces.ServiceRegistry {
	return sm.registry
}

// GetServiceByType returns a service by its type
func (sm *ServiceManager) GetServiceByType(serviceType string) (interfaces.Service, bool) {
	service, exists := sm.serviceTypeMap[serviceType]
	return service, exists
}

// GetServiceByConfigID returns a service by looking up the service type from the service config ID
func (sm *ServiceManager) GetServiceByConfigID(configID string) (interfaces.Service, bool) {
	// Get the service config to find the service type
	serviceConfig, exists := sm.serviceConfigManager.GetServiceConfig(configID)
	if !exists {
		return nil, false
	}

	// Get the service by type
	return sm.GetServiceByType(serviceConfig.Type)
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
		// Get service by looking up the service type from the service config
		service, exists := sm.GetServiceByConfigID(serviceConfigID)
		if !exists {
			// Log warning but continue with other services
			fmt.Printf("Warning: Service for config ID %s not found during cleanup for lab %s\n", serviceConfigID, ctx.LabID)
			continue
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
