package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wcrum/labby/internal/interfaces"
	"github.com/wcrum/labby/internal/models"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/sethvargo/go-password/password"
	"github.com/spectrocloud/palette-sdk-go/api/client/version1"
	palettemodels "github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/palette-sdk-go/client"
)

// PaletteProjectService handles setup and cleanup for Palette Project (Spectro Cloud control plane)
type PaletteProjectService struct {
	host       string
	apiKey     string
	projectUID string
	// Service config credentials (preferred)
	serviceConfig *models.ServiceConfig
}

// NewPaletteProjectService creates a new Palette Project service instance
func NewPaletteProjectService() *PaletteProjectService {
	return &PaletteProjectService{
		host:       os.Getenv("PALETTE_HOST"),
		apiKey:     os.Getenv("PALETTE_API_KEY"),
		projectUID: os.Getenv("PALETTE_PROJECT_UID"),
	}
}

// ConfigureFromServiceConfig configures the service with credentials from service config
func (v *PaletteProjectService) ConfigureFromServiceConfig(serviceConfig *models.ServiceConfig) {
	v.serviceConfig = serviceConfig

	// Override environment variables with service config values
	if host, ok := serviceConfig.Config["host"]; ok {
		v.host = host
	}
	if apiKey, ok := serviceConfig.Config["api_key"]; ok {
		v.apiKey = apiKey
	}
	if projectUID, ok := serviceConfig.Config["project_uid"]; ok {
		v.projectUID = projectUID
	}
}

// GetName returns the service name
func (v *PaletteProjectService) GetName() string {
	return "palette"
}

// GetDescription returns the service description
func (v *PaletteProjectService) GetDescription() string {
	return "Spectro Cloud Project access"
}

// GetRequiredParams returns the required parameters for this service
func (v *PaletteProjectService) GetRequiredParams() []string {
	return []string{"PALETTE_HOST", "PALETTE_API_KEY"}
}

// Name returns the service name (implements Setup interface)
func (v *PaletteProjectService) Name() string {
	return v.GetName()
}

