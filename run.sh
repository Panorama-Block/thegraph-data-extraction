#!/bin/bash

# TheGraph Data Extractor - Docker Runner Script
# Usage: ./run.sh [command] [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if .env file exists
check_env() {
    if [ ! -f .env ]; then
        warn ".env file not found!"
        echo "Please create a .env file based on .env.example"
        echo ""
        echo "Required variables:"
        echo "  GRAPHQL_AUTH_TOKEN=your_token_here"
        echo "  ENDPOINTS_JSON=[\"endpoint1\", \"endpoint2\"]"
        echo ""
        echo "Example:"
        echo "  cp .env.example .env"
        echo "  # Edit .env with your values"
        echo ""
        exit 1
    fi
}

# Function to build the image
build() {
    log "Building TheGraph Data Extractor Docker image..."
    docker-compose build
    log "Build completed!"
}

# Function to run once
run_once() {
    check_env
    log "Running TheGraph Data Extractor once..."
    docker-compose run --rm thegraph-extractor --once
}

# Function to start the cron service
start() {
    check_env
    log "Starting TheGraph Data Extractor with cron scheduler..."
    docker-compose up -d
    log "Service started! Use 'docker-compose logs -f' to view logs"
}

# Function to stop the service
stop() {
    log "Stopping TheGraph Data Extractor..."
    docker-compose down
    log "Service stopped!"
}

# Function to show logs
logs() {
    docker-compose logs -f thegraph-extractor
}

# Function to show status
status() {
    docker-compose ps
}

# Function to run with custom flags
run_custom() {
    check_env
    shift # Remove the 'custom' command
    log "Running TheGraph Data Extractor with custom flags: $@"
    docker-compose run --rm thegraph-extractor "$@"
}

# Function to show help
show_help() {
    echo "TheGraph Data Extractor - Docker Runner"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  build         Build the Docker image"
    echo "  start         Start the service with cron scheduler"
    echo "  stop          Stop the service"
    echo "  once          Run extraction once and exit"
    echo "  logs          Show service logs"
    echo "  status        Show service status"
    echo "  custom        Run with custom flags (e.g., ./run.sh custom --help)"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 build                    # Build the image"
    echo "  $0 once                     # Run extraction once"
    echo "  $0 start                    # Start cron service"
    echo "  $0 custom --once --debug    # Run once with debug"
    echo "  $0 logs                     # View logs"
    echo ""
    echo "Environment:"
    echo "  Make sure to create a .env file with your configuration"
    echo "  See .env.example for required variables"
}

# Main script logic
case "${1:-help}" in
    build)
        build
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    once)
        run_once
        ;;
    logs)
        logs
        ;;
    status)
        status
        ;;
    custom)
        run_custom "$@"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac 