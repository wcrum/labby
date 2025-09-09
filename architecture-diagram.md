# Spectro Lab Architecture Diagram (Simplified)

```mermaid
graph TB
    %% Frontend
    subgraph "Frontend (Next.js)"
        UI[Web Interface]
        Auth[Authentication]
    end

    %% Backend
    subgraph "Backend (Go/Gin)"
        API[REST API]
        Handlers[Request Handlers]
        Services[Business Logic]
    end

    %% External Services
    subgraph "External Services"
        Palette[Spectro Cloud]
        Proxmox[Proxmox]
        Terraform[Terraform Cloud]
        Guacamole[Guacamole]
    end

    %% Database
    subgraph "Database (PostgreSQL)"
        DB[(PostgreSQL)]
    end

    %% Lab Lifecycle
    subgraph "Lab Lifecycle"
        Create[Create Lab]
        Provision[Provision Services]
        Manage[Manage Lab]
        Cleanup[Cleanup Resources]
        
        Create --> Provision
        Provision --> Manage
        Manage --> Cleanup
    end

    %% Main Flow
    UI -->|HTTP Requests| API
    API --> Handlers
    Handlers --> Services
    Services --> DB
    Services --> Palette
    Services --> Proxmox
    Services --> Terraform
    Services --> Guacamole
    
    Services --> Create
    Services --> Provision
    Services --> Manage
    Services --> Cleanup

    %% Styling
    classDef frontend fill:#e1f5fe
    classDef backend fill:#f3e5f5
    classDef database fill:#e8f5e8
    classDef external fill:#fff3e0
    classDef lifecycle fill:#fce4ec
    
    class UI,Auth frontend
    class API,Handlers,Services backend
    class DB database
    class Palette,Proxmox,Terraform,Guacamole external
    class Create,Provision,Manage,Cleanup lifecycle
```

## Simplified Architecture Overview

### **Frontend (Next.js)**
- **Web Interface**: React-based UI for lab management
- **Authentication**: User login and session management

### **Backend (Go/Gin)**
- **REST API**: HTTP endpoints for all operations
- **Request Handlers**: Process incoming requests
- **Business Logic**: Core application logic and orchestration

### **External Services**
- **Spectro Cloud**: Cloud platform management
- **Proxmox**: Infrastructure virtualization
- **Terraform Cloud**: Infrastructure as code
- **Guacamole**: Remote desktop access

### **Database (PostgreSQL)**
- Stores users, labs, credentials, and configurations

### **Lab Lifecycle**
1. **Create Lab**: User creates a new lab session
2. **Provision Services**: System sets up external services
3. **Manage Lab**: User works with the lab environment
4. **Cleanup Resources**: System cleans up when lab expires

## Key Features

- **Template-based Labs**: Predefined lab configurations
- **Multi-service Integration**: Automatically provisions multiple external services
- **Real-time Progress**: Live updates during lab setup
- **Automatic Cleanup**: Resources are cleaned up when labs expire
- **User Management**: Authentication and organization support
- **Admin Controls**: Administrative interface for system management
