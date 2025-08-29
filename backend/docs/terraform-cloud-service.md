# Terraform Cloud Service

The Terraform Cloud service allows you to create and manage Terraform Cloud workspaces for infrastructure provisioning without requiring GitHub-stored folders.

## Overview

This service integrates with Terraform Cloud's API to:
1. Create new workspaces
2. Upload Terraform configurations directly
3. Trigger Terraform runs
4. Monitor run status
5. Clean up resources when labs expire

## Configuration

### Service Configuration

Create a service configuration in `backend/service-configs/terraform-cloud.yaml`:

```yaml
id: "terraform-cloud-1"
name: "Terraform Cloud Workspace"
type: "terraform_cloud"
description: "Terraform Cloud workspace for infrastructure provisioning"
config:
  host: "https://app.terraform.io"
  api_token: "your-terraform-cloud-api-token"
  organization: "your-organization"
  # Terraform configuration settings
  terraform_config:
    source_directory: "spacewalk/bm-maas-connected-pcg"
    variables:
      pm_api_url: "https://172.16.10.101:8006/api2/json"
      pm_node: "swlk-prxmx03"
      template_name: "ubuntu-24.04-cloudinit"
      storage_pool: "data"
      network_bridge: "vmbr0"
      vm_user: "ubuntu"
      vm_password: "Spectro2025!"
      ubuntu_iso: "local:iso/ubuntu-24.04.2-live-server-amd64.iso"
      resource_pool: "lab-pool"
      # Sensitive variables (will be set as sensitive in Terraform Cloud)
      sensitive_variables:
        pm_api_token_id: "your-proxmox-token-id"
        pm_api_token_secret: "your-proxmox-token-secret"
        ssh_key: "your-ssh-public-key"
is_active: true
created_at: "2025-01-18T23:28:50Z"
updated_at: "2025-01-18T23:28:50Z"
```

### Environment Variables

The service uses these environment variables (set from service config):

- `TF_CLOUD_HOST`: Terraform Cloud host URL (default: https://app.terraform.io)
- `TF_CLOUD_API_TOKEN`: Your Terraform Cloud API token
- `TF_CLOUD_ORGANIZATION`: Your Terraform Cloud organization name

## Usage

### Basic Usage

1. **Create a Lab Template** that includes the Terraform Cloud service:

```yaml
name: "Terraform Cloud Lab"
id: "terraform-cloud-lab"
description: "A lab that creates a Terraform Cloud workspace"
expiration_duration: "4h"
owner: "admin@spectrocloud.com"
services:
  - name: "Terraform Cloud Workspace"
    service_id: "terraform-cloud-1"
    description: "Terraform Cloud workspace for infrastructure provisioning"
```

2. **Start a Lab** using the template. The service will:
   - Create a new workspace named `lab-{shortID}`
   - Upload a default Terraform configuration (AWS EC2 instance)
   - Trigger a Terraform run
   - Provide credentials with workspace access

### Using Spacewalk Terraform Configurations

The service can automatically load and templatize Terraform configurations from the `spacewalk` directory:

#### 1. **Configure Source Directory**

In your service configuration, specify which Terraform configuration to use:

```yaml
terraform_config:
  source_directory: "spacewalk/bm-maas-connected-pcg"  # Path relative to spacewalk/
```

#### 2. **Define Variables**

Set variables that will be templatized into your Terraform configuration:

```yaml
variables:
  pm_api_url: "https://172.16.10.101:8006/api2/json"
  pm_node: "swlk-prxmx03"
  template_name: "ubuntu-24.04-cloudinit"
  storage_pool: "data"
  network_bridge: "vmbr0"
  vm_user: "ubuntu"
  vm_password: "Spectro2025!"
  ubuntu_iso: "local:iso/ubuntu-24.04.2-live-server-amd64.iso"
  resource_pool: "lab-pool"
```

#### 3. **Handle Sensitive Variables**

Store sensitive information securely:

```yaml
sensitive_variables:
  pm_api_token_id: "your-proxmox-token-id"
  pm_api_token_secret: "your-proxmox-token-secret"
  ssh_key: "your-ssh-public-key"
```

#### 4. **Variable Templating**

The service automatically replaces variables in your Terraform files:

- `${lab_id}` → Short lab ID (e.g., "abc123")
- `${lab_name}` → Full lab name (e.g., "lab-abc123")
- `${lab_owner}` → Lab owner ID
- `${variable_name}` → Any variable from your service config

**Example terraform.tfvars templating:**
```hcl
pm_api_url           = "https://172.16.10.101:8006/api2/json"
pm_node              = "swlk-prxmx03"
template_name        = "ubuntu-24.04-cloudinit"
storage_pool         = "data"
network_bridge       = "vmbr0"
vm_user              = "ubuntu"
vm_password          = "Spectro2025!"
ubuntu_iso           = "local:iso/ubuntu-24.04.2-live-server-amd64.iso"
lab_id               = ${lab_id}  # Will be replaced with actual lab ID
resource_pool        = "lab-${lab_id}-pool"  # Will be templatized
```

### Custom Terraform Configurations

You can also upload custom Terraform configurations programmatically:

```go
// Example: Upload custom configuration
terraformService := services.NewTerraformCloudService()

customConfig := map[string]string{
    "main.tf": `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
  
  tags = {
    Name = "Lab VPC"
  }
}
`,
    "variables.tf": `
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "lab"
}
`,
}

err := terraformService.UploadCustomConfiguration(workspaceID, customConfig)
```

