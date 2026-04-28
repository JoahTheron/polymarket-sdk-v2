package clob

import (
	"time"

	"github.com/JoahTheron/polymarket-sdk-v2/internal/polyauth"
)

// Auth stores the signer and API credentials used for authenticated CLOB requests.
type Auth struct {
	Signer        *polyauth.Signer
	ChainID       int64
	Credentials   *Credentials
	UseServerTime bool
}

// BuildHMACSignature returns the Polymarket L2 HMAC signature for a request.
func BuildHMACSignature(secret string, timestamp int64, method, requestPath string, body []byte) (string, error) {
	return polyauth.HMACSignature(secret, timestamp, method, requestPath, body)
}

func nowUnix() int64 { return time.Now().Unix() }

// ParsePrivateKey parses a hex-encoded Ethereum private key into a Polymarket signer.
func ParsePrivateKey(raw string) (*polyauth.Signer, error) { return polyauth.ParsePrivateKey(raw) }

// GenerateKey creates a new hex-encoded Ethereum private key.
func GenerateKey() (string, error) { return polyauth.GenerateKey() }
