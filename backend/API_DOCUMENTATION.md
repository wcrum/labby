# Spectro Lab API Documentation

## Overview

The Spectro Lab API provides a RESTful interface for managing lab environments. The API is documented using OpenAPI 3.0 specification and can be accessed through Swagger UI.

## Accessing the API Documentation

### Swagger UI

Once the server is running, you can access the interactive API documentation at:

```
http://localhost:8080/swagger/index.html
```

This provides a user-friendly interface where you can:
- Browse all available endpoints
- See request/response schemas
- Test API calls directly from the browser
- View authentication requirements

### OpenAPI Specification Files

The API specification is also available in different formats:

- **JSON**: `http://localhost:8080/swagger/doc.json`
- **YAML**: `http://localhost:8080/swagger/doc.yaml`

## Authentication

Most API endpoints require authentication using JWT tokens. To authenticate:

1. Use the `/api/auth/login` endpoint to get a JWT token
2. Include the token in the `Authorization` header as `Bearer <token>`

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login and get JWT token
- `GET /api/auth/me` - Get current user information

### Labs
- `GET /api/labs` - Get user's labs
- `GET /api/labs/{id}` - Get specific lab
- `GET /api/labs/{id}/progress` - Get lab progress
- `DELETE /api/labs/{id}` - Delete lab
- `POST /api/labs/{id}/stop` - Stop lab

### Templates
- `GET /api/templates` - Get all lab templates
- `GET /api/templates/{id}` - Get specific template
- `POST /api/templates/{id}/labs` - Create lab from template

### Admin Endpoints
- `GET /api/admin/labs` - Get all labs (admin only)
- `POST /api/admin/labs/{id}/stop` - Stop any lab (admin only)
- `DELETE /api/admin/labs/{id}` - Delete any lab (admin only)
- `POST /api/admin/labs/{id}/cleanup` - Cleanup lab resources (admin only)
- `GET /api/admin/users` - Get all users (admin only)
- `POST /api/admin/users` - Create user (admin only)
- `PUT /api/admin/users/{id}/role` - Update user role (admin only)
- `DELETE /api/admin/users/{id}` - Delete user (admin only)
- `POST /api/admin/palette-project/cleanup` - Cleanup palette project (admin only)

### System
- `GET /health` - Health check

## Regenerating Documentation

If you make changes to the API (add new endpoints, modify request/response models, etc.), you need to regenerate the Swagger documentation:

```bash
./generate-swagger.sh
```

Or manually:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
swag init -g cmd/server/main.go
```

## Development

### Adding New Endpoints

When adding new endpoints, make sure to include Swagger annotations:

```go
// @Summary Endpoint summary
// @Description Detailed description
// @Tags tag-name
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Parameter description"
// @Success 200 {object} ResponseType
// @Failure 400 {object} map[string]interface{} "Error description"
// @Router /endpoint/path [method]
func (h *Handler) YourHandler(c *gin.Context) {
    // Handler implementation
}
```

### Available Tags

- `auth` - Authentication endpoints
- `labs` - Lab management endpoints
- `templates` - Template management endpoints
- `admin` - Admin-only endpoints
- `system` - System endpoints

## Testing the API

You can test the API using:

1. **Swagger UI**: Use the interactive interface at `/swagger/index.html`
2. **curl**: Use command line tools
3. **Postman**: Import the OpenAPI specification
4. **Frontend**: Use the provided frontend application

## Error Responses

The API returns consistent error responses:

```json
{
  "error": "Error message description"
}
```

Common HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error
