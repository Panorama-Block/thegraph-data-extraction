package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/panoramablock/thegraph-data-extraction/internal/config"
	"github.com/panoramablock/thegraph-data-extraction/pkg/client"
	"github.com/panoramablock/thegraph-data-extraction/pkg/extraction"
)

func main() {
	// Define command-line flags
	outputDir := flag.String("output", "data", "Output directory for extracted data")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent queries")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if len(cfg.Endpoints) == 0 {
		log.Fatal("No endpoints configured. Check your ENDPOINTS_JSON environment variable.")
	}
	if cfg.AuthToken == "" {
		log.Fatal("No auth token provided. Check your GRAPHQL_AUTH_TOKEN environment variable.")
	}

	// Create client
	graphClient := client.NewTheGraphClient(cfg.AuthToken)

	// Create extraction service
	service := extraction.NewService(graphClient, cfg.Endpoints)
	service.SetOutputDir(*outputDir)
	service.SetConcurrency(*concurrency)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Start extraction
	fmt.Printf("Starting data extraction from %d endpoints\n", len(cfg.Endpoints))
	if err := service.ExtractAll(); err != nil {
		log.Fatalf("Extraction failed: %v", err)
	}

	fmt.Println("Data extraction completed successfully")
}
