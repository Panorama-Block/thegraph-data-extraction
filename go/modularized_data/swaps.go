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
		``,
		`{swaps {
    amount0
    amount1
    amountUSD
    id
    logIndex
    origin
    recipient
    sender
    sqrtPriceX96
    tick
    timestamp
    pool {
      balanceOfBlock
      collectedFeesToken0
      collectedFeesToken1
      collectedFeesUSD
      createdAtBlockNumber
      createdAtTimestamp
      feeGrowthBlock
      feeGrowthGlobal0X128
      feeGrowthGlobal1X128
      feeTier
      feesUSD
      id
      liquidity
      liquidityProviderCount
      observationIndex
      protocolFeeToken0
      protocolFeeToken1
      sqrtPrice
      tick
      token0Price
      token1Price
      totalValueLockedNative
      totalValueLockedToken0
      totalValueLockedToken1
      totalValueLockedUSD
      txCount
      untrackedVolumeUSD
      volumeToken0
      volumeToken1
      volumeUSD
    }
    transaction {
      blockNumber
      gasPrice
      gasUsed
      id
      timestamp
    }
  }}`,
		`{swaps {
    amount0
    amount1
    amountUSD
    id
    logIndex
    origin
    recipient
    sender
    sqrtPriceX96
    tick
    timestamp
    pool {
      collectedFeesToken0
      collectedFeesToken1
      collectedFeesUSD
      createdAtBlockNumber
      createdAtTimestamp
      feeGrowthGlobal0X128
      feeGrowthGlobal1X128
      feeTier
      feesUSD
      id
      initialFee
      liquidity
      liquidityProviderCount
      observationIndex
      sqrtPrice
      tick
      token0Price
      token1Price
      totalValueLockedETH
      totalValueLockedToken0
      totalValueLockedToken1
      totalValueLockedUSD
      totalValueLockedUSDUntracked
      txCount
      untrackedVolumeUSD
      volumeToken0
      volumeToken1
      volumeUSD
    }
  }}`,
	}

	for i, endpoint := range endpoints {
		if i == 0 {
			i++;
		}
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