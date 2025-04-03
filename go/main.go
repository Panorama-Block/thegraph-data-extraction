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
  transactions(first: 1000) {
    id
    logIndex
    event
    from
    gasLimit
    gasPrice
    gasSent
    hash
    index
    timestamp
    to
    value
  }
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
  _meta {
    deployment
    hasIndexingErrors
  }
  account(id: "") {
    id
  }
  accounts {
    id
  }
  alienTransfer(id: "") {
    amount
    baseChainId
    baseToken
    decimals
    expectedEvers
    id
    name
    payload
    recipientAddr
    recipientWid
    symbol
    value
  }
  alienTransfers {
    amount
    baseChainId
    baseToken
    decimals
    expectedEvers
    id
    name
    payload
    recipientWid
    recipientAddr
    symbol
    value
  }
  token(id: "") {
    activationBlock
    blacklisted
    decimals
    depositFee
    id
    isNative
    name
    namePrefix
    nativeWid
    nativeAddr
    symbol
    symbolPrefix
    withdrawFee
  }
  vault(id: "") {
    defaultAlienDepositFee
    defaultAlienWithdrawFee
    defaultNativeDepositFee
    defaultNativeWithdrawFee
    id
  }
  vaults
  withdraw(id: "") {
    fee
    amount
    isNative
    id
    payloadId
  }
}`,
		`{
  factories(first: 1000) {
    id
    poolCount
    txCount
    totalVolumeUSD
    owner
    totalFeesNative
    totalFeesUSD
    totalValueLockedNative
    totalValueLockedUSD
    totalVolumeNative
    untrackedVolumeUSD
  }
  bundles(first: 1000) {
    id
    nativePriceUSD
  }
  TokenSearch(text: "") {
    decimals
    derivedNative
    feesUSD
    id
    name
    poolCount
    symbol
    tokenAddress
    totalSupply
    totalValueLocked
    totalValueLockedUSD
    txCount
    volume
    untrackedVolumeUSD
    volumeUSD
  }
  _meta {
    deployment
    hasIndexingErrors
  }
  bundle(id: "") {
    id
    nativePriceUSD
  }
  burn(id: "") {
    amount
    amount0
    amount1
    amountUSD
    id
    origin
    logIndex
    owner
    tickLower
    tickUpper
    timestamp
  }
  swap(id: "") {
    origin
    recipient
    sender
    sqrtPriceX96
    tick
    timestamp
  }
  swaps {
    amountUSD
    id
    logIndex
    origin
    recipient
    sender
    sqrtPriceX96
    tick
    timestamp
  }
  token(id: "") {
    decimals
    derivedNative
    feesUSD
    id
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
  factories(first: 1000) {
    id
    poolCount
    txCount
    totalVolumeUSD
    owner
    totalFeesETH
    totalFeesUSD
    totalValueLockedETH
    totalValueLockedETHUntracked
    totalValueLockedUSD
    totalValueLockedUSDUntracked
    totalVolumeETH
    untrackedVolumeUSD
  }
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
  _meta {
    deployment
    hasIndexingErrors
  }
  bundle(id: "") {
    ethPriceUSD
    id
  }
  bundles {
    ethPriceUSD
    id
  }
  burn(id: "") {
    amount
    amount0
    amount1
    amountUSD
    id
    logIndex
    origin
    owner
    tickLower
    tickUpper
    timestamp
  }
  burns {
    amount
    amount0
    amount1
    amountUSD
    id
    logIndex
    origin
    owner
    tickLower
    tickUpper
    timestamp
  }
  collect(id: "") {
    amount0
    amount1
    amountUSD
    id
    logIndex
    owner
    tickLower
    tickUpper
    timestamp
  }
  collects {
    amount0
    amount1
    amountUSD
    id
    logIndex
    owner
    tickLower
    tickUpper
    timestamp
  }
  factory(id: "") {
    id
    owner
    totalFeesETH
    poolCount
    totalFeesUSD
    totalValueLockedETH
    totalValueLockedETHUntracked
    totalValueLockedUSDUntracked
    totalValueLockedUSD
    totalVolumeETH
    totalVolumeUSD
    txCount
    untrackedVolumeUSD
  }
  flash(id: "") {
    amount0
    amount0Paid
    amount1
    amount1Paid
    amountUSD
    logIndex
    id
    recipient
    sender
    timestamp
  }
  transaction(id: "") {
    blockNumber
    id
    timestamp
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