package embedded_test

import (
	"fmt"
	"log"
	"math/big"

	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example_queryToken demonstrates retrieving token information by ZTS.
func Example_queryToken() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Query ZNN token info
	token, err := client.TokenApi.GetByZts(types.ZnnTokenStandard)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Token: %s (%s)\n", token.TokenName, token.TokenSymbol)
	fmt.Printf("Total Supply: %s\n", token.TotalSupply)
	fmt.Printf("Max Supply: %s\n", token.MaxSupply)
	fmt.Printf("Decimals: %d\n", token.Decimals)
	fmt.Printf("Owner: %s\n", token.Owner)
	fmt.Printf("Mintable: %t, Burnable: %t\n", token.IsMintable, token.IsBurnable)
}

// Example_listAllTokens demonstrates listing all tokens on the network.
func Example_listAllTokens() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Get first page of tokens
	tokens, err := client.TokenApi.GetAll(0, 25)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total tokens: %d\n", tokens.Count)
	fmt.Println("\nTokens on network:")
	for i, token := range tokens.List {
		fmt.Printf("%d. %s (%s)\n", i+1, token.TokenName, token.TokenSymbol)
		fmt.Printf("   ZTS: %s\n", token.ZenonTokenStandard)
		fmt.Printf("   Supply: %s / %s\n", token.TotalSupply, token.MaxSupply)
	}
}

// Example_listTokensByOwner demonstrates listing tokens owned by an address.
func Example_listTokensByOwner() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	ownerAddr := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get tokens owned by address
	tokens, err := client.TokenApi.GetByOwner(ownerAddr, 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tokens owned by %s: %d\n", ownerAddr, tokens.Count)

	if tokens.Count > 0 {
		fmt.Println("\nOwned tokens:")
		for _, token := range tokens.List {
			fmt.Printf("- %s (%s)\n", token.TokenName, token.TokenSymbol)
			fmt.Printf("  Can mint: %t, Can burn: %t\n", token.IsMintable, token.IsBurnable)
		}
	} else {
		fmt.Println("No tokens owned by this address")
	}
}

// Example_issueToken demonstrates creating a new ZTS token.
func Example_issueToken() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Token parameters
	tokenName := "My Token"
	tokenSymbol := "MTK"
	tokenDomain := "example.com"

	// Initial supply: 1M tokens with 8 decimals
	totalSupply := big.NewInt(1000000 * 100000000)
	// Max supply: 10M tokens
	maxSupply := big.NewInt(10000000 * 100000000)

	// Create token issuance template
	template := client.TokenApi.IssueToken(
		tokenName,
		tokenSymbol,
		tokenDomain,
		totalSupply,
		maxSupply,
		8,     // decimals
		true,  // mintable
		true,  // burnable
		false, // not utility token
	)

	fmt.Println("Token issuance transaction created")
	fmt.Printf("Token: %s (%s)\n", tokenName, tokenSymbol)
	fmt.Printf("Initial supply: %s\n", totalSupply)
	fmt.Printf("Maximum supply: %s\n", maxSupply)
	fmt.Println("Properties: Mintable, Burnable")
	fmt.Printf("Cost: 1 ZNN (to %s)\n", template.ToAddress)

	// Note: Template must be autofilled, enhanced with PoW, signed, and published
	// Token ZTS will be generated after confirmation
}

// Example_issueFixedSupplyToken demonstrates creating a token with fixed supply.
func Example_issueFixedSupplyToken() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Fixed supply: 21M tokens (like Bitcoin)
	supply := big.NewInt(21000000 * 100000000)

	template := client.TokenApi.IssueToken(
		"Fixed Token",
		"FTK",
		"",
		supply, // total supply
		supply, // max = total (fixed)
		8,
		false, // NOT mintable
		false, // NOT burnable
		false,
	)

	fmt.Println("Fixed supply token created")
	fmt.Printf("Total supply: %s (immutable)\n", supply)
	fmt.Printf("Cost: %s %s\n", template.Amount, template.TokenStandard)
	fmt.Println("No minting or burning allowed")

	// This token's supply cannot change after issuance
}

