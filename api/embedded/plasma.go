package embedded

import (
	"math/big"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type PlasmaApi struct {
	client *server.Client
}

func NewPlasmaApi(client *server.Client) *PlasmaApi {
	return &PlasmaApi{
		client: client,
	}
}

// Get retrieves the current plasma information for an address.
//
// Plasma is Zenon's feeless transaction mechanism. Instead of generating computational
// PoW for each transaction, users can fuse QSR to generate plasma, enabling feeless
// transactions.
//
// Returns PlasmaInfo containing:
//   - CurrentPlasma: Available plasma units
//   - MaxPlasma: Maximum plasma capacity
//   - QsrAmount: Amount of QSR currently fused
//
// Parameters:
//   - address: Account address to query
//
// Example:
//
//	plasmaInfo, err := client.PlasmaApi.Get(address)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Current Plasma: %d\n", plasmaInfo.CurrentPlasma)
//	fmt.Printf("Max Plasma: %d\n", plasmaInfo.MaxPlasma)
//	fmt.Printf("QSR Fused: %s\n", plasmaInfo.QsrAmount)
//
// Use this to check if an address has sufficient plasma for feeless transactions.
func (pa *PlasmaApi) Get(address types.Address) (*PlasmaInfo, error) {
	ans := new(PlasmaInfo)
	if err := pa.client.Call(ans, "embedded.plasma.get", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PlasmaApi) GetEntriesByAddress(address types.Address, pageIndex, pageSize uint32) (*FusionEntryList, error) {
	ans := new(FusionEntryList)
	if err := pa.client.Call(ans, "embedded.plasma.getEntriesByAddress", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetRequiredPoWForAccountBlock calculates the PoW difficulty required for a transaction
// based on available plasma.
//
// This determines whether:
//   - The transaction can be feeless (sufficient plasma)
//   - PoW is needed and how much difficulty
//
// The required PoW decreases as plasma increases. With enough plasma, no PoW is needed.
//
// Parameters:
//   - param: GetRequiredParam containing the account block to check
//
// Returns GetRequiredResult with:
//   - AvailablePlasma: Current plasma available
//   - BasePlasma: Base plasma requirement
//   - RequiredDifficulty: PoW difficulty needed (0 if plasma sufficient)
//
// Example:
//
//	param := embedded.GetRequiredParam{
//	    AccountBlock: accountBlock,
//	    Address:      address,
//	}
//
//	result, err := client.PlasmaApi.GetRequiredPoWForAccountBlock(param)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if result.RequiredDifficulty == 0 {
//	    fmt.Println("Transaction can be feeless")
//	} else {
//	    fmt.Printf("PoW difficulty required: %d\n", result.RequiredDifficulty)
//	    // Generate PoW with this difficulty
//	}
//
// Call this before publishing a transaction to determine if PoW generation is needed.
func (pa *PlasmaApi) GetRequiredPoWForAccountBlock(param GetRequiredParam) (*GetRequiredResult, error) {
	ans := new(GetRequiredResult)
	if err := pa.client.Call(ans, "embedded.plasma.getRequiredPoWForAccountBlock", param); err != nil {
		return nil, err
	}
	return ans, nil
}

// Fuse creates a transaction template to fuse QSR for plasma generation.
//
// Fusing QSR locks it in the plasma contract and generates plasma for the beneficiary
// address. The plasma enables feeless transactions without PoW generation.
//
// Fusion details:
//   - Minimum: 10 QSR
//   - Fused QSR is locked for a period
//   - Plasma generation is proportional to QSR amount
//   - Can be canceled after lock period expires
//
// Parameters:
//   - address: Beneficiary address that will receive the plasma
//   - amount: Amount of QSR to fuse (in base units: 1 QSR = 10^8)
//
// Returns an unsigned AccountBlock template that must be:
//  1. Autofilled with account details
//  2. Enhanced with PoW (since you need plasma to get plasma!)
//  3. Signed with keypair
//  4. Published via PublishRawTransaction
//
// Example - Fuse 10 QSR:
//
//	amount := big.NewInt(10 * 100000000) // 10 QSR
//	template := client.PlasmaApi.Fuse(myAddress, amount)
//	// Process through transaction pipeline and publish
//
// Example - Fuse for another address:
//
//	// Fuse QSR but give plasma to a different address
//	template := client.PlasmaApi.Fuse(beneficiaryAddress, amount)
//	// The sender pays QSR, beneficiary gets plasma
//
// Note: The first fusion requires PoW since you don't have plasma yet. After that,
// the generated plasma enables feeless transactions.
func (pa *PlasmaApi) Fuse(address types.Address, amount *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PlasmaContract,
		TokenStandard: types.QsrTokenStandard,
		Amount:        amount,
		Data:          definition.ABIPlasma.PackMethodPanic(definition.FuseMethodName, address),
	}
}

// Cancel creates a transaction template to cancel a plasma fusion and reclaim QSR.
//
// After the fusion lock period expires, you can cancel the fusion to:
//   - Reclaim your fused QSR
//   - Remove the plasma (plasma will be deducted)
//
// Parameters:
//   - id: Hash ID of the fusion entry to cancel (from GetEntriesByAddress)
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example:
//
//	// Get fusion entries
//	entries, _ := client.PlasmaApi.GetEntriesByAddress(address, 0, 10)
//
//	// Cancel first entry (check if lock period expired first)
//	if len(entries.List) > 0 {
//	    entry := entries.List[0]
//	    template := client.PlasmaApi.Cancel(entry.Id)
//	    // Process and publish transaction
//	}
//
// Note: Canceling requires the fusion lock period to have elapsed. Attempting to
// cancel before the period expires will fail.
func (pa *PlasmaApi) Cancel(id types.Hash) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PlasmaContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABIPlasma.PackMethodPanic(definition.CancelFuseMethodName, id),
	}
}
