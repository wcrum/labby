package lab

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/wcrum/labby/internal/models"

	"gopkg.in/yaml.v3"
)

// ServiceConfigLoader loads service configurations from files
type ServiceConfigLoader struct {
	serviceConfigManager *models.ServiceConfigManager
}

// NewServiceConfigLoader creates a new service configuration loader
func NewServiceConfigLoader(serviceConfigManager *models.ServiceConfigManager) *ServiceConfigLoader {
	return &ServiceConfigLoader{
		serviceConfigManager: serviceConfigManager,
	}
}

// LoadServiceConfigsFromDirectory loads service configurations from a directory
func (scl *ServiceConfigLoader) LoadServiceConfigsFromDirectory(dirPath string) error {
	return filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		// Skip template files and limits file
		if filepath.Base(path) == "templates" || filepath.Base(path) == "limits.yaml" {
			return nil
		}

		return scl.LoadServiceConfigFromFile(path)
	})
}

// LoadServiceConfigFromFile loads a service configuration from a file
func (scl *ServiceConfigLoader) LoadServiceConfigFromFile(filePath string) error {
	fmt.Printf("ServiceConfigLoader.LoadServiceConfigFromFile: Loading from %s\n", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("ServiceConfigLoader.LoadServiceConfigFromFile: Failed to read file %s: %v\n", filePath, err)
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var serviceConfig models.ServiceConfig
	if err := yaml.Unmarshal(data, &serviceConfig); err != nil {
		fmt.Printf("ServiceConfigLoader.LoadServiceConfigFromFile: Failed to unmarshal YAML from %s: %v\n", filePath, err)
		return fmt.Errorf("failed to unmarshal YAML from %s: %w", filePath, err)
	}

	// Set default values if not provided
	if serviceConfig.CreatedAt.IsZero() {
		serviceConfig.CreatedAt = time.Now()
	}
	if serviceConfig.UpdatedAt.IsZero() {
		serviceConfig.UpdatedAt = time.Now()
	}

	fmt.Printf("ServiceConfigLoader.LoadServiceConfigFromFile: Loaded service config %s (ID: %s, Type: %s, Active: %v)\n",
		serviceConfig.Name, serviceConfig.ID, serviceConfig.Type, serviceConfig.IsActive)
	fmt.Printf("ServiceConfigLoader.LoadServiceConfigFromFile: Debug - IsActive field value: %v, type: %T\n",
		serviceConfig.IsActive, serviceConfig.IsActive)

	// Validate service configuration
	if err := scl.validateServiceConfig(&serviceConfig); err != nil {
		fmt.Printf("ServiceConfigLoader.LoadServiceConfigFromFile: Validation failed for %s: %v\n", filePath, err)
		return fmt.Errorf("invalid service config in %s: %w", filePath, err)
	}

	scl.serviceConfigManager.AddServiceConfig(&serviceConfig)
	fmt.Printf("ServiceConfigLoader.LoadServiceConfigFromFile: Successfully added service config %s\n", serviceConfig.ID)
	return nil
}

// validateServiceConfig validates a service configuration
func (scl *ServiceConfigLoader) validateServiceConfig(config *models.ServiceConfig) error {
	if config.ID == "" {
		return fmt.Errorf("service config ID is required")
	}

	if config.Name == "" {
		return fmt.Errorf("service config name is required")
	}

	if config.Type == "" {
		return fmt.Errorf("service config type is required")
	}

	// Validate service type
	switch config.Type {
	case "palette_project", "proxmox_user", "palette_tenant":
		// Valid service types
	default:
		return fmt.Errorf("unsupported service type: %s", config.Type)
	}

	return nil
}

// LoadServiceLimitsFromDirectory loads service limits from a directory
func (scl *ServiceConfigLoader) LoadServiceLimitsFromDirectory(dirPath string) error {
	return filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		// Only process limit files
		if filepath.Base(path) != "limits.yaml" {
			return nil
		}

		return scl.LoadServiceLimitsFromFile(path)
	})
}

// LoadServiceLimitsFromFile loads service limits from a file
func (scl *ServiceConfigLoader) LoadServiceLimitsFromFile(filePath string) error {
	fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Loading from %s\n", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Failed to read file %s: %v\n", filePath, err)
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Raw YAML content:\n%s\n", string(data))

	var limits []models.ServiceLimit
	if err := yaml.Unmarshal(data, &limits); err != nil {
		fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Failed to unmarshal YAML from %s: %v\n", filePath, err)
		return fmt.Errorf("failed to unmarshal YAML from %s: %w", filePath, err)
	}

	fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Loaded %d service limits\n", len(limits))

	for i := range limits {
		// Set default values if not provided
		if limits[i].CreatedAt.IsZero() {
			limits[i].CreatedAt = time.Now()
		}
		if limits[i].UpdatedAt.IsZero() {
			limits[i].UpdatedAt = time.Now()
		}

		fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Processing limit %d: ServiceID=%s, MaxLabs=%d, Active=%v\n",
			i, limits[i].ServiceID, limits[i].MaxLabs, limits[i].IsActive)
		fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Debug - IsActive field value: %v, type: %T\n",
			limits[i].IsActive, limits[i].IsActive)

		// Validate service limit
		if err := scl.validateServiceLimit(&limits[i]); err != nil {
			fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Validation failed for limit %d: %v\n", i, err)
			return fmt.Errorf("invalid service limit at index %d: %w", i, err)
		}

		scl.serviceConfigManager.AddServiceLimit(&limits[i])
		fmt.Printf("ServiceConfigLoader.LoadServiceLimitsFromFile: Successfully added limit for service %s\n", limits[i].ServiceID)
	}

	return nil
}

// validateServiceLimit validates a service limit
func (scl *ServiceConfigLoader) validateServiceLimit(limit *models.ServiceLimit) error {
	fmt.Printf("validateServiceLimit: Validating limit ID=%s, ServiceID=%s, MaxLabs=%d, MaxDuration=%d, Active=%v\n",
		limit.ID, limit.ServiceID, limit.MaxLabs, limit.MaxDuration, limit.IsActive)

	if limit.ID == "" {
		return fmt.Errorf("service limit ID is required")
	}

	if limit.ServiceID == "" {
		return fmt.Errorf("service limit service_id is required")
	}

	if limit.MaxLabs <= 0 {
		return fmt.Errorf("service limit max_labs must be greater than 0")
	}

	if limit.MaxDuration <= 0 {
		return fmt.Errorf("service limit max_duration must be greater than 0")
	}

	fmt.Printf("validateServiceLimit: Limit validation passed for %s\n", limit.ID)
	return nil
}
