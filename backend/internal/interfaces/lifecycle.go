package interfaces

import (
	"context"
	"spectro-lab-backend/internal/models"
	"time"
)

// Credential represents a credential that can be added during setup
type Credential struct {
	ID        string    `json:"id"`
	LabID     string    `json:"lab_id"`
	Label     string    `json:"label"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	URL       string    `json:"url,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SetupContext provides context and utilities for setup operations
type SetupContext struct {
	LabID          string
	LabName        string
	Duration       int // Duration in minutes
	OwnerID        string
	Context        context.Context
	Lab            *models.Lab // Reference to the lab for persistent data storage
	AddCredential  func(credential *Credential) error
	UpdateProgress func(stepName, status, message string) // Function to update progress steps
}

// CleanupContext provides context and utilities for cleanup operations
type CleanupContext struct {
	LabID   string
	Context context.Context
	Lab     *models.Lab // Reference to the lab for accessing stored service data
}

// Setup defines the contract for setup actions
type Setup interface {
	ExecuteSetup(ctx *SetupContext) error
	Name() string
}

// Cleanup defines the contract for cleanup actions
type Cleanup interface {
	ExecuteCleanup(ctx *CleanupContext) error
	Name() string
}

// Lifecycle combines setup and cleanup actions
type Lifecycle interface {
	Setup
	Cleanup
}

// Service represents a service that can be set up and cleaned up
type Service interface {
	Lifecycle
	GetName() string
	GetDescription() string
	GetRequiredParams() []string
}

// ServiceRegistry manages all available services
type ServiceRegistry struct {
	services map[string]Service
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]Service),
	}
}

// RegisterService adds a service to the registry
func (sr *ServiceRegistry) RegisterService(service Service) {
	sr.services[service.GetName()] = service
}

// GetService retrieves a service by name
func (sr *ServiceRegistry) GetService(name string) (Service, bool) {
	service, exists := sr.services[name]
	return service, exists
}

// GetAllServices returns all registered services
func (sr *ServiceRegistry) GetAllServices() map[string]Service {
	return sr.services
}

// GetServiceNames returns a list of all service names
func (sr *ServiceRegistry) GetServiceNames() []string {
	names := make([]string, 0, len(sr.services))
	for name := range sr.services {
		names = append(names, name)
	}
	return names
}
