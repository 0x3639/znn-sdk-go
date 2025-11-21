package api_test

import (
	"fmt"
	"log"
	"math/big"

	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example_queryAccountBalance demonstrates checking ZNN and QSR balances.
func Example_queryAccountBalance() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Parse address
	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get account info
	info, err := client.LedgerApi.GetAccountInfoByAddress(address)
	if err != nil {
		log.Fatal(err)
	}

	// Display balances
	if znnBalance, ok := info.BalanceInfoMap[types.ZnnTokenStandard]; ok {
		fmt.Printf("ZNN Balance: %s\n", znnBalance.Balance)
	} else {
		fmt.Println("ZNN Balance: 0")
	}

	if qsrBalance, ok := info.BalanceInfoMap[types.QsrTokenStandard]; ok {
		fmt.Printf("QSR Balance: %s\n", qsrBalance.Balance)
	} else {
		fmt.Println("QSR Balance: 0")
	}

	fmt.Printf("Account Height: %d\n", info.AccountHeight)
}

// Example_checkUnreceivedBlocks demonstrates checking for unreceived transactions.
func Example_checkUnreceivedBlocks() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get unreceived blocks
	blocks, err := client.LedgerApi.GetUnreceivedBlocksByAddress(address, 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Unreceived blocks: %d\n", blocks.Count)

	if blocks.Count > 0 {
		fmt.Println("Pending incoming transactions detected")
		for _, block := range blocks.List {
			fmt.Printf("From: %s, Amount: %s\n", block.Address, block.Amount)
		}
	} else {
		fmt.Println("No pending transactions")
	}
}

// Example_getCurrentBlockchainHeight demonstrates getting the latest momentum.
func Example_getCurrentBlockchainHeight() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Get latest momentum
	momentum, err := client.LedgerApi.GetFrontierMomentum()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Current blockchain height: %d\n", momentum.Height)
	fmt.Printf("Timestamp: %v\n", momentum.Timestamp)
	fmt.Printf("Hash: %s...\n", momentum.Hash.String()[:16])
}

// Example_createSendTransaction demonstrates creating a send transaction template.
func Example_createSendTransaction() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Recipient address
	toAddress := types.ParseAddressPanic("z1qqga8s8rkypgsg5qg2g7rp68nqh3r4lkm54tta")

	// Amount: 10 ZNN (10 * 10^8 base units)
	amount := big.NewInt(10 * 100000000)

	// Create send template
	template := client.LedgerApi.SendTemplate(
		toAddress,
		types.ZnnTokenStandard,
		amount,
		[]byte{}, // no data
	)

	fmt.Println("Send transaction template created")
	fmt.Printf("To: %s\n", template.ToAddress)
	fmt.Printf("Amount: %s\n", template.Amount)
	fmt.Printf("Token: %s\n", template.TokenStandard)

	// Note: Template must be autofilled, enhanced with PoW, signed, and published
	// This example only shows template creation
}

// Example_createReceiveTransaction demonstrates creating a receive transaction template.
func Example_createReceiveTransaction() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get unreceived blocks
	blocks, err := client.LedgerApi.GetUnreceivedBlocksByAddress(address, 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	if blocks.Count == 0 {
		fmt.Println("No blocks to receive")
		return
	}

	// Create receive template for first unreceived block
	firstBlock := blocks.List[0]
	template := client.LedgerApi.ReceiveTemplate(firstBlock.Hash)

	fmt.Println("Receive transaction template created")
	fmt.Printf("Receiving block: %s\n", template.FromBlockHash)
	fmt.Printf("Block type: %d\n", template.BlockType)

	// Note: Template must be autofilled, enhanced with PoW, signed, and published
}

// Example_checkTransactionConfirmation demonstrates verifying a transaction was confirmed.
func Example_checkTransactionConfirmation() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Example transaction hash (would be from a real transaction)
	blockHash := types.HexToHashPanic("0000000000000000000000000000000000000000000000000000000000000001")

	// Try to get the block
	block, err := client.LedgerApi.GetAccountBlockByHash(blockHash)
	if err != nil {
		fmt.Println("Transaction not found or not yet confirmed")
		return
	}

	fmt.Println("Transaction confirmed")
	fmt.Printf("Confirmed at height: %d\n", block.Height)
	fmt.Printf("Amount: %s\n", block.Amount)
}

// Example_monitorPendingTransactions demonstrates checking unconfirmed blocks.
func Example_monitorPendingTransactions() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get unconfirmed blocks
	blocks, err := client.LedgerApi.GetUnconfirmedBlocksByAddress(address, 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Pending transactions: %d\n", blocks.Count)

	if blocks.Count > 0 {
		fmt.Println("Transactions waiting for confirmation:")
		for _, block := range blocks.List {
			fmt.Printf("- Hash: %s..., Type: %d\n", block.Hash.String()[:16], block.BlockType)
		}
	} else {
		fmt.Println("No pending transactions")
	}
}

// Example_sendWithData demonstrates creating a transaction with arbitrary data.
func Example_sendWithData() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	toAddress := types.ParseAddressPanic("z1qqga8s8rkypgsg5qg2g7rp68nqh3r4lkm54tta")
	amount := big.NewInt(1 * 100000000) // 1 ZNN

	// Add custom data to transaction
	data := []byte("Payment for invoice #12345")

	template := client.LedgerApi.SendTemplate(
		toAddress,
		types.ZnnTokenStandard,
		amount,
		data,
	)

	fmt.Println("Transaction with data created")
	fmt.Printf("Data length: %d bytes\n", len(template.Data))
	fmt.Printf("Data: %s\n", string(template.Data))

	// Template must still be processed through full transaction pipeline
}
