package services

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wcrum/labby/internal/interfaces"
)

// TerraformCloudService handles setup and cleanup for Terraform Cloud workspaces
type TerraformCloudService struct {
	host         string
	apiToken     string
	organization string
	workspaceID  string
	uploadURL    string
	// Terraform configuration settings
	sourceDirectory string
	agentPoolID     string
	executionMode   string
	variables       map[string]string
	sensitiveVars   map[string]string
}

// Global VLAN tag tracking (in a real production environment, this should be in a database)
var (
	vlanTagMutex sync.Mutex
	usedVlanTags = make(map[int]bool)
)

// NewTerraformCloudService creates a new Terraform Cloud service instance
func NewTerraformCloudService() *TerraformCloudService {
	return &TerraformCloudService{
		host:          os.Getenv("TF_CLOUD_HOST"),
		apiToken:      os.Getenv("TF_CLOUD_API_TOKEN"),
		organization:  os.Getenv("TF_CLOUD_ORGANIZATION"),
		variables:     make(map[string]string),
		sensitiveVars: make(map[string]string),
	}
}

// generateUniqueVlanTag generates a unique VLAN tag in the specified range
func (v *TerraformCloudService) generateUniqueVlanTag(min, max int) string {
	vlanTagMutex.Lock()
	defer vlanTagMutex.Unlock()

	// Log current usage
	usedCount := len(usedVlanTags)
	fmt.Printf("Current VLAN tag usage: %d tags in use\n", usedCount)

	// Try to find an available VLAN tag
	for vlanTag := min; vlanTag <= max; vlanTag++ {
		if !usedVlanTags[vlanTag] {
			usedVlanTags[vlanTag] = true
			fmt.Printf("Generated unique VLAN tag: %d (range: %d-%d, total used: %d)\n", vlanTag, min, max, usedCount+1)
			return fmt.Sprintf("%d", vlanTag)
		}
	}

	// If no VLAN tags are available, return an error
	// In a production environment, you might want to implement a cleanup mechanism
	// or expand the range dynamically
	fmt.Printf("Warning: No available VLAN tags in range %d-%d, reusing first available\n", min, max)

	// For now, find the first available tag by clearing old entries
	// This is a simple approach - in production, you'd want proper cleanup
	for vlanTag := min; vlanTag <= max; vlanTag++ {
		usedVlanTags[vlanTag] = true
		fmt.Printf("Reusing VLAN tag: %d (range: %d-%d)\n", vlanTag, min, max)
		return fmt.Sprintf("%d", vlanTag)
	}

	// Fallback to min value if all else fails
	return fmt.Sprintf("%d", min)
}

// processTemplateString processes template strings like "${unique_integer(3100,3149)}" and "${lab_uuid}"
func (v *TerraformCloudService) processTemplateString(value string, labID string) string {
	// Check for template patterns
	if strings.Contains(value, "${") && strings.Contains(value, "}") {
		fmt.Printf("TerraformCloudService: Processing template string: %s\n", value)

		// Process unique_integer template
		if strings.Contains(value, "${unique_integer(") {
			// Extract the range from "${unique_integer(3100,3149)}"
			start := strings.Index(value, "(") + 1
			end := strings.Index(value, ")")
			if start > 0 && end > start {
				rangeStr := value[start:end]
				parts := strings.Split(rangeStr, ",")
				if len(parts) == 2 {
					var min, max int
					if _, err := fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &min); err == nil {
						if _, err := fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &max); err == nil {
							result := v.generateUniqueVlanTag(min, max)
							fmt.Printf("TerraformCloudService: Generated unique integer: %s (range: %d-%d)\n", result, min, max)
							return result
						}
					}
				}
			}
		}

		// Process lab_uuid template
		if strings.Contains(value, "${lab_uuid}") {
			result := strings.ReplaceAll(value, "${lab_uuid}", labID)
			fmt.Printf("TerraformCloudService: Replaced lab_uuid with: %s\n", labID)
			return result
		}

	}

	return value
}

