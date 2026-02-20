# Multi-stage build for spectro-lab application
FROM node:20-alpine AS frontend-builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json pnpm-lock.yaml ./

# Install pnpm
RUN npm install -g pnpm

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy frontend source code
COPY . .

# Build the frontend
RUN pnpm run build

# Go backend build stage
FROM golang:1.24-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy .netrc for private repository access
COPY .netrc /root/.netrc
RUN chmod 600 /root/.netrc

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# TEMPORARY: Copy local palette-sdk-go-internal dependency
# This is a temporary workaround until upstream changes are made to the official palette-sdk-go-internal package.
# The go.mod file contains a replace directive pointing to ../palette-sdk-go-internal
# TODO: Remove this temporary copy once upstream changes are available
COPY ../palette-sdk-go-internal /palette-sdk-go-internal

# Download dependencies
RUN go mod download

# Copy backend source code
COPY backend/ .

# Build the backend binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create app user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nextjs -u 1001

# Set working directory
WORKDIR /app

# Copy the backend binary from the builder stage
COPY --from=backend-builder /app/main .

# Copy the frontend build from the frontend-builder stage
COPY --from=frontend-builder /app/out ./static

# Copy backend templates
COPY backend/templates ./templates

COPY backend/service-configs ./service-configs

# Copy environment example (optional, for reference)
COPY backend/env.example ./

# Change ownership to the nextjs user
RUN chown -R nextjs:nodejs /app

# Switch to non-root user
USER nextjs

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV JWT_SECRET=your-secret-key-change-in-production

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
