package relayer

import pmtypes "github.com/bububa/polymarket-client/shared"

// NonceType identifies the relayer nonce domain to query.
type NonceType string

const (
	// NonceTypeProxy selects the proxy-wallet nonce.
	NonceTypeProxy NonceType = "PROXY"
	// NonceTypeSafe selects the Safe-wallet nonce.
	NonceTypeSafe NonceType = "SAFE"
)

// Credentials contains Relayer API-key authentication headers.
type Credentials struct {
	// APIKey is the relayer API key.
	APIKey string
	// Address is the owner address for APIKey.
	Address string
}

// BuilderCredentials contains builder-key authentication material for relayer requests.
type BuilderCredentials struct {
	// APIKey is the builder API key.
	APIKey string
	// Secret is the URL-safe base64 builder API secret.
	Secret string
	// Passphrase is the builder API passphrase.
	Passphrase string
}

// SignatureParams contains Safe transaction signature parameters.
type SignatureParams struct {
	// GasPrice is the Safe gas price parameter.
	GasPrice string `json:"gasPrice"`
	// Operation is the Safe operation parameter.
	Operation string `json:"operation"`
	// SafeTxGas is the Safe transaction gas parameter.
	SafeTxGas string `json:"safeTxnGas"`
	// BaseGas is the Safe base gas parameter.
	BaseGas string `json:"baseGas"`
	// GasToken is the token used for gas refunds.
	GasToken string `json:"gasToken"`
	// RefundReceiver is the address receiving gas refunds.
	RefundReceiver string `json:"refundReceiver"`
}

// SubmitTransactionRequest is the request body for POST /submit.
type SubmitTransactionRequest struct {
	// From is the signer address.
	From string `json:"from"`
	// To is the target contract address.
	To string `json:"to"`
	// ProxyWallet is the user's Polymarket proxy wallet.
	ProxyWallet string `json:"proxyWallet"`
	// Data is 0x-prefixed encoded transaction calldata.
	Data string `json:"data"`
	// Nonce is the relayer transaction nonce.
	Nonce string `json:"nonce"`
	// Signature is the 0x-prefixed transaction signature.
	Signature string `json:"signature"`
	// SignatureParams are Safe transaction parameters.
	SignatureParams SignatureParams `json:"signatureParams"`
	// Type is the transaction type, typically SAFE or PROXY.
	Type string `json:"type"`
	// Metadata is optional caller-provided transaction metadata.
	Metadata string `json:"metadata,omitempty"`
	// Value is the native token value for the call when required.
	Value string `json:"value,omitempty"`
}

// SubmitTransactionResponse is returned immediately after a transaction is accepted.
type SubmitTransactionResponse struct {
	// TransactionID is the relayer transaction identifier.
	TransactionID string `json:"transactionID"`
	// State is the current relayer state.
	State string `json:"state"`
}

// Transaction describes a relayer transaction.
type Transaction struct {
	// TransactionID is the relayer transaction identifier.
	TransactionID string `json:"transactionID"`
	// State is the current relayer state.
	State string `json:"state"`
	// TransactionHash is the on-chain hash after broadcast.
	TransactionHash string `json:"transactionHash"`
	// From is the signer address.
	From string `json:"from"`
	// To is the target contract address.
	To string `json:"to"`
	// ProxyWallet is the user's Polymarket proxy wallet.
	ProxyWallet string `json:"proxyWallet"`
	// Data is the 0x-prefixed calldata submitted to the relayer.
	Data string `json:"data"`
	// Nonce is the transaction nonce.
	Nonce pmtypes.String `json:"nonce"`
	// Value is the transaction value.
	Value pmtypes.String `json:"value"`
	// Signature is the 0x-prefixed transaction signature.
	Signature string `json:"signature"`
	// Type is the transaction type.
	Type string `json:"type"`
	// Owner is the transaction owner.
	Owner string `json:"owner"`
	// Metadata is the transaction metadata.
	Metadata string `json:"metadata"`
	// CreatedAt is the creation timestamp.
	CreatedAt pmtypes.Time `json:"createdAt"`
	// UpdatedAt is the last update timestamp.
	UpdatedAt pmtypes.Time `json:"updatedAt"`
}

// NonceResponse is returned by nonce endpoints.
type NonceResponse struct {
	// Nonce is the current relayer nonce.
	Nonce pmtypes.String `json:"nonce"`
	// Address is the associated relayer or user address when present.
	Address string `json:"address,omitempty"`
}

// SafeDeployedResponse reports whether a Safe wallet is deployed.
type SafeDeployedResponse struct {
	// Address is the Safe wallet address queried.
	Address string `json:"address,omitempty"`
	// Deployed is true when the Safe is deployed.
	Deployed bool `json:"deployed"`
}

// APIKey describes a relayer API key.
type APIKey struct {
	// Key is the API key.
	Key string `json:"apiKey"`
	// Address is the owner address.
	Address string `json:"address"`
	// CreatedAt is the key creation timestamp.
	CreatedAt pmtypes.Time `json:"createdAt"`
	// UpdatedAt is the key update timestamp.
	UpdatedAt pmtypes.Time `json:"updatedAt"`
}