// ConfigureFromServiceConfig configures the service from a service configuration
func (v *TerraformCloudService) ConfigureFromServiceConfig(config map[string]string, labID string) {
	// Set basic configuration
	if host, ok := config["host"]; ok {
		v.host = host
	}
	if apiToken, ok := config["api_token"]; ok {
		v.apiToken = apiToken
	}
	if organization, ok := config["organization"]; ok {
		v.organization = organization
	}

	// Set source directory
	if sourceDir, ok := config["source_directory"]; ok {
		v.sourceDirectory = sourceDir
	}

	// Set agent pool ID and execution mode
	if agentPoolID, ok := config["agent_pool_id"]; ok {
		v.agentPoolID = agentPoolID
		fmt.Printf("TerraformCloudService: Agent pool ID set to: %s\n", agentPoolID)
	}
	if executionMode, ok := config["execution_mode"]; ok {
		v.executionMode = executionMode
		fmt.Printf("TerraformCloudService: Execution mode set to: %s\n", executionMode)
	}

	// Extract variables from flat YAML structure
	// Define which keys should be treated as Terraform variables
	terraformVariableKeys := []string{
		"pm_api_url",
		"pm_node",
		"template_name",
		"storage_pool",
		"network_bridge",
		"vm_user",
		"vm_password",
		"ubuntu_iso",
		"resource_pool",
		"vlan_tag",
		"lab_id",
		// Add any other variables that should be passed to Terraform
	}

	// Extract regular variables
	for key, value := range config {
		for _, varKey := range terraformVariableKeys {
			if key == varKey {
				// Process template strings for all variables
				processedValue := v.processTemplateString(value, labID)
				v.variables[key] = processedValue
				if value != processedValue {
					fmt.Printf("TerraformCloudService: Processed variable %s: '%s' -> '%s'\n", key, value, processedValue)
				}
				break
			}
		}
	}

	// Calculate IP octet from VLAN tag if both exist
	if vlanTagStr, exists := v.variables["vlan_tag"]; exists {
		var vlanTag int
		if _, err := fmt.Sscanf(vlanTagStr, "%d", &vlanTag); err == nil {
			// Extract last 3 digits of VLAN tag
			ipOctet := vlanTag % 1000
			v.variables["ip_octet"] = fmt.Sprintf("%d", ipOctet)
			fmt.Printf("TerraformCloudService: Calculated IP octet %d from VLAN tag %s\n", ipOctet, vlanTagStr)
		}
	}

	// Extract sensitive variables
	sensitiveVariableKeys := []string{
		"pm_api_token_id",
		"pm_api_token_secret",
		"ssh_key",
	}

	for key, value := range config {
		for _, varKey := range sensitiveVariableKeys {
			if key == varKey {
				// Process template strings for sensitive variables too
				processedValue := v.processTemplateString(value, labID)
				v.sensitiveVars[key] = processedValue
				break
			}
		}
	}

	// Also check for nested structure for backward compatibility
	for key, value := range config {
		if strings.HasPrefix(key, "terraform_config.variables.") {
			varName := strings.TrimPrefix(key, "terraform_config.variables.")
			v.variables[varName] = value
		}
		if strings.HasPrefix(key, "terraform_config.sensitive_variables.") {
			varName := strings.TrimPrefix(key, "terraform_config.sensitive_variables.")
			v.sensitiveVars[varName] = value
		}
	}

	// Debug logging
	fmt.Printf("TerraformCloudService: Extracted %d regular variables: %v\n", len(v.variables), v.variables)
	fmt.Printf("TerraformCloudService: Extracted %d sensitive variables: %v\n", len(v.sensitiveVars), v.sensitiveVars)

	// Special logging for vlan_tag
	if vlanTag, exists := v.variables["vlan_tag"]; exists {
		fmt.Printf("TerraformCloudService: VLAN tag set to: %s\n", vlanTag)
	}
}

// GetName returns the service name
func (v *TerraformCloudService) GetName() string {
	return "terraform_cloud"
}

// GetDescription returns the service description
func (v *TerraformCloudService) GetDescription() string {
	return "Terraform Cloud workspace for infrastructure provisioning"
}

// GetRequiredParams returns the required parameters for this service
func (v *TerraformCloudService) GetRequiredParams() []string {
	return []string{"TF_CLOUD_HOST", "TF_CLOUD_API_TOKEN", "TF_CLOUD_ORGANIZATION"}
}

// ReleaseVlanTag releases a VLAN tag back to the pool
func (v *TerraformCloudService) ReleaseVlanTag(vlanTag string) {
	vlanTagMutex.Lock()
	defer vlanTagMutex.Unlock()

	// Convert string to int for map lookup
	var vlanTagInt int
	if _, err := fmt.Sscanf(vlanTag, "%d", &vlanTagInt); err == nil {
		delete(usedVlanTags, vlanTagInt)
		fmt.Printf("Released VLAN tag: %s\n", vlanTag)
	}
}

