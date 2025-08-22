package models

import (
	"errors"
	"fmt"
	"sync"
)

// ServiceConfigManager manages service configurations and limits
type ServiceConfigManager struct {
	configs map[string]*ServiceConfig
	limits  map[string]*ServiceLimit
	mu      sync.RWMutex
}

// NewServiceConfigManager creates a new service configuration manager
func NewServiceConfigManager() *ServiceConfigManager {
	return &ServiceConfigManager{
		configs: make(map[string]*ServiceConfig),
		limits:  make(map[string]*ServiceLimit),
	}
}

// AddServiceConfig adds a service configuration
func (scm *ServiceConfigManager) AddServiceConfig(config *ServiceConfig) {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	scm.configs[config.ID] = config
}

// GetServiceConfig retrieves a service configuration by ID
func (scm *ServiceConfigManager) GetServiceConfig(id string) (*ServiceConfig, bool) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()
	config, exists := scm.configs[id]
	return config, exists
}

// GetAllServiceConfigs returns all service configurations
func (scm *ServiceConfigManager) GetAllServiceConfigs() []*ServiceConfig {
	scm.mu.RLock()
	defer scm.mu.RUnlock()
	configs := make([]*ServiceConfig, 0, len(scm.configs))
	for _, config := range scm.configs {
		configs = append(configs, config)
	}
	return configs
}

// GetActiveServiceConfigs returns only active service configurations
func (scm *ServiceConfigManager) GetActiveServiceConfigs() []*ServiceConfig {
	scm.mu.RLock()
	defer scm.mu.RUnlock()
	configs := make([]*ServiceConfig, 0)
	for _, config := range scm.configs {
		if config.IsActive {
			configs = append(configs, config)
		}
	}
	return configs
}

// RemoveServiceConfig removes a service configuration
func (scm *ServiceConfigManager) RemoveServiceConfig(id string) {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	delete(scm.configs, id)
}

// AddServiceLimit adds a service limit
func (scm *ServiceConfigManager) AddServiceLimit(limit *ServiceLimit) {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	scm.limits[limit.ServiceID] = limit
}

// GetServiceLimit retrieves a service limit by service ID
func (scm *ServiceConfigManager) GetServiceLimit(serviceID string) (*ServiceLimit, bool) {
	scm.mu.RLock()
	defer scm.mu.RUnlock()
	limit, exists := scm.limits[serviceID]
	return limit, exists
}

// GetAllServiceLimits returns all service limits
func (scm *ServiceConfigManager) GetAllServiceLimits() []*ServiceLimit {
	scm.mu.RLock()
	defer scm.mu.RUnlock()
	limits := make([]*ServiceLimit, 0, len(scm.limits))
	for _, limit := range scm.limits {
		limits = append(limits, limit)
	}
	return limits
}

// RemoveServiceLimit removes a service limit
func (scm *ServiceConfigManager) RemoveServiceLimit(serviceID string) {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	delete(scm.limits, serviceID)
}

// CheckServiceAvailability checks if a service is available and within limits
func (scm *ServiceConfigManager) CheckServiceAvailability(serviceID string, currentUsage int) error {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	fmt.Printf("CheckServiceAvailability: Checking service %s with current usage %d\n", serviceID, currentUsage)

	// Check if service config exists and is active
	config, exists := scm.configs[serviceID]
	if !exists {
		fmt.Printf("CheckServiceAvailability: Service config not found for %s\n", serviceID)
		return errors.New("service configuration not found")
	}
	fmt.Printf("CheckServiceAvailability: Found service config %s (active: %v)\n", config.Name, config.IsActive)
	fmt.Printf("CheckServiceAvailability: Debug - IsActive field value: %v, type: %T\n", config.IsActive, config.IsActive)

	if !config.IsActive {
		fmt.Printf("CheckServiceAvailability: Service config %s is not active\n", serviceID)
		return errors.New("service configuration is not active")
	}

	// Check if limit exists and is active
	limit, exists := scm.limits[serviceID]
	if !exists {
		fmt.Printf("CheckServiceAvailability: Service limit not found for %s\n", serviceID)
		return errors.New("service limit not found")
	}
	fmt.Printf("CheckServiceAvailability: Found service limit for %s (active: %v, max_labs: %d)\n", serviceID, limit.IsActive, limit.MaxLabs)

	if !limit.IsActive {
		fmt.Printf("CheckServiceAvailability: Service limit %s is not active\n", serviceID)
		return errors.New("service limit is not active")
	}

	// Check if current usage is within limits
	if currentUsage >= limit.MaxLabs {
		fmt.Printf("CheckServiceAvailability: Service %s limit exceeded (usage: %d, limit: %d)\n", serviceID, currentUsage, limit.MaxLabs)
		return errors.New("service limit exceeded")
	}

	fmt.Printf("CheckServiceAvailability: Service %s availability check passed (usage: %d, limit: %d)\n", serviceID, currentUsage, limit.MaxLabs)
	return nil
}

// GetServiceUsage returns the usage information for a service
func (scm *ServiceConfigManager) GetServiceUsage(serviceID string, currentUsage int) *ServiceUsage {
	scm.mu.RLock()
	defer scm.mu.RUnlock()

	limit, exists := scm.limits[serviceID]
	if !exists {
		return &ServiceUsage{
			ServiceID:  serviceID,
			ActiveLabs: currentUsage,
			Limit:      0,
		}
	}

	return &ServiceUsage{
		ServiceID:  serviceID,
		ActiveLabs: currentUsage,
		Limit:      limit.MaxLabs,
	}
}
