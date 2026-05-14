#!/bin/bash

# BookRank Frontend Deployment Script

set -e

echo "🚀 Starting BookRank Frontend Deployment..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ .env file not found. Please create one based on .env.example"
    exit 1
fi

# Install dependencies
echo "📦 Installing dependencies..."
npm ci

# Run linting
echo "🔍 Running linter..."
npm run lint

# Build the application
echo "🏗️  Building application..."
npm run build

# Check if build was successful
if [ ! -d "dist" ]; then
    echo "❌ Build failed - dist directory not found"
    exit 1
fi

echo "✅ Build completed successfully!"

# Optional: Deploy to specific environments
case "${DEPLOY_ENV:-production}" in
    "staging")
        echo "🚀 Deploying to staging..."
        # Add staging deployment commands here
        # e.g., rsync, scp, or cloud deployment commands
        ;;
    "production")
        echo "🚀 Deploying to production..."
        # Add production deployment commands here
        # e.g., rsync, scp, or cloud deployment commands
        ;;
    *)
        echo "ℹ️  Build ready for deployment. Files are in ./dist/"
        echo "   Set DEPLOY_ENV=staging or DEPLOY_ENV=production to auto-deploy"
        ;;
esac

echo "🎉 Deployment process completed!"

# Print deployment info
echo ""
echo "📋 Deployment Summary:"
echo "   - Environment: ${DEPLOY_ENV:-local}"
echo "   - Build directory: ./dist/"
echo "   - API URL: ${VITE_API_URL:-http://localhost:8080}"
echo ""
echo "🔧 Next steps:"
echo "   1. Upload ./dist/ contents to your web server"
echo "   2. Configure web server to serve index.html for all routes"
echo "   3. Ensure API server is running and accessible"
echo "   4. Test the application in your deployment environment"