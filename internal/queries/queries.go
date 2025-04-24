package queries

import "strings"

// Query variants for different schema types
var queryVariants = map[string]map[string]string{
	"tokens": {
		"default": `{
  tokens(first: 1000) {
    id
    decimals
    name
    symbol
  }
}`,
		"9cT3GzNxcLWFXGAgqdJsydZkh9ajKEXn4hKvkRLJHgwv": `{
  tokens(first: 1000) {
    id
    decimals
    name
    symbol
    vault {
      id
    }
    isNative
  }
}`,
		"9EAxYE17Cc478uzFXRbM7PVnMUSsgb99XZiGxodbtpbk": `{
  tokens(first: 1000) {
    id
    decimals
    name
    symbol
    totalValueLockedUSD
    volume
    volumeUSD
  }
}`,
	},
	"transactions": {
		"default": `{
  transactions(first: 1000) {
    id
    blockNumber
    timestamp
  }
}`,
		"9cT3GzNxcLWFXGAgqdJsydZkh9ajKEXn4hKvkRLJHgwv": `{
  transactions(first: 1000) {
    id
    blockNumber
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
}`,
	},
	"factories": {
		"9EAxYE17Cc478uzFXRbM7PVnMUSsgb99XZiGxodbtpbk": `{
  factories(first: 1000) {
    id
    poolCount
    txCount
    totalVolumeUSD
    owner
    totalFeesUSD
    untrackedVolumeUSD
  }}`,
		"EMnAvnfc1fwGSU6ToqYJCeEkXmSgmDmhwtyaha1tM5oi": `{
  factories(first: 1000) {
    id
    poolCount
    txCount
    totalVolumeUSD
    owner
    totalFeesUSD
    untrackedVolumeUSD
  }}`},
	"swaps": {
		"9EAxYE17Cc478uzFXRbM7PVnMUSsgb99XZiGxodbtpbk": `{
  swaps(first: 1000) {
    amountUSD
    id
    origin
    recipient
    sender
    timestamp
  }
}`,
		"EMnAvnfc1fwGSU6ToqYJCeEkXmSgmDmhwtyaha1tM5oi": `{
  swaps(first: 1000) {
    id
    timestamp
    amountUSD
  }
}`,
	},
	"_meta": {
		"9cT3GzNxcLWFXGAgqdJsydZkh9ajKEXn4hKvkRLJHgwv": `{
  _meta {
    deployment
    hasIndexingErrors
  }
}`,
		"9EAxYE17Cc478uzFXRbM7PVnMUSsgb99XZiGxodbtpbk": `{
   _meta {
    deployment
    hasIndexingErrors
  }
}`,
		"EMnAvnfc1fwGSU6ToqYJCeEkXmSgmDmhwtyaha1tM5oi": `{
  _meta {
    deployment
    hasIndexingErrors
  }
}`,
	},
	"vaults": {
		"9cT3GzNxcLWFXGAgqdJsydZkh9ajKEXn4hKvkRLJHgwv": `{
  vaults {
    defaultAlienDepositFee
    defaultAlienWithdrawFee
    defaultNativeDepositFee
    defaultNativeWithdrawFee
    id
  }
}`,
	},
	"withdraws": {
		"9cT3GzNxcLWFXGAgqdJsydZkh9ajKEXn4hKvkRLJHgwv": `{
  withdraws {
    amount
    fee
    id
    isNative
    payloadId
  }
}`,
	},
	"burns": {
		"9EAxYE17Cc478uzFXRbM7PVnMUSsgb99XZiGxodbtpbk": `{
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
}`,
		"EMnAvnfc1fwGSU6ToqYJCeEkXmSgmDmhwtyaha1tM5oi": `{
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
}`,
	},
	"accounts": {
		"9cT3GzNxcLWFXGAgqdJsydZkh9ajKEXn4hKvkRLJHgwv": `{
  accounts {
    id
  }
}`,
	},
	"pools": {
		"9EAxYE17Cc478uzFXRbM7PVnMUSsgb99XZiGxodbtpbk": `{
   pools {
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
}`,
		"EMnAvnfc1fwGSU6ToqYJCeEkXmSgmDmhwtyaha1tM5oi": `{
   pools {
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
    volumeToken1
    volumeToken0
    volumeUSD
  }
}`,
	},
	"skimFees": {
		"9cT3GzNxcLWFXGAgqdJsydZkh9ajKEXn4hKvkRLJHgwv": `{
  skimFees {
    amount
    id
    skimToEverscale
  }
}`,
	},
}

// GetQueryVariants returns the map of query variants for use in other components
func GetQueryVariants() map[string]map[string]string {
	return queryVariants
}

// GetEndpointID returns a shortened endpoint ID for use in logs and filenames
func GetEndpointID(endpoint string) string {
	// Use just the first 8 characters of the endpoint for readability
	if len(endpoint) > 8 {
		return endpoint[:8]
	}
	return endpoint
}

// GetQueryForEndpoint returns the appropriate query for a given endpoint
func GetQueryForEndpoint(endpoint string, queryType string) string {
	// Check if we have a specific query for this endpoint
	if variants, ok := queryVariants[queryType]; ok {
		if query, ok := variants[endpoint]; ok {
			return query
		}

		// If no exact match, try to find an endpoint that contains this one
		// (for example, if endpoint is shortened)
		for variantEndpoint, query := range variants {
			if strings.Contains(endpoint, variantEndpoint) ||
				strings.Contains(variantEndpoint, endpoint) {
				return query
			}
		}

		// Fall back to default if available
		if defaultQuery, ok := variants["default"]; ok {
			return defaultQuery
		}
	}

	// If no query is found or no default, return empty
	return ""
}
