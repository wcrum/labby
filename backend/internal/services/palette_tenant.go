package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wcrum/labby/internal/interfaces"

	internalclient "github.com/spectrocloud/palette-sdk-go-internal/client"

	"github.com/sethvargo/go-password/password"
	hapimodels "github.com/spectrocloud/hapi/models"
)

// Constants
const (
	PlaceholderTenantID = "TODO_TENANT_ID"
)

// PaletteTenantService handles setup and cleanup for Palette Tenant user accounts
type PaletteTenantService struct {
	host           string
	systemUsername string
	systemPassword string
}

// NewPaletteTenantService creates a new Palette Tenant service instance
func NewPaletteTenantService() *PaletteTenantService {
	return &PaletteTenantService{
		host:           os.Getenv("palette_host"),
		systemUsername: os.Getenv("palette_system_username"),
		systemPassword: os.Getenv("palette_system_password"),
	}
}

// GetName returns the service name
func (v *PaletteTenantService) GetName() string {
	return "palette_tenant"
}

// GetDescription returns the service description
func (v *PaletteTenantService) GetDescription() string {
	return "Palette Tenant user account access"
}

// GetRequiredParams returns the required parameters for this service
func (v *PaletteTenantService) GetRequiredParams() []string {
	return []string{"palette_host", "palette_system_username", "palette_system_password"}
}

// Name returns the service name (implements Setup interface)
func (v *PaletteTenantService) Name() string {
	return v.GetName()
}

