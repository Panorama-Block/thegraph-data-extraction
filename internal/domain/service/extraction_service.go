package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/panoramablock/thegraph-data-extraction/internal/domain/entity"
	"github.com/panoramablock/thegraph-data-extraction/internal/domain/ports"
)

// ExtractionService implements the core extraction logic
type ExtractionService struct {
	client         ports.GraphQLClient
	publisher      ports.EventPublisher
	repository     ports.Repository
	queryGenerator ports.QueryGenerator
	rateLimiter    ports.RateLimiter
	workerPool     ports.WorkerPool
	
	endpoints      []string
	queryTypes     []string
	pageSize       int
	maxRetries     int
	retryDelay     time.Duration
}

// ExtractionConfig holds the configuration for the extraction service
type ExtractionConfig struct {
	PageSize   int
	MaxRetries int
	RetryDelay time.Duration
}

// NewExtractionService creates a new extraction service
func NewExtractionService(
	ctx context.Context,
	client ports.GraphQLClient,
	publisher ports.EventPublisher,
	repository ports.Repository,
	queryGenerator ports.QueryGenerator,
	rateLimiter ports.RateLimiter,
	workerPool ports.WorkerPool,
	endpoints []string,
	queryTypes []string,
	config ExtractionConfig,
) *ExtractionService {
	if config.PageSize <= 0 {
		config.PageSize = 100 // Default page size
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3 // Default max retries
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 5 * time.Second // Default retry delay
	}
	
	return &ExtractionService{
		client:         client,
		publisher:      publisher,
		repository:     repository,
		queryGenerator: queryGenerator,
		rateLimiter:    rateLimiter,
		workerPool:     workerPool,
		endpoints:      endpoints,
		queryTypes:     queryTypes,
		pageSize:       config.PageSize,
		maxRetries:     config.MaxRetries,
		retryDelay:     config.RetryDelay,
	}
}

// ExtractAll extracts all configured entity types from all endpoints
func (s *ExtractionService) ExtractAll(ctx context.Context) error {
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var errs []error
	
	for _, endpoint := range s.endpoints {
		for _, queryType := range s.queryTypes {
			wg.Add(1)
			
			// Submit extraction task to worker pool
			err := s.workerPool.Submit(func() error {
				defer wg.Done()
				
				// Get the latest cursor to perform delta extraction
				cursor, err := s.repository.GetLatestCursor(ctx, queryType, endpoint)
				if err != nil {
					log.Error().
						Str("endpoint", endpoint).
						Str("queryType", queryType).
						Err(err).
						Msg("Failed to get latest cursor")
					// Continue with empty cursor (full extraction)
					cursor = ""
				}
				
				// Extract entities with delta if cursor exists
				var entities []*entity.Entity
				if cursor != "" {
					entities, err = s.ExtractWithDelta(ctx, endpoint, queryType, cursor)
				} else {
					entities, err = s.ExtractEntities(ctx, endpoint, queryType)
				}
				
				if err != nil {
					errMu.Lock()
					errs = append(errs, fmt.Errorf("error extracting %s from %s: %w", queryType, endpoint, err))
					errMu.Unlock()
					return err
				}
				
				// Publish entities to message bus
				topic := fmt.Sprintf("%s.%s", endpoint, queryType)
				for _, e := range entities {
					if err := s.publisher.PublishEntity(ctx, e, topic); err != nil {
						log.Error().
							Str("endpoint", endpoint).
							Str("queryType", queryType).
							Str("entityId", e.ID).
							Err(err).
							Msg("Failed to publish entity")
						errMu.Lock()
						errs = append(errs, fmt.Errorf("error publishing entity %s: %w", e.ID, err))
						errMu.Unlock()
					}
				}
				
				log.Info().
					Str("endpoint", endpoint).
					Str("queryType", queryType).
					Int("entityCount", len(entities)).
					Msg("Successfully extracted and published entities")
					
				return nil
			})
			
			if err != nil {
				log.Error().
					Str("endpoint", endpoint).
					Str("queryType", queryType).
					Err(err).
					Msg("Failed to submit extraction task")
				errMu.Lock()
				errs = append(errs, fmt.Errorf("error submitting task for %s from %s: %w", queryType, endpoint, err))
				errMu.Unlock()
			}
		}
	}
	
	// Wait for all extraction tasks to complete
	if err := s.workerPool.Wait(); err != nil {
		return fmt.Errorf("error waiting for worker pool completion: %w", err)
	}
	
	// Check if there were any errors
	if len(errs) > 0 {
		log.Error().
			Int("errorCount", len(errs)).
			Msg("Extraction completed with errors")
		return fmt.Errorf("completed with %d errors", len(errs))
	}
	
	log.Info().Msg("All data extracted and published successfully")
	return nil
}

