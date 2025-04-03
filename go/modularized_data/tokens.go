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
  tokens(first: 1000) {
    id
    transaction {
      id
      blockNumber
      event
      from
      gasLimit
      gasPrice
      gasSent
      hash
      index
      logIndex
      timestamp
      to
      value
    }
    vault {
      id
    }
    activationBlock
    blacklisted
    decimals
    depositFee
    isNative
    name
    namePrefix
    nativeAddr
    nativeWid
    symbol
    symbolPrefix
  }
}`,
		`{
  tokens {
    decimals
    derivedNative
    id
    feesUSD
    name
    poolCount
    symbol
    tokenAddress
    totalSupply
    totalValueLocked
    totalValueLockedUSD
    txCount
    untrackedVolumeUSD
    volume
    volumeUSD
  }
}`,
		`{
  tokens(first: 1000) {
    id
    symbol
    name
    decimals
    derivedETH
    feesUSD
    poolCount
    totalValueLocked
    totalValueLockedUSD
    totalValueLockedUSDUntracked
    txCount
    untrackedVolumeUSD
    volume
    volumeUSD
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