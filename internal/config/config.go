package config

import (
	"encoding/json"
	"os"

	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	Endpoints []string
	AuthToken string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	// Parse endpoints from environment variable
	endpointsJSON := os.Getenv("ENDPOINTS_JSON")
	var endpoints []string
	if err := json.Unmarshal([]byte(endpointsJSON), &endpoints); err != nil {
		return nil, err
	}

	// Get authentication token
	authToken := os.Getenv("GRAPHQL_AUTH_TOKEN")

	return &Config{
		Endpoints: endpoints,
		AuthToken: authToken,
	}, nil
} 