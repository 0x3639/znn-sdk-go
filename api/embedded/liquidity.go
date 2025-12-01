package embedded

import (
	"math/big"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type LiquidityApi struct {
	client *server.Client
}

func NewLiquidityApi(client *server.Client) *LiquidityApi {
	return &LiquidityApi{
		client: client,
	}
}

func (sa *LiquidityApi) GetUncollectedReward(address types.Address) (*UncollectedReward, error) {
	ans := new(UncollectedReward)
	if err := sa.client.Call(ans, "embedded.liquidity.getUncollectedReward", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *LiquidityApi) GetFrontierRewardByPage(address types.Address, pageIndex, pageSize uint32) (*RewardHistoryList, error) {
	ans := new(RewardHistoryList)
	if err := sa.client.Call(ans, "embedded.liquidity.getFrontierRewardByPage", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *LiquidityApi) GetLiquidityInfo() (*LiquidityInfo, error) {
	ans := new(LiquidityInfo)
	if err := sa.client.Call(ans, "embedded.liquidity.getLiquidityInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *LiquidityApi) GetSecurityInfo() (*SecurityInfo, error) {
	ans := new(SecurityInfo)
	if err := sa.client.Call(ans, "embedded.liquidity.getSecurityInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *LiquidityApi) GetLiquidityStakeEntriesByAddress(address types.Address, pageIndex, pageSize uint32) (*LiquidityStakeList, error) {
	ans := new(LiquidityStakeList)
	if err := sa.client.Call(ans, "embedded.liquidity.getLiquidityStakeEntriesByAddress", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *LiquidityApi) GetTimeChallengesInfo() (*TimeChallengesList, error) {
	ans := new(TimeChallengesList)
	if err := sa.client.Call(ans, "embedded.liquidity.getTimeChallengesInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

// Contract calls

func (sa *LiquidityApi) SetTokenTupleMethod(tokenStandards []string, znnPercentages []uint32, qsrPercentages []uint32, minAmounts []*big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.LiquidityContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABILiquidity.PackMethodPanic(definition.SetTokenTupleMethodName,
			tokenStandards, znnPercentages, qsrPercentages, minAmounts),
	}
}

func (sa *LiquidityApi) LiquidityStake(durationInSec int64, amount *big.Int, zts types.ZenonTokenStandard) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.LiquidityContract,
		TokenStandard: zts,
		Amount:        amount,
		Data:          definition.ABILiquidity.PackMethodPanic(definition.LiquidityStakeMethodName, durationInSec),
	}
}

func (sa *LiquidityApi) SetIsHalted(value bool) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.LiquidityContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABILiquidity.PackMethodPanic(
			definition.SetIsHaltedMethodName,
			value,
		),
	}
}

func (sa *LiquidityApi) CollectReward() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.LiquidityContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABILiquidity.PackMethodPanic(definition.CollectRewardMethodName),
	}
}

func (sa *LiquidityApi) CancelLiquidity(id types.Hash) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.LiquidityContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABILiquidity.PackMethodPanic(definition.CancelLiquidityStakeMethodName,
			id),
	}
}

func (sa *LiquidityApi) UnlockLiquidityStakeEntries(zts types.ZenonTokenStandard) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.LiquidityContract,
		TokenStandard: zts,
		Amount:        common.Big0,
		Data:          definition.ABILiquidity.PackMethodPanic(definition.UnlockLiquidityStakeEntriesMethodName),
	}
}

func (sa *LiquidityApi) SetAdditionalReward(znnReward *big.Int, qsrAmount *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.LiquidityContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABILiquidity.PackMethodPanic(definition.SetAdditionalRewardMethodName,
			znnReward, qsrAmount),
	}
}

func (sa *LiquidityApi) NominateGuardians(guardians []types.Address) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.LiquidityContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABILiquidity.PackMethodPanic(
			definition.NominateGuardiansMethodName,
			guardians,
		),
	}
}
