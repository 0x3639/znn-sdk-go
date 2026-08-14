package utils

import (
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
)

// Regression test for finding #25: GetTransactionBytes serialized the amount
// with a non-injective encoder, so a negative or over-wide amount produced
// the same transaction hash as a small positive amount (e.g. -5 aliased 251,
// 2^256+5 aliased 5). The signing path must reject out-of-domain amounts
// instead of silently binding a different value than the caller intended.
func TestGetTransactionHashRejectsOutOfDomainAmount(t *testing.T) {
	outOfDomain := []*big.Int{
		big.NewInt(-5),
		new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(5)), // 2^256+5
	}
	for _, amount := range outOfDomain {
		block := &nom.AccountBlock{Amount: amount}
		if _, err := GetTransactionBytes(block); err == nil {
			t.Errorf("GetTransactionBytes accepted out-of-domain amount %s", amount)
		}
		if _, err := GetTransactionHash(block); err == nil {
			t.Errorf("GetTransactionHash accepted out-of-domain amount %s", amount)
		}
	}
}

// A valid amount (and a nil amount treated as zero) must still hash.
func TestGetTransactionHashAcceptsValidAmount(t *testing.T) {
	for _, amount := range []*big.Int{nil, big.NewInt(0), big.NewInt(150000000)} {
		block := &nom.AccountBlock{Amount: amount}
		hash, err := GetTransactionHash(block)
		if err != nil {
			t.Fatalf("GetTransactionHash(amount=%v) error = %v", amount, err)
		}
		if hash == (nom.AccountBlock{}).Hash {
			t.Errorf("GetTransactionHash(amount=%v) returned zero hash", amount)
		}
	}
}
