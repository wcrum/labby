#!/bin/bash

# Generate Swagger documentation for the Spectro Lab API
echo "Generating Swagger documentation..."

# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Generate the docs
swag init -g cmd/server/main.go

echo "Swagger documentation generated successfully!"
echo "You can now access the Swagger UI at: http://localhost:8080/swagger/index.html"
