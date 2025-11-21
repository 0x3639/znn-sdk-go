package embedded

import (
	"math/big"

	sdkembedded "github.com/0x3639/znn-sdk-go/embedded"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

// HtlcApi provides access to the HTLC (Hashed Timelock Contract) embedded contract
type HtlcApi struct {
	client *server.Client
}

// NewHtlcApi creates a new HTLC API instance
func NewHtlcApi(client *server.Client) *HtlcApi {
	return &HtlcApi{
		client: client,
	}
}

// GetById retrieves HTLC information by ID
func (h *HtlcApi) GetById(id types.Hash) (*definition.HtlcInfo, error) {
	ans := new(definition.HtlcInfo)
	if err := h.client.Call(ans, "embedded.htlc.getById", id.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

// GetProxyUnlockStatus retrieves the proxy unlock status for an address
func (h *HtlcApi) GetProxyUnlockStatus(address types.Address) (bool, error) {
	var ans bool
	if err := h.client.Call(&ans, "embedded.htlc.getProxyUnlockStatus", address.String()); err != nil {
		return false, err
	}
	return ans, nil
}

// Contract method templates

// Create creates a new HTLC contract
// Parameters:
//   - hashLocked: The address that can unlock with preimage
//   - expirationTime: Unix timestamp when HTLC expires
//   - hashType: 0 for SHA3-256, 1 for SHA-256
//   - keyMaxSize: Maximum size of the preimage key
//   - hashLock: Hash of the preimage
func (h *HtlcApi) Create(
	token types.ZenonTokenStandard,
	amount *big.Int,
	hashLocked types.Address,
	expirationTime int64,
	hashType uint8,
	keyMaxSize uint8,
	hashLock []byte,
) *nom.AccountBlock {
	data, err := sdkembedded.Htlc.EncodeFunction("Create", []interface{}{hashLocked, expirationTime, hashType, keyMaxSize, hashLock})
	if err != nil {
		panic(err)
	}

	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.HtlcContract,
		TokenStandard: token,
		Amount:        amount,
		Data:          data,
	}
}

// Reclaim reclaims an expired HTLC
func (h *HtlcApi) Reclaim(id types.Hash) *nom.AccountBlock {
	data, err := sdkembedded.Htlc.EncodeFunction("Reclaim", []interface{}{id})
	if err != nil {
		panic(err)
	}

	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.HtlcContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          data,
	}
}

// Unlock unlocks an HTLC with the preimage
func (h *HtlcApi) Unlock(id types.Hash, preimage []byte) *nom.AccountBlock {
	data, err := sdkembedded.Htlc.EncodeFunction("Unlock", []interface{}{id, preimage})
	if err != nil {
		panic(err)
	}

	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.HtlcContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          data,
	}
}

// DenyProxyUnlock denies proxy unlock for the caller's address
func (h *HtlcApi) DenyProxyUnlock() *nom.AccountBlock {
	data, err := sdkembedded.Htlc.EncodeFunction("DenyProxyUnlock", []interface{}{})
	if err != nil {
		panic(err)
	}

	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.HtlcContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          data,
	}
}

// AllowProxyUnlock allows proxy unlock for the caller's address
func (h *HtlcApi) AllowProxyUnlock() *nom.AccountBlock {
	data, err := sdkembedded.Htlc.EncodeFunction("AllowProxyUnlock", []interface{}{})
	if err != nil {
		panic(err)
	}

	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.HtlcContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          data,
	}
}
