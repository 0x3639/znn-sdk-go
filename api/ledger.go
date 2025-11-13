package api

import (
	"math/big"

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

func (la *LedgerApi) PublishRawTransaction(transaction *nom.AccountBlock) error {
	var ans interface{}
	if err := la.client.Call(&ans, "ledger.publishRawTransaction", transaction); err != nil {
		return err
	}
	return nil
}

// Unconfirmed AccountBlocks
func (la *LedgerApi) GetUnconfirmedBlocksByAddress(address types.Address, pageIndex, pageSize uint32) (*api.AccountBlockList, error) {
	ans := new(api.AccountBlockList)
	if err := la.client.Call(ans, "ledger.getUnconfirmedBlocksByAddress", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// AccountBlocks
func (la *LedgerApi) GetFrontierAccountBlock(address types.Address) (*api.AccountBlock, error) {
	ans := new(api.AccountBlock)
	if err := la.client.Call(ans, "ledger.getFrontierAccountBlock", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

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

func (la *LedgerApi) GetAccountInfoByAddress(address types.Address) (*api.AccountInfo, error) {
	ans := new(api.AccountInfo)
	if err := la.client.Call(ans, "ledger.getAccountInfoByAddress", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (la *LedgerApi) GetUnreceivedBlocksByAddress(address types.Address, pageIndex, pageSize uint32) (*api.AccountBlockList, error) {
	ans := new(api.AccountBlockList)
	if err := la.client.Call(ans, "ledger.getUnreceivedBlocksByAddress", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// Momentum
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

func (la *LedgerApi) SendTemplate(toAddress types.Address, tokenStandard types.ZenonTokenStandard, amount *big.Int, data []byte) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     toAddress,
		TokenStandard: tokenStandard,
		Amount:        amount,
		Data:          data,
	}
}

func (la *LedgerApi) ReceiveTemplate(fromBlockHash types.Hash) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserReceive,
		FromBlockHash: fromBlockHash,
	}
}
