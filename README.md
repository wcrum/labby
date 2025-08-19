# Spectro Lab

A full-stack lab management application with Go backend and Next.js frontend, featuring real-time lab session management, credential handling, and admin dashboard.

## 🚀 Quick Start

### Option 1: Simple Runner (Recommended)

Run everything with a single command:

```bash
go run simple-run.go
```

This will automatically:
- Build the frontend with `npm run build`
- Copy the built frontend to the backend
- Start the Go server on `http://localhost:8080`

### Option 2: Manual Steps

```bash
# Build frontend
npm run build

# Copy to backend
cp -r out backend/static

# Run server
cd backend && go run cmd/server/main.go
```

### Option 3: Development Mode

For development with hot reloading:

```bash
# Terminal 1: Backend
cd backend && go run cmd/server/main.go

# Terminal 2: Frontend
npm run dev
```

## 📋 Features

- **Lab Management**: Create, monitor, and manage lab sessions
- **Real-time Updates**: Live countdown timers and status updates
- **Credential Management**: Secure display and management of lab credentials
- **Admin Dashboard**: Comprehensive admin interface for managing all labs
- **Authentication**: JWT-based authentication with role-based access
- **API Integration**: RESTful API for lab operations

## 🏗️ Architecture

- **Frontend**: Next.js 15 with TypeScript, Tailwind CSS, and Radix UI
- **Backend**: Go with Gin framework
- **Authentication**: JWT tokens with role-based access control
- **Database**: In-memory storage (can be extended to persistent storage)
- **Service Architecture**: Plugin-based service system with lifecycle management

### Service Interface System

The application uses a sophisticated service interface system for managing lab resources. This allows for easy extension and pluggable services.

#### Core Interfaces

The system is built around several key interfaces in `backend/internal/interfaces/lifecycle.go`:

```go
// Setup defines the contract for setup actions
type Setup interface {
    ExecuteSetup(ctx *SetupContext) error
    Name() string
}

// Cleanup defines the contract for cleanup actions  
type Cleanup interface {
    ExecuteCleanup(ctx *CleanupContext) error
    Name() string
}

// Service represents a complete service lifecycle
type Service interface {
    Setup
    Cleanup
    GetName() string
    GetDescription() string
    GetRequiredParams() []string
}
```

#### How Services Work

1. **Service Registration**: Services are registered in the `ServiceManager` (see `backend/internal/services/manager.go`)
2. **Lab Provisioning**: When a lab is created, all registered services are executed during setup
3. **Resource Management**: Services can add credentials and store persistent data
4. **Cleanup**: When labs are stopped or deleted, services handle resource cleanup

#### Example Service Implementation

The `PaletteProjectService` demonstrates the pattern:

```go
type PaletteProjectService struct {
    host       string
    apiKey     string
    projectUID string
}

// Setup creates Spectro Cloud projects, users, and credentials
func (v *PaletteProjectService) ExecuteSetup(ctx *SetupContext) error {
    // 1. Create project in Spectro Cloud
    // 2. Create user account
    // 3. Generate API keys
    // 4. Add credentials via ctx.AddCredential()
    // 5. Store service data in ctx.Lab.ServiceData
}

// Cleanup removes all created resources
func (v *PaletteProjectService) ExecuteCleanup(ctx *CleanupContext) error {
    // 1. Delete API keys
    // 2. Remove user accounts  
    // 3. Delete projects
    // 4. Clean up any remaining resources
}
```

#### Context Objects

**SetupContext** provides:
- Lab metadata (ID, name, duration, owner)
- `AddCredential()` function to add lab credentials
- Lab reference for storing service-specific data
- Go context for cancellation

**CleanupContext** provides:
- Lab metadata for identifying resources
- Lab reference to access stored service data
- Go context for operations

#### Adding New Services

To add a new service:

1. **Implement the Service interface**:
   ```go
   type MyCustomService struct {
       // service configuration
   }
   
   func (s *MyCustomService) ExecuteSetup(ctx *SetupContext) error {
       // Setup logic
   }
   
   func (s *MyCustomService) ExecuteCleanup(ctx *CleanupContext) error {
       // Cleanup logic
   }
   ```

2. **Register in ServiceManager**:
   ```go
   // In services/manager.go
   registry.RegisterService(NewMyCustomService())
   ```

3. **Configure Environment Variables**: Add any required configuration to your environment

#### Service Data Persistence

Services can store persistent data in `lab.ServiceData` map for use during cleanup:

```go
// During setup
ctx.Lab.ServiceData = map[string]string{
    "my_service_resource_id": "created-resource-123",
    "my_service_api_key": "generated-key-456",
}

// During cleanup - retrieve stored data
resourceID := ctx.Lab.ServiceData["my_service_resource_id"]
```

This architecture enables the lab system to support any cloud provider or service by implementing the standard interfaces.

## 🔧 Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `JWT_SECRET`: JWT signing secret
- `NEXT_PUBLIC_API_URL`: Frontend API URL

### Default Admin User

- **Email**: admin@spectrocloud.com
- **Role**: admin

## 📚 Documentation

- Simple runner approach with integrated static file serving
- Real-time admin dashboard with live lab monitoring

## 🛠️ Development

### Prerequisites

- Go 1.19+
- Node.js 18+
- npm

### Project Structure

```
spectro-lab/
├── backend/           # Go backend
│   ├── cmd/server/   # Main server binary
│   ├── internal/     # Internal packages
│   └── static/       # Frontend build output
├── src/              # Next.js frontend
│   ├── app/          # App router pages
│   ├── components/   # React components
│   └── lib/          # Utilities and API client
├── simple-run.go     # Simple build and run script
└── build.sh         # Build script
```

## 🚀 Deployment

### Production Build

```bash
# Build and run production version
go run simple-run.go

# Or build manually
./build.sh
cd backend && ./server
```

The application will be available at `http://localhost:8080` with:
- Frontend: `http://localhost:8080/`
- API: `http://localhost:8080/api/`
- Health: `http://localhost:8080/health`

### API Endpoints

**User Endpoints:**
- `GET /health` - Health check
- `POST /api/auth/login` - User login
- `GET /api/labs` - Get user's labs
- `POST /api/labs` - Create new lab
- `DELETE /api/labs/{id}` - Delete user's own lab
- `POST /api/labs/{id}/stop` - Stop user's own lab

**Admin Endpoints** (require admin privileges):
- `GET /api/admin/labs` - Get all labs
- `POST /api/admin/labs/{id}/stop` - Stop any lab
- `DELETE /api/admin/labs/{id}` - Delete any lab
- `POST /api/admin/labs/{id}/cleanup` - Cleanup lab resources
- `POST /api/admin/palette-project/cleanup` - Cleanup specific palette projects
- `GET /api/admin/users` - Get all users
- `POST /api/admin/users` - Create user
- `PUT /api/admin/users/{id}/role` - Update user role
- `DELETE /api/admin/users/{id}` - Delete user

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.
