package embedded

import (
	"math/big"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type SentinelApi struct {
	client *server.Client
}

func NewSentinelApi(client *server.Client) *SentinelApi {
	return &SentinelApi{
		client: client,
	}
}

func (sa *SentinelApi) GetByOwner(address types.Address) (*SentinelInfo, error) {
	ans := new(SentinelInfo)
	if err := sa.client.Call(ans, "embedded.sentinel.getByOwner", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *SentinelApi) GetAllActive(pageIndex, pageSize uint32) (*SentinelInfoList, error) {
	ans := new(SentinelInfoList)
	if err := sa.client.Call(ans, "embedded.sentinel.getAllActive", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *SentinelApi) GetDepositedQsr(address types.Address) (*big.Int, error) {
	var ans string
	if err := sa.client.Call(ans, "embedded.sentinel.getDepositedQsr", address); err != nil {
		return nil, err
	}
	return common.StringToBigInt(ans), nil
}

func (sa *SentinelApi) GetUncollectedReward(address types.Address) (*UncollectedReward, error) {
	ans := new(UncollectedReward)
	if err := sa.client.Call(ans, "embedded.sentinel.getUncollectedReward", address); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *SentinelApi) GetFrontierRewardByPage(address types.Address, pageIndex, pageSize uint32) (*RewardHistoryList, error) {
	ans := new(RewardHistoryList)
	if err := sa.client.Call(ans, "embedded.sentinel.getFrontierRewardByPage", address, pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// Contract calls

// Register creates a transaction template to register a new Sentinel.
//
// Sentinels are network infrastructure nodes that provide reliability and support.
// Running a Sentinel requires:
//   - 5,000 ZNN collateral (locked, returned upon revocation)
//   - 50,000 QSR collateral (locked, returned upon revocation)
//   - Dedicated node infrastructure
//
// Sentinel benefits:
//   - Earn ZNN and QSR rewards
//   - Support network infrastructure
//   - Lower barrier than Pillar operation
//   - Collateral fully returned on revocation
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example:
//
//	template := client.SentinelApi.Register()
//	// Sign and publish transaction
//	// Requires 5,000 ZNN + 50,000 QSR
//
// Note: Ensure you have sufficient ZNN and QSR before attempting registration.
func (sa *SentinelApi) Register() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.SentinelContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        constants.SentinelZnnRegisterAmount,
		Data:          definition.ABISentinel.PackMethodPanic(definition.RegisterSentinelMethodName),
	}
}

func (sa *SentinelApi) Revoke() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.SentinelContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABISentinel.PackMethodPanic(definition.RevokeSentinelMethodName),
	}
}

func (sa *SentinelApi) DepositQsr(amount *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.SentinelContract,
		TokenStandard: types.QsrTokenStandard,
		Amount:        amount,
		Data:          definition.ABISentinel.PackMethodPanic(definition.DepositQsrMethodName),
	}
}

func (sa *SentinelApi) WithdrawQsr() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.SentinelContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABISentinel.PackMethodPanic(definition.WithdrawQsrMethodName),
	}
}

func (sa *SentinelApi) CollectReward() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.SentinelContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABISentinel.PackMethodPanic(definition.CollectRewardMethodName),
	}
}
