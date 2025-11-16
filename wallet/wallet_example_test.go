package wallet_test

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/0x3639/znn-sdk-go/wallet"
)

// Example demonstrates creating a new wallet with random mnemonic.
func Example() {
	// Create temporary directory for example
	tempDir, _ := os.MkdirTemp("", "wallet-example-*")
	defer os.RemoveAll(tempDir)

	// Create wallet manager
	manager, err := wallet.NewKeyStoreManager(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	// Create new wallet with random mnemonic
	keystore, err := manager.CreateNew("secure-password", "my-wallet")
	if err != nil {
		log.Fatal(err)
	}

	// Display mnemonic (MUST be backed up securely!)
	fmt.Println("Wallet created successfully")

	// Verify mnemonic was generated
	if len(keystore.Mnemonic) > 0 {
		fmt.Println("Mnemonic generated")
	}

	// Get base address (first address, index 0)
	_, err = keystore.GetBaseAddress()
	if err == nil {
		fmt.Println("Base address derived")
	}

	// Output:
	// Wallet created successfully
	// Mnemonic generated
	// Base address derived
}

// Example_importMnemonic demonstrates restoring a wallet from an existing mnemonic.
func Example_importMnemonic() {
	tempDir, _ := os.MkdirTemp("", "wallet-import-*")
	defer os.RemoveAll(tempDir)

	manager, err := wallet.NewKeyStoreManager(tempDir)
	if err != nil {
		log.Fatal(err)
	}

	// Import wallet from existing mnemonic
	mnemonic := "route become dream access impulse price inform obtain engage ski believe awful"
	keystore, err := manager.CreateFromMnemonic(mnemonic, "new-password", "imported-wallet")
	if err != nil {
		log.Fatal(err)
	}

	// Verify import succeeded
	baseAddr, _ := keystore.GetBaseAddress()
	fmt.Println("Wallet imported successfully")
	fmt.Printf("Address: %s...\n", baseAddr.String()[:12])

	// Same mnemonic always generates same addresses
	fmt.Println("Mnemonic restored consistently")
}

// Example_deriveMultipleAddresses demonstrates deriving multiple addresses from one wallet.
func Example_deriveMultipleAddresses() {
	tempDir, _ := os.MkdirTemp("", "wallet-derive-*")
	defer os.RemoveAll(tempDir)

	manager, _ := wallet.NewKeyStoreManager(tempDir)
	keystore, _ := manager.CreateNew("password", "multi-address-wallet")

	// Derive first 3 addresses
	for i := 0; i < 3; i++ {
		keypair, err := keystore.GetKeyPair(i)
		if err != nil {
			log.Fatal(err)
		}

		address, _ := keypair.GetAddress()
		fmt.Printf("Address %d: %s...\n", i, address.String()[:12])
	}

	// All addresses derived from same mnemonic
	fmt.Println("All addresses from single mnemonic")
}

// Example_bulkAddressGeneration demonstrates efficient bulk address derivation.
func Example_bulkAddressGeneration() {
	tempDir, _ := os.MkdirTemp("", "wallet-bulk-*")
	defer os.RemoveAll(tempDir)

	manager, _ := wallet.NewKeyStoreManager(tempDir)
	keystore, _ := manager.CreateNew("password", "bulk-wallet")

	// Derive addresses 0-4 in one call
	addresses, err := keystore.DeriveAddressesByRange(0, 5)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated %d addresses\n", len(addresses))
	for i, addr := range addresses {
		fmt.Printf("Address %d: %s...\n", i, addr.String()[:12])
	}
}

// Example_findAddress demonstrates searching for an address in the wallet.
func Example_findAddress() {
	tempDir, _ := os.MkdirTemp("", "wallet-find-*")
	defer os.RemoveAll(tempDir)

	manager, _ := wallet.NewKeyStoreManager(tempDir)
	keystore, _ := manager.CreateNew("password", "search-wallet")

	// Get an address to search for
	targetKeypair, _ := keystore.GetKeyPair(5)
	targetAddr, _ := targetKeypair.GetAddress()

	// Find which index this address belongs to
	result, err := keystore.FindAddress(*targetAddr, 10)
	if errors.Is(err, wallet.ErrAddressNotFound) {
		fmt.Println("Address not found")
	} else if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Address found at index: %d\n", result.Index)
		// Can use result.KeyPair to sign transactions
	}

	// Output:
	// Address found at index: 5
}

// Example_saveAndLoadWallet demonstrates wallet persistence.
func Example_saveAndLoadWallet() {
	tempDir, _ := os.MkdirTemp("", "wallet-persist-*")
	defer os.RemoveAll(tempDir)

	manager, _ := wallet.NewKeyStoreManager(tempDir)

	// Create and save wallet
	keystore1, _ := manager.CreateNew("password123", "persistent-wallet")
	addr1, _ := keystore1.GetBaseAddress()
	fmt.Printf("Created wallet: %s...\n", addr1.String()[:12])

	// Load wallet from disk
	keystore2, err := manager.ReadKeyStore("password123", "persistent-wallet")
	if err != nil {
		log.Fatal(err)
	}
	addr2, _ := keystore2.GetBaseAddress()
	fmt.Printf("Loaded wallet: %s...\n", addr2.String()[:12])

	// Verify they match
	if addr1.String() == addr2.String() {
		fmt.Println("Wallet persisted correctly")
	}
}

// Example_listWallets demonstrates listing all wallets in a directory.
func Example_listWallets() {
	tempDir, _ := os.MkdirTemp("", "wallet-list-*")
	defer os.RemoveAll(tempDir)

	manager, _ := wallet.NewKeyStoreManager(tempDir)

	// Create multiple wallets
	manager.CreateNew("password", "wallet-1")
	manager.CreateNew("password", "wallet-2")
	manager.CreateNew("password", "wallet-3")

	// List all wallets
	wallets, err := manager.ListAllKeyStores()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d wallets:\n", len(wallets))
	for _, name := range wallets {
		fmt.Printf("- %s\n", name)
	}
}

// Example_signAndVerify demonstrates cryptographic signing with a keypair.
func Example_signAndVerify() {
	tempDir, _ := os.MkdirTemp("", "wallet-sign-*")
	defer os.RemoveAll(tempDir)

	manager, _ := wallet.NewKeyStoreManager(tempDir)
	keystore, _ := manager.CreateNew("password", "signing-wallet")

	// Get keypair for signing
	keypair, _ := keystore.GetKeyPair(0)

	// Sign a message
	message := []byte("Hello Zenon Network")
	signature, err := keypair.Sign(message)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Signature length: %d bytes\n", len(signature))

	// Verify signature
	valid, err := keypair.Verify(signature, message)
	if err != nil {
		log.Fatal(err)
	}

	if valid {
		fmt.Println("Signature verified successfully")
	} else {
		fmt.Println("Signature verification failed")
	}

	// Output:
	// Signature length: 64 bytes
	// Signature verified successfully
}
