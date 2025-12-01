package embedded

import (
	"encoding/base64"
	"encoding/json"
	"math/big"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
)

// HtlcInfo represents HTLC (Hashed Timelock Contract) information.
//
// HTLCs enable trustless atomic swaps between parties by locking tokens with
// both a hash lock (requires revealing a preimage) and a time lock (expires
// after a deadline). This allows for cross-chain atomic swaps and payment
// channels.
//
// Fields:
//   - Id: Unique identifier for this HTLC
//   - TimeLocked: Address that can reclaim tokens after expiration
//   - HashLocked: Address that can claim tokens by revealing the preimage
//   - TokenStandard: ZTS identifier of the locked tokens
//   - Amount: Locked amount (in base units, 8 decimals)
//   - ExpirationTime: Unix timestamp when the time lock expires
//   - HashType: Hash algorithm used (0 = SHA256, 1 = SHA3)
//   - KeyMaxSize: Maximum size of the preimage in bytes
//   - HashLock: Hash that must be satisfied to claim tokens
//
// HTLC Flow:
//  1. Creator locks tokens with a hash lock and time lock
//  2. HashLocked address reveals preimage to claim tokens
//  3. If not claimed before expiration, TimeLocked can reclaim
//
// Example:
//
//	htlc, err := client.HtlcApi.GetById(htlcId)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if htlc.ExpirationTime > time.Now().Unix() {
//	    fmt.Printf("HTLC active, amount: %s\n", htlc.Amount)
//	}
type HtlcInfo struct {
	Id             types.Hash               `json:"id"`
	TimeLocked     types.Address            `json:"timeLocked"`
	HashLocked     types.Address            `json:"hashLocked"`
	TokenStandard  types.ZenonTokenStandard `json:"tokenStandard"`
	Amount         *big.Int                 `json:"amount"`
	ExpirationTime int64                    `json:"expirationTime"`
	HashType       uint8                    `json:"hashType"`
	KeyMaxSize     uint8                    `json:"keyMaxSize"`
	HashLock       []byte                   `json:"hashLock"`
}

// htlcInfoJSON is used for JSON unmarshaling with string amounts
type htlcInfoJSON struct {
	Id             types.Hash               `json:"id"`
	TimeLocked     types.Address            `json:"timeLocked"`
	HashLocked     types.Address            `json:"hashLocked"`
	TokenStandard  types.ZenonTokenStandard `json:"tokenStandard"`
	Amount         string                   `json:"amount"`
	ExpirationTime int64                    `json:"expirationTime"`
	HashType       uint8                    `json:"hashType"`
	KeyMaxSize     uint8                    `json:"keyMaxSize"`
	HashLock       string                   `json:"hashLock"`
}

func (h *HtlcInfo) UnmarshalJSON(data []byte) error {
	var aux htlcInfoJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	h.Id = aux.Id
	h.TimeLocked = aux.TimeLocked
	h.HashLocked = aux.HashLocked
	h.TokenStandard = aux.TokenStandard
	h.Amount = common.StringToBigInt(aux.Amount)
	h.ExpirationTime = aux.ExpirationTime
	h.HashType = aux.HashType
	h.KeyMaxSize = aux.KeyMaxSize
	// Decode base64 hashLock
	if aux.HashLock != "" {
		decoded, err := base64.StdEncoding.DecodeString(aux.HashLock)
		if err != nil {
			return err
		}
		h.HashLock = decoded
	}
	return nil
}
