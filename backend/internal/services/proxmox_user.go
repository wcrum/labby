package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/wcrum/labby/internal/interfaces"

	"github.com/sethvargo/go-password/password"
)

// ProxmoxUserService handles setup and cleanup for Proxmox user accounts
type ProxmoxUserService struct {
	uri           string
	adminUser     string
	adminPass     string
	skipTLSVerify bool
}

// NewProxmoxUserService creates a new Proxmox user service instance
func NewProxmoxUserService() *ProxmoxUserService {
	return &ProxmoxUserService{
		uri:           os.Getenv("PROXMOX_URI"),
		adminUser:     os.Getenv("PROXMOX_ADMIN_USER"),
		adminPass:     os.Getenv("PROXMOX_ADMIN_PASS"),
		skipTLSVerify: os.Getenv("PROXMOX_SKIP_TLS_VERIFY") == "true",
	}
}

// ConfigureFromServiceConfig configures the service from a service configuration
func (v *ProxmoxUserService) ConfigureFromServiceConfig(config map[string]string) {
	if uri, ok := config["uri"]; ok {
		v.uri = uri
	}
	if adminUser, ok := config["admin_user"]; ok {
		v.adminUser = adminUser
	}
	if adminPass, ok := config["admin_pass"]; ok {
		v.adminPass = adminPass
	}
	if skipTLSVerify, ok := config["skip_tls_verify"]; ok {
		v.skipTLSVerify = skipTLSVerify == "true"
	}
}

// GetName returns the service name
func (v *ProxmoxUserService) GetName() string {
	return "proxmox_user"
}

// GetDescription returns the service description
func (v *ProxmoxUserService) GetDescription() string {
	return "Proxmox VE user account access and resource pool"
}

// GetRequiredParams returns the required parameters for this service
func (v *ProxmoxUserService) GetRequiredParams() []string {
	return []string{"PROXMOX_URI", "PROXMOX_ADMIN_USER", "PROXMOX_ADMIN_PASS"}
}

// Name returns the service name (implements Setup interface)
func (v *ProxmoxUserService) Name() string {
	return v.GetName()
}

// ProxmoxClient represents a Proxmox API client
type ProxmoxClient struct {
	baseURL    string
	httpClient *http.Client
	ticket     string
	csrfToken  string
}

// NewProxmoxClient creates a new Proxmox client
func NewProxmoxClient(baseURL, username, password string, skipTLSVerify bool) (*ProxmoxClient, error) {
	// Create HTTP client with TLS configuration
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTLSVerify,
			},
		},
	}

	client := &ProxmoxClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}

	// Authenticate and get ticket
	if err := client.authenticate(username, password); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return client, nil
}

