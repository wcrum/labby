#!/bin/bash

# Generate Swagger documentation for the Spectro Lab API

echo "Generating Swagger documentation..."

# Generate swagger docs using go run (works even if swag is not in PATH)
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs

echo "Swagger documentation generated successfully!"
echo "Files created:"
echo "  - docs/docs.go"
echo "  - docs/swagger.json"
echo "  - docs/swagger.yaml"
