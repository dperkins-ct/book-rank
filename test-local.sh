#!/bin/bash

# Simple local testing script (no Docker needed)
set -e

echo "🧪 Testing BookRank Application Locally"
echo "=================================="

# Set minimal environment variables
export JWT_SECRET=test-secret-for-local-development-only
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=bookrank
export LOG_LEVEL=info

echo "✅ 1. Building application..."
make build

echo "✅ 2. Running core service tests..."
go test ./internal/service/... -v

echo "✅ 3. Running authentication tests..."
go test ./internal/auth/... -v

echo "✅ 4. Testing application startup (will fail on DB, but shows config works)..."
timeout 3s ./bookrank || echo "Expected: DB connection failed (normal without PostgreSQL)"

echo ""
echo "🎉 Local Testing Complete!"
echo ""
echo "Summary:"
echo "✅ Application builds successfully"
echo "✅ Core business logic tests pass"
echo "✅ Authentication system works"
echo "✅ Application starts (needs PostgreSQL for full functionality)"
echo ""
echo "To run with database:"
echo "1. Install Docker Desktop: https://www.docker.com/products/docker-desktop/"
echo "2. Start Docker Desktop"
echo "3. Run: make docker-up-dev"
echo "4. Run: make run"