## API Integration

The service uses Terraform Cloud's REST API v2:

### Workspace Management
- **Create Workspace**: `POST /api/v2/organizations/{org}/workspaces`
- **Delete Workspace**: `DELETE /api/v2/workspaces/{id}`

### Configuration Management
- **Create Configuration Version**: `POST /api/v2/workspaces/{id}/configuration-versions`
- **Upload Configuration**: `PUT {upload-url}` (multipart form data)

### Run Management
- **Trigger Run**: `POST /api/v2/runs`
- **Get Run Status**: `GET /api/v2/runs/{id}`

## Default Configuration

The service includes a default Terraform configuration that creates:
- AWS EC2 instance (t2.micro)
- Basic networking
- Outputs for instance ID and public IP

## Credentials

When a lab is created, the service adds credentials including:
- **Workspace URL**: Direct link to the Terraform Cloud workspace
- **API Token**: For programmatic access
- **Workspace ID**: For API operations
- **Organization**: For reference

## Cleanup

When a lab expires or is deleted:
1. The service retrieves the workspace ID from stored lab data
2. Deletes the workspace and all associated resources
3. Cleans up any Terraform-managed infrastructure

## Monitoring

The service provides progress updates during:
- Workspace creation
- Configuration upload
- Run triggering
- Cleanup operations

## Error Handling

The service handles common errors:
- Invalid API tokens
- Organization not found
- Workspace creation failures
- Configuration upload issues
- Run trigger failures

## Security Considerations

- API tokens are stored securely in lab credentials
- Workspaces are isolated per lab
- All resources are tagged with lab information
- Automatic cleanup prevents resource leaks

## Examples

### Proxmox Lab Configuration

For a Proxmox-based lab, you might upload:

```hcl
terraform {
  required_providers {
    proxmox = {
      source  = "telmate/proxmox"
      version = "~> 2.9"
    }
  }
}

provider "proxmox" {
  pm_api_url          = var.proxmox_url
  pm_api_token_id     = var.proxmox_token_id
  pm_api_token_secret = var.proxmox_token_secret
  pm_tls_insecure     = true
}

resource "proxmox_vm_qemu" "lab_vm" {
  name        = "lab-${var.lab_id}-vm"
  target_node = var.proxmox_node
  clone       = var.template_name
  full_clone  = true
  
  cores   = 2
  memory  = 4096
  agent   = 1
  
  network {
    bridge = "vmbr0"
  }
  
  disk {
    storage = "local-lvm"
    size    = "20G"
  }
}
```

### Multi-Cloud Configuration

For a multi-cloud lab:

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

# AWS Resources
provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "aws_vm" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"
}

# Azure Resources
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "rg" {
  name     = "lab-rg"
  location = "West US 2"
}
```

## Troubleshooting

### Common Issues

1. **API Token Invalid**
   - Verify the token has appropriate permissions
   - Check organization access

2. **Workspace Creation Fails**
   - Ensure organization exists
   - Check workspace naming conflicts

3. **Configuration Upload Fails**
   - Verify file format and content
   - Check upload URL validity

4. **Run Trigger Fails**
   - Ensure configuration is valid
   - Check workspace state

### Debug Information

The service logs detailed information including:
- API request/response details
- Workspace and run IDs
- Configuration upload status
- Error messages and stack traces
