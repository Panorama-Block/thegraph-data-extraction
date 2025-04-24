package graphql

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

// QueryGenerator implements a GraphQL query generator with pagination support
type QueryGenerator struct {
	queryTemplates     map[string]map[string]string
	paginatedTemplates map[string]map[string]string
	defaultPageSize    int
	mu                 sync.RWMutex
}

// QueryGeneratorConfig holds configuration for the query generator
type QueryGeneratorConfig struct {
	DefaultPageSize int
}

// NewQueryGenerator creates a new GraphQL query generator
func NewQueryGenerator(config QueryGeneratorConfig) *QueryGenerator {
	// Set default page size if not provided
	if config.DefaultPageSize <= 0 {
		config.DefaultPageSize = 100
	}
	
	return &QueryGenerator{
		queryTemplates:     make(map[string]map[string]string),
		paginatedTemplates: make(map[string]map[string]string),
		defaultPageSize:    config.DefaultPageSize,
	}
}

// RegisterQueryTemplate registers a query template for a specific query type and endpoint
func (g *QueryGenerator) RegisterQueryTemplate(queryType, endpoint, template string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	// Initialize the map for this query type if it doesn't exist
	if g.queryTemplates[queryType] == nil {
		g.queryTemplates[queryType] = make(map[string]string)
	}
	
	g.queryTemplates[queryType][endpoint] = template
	
	// Generate and register the paginated version of this template
	paginatedTemplate := g.generatePaginatedTemplate(template, queryType)
	
	if g.paginatedTemplates[queryType] == nil {
		g.paginatedTemplates[queryType] = make(map[string]string)
	}
	
	g.paginatedTemplates[queryType][endpoint] = paginatedTemplate
	
	log.Debug().
		Str("queryType", queryType).
		Str("endpoint", endpoint).
		Msg("Registered query template")
}

// RegisterDefaultQueryTemplate registers a default query template for a query type
func (g *QueryGenerator) RegisterDefaultQueryTemplate(queryType, template string) {
	g.RegisterQueryTemplate(queryType, "default", template)
}

// GenerateQuery generates a GraphQL query for a given endpoint and type
func (g *QueryGenerator) GenerateQuery(endpoint, queryType string) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	// Check if we have a specific query for this endpoint
	if templates, ok := g.queryTemplates[queryType]; ok {
		if query, ok := templates[endpoint]; ok {
			return query
		}
		
		// Try to find an endpoint that contains this one (for shortened endpoints)
		for templateEndpoint, query := range templates {
			if strings.Contains(endpoint, templateEndpoint) || 
			   strings.Contains(templateEndpoint, endpoint) {
				return query
			}
		}
		
		// Fall back to default if available
		if defaultQuery, ok := templates["default"]; ok {
			return defaultQuery
		}
	}
	
	// If no query is found, return empty string
	return ""
}

// GeneratePaginatedQuery generates a paginated query with cursor
func (g *QueryGenerator) GeneratePaginatedQuery(endpoint, queryType, cursor string, first int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	// Use default page size if first is not positive
	if first <= 0 {
		first = g.defaultPageSize
	}
	
	// Get the paginated template
	var template string
	
	if templates, ok := g.paginatedTemplates[queryType]; ok {
		if query, ok := templates[endpoint]; ok {
			template = query
		} else {
			// Try to find an endpoint that contains this one
			for templateEndpoint, query := range templates {
				if strings.Contains(endpoint, templateEndpoint) || 
				   strings.Contains(templateEndpoint, endpoint) {
					template = query
					break
				}
			}
			
			// Fall back to default if available
			if template == "" {
				if defaultQuery, ok := templates["default"]; ok {
					template = defaultQuery
				}
			}
		}
	}
	
	if template == "" {
		return ""
	}
	
	// Replace placeholders in the template
	query := template
	query = strings.Replace(query, "{FIRST}", fmt.Sprintf("%d", first), -1)
	
	// Add cursor if provided
	if cursor != "" {
		cursorArg := fmt.Sprintf(`, where: {id_gt: "%s"}`, cursor)
		query = strings.Replace(query, "{CURSOR}", cursorArg, -1)
	} else {
		query = strings.Replace(query, "{CURSOR}", "", -1)
	}
	
	return query
}

// generatePaginatedTemplate converts a regular query template to a paginated one
func (g *QueryGenerator) generatePaginatedTemplate(template, queryType string) string {
	// Look for first: N in the template
	firstPattern := `first: \d+`
	if strings.Contains(template, firstPattern) {
		// Replace it with first: {FIRST}
		paginatedTemplate := strings.Replace(template, "first: 1000", "first: {FIRST}{CURSOR}", 1)
		return paginatedTemplate
	}
	
	// If the template doesn't have a first parameter, try to add one
	// This is a simplistic approach and might need customization for complex queries
	entityPattern := fmt.Sprintf(`%s\(`, queryType)
	if strings.Contains(template, entityPattern) {
		// Add pagination to the entity query
		paginatedTemplate := strings.Replace(
			template,
			fmt.Sprintf("%s(", queryType),
			fmt.Sprintf("%s(first: {FIRST}{CURSOR}", queryType),
			1,
		)
		return paginatedTemplate
	}
	
	// If we can't automatically convert it, just return the original
	log.Warn().
		Str("queryType", queryType).
		Msg("Could not automatically generate paginated template")
	return template
}

// LoadQueryVariants loads query variants from a map structure
func (g *QueryGenerator) LoadQueryVariants(queryVariants map[string]map[string]string) {
	for queryType, variants := range queryVariants {
		for endpoint, query := range variants {
			g.RegisterQueryTemplate(queryType, endpoint, query)
		}
	}
}

// AddMetaDeploymentToQueries modifies queries to include _meta { deployment } field
func (g *QueryGenerator) AddMetaDeploymentToQueries() {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	// Add _meta { deployment } to all query templates
	for queryType, templates := range g.queryTemplates {
		for endpoint, query := range templates {
			// Check if query already has _meta
			if !strings.Contains(query, "_meta") {
				// Find the closing bracket of the query
				lastBraceIndex := strings.LastIndex(query, "}")
				if lastBraceIndex >= 0 {
					// Insert _meta { deployment } before the last closing brace
					modifiedQuery := query[:lastBraceIndex] + 
						"\n  _meta {\n    deployment\n  }\n" + 
						query[lastBraceIndex:]
					
					g.queryTemplates[queryType][endpoint] = modifiedQuery
					
					// Update the paginated template too
					if g.paginatedTemplates[queryType] != nil {
						g.paginatedTemplates[queryType][endpoint] = g.generatePaginatedTemplate(modifiedQuery, queryType)
					}
					
					log.Debug().
						Str("queryType", queryType).
						Str("endpoint", endpoint).
						Msg("Added _meta.deployment to query")
				}
			}
		}
	}
} 