// ExecuteSetup sets up Palette Tenant user access and adds credentials
func (v *PaletteTenantService) ExecuteSetup(ctx *interfaces.SetupContext) error {
	fmt.Printf("PaletteTenantService.ExecuteSetup called for lab: %s\n", ctx.LabName)

	// Update progress: Connecting to Palette
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Connecting to Palette", "running", "Connecting to Palette tenant...")
	}

	fmt.Printf("Environment variables check:\n")
	fmt.Printf("  palette_host: %s\n", v.host)
	fmt.Printf("  palette_system_username: %s\n", v.systemUsername)
	fmt.Printf("  palette_system_password: [%s]\n", func() string {
		if v.systemPassword == "" {
			return "EMPTY"
		}
		return "SET"
	}())

	if v.host == "" || v.systemUsername == "" || v.systemPassword == "" {
		hostStatus := "EMPTY"
		if v.host != "" {
			hostStatus = "SET"
		}
		usernameStatus := "EMPTY"
		if v.systemUsername != "" {
			usernameStatus = "SET"
		}
		passwordStatus := "EMPTY"
		if v.systemPassword != "" {
			passwordStatus = "SET"
		}
		errMsg := fmt.Sprintf("Missing required environment variables: palette_host=%s, palette_system_username=%s, palette_system_password=%s",
			hostStatus, usernameStatus, passwordStatus)
		fmt.Printf("ERROR: %s\n", errMsg)
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Connecting to Palette", "failed", errMsg)
		}
		return fmt.Errorf("%s", errMsg)
	}

	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	fmt.Printf("Setting up Palette Tenant for lab %s...\n", ctx.LabName)

	// Initialize Palette client with internal SDK using system credentials
	fmt.Printf("- Initializing Palette client with host: %s\n", v.host)
	pc := internalclient.New(
		internalclient.WithHubbleURI(v.host),
		internalclient.WithUsername(v.systemUsername),
		internalclient.WithPassword(v.systemPassword),
		internalclient.WithScopeSystem(v.systemUsername, v.systemPassword),
	)

	// Validate that the client was created successfully
	if pc == nil {
		err := fmt.Errorf("failed to create Palette client: client is nil")
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Connecting to Palette", "failed", err.Error())
		}
		return err
	}
	fmt.Printf("  Palette client initialized successfully\n")

	// Authenticate with system credentials
	fmt.Printf("- Authenticating with system credentials (username: %s)\n", v.systemUsername)
	_, err := pc.SysAdminLogin(v.systemUsername, v.systemPassword)
	if err != nil {
		fmt.Printf("ERROR: Authentication failed: %v\n", err)
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Connecting to Palette", "failed", fmt.Sprintf("Authentication failed: %v", err))
		}
		return fmt.Errorf("failed to authenticate with system credentials: %w", err)
	}
	fmt.Printf("  Authentication successful\n")

	// Update progress: Connecting to Palette completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Connecting to Palette", "completed", "Successfully connected to Palette")
	}

	// Update progress: Creating User Account
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating User Account", "running", "Creating Palette tenant user account...")
	}

	// Create new tenant
	tenantName := fmt.Sprintf("lab-%s", shortID)
	fmt.Printf("- Creating tenant: %s\n", tenantName)

	// Create tenant entity with spec
	fmt.Printf("- Creating tenant entity with name: %s\n", tenantName)

	// Validate that we have all required data before creating the entity
	if shortID == "" {
		err := fmt.Errorf("shortID is empty, cannot create tenant entity")
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating User Account", "failed", err.Error())
		}
		return err
	}

	tenantEntity := &hapimodels.V1TenantEntity{
		Metadata: &hapimodels.V1ObjectMeta{
			Name: tenantName,
		},
		Spec: &hapimodels.V1TenantSpecEntity{
			OrgName:   tenantName,
			FirstName: "Lab",
			LastName:  "Admin",
			EmailID:   fmt.Sprintf("lab-admin-%s@spectrocloud.com", shortID),
			AuthType:  "password",
		},
	}

	// Validate the created entity
	if tenantEntity.Metadata == nil {
		return fmt.Errorf("failed to create tenant entity: Metadata is nil")
	}
	if tenantEntity.Spec == nil {
		return fmt.Errorf("failed to create tenant entity: Spec is nil")
	}
	if tenantEntity.Spec.EmailID == "" {
		return fmt.Errorf("failed to create tenant entity: EmailID is empty")
	}

	fmt.Printf("  Tenant entity details:\n")
	fmt.Printf("    OrgName: %s\n", tenantEntity.Spec.OrgName)
	fmt.Printf("    OrgEmailID: %s\n", tenantEntity.Spec.OrgEmailID)
	fmt.Printf("    EmailID: %s\n", tenantEntity.Spec.EmailID)
	fmt.Printf("    AuthType: %s\n", tenantEntity.Spec.AuthType)
	fmt.Printf("    Roles: %v\n", tenantEntity.Spec.Roles)

	// Create tenant using internal SDK
	fmt.Printf("- Calling CreateTenant API...\n")
	tenantID, err := pc.CreateTenant(tenantEntity)
	if err != nil {
		fmt.Printf("ERROR: CreateTenant failed: %v\n", err)
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating User Account", "failed", fmt.Sprintf("Failed to create tenant: %v", err))
		}
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	fmt.Printf("  Tenant created successfully with ID: %s\n", tenantID)

	// Update progress: Setting Password
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Setting Password", "running", "Setting up tenant password...")
	}

	// Get password token for the tenant
	fmt.Printf("- Getting password token for tenant: %s\n", tenantID)
	passwordToken, err := pc.GetPasswordToken(tenantID)
	if err != nil {
		fmt.Printf("ERROR: Failed to get password token: %v\n", err)
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Setting Password", "failed", fmt.Sprintf("Failed to get password token: %v", err))
		}
		return fmt.Errorf("failed to get password token: %w", err)
	}
	fmt.Printf("  Password token retrieved successfully\n")

	// Generate a secure password for the tenant admin
	fmt.Printf("- Generating secure password for tenant admin...\n")
	mySecretPassword, err := password.Generate(16, 4, 4, false, false)
	if err != nil {
		fmt.Printf("ERROR: Failed to generate password: %v\n", err)
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Setting Password", "failed", fmt.Sprintf("Failed to generate password: %v", err))
		}
		return fmt.Errorf("failed to generate password: %w", err)
	}

	// Validate that password was generated
	if mySecretPassword == "" {
		err := fmt.Errorf("generated password is empty")
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Setting Password", "failed", err.Error())
		}
		return err
	}

	goodPassword := "L3@rN-" + mySecretPassword
	fmt.Printf("  Password generated successfully\n")

	// Activate the user with the password
	fmt.Printf("- Activating tenant admin user with password...\n")
	err = pc.ActivateUser(goodPassword, passwordToken)
	if err != nil {
		fmt.Printf("ERROR: Failed to activate user: %v\n", err)
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Setting Password", "failed", fmt.Sprintf("Failed to activate user: %v", err))
		}
		return fmt.Errorf("failed to activate user: %w", err)
	}
	fmt.Printf("  Tenant admin user activated successfully\n")

	// Store the password for credential creation
	tenantPassword := goodPassword

	// Store tenant spec entity content for later use
	tenantSpecData := map[string]interface{}{
		"orgName":   tenantEntity.Spec.OrgName,
		"firstName": tenantEntity.Spec.FirstName,
		"lastName":  tenantEntity.Spec.LastName,
		"emailID":   tenantEntity.Spec.EmailID,
		"authType":  tenantEntity.Spec.AuthType,
	}

	// Safely handle LoginMode which might be nil
	if tenantEntity.Spec.LoginMode != nil {
		tenantSpecData["loginMode"] = *tenantEntity.Spec.LoginMode
	} else {
		tenantSpecData["loginMode"] = "password" // Default value
		fmt.Printf("  Warning: LoginMode was nil, using default value 'password'\n")
	}

	// Store lab-specific data in context for cleanup
	ctx.Context = context.WithValue(ctx.Context, "palette_tenant_id", tenantID)
	ctx.Context = context.WithValue(ctx.Context, "palette_tenant_spec", tenantSpecData)

	// Store in lab's ServiceData for persistence
	if ctx.Lab != nil {
		if ctx.Lab.ServiceData == nil {
			ctx.Lab.ServiceData = make(map[string]string)
		}
		ctx.Lab.ServiceData["palette_tenant_id"] = tenantID
		// Store tenant spec data as JSON string
		ctx.Lab.ServiceData["palette_tenant_spec_org_name"] = tenantEntity.Spec.OrgName
		ctx.Lab.ServiceData["palette_tenant_spec_org_email"] = tenantEntity.Spec.OrgEmailID
		ctx.Lab.ServiceData["palette_tenant_spec_first_name"] = tenantEntity.Spec.FirstName
		ctx.Lab.ServiceData["palette_tenant_spec_last_name"] = tenantEntity.Spec.LastName
		ctx.Lab.ServiceData["palette_tenant_spec_email"] = tenantEntity.Spec.EmailID
		ctx.Lab.ServiceData["palette_tenant_spec_auth_type"] = tenantEntity.Spec.AuthType
		ctx.Lab.ServiceData["palette_tenant_spec_roles"] = strings.Join(tenantEntity.Spec.Roles, ",")
		// Store configuration for cleanup
		ctx.Lab.ServiceData["palette_tenant_host"] = v.host
		ctx.Lab.ServiceData["palette_tenant_system_username"] = v.systemUsername
		ctx.Lab.ServiceData["palette_tenant_system_password"] = v.systemPassword
		// Store tenant password for reference
		ctx.Lab.ServiceData["palette_tenant_password"] = tenantPassword
	}

	// Add credential to lab
	fmt.Printf("- Creating credential for lab access...\n")
	credential := &interfaces.Credential{
		ID:        fmt.Sprintf("palette-tenant-%s", shortID),
		LabID:     ctx.LabID,
		Label:     "Palette Tenant",
		Username:  tenantEntity.Spec.EmailID,
		Password:  tenantPassword, // Use the generated password for tenant admin
		URL:       v.host,
		ExpiresAt: time.Now().Add(time.Duration(ctx.Duration) * time.Minute),
		Notes:     fmt.Sprintf("Palette Tenant access (Tenant: %s, Org: %s)", tenantName, tenantEntity.Spec.OrgName),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fmt.Printf("  Credential details:\n")
	fmt.Printf("    ID: %s\n", credential.ID)
	fmt.Printf("    Username: %s\n", credential.Username)
	fmt.Printf("    URL: %s\n", credential.URL)
	fmt.Printf("    Notes: %s\n", credential.Notes)

	// Validate that AddCredential function is available
	if ctx.AddCredential == nil {
		return fmt.Errorf("AddCredential function is nil, cannot add credential")
	}

	if err := ctx.AddCredential(credential); err != nil {
		fmt.Printf("ERROR: Failed to add credential: %v\n", err)
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating User Account", "failed", fmt.Sprintf("Failed to add credential: %v", err))
		}
		return fmt.Errorf("failed to add Palette Tenant credential: %w", err)
	}
	fmt.Printf("  Credential added successfully\n")

	// Update progress: All steps completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating User Account", "completed", "Palette tenant user created successfully")
		ctx.UpdateProgress("Setting Password", "completed", "Password set successfully")
	}

	fmt.Printf("Palette Tenant setup completed for lab %s\n", ctx.LabName)
	return nil
}

