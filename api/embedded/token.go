package embedded

import (
	"math/big"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/api"
	"github.com/zenon-network/go-zenon/rpc/api/embedded"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type TokenApi struct {
	client *server.Client
}

func NewTokenApi(client *server.Client) *TokenApi {
	return &TokenApi{
		client: client,
	}
}

// GetAll retrieves a paginated list of all tokens issued on the Zenon Network.
//
// This includes both native tokens (ZNN, QSR) and all ZTS (Zenon Token Standard)
// tokens created by users. Each token entry contains metadata like name, symbol,
// domain, supply information, and ownership details.
//
// Parameters:
//   - pageIndex: Page number (0-indexed)
//   - pageSize: Number of tokens per page (typically 10-50)
//
// Returns a TokenList containing token entries or an error if the query fails.
//
// Example:
//
//	// Get first page of all tokens
//	tokens, err := client.TokenApi.GetAll(0, 25)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Total tokens: %d\n", tokens.Count)
//	for _, token := range tokens.List {
//	    fmt.Printf("%s (%s): %s\n", token.Name, token.Symbol, token.TokenStandard)
//	}
func (ta *TokenApi) GetAll(pageIndex, pageSize uint32) (*embedded.TokenList, error) {
	ans := new(embedded.TokenList)
	if err := ta.client.Call(ans, "embedded.token.getAll", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetByOwner retrieves a paginated list of tokens owned by a specific address.
//
// Token owners can:
//   - Update token properties (if enabled)
//   - Mint additional supply (if mintable)
//   - Transfer ownership to another address
//
// This is useful for:
//   - Managing your own tokens
//   - Auditing tokens owned by an address
//   - Tracking token portfolios
//
// Parameters:
//   - address: Token owner address
//   - pageIndex: Page number (0-indexed)
//   - pageSize: Number of tokens per page
//
// Returns a TokenList of owned tokens or an error.
//
// Example:
//
//	ownerAddr := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
//	tokens, err := client.TokenApi.GetByOwner(ownerAddr, 0, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Tokens owned: %d\n", tokens.Count)
//	for _, token := range tokens.List {
//	    fmt.Printf("- %s (%s)\n", token.Name, token.Symbol)
//	}
func (ta *TokenApi) GetByOwner(address types.Address, pageIndex, pageSize uint32) (*embedded.TokenList, error) {
	ans := new(embedded.TokenList)
	if err := ta.client.Call(ans, "embedded.token.getByOwner", address, pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetByZts retrieves detailed information about a specific token by its ZTS identifier.
//
// Returns complete token metadata including:
//   - Name, symbol, and domain
//   - Total supply, max supply, and decimals
//   - Owner address
//   - Properties: isMintable, isBurnable, isUtility
//   - Token standard identifier (ZTS)
//
// Use this to:
//   - Verify token properties before transactions
//   - Check supply limits before minting
//   - Validate token ownership
//   - Display token information in applications
//
// Parameters:
//   - zts: Token standard identifier (types.ZnnTokenStandard, types.QsrTokenStandard, or custom ZTS)
//
// Returns detailed Token information or an error if the token doesn't exist.
//
// Example:
//
//	// Get ZNN token info
//	token, err := client.TokenApi.GetByZts(types.ZnnTokenStandard)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Token: %s (%s)\n", token.Name, token.Symbol)
//	fmt.Printf("Total Supply: %s\n", token.TotalSupply)
//	fmt.Printf("Decimals: %d\n", token.Decimals)
//	fmt.Printf("Mintable: %t, Burnable: %t\n", token.IsMintable, token.IsBurnable)
func (ta *TokenApi) GetByZts(zts types.ZenonTokenStandard) (*api.Token, error) {
	ans := new(api.Token)
	if err := ta.client.Call(ans, "embedded.token.getByZts", zts.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

// Contract calls

// IssueToken creates a transaction template to issue a new ZTS token on Zenon Network.
//
// Token issuance allows anyone to create custom tokens following the ZTS (Zenon Token
// Standard). The issuing address becomes the token owner with special privileges.
//
// Token requirements:
//   - Cost: 1 ZNN (burned as protocol fee)
//   - Name: 1-40 characters
//   - Symbol: 1-10 uppercase characters
//   - Domain: Optional metadata field (max 128 chars)
//   - TotalSupply: Initial minted supply (can be 0)
//   - MaxSupply: Maximum possible supply (must be >= totalSupply)
//   - Decimals: 0-18 (typically 8 for ZNN/QSR compatibility)
//
// Token properties (permanent once set false):
//   - isMintable: Owner can mint additional supply up to maxSupply
//   - isBurnable: Token holders can burn their tokens
//   - isUtility: Marks token as utility token (metadata flag)
//
// Parameters:
//   - tokenName: Human-readable token name
//   - tokenSymbol: Short trading symbol (uppercase)
//   - tokenDomain: Optional domain/URL for token metadata
//   - totalSupply: Initial supply to mint (in base units)
//   - maxSupply: Maximum total supply allowed (in base units)
//   - decimals: Number of decimal places (0-18)
//   - isMintable: Enable future minting by owner
//   - isBurnable: Enable token burning by holders
//   - isUtility: Flag as utility token
//
// Returns an unsigned AccountBlock template that must be:
//  1. Autofilled with account details
//  2. Enhanced with PoW/plasma
//  3. Signed with keypair
//  4. Published via PublishRawTransaction
//
// Example - Issue mintable token:
//
//	totalSupply := big.NewInt(1000000 * 100000000) // 1M tokens with 8 decimals
//	maxSupply := big.NewInt(10000000 * 100000000)  // 10M max
//
//	template := client.TokenApi.IssueToken(
//	    "My Token",           // name
//	    "MTK",                // symbol
//	    "example.com",        // domain
//	    totalSupply,          // initial supply
//	    maxSupply,            // maximum supply
//	    8,                    // decimals
//	    true,                 // mintable
//	    true,                 // burnable
//	    false,                // not utility token
//	)
//	// Process through transaction pipeline and publish
//
// Example - Issue fixed supply token:
//
//	supply := big.NewInt(21000000 * 100000000) // 21M tokens (Bitcoin-style)
//
//	template := client.TokenApi.IssueToken(
//	    "Fixed Token",
//	    "FTK",
//	    "",
//	    supply,              // total supply
//	    supply,              // max = total (fixed)
//	    8,
//	    false,               // not mintable
//	    false,               // not burnable
//	    false,
//	)
//
// Note: Token issuance costs 1 ZNN which is burned. Ensure you have sufficient balance
// plus plasma/PoW for the transaction.
func (ta *TokenApi) IssueToken(tokenName, tokenSymbol, tokenDomain string, totalSupply, maxSupply *big.Int, decimals uint8, isMintable, isBurnable, isUtility bool) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.TokenContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        constants.TokenIssueAmount,
		Data: definition.ABIToken.PackMethodPanic(
			definition.IssueMethodName,
			tokenName,
			tokenSymbol,
			tokenDomain,
			totalSupply,
			maxSupply,
			decimals,
			isMintable,
			isBurnable,
			isUtility,
		),
	}
}

// Mint creates a transaction template to mint additional tokens.
//
// Minting requirements:
//   - Must be called by token owner
//   - Token must have isMintable = true
//   - Cannot exceed maxSupply limit
//   - Newly minted tokens sent to receiver address
//
// This is commonly used for:
//   - Increasing token supply over time
//   - Reward distribution programs
//   - Protocol-controlled token emission
//   - Gradual token launches
//
// Parameters:
//   - tokenStandard: ZTS identifier of the token to mint
//   - amount: Number of tokens to mint (in base units)
//   - receiver: Address that will receive the newly minted tokens
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example - Mint tokens to specific address:
//
//	zts := types.ParseZTS("zts1your-token-standard")
//	amount := big.NewInt(1000 * 100000000) // 1000 tokens with 8 decimals
//	receiver := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")
//
//	template := client.TokenApi.Mint(zts, amount, receiver)
//	// Sign and publish transaction
//
// Example - Mint to own address:
//
//	template := client.TokenApi.Mint(myTokenZts, amount, myAddress)
//	// Process and publish
//
// Note: Only the token owner can mint. Attempting to mint as non-owner will fail.
// Verify token properties and current supply with GetByZts() before minting.
func (ta *TokenApi) Mint(tokenStandard types.ZenonTokenStandard, amount *big.Int, receiver types.Address) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.TokenContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIToken.PackMethodPanic(
			definition.MintMethodName,
			tokenStandard,
			amount,
			receiver,
		),
	}
}

// Burn creates a transaction template to permanently destroy tokens.
//
// Burning requirements:
//   - Token must have isBurnable = true
//   - Can be called by any token holder
//   - Burns tokens from caller's balance
//   - Reduces total supply permanently
//
// Common burn use cases:
//   - Deflationary tokenomics
//   - Fee mechanisms (burn fees)
//   - Supply management
//   - Protocol buyback and burn programs
//
// Parameters:
//   - tokenStandard: ZTS identifier of the token to burn
//   - amount: Number of tokens to burn (in base units)
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example - Burn tokens from your balance:
//
//	zts := types.ParseZTS("zts1your-token-standard")
//	amount := big.NewInt(500 * 100000000) // Burn 500 tokens
//
//	template := client.TokenApi.Burn(zts, amount)
//	// Sign and publish transaction
//
// Example - Burn 10% of holdings:
//
//	// Get current balance first
//	info, _ := client.LedgerApi.GetAccountInfoByAddress(myAddress)
//	balance := info.BalanceInfoMap[myTokenZts].Balance
//
//	// Burn 10%
//	burnAmount := new(big.Int).Div(balance, big.NewInt(10))
//	template := client.TokenApi.Burn(myTokenZts, burnAmount)
//
// Note: Burning is permanent and irreversible. Ensure you have sufficient balance before
// attempting to burn. Transaction will fail if amount exceeds your balance.
func (ta *TokenApi) Burn(tokenStandard types.ZenonTokenStandard, amount *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.TokenContract,
		TokenStandard: tokenStandard,
		Amount:        amount,
		Data:          definition.ABIToken.PackMethodPanic(definition.BurnMethodName),
	}
}

// UpdateToken creates a transaction template to update token properties.
//
// Update capabilities:
//   - Transfer ownership to a new address
//   - Disable minting (cannot be re-enabled)
//   - Disable burning (cannot be re-enabled)
//
// Update requirements:
//   - Must be called by current token owner
//   - Cannot enable properties that were disabled
//   - Can only disable features, not enable them
//   - Ownership transfer is permanent
//
// Common update scenarios:
//   - Finalize token by disabling minting
//   - Transfer ownership to DAO/multisig
//   - Lock token properties permanently
//   - Decentralize token control
//
// Parameters:
//   - tokenStandard: ZTS identifier of the token to update
//   - owner: New owner address (or current owner to keep unchanged)
//   - isMintable: Minting status (can only change true -> false)
//   - isBurnable: Burning status (can only change true -> false)
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example - Transfer ownership:
//
//	newOwner := types.ParseAddressPanic("z1qqga8s8rkypgsg5qg2g7rp68nqh3r4lkm54tta")
//	template := client.TokenApi.UpdateToken(
//	    myTokenZts,
//	    newOwner,
//	    true,  // keep mintable
//	    true,  // keep burnable
//	)
//	// Sign and publish
//
// Example - Finalize token (disable minting):
//
//	template := client.TokenApi.UpdateToken(
//	    myTokenZts,
//	    myAddress,  // keep same owner
//	    false,      // disable minting permanently
//	    true,       // keep burning enabled
//	)
//	// This makes the token fixed supply
//
// Example - Lock all properties:
//
//	template := client.TokenApi.UpdateToken(
//	    myTokenZts,
//	    myAddress,
//	    false,  // disable minting
//	    false,  // disable burning
//	)
//	// Token is now immutable
//
// Note: Property changes are permanent. Once minting or burning is disabled, it cannot
// be re-enabled. Use with caution.
func (ta *TokenApi) UpdateToken(tokenStandard types.ZenonTokenStandard, owner types.Address, isMintable, isBurnable bool) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.TokenContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIToken.PackMethodPanic(
			definition.UpdateTokenMethodName,
			tokenStandard,
			owner,
			isMintable,
			isBurnable,
		),
	}
}
