package extraction

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/panoramablock/thegraph-data-extraction/internal/queries"
	"github.com/panoramablock/thegraph-data-extraction/pkg/client"
)

// Service handles data extraction from The Graph API
type Service struct {
	client      *client.TheGraphClient
	endpoints   []string
	outputDir   string
	queryTypes  []string
	concurrency int
}

// NewService creates a new extraction service
func NewService(client *client.TheGraphClient, endpoints []string) *Service {
	return &Service{
		client:      client,
		endpoints:   endpoints,
		outputDir:   "data",
		queryTypes:  []string{"tokens", "transactions", "factories", "swaps", "_meta"},
		concurrency: 5, // Number of concurrent queries
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

// ExtractAll extracts all data types from all endpoints
func (s *Service) ExtractAll() error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return err
	}

	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrency
	semaphore := make(chan struct{}, s.concurrency)

	// Track errors
	var errorsMu sync.Mutex
	var errors []error

	// Process each endpoint and query type
	for _, endpoint := range s.endpoints {
		for _, queryType := range s.queryTypes {
			wg.Add(1)

			// Get the query for this endpoint and type
			query := queries.GetQueryForEndpoint(endpoint, queryType)
			if query == "" {
				log.Printf("No query defined for %s on endpoint %s, skipping", queryType, endpoint)
				wg.Done()
				continue
			}

			// Execute the query in a goroutine
			go func(endpoint, queryType, query string) {
				defer wg.Done()

				// Acquire semaphore
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Set the client endpoint
				s.client.SetEndpoint(endpoint)

				// Execute the query
				response := make(map[string]interface{})
				if err := s.client.QueryWithTimeout(query, &response, 30*time.Second); err != nil {
					errorMsg := fmt.Errorf("error querying %s from %s: %w", queryType, endpoint, err)
					log.Printf("ERROR: %v", errorMsg)
					errorsMu.Lock()
					errors = append(errors, errorMsg)
					errorsMu.Unlock()
					return
				}

				// Get a shorter endpoint ID for the filename
				endpointID := queries.GetEndpointID(endpoint)

				// Save the response to a file
				filename := filepath.Join(s.outputDir, fmt.Sprintf("%s_%s.json", endpointID, queryType))
				if err := s.saveJSON(filename, response); err != nil {
					errorMsg := fmt.Errorf("error saving %s: %w", filename, err)
					log.Printf("ERROR: %v", errorMsg)
					errorsMu.Lock()
					errors = append(errors, errorMsg)
					errorsMu.Unlock()
					return
				}

				// Print successful extraction
				log.Printf("✅ Successfully extracted %s from %s", queryType, endpointID)
			}(endpoint, queryType, query)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check if there were any errors
	if len(errors) > 0 {
		log.Printf("Completed with %d errors", len(errors))
		for i, err := range errors {
			log.Printf("Error %d: %v", i+1, err)
		}
		return fmt.Errorf("encountered %d errors during extraction", len(errors))
	}

	log.Printf("All data extracted successfully")
	return nil
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
