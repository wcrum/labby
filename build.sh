#!/bin/bash

# Build and run spectro-lab in Docker

set -e

echo "Building spectro-lab Docker image..."

# Build the Docker image
docker build -t spectro-lab .

echo "Docker image built successfully!"
echo ""
echo "To run the application:"
echo "  docker run -p 8080:8080 spectro-lab"
echo ""
echo "Or use docker-compose:"
echo "  docker-compose up"
echo ""
echo "The application will be available at: http://localhost:8080"
