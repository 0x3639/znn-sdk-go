package wallet

import "errors"

// WalletError represents a wallet-related error
type WalletError struct {
	Message string
}

func (e *WalletError) Error() string {
	return e.Message
}

// NewWalletError creates a new wallet error
func NewWalletError(message string) error {
	return &WalletError{Message: message}
}

// Common wallet errors
var (
	ErrWalletManagerStopped = errors.New("wallet manager has not started")
	ErrIncorrectPassword    = errors.New("incorrect password")
	ErrInvalidMnemonic      = errors.New("invalid mnemonic")
	ErrInvalidEntropy       = errors.New("invalid entropy")
	ErrWalletNotFound       = errors.New("wallet not found")
	ErrWalletAlreadyExists  = errors.New("wallet already exists")
	ErrInvalidKeyStore      = errors.New("invalid keystore")
	ErrInvalidPrivateKey    = errors.New("invalid private key")
	ErrAddressNotFound      = errors.New("address not found in wallet")
	ErrKeystoreNotFound     = errors.New("keystore not found")
)
