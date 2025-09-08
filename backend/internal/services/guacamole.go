package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/wcrum/labby/internal/interfaces"

	"github.com/sethvargo/go-password/password"
)

// GuacamoleService handles setup and cleanup for Guacamole user accounts
type GuacamoleService struct {
	host          string
	adminUsername string
	adminPassword string
	skipTLSVerify bool
}

// NewGuacamoleService creates a new Guacamole service instance
func NewGuacamoleService() *GuacamoleService {
	return &GuacamoleService{
		host:          os.Getenv("GUACAMOLE_HOST"),
		adminUsername: os.Getenv("GUACAMOLE_ADMIN_USERNAME"),
		adminPassword: os.Getenv("GUACAMOLE_ADMIN_PASSWORD"),
		skipTLSVerify: os.Getenv("GUACAMOLE_SKIP_TLS_VERIFY") == "true",
	}
}

// ConfigureFromServiceConfig configures the service from a service configuration
func (v *GuacamoleService) ConfigureFromServiceConfig(config map[string]string) {
	if host, ok := config["host"]; ok {
		v.host = host
	}
	if adminUsername, ok := config["admin_username"]; ok {
		v.adminUsername = adminUsername
	}
	if adminPassword, ok := config["admin_password"]; ok {
		v.adminPassword = adminPassword
	}
	if skipTLSVerify, ok := config["skip_tls_verify"]; ok {
		v.skipTLSVerify = skipTLSVerify == "true"
	}
}

// GetName returns the service name
func (v *GuacamoleService) GetName() string {
	return "guacamole"
}

// GetDescription returns the service description
func (v *GuacamoleService) GetDescription() string {
	return "Apache Guacamole remote desktop access"
}

// GetRequiredParams returns the required parameters for this service
func (v *GuacamoleService) GetRequiredParams() []string {
	return []string{"GUACAMOLE_HOST", "GUACAMOLE_ADMIN_USERNAME", "GUACAMOLE_ADMIN_PASSWORD"}
}

// Name returns the service name (implements Setup interface)
func (v *GuacamoleService) Name() string {
	return v.GetName()
}

// GuacamoleClient represents a Guacamole API client
type GuacamoleClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// NewGuacamoleClient creates a new Guacamole client
func NewGuacamoleClient(baseURL, username, password string, skipTLSVerify bool) (*GuacamoleClient, error) {
	// Create HTTP client with TLS configuration
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTLSVerify,
			},
		},
	}

	client := &GuacamoleClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}

	// Authenticate and get token
	if err := client.authenticate(username, password); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return client, nil
}

// GuacamoleTokenResponse represents the response from the token endpoint
type GuacamoleTokenResponse struct {
	AuthToken            string   `json:"authToken"`
	Username             string   `json:"username"`
	DataSource           string   `json:"dataSource"`
	AvailableDataSources []string `json:"availableDataSources"`
}

// authenticate performs authentication and gets auth token
func (gc *GuacamoleClient) authenticate(username, password string) error {
	loginURL := fmt.Sprintf("%s/guacamole/api/tokens", gc.baseURL)

	// Create form data
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	fmt.Printf("Authenticating to Guacamole at: %s\n", loginURL)
	fmt.Printf("Using admin user: %s\n", username)

	resp, err := gc.httpClient.PostForm(loginURL, data)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read authentication response: %w", err)
	}

	fmt.Printf("Authentication response status: %d\n", resp.StatusCode)
	fmt.Printf("Authentication response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	var result GuacamoleTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to decode authentication response: %w", err)
	}

	gc.authToken = result.AuthToken

	fmt.Printf("Authentication successful, auth token length: %d\n", len(gc.authToken))

	return nil
}

// GuacamoleUserRequest represents the request to create a Guacamole user
type GuacamoleUserRequest struct {
	Username   string                 `json:"username"`
	Password   string                 `json:"password"`
	Attributes map[string]interface{} `json:"attributes"`
}

