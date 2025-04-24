#!/bin/bash

echo "Installing TheGraph Data Extraction Tool..."

# Ensure Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

# Get Go dependencies
echo "Downloading dependencies..."
go get github.com/rs/zerolog/log
go get github.com/google/uuid
go get github.com/segmentio/kafka-go
go get golang.org/x/time/rate
go mod tidy

# Build the tool
echo "Building extraction tool..."
go build -o thegraph-extract ./cmd/main.go

# Create data directory if it doesn't exist
mkdir -p data

echo "Installation completed successfully."
echo "You can now run the tool with: ./thegraph-extract"
echo "Use -h flag to see available options: ./thegraph-extract -h" 