// ExecuteSetup sets up Palette Project access and adds credentials
func (v *PaletteProjectService) ExecuteSetup(ctx *interfaces.SetupContext) error {
	// Update progress: Creating Project
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating Project", "running", "Creating project in Spectro Cloud...")
	}

	if v.host == "" || v.apiKey == "" {
		err := fmt.Errorf("PALETTE_HOST and PALETTE_API_KEY environment variables are required")
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating Project", "failed", err.Error())
		}
		return err
	}

	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	fmt.Printf("Setting up Palette Project for lab %s...\n", ctx.LabName)

	// Initialize Palette client
	pc := client.New(
		client.WithPaletteURI(v.host),
		client.WithAPIKey(v.apiKey),
	)

	// Set scope based on project UID
	scope := "tenant"
	if v.projectUID != "" {
		client.WithScopeProject(v.projectUID)(pc)
		scope = "project"
	} else {
		client.WithScopeTenant()(pc)
	}

	fmt.Printf("Using %s scope\n", scope)

	// Create Project Entity
	projectEntity := palettemodels.V1ProjectEntity{
		Metadata: &palettemodels.V1ObjectMeta{
			Name: fmt.Sprintf("lab-%s", shortID),
		},
	}

	// Create User Entity
	userEntity := palettemodels.V1UserEntity{
		Spec: &palettemodels.V1UserSpecEntity{
			EmailID:   fmt.Sprintf("lab+%s@spectrocloud.com", shortID),
			FirstName: "Lab",
			LastName:  "User",
		},
	}

	// Create Project
	fmt.Printf("- Creating project: %s\n", projectEntity.Metadata.Name)
	projectID, err := pc.CreateProject(&projectEntity)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating Project", "failed", fmt.Sprintf("Failed to create project: %v", err))
		}
		return fmt.Errorf("failed to create project: %w", err)
	}
	fmt.Printf("  Project created with ID: %s\n", projectID)

	// Update progress: Creating Project completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating Project", "completed", "Project created successfully")
	}

	// Update progress: Setting up User Account
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Setting up User Account", "running", "Setting up user account...")
	}

	// Create User
	fmt.Printf("- Creating user: %s\n", userEntity.Spec.EmailID)
	userID, err := pc.CreateUser(&userEntity)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Setting up User Account", "failed", fmt.Sprintf("Failed to create user: %v", err))
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	fmt.Printf("  User created with ID: %s\n", userID)

	// Update progress: Configuring Access Permissions
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Configuring Access Permissions", "running", "Configuring access permissions...")
	}

	// Get Project Admin Role
	fmt.Printf("- Getting Project Admin role\n")
	projectAdmin, err := pc.GetRole("Project Admin")
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Configuring Access Permissions", "failed", fmt.Sprintf("Failed to get Project Admin role: %v", err))
		}
		return fmt.Errorf("failed to get Project Admin role: %w", err)
	}

	// Associate User with Project Role
	projectRolePatch := palettemodels.V1ProjectRolesPatch{
		Projects: []*palettemodels.V1ProjectRolesPatchProjectsItems0{{
			ProjectUID: projectID,
			Roles: []string{
				projectAdmin.Metadata.UID,
			},
		}},
	}

	fmt.Printf("- Assigning Project Admin role to user\n")
	if err = pc.AssociateUserProjectRole(userID, &projectRolePatch); err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Configuring Access Permissions", "failed", fmt.Sprintf("Failed to associate user with project role: %v", err))
		}
		return fmt.Errorf("failed to associate user with project role: %w", err)
	}

	// Update progress: Configuring Access Permissions completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Configuring Access Permissions", "completed", "Permissions configured")
	}

	// Update progress: Setting up User Account completed (after user creation and role assignment)
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Setting up User Account", "completed", "User account created")
	}

	// Get User to get activation link
	user, err := pc.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Extract token from activation link
	token := ""
	if user.Status.ActivationLink != "" {
		fmt.Printf("  Activation link: %s\n", user.Status.ActivationLink)
		// Parse activation link to extract token using the original method
		parts := strings.Split(user.Status.ActivationLink, "/")
		if len(parts) > 5 {
			token = parts[5]
		}
		fmt.Printf("  Extracted token: %s\n", token)
	}

	// Generate secure password
	mySecretPassword, err := password.Generate(16, 4, 4, false, false)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}
	goodPassword := "L3@rN-" + mySecretPassword

	// Activate user password if token is available
	if token != "" {
		fmt.Printf("- Setting user password\n")
		passwordActivateParams := version1.NewV1PasswordActivateParams()
		passwordValues := strfmt.Password(goodPassword)
		passwordActivateParams.Body.Password = &passwordValues
		passwordActivateParams.PasswordToken = token

		// Try password activation with retry logic
		var respPass interface{}
		var err error
		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			respPass, err = pc.Client.V1PasswordActivate(passwordActivateParams)
			if err == nil {
				fmt.Printf("  Password activated successfully on attempt %d: %v\n", attempt, respPass)
				break
			} else {
				fmt.Printf("  Password activation attempt %d failed: %v\n", attempt, err)
				if attempt < maxRetries {
					time.Sleep(2 * time.Second) // Wait before retry
				}
			}
		}
		if err != nil {
			fmt.Printf("Warning: All password activation attempts failed: %v\n", err)
		}
	} else {
		fmt.Printf("Warning: No activation token found, attempting direct password update\n")
		// Try to update password directly if activation token is not available
		// This might require additional API calls depending on the Palette SDK
		fmt.Printf("  Note: Password may need to be set manually in Palette UI\n")
	}

	// Update progress: Generating API Keys
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Generating API Keys", "running", "Generating API keys...")
	}

	// Create API Key
	fmt.Printf("- Creating API key for user\n")
	body := &palettemodels.V1APIKeyEntity{
		Metadata: &palettemodels.V1ObjectMeta{
			Name:        fmt.Sprintf("lab-%s-api-key", shortID),
			Annotations: make(map[string]string),
		},
		Spec: &palettemodels.V1APIKeySpecEntity{
			UserUID: userID,
			Expiry:  palettemodels.V1Time(time.Now().Add(time.Duration(7 * 24 * time.Hour))),
		},
	}
	body.Metadata.Annotations["description"] = "Autogenerated Lab API Key"

	params := version1.NewV1APIKeysCreateParams().WithBody(body)
	resp, err := pc.Client.V1APIKeysCreate(params)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Generating API Keys", "failed", fmt.Sprintf("Failed to create API key: %v", err))
		}
		return fmt.Errorf("failed to create API key: %w", err)
	}
	fmt.Printf("  API key created: %s\n", resp.Payload.APIKey)

	// Update progress: Generating API Keys completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Generating API Keys", "completed", "API keys generated")
	}

	// Update progress: Creating Edge Tokens
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating Edge Tokens", "running", "Creating edge tokens...")
	}

	// Create Edge Token
	fmt.Printf("- Creating edge registration token\n")
	edgeEntity := &palettemodels.V1EdgeTokenEntity{
		Metadata: &palettemodels.V1ObjectMeta{
			Name: fmt.Sprintf("lab-%s", shortID),
		},
		Spec: &palettemodels.V1EdgeTokenSpecEntity{
			DefaultProjectUID: projectID,
			Expiry:            palettemodels.V1Time(time.Now().Add(time.Duration(7 * 24 * time.Hour))),
		},
	}
	edgeTokenParams := version1.NewV1EdgeTokensCreateParams().WithBody(edgeEntity)
	registrationTokenUid, err := pc.Client.V1EdgeTokensCreate(edgeTokenParams)
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating Edge Tokens", "failed", fmt.Sprintf("Failed to create edge token: %v", err))
		}
		return fmt.Errorf("failed to create edge token: %w", err)
	}

	// Get Edge Token details
	edgeTokenGet, err := pc.Client.V1EdgeTokensUIDGet(version1.NewV1EdgeTokensUIDGetParams().WithUID(*registrationTokenUid.Payload.UID))
	if err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Creating Edge Tokens", "failed", fmt.Sprintf("Failed to get edge token: %v", err))
		}
		return fmt.Errorf("failed to get edge token: %w", err)
	}
	fmt.Printf("  Edge token created: %s\n", edgeTokenGet.Payload.Spec.Token)

	// Update progress: Creating Edge Tokens completed
	if ctx.UpdateProgress != nil {
		ctx.UpdateProgress("Creating Edge Tokens", "completed", "Edge tokens created")
	}

	// Store lab-specific data in context for cleanup
	ctx.Context = context.WithValue(ctx.Context, "palette_project_sandbox_id", shortID)
	ctx.Context = context.WithValue(ctx.Context, "palette_project_id", projectID)
	ctx.Context = context.WithValue(ctx.Context, "palette_project_user_id", userID)
	ctx.Context = context.WithValue(ctx.Context, "palette_project_name", projectEntity.Metadata.Name)
	ctx.Context = context.WithValue(ctx.Context, "palette_project_user_email", userEntity.Spec.EmailID)
	ctx.Context = context.WithValue(ctx.Context, "palette_project_api_key_name", body.Metadata.Name)

	// Also store data in lab's ServiceData for persistent access during cleanup
	if ctx.Lab != nil {
		if ctx.Lab.ServiceData == nil {
			ctx.Lab.ServiceData = make(map[string]string)
		}
		ctx.Lab.ServiceData["palette_project_sandbox_id"] = shortID
		ctx.Lab.ServiceData["palette_project_id"] = projectID
		ctx.Lab.ServiceData["palette_project_user_id"] = userID
		ctx.Lab.ServiceData["palette_project_name"] = projectEntity.Metadata.Name
		ctx.Lab.ServiceData["palette_project_user_email"] = userEntity.Spec.EmailID
		ctx.Lab.ServiceData["palette_project_api_key_name"] = body.Metadata.Name
	}

	// Add credentials to the lab
	credential := &interfaces.Credential{
		ID:        uuid.New().String(),
		LabID:     ctx.LabID,
		Label:     "Palette Project",
		Username:  userEntity.Spec.EmailID,
		Password:  goodPassword,
		URL:       fmt.Sprintf("%s/login", v.host),
		ExpiresAt: time.Now().Add(time.Duration(ctx.Duration) * time.Minute),
		Notes: fmt.Sprintf("Spectro Cloud Project access. Project: %s, API Key: %s, Edge Token: %s",
			projectEntity.Metadata.Name, resp.Payload.APIKey, edgeTokenGet.Payload.Spec.Token),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ctx.AddCredential(credential); err != nil {
		if ctx.UpdateProgress != nil {
			ctx.UpdateProgress("Setting up User Account", "failed", fmt.Sprintf("Failed to add credential: %v", err))
		}
		return fmt.Errorf("failed to add Palette Project credential: %w", err)
	}

	fmt.Printf("Palette Project setup completed for lab %s\n", ctx.LabName)
	return nil
}

