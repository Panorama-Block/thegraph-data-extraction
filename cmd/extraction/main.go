package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"

	"github.com/panoramablock/thegraph-data-extraction/internal/config"
	"github.com/panoramablock/thegraph-data-extraction/pkg/client"
	"github.com/panoramablock/thegraph-data-extraction/pkg/extraction"
)

// Helper functions for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

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

	// Define command-line flags with environment variable fallbacks
	outputDir := flag.String("output", getEnvOrDefault("OUTPUT_DIR", "data"), "Output directory for extracted data")
	concurrency := flag.Int("concurrency", getEnvInt("CONCURRENCY", 8), "Number of concurrent workers")
	kafkaBrokers := flag.String("kafka", getEnvOrDefault("KAFKA_BROKERS", "localhost:9092"), "Comma-separated list of Kafka brokers")
	topicPrefix := flag.String("topic-prefix", getEnvOrDefault("KAFKA_TOPIC_PREFIX", "thegraph"), "Prefix for Kafka topics")
	cronSchedule := flag.String("cron", getEnvOrDefault("CRON_SCHEDULE", "*/5 * * * *"), "Cron schedule for automatic extraction (default: every 5 minutes)")
	runOnce := flag.Bool("once", getEnvBool("RUN_ONCE", false), "Run extraction once and exit (disable cron)")
	enableKafka := flag.Bool("enable-kafka", getEnvBool("ENABLE_KAFKA", true), "Enable Kafka publishing")
	flag.Parse()

	log.Info().
		Str("outputDir", *outputDir).
		Int("concurrency", *concurrency).
		Str("kafkaBrokers", *kafkaBrokers).
		Str("topicPrefix", *topicPrefix).
		Str("cronSchedule", *cronSchedule).
		Bool("runOnce", *runOnce).
		Bool("enableKafka", *enableKafka).
		Msg("Starting TheGraph Data Extraction Service")

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

	// Create GraphQL client
	graphClient := client.NewTheGraphClient(cfg.AuthToken)

	// Create extraction service
	service := extraction.NewService(graphClient, cfg.Endpoints)
	service.SetOutputDir(*outputDir)
	service.SetConcurrency(*concurrency)

	// Setup Kafka if enabled
	var kafkaWriter *kafka.Writer
	if *enableKafka {
		kafkaWriter = &kafka.Writer{
			Addr:         kafka.TCP(strings.Split(*kafkaBrokers, ",")...),
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			BatchSize:    100,
		}
		service.SetKafkaWriter(kafkaWriter)
		service.SetKafkaTopicPrefix(*topicPrefix)
		
		log.Info().
			Strs("brokers", strings.Split(*kafkaBrokers, ",")).
			Str("topicPrefix", *topicPrefix).
			Msg("Kafka publishing enabled")
	} else {
		log.Info().Msg("Kafka publishing disabled")
	}

	// Ensure cleanup on exit
	defer func() {
		if err := service.Close(); err != nil {
			log.Error().Err(err).Msg("Error during service shutdown")
		}
		if kafkaWriter != nil {
			if err := kafkaWriter.Close(); err != nil {
				log.Error().Err(err).Msg("Error closing Kafka writer")
			}
		}
	}()

	// Create output directory if it doesn't exist
	/* DISABLED: No longer saving files to disk
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatal().Err(err).Str("dir", *outputDir).Msg("Failed to create output directory")
	}
	*/

	// Define extraction function
	extractionFunc := func() {
		log.Info().Msg("Starting scheduled data extraction")
		startTime := time.Now()

		if err := service.ExtractAllWithContext(ctx); err != nil {
			log.Error().Err(err).Msg("Extraction failed")
			return
		}

		duration := time.Since(startTime)
		log.Info().
			Dur("duration", duration).
			Msg("Scheduled data extraction completed successfully")
	}

	if *runOnce {
		// Run extraction once and exit
		log.Info().
			Int("endpoints", len(cfg.Endpoints)).
			Int("workers", *concurrency).
			Str("output", *outputDir).
			Msg("Running single extraction")

		extractionFunc()
		log.Info().Msg("Single extraction completed, exiting")
		return
	}

	// Setup cron scheduler
	c := cron.New() // Standard 5-field format: minute hour day month weekday
	
	// Add extraction job to cron
	_, err = c.AddFunc(*cronSchedule, extractionFunc)
	if err != nil {
		log.Fatal().Err(err).Str("schedule", *cronSchedule).Msg("Failed to add cron job")
	}

	log.Info().
		Int("endpoints", len(cfg.Endpoints)).
		Int("workers", *concurrency).
		Str("output", *outputDir).
		Str("schedule", *cronSchedule).
		Msg("Starting cron scheduler for automatic data extraction")

	// Start the cron scheduler
	c.Start()
	defer c.Stop()

	// Run initial extraction immediately
	log.Info().Msg("Running initial extraction...")
	extractionFunc()

	// Keep the application running until interrupted
	log.Info().Msg("Cron scheduler started. Press Ctrl+C to stop.")
	
	// Wait for context cancellation (SIGINT/SIGTERM)
	<-ctx.Done()
	
	log.Info().Msg("Shutdown signal received, stopping cron scheduler...")
	
	// Give ongoing extractions time to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// Wait for shutdown
	select {
	case <-shutdownCtx.Done():
		log.Warn().Msg("Shutdown timeout exceeded")
	case <-time.After(2 * time.Second):
		log.Info().Msg("Graceful shutdown completed")
	}
}