// ExtractEntities extracts entities from a given endpoint and query type
func (s *ExtractionService) ExtractEntities(ctx context.Context, endpoint, queryType string) ([]*entity.Entity, error) {
	// Generate initial query
	query := s.queryGenerator.GenerateQuery(endpoint, queryType)
	if query == "" {
		return nil, fmt.Errorf("no query defined for %s on endpoint %s", queryType, endpoint)
	}
	
	// Set client endpoint
	s.client.SetEndpoint(endpoint)
	
	// Execute query with pagination
	return s.executeQueryWithPagination(ctx, endpoint, queryType, query, "")
}

// ExtractWithDelta extracts only new entities since the last extraction
func (s *ExtractionService) ExtractWithDelta(ctx context.Context, endpoint, queryType, cursor string) ([]*entity.Entity, error) {
	// Generate paginated query with cursor
	query := s.queryGenerator.GeneratePaginatedQuery(endpoint, queryType, cursor, s.pageSize)
	if query == "" {
		return nil, fmt.Errorf("no paginated query defined for %s on endpoint %s", queryType, endpoint)
	}
	
	// Set client endpoint
	s.client.SetEndpoint(endpoint)
	
	// Execute query with pagination
	return s.executeQueryWithPagination(ctx, endpoint, queryType, query, cursor)
}

// executeQueryWithPagination executes a query with pagination to retrieve all results
func (s *ExtractionService) executeQueryWithPagination(
	ctx context.Context,
	endpoint, queryType, query, startCursor string,
) ([]*entity.Entity, error) {
	var allEntities []*entity.Entity
	var currentCursor = startCursor
	hasMore := true
	
	for hasMore {
		// Rate limit the request
		if err := s.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limit error: %w", err)
		}
		
		startTime := time.Now()
		var response entity.GraphResponse
		var err error
		var success bool
		
		// Retry logic
		for retry := 0; retry <= s.maxRetries; retry++ {
			if retry > 0 {
				log.Warn().
					Str("endpoint", endpoint).
					Str("queryType", queryType).
					Int("retry", retry).
					Err(err).
					Msg("Retrying query")
				time.Sleep(s.retryDelay)
			}
			
			// Execute the query
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			err = s.client.Query(ctx, query, nil, &response)
			cancel()
			
			if err == nil {
				success = true
				break
			}
		}
		
		// Report request completion to rate limiter
		latency := time.Since(startTime)
		s.rateLimiter.Done(success, latency)
		
		if !success {
			return nil, fmt.Errorf("query failed after %d retries: %w", s.maxRetries, err)
		}
		
		// Process the response into entities
		entities, nextCursor, more := s.processResponse(endpoint, queryType, response.Data)
		allEntities = append(allEntities, entities...)
		
		// Check if we have more pages
		if !more || nextCursor == currentCursor || nextCursor == "" {
			hasMore = false
		} else {
			currentCursor = nextCursor
			query = s.queryGenerator.GeneratePaginatedQuery(endpoint, queryType, currentCursor, s.pageSize)
		}
	}
	
	return allEntities, nil
}

// processResponse processes a GraphQL response into domain entities
func (s *ExtractionService) processResponse(endpoint, queryType string, data map[string]interface{}) ([]*entity.Entity, string, bool) {
	var entities []*entity.Entity
	var nextCursor string
	hasMore := false
	
	// Extract the data array for the query type
	if data == nil {
		return entities, nextCursor, hasMore
	}
	
	if items, ok := data[queryType].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				id, _ := itemMap["id"].(string)
				if id == "" {
					id = uuid.New().String()
				}
				
				// Create entity
				entity := &entity.Entity{
					ID:         id,
					Type:       queryType,
					Deployment: endpoint,
					Timestamp:  time.Now().UTC(),
					Data:       itemMap,
				}
				
				entities = append(entities, entity)
				
				// Extract cursor from the last item
				if cursor, ok := itemMap["id"].(string); ok {
					nextCursor = cursor
				}
			}
		}
	}
	
	// Check if there are more pages
	if pageInfo, ok := data["pageInfo"].(map[string]interface{}); ok {
		if hasNextPage, ok := pageInfo["hasNextPage"].(bool); ok {
			hasMore = hasNextPage
		}
		if endCursor, ok := pageInfo["endCursor"].(string); ok && endCursor != "" {
			nextCursor = endCursor
		}
	} else {
		// If we don't have explicit pageInfo, assume there's more if we got a full page
		hasMore = len(entities) >= s.pageSize
	}
	
	return entities, nextCursor, hasMore
} 