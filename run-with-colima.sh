#!/bin/bash

echo "🐳 Starting BookRank with Colima"
echo "==============================="

# Start colima if not running
echo "📦 Starting Colima..."
colima start 2>/dev/null || echo "Colima already running"

# Create network
echo "🌐 Creating network..."
docker network create bookrank-network 2>/dev/null || echo "Network already exists"

# Start PostgreSQL
echo "🗄️ Starting PostgreSQL..."
docker run -d \
  --name bookrank-postgres \
  --network bookrank-network \
  -e POSTGRES_DB=bookrank \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:16-alpine

# Start Redis
echo "🔴 Starting Redis..."
docker run -d \
  --name bookrank-redis \
  --network bookrank-network \
  -p 6379:6379 \
  redis:7-alpine

# Wait for databases to be ready
echo "⏳ Waiting for databases to start..."
sleep 10

# Set environment variables
export JWT_SECRET=development-secret-change-in-production
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=bookrank
export DB_SSL_MODE=disable
export REDIS_URL=redis://localhost:6379
export LOG_LEVEL=info

echo "🚀 Starting BookRank application..."
echo "Application will be available at: http://localhost:8080"
echo "Health check: http://localhost:8080/health"
echo ""
echo "Press Ctrl+C to stop"

# Build and run the application
make build && ./bookrank