// GetVlanTagUsage returns the current VLAN tag usage statistics
func (v *TerraformCloudService) GetVlanTagUsage() map[int]bool {
	vlanTagMutex.Lock()
	defer vlanTagMutex.Unlock()

	// Create a copy of the map to avoid race conditions
	usage := make(map[int]bool)
	for k, v := range usedVlanTags {
		usage[k] = v
	}
	return usage
}

// Name returns the service name (implements Setup interface)
func (v *TerraformCloudService) Name() string {
	return v.GetName()
}

// ExecuteSetup sets up Terraform Cloud workspace and adds credentials
func (v *TerraformCloudService) ExecuteSetup(ctx *interfaces.SetupContext) error {
	// Update progress: Creating Workspace
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating Workspace", "running", "Creating workspace in Terraform Cloud...")
	}

	if v.host == "" || v.apiToken == "" || v.organization == "" {
		err := fmt.Errorf("TF_CLOUD_HOST, TF_CLOUD_API_TOKEN, and TF_CLOUD_ORGANIZATION environment variables are required")
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating Workspace", "failed", err.Error())
		}
		return err
	}

	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	fmt.Printf("Setting up Terraform Cloud workspace for lab %s...\n", ctx.LabName)

	// Create workspace
	workspaceID, err := v.createWorkspace(ctx, shortID)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating Workspace", "failed", fmt.Sprintf("Failed to create workspace: %v", err))
		}
		return err
	}

	v.workspaceID = workspaceID

	// Store workspace ID in lab data
	if ctx.Lab != nil {
		if ctx.Lab.ServiceData == nil {
			ctx.Lab.ServiceData = make(map[string]string)
		}
		ctx.Lab.ServiceData["terraform_cloud_workspace_id"] = workspaceID
	}

	// Update progress: Workspace Created
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating Workspace", "completed", "Workspace created successfully")
	}

	// Upload Terraform configuration if provided
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Uploading Configuration", "running", "Uploading Terraform configuration...")
	}

	// Load Terraform configuration from spacewalk directory or use default
	configFiles, err := v.LoadTerraformConfiguration(ctx)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Loading Configuration", "failed", fmt.Sprintf("Failed to load configuration: %v", err))
		}
		return err
	}

	if err := v.uploadConfiguration(workspaceID, configFiles); err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Uploading Configuration", "failed", fmt.Sprintf("Failed to upload configuration: %v", err))
		}
		return err
	}

	// Set workspace variables from service configuration
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Setting Variables", "running", "Setting workspace variables...")
	}

	if err := v.SetWorkspaceVariables(workspaceID); err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Setting Variables", "failed", fmt.Sprintf("Failed to set variables: %v", err))
		}
		return err
	}

	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Setting Variables", "completed", "Workspace variables set successfully")
	}

	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Uploading Configuration", "completed", "Configuration uploaded successfully")
	}

	// Trigger a Terraform run
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Triggering Run", "running", "Triggering Terraform apply...")
	}

	runID, err := v.triggerRun(workspaceID, "Initial lab setup")
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Triggering Run", "failed", fmt.Sprintf("Failed to trigger run: %v", err))
		}
		return err
	}

	// Store run ID in lab data
	if ctx.Lab != nil {
		ctx.Lab.ServiceData["terraform_cloud_run_id"] = runID
	}

	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Triggering Run", "completed", "Terraform run triggered successfully")
	}

	// Add credentials
	workspaceURL := fmt.Sprintf("%s/app/%s/workspaces/%s", v.host, v.organization, workspaceID)

	credential := &interfaces.Credential{
		ID:        uuid.New().String(),
		LabID:     ctx.LabID,
		Label:     "Terraform Cloud Workspace",
		Username:  "API Token",
		Password:  v.apiToken,
		URL:       workspaceURL,
		ExpiresAt: time.Now().Add(time.Duration(ctx.Duration) * time.Minute),
		Notes:     fmt.Sprintf("Workspace ID: %s\nOrganization: %s", workspaceID, v.organization),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ctx.AddCredential(credential); err != nil {
		fmt.Printf("Warning: Failed to add credential: %v\n", err)
	}

	fmt.Printf("Terraform Cloud workspace setup completed for lab %s\n", ctx.LabName)
	return nil
}