// Example_mintTokens demonstrates minting additional tokens.
func Example_mintTokens() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Token to mint (must be mintable and you must be owner)
	myTokenZts := types.ParseZTSPanic("zts1your-token-standard")

	// Mint 1000 tokens
	amount := big.NewInt(1000 * 100000000)
	receiver := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	_ = client.TokenApi.Mint(myTokenZts, amount, receiver)

	fmt.Println("Token mint transaction created")
	fmt.Printf("Minting: %s tokens\n", amount)
	fmt.Printf("Receiver: %s\n", receiver)
	fmt.Printf("Token: %s\n", myTokenZts)

	// Note: Only token owner can mint
	// Must be processed through transaction pipeline and published
}

// Example_burnTokens demonstrates burning tokens permanently.
func Example_burnTokens() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Token to burn (must be burnable)
	tokenZts := types.ParseZTSPanic("zts1your-token-standard")

	// Burn 500 tokens from your balance
	amount := big.NewInt(500 * 100000000)

	template := client.TokenApi.Burn(tokenZts, amount)

	fmt.Println("Token burn transaction created")
	fmt.Printf("Burning: %s tokens\n", amount)
	fmt.Printf("Token: %s\n", template.TokenStandard)
	fmt.Printf("Amount to burn: %s\n", template.Amount)

	// Burning reduces total supply permanently
	fmt.Println("\nWarning: Burning is irreversible!")

	// Note: You must have sufficient balance to burn
}

// Example_transferTokenOwnership demonstrates transferring token ownership.
func Example_transferTokenOwnership() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	myTokenZts := types.ParseZTSPanic("zts1your-token-standard")
	newOwner := types.ParseAddressPanic("z1qqga8s8rkypgsg5qg2g7rp68nqh3r4lkm54tta")

	// Transfer ownership while keeping properties unchanged
	_ = client.TokenApi.UpdateToken(
		myTokenZts,
		newOwner,
		true, // keep mintable
		true, // keep burnable
	)

	fmt.Println("Token ownership transfer created")
	fmt.Printf("New owner: %s\n", newOwner)
	fmt.Println("Token properties unchanged")

	// New owner will control minting and further updates
}

// Example_finalizeToken demonstrates making a token immutable.
func Example_finalizeToken() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	myTokenZts := types.ParseZTSPanic("zts1your-token-standard")
	myAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Disable minting to finalize supply
	_ = client.TokenApi.UpdateToken(
		myTokenZts,
		myAddress, // keep same owner
		false,     // disable minting permanently
		true,      // keep burning enabled
	)

	fmt.Println("Token finalization transaction created")
	fmt.Println("Minting will be permanently disabled")
	fmt.Println("Supply becomes fixed after confirmation")

	// Warning: This is permanent and cannot be undone!
	fmt.Println("\nAfter confirmation:")
	fmt.Println("- No more tokens can be minted")
	fmt.Println("- Supply is capped at current total")
	fmt.Println("- Change is irreversible")
}

// Example_checkTokenProperties demonstrates inspecting token details.
func Example_checkTokenProperties() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Check QSR token properties
	token, err := client.TokenApi.GetByZts(types.QsrTokenStandard)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Token Analysis: %s (%s)\n", token.TokenName, token.TokenSymbol)
	fmt.Println("----------------------------------------")

	// Supply information
	fmt.Println("\nSupply:")
	fmt.Printf("  Current: %s\n", token.TotalSupply)
	fmt.Printf("  Maximum: %s\n", token.MaxSupply)

	// Calculate utilization
	if token.MaxSupply.Cmp(big.NewInt(0)) > 0 {
		utilization := new(big.Int).Mul(token.TotalSupply, big.NewInt(100))
		utilization.Div(utilization, token.MaxSupply)
		fmt.Printf("  Utilization: %s%%\n", utilization)
	}

	// Properties
	fmt.Println("\nProperties:")
	fmt.Printf("  Decimals: %d\n", token.Decimals)
	fmt.Printf("  Mintable: %t\n", token.IsMintable)
	fmt.Printf("  Burnable: %t\n", token.IsBurnable)
	fmt.Printf("  Utility: %t\n", token.IsUtility)

	// Ownership
	fmt.Println("\nOwnership:")
	fmt.Printf("  Owner: %s\n", token.Owner)
	fmt.Printf("  Domain: %s\n", token.TokenDomain)

	// Capabilities
	fmt.Println("\nCapabilities:")
	if token.IsMintable {
		fmt.Println("  - Owner can mint additional supply")
	}
	if token.IsBurnable {
		fmt.Println("  - Holders can burn tokens")
	}
	if !token.IsMintable && !token.IsBurnable {
		fmt.Println("  - Token is immutable (fixed supply)")
	}
}
