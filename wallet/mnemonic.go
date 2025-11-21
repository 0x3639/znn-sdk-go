package wallet

import (
	"strings"

	"github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic generates a BIP39 mnemonic with the given entropy strength
// strength must be 128, 160, 192, 224, or 256 bits
// 128 bits = 12 words, 256 bits = 24 words
func GenerateMnemonic(strength int) (string, error) {
	entropy, err := bip39.NewEntropy(strength)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

// ValidateMnemonic validates a BIP39 mnemonic phrase
func ValidateMnemonic(words []string) bool {
	mnemonic := strings.Join(words, " ")
	return bip39.IsMnemonicValid(mnemonic)
}

// ValidateMnemonicString validates a BIP39 mnemonic phrase from a string
func ValidateMnemonicString(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// IsValidWord checks if a word is in the BIP39 wordlist
func IsValidWord(word string) bool {
	wordlist := bip39.GetWordList()
	for _, w := range wordlist {
		if w == word {
			return true
		}
	}
	return false
}

// MnemonicToEntropy converts a mnemonic to its entropy bytes
func MnemonicToEntropy(mnemonic string) ([]byte, error) {
	return bip39.EntropyFromMnemonic(mnemonic)
}

// EntropyToMnemonic converts entropy bytes to a mnemonic
func EntropyToMnemonic(entropy []byte) (string, error) {
	return bip39.NewMnemonic(entropy)
}

// MnemonicToSeed converts a mnemonic to a seed for key derivation
// passphrase can be empty string
func MnemonicToSeed(mnemonic string, passphrase string) []byte {
	return bip39.NewSeed(mnemonic, passphrase)
}