// ExecuteCleanup cleans up Terraform Cloud workspace
func (v *TerraformCloudService) ExecuteCleanup(ctx *interfaces.CleanupContext) error {
	fmt.Printf("Cleaning up Terraform Cloud workspace for lab %s...\n", ctx.LabID)

	// Get workspace ID from lab data
	var workspaceID string
	if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
		if id, exists := ctx.Lab.ServiceData["terraform_cloud_workspace_id"]; exists {
			workspaceID = id
		}
	}

	if workspaceID == "" {
		fmt.Printf("Warning: No workspace ID found for lab %s\n", ctx.LabID)
		return nil
	}

	// Delete workspace
	if err := v.deleteWorkspace(workspaceID); err != nil {
		fmt.Printf("Warning: Failed to delete workspace %s: %v\n", workspaceID, err)
		return err
	}

	fmt.Printf("Terraform Cloud workspace cleanup completed for lab %s\n", ctx.LabID)
	return nil
}

// createWorkspace creates a new Terraform Cloud workspace
func (v *TerraformCloudService) createWorkspace(ctx *interfaces.SetupContext, shortID string) (string, error) {
	// Validate required configuration
	if v.executionMode == "" {
		return "", fmt.Errorf("execution_mode is required but not configured")
	}
	if v.agentPoolID == "" {
		return "", fmt.Errorf("agent_pool_id is required but not configured")
	}

	workspaceName := fmt.Sprintf("lab-%s", shortID)

	// Prepare workspace data
	workspaceData := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "workspaces",
			"attributes": map[string]interface{}{
				"name":                  workspaceName,
				"description":           fmt.Sprintf("Lab workspace for %s", ctx.LabName),
				"auto-apply":            false,
				"file-triggers-enabled": true,
				"terraform-version":     "1.5.0",
				"execution-mode":        v.executionMode,
				"agent-pool-id":         v.agentPoolID,
			},
		},
	}

	jsonData, err := json.Marshal(workspaceData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workspace data: %v", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/v2/organizations/%s/workspaces", v.host, v.organization)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+v.apiToken)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create workspace: %s - %s", resp.Status, string(body))
	}

	// Parse response to get workspace ID
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	workspaceID, ok := data["id"].(string)
	if !ok {
		return "", fmt.Errorf("workspace ID not found in response")
	}

	fmt.Printf("Created Terraform Cloud workspace: %s (ID: %s)\n", workspaceName, workspaceID)
	return workspaceID, nil
}

// deleteWorkspace deletes a Terraform Cloud workspace
func (v *TerraformCloudService) deleteWorkspace(workspaceID string) error {
	url := fmt.Sprintf("%s/api/v2/workspaces/%s", v.host, workspaceID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+v.apiToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make delete request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete workspace: %s - %s", resp.Status, string(body))
	}

	fmt.Printf("Deleted Terraform Cloud workspace: %s\n", workspaceID)
	return nil
}

// uploadConfiguration uploads Terraform configuration to the workspace
func (v *TerraformCloudService) uploadConfiguration(workspaceID string, configFiles map[string]string) error {
	// Create a configuration version
	_, err := v.createConfigurationVersion(workspaceID)
	if err != nil {
		return fmt.Errorf("failed to create configuration version: %v", err)
	}

	// Create a temporary directory to store files
	tempDir, err := os.MkdirTemp("", "terraform-config-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Validate that we have at least one configuration file
	if len(configFiles) == 0 {
		return fmt.Errorf("no configuration files to upload")
	}

	// Write all configuration files to temp directory
	for fileName, content := range configFiles {
		// Validate file name
		if fileName == "" {
			return fmt.Errorf("empty file name not allowed")
		}

		// Validate content
		if len(content) == 0 {
			fmt.Printf("Warning: Empty file content for %s\n", fileName)
		}

		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %v", fileName, err)
		}
		fmt.Printf("Wrote configuration file: %s (%d bytes)\n", fileName, len(content))
	}

	// Create a tar.gz file containing all configuration files
	tarGzPath := filepath.Join(tempDir, "config.tar.gz")
	if err := v.createTarGzFile(tempDir, tarGzPath); err != nil {
		return fmt.Errorf("failed to create tar.gz file: %v", err)
	}

	// Verify tar.gz file was created and has content
	tarGzInfo, err := os.Stat(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to stat tar.gz file: %v", err)
	}
	if tarGzInfo.Size() == 0 {
		return fmt.Errorf("created tar.gz file is empty")
	}
	fmt.Printf("Created tar.gz file: %s (%d bytes)\n", tarGzPath, tarGzInfo.Size())

	// Upload the tar.gz file
	if err := v.uploadTarGzFile(tarGzPath); err != nil {
		return fmt.Errorf("failed to upload tar.gz file: %v", err)
	}

	fmt.Printf("Successfully uploaded %d configuration files to workspace %s\n", len(configFiles), workspaceID)
	return nil
}

