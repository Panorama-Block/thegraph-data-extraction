package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/panoramablock/thegraph-data-extraction/internal/config"
	"github.com/panoramablock/thegraph-data-extraction/pkg/client"
	"github.com/panoramablock/thegraph-data-extraction/pkg/extraction"
)

func main() {
	// Define command-line flags
	outputDir := flag.String("output", "data", "Output directory for extracted data")
	concurrency := flag.Int("concurrency", 4, "Number of concurrent queries")
	flag.Parse()

	for {
		fmt.Println("Starting data extraction loop...")

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Printf("Failed to load configuration: %v\n", err)
			waitForNextRun()
			continue
		}

		// Validate configuration
		if len(cfg.Endpoints) == 0 {
			log.Println("No endpoints configured. Check your ENDPOINTS_JSON environment variable.")
			waitForNextRun()
			continue
		}
		if cfg.AuthToken == "" {
			log.Println("No auth token provided. Check your GRAPHQL_AUTH_TOKEN environment variable.")
			waitForNextRun()
			continue
		}

		// Create client
		graphClient := client.NewTheGraphClient(cfg.AuthToken)

		// Create extraction service
		service := extraction.NewService(graphClient, cfg.Endpoints)
		service.SetOutputDir(*outputDir)
		service.SetConcurrency(*concurrency)

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Printf("Failed to create output directory: %v\n", err)
			waitForNextRun()
			continue
		}

		// Start extraction
		fmt.Printf("Extracting from %d endpoints...\n", len(cfg.Endpoints))
		if err := service.ExtractAll(); err != nil {
			log.Printf("Extraction failed: %v\n", err)
		} else {
			fmt.Println("Data extraction completed successfully")
		}

		waitForNextRun()
	}
}

func waitForNextRun() {
	fmt.Println("Waiting 30 minutes for the next run...")
	time.Sleep(30 * time.Minute)
}
