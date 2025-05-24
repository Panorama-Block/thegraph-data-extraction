# TheGraph AVAX DEX Data Extraction

A powerful tool for extracting AVAX DEX data from The Graph API endpoints using a modular hexagonal architecture.

## Features

- **Hexagonal Architecture**: Clean separation of core domain, adapters, and infrastructure
- **Event-Driven Design**: Publishes extraction results to Kafka for downstream processing
- **Dynamic Pagination**: Efficiently handles datasets of any size with cursor-based pagination
- **Adaptive Rate Limiting**: Automatically adjusts request rates based on API response patterns
- **Delta Extraction**: Only extracts data since the last run, optimizing for efficiency
- **Dynamic Worker Pool**: Scales worker count based on API latency and performance metrics
- **Structured Logging**: Comprehensive, well-formatted logs for monitoring and debugging
- **Streaming JSON Processing**: Optimized memory usage with streaming encoders/decoders

## Architecture

This project follows a hexagonal (ports and adapters) architecture pattern:

```
┌────────────────────────────────────────────────┐
│                                                │
│  ┌────────────────────────────────────────┐    │
│  │                                        │    │
│  │  ┌────────────────────────────────┐    │    │
│  │  │                                │    │    │
│  │  │     Core Domain                │    │    │
│  │  │  (Extraction Service)          │    │    │
│  │  │                                │    │    │
│  │  └─────────────┬─────────────────┘    │    │
│  │                │                       │    │
│  │     Domain     │                       │    │
│  │                │                       │    │
│  └────────────────┼───────────────────────┘    │
│                   │                             │
│  ┌────────────────┼───────────────────────┐    │
│  │                ▼                       │    │
│  │  ┌─────────────────────┐  ┌─────────┐  │    │
│  │  │    Port Interfaces  │  │         │  │    │
│  │  │  ┌─────────────────┐│  │         │  │    │
│  │  │  │ GraphQLClient   ││  │         │  │    │
│  │  │  ├─────────────────┤│  │         │  │    │
│  │  │  │ EventPublisher  ││  │         │  │    │
│  │  │  ├─────────────────┤│  │         │  │    │
│  │  │  │ Repository      ││  │         │  │    │
│  │  │  ├─────────────────┤│  │         │  │    │
│  │  │  │ RateLimiter     ││  │         │  │    │
│  │  │  ├─────────────────┤│  │         │  │    │
│  │  │  │ WorkerPool      ││  │         │  │    │
│  │  │  └─────────────────┘│  │         │  │    │
│  │  └─────────────────────┘  │         │  │    │
│  │                           │         │  │    │
│  │            Adapters       │         │  │    │
│  │                           │         │  │    │
│  └───────────────────────────┼─────────┘    │  │
│                              │              │  │
└──────────────────────────────┼──────────────┘  │
                               │                 │
┌──────────────────────────────┼─────────────────┤
│  ┌────────────────┐  ┌───────▼──────┐  ┌─────┐ │
│  │  GraphQL API   │  │  Kafka       │  │ File│ │
│  └────────────────┘  └──────────────┘  └─────┘ │
│                                                │
│                 Infrastructure                 │
└────────────────────────────────────────────────┘
```

### Key Components

- **Domain Layer**: Contains the core business logic for extraction
  - `Entity`: Core domain entities
  - `ExtractionService`: Main business logic for data extraction

- **Ports**: Define interfaces for external dependencies
  - `GraphQLClient`: Interface for GraphQL API communication
  - `EventPublisher`: Interface for publishing events
  - `Repository`: Interface for data persistence
  - `RateLimiter`: Interface for rate limiting
  - `WorkerPool`: Interface for concurrent task execution

- **Adapters**: Implement the ports interfaces with concrete technologies
  - `GraphQL Client`: Connects to The Graph API
  - `Kafka Publisher`: Sends data to Kafka topics
  - `File Repository`: Stores data and cursors in files
  - `Adaptive Rate Limiter`: Smart API request rate control
  - `Dynamic Worker Pool`: Adjustable concurrent task execution

## Installation

### Using install script

```bash
# Run the install script
./install.sh
```

### Manual setup

```bash
# Download dependencies
go get github.com/rs/zerolog/log
go get github.com/google/uuid
go get github.com/segmentio/kafka-go
go get golang.org/x/time/rate
go mod tidy

# Build the CLI tool
go build -o thegraph-extract ./cmd/main.go
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
# Run with default settings
./thegraph-extract

# Specify custom output directory and concurrency
./thegraph-extract -output ./custom-data -concurrency 8

# Configure Kafka brokers
./thegraph-extract -kafka "broker1:9092,broker2:9092" -topic-prefix "mycompany.thegraph"

# Set page size for pagination
./thegraph-extract -page-size 200

# Enable debug logging
DEBUG=true ./thegraph-extract
```

### CLI Options

- `-output`: Output directory for data files (default: "data")
- `-concurrency`: Initial number of concurrent workers (default: 4)
- `-kafka`: Comma-separated list of Kafka brokers (default: "localhost:9092")
- `-topic-prefix`: Prefix for Kafka topics (default: "thegraph")
- `-page-size`: Number of items per page in GraphQL queries (default: 100)

## Extending the Project

### Adding a New Query Type

1. Add the query template to `internal/queries/queries.go`:

```go
"pools": {
    "default": `{
      pools(first: 1000) {
        id
        token0 { id, symbol }
        token1 { id, symbol }
        volumeUSD
      }
    }`,
}
```

2. Add the new query type to the app configuration in `cmd/main.go`:

```go
appConfig := app.Config{
    // ...
    QueryTypes: []string{"tokens", "transactions", "factories", "swaps", "pools"},
    // ...
}
```

### Implementing a New Adapter

To replace Kafka with another message broker (e.g., RabbitMQ):

1. Create a new adapter in `internal/adapters/rabbitmq/publisher.go`
2. Implement the `EventPublisher` interface
3. Update the app factory to use your new adapter

## License

MIT 