// createConfigurationVersion creates a new configuration version for the workspace
func (v *TerraformCloudService) createConfigurationVersion(workspaceID string) (string, error) {
	configData := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "configuration-versions",
			"attributes": map[string]interface{}{
				"auto-queue-runs": false,
			},
		},
	}

	jsonData, err := json.Marshal(configData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal configuration version data: %v", err)
	}

	url := fmt.Sprintf("%s/api/v2/workspaces/%s/configuration-versions", v.host, workspaceID)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create configuration version request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+v.apiToken)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make configuration version request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read configuration version response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create configuration version: %s - %s", resp.Status, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse configuration version response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid configuration version response format")
	}

	configVersionID, ok := data["id"].(string)
	if !ok {
		return "", fmt.Errorf("configuration version ID not found in response")
	}

	// Get upload URL
	attributes, ok := data["attributes"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("configuration version attributes not found")
	}

	uploadURL, ok := attributes["upload-url"].(string)
	if !ok {
		return "", fmt.Errorf("upload URL not found in configuration version")
	}

	// Store upload URL in the service for later use
	v.uploadURL = uploadURL

	return configVersionID, nil
}

// createZipFile creates a zip file containing all configuration files
func (v *TerraformCloudService) createZipFile(sourceDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)

	// Walk through the source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the zip file itself
		if info.IsDir() || filepath.Base(path) == "config.zip" {
			return nil
		}

		// Create a relative path for the zip file
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Create zip file entry
		zipEntry, err := zipWriter.Create(relPath)
		if err != nil {
			return fmt.Errorf("failed to create zip entry: %v", err)
		}

		// Read and write file content
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", path, err)
		}

		_, err = zipEntry.Write(fileContent)
		if err != nil {
			return fmt.Errorf("failed to write to zip entry: %v", err)
		}

		fmt.Printf("Added to zip: %s (%d bytes)\n", relPath, len(fileContent))
		return nil
	})

	if err != nil {
		zipWriter.Close()
		return fmt.Errorf("failed to walk directory: %v", err)
	}

	// Close the zip writer to finalize the zip file
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %v", err)
	}

	return nil
}

// createTarGzFile creates a tar.gz file containing all configuration files
func (v *TerraformCloudService) createTarGzFile(sourceDir, tarGzPath string) error {
	tarGzFile, err := os.Create(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to create tar.gz file: %v", err)
	}
	defer tarGzFile.Close()

	gzipWriter := gzip.NewWriter(tarGzFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk through the source directory
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the tar.gz file itself
		if info.IsDir() || filepath.Base(path) == "config.tar.gz" {
			return nil
		}

		// Create a relative path for the tar file
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		// Read file content
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %v", path, err)
		}

		// Create tar header
		header := &tar.Header{
			Name: relPath,
			Mode: int64(info.Mode()),
			Size: int64(len(fileContent)),
		}

		// Write tar header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %v", relPath, err)
		}

		// Write file content
		if _, err := tarWriter.Write(fileContent); err != nil {
			return fmt.Errorf("failed to write to tar entry: %v", err)
		}

		fmt.Printf("Added to tar.gz: %s (%d bytes)\n", relPath, len(fileContent))
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %v", err)
	}

	return nil
}

// validateZipFile validates that the zip file is properly formatted
func (v *TerraformCloudService) validateZipFile(zipPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file for validation: %v", err)
	}
	defer reader.Close()

	fileCount := 0
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			fileCount++
			fmt.Printf("Validated zip entry: %s (%d bytes)\n", file.Name, file.UncompressedSize64)
		}
	}

	if fileCount == 0 {
		return fmt.Errorf("zip file contains no files")
	}

	fmt.Printf("Zip file validation passed: %d files found\n", fileCount)
	return nil
}

