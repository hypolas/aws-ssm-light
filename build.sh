#!/bin/bash
# Script de build pour light_ssm

set -e

echo "Building light_ssm for Linux (container)..."

# Initialiser le module Go si nécessaire
if [ ! -f "go.sum" ]; then
    go mod tidy
fi

# Build pour Linux (architecture du conteneur)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o light_ssm main.go

echo "✅ light_ssm built successfully!"
echo "Binary size: $(du -h light_ssm | cut -f1)"

# Test basique
echo "Testing binary..."
./light_ssm --help 2>/dev/null || echo "Binary is ready (help not implemented, normal)"

echo "✅ Ready to be copied into Docker container"