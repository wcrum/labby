package models

import (
	"time"
)

// LabTemplate represents a lab template definition
type LabTemplate struct {
	Name               string             `yaml:"name" json:"name"`
	ID                 string             `yaml:"id" json:"id"`
	Description        string             `yaml:"description" json:"description"`
	ExpirationDuration string             `yaml:"expiration_duration" json:"expiration_duration"`
	Owner              string             `yaml:"owner" json:"owner"`
	CreatedAt          time.Time          `yaml:"created_at" json:"created_at"`
	Services           []ServiceReference `yaml:"services" json:"services"`
}

// ServiceReference represents a reference to a preconfigured service
type ServiceReference struct {
	Name        string `yaml:"name" json:"name"`
	ServiceID   string `yaml:"service_id" json:"service_id"` // Reference to ServiceConfig
	Description string `yaml:"description" json:"description"`
	Type        string `yaml:"type" json:"type,omitempty"` // Service type (enriched from ServiceConfig)
	Logo        string `yaml:"logo" json:"logo,omitempty"` // Service logo (enriched from ServiceConfig)
}

// ServiceTemplate represents a service configuration in a lab template (legacy, kept for backward compatibility)
type ServiceTemplate struct {
	Name        string            `yaml:"name" json:"name"`
	Type        string            `yaml:"type" json:"type"`
	Description string            `yaml:"description" json:"description"`
	Config      map[string]string `yaml:"config" json:"config"`
}

// LabTemplateManager manages lab templates
type LabTemplateManager struct {
	templates map[string]*LabTemplate
}

// NewLabTemplateManager creates a new lab template manager
func NewLabTemplateManager() *LabTemplateManager {
	return &LabTemplateManager{
		templates: make(map[string]*LabTemplate),
	}
}

// AddTemplate adds a lab template
func (ltm *LabTemplateManager) AddTemplate(template *LabTemplate) {
	ltm.templates[template.ID] = template
}

// GetTemplate retrieves a lab template by ID
func (ltm *LabTemplateManager) GetTemplate(id string) (*LabTemplate, bool) {
	template, exists := ltm.templates[id]
	return template, exists
}

// GetAllTemplates returns all lab templates
func (ltm *LabTemplateManager) GetAllTemplates() []*LabTemplate {
	templates := make([]*LabTemplate, 0, len(ltm.templates))
	for _, template := range ltm.templates {
		templates = append(templates, template)
	}
	return templates
}

// EnrichTemplatesWithServiceTypes enriches all templates with service type information
func (ltm *LabTemplateManager) EnrichTemplatesWithServiceTypes(repo interface {
	GetServiceConfigByID(id string) (*ServiceConfig, error)
}) {
	for _, template := range ltm.templates {
		for i := range template.Services {
			serviceRef := &template.Services[i]
			if serviceConfig, err := repo.GetServiceConfigByID(serviceRef.ServiceID); err == nil {
				serviceRef.Type = serviceConfig.Type
				serviceRef.Logo = serviceConfig.Logo
			}
		}
	}
}
