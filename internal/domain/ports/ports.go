package ports

import (
	"context"
	"time"

	"github.com/panoramablock/thegraph-data-extraction/internal/domain/entity"
)

// GraphQLClient defines the interface for interacting with GraphQL APIs
type GraphQLClient interface {
	// Query executes a GraphQL query and returns the result
	Query(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error
	
	// SetEndpoint configures the client to use a specific endpoint
	SetEndpoint(endpoint string)
}

// EventPublisher defines the interface for publishing events to a message bus
type EventPublisher interface {
	// PublishEntity publishes an entity to the message bus
	PublishEntity(ctx context.Context, entity *entity.Entity, topic string) error
	
	// PublishRaw publishes raw data to the message bus
	PublishRaw(ctx context.Context, key string, data []byte, topic string) error
	
	// Close closes the publisher connection
	Close() error
}

// Repository defines the interface for data persistence
type Repository interface {
	// SaveEntity saves an entity to the repository
	SaveEntity(ctx context.Context, entity *entity.Entity) error
	
	// GetLatestCursor gets the latest cursor for a given entity type and deployment
	GetLatestCursor(ctx context.Context, entityType, deployment string) (string, error)
	
	// Close closes the repository connection
	Close() error
}

// ExtractionService defines the interface for the core extraction logic
type ExtractionService interface {
	// ExtractEntities extracts entities from a given endpoint and query type
	ExtractEntities(ctx context.Context, endpoint, queryType string) ([]*entity.Entity, error)
	
	// ExtractAll extracts all configured entity types from all endpoints
	ExtractAll(ctx context.Context) error
	
	// ExtractWithDelta extracts only new entities since the last extraction
	ExtractWithDelta(ctx context.Context, endpoint, queryType, cursor string) ([]*entity.Entity, error)
}

// QueryGenerator defines the interface for generating GraphQL queries
type QueryGenerator interface {
	// GenerateQuery generates a GraphQL query for a given endpoint and type
	GenerateQuery(endpoint, queryType string) string
	
	// GeneratePaginatedQuery generates a paginated query with cursor
	GeneratePaginatedQuery(endpoint, queryType, cursor string, first int) string
}

// RateLimiter defines the interface for rate limiting API requests
type RateLimiter interface {
	// Wait blocks until a request is allowed according to rate limits
	Wait(ctx context.Context) error
	
	// Done signals that a request has completed
	Done(success bool, latency time.Duration)
	
	// UpdateRateLimit updates the rate limit based on API response
	UpdateRateLimit(rateLimit, remaining int, resetAt time.Time)
}

// WorkerPool defines the interface for managing a dynamic pool of workers
type WorkerPool interface {
	// Submit submits a task to the worker pool
	Submit(task func() error) error
	
	// Wait waits for all tasks to complete
	Wait() error
	
	// SetPoolSize dynamically adjusts the worker pool size
	SetPoolSize(size int)
	
	// Close shuts down the worker pool
	Close() error
} 