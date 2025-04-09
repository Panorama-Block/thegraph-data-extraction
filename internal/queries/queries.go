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
		"default": `{
  factories(first: 1000) {
    id
    poolCount
    txCount
    totalVolumeUSD
    owner
  }
}`,
		"9EAxYE17Cc478uzFXRbM7PVnMUSsgb99XZiGxodbtpbk": `{
  factories(first: 1000) {
    id
    poolCount
    txCount
    totalVolumeUSD
    owner
    totalFeesUSD
    totalValueLockedUSD
  }
}`,
	},
	"swaps": {
		"default": `{
  swaps(first: 1000) {
    id
    timestamp
  }
}`,
		"9cT3GzNxcLWFXGAgqdJsydZkh9ajKEXn4hKvkRLJHgwv": `{
  swaps(first: 1000) {
    id
    origin
    recipient
    sender
    timestamp
  }
}`,
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