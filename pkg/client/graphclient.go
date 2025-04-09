package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/machinebox/graphql"
)

// TheGraphClient represents a client for The Graph API
type TheGraphClient struct {
	client    *graphql.Client
	authToken string
	maxRetries int
	retryDelay time.Duration
}

// NewTheGraphClient creates a new client for The Graph API
func NewTheGraphClient(authToken string) *TheGraphClient {
	return &TheGraphClient{
		authToken:  authToken,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}
}

// SetEndpoint configures the endpoint for the client
func (c *TheGraphClient) SetEndpoint(endpoint string) {
	c.client = graphql.NewClient(fmt.Sprintf("https://gateway.thegraph.com/api/subgraphs/id/%s", endpoint))
}

// Query executes a GraphQL query with retry logic
func (c *TheGraphClient) Query(ctx context.Context, query string, response interface{}) error {
	request := graphql.NewRequest(query)
	request.Header.Set("Authorization", "Bearer "+c.authToken)

	var err error
	for retry := 0; retry <= c.maxRetries; retry++ {
		if retry > 0 {
			log.Printf("Retrying query (attempt %d/%d) after error: %v", retry, c.maxRetries, err)
			time.Sleep(c.retryDelay)
		}

		err = c.client.Run(ctx, request, response)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("query failed after %d retries: %w", c.maxRetries, err)
}

// QueryWithTimeout executes a GraphQL query with a timeout
func (c *TheGraphClient) QueryWithTimeout(query string, response interface{}, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return c.Query(ctx, query, response)
} 