# TheGraph Data Extraction

A tool for extracting data from The Graph API endpoints.

## Features

- Concurrent and efficient data extraction from multiple endpoints
- Proper error handling and retry logic
- Modular architecture for maintainability
- Configurable concurrency and output directories
- Automatic rate limiting to avoid API throttling

## Installation

```bash
go get github.com/panoramablock/thegraph-data-extraction
```

## Configuration

Create a `.env` file in the root directory with the following variables:

```
GRAPHQL_AUTH_TOKEN=your_auth_token
ENDPOINTS_JSON=["endpoint1", "endpoint2", "endpoint3"]
```

## Usage

### CLI

```bash
# Build the CLI tool
go build -o thegraph-extract ./cmd/extraction

# Run with default settings
./thegraph-extract

# Specify custom output directory and concurrency
./thegraph-extract -output ./custom-data -concurrency 8
```

### As a library

```go
package main

import (
    "log"
    
    "github.com/panoramablock/thegraph-data-extraction/pkg/client"
    "github.com/panoramablock/thegraph-data-extraction/pkg/extraction"
)

func main() {
    // Create a client
    graphClient := client.NewTheGraphClient("your_auth_token")
    
    // Specify endpoints
    endpoints := []string{"endpoint1", "endpoint2"}
    
    // Create extraction service
    service := extraction.NewService(graphClient, endpoints)
    service.SetOutputDir("./data")
    service.SetConcurrency(4)
    
    // Start extraction
    if err := service.ExtractAll(); err != nil {
        log.Fatalf("Extraction failed: %v", err)
    }
}
```

## Project Structure

```
├── cmd/                  # Command-line tools
│   └── main              # Main CLI tool
├── internal/             # Private application packages
│   ├── config/           # Configuration handling
│   └── queries/          # GraphQL queries
├── pkg/                  # Public API packages
│   ├── client/           # TheGraph API client
│   ├── extraction/       # Data extraction service
│   └── models/           # Data models
└── data/                 # Default output directory
```

## License

MIT 