#!/bin/bash

# Development setup script for BookRank

set -e

echo "🚀 Setting up BookRank development environment..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
MIN_VERSION="1.21"
if [[ "$(printf '%s\n' "$MIN_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$MIN_VERSION" ]]; then
    echo "❌ Go version $GO_VERSION is installed, but version $MIN_VERSION or later is required."
    exit 1
fi

echo "✅ Go version $GO_VERSION is installed"

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.example..."
    cp .env.example .env
    echo "✅ .env file created. Please review and update the configuration."
else
    echo "✅ .env file already exists"
fi

# Download Go dependencies
echo "📦 Downloading Go dependencies..."
go mod download
go mod tidy

# Build the application
echo "🔨 Building the application..."
go build -o bookrank cmd/server/main.go

echo "✅ Build successful!"

# Check if Docker is available
if command -v docker &> /dev/null; then
    echo "🐳 Docker is available"

    if command -v docker-compose &> /dev/null || docker compose version &> /dev/null; then
        echo "📋 Docker Compose is available"
        echo "💡 You can start the database with: make docker-up-dev"
        echo "💡 Or use Docker Compose directly: docker compose --profile dev up -d"
    else
        echo "⚠️  Docker Compose not found. Please install Docker Compose for database setup."
    fi
else
    echo "⚠️  Docker not found. You'll need to set up PostgreSQL manually."
    echo "💡 Database connection string: postgres://postgres:postgres@localhost:5432/bookrank?sslmode=disable"
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "📝 Next steps:"
echo "   1. Review the .env file and update configuration as needed"
echo "   2. Start the database: make docker-up-dev (if Docker is available)"
echo "   3. Run the application: make run"
echo "   4. Visit http://localhost:8080/health to check if it's running"
echo ""
echo "📚 Available commands:"
echo "   make help       - Show all available commands"
echo "   make run        - Build and run the application"
echo "   make test       - Run tests"
echo "   make docker-up-dev - Start development database"
echo ""