# Spectro Lab Backend

A Golang-based backend service for managing lab sessions with authentication and role-based access control.

## Features

- **Authentication**: JWT-based authentication with dummy user management
- **Lab Management**: Create, manage, and cleanup lab sessions
- **Role-Based Access Control**: Separate admin and user roles
- **Credential Management**: Generate credentials for lab services
- **Service Integration**: Modular service architecture for different lab environments
- **Automatic Cleanup**: Background cleanup of expired labs and resources

## Services

### Palette Project Service

The Palette Project service provides access to Spectro Cloud control plane functionality. It creates isolated projects and users for each lab session.

**Setup Process:**
1. Creates a new project with lab-specific naming
2. Creates a new user with lab-specific email
3. Assigns Project Admin role to the user
4. Generates secure password for the user
5. Creates API key for programmatic access
6. Creates edge registration token for device registration
7. Adds all credentials to the lab session

**Cleanup Process:**
1. Deletes API keys associated with the lab
2. Deletes the lab user
3. Cleans up any clusters in the project
4. Cleans up any edge devices in the project
5. Deletes registration tokens for the project
6. Deletes the project itself

**Required Environment Variables:**
- `PALETTE_HOST`: The URL of your Palette instance
- `PALETTE_API_KEY`: API key with tenant-level access
- `PALETTE_PROJECT_UID`: (Optional) Specific project UID for scoped access

## API Endpoints

### Authentication
- `POST /api/auth/login` - User login

### Lab Management
- `POST /api/labs` - Create a new lab
- `GET /api/labs/:id` - Get lab details
- `GET /api/labs` - Get user's labs
- `DELETE /api/labs/:id` - Delete a lab
- `POST /api/labs/:id/stop` - Stop a lab


### Admin Endpoints
- `GET /api/admin/labs` - Get all labs
- `GET /api/admin/users` - Get all users
- `POST /api/admin/users` - Create a user
- `PUT /api/admin/users/:id/role` - Update user role
- `DELETE /api/admin/users/:id` - Delete a user

### Health Check
- `GET /health` - Health check endpoint

## Setup

1. Copy `env.example` to `.env` and configure your environment variables
2. Run `go mod tidy` to install dependencies
3. Run `go run cmd/server/main.go` to start the server

## Architecture

The backend uses a modular service architecture where different lab environments (Palette Project, Proxmox, Kubernetes) are implemented as separate services that can be registered and managed through a service registry. Each service implements setup and cleanup operations that are called during lab creation and deletion.