// createUser creates a new Guacamole user
func (gc *GuacamoleClient) createUser(username, password string) error {
	createURL := fmt.Sprintf("%s/guacamole/api/session/data/mysql/users", gc.baseURL)

	// Create user request
	userReq := GuacamoleUserRequest{
		Username: username,
		Password: password,
		Attributes: map[string]interface{}{
			"expired":             "",
			"access-window-start": "",
			"access-window-end":   "",
			"valid-from":          "",
			"valid-until":         "",
			"timezone":            nil,
		},
	}

	jsonData, err := json.Marshal(userReq)
	if err != nil {
		return fmt.Errorf("failed to marshal user request: %w", err)
	}

	req, err := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Guacamole-Token", gc.authToken)

	fmt.Printf("Creating Guacamole user: %s\n", username)
	fmt.Printf("Request URL: %s\n", createURL)
	fmt.Printf("Request body: %s\n", string(jsonData))

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("create user request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Create user response status: %d\n", resp.StatusCode)
	fmt.Printf("Create user response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("create user failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

// deleteUser deletes a Guacamole user
func (gc *GuacamoleClient) deleteUser(username string) error {
	deleteURL := fmt.Sprintf("%s/guacamole/api/session/data/mysql/users/%s", gc.baseURL, url.PathEscape(username))

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Guacamole-Token", gc.authToken)

	fmt.Printf("Deleting Guacamole user: %s\n", username)
	fmt.Printf("Request URL: %s\n", deleteURL)

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete user request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("Delete user response status: %d\n", resp.StatusCode)
	fmt.Printf("Delete user response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete user failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ExecuteSetup sets up Guacamole user access and adds credentials
func (v *GuacamoleService) ExecuteSetup(ctx *interfaces.SetupContext) error {
	// Update progress: Connecting to Guacamole
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Connecting to Guacamole", "running", "Connecting to Guacamole instance...")
	}

	if v.host == "" || v.adminUsername == "" || v.adminPassword == "" {
		err := fmt.Errorf("GUACAMOLE_HOST, GUACAMOLE_ADMIN_USERNAME, and GUACAMOLE_ADMIN_PASSWORD environment variables are required")
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Connecting to Guacamole", "failed", err.Error())
		}
		return err
	}

	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	fmt.Printf("Setting up Guacamole user for lab %s...\n", ctx.LabName)

	// Create Guacamole client
	client, err := NewGuacamoleClient(v.host, v.adminUsername, v.adminPassword, v.skipTLSVerify)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Connecting to Guacamole", "failed", fmt.Sprintf("Failed to connect: %v", err))
		}
		return fmt.Errorf("failed to create Guacamole client: %w", err)
	}

	// Update progress: Connecting to Guacamole completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Connecting to Guacamole", "completed", "Successfully connected to Guacamole instance")
	}

	// Update progress: Creating User Account
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating User Account", "running", "Creating Guacamole user account...")
	}

	// Generate username and password for lab user
	labUsername := fmt.Sprintf("lab-%s", shortID)
	labPassword, err := password.Generate(16, 4, 4, false, false)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating User Account", "failed", fmt.Sprintf("Failed to generate password: %v", err))
		}
		return fmt.Errorf("failed to generate password: %w", err)
	}

	// Create user
	fmt.Printf("- Creating user: %s\n", labUsername)
	if err := client.createUser(labUsername, labPassword); err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating User Account", "failed", fmt.Sprintf("Failed to create user: %v", err))
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Printf("  User created successfully\n")

	// Update progress: Creating User Account completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating User Account", "completed", "Guacamole user account created successfully")
	}

	// Store lab-specific data in context for cleanup
	ctx.Context = context.WithValue(ctx.Context, "guacamole_user_username", labUsername)
	ctx.Context = context.WithValue(ctx.Context, "guacamole_user_password", labPassword)

	// Store in lab's ServiceData for persistence
	if ctx.Lab != nil {
		if ctx.Lab.ServiceData == nil {
			ctx.Lab.ServiceData = make(map[string]string)
		}
		ctx.Lab.ServiceData["guacamole_user_username"] = labUsername
		ctx.Lab.ServiceData["guacamole_user_password"] = labPassword
		// Store configuration for cleanup
		ctx.Lab.ServiceData["guacamole_host"] = v.host
		ctx.Lab.ServiceData["guacamole_admin_username"] = v.adminUsername
		ctx.Lab.ServiceData["guacamole_admin_password"] = v.adminPassword
		ctx.Lab.ServiceData["guacamole_skip_tls_verify"] = fmt.Sprintf("%t", v.skipTLSVerify)
	}

	// Add credential to lab
	credential := &interfaces.Credential{
		ID:        fmt.Sprintf("guacamole-%s", shortID),
		LabID:     ctx.LabID,
		Label:     "Guacamole Access",
		Username:  labUsername,
		Password:  labPassword,
		URL:       v.host,
		ExpiresAt: time.Now().Add(time.Duration(ctx.Duration) * time.Minute),
		Notes:     "Apache Guacamole remote desktop access",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ctx.AddCredential(credential); err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating User Account", "failed", fmt.Sprintf("Failed to add credential: %v", err))
		}
		return fmt.Errorf("failed to add Guacamole credential: %w", err)
	}

	fmt.Printf("Guacamole user setup completed for lab %s\n", ctx.LabName)
	return nil
}

