package bridge

import (
	"encoding/json"
	"fmt"
	"strconv"

	pmtypes "github.com/JoahTheron/polymarket-sdk-v2/shared"
)

// ChainID represents an EVM/SVM chain identifier that accepts both
// JSON numbers and quoted strings.
type ChainID int64

func (c ChainID) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(c), 10))
}

func (c *ChainID) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err == nil {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("parse chain id: %w", err)
		}
		*c = ChainID(parsed)
		return nil
	}
	var numeric int64
	if err := json.Unmarshal(data, &numeric); err != nil {
		return fmt.Errorf("decode chain id: %w", err)
	}
	*c = ChainID(numeric)
	return nil
}

// DepositRequest is the body for creating deposit addresses.
type DepositRequest struct {
	// Address is the user's Polymarket proxy wallet.
	Address string `json:"address"`
}

// DepositAddresses holds deposit addresses across chain families.
type DepositAddresses struct {
	// EVM is the Ethereum-compatible deposit address.
	EVM string `json:"evm"`
	// SVM is the Solana-compatible deposit address.
	SVM string `json:"svm"`
	// BTC is the Bitcoin deposit address.
	BTC string `json:"btc"`
}

// DepositResponse is returned after requesting deposit addresses.
type DepositResponse struct {
	// Address contains deposit addresses by chain family.
	Address DepositAddresses `json:"address"`
	// Note contains optional informational text.
	Note string `json:"note,omitempty"`
}

// Token describes a supported ERC-20 or SPL token.
type Token struct {
	// Name is the token display name.
	Name string `json:"name"`
	// Symbol is the ticker symbol.
	Symbol string `json:"symbol"`
	// Address is the contract address.
	Address string `json:"address"`
	// Decimals is the token precision.
	Decimals pmtypes.Int `json:"decimals"`
}

// SupportedAsset describes a bridgeable asset on a specific chain.
type SupportedAsset struct {
	// ChainID is the source chain identifier.
	ChainID ChainID `json:"chainId"`
	// ChainName is the human-readable chain name.
	ChainName string `json:"chainName"`
	// Token describes the token on this chain.
	Token Token `json:"token"`
	// MinCheckoutUSD is the minimum checkout amount in USD.
	MinCheckoutUSD pmtypes.Float64 `json:"minCheckoutUsd"`
}

// SupportedAssetsResponse lists all bridgeable assets.
type SupportedAssetsResponse struct {
	// SupportedAssets lists the bridgeable assets.
	SupportedAssets []SupportedAsset `json:"supportedAssets"`
	// Note contains optional informational text.
	Note string `json:"note,omitempty"`
}

// QuoteRequest is the body for requesting a bridge quote.
type QuoteRequest struct {
	// FromAmountBaseUnit is the source amount in smallest units.
	FromAmountBaseUnit string `json:"fromAmountBaseUnit"`
	// FromChainID is the source chain identifier.
	FromChainID ChainID `json:"fromChainId"`
	// FromTokenAddress is the source token contract address.
	FromTokenAddress string `json:"fromTokenAddress"`
	// RecipientAddress is the destination wallet address.
	RecipientAddress string `json:"recipientAddress"`
	// ToChainID is the destination chain identifier.
	ToChainID ChainID `json:"toChainId"`
	// ToTokenAddress is the destination token contract address.
	ToTokenAddress string `json:"toTokenAddress"`
}

