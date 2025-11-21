package api

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/api"
	"github.com/zenon-network/go-zenon/rpc/server"
)

type LedgerApi struct {
	client *server.Client
}

func NewLedgerApi(client *server.Client) *LedgerApi {
	return &LedgerApi{
		client: client,
	}
}

// PublishRawTransaction submits a signed and finalized transaction to the Zenon Network.
//
// This is the final step in the transaction flow. Before calling this method, the
// transaction must be:
//  1. Created (via SendTemplate, ReceiveTemplate, or embedded contract API)
//  2. Autofilled with height, previous hash, and momentum acknowledgment
//  3. Enhanced with PoW nonce or plasma
//  4. Signed with a keypair
//  5. Hash computed
//
// Parameters:
//   - transaction: Fully prepared AccountBlock ready for submission
//
// Returns an error if the transaction is rejected by the node. Common rejection reasons:
//   - Insufficient PoW/plasma
//   - Invalid signature
//   - Incorrect height or previous hash
//   - Insufficient balance
//   - Invalid contract call parameters
//
// Example:
//
//	// Assuming transaction is fully prepared
//	err := client.LedgerApi.PublishRawTransaction(transaction)
//	if err != nil {
//	    log.Printf("Transaction failed: %v", err)
//	    return err
//	}
//	fmt.Println("Transaction published successfully")
//
// Note: A successful publish means the transaction was accepted by the node, but it
// still needs to be confirmed in a momentum. Use GetAccountBlockByHash to check confirmation.
func (la *LedgerApi) PublishRawTransaction(transaction *nom.AccountBlock) error {
	var ans interface{}
	if err := la.client.Call(&ans, "ledger.publishRawTransaction", transaction); err != nil {
		return err
	}
	return nil
}

// PublishRawTransactionWithRetry publishes a transaction with automatic retry logic
// for transient failures.
//
// This method wraps PublishRawTransaction with exponential backoff retry logic,
// making it more resilient to temporary network issues, node synchronization delays,
// or transient errors.
//
// Retry behavior:
//   - Retries only on transient errors (network errors, timeouts, temporary unavailability)
//   - Does NOT retry on permanent errors (invalid signature, insufficient balance, etc.)
//   - Uses exponential backoff: 1s, 2s, 4s, 8s, ...
//   - Maximum backoff delay capped at 30 seconds
//
// Parameters:
//   - transaction: Fully prepared AccountBlock ready for submission
//   - maxRetries: Maximum number of retry attempts (0 = no retries, just one attempt)
//
// Returns an error if all retry attempts fail or if a permanent error is encountered.
//
// Example - Standard retry (3 attempts):
//
//	err := client.LedgerApi.PublishRawTransactionWithRetry(transaction, 3)
//	if err != nil {
//	    log.Printf("Transaction failed after retries: %v", err)
//	    return err
//	}
//	fmt.Println("Transaction published successfully")
//
// Example - No retries (same as PublishRawTransaction):
//
//	err := client.LedgerApi.PublishRawTransactionWithRetry(transaction, 0)
//
// Example - Aggressive retry for critical transactions:
//
//	err := client.LedgerApi.PublishRawTransactionWithRetry(transaction, 5)
//
// Common transient errors that trigger retry:
//   - Connection errors
//   - Timeout errors
//   - "connection refused" or "connection reset"
//   - Temporary node unavailability
//
// Permanent errors (no retry):
//   - Invalid signature
//   - Insufficient balance
//   - Incorrect account height
//   - Invalid PoW/plasma
//   - Malformed transaction data
func (la *LedgerApi) PublishRawTransactionWithRetry(transaction *nom.AccountBlock, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Attempt to publish
		err := la.PublishRawTransaction(transaction)
		if err == nil {
			// Success!
			return nil
		}

		// Save the error
		lastErr = err

		// Check if this is a transient error worth retrying
		if !isTransientError(err) {
			// Permanent error - don't retry
			return fmt.Errorf("permanent error, not retrying: %w", err)
		}

		// Check if we have retries left
		if attempt >= maxRetries {
			// No more retries
			break
		}

		// Calculate backoff delay with exponential backoff
		// 1s, 2s, 4s, 8s, 16s, 30s (capped)
		// Cap shift amount to prevent overflow (attempt is always >= 0 in our loop)
		shiftAmount := attempt
		if shiftAmount > 5 {
			shiftAmount = 5 // Cap at 2^5 = 32 seconds
		}
		// #nosec G115 - shiftAmount is capped to prevent overflow
		backoff := time.Duration(1<<uint(shiftAmount)) * time.Second
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}

		// Wait before retry
		time.Sleep(backoff)
	}

	// All retries exhausted
	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isTransientError determines if an error is transient (retry-worthy) or permanent.
