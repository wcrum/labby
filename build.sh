#!/bin/bash

# Simple build script for Spectro Lab
set -e

echo "🎨 Building frontend..."
npm run build

echo "📁 Copying frontend to backend..."
rm -rf backend/static
cp -r out backend/static

echo "🔨 Building backend..."
cd backend
go build -o server ./cmd/server

echo "✅ Build complete! Run with: cd backend && ./server"