// ExecuteCleanup cleans up Guacamole user resources
func (v *GuacamoleService) ExecuteCleanup(ctx *interfaces.CleanupContext) error {
	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	// Get configuration from lab's ServiceData
	var host, adminUsername, adminPassword string
	var skipTLSVerify bool

	if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
		host = ctx.Lab.ServiceData["guacamole_host"]
		adminUsername = ctx.Lab.ServiceData["guacamole_admin_username"]
		adminPassword = ctx.Lab.ServiceData["guacamole_admin_password"]
		skipTLSVerifyStr := ctx.Lab.ServiceData["guacamole_skip_tls_verify"]
		skipTLSVerify = skipTLSVerifyStr == "true"
	}

	// Fallback to environment variables if not in ServiceData
	if host == "" {
		host = v.host
	}
	if adminUsername == "" {
		adminUsername = v.adminUsername
	}
	if adminPassword == "" {
		adminPassword = v.adminPassword
	}

	// Validate required configuration
	if host == "" || adminUsername == "" || adminPassword == "" {
		return fmt.Errorf("GUACAMOLE_HOST, GUACAMOLE_ADMIN_USERNAME, and GUACAMOLE_ADMIN_PASSWORD configuration not found in lab data or environment")
	}

	// Get lab-specific data from context
	username, ok := ctx.Context.Value("guacamole_user_username").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedUsername, exists := ctx.Lab.ServiceData["guacamole_user_username"]; exists {
				username = storedUsername
				fmt.Printf("Retrieved username from lab ServiceData: %s\n", username)
			} else {
				// If username is not in context, construct it from lab ID
				username = fmt.Sprintf("lab-%s", shortID)
				fmt.Printf("Warning: guacamole user username not found in context or lab data, using constructed username: %s\n", username)
			}
		} else {
			// If username is not in context, construct it from lab ID
			username = fmt.Sprintf("lab-%s", shortID)
			fmt.Printf("Warning: guacamole user username not found in context, using constructed username: %s\n", username)
		}
	}

	// Create Guacamole client for cleanup
	client, err := NewGuacamoleClient(host, adminUsername, adminPassword, skipTLSVerify)
	if err != nil {
		return fmt.Errorf("failed to create Guacamole client for cleanup: %w", err)
	}

	fmt.Printf("Cleaning up Guacamole user resources for lab %s:\n", ctx.LabID)

	// Delete user
	fmt.Printf("- Deleting user: %s\n", username)
	if err := client.deleteUser(username); err != nil {
		fmt.Printf("Warning: Failed to delete user: %v\n", err)
	} else {
		fmt.Printf("  User deleted successfully\n")
	}

	fmt.Printf("Guacamole user cleanup completed for lab %s\n", ctx.LabID)
	return nil
}