// uploadZipFile uploads the zip file to Terraform Cloud
func (v *TerraformCloudService) uploadZipFile(zipPath string) error {
	if v.uploadURL == "" {
		return fmt.Errorf("upload URL not available")
	}

	// Read the zip file
	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		return fmt.Errorf("failed to read zip file: %v", err)
	}

	fmt.Printf("Uploading zip file: %d bytes\n", len(zipData))

	// Create request with zip data
	req, err := http.NewRequest("PUT", v.uploadURL, bytes.NewReader(zipData))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 120 * time.Second} // Increased timeout for large files
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload zip file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload zip file: %s - %s", resp.Status, string(body))
	}

	fmt.Printf("Successfully uploaded zip file to Terraform Cloud\n")
	return nil
}

// uploadTarGzFile uploads the tar.gz file to Terraform Cloud
func (v *TerraformCloudService) uploadTarGzFile(tarGzPath string) error {
	if v.uploadURL == "" {
		return fmt.Errorf("upload URL not available")
	}

	// Read the tar.gz file
	tarGzData, err := os.ReadFile(tarGzPath)
	if err != nil {
		return fmt.Errorf("failed to read tar.gz file: %v", err)
	}

	fmt.Printf("Uploading tar.gz file: %d bytes\n", len(tarGzData))

	// Create request with tar.gz data
	req, err := http.NewRequest("PUT", v.uploadURL, bytes.NewReader(tarGzData))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 120 * time.Second} // Increased timeout for large files
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload tar.gz file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload tar.gz file: %s - %s", resp.Status, string(body))
	}

	fmt.Printf("Successfully uploaded tar.gz file to Terraform Cloud\n")
	return nil
}

// triggerRun triggers a Terraform run in the workspace
func (v *TerraformCloudService) triggerRun(workspaceID, message string) (string, error) {
	runData := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "runs",
			"attributes": map[string]interface{}{
				"message": message,
			},
			"relationships": map[string]interface{}{
				"workspace": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "workspaces",
						"id":   workspaceID,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(runData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal run data: %v", err)
	}

	url := fmt.Sprintf("%s/api/v2/runs", v.host)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create run request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+v.apiToken)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make run request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read run response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to trigger run: %s - %s", resp.Status, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse run response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid run response format")
	}

	runID, ok := data["id"].(string)
	if !ok {
		return "", fmt.Errorf("run ID not found in response")
	}

	fmt.Printf("Triggered Terraform run: %s\n", runID)
	return runID, nil
}

// getRunStatus gets the status of a Terraform run
func (v *TerraformCloudService) getRunStatus(runID string) (string, error) {
	url := fmt.Sprintf("%s/api/v2/runs/%s", v.host, runID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create status request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+v.apiToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make status request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read status response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get run status: %s - %s", resp.Status, string(body))
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse status response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid status response format")
	}

	attributes, ok := data["attributes"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("status attributes not found")
	}

	status, ok := attributes["status"].(string)
	if !ok {
		return "", fmt.Errorf("status not found in response")
	}

	return status, nil
}

// UploadCustomConfiguration uploads custom Terraform configuration files to the workspace
func (v *TerraformCloudService) UploadCustomConfiguration(workspaceID string, configFiles map[string]string) error {
	return v.uploadConfiguration(workspaceID, configFiles)
}

// LoadTerraformConfiguration loads Terraform configuration from the spacewalk directory
func (v *TerraformCloudService) LoadTerraformConfiguration(ctx *interfaces.SetupContext) (map[string]string, error) {
	if v.sourceDirectory == "" {
		// Fallback to default configuration
		return v.GetDefaultConfiguration(ctx.LabName), nil
	}

	// Construct the full path to the Terraform configuration
	// The sourceDirectory is relative to the project root, but we're running from the backend directory
	configPath := filepath.Join("..", v.sourceDirectory)

	// Check if the directory exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Terraform configuration directory not found: %s", configPath)
	}

	configFiles := make(map[string]string)

	// Read all .tf files
	tfFiles, err := filepath.Glob(filepath.Join(configPath, "*.tf"))
	if err != nil {
		return nil, fmt.Errorf("failed to find .tf files: %v", err)
	}

	for _, tfFile := range tfFiles {
		content, err := os.ReadFile(tfFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", tfFile, err)
		}

		// Get just the filename
		filename := filepath.Base(tfFile)

		// Templatize the content with lab-specific variables
		templatizedContent := v.templatizeContent(string(content), ctx)
		configFiles[filename] = templatizedContent
	}

	// Read .tfvars file if it exists and templatize it
	tfvarsPath := filepath.Join(configPath, "terraform.tfvars")
	if _, err := os.Stat(tfvarsPath); err == nil {
		content, err := os.ReadFile(tfvarsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read terraform.tfvars: %v", err)
		}

		// Templatize the tfvars content
		templatizedTfvars := v.templatizeContent(string(content), ctx)
		configFiles["terraform.tfvars"] = templatizedTfvars
	}

	// Read versions.tf if it exists
	versionsPath := filepath.Join(configPath, "versions.tf")
	if _, err := os.Stat(versionsPath); err == nil {
		content, err := os.ReadFile(versionsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read versions.tf: %v", err)
		}
		configFiles["versions.tf"] = string(content)
	}

	return configFiles, nil
}

