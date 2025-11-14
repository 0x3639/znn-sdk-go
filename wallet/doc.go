// Package wallet provides hierarchical deterministic (HD) wallet functionality for the
// Zenon Network, including BIP39 mnemonic generation, BIP32/BIP44 key derivation, and
// encrypted keystore management.
//
// The wallet package enables secure storage and management of Zenon Network keypairs
// using industry-standard cryptographic practices. Wallets are encrypted with Argon2
// key derivation and stored as JSON keyfiles.
//
// # Basic Usage
//
// Create a new wallet with a random mnemonic:
//
//	manager, err := wallet.NewKeyStoreManager("./wallets")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create new wallet with password protection
//	keystore, err := manager.CreateNew("my-secure-password", "main-wallet")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Println("Mnemonic:", keystore.Mnemonic)
//	fmt.Println("Base address:", keystore.GetBaseAddress())
//
// # Key Derivation
//
// The wallet follows BIP44 derivation path: m/44'/73404'/account'/0'/0'
// where 73404 is Zenon's registered coin type.
//
// Derive keypairs at different indices:
//
//	// Get default keypair (index 0)
//	keypair0, err := keystore.GetKeyPair(0)
//	address0, _ := keypair0.GetAddress()
//
//	// Derive multiple addresses
//	keypair1, _ := keystore.GetKeyPair(1)
//	keypair2, _ := keystore.GetKeyPair(2)
//
// # Importing Existing Mnemonics
//
// Import a wallet from an existing BIP39 mnemonic:
//
//	mnemonic := "route become dream access impulse price inform obtain engage ski believe awful"
//	keystore, err := manager.CreateFromMnemonic(mnemonic, "password", "imported-wallet")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Wallet Persistence
//
// Wallets are automatically saved as encrypted keyfiles. Load an existing wallet:
//
//	keystore, err := manager.ReadKeyStore("password", "main-wallet")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// List all wallets in directory
//	wallets, err := manager.ListAllKeyStores()
//	for _, name := range wallets {
//	    fmt.Println("Wallet:", name)
//	}
//
// # Cryptographic Operations
//
// Sign and verify messages with Ed25519:
//
//	keypair, _ := keystore.GetKeyPair(0)
//	message := []byte("Hello Zenon")
//
//	// Sign message
//	signature, err := keypair.Sign(message)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Verify signature
//	valid, err := keypair.Verify(signature, message)
//	if err != nil || !valid {
//	    log.Fatal("Invalid signature")
//	}
//
// # Security Considerations
//
// - Mnemonics should be backed up securely and never shared
// - Passwords should be strong and unique
// - Keyfiles are encrypted but filesystem permissions should be restricted
// - Never commit keyfiles to version control
//
// For more examples, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/wallet
package wallet
