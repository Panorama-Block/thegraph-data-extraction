package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"

	"github.com/panoramablock/thegraph-data-extraction/internal/queries"
	"github.com/panoramablock/thegraph-data-extraction/pkg/client"
)

// DataCallback is a function type for handling extracted data
type DataCallback func(endpoint, queryType string, data map[string]interface{}) error

// Service handles data extraction from The Graph API
type Service struct {
	client          *client.TheGraphClient
	endpoints       []string
	outputDir       string
	queryTypes      []string
	concurrency     int
	kafkaWriter     *kafka.Writer
	kafkaTopicPrefix string
	dataCallback    DataCallback
}

// NewService creates a new extraction service
func NewService(client *client.TheGraphClient, endpoints []string) *Service {
	return &Service{
		client:      client,
		endpoints:   endpoints,
		outputDir:   "data",
		queryTypes:  []string{"tokens", "transactions", "factories", "swaps", "_meta", "vaults", "withdraws", "burns", "accounts", "pools", "skimFees"},
		concurrency: 11, // Number of concurrent queries
	}
}

// SetOutputDir sets the output directory for extracted data
func (s *Service) SetOutputDir(dir string) {
	s.outputDir = dir
}

// SetQueryTypes sets the types of queries to execute
func (s *Service) SetQueryTypes(types []string) {
	s.queryTypes = types
}

// SetConcurrency sets the number of concurrent queries
func (s *Service) SetConcurrency(n int) {
	s.concurrency = n
}

// SetKafkaWriter sets the Kafka writer for publishing data
func (s *Service) SetKafkaWriter(writer *kafka.Writer) {
	s.kafkaWriter = writer
}

// SetKafkaTopicPrefix sets the prefix for Kafka topics
func (s *Service) SetKafkaTopicPrefix(prefix string) {
	s.kafkaTopicPrefix = prefix
}

// SetDataCallback sets a callback function to be called with extracted data
func (s *Service) SetDataCallback(callback DataCallback) {
	s.dataCallback = callback
}

// ExtractAll extracts all data types from all endpoints
func (s *Service) ExtractAll() error {
	return s.ExtractAllWithContext(context.Background())
}

// ExtractAllWithContext extracts all data types from all endpoints with context support
func (s *Service) ExtractAllWithContext(ctx context.Context) error {
	/* DISABLED: Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return err
	}
	*/

	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrency
	semaphore := make(chan struct{}, s.concurrency)

	// Track errors
	var errorsMu sync.Mutex
	var errors []error

	log.Info().
		Int("endpoints", len(s.endpoints)).
		Int("queryTypes", len(s.queryTypes)).
		Int("concurrency", s.concurrency).
		Msg("Starting data extraction")

	// Process each endpoint and query type
	for _, endpoint := range s.endpoints {
		for _, queryType := range s.queryTypes {
			wg.Add(1)

			// Get the query for this endpoint and type
			query := queries.GetQueryForEndpoint(endpoint, queryType)
			if query == "" {
				log.Debug().
					Str("queryType", queryType).
					Str("endpoint", endpoint).
					Msg("No query defined, skipping")
				wg.Done()
				continue
			}

			// Execute the query in a goroutine
			go func(endpoint, queryType, query string) {
				defer wg.Done()

				// Check for context cancellation
				select {
				case <-ctx.Done():
					log.Warn().Msg("Context cancelled, stopping extraction")
					return
				default:
				}

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Set the client endpoint
				s.client.SetEndpoint(endpoint)

				// Execute the query
				response := make(map[string]interface{})
				if err := s.client.QueryWithTimeout(query, &response, 30*time.Second); err != nil {
					errorMsg := fmt.Errorf("error querying %s from %s: %w", queryType, endpoint, err)
					log.Error().
						Err(err).
						Str("queryType", queryType).
						Str("endpoint", endpoint).
						Msg("Query failed")
					errorsMu.Lock()
					errors = append(errors, errorMsg)
					errorsMu.Unlock()
					return
				}

				// Get a shorter endpoint ID for the filename
				endpointID := queries.GetEndpointID(endpoint)

				// Print the extracted data to console instead of saving to file
				jsonData, err := json.MarshalIndent(response, "", "  ")
				if err != nil {
					log.Error().
						Err(err).
						Str("queryType", queryType).
						Str("endpoint", endpoint).
						Msg("Failed to marshal JSON data")
				} else {
					log.Info().
						Str("queryType", queryType).
						Str("endpointID", endpointID).
						RawJSON("data", jsonData).
						Msg("Extracted data")
				}

				// Send data to Kafka if writer is configured
				if s.kafkaWriter != nil {
					if err := s.publishToKafka(ctx, endpointID, queryType, response); err != nil {
						log.Error().
							Err(err).
							Str("endpointID", endpointID).
							Str("queryType", queryType).
							Msg("Failed to publish to Kafka")
						// Don't treat Kafka errors as fatal
					} else {
						log.Debug().
							Str("endpointID", endpointID).
							Str("queryType", queryType).
							Msg("Successfully published to Kafka")
					}
				}

				// Call data callback if configured
				if s.dataCallback != nil {
					if err := s.dataCallback(endpoint, queryType, response); err != nil {
						log.Error().
							Err(err).
							Str("endpoint", endpoint).
							Str("queryType", queryType).
							Msg("Data callback failed")
						// Don't treat callback errors as fatal
					}
				}

				// Print successful extraction
				log.Info().
					Str("queryType", queryType).
					Str("endpointID", endpointID).
					Msg("Successfully extracted data")
			}(endpoint, queryType, query)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if there were any errors
	if len(errors) > 0 {
		log.Warn().
			Int("errorCount", len(errors)).
			Msg("Completed with errors")
		for i, err := range errors {
			log.Error().
				Int("errorIndex", i+1).
				Err(err).
				Msg("Extraction error")
		}
		return fmt.Errorf("encountered %d errors during extraction", len(errors))
	}

	log.Info().Msg("All data extracted successfully")
	return nil
}

// publishToKafka publishes extracted data to Kafka
func (s *Service) publishToKafka(ctx context.Context, endpointID, queryType string, data map[string]interface{}) error {
	if s.kafkaWriter == nil {
		return fmt.Errorf("kafka writer not configured")
	}

	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// Create Kafka message
	topic := fmt.Sprintf("%s_%s_%s", s.kafkaTopicPrefix, endpointID, queryType)
	message := kafka.Message{
		Topic: topic,
		Key:   []byte(fmt.Sprintf("%s-%s", endpointID, queryType)),
		Value: jsonData,
		Time:  time.Now(),
	}

	// Publish message with context
	return s.kafkaWriter.WriteMessages(ctx, message)
}

// saveJSON saves data to a JSON file
func (s *Service) saveJSON(filename string, data interface{}) error {
	// Marshal the data with indentation for readability
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write the data to the file
	return os.WriteFile(filename, jsonData, 0644)
}

// Close closes the Kafka writer if configured
func (s *Service) Close() error {
	if s.kafkaWriter != nil {
		return s.kafkaWriter.Close()
	}
	return nil
}