// templatizeContent replaces variables in Terraform configuration with values from service config
func (v *TerraformCloudService) templatizeContent(content string, ctx *interfaces.SetupContext) string {
	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	// Replace common lab variables
	replacements := map[string]string{
		"${lab_id}":    shortID,
		"${lab_name}":  ctx.LabName,
		"${lab_owner}": ctx.OwnerID,
	}

	// Add variables from service configuration
	for key, value := range v.variables {
		replacements["${"+key+"}"] = value
	}

	// Apply replacements
	result := content
	for placeholder, value := range replacements {
		oldResult := result
		result = strings.ReplaceAll(result, placeholder, value)
		if oldResult != result {
			fmt.Printf("Templatized: %s -> %s\n", placeholder, value)
		}
	}

	return result
}

// SetWorkspaceVariables sets variables in the Terraform Cloud workspace
func (v *TerraformCloudService) SetWorkspaceVariables(workspaceID string) error {
	fmt.Printf("Setting %d regular variables in workspace %s\n", len(v.variables), workspaceID)

	// Set regular variables
	for key, value := range v.variables {
		fmt.Printf("Setting variable %s = %s\n", key, value)
		if err := v.setWorkspaceVariable(workspaceID, key, value, false); err != nil {
			return fmt.Errorf("failed to set variable %s: %v", key, err)
		}
	}

	fmt.Printf("Setting %d sensitive variables in workspace %s\n", len(v.sensitiveVars), workspaceID)

	// Set sensitive variables
	for key, value := range v.sensitiveVars {
		fmt.Printf("Setting sensitive variable %s = [REDACTED]\n", key)
		if err := v.setWorkspaceVariable(workspaceID, key, value, true); err != nil {
			return fmt.Errorf("failed to set sensitive variable %s: %v", key, err)
		}
	}

	return nil
}

// setWorkspaceVariable sets a single variable in the Terraform Cloud workspace
func (v *TerraformCloudService) setWorkspaceVariable(workspaceID, key, value string, sensitive bool) error {
	varData := map[string]interface{}{
		"data": map[string]interface{}{
			"type": "vars",
			"attributes": map[string]interface{}{
				"key":       key,
				"value":     value,
				"sensitive": sensitive,
				"category":  "terraform",
			},
		},
	}

	jsonData, err := json.Marshal(varData)
	if err != nil {
		return fmt.Errorf("failed to marshal variable data: %v", err)
	}

	url := fmt.Sprintf("%s/api/v2/workspaces/%s/vars", v.host, workspaceID)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create variable request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+v.apiToken)
	req.Header.Set("Content-Type", "application/vnd.api+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make variable request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set variable %s: %s - %s", key, resp.Status, string(body))
	}

	return nil
}

// GetDefaultConfiguration returns a default Terraform configuration for labs
func (v *TerraformCloudService) GetDefaultConfiguration(labName string) map[string]string {
	return map[string]string{
		"main.tf": `terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = "Lab Instance"
    Lab  = "` + labName + `"
  }
}

output "instance_id" {
  value = aws_instance.example.id
}

output "public_ip" {
  value = aws_instance.example.public_ip
}`,
		"variables.tf": `variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}`,
		"outputs.tf": `output "instance_id" {
  description = "ID of the EC2 instance"
  value       = aws_instance.example.id
}

output "public_ip" {
  description = "Public IP address of the EC2 instance"
  value       = aws_instance.example.public_ip
}`,
	}
}