// ExecuteCleanup cleans up Palette Project resources
func (v *PaletteProjectService) ExecuteCleanup(ctx *interfaces.CleanupContext) error {
	// Validate required environment variables
	if v.host == "" || v.apiKey == "" {
		return fmt.Errorf("PALETTE_HOST and PALETTE_API_KEY environment variables are required")
	}

	// Use lab ID directly as it's already the short ID
	shortID := ctx.LabID

	// Get lab-specific data from context
	sandboxID, ok := ctx.Context.Value("palette_project_sandbox_id").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedSandboxID, exists := ctx.Lab.ServiceData["palette_project_sandbox_id"]; exists {
				sandboxID = storedSandboxID
				fmt.Printf("Retrieved sandbox ID from lab ServiceData: %s\n", sandboxID)
			} else {
				// If sandbox ID is not in context, use the short ID from lab name
				sandboxID = shortID
				fmt.Printf("Warning: palette project sandbox ID not found in context or lab data, using short ID from lab name: %s\n", sandboxID)
			}
		} else {
			// If sandbox ID is not in context, use the short ID from lab name
			sandboxID = shortID
			fmt.Printf("Warning: palette project sandbox ID not found in context, using short ID from lab name: %s\n", sandboxID)
		}
	}

	projectID, ok := ctx.Context.Value("palette_project_id").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedProjectID, exists := ctx.Lab.ServiceData["palette_project_id"]; exists {
				projectID = storedProjectID
				fmt.Printf("Retrieved project ID from lab ServiceData: %s\n", projectID)
			} else {
				// If project ID is not in context or lab data, we'll need to find it by name
				projectID = ""
				fmt.Printf("Warning: palette project ID not found in context or lab data, will search by project name\n")
			}
		} else {
			// If project ID is not in context, we'll need to find it by name
			projectID = ""
			fmt.Printf("Warning: palette project ID not found in context, will search by project name\n")
		}
	}

	userID, ok := ctx.Context.Value("palette_project_user_id").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedUserID, exists := ctx.Lab.ServiceData["palette_project_user_id"]; exists {
				userID = storedUserID
				fmt.Printf("Retrieved user ID from lab ServiceData: %s\n", userID)
			} else {
				// If user ID is not in context or lab data, we'll need to find it by email
				userID = ""
				fmt.Printf("Warning: palette project user ID not found in context or lab data, will search by email\n")
			}
		} else {
			// If user ID is not in context, we'll need to find it by email
			userID = ""
			fmt.Printf("Warning: palette project user ID not found in context, will search by email\n")
		}
	}

	projectName, ok := ctx.Context.Value("palette_project_name").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedProjectName, exists := ctx.Lab.ServiceData["palette_project_name"]; exists {
				projectName = storedProjectName
				fmt.Printf("Retrieved project name from lab ServiceData: %s\n", projectName)
			} else {
				projectName = fmt.Sprintf("lab-%s", sandboxID)
			}
		} else {
			projectName = fmt.Sprintf("lab-%s", sandboxID)
		}
	}

	userEmail, ok := ctx.Context.Value("palette_project_user_email").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedUserEmail, exists := ctx.Lab.ServiceData["palette_project_user_email"]; exists {
				userEmail = storedUserEmail
				fmt.Printf("Retrieved user email from lab ServiceData: %s\n", userEmail)
			} else {
				userEmail = fmt.Sprintf("lab+%s@spectrocloud.com", sandboxID)
			}
		} else {
			userEmail = fmt.Sprintf("lab+%s@spectrocloud.com", sandboxID)
		}
	}

	apiKeyName, ok := ctx.Context.Value("palette_project_api_key_name").(string)
	if !ok {
		// Try to get from lab's ServiceData
		if ctx.Lab != nil && ctx.Lab.ServiceData != nil {
			if storedApiKeyName, exists := ctx.Lab.ServiceData["palette_project_api_key_name"]; exists {
				apiKeyName = storedApiKeyName
				fmt.Printf("Retrieved API key name from lab ServiceData: %s\n", apiKeyName)
			} else {
				apiKeyName = fmt.Sprintf("lab-%s-api-key", sandboxID)
			}
		} else {
			apiKeyName = fmt.Sprintf("lab-%s-api-key", sandboxID)
		}
	}

	// Initialize Palette client
	pc := client.New(
		client.WithPaletteURI(v.host),
		client.WithAPIKey(v.apiKey),
	)

	// Set scope based on project UID
	if v.projectUID != "" {
		client.WithScopeProject(v.projectUID)(pc)
	} else {
		client.WithScopeTenant()(pc)
	}

	fmt.Printf("Cleaning up Palette Project resources for lab %s:\n", ctx.LabID)

	// If we don't have the project ID, try to find it by name
	if projectID == "" {
		fmt.Printf("- Searching for project by name: %s\n", projectName)
		foundProjectID, err := pc.GetProjectUID(projectName)
		if err != nil {
			fmt.Printf("Warning: Could not find project by name %s: %v\n", projectName, err)
			// Continue with cleanup using the name pattern
		} else {
			projectID = foundProjectID
			fmt.Printf("  Found project ID: %s\n", projectID)
		}
	}

	// If we don't have the user ID, try to find it by email
	if userID == "" {
		fmt.Printf("- Searching for user by email: %s\n", userEmail)
		user, err := pc.GetUserByEmail(userEmail)
		if err != nil {
			fmt.Printf("Warning: Could not find user by email %s: %v\n", userEmail, err)
		} else {
			userID = user.Metadata.UID
			fmt.Printf("  Found user ID: %s\n", userID)
		}
	}

	// Delete API Key
	fmt.Printf("- Deleting API key: %s\n", apiKeyName)
	if err := pc.DeleteAPIKeyByName(apiKeyName); err != nil {
		fmt.Printf("Warning: Failed to delete API key: %v\n", err)
	}

	// Delete User (only if we found the user ID)
	if userID != "" {
		fmt.Printf("- Deleting user: %s (ID: %s)\n", userEmail, userID)
		if err := pc.DeleteUser(userID); err != nil {
			fmt.Printf("Warning: Failed to delete user: %v\n", err)
		}
	} else {
		fmt.Printf("- Skipping user deletion (user ID not found)\n")
	}

	// Switch to project scope for cleanup (only if we have a project ID)
	if projectID != "" {
		client.WithScopeProject(projectID)(pc)

		// Clean up clusters
		fmt.Printf("- Cleaning up clusters in project: %s\n", projectName)
		edgeClusters, err := pc.GetClusterGroupSummaries()
		if err != nil {
			fmt.Printf("Warning: Failed to get clusters: %v\n", err)
		} else {
			for _, cluster := range edgeClusters {
				fmt.Printf("  Deleting cluster: %s\n", cluster.Metadata.UID)
				if err = pc.ForceDeleteCluster(cluster.Metadata.UID, true); err != nil {
					fmt.Printf("Warning: Failed to delete cluster %s: %v\n", cluster.Metadata.UID, err)
				}
			}
		}

		// Clean up edge devices
		fmt.Printf("- Cleaning up edge devices in project: %s\n", projectName)
		edgeDevices, err := pc.ListEdgeHosts()
		if err != nil {
			fmt.Printf("Warning: Failed to get edge devices: %v\n", err)
		} else {
			for _, edgeDevice := range edgeDevices {
				fmt.Printf("  Deleting edge device: %s\n", edgeDevice.Metadata.UID)
				if err = pc.DeleteAppliance(edgeDevice.Metadata.UID); err != nil {
					fmt.Printf("Warning: Failed to delete edge device %s: %v\n", edgeDevice.Metadata.UID, err)
				}
			}
		}

		// Clean up registration tokens
		fmt.Printf("- Cleaning up registration tokens for project: %s\n", projectName)
		params := &version1.V1EdgeTokensListParams{}
		edgeTokens, err := pc.Client.V1EdgeTokensList(params)
		if err != nil {
			fmt.Printf("Warning: Failed to get edge tokens: %v\n", err)
		} else {
			for _, token := range edgeTokens.Payload.Items {
				if token.Spec.DefaultProject.UID == projectID {
					fmt.Printf("  Deleting registration token: %s\n", token.Metadata.UID)
					if err = pc.DeleteRegistrationToken(token.Metadata.UID); err != nil {
						fmt.Printf("Warning: Failed to delete registration token %s: %v\n", token.Metadata.UID, err)
					}
				}
			}
		}

		// Switch back to tenant scope for project deletion
		if v.projectUID != "" {
			client.WithScopeProject(v.projectUID)(pc)
		} else {
			client.WithScopeTenant()(pc)
		}

		// Delete Project
		fmt.Printf("- Deleting project: %s (ID: %s)\n", projectName, projectID)
		if err = pc.DeleteProject(projectID); err != nil {
			fmt.Printf("Warning: Failed to delete project: %v\n", err)
		}
	} else {
		fmt.Printf("- Skipping project cleanup (project ID not found)\n")
	}

	fmt.Printf("Palette Project cleanup completed for lab %s\n", sandboxID)
	return nil
}
