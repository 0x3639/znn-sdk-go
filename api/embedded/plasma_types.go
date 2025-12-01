package embedded

import (
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// PlasmaInfo represents plasma information for an address.
//
// Plasma is the resource required to send transactions on the Zenon Network.
// It can be obtained by fusing QSR (generating plasma over time) or by
// computing Proof-of-Work for individual transactions.
//
// Fields:
//   - CurrentPlasma: Available plasma units for transactions
//   - MaxPlasma: Maximum plasma capacity based on fused QSR
//   - QsrAmount: Total QSR fused for this address (in base units, 8 decimals)
//
// Plasma regenerates over time proportional to the fused QSR amount.
// More fused QSR means faster regeneration and higher maximum capacity.
//
// Example:
//
//	plasma, err := client.PlasmaApi.Get(address)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Plasma: %d / %d (Fused: %s QSR)\n",
//	    plasma.CurrentPlasma, plasma.MaxPlasma, plasma.QsrAmount)
type PlasmaInfo struct {
	CurrentPlasma uint64   `json:"currentPlasma"`
	MaxPlasma     uint64   `json:"maxPlasma"`
	QsrAmount     *big.Int `json:"qsrAmount"`
}

// plasmaInfoJSON is used for JSON unmarshaling with string amounts
type plasmaInfoJSON struct {
	CurrentPlasma uint64 `json:"currentPlasma"`
	MaxPlasma     uint64 `json:"maxPlasma"`
	QsrAmount     string `json:"qsrAmount"`
}

func (p *PlasmaInfo) UnmarshalJSON(data []byte) error {
	var aux plasmaInfoJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.CurrentPlasma = aux.CurrentPlasma
	p.MaxPlasma = aux.MaxPlasma
	p.QsrAmount = common.StringToBigInt(aux.QsrAmount)
	return nil
}

// FusionEntry represents a single plasma fusion entry.
//
// When QSR is fused for plasma, a FusionEntry is created tracking the fusion.
// The beneficiary receives plasma from this fusion, which may be different
// from the address that provided the QSR.
//
// Fields:
//   - QsrAmount: Amount of QSR fused (in base units, 8 decimals)
//   - Beneficiary: Address that receives plasma from this fusion
//   - ExpirationHeight: Momentum height when the fusion can be canceled
//   - Id: Unique identifier for this fusion entry
//
// After the lock period expires (ExpirationHeight), the fusion can be canceled
// to reclaim the fused QSR.
type FusionEntry struct {
	QsrAmount        *big.Int      `json:"qsrAmount"`
	Beneficiary      types.Address `json:"beneficiary"`
	ExpirationHeight uint64        `json:"expirationHeight"`
	Id               types.Hash    `json:"id"`
}

// fusionEntryJSON is used for JSON unmarshaling with string amounts
type fusionEntryJSON struct {
	QsrAmount        string        `json:"qsrAmount"`
	Beneficiary      types.Address `json:"beneficiary"`
	ExpirationHeight uint64        `json:"expirationHeight"`
	Id               types.Hash    `json:"id"`
}

func (f *FusionEntry) UnmarshalJSON(data []byte) error {
	var aux fusionEntryJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	f.QsrAmount = common.StringToBigInt(aux.QsrAmount)
	f.Beneficiary = aux.Beneficiary
	f.ExpirationHeight = aux.ExpirationHeight
	f.Id = aux.Id
	return nil
}

// FusionEntryList represents a paginated list of fusion entries.
//
// This type is returned by methods that list fusions, such as GetEntriesByAddress.
// It includes the total QSR amount across all fusions for the queried address.
//
// Fields:
//   - QsrAmount: Total QSR fused across all entries (in base units, 8 decimals)
//   - Count: Total number of fusion entries matching the query
//   - List: Slice of FusionEntry entries for the current page
type FusionEntryList struct {
	QsrAmount *big.Int       `json:"qsrAmount"`
	Count     int            `json:"count"`
	List      []*FusionEntry `json:"list"`
}

// fusionEntryListJSON is used for JSON unmarshaling with string amounts
type fusionEntryListJSON struct {
	QsrAmount string         `json:"qsrAmount"`
	Count     int            `json:"count"`
	List      []*FusionEntry `json:"list"`
}

func (f *FusionEntryList) UnmarshalJSON(data []byte) error {
	var aux fusionEntryListJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	f.QsrAmount = common.StringToBigInt(aux.QsrAmount)
	f.Count = aux.Count
	f.List = aux.List
	return nil
}

// GetRequiredParam represents parameters for GetRequiredPoWForAccountBlock.
//
// This type is used when querying the PoW difficulty required for a transaction
// based on the sender's available plasma and the transaction details.
//
// Fields:
//   - Address: Sender address
//   - BlockType: Type of account block being created
//   - ToAddress: Destination address for the transaction
//   - Data: Transaction data payload
type GetRequiredParam struct {
	Address   types.Address `json:"address"`
	BlockType uint64        `json:"blockType"`
	ToAddress types.Address `json:"toAddress"`
	Data      []byte        `json:"data"`
}

// GetRequiredResult represents the result of GetRequiredPoWForAccountBlock.
//
// This type indicates whether a transaction can be sent using available plasma
// or if Proof-of-Work must be computed. If RequiredDifficulty is 0, the
// transaction can proceed using plasma alone.
//
// Fields:
//   - AvailablePlasma: Current plasma available for the address
//   - BasePlasma: Minimum plasma required for this transaction type
//   - RequiredDifficulty: PoW difficulty to compute (0 if plasma is sufficient)
//
// Example:
//
//	result, err := client.PlasmaApi.GetRequiredPoWForAccountBlock(params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result.RequiredDifficulty > 0 {
//	    fmt.Printf("PoW required, difficulty: %d\n", result.RequiredDifficulty)
//	} else {
//	    fmt.Println("Sufficient plasma, no PoW needed")
//	}
type GetRequiredResult struct {
	AvailablePlasma    uint64 `json:"availablePlasma"`
	BasePlasma         uint64 `json:"basePlasma"`
	RequiredDifficulty uint64 `json:"requiredDifficulty"`
}