// EstimatedFeeBreakdown describes the fees for a bridge quote.
type EstimatedFeeBreakdown struct {
	// AppFeeLabel is the application fee description.
	AppFeeLabel string `json:"appFeeLabel"`
	// AppFeePercent is the fee percentage of the transaction.
	AppFeePercent pmtypes.Float64 `json:"appFeePercent"`
	// AppFeeUSD is the application fee in USD.
	AppFeeUSD pmtypes.Float64 `json:"appFeeUsd"`
	// FillCostPercent is the liquidity provider cost percentage.
	FillCostPercent pmtypes.Float64 `json:"fillCostPercent"`
	// FillCostUSD is the liquidity provider cost in USD.
	FillCostUSD pmtypes.Float64 `json:"fillCostUsd"`
	// GasUSD is the estimated gas cost in USD.
	GasUSD pmtypes.Float64 `json:"gasUsd"`
	// MaxSlippage is the maximum expected slippage percentage.
	MaxSlippage pmtypes.Float64 `json:"maxSlippage"`
	// MinReceived is the minimum tokens received after fees.
	MinReceived pmtypes.Float64 `json:"minReceived"`
	// SwapImpact is the swap slippage percentage.
	SwapImpact pmtypes.Float64 `json:"swapImpact"`
	// SwapImpactUSD is the swap slippage in USD.
	SwapImpactUSD pmtypes.Float64 `json:"swapImpactUsd"`
	// TotalImpact is the total fee percentage.
	TotalImpact pmtypes.Float64 `json:"totalImpact"`
	// TotalImpactUSD is the total fee in USD.
	TotalImpactUSD pmtypes.Float64 `json:"totalImpactUsd"`
}

// QuoteResponse is the estimated bridge checkout details.
type QuoteResponse struct {
	// QuoteID is the unique quote identifier.
	QuoteID string `json:"quoteId"`
	// EstCheckoutTimeMs is the estimated time to receive funds.
	EstCheckoutTimeMs pmtypes.Uint64 `json:"estCheckoutTimeMs"`
	// EstFeeBreakdown lists all fees and costs.
	EstFeeBreakdown EstimatedFeeBreakdown `json:"estFeeBreakdown"`
	// EstInputUSD is the input value in USD.
	EstInputUSD pmtypes.Float64 `json:"estInputUsd"`
	// EstOutputUSD is the estimated output value in USD.
	EstOutputUSD pmtypes.Float64 `json:"estOutputUsd"`
	// EstToTokenBaseUnit is the estimated output in smallest units.
	EstToTokenBaseUnit string `json:"estToTokenBaseUnit"`
}

// WithdrawRequest is the body for creating withdrawal addresses.
type WithdrawRequest struct {
	// Address is the user's Polymarket proxy wallet.
	Address string `json:"address"`
	// ToChainID is the destination chain identifier.
	ToChainID ChainID `json:"toChainId"`
	// ToTokenAddress is the destination token contract address.
	ToTokenAddress string `json:"toTokenAddress"`
	// RecipientAddr is the external recipient address.
	RecipientAddr string `json:"recipientAddr"`
}

// WithdrawalAddresses holds withdrawal addresses across chain families.
type WithdrawalAddresses struct {
	// EVM is the Ethereum-compatible withdrawal address.
	EVM string `json:"evm"`
	// SVM is the Solana-compatible withdrawal address.
	SVM string `json:"svm"`
	// BTC is the Bitcoin withdrawal address.
	BTC string `json:"btc"`
}

// WithdrawResponse is returned after requesting withdrawal addresses.
type WithdrawResponse struct {
	// Address contains withdrawal addresses by chain family.
	Address WithdrawalAddresses `json:"address"`
	// Note contains informational text.
	Note string `json:"note"`
}

// DepositTransaction describes a single bridge deposit.
type DepositTransaction struct {
	// FromChainID is the source chain identifier.
	FromChainID ChainID `json:"fromChainId"`
	// FromTokenAddress is the source token contract address.
	FromTokenAddress string `json:"fromTokenAddress"`
	// FromAmountBaseUnit is the deposited amount in smallest units.
	FromAmountBaseUnit string `json:"fromAmountBaseUnit"`
	// ToChainID is the destination chain identifier.
	ToChainID ChainID `json:"toChainId"`
	// ToTokenAddress is the destination token contract address.
	ToTokenAddress string `json:"toTokenAddress"`
	// Status is the transaction processing status.
	Status string `json:"status"`
	// TxHash is the on-chain transaction hash.
	TxHash string `json:"txHash"`
	// CreatedTimeMs is when the transaction was created.
	CreatedTimeMs pmtypes.Uint64 `json:"createdTimeMs"`
}

// StatusResponse lists deposit transactions for a user.
type StatusResponse struct {
	// Transactions lists the user's bridge deposits.
	Transactions []DepositTransaction `json:"transactions"`
}