// ExecuteCleanup cleans up Palette Tenant user resources
func (v *PaletteTenantService) ExecuteCleanup(ctx *interfaces.CleanupContext) error {
	fmt.Printf("PaletteTenantService.ExecuteCleanup called for lab: %s\n", ctx.LabID)
	fmt.Printf("PaletteTenantService: Starting cleanup process...\n")

	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID
	fmt.Printf("Extracted short ID: %s\n", shortID)

	// Get configuration from lab's ServiceData
	fmt.Printf("- Retrieving configuration from lab ServiceData...\n")
	var host, systemUsername, systemPassword string

	if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
		host = ctx.Lab.ServiceData["palette_tenant_host"]
		systemUsername = ctx.Lab.ServiceData["palette_tenant_system_username"]
		systemPassword = ctx.Lab.ServiceData["palette_tenant_system_password"]
		fmt.Printf("  Retrieved from ServiceData:\n")
		fmt.Printf("    host: %s\n", host)
		fmt.Printf("    systemUsername: %s\n", systemUsername)
		fmt.Printf("    systemPassword: [%s]\n", func() string {
			if systemPassword == "" {
				return "EMPTY"
			}
			return "SET"
		}())
	} else {
		fmt.Printf("  Warning: Lab or ServiceData is nil\n")
	}

	// Fallback to environment variables if not in ServiceData
	if host == "" {
		host = v.host
		fmt.Printf("  Using environment variable for host: %s\n", host)
	}
	if systemUsername == "" {
		systemUsername = v.systemUsername
		fmt.Printf("  Using environment variable for systemUsername: %s\n", systemUsername)
	}
	if systemPassword == "" {
		systemPassword = v.systemPassword
		fmt.Printf("  Using environment variable for systemPassword: [%s]\n", func() string {
			if systemPassword == "" {
				return "EMPTY"
			}
			return "SET"
		}())
	}

	// Validate required configuration
	if host == "" || systemUsername == "" || systemPassword == "" {
		hostStatus := "EMPTY"
		if host != "" {
			hostStatus = "SET"
		}
		usernameStatus := "EMPTY"
		if systemUsername != "" {
			usernameStatus = "SET"
		}
		passwordStatus := "EMPTY"
		if systemPassword != "" {
			passwordStatus = "SET"
		}
		errMsg := fmt.Sprintf("Missing required configuration: host=%s, systemUsername=%s, systemPassword=%s",
			hostStatus, usernameStatus, passwordStatus)
		fmt.Printf("ERROR: %s\n", errMsg)
		return fmt.Errorf("%s", errMsg)
	}

	// Get lab-specific data from context
	tenantID, ok := ctx.Context.Value("palette_tenant_id").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedTenantID, exists := ctx.Lab.ServiceData["palette_tenant_id"]; exists {
				tenantID = storedTenantID
				fmt.Printf("Retrieved tenant ID from lab ServiceData: %s\n", tenantID)
			} else {
				// If tenant ID is not in context, construct it from lab ID
				tenantID = fmt.Sprintf("tenant-%s", shortID)
				fmt.Printf("Warning: palette tenant ID not found in context or lab data, using constructed tenant ID: %s\n", tenantID)
			}
		} else {
			// If tenant ID is not in context, construct it from lab ID
			tenantID = fmt.Sprintf("tenant-%s", shortID)
			fmt.Printf("Warning: palette tenant ID not found in context, using constructed tenant ID: %s\n", tenantID)
		}
	}

	// Initialize Palette client with internal SDK using system credentials
	pc := internalclient.New(
		internalclient.WithHubbleURI(host),
		internalclient.WithUsername(systemUsername),
		internalclient.WithPassword(systemPassword),
		internalclient.WithScopeSystem(systemUsername, systemPassword),
	)

	// Validate that the client was created successfully
	if pc == nil {
		return fmt.Errorf("failed to create Palette client for cleanup: client is nil")
	}

	// Authenticate with system credentials
	fmt.Printf("- Authenticating with system credentials for cleanup\n")
	_, err := pc.SysAdminLogin(systemUsername, systemPassword)
	if err != nil {
		return fmt.Errorf("failed to authenticate with system credentials for cleanup: %w", err)
	}

	fmt.Printf("Cleaning up Palette Tenant resources for lab %s:\n", ctx.LabID)

	// Delete Tenant (only if we found the tenant ID)
	if tenantID != "" && tenantID != PlaceholderTenantID {
		fmt.Printf("- Deleting tenant: %s\n", tenantID)
		if err := pc.DeleteTenant(tenantID); err != nil {
			fmt.Printf("Warning: Failed to delete tenant: %v\n", err)
		} else {
			fmt.Printf("  Tenant deleted successfully\n")
		}
	} else {
		fmt.Printf("- Skipping tenant deletion (no valid tenant ID found)\n")
	}

	fmt.Printf("Palette Tenant cleanup completed for lab %s\n", ctx.LabID)
	fmt.Printf("PaletteTenantService: Cleanup process completed successfully\n")
	return nil
}
