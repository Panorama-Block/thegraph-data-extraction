package graphql

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/machinebox/graphql"
	"github.com/rs/zerolog/log"
)

// Client is an adapter for the GraphQL client that implements the ports.GraphQLClient interface
type Client struct {
	client    *graphql.Client
	endpoint  string
	authToken string
	headers   map[string]string
	httpClient *http.Client
}

// ClientConfig holds the configuration for the GraphQL client
type ClientConfig struct {
	BaseURL      string
	AuthToken    string
	ExtraHeaders map[string]string
	Timeout      time.Duration
}

// NewClient creates a new GraphQL client
func NewClient(config ClientConfig) *Client {
	// Set default timeout if not provided
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}
	
	// Create headers map if nil
	if config.ExtraHeaders == nil {
		config.ExtraHeaders = make(map[string]string)
	}
	
	return &Client{
		authToken: config.AuthToken,
		headers:   config.ExtraHeaders,
		httpClient: httpClient,
	}
}

// SetEndpoint configures the client to use a specific endpoint
func (c *Client) SetEndpoint(endpoint string) {
	c.endpoint = endpoint
	c.client = graphql.NewClient(
		fmt.Sprintf("https://gateway.thegraph.com/api/subgraphs/id/%s", endpoint),
		graphql.WithHTTPClient(c.httpClient),
	)
}

// Query executes a GraphQL query and returns the result
func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	if c.client == nil {
		return fmt.Errorf("client endpoint not set, call SetEndpoint first")
	}
	
	// Create GraphQL request
	request := graphql.NewRequest(query)
	
	// Add auth header
	if c.authToken != "" {
		request.Header.Set("Authorization", "Bearer "+c.authToken)
	}
	
	// Add variables if provided
	if variables != nil {
		for key, value := range variables {
			request.Var(key, value)
		}
	}
	
	// Add extra headers
	for key, value := range c.headers {
		request.Header.Set(key, value)
	}
	
	// Log the query (debug level)
	log.Debug().
		Str("endpoint", c.endpoint).
		Str("query", query).
		Interface("variables", variables).
		Msg("Executing GraphQL query")
	
	// Execute the query
	startTime := time.Now()
	err := c.client.Run(ctx, request, response)
	duration := time.Since(startTime)
	
	if err != nil {
		log.Error().
			Str("endpoint", c.endpoint).
			Str("query", query).
			Err(err).
			Dur("duration", duration).
			Msg("GraphQL query failed")
		return err
	}
	
	log.Debug().
		Str("endpoint", c.endpoint).
		Dur("duration", duration).
		Msg("GraphQL query completed successfully")
	
	return nil
} 