// Package embedded provides API interfaces for interacting with Zenon Network's embedded
// smart contracts. These protocol-level contracts handle core network functionality including
// staking, pillar management, token operations, plasma fusion, and more.
//
// Embedded contracts are built into the Zenon Network protocol and provide essential
// functionality for network operations and governance. Each contract has a dedicated API
// that creates transaction templates for contract calls.
//
// # Available Embedded Contract APIs
//
// Token Operations (TokenApi):
//   - Issue new ZTS tokens
//   - Mint and burn tokens
//   - Update token properties
//
// Staking (StakeApi):
//   - Stake ZNN for network rewards
//   - Cancel stake entries
//   - Collect staking rewards
//
// Plasma (PlasmaApi):
//   - Fuse QSR for feeless transactions
//   - Cancel plasma fusion
//   - Query plasma availability
//
// Pillar Operations (PillarApi):
//   - Register network pillars
//   - Update pillar information
//   - Delegate to pillars
//
// Sentinel Operations (SentinelApi):
//   - Register sentinels
//   - Revoke sentinels
//
// Additional Contract APIs:
//   - AcceleratorApi: Project funding and voting
//   - BridgeApi: Cross-chain bridge operations
//   - LiquidityApi: Liquidity provision
//   - SwapApi: Token swaps
//   - HTLCApi: Hash Time-Locked Contracts for atomic swaps
//   - SporkApi: Network upgrades and governance
//
// # Basic Usage Pattern
//
// All embedded contract methods follow the same pattern:
//  1. Call the contract method to create a transaction template
//  2. Sign and enhance the transaction with PoW/plasma
//  3. Publish the transaction
//
// Example - Staking ZNN:
//
//	// Create stake transaction template
//	template := client.StakeApi.Stake(
//	    durationInDays,
//	    amount,
//	)
//
//	// Sign and publish (implementation depends on your setup)
//	// See examples directory for complete workflows
//
// # Token Operations
//
// Issue a new ZTS token:
//
//	template := client.TokenApi.IssueToken(
//	    "MyToken",           // token name
//	    "MTK",               // token symbol
//	    "example.com",       // domain
//	    totalSupply,         // total supply
//	    maxSupply,           // max supply
//	    8,                   // decimals
//	    true,                // mintable
//	    true,                // burnable
//	    false,               // utility
//	)
//
// # Plasma for Feeless Transactions
//
// Fuse QSR to generate plasma for feeless transactions:
//
//	// Fuse 10 QSR
//	amount := big.NewInt(10 * 100000000) // 10 QSR in base units
//	template := client.PlasmaApi.Fuse(beneficiaryAddress, amount)
//
//	// Check required PoW for a transaction
//	difficulty, err := client.PlasmaApi.GetRequiredPoWForAccountBlock(accountBlock)
//
// # Pillar Delegation
//
// Delegate weight to a pillar:
//
//	template := client.PillarApi.Delegate(pillarName)
//
//	// Undelegate
//	template = client.PillarApi.Undelegate()
//
// # Staking Workflow
//
// Stake ZNN and collect rewards:
//
//	// Stake 100 ZNN for 30 days
//	amount := big.NewInt(100 * 100000000)
//	template := client.StakeApi.Stake(30, amount)
//
//	// Later, collect rewards
//	rewardTemplate := client.StakeApi.CollectReward()
//
// # Accelerator (Zenon Hyperspace)
//
// Create and vote on funding projects:
//
//	template := client.AcceleratorApi.CreateProject(
//	    name,
//	    description,
//	    url,
//	    znnFundsNeeded,
//	    qsrFundsNeeded,
//	)
//
//	// Vote on project
//	voteTemplate := client.AcceleratorApi.VoteByProdAddress(
//	    projectId,
//	    voteOption,
//	)
//
// # Important Notes
//
// - All methods return *nom.AccountBlock templates (unsigned transactions)
// - Templates must be signed with a keypair before submission
// - Most operations require PoW generation or fused plasma
// - Some operations have minimum requirements (e.g., pillar registration requires ZNN deposit)
// - Contract calls may fail if prerequisites aren't met (check embedded contract rules)
//
// # Transaction Requirements
//
// Different operations require different amounts of PoW or plasma:
//   - Simple transfers: Low PoW requirement
//   - Contract calls: Higher PoW requirement
//   - Token issuance: Requires ZNN burn and significant PoW
//
// Use PlasmaApi.GetRequiredPoWForAccountBlock to check requirements before submitting.
//
// For complete examples of embedded contract usage, see the examples directory.
//
// For more information, see https://pkg.go.dev/github.com/0x3639/znn-sdk-go/api/embedded
package embedded
