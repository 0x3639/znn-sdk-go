package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/0x3639/znn-sdk-go/wallet"
)

func main() {
	fmt.Println("Zenon Go SDK - Wallet Management Example")
	fmt.Println("=========================================")

	// Create temporary wallet directory
	tempDir := "./temp_wallets"
	os.MkdirAll(tempDir, 0700)
	defer os.RemoveAll(tempDir)

	// Initialize wallet manager
	manager, err := wallet.NewKeyStoreManager(tempDir)
	if err != nil {
		fmt.Printf("Error initializing manager: %v\n", err)
		return
	}
	fmt.Printf("✓ Wallet manager initialized (directory: %s)\n\n", tempDir)

	// Create new wallet with random mnemonic
	fmt.Println("Creating new wallet...")
	keystore, err := manager.CreateNew("password123", "my-wallet")
	if err != nil {
		fmt.Printf("Error creating wallet: %v\n", err)
		return
	}

	fmt.Printf("✓ Wallet created successfully\n")
	baseAddr, err := keystore.GetBaseAddress()
	if err != nil {
		fmt.Printf("Error getting base address: %v\n", err)
		return
	}
	fmt.Printf("  Base address: %s\n\n", baseAddr.String())

	// Derive keypairs at different indices
	fmt.Println("Deriving keypairs...")
	for i := 0; i < 3; i++ {
		keypair, err := keystore.GetKeyPair(i)
		if err != nil {
			fmt.Printf("Error deriving keypair %d: %v\n", i, err)
			continue
		}

		addr, _ := keypair.GetAddress()
		pubKey, _ := keypair.GetPublicKey()

		fmt.Printf("Account %d:\n", i)
		fmt.Printf("  Address:    %s\n", addr.String())
		fmt.Printf("  Public key: %s...\n", hex.EncodeToString(pubKey)[:32])
	}
	fmt.Println()

	// Sign a message
	fmt.Println("Signing message...")
	keypair, _ := keystore.GetKeyPair(0)
	message := []byte("Hello Zenon Network!")
	signature, err := keypair.Sign(message)
	if err != nil {
		fmt.Printf("Error signing: %v\n", err)
		return
	}
	fmt.Printf("✓ Message signed\n")
	fmt.Printf("  Message: %s\n", string(message))
	fmt.Printf("  Signature: %s...\n\n", hex.EncodeToString(signature)[:32])

	// Verify signature
	valid, err := keypair.Verify(signature, message)
	if err != nil {
		fmt.Printf("Error verifying: %v\n", err)
		return
	}
	if valid {
		fmt.Println("✓ Signature verification: PASSED")
	} else {
		fmt.Println("✗ Signature verification: FAILED")
	}

	// List all wallets
	fmt.Println("\nListing all wallets...")
	wallets, err := manager.ListAllKeyStores()
	if err != nil {
		fmt.Printf("Error listing wallets: %v\n", err)
		return
	}

	fmt.Printf("Found %d wallet(s):\n", len(wallets))
	for _, w := range wallets {
		fmt.Printf("  - %s\n", w)
	}

	// Load existing wallet
	fmt.Println("\nLoading wallet...")
	loadedKeystore, err := manager.ReadKeyStore("password123", "my-wallet")
	if err != nil {
		fmt.Printf("Error loading wallet: %v\n", err)
		return
	}

	fmt.Printf("✓ Wallet loaded successfully\n")
	loadedAddr, _ := loadedKeystore.GetBaseAddress()
	originalAddr, _ := keystore.GetBaseAddress()
	fmt.Printf("  Base address matches: %v\n", loadedAddr.String() == originalAddr.String())

	fmt.Println("\n✓ Example completed successfully")
}
