#!/bin/bash

echo "🛑 Stopping BookRank services"
echo "============================="

# Stop containers
echo "📦 Stopping containers..."
docker stop bookrank-postgres bookrank-redis 2>/dev/null || echo "Containers already stopped"

# Remove containers
echo "🗑️ Removing containers..."
docker rm bookrank-postgres bookrank-redis 2>/dev/null || echo "Containers already removed"

# Remove network
echo "🌐 Removing network..."
docker network rm bookrank-network 2>/dev/null || echo "Network already removed"

echo "✅ Cleanup complete"
echo ""
echo "To stop colima completely:"
echo "colima stop"