package models

// Transaction represents a blockchain transaction
type Transaction struct {
	ID          string `json:"id"`
	BlockNumber string `json:"blockNumber,omitempty"`
	LogIndex    string `json:"logIndex,omitempty"`
	Event       string `json:"event,omitempty"`
	From        string `json:"from,omitempty"`
	GasLimit    string `json:"gasLimit,omitempty"`
	GasPrice    string `json:"gasPrice,omitempty"`
	GasSent     string `json:"gasSent,omitempty"`
	Hash        string `json:"hash,omitempty"`
	Index       string `json:"index,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
	To          string `json:"to,omitempty"`
	Value       string `json:"value,omitempty"`
}

// Token represents a token entity
type Token struct {
	ID                         string       `json:"id"`
	Transaction                *Transaction `json:"transaction,omitempty"`
	VaultID                    string       `json:"vault,omitempty"`
	ActivationBlock            string       `json:"activationBlock,omitempty"`
	Blacklisted                bool         `json:"blacklisted,omitempty"`
	Decimals                   string       `json:"decimals,omitempty"`
	DepositFee                 string       `json:"depositFee,omitempty"`
	IsNative                   bool         `json:"isNative,omitempty"`
	Name                       string       `json:"name,omitempty"`
	NamePrefix                 string       `json:"namePrefix,omitempty"`
	NativeAddr                 string       `json:"nativeAddr,omitempty"`
	NativeWid                  string       `json:"nativeWid,omitempty"`
	Symbol                     string       `json:"symbol,omitempty"`
	SymbolPrefix               string       `json:"symbolPrefix,omitempty"`
	TokenAddress               string       `json:"tokenAddress,omitempty"`
	TotalSupply                string       `json:"totalSupply,omitempty"`
	TotalValueLocked           string       `json:"totalValueLocked,omitempty"`
	TotalValueLockedUSD        string       `json:"totalValueLockedUSD,omitempty"`
	TxCount                    string       `json:"txCount,omitempty"`
	Volume                     string       `json:"volume,omitempty"`
	VolumeUSD                  string       `json:"volumeUSD,omitempty"`
	UntrackedVolumeUSD         string       `json:"untrackedVolumeUSD,omitempty"`
	FeesUSD                    string       `json:"feesUSD,omitempty"`
	PoolCount                  string       `json:"poolCount,omitempty"`
	DerivedNative              string       `json:"derivedNative,omitempty"`
	DerivedETH                 string       `json:"derivedETH,omitempty"`
	TotalValueLockedUSDUntracked string     `json:"totalValueLockedUSDUntracked,omitempty"`
}

// Vault represents a vault entity
type Vault struct {
	ID                       string `json:"id"`
	DefaultAlienDepositFee   string `json:"defaultAlienDepositFee,omitempty"`
	DefaultAlienWithdrawFee  string `json:"defaultAlienWithdrawFee,omitempty"`
	DefaultNativeDepositFee  string `json:"defaultNativeDepositFee,omitempty"`
	DefaultNativeWithdrawFee string `json:"defaultNativeWithdrawFee,omitempty"`
}

// Factory represents a factory entity
type Factory struct {
	ID                          string `json:"id"`
	PoolCount                   string `json:"poolCount,omitempty"`
	TxCount                     string `json:"txCount,omitempty"`
	TotalVolumeUSD              string `json:"totalVolumeUSD,omitempty"`
	Owner                       string `json:"owner,omitempty"`
	TotalFeesNative             string `json:"totalFeesNative,omitempty"`
	TotalFeesETH                string `json:"totalFeesETH,omitempty"`
	TotalFeesUSD                string `json:"totalFeesUSD,omitempty"`
	TotalValueLockedNative      string `json:"totalValueLockedNative,omitempty"`
	TotalValueLockedETH         string `json:"totalValueLockedETH,omitempty"`
	TotalValueLockedUSD         string `json:"totalValueLockedUSD,omitempty"`
	TotalVolumeNative           string `json:"totalVolumeNative,omitempty"`
	TotalVolumeETH              string `json:"totalVolumeETH,omitempty"`
	UntrackedVolumeUSD          string `json:"untrackedVolumeUSD,omitempty"`
	TotalValueLockedETHUntracked string `json:"totalValueLockedETHUntracked,omitempty"`
	TotalValueLockedUSDUntracked string `json:"totalValueLockedUSDUntracked,omitempty"`
}

// Swap represents a swap transaction
type Swap struct {
	ID           string `json:"id"`
	Origin       string `json:"origin,omitempty"`
	Recipient    string `json:"recipient,omitempty"`
	Sender       string `json:"sender,omitempty"`
	SqrtPriceX96 string `json:"sqrtPriceX96,omitempty"`
	Tick         string `json:"tick,omitempty"`
	Timestamp    string `json:"timestamp,omitempty"`
	AmountUSD    string `json:"amountUSD,omitempty"`
	LogIndex     string `json:"logIndex,omitempty"`
}

// Bundle represents price data
type Bundle struct {
	ID            string `json:"id"`
	NativePriceUSD string `json:"nativePriceUSD,omitempty"`
	EthPriceUSD   string `json:"ethPriceUSD,omitempty"`
}

// MetaData represents The Graph metadata
type MetaData struct {
	Deployment        string `json:"deployment"`
	HasIndexingErrors bool   `json:"hasIndexingErrors"`
}

// QueryResponse is a generic response structure for GraphQL queries
type QueryResponse struct {
	Transactions []Transaction `json:"transactions,omitempty"`
	Tokens       []Token       `json:"tokens,omitempty"`
	Vaults       []Vault       `json:"vaults,omitempty"`
	Factories    []Factory     `json:"factories,omitempty"`
	Swaps        []Swap        `json:"swaps,omitempty"`
	Bundles      []Bundle      `json:"bundles,omitempty"`
	Meta         MetaData      `json:"_meta,omitempty"`
} 