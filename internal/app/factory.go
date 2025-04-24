package app

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	
	"github.com/panoramablock/thegraph-data-extraction/internal/adapters/graphql"
	"github.com/panoramablock/thegraph-data-extraction/internal/adapters/kafka"
	"github.com/panoramablock/thegraph-data-extraction/internal/adapters/ratelimit"
	"github.com/panoramablock/thegraph-data-extraction/internal/adapters/repository"
	"github.com/panoramablock/thegraph-data-extraction/internal/adapters/worker"
	"github.com/panoramablock/thegraph-data-extraction/internal/domain/service"
	"github.com/panoramablock/thegraph-data-extraction/internal/queries"
)

// Config holds the application configuration
type Config struct {
	// API settings
	GraphQLAuthToken string
	Endpoints        []string
	QueryTypes       []string
	
	// Output settings
	OutputDir string
	
	// Kafka settings
	KafkaBrokers    []string
	KafkaTopicPrefix string
	KafkaProducer   string
	
	// Performance settings
	PageSize       int
	MaxRetries     int
	MinWorkers     int
	MaxWorkers     int
	InitialWorkers int
	InitialRate    float64
	MaxRate        float64
}

// Application holds all components of the application
type Application struct {
	// Domain services
	ExtractionService *service.ExtractionService
	
	// Adapters
	GraphQLClient  *graphql.Client
	Repository     *repository.FileRepository
	Publisher      *kafka.Publisher
	QueryGenerator *graphql.QueryGenerator
	RateLimiter    *ratelimit.AdaptiveLimiter
	WorkerPool     *worker.DynamicPool
}

// NewApplication creates a new application with all components
func NewApplication(ctx context.Context, config Config) (*Application, error) {
	// Create GraphQL client
	graphQLClient := graphql.NewClient(graphql.ClientConfig{
		AuthToken: config.GraphQLAuthToken,
	})
	
	// Create file repository
	fileRepo, err := repository.NewFileRepository(repository.FileRepositoryConfig{
		BaseDir: config.OutputDir,
	})
	if err != nil {
		return nil, err
	}
	
	// Create Kafka publisher
	kafkaPublisher := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:     config.KafkaBrokers,
		TopicPrefix: config.KafkaTopicPrefix,
		Producer:    config.KafkaProducer,
	})
	
	// Create query generator and load queries
	queryGenerator := graphql.NewQueryGenerator(graphql.QueryGeneratorConfig{
		DefaultPageSize: config.PageSize,
	})
	queryGenerator.LoadQueryVariants(queries.GetQueryVariants())
	queryGenerator.AddMetaDeploymentToQueries()
	
	// Create rate limiter
	rateLimiter := ratelimit.NewAdaptiveLimiter(ratelimit.AdaptiveLimiterConfig{
		InitialRate: config.InitialRate,
		MaxRate:     config.MaxRate,
	})
	
	// Create worker pool
	workerPool := worker.NewDynamicPool(worker.PoolConfig{
		InitialWorkers: config.InitialWorkers,
		MinWorkers:     config.MinWorkers,
		MaxWorkers:     config.MaxWorkers,
	})
	
	// Create extraction service
	extractionService := service.NewExtractionService(
		ctx,
		graphQLClient,
		kafkaPublisher,
		fileRepo,
		queryGenerator,
		rateLimiter,
		workerPool,
		config.Endpoints,
		config.QueryTypes,
		service.ExtractionConfig{
			PageSize:   config.PageSize,
			MaxRetries: config.MaxRetries,
		},
	)
	
	// Log configuration
	log.Info().
		Strs("endpoints", config.Endpoints).
		Strs("queryTypes", config.QueryTypes).
		Int("pageSize", config.PageSize).
		Int("maxRetries", config.MaxRetries).
		Int("minWorkers", config.MinWorkers).
		Int("maxWorkers", config.MaxWorkers).
		Float64("initialRate", config.InitialRate).
		Strs("kafkaBrokers", config.KafkaBrokers).
		Msg("Application initialized")
	
	return &Application{
		ExtractionService: extractionService,
		GraphQLClient:     graphQLClient,
		Repository:        fileRepo,
		Publisher:         kafkaPublisher,
		QueryGenerator:    queryGenerator,
		RateLimiter:       rateLimiter,
		WorkerPool:        workerPool,
	}, nil
}

// DefaultConfig creates a default configuration
func DefaultConfig() Config {
	return Config{
		QueryTypes:     []string{"tokens", "transactions", "factories", "swaps"},
		OutputDir:      "data",
		PageSize:       100,
		MaxRetries:     3,
		MinWorkers:     2,
		MaxWorkers:     10,
		InitialWorkers: 4,
		InitialRate:    5.0,
		MaxRate:        20.0,
		KafkaBrokers:   []string{"localhost:9092"},
		KafkaTopicPrefix: "thegraph",
		KafkaProducer:  "thegraph-extractor",
	}
}

// ConfigFromEnvironment loads configuration from environment variables
func ConfigFromEnvironment() Config {
	config := DefaultConfig()
	
	// Load environment variables using godotenv if necessary
	// ...
	
	// Override from environment variables if set
	// Example: config.OutputDir = getEnvOrDefault("OUTPUT_DIR", config.OutputDir)
	
	return config
}

// Close closes all components of the application
func (a *Application) Close() error {
	var errors []error
	
	// Close all components
	if err := a.WorkerPool.Close(); err != nil {
		errors = append(errors, err)
	}
	
	if err := a.Publisher.Close(); err != nil {
		errors = append(errors, err)
	}
	
	if err := a.Repository.Close(); err != nil {
		errors = append(errors, err)
	}
	
	// Log errors
	if len(errors) > 0 {
		errorStrings := make([]string, len(errors))
		for i, err := range errors {
			errorStrings[i] = err.Error()
		}
		log.Error().
			Str("errors", strings.Join(errorStrings, "; ")).
			Msg("Errors occurred while closing application")
		return errors[0]
	}
	
	log.Info().Msg("Application closed successfully")
	return nil
} 