// authenticate performs authentication and gets ticket/CSRF token
func (pc *ProxmoxClient) authenticate(username, password string) error {
	loginURL := fmt.Sprintf("%s/api2/json/access/ticket", pc.baseURL)

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	fmt.Printf("Authenticating to Proxmox at: %s\n", loginURL)
	fmt.Printf("Using admin user: %s\n", username)

	resp, err := pc.httpClient.PostForm(loginURL, data)
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

	var result struct {
		Data struct {
			Ticket              string `json:"ticket"`
			CSRFPreventionToken string `json:"CSRFPreventionToken"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to decode authentication response: %w", err)
	}

	pc.ticket = result.Data.Ticket
	pc.csrfToken = result.Data.CSRFPreventionToken

	fmt.Printf("Authentication successful, ticket length: %d, CSRF token length: %d\n", len(pc.ticket), len(pc.csrfToken))

	return nil
}

// createUser creates a new Proxmox user
func (pc *ProxmoxClient) createUser(username, password string) error {
	createURL := fmt.Sprintf("%s/api2/json/access/users", pc.baseURL)

	data := url.Values{}
	data.Set("userid", username)
	data.Set("password", password)
	data.Set("comment", "Lab user account")

	req, err := http.NewRequest("POST", createURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", pc.ticket))
	req.Header.Set("CSRFPreventionToken", pc.csrfToken)

	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("create user request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create user failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

// resetUserPassword resets a user's password
func (pc *ProxmoxClient) resetUserPassword(username, newPassword string) error {
	resetURL := fmt.Sprintf("%s/api2/json/access/users/%s", pc.baseURL, username)

	data := url.Values{}
	data.Set("password", newPassword)

	req, err := http.NewRequest("PUT", resetURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", pc.ticket))
	req.Header.Set("CSRFPreventionToken", pc.csrfToken)

	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("reset password request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reset password failed with status: %d", resp.StatusCode)
	}

	return nil
}

// deleteUser deletes a Proxmox user
func (pc *ProxmoxClient) deleteUser(username string) error {
	deleteURL := fmt.Sprintf("%s/api2/json/access/users/%s", pc.baseURL, username)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", pc.ticket))
	req.Header.Set("CSRFPreventionToken", pc.csrfToken)

	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete user request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete user failed with status: %d", resp.StatusCode)
	}

	return nil
}

// createPool creates a new Proxmox pool
func (pc *ProxmoxClient) createPool(poolName string) error {
	createURL := fmt.Sprintf("%s/api2/json/pools", pc.baseURL)

	data := url.Values{}
	data.Set("poolid", poolName)
	data.Set("comment", "Lab resource pool")

	req, err := http.NewRequest("POST", createURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", pc.ticket))
	req.Header.Set("CSRFPreventionToken", pc.csrfToken)

	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("create pool request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("create pool failed with status: %d, response: %s", resp.StatusCode, string(body))
	}

	return nil
}

// deletePool deletes a Proxmox pool
func (pc *ProxmoxClient) deletePool(poolName string) error {
	deleteURL := fmt.Sprintf("%s/api2/json/pools/%s", pc.baseURL, poolName)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", pc.ticket))
	req.Header.Set("CSRFPreventionToken", pc.csrfToken)

	resp, err := pc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete pool request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete pool failed with status: %d", resp.StatusCode)
	}

	return nil
}

// ExecuteSetup sets up Proxmox user access and adds credentials
func (v *ProxmoxUserService) ExecuteSetup(ctx *interfaces.SetupContext) error {
	// Update progress: Connecting to Proxmox
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Connecting to Proxmox", "running", "Connecting to Proxmox cluster...")
	}

	if v.uri == "" || v.adminUser == "" || v.adminPass == "" {
		err := fmt.Errorf("PROXMOX_URI, PROXMOX_ADMIN_USER, and PROXMOX_ADMIN_PASS environment variables are required")
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Connecting to Proxmox", "failed", err.Error())
		}
		return err
	}

	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	fmt.Printf("Setting up Proxmox user for lab %s...\n", ctx.LabName)

	// Create Proxmox client
	client, err := NewProxmoxClient(v.uri, v.adminUser, v.adminPass, v.skipTLSVerify)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Connecting to Proxmox", "failed", fmt.Sprintf("Failed to connect: %v", err))
		}
		return fmt.Errorf("failed to create Proxmox client: %w", err)
	}

	// Update progress: Connecting to Proxmox completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Connecting to Proxmox", "completed", "Successfully connected to Proxmox cluster")
	}

	// Update progress: Creating User Account
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating User Account", "running", "Creating Proxmox user account...")
	}

	// Generate username and password for lab user
	labUsername := fmt.Sprintf("lab-%s@pve", shortID)
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
		ctx.UpdateProgress("Creating User Account", "completed", "Proxmox user account created successfully")
	}

	// Update progress: Creating Resource Pool
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating Resource Pool", "running", "Creating Proxmox resource pool...")
	}

	// Generate pool name
	poolName := fmt.Sprintf("lab-%s-pool", shortID)

	// Create pool
	fmt.Printf("- Creating pool: %s\n", poolName)
	if err := client.createPool(poolName); err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating Resource Pool", "failed", fmt.Sprintf("Failed to create pool: %v", err))
		}
		return fmt.Errorf("failed to create pool: %w", err)
	}
	fmt.Printf("  Pool created successfully\n")

	// Update progress: Creating Resource Pool completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating Resource Pool", "completed", "Proxmox resource pool created successfully")
	}

	// Update progress: Setting Password
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Setting Password", "running", "Setting up user password...")
	}

	// Store lab-specific data in context for cleanup
	ctx.Context = context.WithValue(ctx.Context, "proxmox_user_username", labUsername)
	ctx.Context = context.WithValue(ctx.Context, "proxmox_user_password", labPassword)
	ctx.Context = context.WithValue(ctx.Context, "proxmox_pool_name", poolName)

	// Store in lab's ServiceData for persistence
	if ctx.Lab != nil {
		if ctx.Lab.ServiceData == nil {
			ctx.Lab.ServiceData = make(map[string]string)
		}
		ctx.Lab.ServiceData["proxmox_user_username"] = labUsername
		ctx.Lab.ServiceData["proxmox_user_password"] = labPassword
		ctx.Lab.ServiceData["proxmox_pool_name"] = poolName
		// Store configuration for cleanup
		ctx.Lab.ServiceData["proxmox_uri"] = v.uri
		ctx.Lab.ServiceData["proxmox_admin_user"] = v.adminUser
		ctx.Lab.ServiceData["proxmox_admin_pass"] = v.adminPass
		ctx.Lab.ServiceData["proxmox_skip_tls_verify"] = fmt.Sprintf("%t", v.skipTLSVerify)
	}

	// Add credential to lab
	credential := &interfaces.Credential{
		ID:        fmt.Sprintf("proxmox-%s", shortID),
		LabID:     ctx.LabID,
		Label:     "Proxmox VE",
		Username:  labUsername,
		Password:  labPassword,
		URL:       v.uri,
		ExpiresAt: time.Now().Add(time.Duration(ctx.Duration) * time.Minute),
		Notes:     fmt.Sprintf("Proxmox VE cluster management access. Resource pool: %s", poolName),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ctx.AddCredential(credential); err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Setting Password", "failed", fmt.Sprintf("Failed to add credential: %v", err))
		}
		return fmt.Errorf("failed to add Proxmox credential: %w", err)
	}

	// Update progress: Setting Password completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Setting Password", "completed", "Password set successfully")
	}

	fmt.Printf("Proxmox user setup completed for lab %s\n", ctx.LabName)
	return nil
}

// ExecuteCleanup cleans up Proxmox user resources
func (v *ProxmoxUserService) ExecuteCleanup(ctx *interfaces.CleanupContext) error {
	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	// Get configuration from lab's ServiceData
	var uri, adminUser, adminPass string
	var skipTLSVerify bool

	if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
		uri = ctx.Lab.ServiceData["proxmox_uri"]
		adminUser = ctx.Lab.ServiceData["proxmox_admin_user"]
		adminPass = ctx.Lab.ServiceData["proxmox_admin_pass"]
		skipTLSVerifyStr := ctx.Lab.ServiceData["proxmox_skip_tls_verify"]
		skipTLSVerify = skipTLSVerifyStr == "true"
	}

	// Fallback to environment variables if not in ServiceData
	if uri == "" {
		uri = v.uri
	}
	if adminUser == "" {
		adminUser = v.adminUser
	}
	if adminPass == "" {
		adminPass = v.adminPass
	}

	// Validate required configuration
	if uri == "" || adminUser == "" || adminPass == "" {
		return fmt.Errorf("PROXMOX_URI, PROXMOX_ADMIN_USER, and PROXMOX_ADMIN_PASS configuration not found in lab data or environment")
	}

	// Get lab-specific data from context
	username, ok := ctx.Context.Value("proxmox_user_username").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedUsername, exists := ctx.Lab.ServiceData["proxmox_user_username"]; exists {
				username = storedUsername
				fmt.Printf("Retrieved username from lab ServiceData: %s\n", username)
			} else {
				// If username is not in context, construct it from lab ID
				username = fmt.Sprintf("lab-%s@pve", shortID)
				fmt.Printf("Warning: proxmox user username not found in context or lab data, using constructed username: %s\n", username)
			}
		} else {
			// If username is not in context, construct it from lab ID
			username = fmt.Sprintf("lab-%s@pve", shortID)
			fmt.Printf("Warning: proxmox user username not found in context, using constructed username: %s\n", username)
		}
	}

	// Get pool name from context or construct it
	poolName, ok := ctx.Context.Value("proxmox_pool_name").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedPoolName, exists := ctx.Lab.ServiceData["proxmox_pool_name"]; exists {
				poolName = storedPoolName
				fmt.Printf("Retrieved pool name from lab ServiceData: %s\n", poolName)
			} else {
				// If pool name is not in context, construct it from lab ID
				poolName = fmt.Sprintf("lab-%s-pool", shortID)
				fmt.Printf("Warning: proxmox pool name not found in context or lab data, using constructed pool name: %s\n", poolName)
			}
		} else {
			// If pool name is not in context, construct it from lab ID
			poolName = fmt.Sprintf("lab-%s-pool", shortID)
			fmt.Printf("Warning: proxmox pool name not found in context, using constructed pool name: %s\n", poolName)
		}
	}

	// Create Proxmox client for cleanup
	client, err := NewProxmoxClient(uri, adminUser, adminPass, skipTLSVerify)
	if err != nil {
		return fmt.Errorf("failed to create Proxmox client for cleanup: %w", err)
	}

	fmt.Printf("Cleaning up Proxmox user resources for lab %s:\n", ctx.LabID)

	// Delete user
	fmt.Printf("- Deleting user: %s\n", username)
	if err := client.deleteUser(username); err != nil {
		fmt.Printf("Warning: Failed to delete user: %v\n", err)
	} else {
		fmt.Printf("  User deleted successfully\n")
	}

	// Delete pool
	fmt.Printf("- Deleting pool: %s\n", poolName)
	if err := client.deletePool(poolName); err != nil {
		fmt.Printf("Warning: Failed to delete pool: %v\n", err)
	} else {
		fmt.Printf("  Pool deleted successfully\n")
	}

	fmt.Printf("Proxmox user cleanup completed for lab %s\n", ctx.LabID)
	return nil
}
