package lab

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/wcrum/labby/internal/models"

	"gopkg.in/yaml.v2"
)

// TemplateLoader loads lab templates from YAML files
type TemplateLoader struct {
	templateManager *models.LabTemplateManager
}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader(templateManager *models.LabTemplateManager) *TemplateLoader {
	return &TemplateLoader{
		templateManager: templateManager,
	}
}

// LoadTemplatesFromDirectory loads all lab templates from a directory
func (tl *TemplateLoader) LoadTemplatesFromDirectory(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}

		filePath := filepath.Join(dirPath, file.Name())
		if err := tl.LoadTemplateFromFile(filePath); err != nil {
			return fmt.Errorf("failed to load template from %s: %w", filePath, err)
		}
	}

	return nil
}

// LoadTemplateFromFile loads a single lab template from a YAML file
func (tl *TemplateLoader) LoadTemplateFromFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var template models.LabTemplate
	if err := yaml.Unmarshal(data, &template); err != nil {
		return fmt.Errorf("failed to unmarshal YAML from %s: %w", filePath, err)
	}

	// Set default values if not provided
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}

	// Validate template
	if err := tl.validateTemplate(&template); err != nil {
		return fmt.Errorf("invalid template in %s: %w", filePath, err)
	}

	tl.templateManager.AddTemplate(&template)
	return nil
}

// LoadTemplateFromString loads a lab template from a YAML string
func (tl *TemplateLoader) LoadTemplateFromString(yamlData string) error {
	var template models.LabTemplate
	if err := yaml.Unmarshal([]byte(yamlData), &template); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Set default values if not provided
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}

	// Validate template
	if err := tl.validateTemplate(&template); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	tl.templateManager.AddTemplate(&template)
	return nil
}

// validateTemplate validates a lab template
func (tl *TemplateLoader) validateTemplate(template *models.LabTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.ID == "" {
		return fmt.Errorf("template ID is required")
	}

	if template.ExpirationDuration == "" {
		return fmt.Errorf("expiration duration is required")
	}

	// Validate expiration duration format
	if _, err := time.ParseDuration(template.ExpirationDuration); err != nil {
		return fmt.Errorf("invalid expiration duration format: %s", template.ExpirationDuration)
	}

	// Validate services
	for i, service := range template.Services {
		if service.Name == "" {
			return fmt.Errorf("service %d name is required", i)
		}

		if service.ServiceID == "" {
			return fmt.Errorf("service %d service_id is required", i)
		}
	}

	return nil
}

// CreateLabFromTemplate creates a lab instance from a template
func (tl *TemplateLoader) CreateLabFromTemplate(templateID, ownerID string) (*models.Lab, error) {
	fmt.Printf("TemplateLoader.CreateLabFromTemplate: Starting for template %s, owner %s\n", templateID, ownerID)

	template, exists := tl.templateManager.GetTemplate(templateID)
	if !exists {
		fmt.Printf("TemplateLoader.CreateLabFromTemplate: Template %s not found\n", templateID)
		return nil, fmt.Errorf("template not found: %s", templateID)
	}
	fmt.Printf("TemplateLoader.CreateLabFromTemplate: Found template %s with %d services\n", templateID, len(template.Services))

	// Parse duration
	duration, err := time.ParseDuration(template.ExpirationDuration)
	if err != nil {
		fmt.Printf("TemplateLoader.CreateLabFromTemplate: Invalid duration %s: %v\n", template.ExpirationDuration, err)
		return nil, fmt.Errorf("invalid duration in template: %w", err)
	}
	fmt.Printf("TemplateLoader.CreateLabFromTemplate: Parsed duration: %v\n", duration)

	// Create lab
	now := time.Now()

	// Extract service IDs from template for tracking
	usedServices := make([]string, 0, len(template.Services))
	for _, service := range template.Services {
		fmt.Printf("TemplateLoader.CreateLabFromTemplate: Adding service %s (ID: %s) to used services\n", service.Name, service.ServiceID)
		usedServices = append(usedServices, service.ServiceID)
	}

	labID := models.GenerateID()
	labName := models.GenerateLabName()

	fmt.Printf("TemplateLoader.CreateLabFromTemplate: Generated lab ID: %s, name: %s\n", labID, labName)

	lab := &models.Lab{
		ID:           labID,
		Name:         labName,
		Status:       models.LabStatusProvisioning,
		OwnerID:      ownerID,
		StartedAt:    now,
		EndsAt:       now.Add(duration),
		CreatedAt:    now,
		UpdatedAt:    now,
		Credentials:  []models.Credential{},
		TemplateID:   templateID,
		UsedServices: usedServices,
	}

	fmt.Printf("TemplateLoader.CreateLabFromTemplate: Lab created successfully with %d used services\n", len(usedServices))
	return lab, nil
}
