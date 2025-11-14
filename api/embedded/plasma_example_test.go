package embedded_test

import (
	"fmt"
	"log"
	"math/big"

	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/zenon-network/go-zenon/common/types"
)

// Example_checkPlasma demonstrates checking plasma availability for an address.
func Example_checkPlasma() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get plasma info
	plasmaInfo, err := client.PlasmaApi.Get(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Current Plasma: %d\n", plasmaInfo.CurrentPlasma)
	fmt.Printf("Max Plasma: %d\n", plasmaInfo.MaxPlasma)
	fmt.Printf("QSR Fused: %s\n", plasmaInfo.QsrAmount)

	if plasmaInfo.CurrentPlasma > 0 {
		fmt.Println("Has plasma for feeless transactions")
	} else {
		fmt.Println("No plasma - PoW required for transactions")
	}
}

// Example_fuseQSR demonstrates creating a plasma fusion transaction.
func Example_fuseQSR() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	myAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Fuse 10 QSR for plasma
	amount := big.NewInt(10 * 100000000) // 10 QSR in base units

	template := client.PlasmaApi.Fuse(myAddress, amount)

	fmt.Println("Plasma fusion transaction created")
	fmt.Printf("Fusing %s QSR\n", template.Amount)
	fmt.Printf("Beneficiary: %s\n", myAddress)
	fmt.Println("Token: QSR")

	// Note: Template must be autofilled, enhanced with PoW, signed, and published
	// First fusion requires PoW since you don't have plasma yet
}

// Example_fuseForAnotherAddress demonstrates fusing QSR for a different address.
func Example_fuseForAnotherAddress() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Sender pays QSR, beneficiary gets plasma
	beneficiary := types.ParseAddressPanic("z1qqga8s8rkypgsg5qg2g7rp68nqh3r4lkm54tta")
	amount := big.NewInt(20 * 100000000) // 20 QSR

	template := client.PlasmaApi.Fuse(beneficiary, amount)

	fmt.Println("Fusing QSR for another address")
	fmt.Printf("Beneficiary will receive plasma: %s\n", beneficiary)
	fmt.Printf("Amount: %s QSR\n", template.Amount)

	// Useful for:
	// - Setting up new accounts with plasma
	// - Gifting plasma to others
	// - Service providers funding client accounts
}

// Example_checkRequiredPoW demonstrates checking PoW requirements for a transaction.
func Example_checkRequiredPoW() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	// Create a sample transaction template
	template := client.LedgerApi.SendTemplate(
		types.ParseAddressPanic("z1qqga8s8rkypgsg5qg2g7rp68nqh3r4lkm54tta"),
		types.ZnnTokenStandard,
		big.NewInt(100000000), // 1 ZNN
		[]byte{},
	)

	// Check required PoW (would need proper param structure in practice)
	fmt.Println("Checking PoW requirements for transaction")
	fmt.Printf("Transaction to: %s\n", template.ToAddress)
	fmt.Printf("Amount: %s\n", template.Amount)

	// In practice, you would:
	// 1. Get plasma info
	// 2. Call GetRequiredPoWForAccountBlock with proper params
	// 3. Generate PoW if needed or use plasma
	fmt.Println("PoW check determines if transaction can be feeless")
}

// Example_listFusionEntries demonstrates listing all plasma fusions for an address.
func Example_listFusionEntries() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	myAddress := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get fusion entries
	entries, err := client.PlasmaApi.GetEntriesByAddress(myAddress, 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total fusion entries: %d\n", entries.Count)

	if entries.Count > 0 {
		fmt.Println("Active fusions:")
		for i, entry := range entries.Fusions {
			fmt.Printf("%d. Amount: %s QSR, ID: %s...\n",
				i+1, entry.QsrAmount, entry.Id.String()[:16])
		}
	} else {
		fmt.Println("No active plasma fusions")
	}
}

// Example_cancelFusion demonstrates canceling a plasma fusion to reclaim QSR.
func Example_cancelFusion() {
	client, err := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	address := types.ParseAddressPanic("z1qqjnwjjpnue8xmmpanz6csze6tcmtzzdtfsww7")

	// Get fusion entries
	entries, err := client.PlasmaApi.GetEntriesByAddress(address, 0, 10)
	if err != nil {
		log.Fatal(err)
	}

	if entries.Count == 0 {
		fmt.Println("No fusions to cancel")
		return
	}

	// Cancel first entry (in practice, check if lock period expired)
	entry := entries.Fusions[0]
	_ = client.PlasmaApi.Cancel(entry.Id)

	fmt.Println("Fusion cancellation transaction created")
	fmt.Printf("Canceling fusion ID: %s...\n", entry.Id.String()[:16])
	fmt.Printf("Will reclaim: %s QSR\n", entry.QsrAmount)

	// Note: Must wait for lock period to expire before canceling
	// Plasma will be deducted when fusion is canceled
}
