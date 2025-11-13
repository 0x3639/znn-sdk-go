package wallet

import "fmt"

// BIP44 https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
//
// m / purpose' / coin_type' / account' / change / address_index
//
// Zenon uses:
// m / 44' / 73404' / account' / 0 / 0

const (
	// CoinType is the BIP44 coin type for Zenon (73404')
	CoinType = "73404"

	// DerivationPath is the base BIP44 path for Zenon wallets
	DerivationPath = "m/44'/" + CoinType + "'"
)

// GetDerivationAccount returns the BIP44 derivation path for a given account index
// For example: account 0 returns "m/44'/73404'/0'"
func GetDerivationAccount(account int) string {
	return fmt.Sprintf("%s/%d'", DerivationPath, account)
}
