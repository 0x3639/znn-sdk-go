package wallet

import (
	"errors"
	"testing"
)

func TestWalletErrorPreservesMessageAndType(t *testing.T) {
	err := NewWalletError("wallet operation failed")
	if err.Error() != "wallet operation failed" {
		t.Fatalf("Error() = %q", err.Error())
	}
	var walletErr *WalletError
	if !errors.As(err, &walletErr) || walletErr.Message != "wallet operation failed" {
		t.Fatalf("wallet error = %#v", err)
	}
}