//
// Transient errors include:
//   - Network connectivity issues
//   - Timeouts
//   - Temporary node unavailability
//   - Connection resets
//
// Permanent errors include:
//   - Invalid signature
//   - Insufficient balance/plasma
//   - Invalid transaction parameters
//   - Invalid account height or hash
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Transient network errors
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"connection closed",
		"timeout",
		"temporary failure",
		"try again",
		"temporarily unavailable",
		"network unreachable",
		"host unreachable",
		"no route to host",
		"broken pipe",
		"i/o timeout",
		"deadline exceeded",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	// Permanent errors (explicit check)
	permanentPatterns := []string{
		"invalid signature",
		"insufficient",
		"invalid hash",
		"invalid height",
		"invalid data",
		"invalid amount",
		"invalid address",
		"invalid block",
		"account chain",
		"invalid parameter",
		"invalid token",
	}

	for _, pattern := range permanentPatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	// Default: treat unknown errors as transient (safer to retry)
	return true
}

// GetUnconfirmedBlocksByAddress retrieves account blocks that have been published but
// not yet confirmed in a momentum.
//
// Unconfirmed blocks are transactions that:
//   - Have been accepted by the node
//   - Are waiting to be included in a momentum
//   - May still fail validation during momentum confirmation
//
// This is useful for:
//   - Checking pending outgoing transactions
//   - Monitoring transaction status before confirmation
//   - Detecting potential issues with transactions
//
// Parameters:
//   - address: Account address to query
//   - pageIndex: Page number (0-indexed)
//   - pageSize: Number of blocks per page (typically 10-50)
//
// Returns a paginated list of unconfirmed blocks or an error.
//
// Example:
//
//	// Get first page of unconfirmed blocks
//	blocks, err := client.LedgerApi.GetUnconfirmedBlocksByAddress(address, 0, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Pending transactions: %d\n", blocks.Count)
//	for _, block := range blocks.List {
//	    fmt.Printf("Block hash: %s\n", block.Hash)
//	}
func (la *LedgerApi) GetUnconfirmedBlocksByAddress(address types.Address, pageIndex, pageSize uint32) (*api.AccountBlockList, error) {
	ans := new(api.AccountBlockList)
	if err := la.client.Call(ans, "ledger.getUnconfirmedBlocksByAddress", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetFrontierAccountBlock returns the frontier account block for an address
func (la *LedgerApi) GetFrontierAccountBlock(address types.Address) (*api.AccountBlock, error) {
	ans := new(api.AccountBlock)
	if err := la.client.Call(ans, "ledger.getFrontierAccountBlock", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetAccountBlockByHash retrieves a specific account block by its hash.
//
// Use this to:
//   - Check if a transaction was confirmed
//   - Get transaction details
//   - Verify transaction status
//   - Retrieve confirmation height
//
// Parameters:
//   - blockHash: Hash of the account block to retrieve
//
// Returns the account block or an error if the block doesn't exist.
//
// Example:
//
//	block, err := client.LedgerApi.GetAccountBlockByHash(blockHash)
//	if err != nil {
//	    fmt.Println("Block not found or not yet confirmed")
//	    return err
//	}
//
//	fmt.Printf("Block confirmed at height: %d\n", block.Height)
//	fmt.Printf("Amount: %s\n", block.Amount)
//	fmt.Printf("Token: %s\n", block.TokenStandard)
//
// A nil error and non-nil block indicates the transaction is confirmed.
func (la *LedgerApi) GetAccountBlockByHash(blockHash types.Hash) (*api.AccountBlock, error) {
	ans := new(api.AccountBlock)
	if err := la.client.Call(ans, "ledger.getAccountBlockByHash", blockHash.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (la *LedgerApi) GetAccountBlocksByHeight(address types.Address, height, count uint64) (*api.AccountBlockList, error) {
	ans := new(api.AccountBlockList)
	if err := la.client.Call(ans, "ledger.getAccountBlocksByHeight", address.String(), height, count); err != nil {
		return nil, err
	}
	return ans, nil
}

func (la *LedgerApi) GetAccountBlocksByPage(address types.Address, pageIndex, pageSize uint32) (*api.AccountBlockList, error) {
	ans := new(api.AccountBlockList)
	if err := la.client.Call(ans, "ledger.getAccountBlocksByPage", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetAccountInfoByAddress retrieves comprehensive account information including balances
// and account chain state.
//
// Returns account information containing:
//   - Address: The account address
//   - AccountHeight: Number of blocks in the account chain
//   - BalanceInfoMap: Map of token standards to balance information
//
// This is the primary method for checking account balances and state.
//
// Parameters:
//   - address: Zenon address to query
//
// Returns account info or an error if the address doesn't exist or query fails.
//
// Example:
//
//	info, err := client.LedgerApi.GetAccountInfoByAddress(address)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get ZNN balance
//	if znnBalance, ok := info.BalanceInfoMap[types.ZnnTokenStandard]; ok {
//	    fmt.Printf("ZNN Balance: %s\n", znnBalance.Balance)
//	}
//
//	// Get QSR balance
//	if qsrBalance, ok := info.BalanceInfoMap[types.QsrTokenStandard]; ok {
//	    fmt.Printf("QSR Balance: %s\n", qsrBalance.Balance)
//	}
//
//	fmt.Printf("Account Height: %d\n", info.AccountHeight)
//
// Balance amounts are returned in base units (1 ZNN = 10^8 base units).
func (la *LedgerApi) GetAccountInfoByAddress(address types.Address) (*api.AccountInfo, error) {
	ans := new(api.AccountInfo)
	if err := la.client.Call(ans, "ledger.getAccountInfoByAddress", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetUnreceivedBlocksByAddress retrieves incoming transactions that have not yet been
// received by the account.
//
// In Zenon's dual-ledger architecture, receiving funds is a two-step process:
//  1. Sender publishes a send block
//  2. Recipient must publish a receive block to accept the funds
//
// This method returns all send blocks waiting to be received. Each unreceived block
// represents incoming funds or contract calls that need to be accepted.
//
// Parameters:
//   - address: Account address to check for unreceived blocks
//   - pageIndex: Page number (0-indexed)
//   - pageSize: Number of blocks per page (typically 10-50)
//
// Returns a paginated list of unreceived blocks or an error.
//
// Example:
//
//	// Check for unreceived blocks
//	blocks, err := client.LedgerApi.GetUnreceivedBlocksByAddress(address, 0, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Unreceived blocks: %d\n", blocks.Count)
//	for _, block := range blocks.List {
//	    // Create receive template for each block
//	    receiveTemplate := client.LedgerApi.ReceiveTemplate(block.Hash)
//	    // Sign and publish receiveTemplate...
//	}
//
// Note: Unreceived blocks must be received to access the funds. They don't expire but
// remain pending until explicitly received.
func (la *LedgerApi) GetUnreceivedBlocksByAddress(address types.Address, pageIndex, pageSize uint32) (*api.AccountBlockList, error) {
	ans := new(api.AccountBlockList)
	if err := la.client.Call(ans, "ledger.getUnreceivedBlocksByAddress", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetFrontierMomentum retrieves the latest momentum (block) from the network.
//
// Momentums are the backbone of Zenon Network, similar to blocks in other blockchains.
// Each momentum:
//   - Contains a batch of confirmed account blocks
//   - Represents a specific height in the chain
//   - Includes timestamp and hash
//   - Is produced by pillars in a coordinated manner
//
// This is commonly used to:
//   - Get the current blockchain height
//   - Check network liveness
//   - Determine the latest confirmed state
//   - Use as momentum acknowledgment for new transactions
//
// Returns the latest momentum or an error if the network is unreachable.
//
// Example:
//
//	momentum, err := client.LedgerApi.GetFrontierMomentum()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Current height: %d\n", momentum.Height)
//	fmt.Printf("Timestamp: %d\n", momentum.Timestamp)
//	fmt.Printf("Hash: %s\n", momentum.Hash)
//
// The frontier momentum height is used when constructing new transactions
// to acknowledge the current state of the network.
func (la *LedgerApi) GetFrontierMomentum() (*api.Momentum, error) {
	ans := new(api.Momentum)
	if err := la.client.Call(ans, "ledger.getFrontierMomentum"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (la *LedgerApi) GetMomentumBeforeTime(timestamp int64) (*api.Momentum, error) {
	ans := new(api.Momentum)
	if err := la.client.Call(ans, "ledger.getMomentumBeforeTime", timestamp); err != nil {
		return nil, err
	}
	return ans, nil
}

func (la *LedgerApi) GetMomentumByHash(hash types.Hash) (*api.Momentum, error) {
	ans := new(api.Momentum)
	if err := la.client.Call(ans, "ledger.getMomentumByHash", hash.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (la *LedgerApi) GetMomentumsByHeight(height, count uint64) (*api.MomentumList, error) {
	ans := new(api.MomentumList)
	if err := la.client.Call(ans, "ledger.getMomentumsByHeight", height, count); err != nil {
		return nil, err
	}
	return ans, nil
}

func (la *LedgerApi) GetMomentumsByPage(pageIndex, pageSize uint32) (*api.MomentumList, error) {
	ans := new(api.MomentumList)
	if err := la.client.Call(ans, "ledger.getMomentumsByPage", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (la *LedgerApi) GetDetailedMomentumsByHeight(height, count uint64) (*api.DetailedMomentumList, error) {
	ans := new(api.DetailedMomentumList)
	if err := la.client.Call(ans, "ledger.getDetailedMomentumsByHeight", height, count); err != nil {
		return nil, err
	}
	return ans, nil
}

// SendTemplate creates an unsigned transaction template for sending tokens.
//
// This is the starting point for all token transfers (ZNN, QSR, or any ZTS token).
// The template must be further processed before submission:
//  1. Autofill with account height, previous hash, and momentum acknowledgment
//  2. Generate PoW or use plasma
//  3. Sign with keypair
//  4. Publish via PublishRawTransaction
//
// Parameters:
//   - toAddress: Recipient address
//   - tokenStandard: Token to send (types.ZnnTokenStandard, types.QsrTokenStandard, or custom ZTS)
//   - amount: Amount in base units (1 ZNN = 10^8 base units, use big.NewInt)
//   - data: Optional arbitrary data (empty []byte{} for simple transfers)
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example - Send 10 ZNN:
//
//	amount := big.NewInt(10 * 100000000) // 10 ZNN in base units
//	template := client.LedgerApi.SendTemplate(
//	    recipientAddress,
//	    types.ZnnTokenStandard,
//	    amount,
//	    []byte{}, // no data
//	)
//	// Now: autofill, add PoW, sign, and publish
//
// Example - Send with data:
//
//	data := []byte("Payment for invoice #123")
//	template := client.LedgerApi.SendTemplate(
//	    recipientAddress,
//	    types.ZnnTokenStandard,
//	    amount,
//	    data,
//	)
//
// Note: The template is NOT a complete transaction. It must be processed through
// the full transaction pipeline before publishing.
func (la *LedgerApi) SendTemplate(toAddress types.Address, tokenStandard types.ZenonTokenStandard, amount *big.Int, data []byte) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     toAddress,
		TokenStandard: tokenStandard,
		Amount:        amount,
		Data:          data,
	}
}

// ReceiveTemplate creates an unsigned transaction template for receiving tokens.
//
// In Zenon's dual-ledger model, receiving funds requires publishing a receive block
// that references the sender's send block. This method creates that receive template.
//
// The receive process:
//  1. Get unreceived blocks via GetUnreceivedBlocksByAddress
//  2. For each unreceived block, create a receive template with its hash
//  3. Autofill, add PoW/plasma, sign, and publish
//
// Parameters:
//   - fromBlockHash: Hash of the send block to receive (from GetUnreceivedBlocksByAddress)
//
// Returns an unsigned AccountBlock template for receiving.
//
// Example - Receive all unreceived blocks:
//
//	// Get unreceived blocks
//	unreceived, _ := client.LedgerApi.GetUnreceivedBlocksByAddress(myAddress, 0, 10)
//
//	for _, block := range unreceived.List {
//	    // Create receive template
//	    template := client.LedgerApi.ReceiveTemplate(block.Hash)
//
//	    // Autofill, add PoW, sign, and publish
//	    // (full transaction flow needed here)
//	}
//
// Example - Receive specific block:
//
//	template := client.LedgerApi.ReceiveTemplate(sendBlockHash)
//	// Process: autofill -> PoW -> sign -> publish
//
// Note: Like SendTemplate, this is just the first step. The template must go through
// the full transaction pipeline before submission.
func (la *LedgerApi) ReceiveTemplate(fromBlockHash types.Hash) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserReceive,
		FromBlockHash: fromBlockHash,
	}
}
