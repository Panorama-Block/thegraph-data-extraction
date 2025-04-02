package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/machinebox/graphql"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Erro ao carregar .env", err)
	}

	endpointsJSON := os.Getenv("ENDPOINTS_JSON")
	var endpoints []string
	if err := json.Unmarshal([]byte(endpointsJSON), &endpoints); err != nil {
		log.Fatal("Erro ao decodificar ENDPOINTS_JSON", err)
	}
	authToken := os.Getenv("GRAPHQL_AUTH_TOKEN")

	queries := []string{
		`{
			transactions(first: 5) {
				id
				logIndex
				event
				from
			}
			tokens(first: 5) {
				id
				transaction { id }
				vault { id }
				activationBlock
			}
			_meta {
				deployment
				hasIndexingErrors
				block { hash number parentHash timestamp }
			}
		}`,
		`{
			factories(first: 5) {
				id
				poolCount
				txCount
				totalVolumeUSD
				owner
			}
		}`,
		`{
			factories(first: 5) {
				id
				poolCount
				txCount
				totalVolumeUSD
				owner
				totalFeesUSD
				totalFeesETH
			}
		}`,
	}

	for i, endpoint := range endpoints {
		if i >= len(queries) {
			break
		}
		fetchData(endpoint, queries[i], authToken)
	}
}

func fetchData(endpoint, query, authToken string) {
	client := graphql.NewClient(fmt.Sprintf("https://gateway.thegraph.com/api/subgraphs/id/%s", endpoint))
	request := graphql.NewRequest(query)
	request.Header.Set("Authorization", "Bearer "+authToken)

	ctx := context.Background()
	var response map[string]interface{}
	if err := client.Run(ctx, request, &response); err != nil {
		log.Printf("Erro ao buscar dados do endpoint %s: %v", endpoint, err)
		return
	}

	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("\nDados do %s: %s\n", endpoint, string(responseJSON))
}