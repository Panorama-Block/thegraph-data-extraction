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
	ID                           string       `json:"id"`
	Transaction                  *Transaction `json:"transaction,omitempty"`
	VaultID                      string       `json:"vault,omitempty"`
	ActivationBlock              string       `json:"activationBlock,omitempty"`
	Blacklisted                  bool         `json:"blacklisted,omitempty"`
	Decimals                     string       `json:"decimals,omitempty"`
	DepositFee                   string       `json:"depositFee,omitempty"`
	IsNative                     bool         `json:"isNative,omitempty"`
	Name                         string       `json:"name,omitempty"`
	NamePrefix                   string       `json:"namePrefix,omitempty"`
	NativeAddr                   string       `json:"nativeAddr,omitempty"`
	NativeWid                    string       `json:"nativeWid,omitempty"`
	Symbol                       string       `json:"symbol,omitempty"`
	SymbolPrefix                 string       `json:"symbolPrefix,omitempty"`
	TokenAddress                 string       `json:"tokenAddress,omitempty"`
	TotalSupply                  string       `json:"totalSupply,omitempty"`
	TotalValueLocked             string       `json:"totalValueLocked,omitempty"`
	TotalValueLockedUSD          string       `json:"totalValueLockedUSD,omitempty"`
	TxCount                      string       `json:"txCount,omitempty"`
	Volume                       string       `json:"volume,omitempty"`
	VolumeUSD                    string       `json:"volumeUSD,omitempty"`
	UntrackedVolumeUSD           string       `json:"untrackedVolumeUSD,omitempty"`
	FeesUSD                      string       `json:"feesUSD,omitempty"`
	PoolCount                    string       `json:"poolCount,omitempty"`
	DerivedNative                string       `json:"derivedNative,omitempty"`
	DerivedETH                   string       `json:"derivedETH,omitempty"`
	TotalValueLockedUSDUntracked string       `json:"totalValueLockedUSDUntracked,omitempty"`
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
	ID                           string `json:"id"`
	PoolCount                    string `json:"poolCount,omitempty"`
	TxCount                      string `json:"txCount,omitempty"`
	TotalVolumeUSD               string `json:"totalVolumeUSD,omitempty"`
	Owner                        string `json:"owner,omitempty"`
	TotalFeesNative              string `json:"totalFeesNative,omitempty"`
	TotalFeesETH                 string `json:"totalFeesETH,omitempty"`
	TotalFeesUSD                 string `json:"totalFeesUSD,omitempty"`
	TotalValueLockedNative       string `json:"totalValueLockedNative,omitempty"`
	TotalValueLockedETH          string `json:"totalValueLockedETH,omitempty"`
	TotalValueLockedUSD          string `json:"totalValueLockedUSD,omitempty"`
	TotalVolumeNative            string `json:"totalVolumeNative,omitempty"`
	TotalVolumeETH               string `json:"totalVolumeETH,omitempty"`
	UntrackedVolumeUSD           string `json:"untrackedVolumeUSD,omitempty"`
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
	ID             string `json:"id"`
	NativePriceUSD string `json:"nativePriceUSD,omitempty"`
	EthPriceUSD    string `json:"ethPriceUSD,omitempty"`
}

// MetaData represents The Graph metadata
type MetaData struct {
	Deployment        string `json:"deployment"`
	HasIndexingErrors bool   `json:"hasIndexingErrors"`
}

// Withdraw represents a withdrawal transaction in avax
type Withdraw struct {
	Amount    string `json:"amount"`
	Fee       string `json:"fee"`
	ID        string `json:"id"`
	IsNative  bool   `json:"isNative"`
	PayloadId string `json:"payloadId"`
}

// Burn represents a burn transaction in dex
type Burn struct {
	Amount    string `json:"amount"`
	Amount0   string `json:"amount0"`
	Amount1   string `json:"amount1"`
	AmountUSD string `json:"amountUSD"`
	ID        string `json:"id"`
	LogIndex  string `json:"logIndex"`
	Origin    string `json:"origin"`
	Owner     string `json:"owner"`
	TickLower string `json:"tickLower"`
	TickUpper string `json:"tickUpper"`
	Timestamp string `json:"timestamp"`
}

// Account represents a user account
type Account struct {
	ID string `json:"id"`
}

// Pools represents a collection of liquidity pools data
type Pools struct {
	BalanceOfBlock         string `json:"balanceOfBlock"`
	CollectedFeesToken0    string `json:"collectedFeesToken0"`
	CollectedFeesToken1    string `json:"collectedFeesToken1"`
	CollectedFeesUSD       string `json:"collectedFeesUSD"`
	CreatedAtBlockNumber   string `json:"createdAtBlockNumber"`
	CreatedAtTimestamp     string `json:"createdAtTimestamp"`
	FeeGrowthBlock         string `json:"feeGrowthBlock"`
	FeeGrowthGlobal0X128   string `json:"feeGrowthGlobal0X128"`
	FeeGrowthGlobal1X128   string `json:"feeGrowthGlobal1X128"`
	FeeTier                string `json:"feeTier"`
	FeesUSD                string `json:"feesUSD"`
	ID                     string `json:"id"`
	Liquidity              string `json:"liquidity"`
	LiquidityProviderCount string `json:"liquidityProviderCount"`
	ObservationIndex       string `json:"observationIndex"`
	ProtocolFeeToken0      string `json:"protocolFeeToken0"`
	ProtocolFeeToken1      string `json:"protocolFeeToken1"`
	SqrtPrice              string `json:"sqrtPrice"`
	Tick                   string `json:"tick"`
	Token0Price            string `json:"token0Price"`
	Token1Price            string `json:"token1Price"`
	TotalValueLockedNative string `json:"totalValueLockedNative"`
	TotalValueLockedToken0 string `json:"totalValueLockedToken0"`
	TotalValueLockedToken1 string `json:"totalValueLockedToken1"`
	TotalValueLockedUSD    string `json:"totalValueLockedUSD"`
	TxCount                string `json:"txCount"`
	UntrackedVolumeUSD     string `json:"untrackedVolumeUSD"`
	VolumeToken0           string `json:"volumeToken0"`
	VolumeToken1           string `json:"volumeToken1"`
	VolumeUSD              string `json:"volumeUSD"`
}

// SkimFees represents skim fees data in avax
type SkimFees struct {
	Amount    string `json:"amount"`
	ID        string `json:"id"`
	SkimToEve string `json:"skimToEve"`
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
	Withdraws    []Withdraw    `json:"withdraws,omitempty"`
	Burns        []Burn        `json:"burns,omitempty"`
	Accounts     []Account     `json:"accounts,omitempty"`
	Pools        []Pools       `json:"pools,omitempty"`
	SkimFees     []SkimFees    `json:"skimFees,omitempty"`
}
