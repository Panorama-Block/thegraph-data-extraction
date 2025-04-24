package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/panoramablock/thegraph-data-extraction/internal/app"
	"github.com/panoramablock/thegraph-data-extraction/internal/config"
)

func main() {
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
	}()

	// Configure logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel := zerolog.InfoLevel
	if os.Getenv("DEBUG") == "true" {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Load .env file if exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Warn().Err(err).Msg("Error loading .env file")
	}

	// Define command-line flags
	outputDir := flag.String("output", "data", "Output directory for extracted data")
	concurrency := flag.Int("concurrency", 4, "Initial number of concurrent workers")
	kafkaBrokers := flag.String("kafka", "localhost:9092", "Comma-separated list of Kafka brokers")
	topicPrefix := flag.String("topic-prefix", "thegraph", "Prefix for Kafka topics")
	pageSize := flag.Int("page-size", 100, "Number of items per page in GraphQL queries")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Validate configuration
	if len(cfg.Endpoints) == 0 {
		log.Fatal().Msg("No endpoints configured. Check your ENDPOINTS_JSON environment variable.")
	}
	if cfg.AuthToken == "" {
		log.Fatal().Msg("No auth token provided. Check your GRAPHQL_AUTH_TOKEN environment variable.")
	}

	// Create application config
	appConfig := app.Config{
		GraphQLAuthToken: cfg.AuthToken,
		Endpoints:        cfg.Endpoints,
		QueryTypes:       []string{"tokens", "transactions", "factories", "swaps"},
		OutputDir:        *outputDir,
		KafkaBrokers:     strings.Split(*kafkaBrokers, ","),
		KafkaTopicPrefix: *topicPrefix,
		KafkaProducer:    "thegraph-extractor",
		PageSize:         *pageSize,
		MaxRetries:       3,
		MinWorkers:       2,
		MaxWorkers:       10,
		InitialWorkers:   *concurrency,
		InitialRate:      5.0,
		MaxRate:          20.0,
	}

	// Create application
	application, err := app.NewApplication(ctx, appConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create application")
	}

	// Ensure cleanup on exit
	defer func() {
		if err := application.Close(); err != nil {
			log.Error().Err(err).Msg("Error during application shutdown")
		}
	}()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatal().Err(err).Str("dir", *outputDir).Msg("Failed to create output directory")
	}

	// Start extraction
	log.Info().
		Int("endpoints", len(cfg.Endpoints)).
		Int("workers", *concurrency).
		Str("output", *outputDir).
		Msg("Starting data extraction")

	startTime := time.Now()
	if err := application.ExtractionService.ExtractAll(ctx); err != nil {
		log.Fatal().Err(err).Msg("Extraction failed")
	}

	duration := time.Since(startTime)
	log.Info().
		Dur("duration", duration).
		Msg("Data extraction completed successfully")
} 