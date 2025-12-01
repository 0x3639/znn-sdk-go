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

type PillarApi struct {
	client *server.Client
}

func NewPillarApi(client *server.Client) *PillarApi {
	return &PillarApi{
		client: client,
	}
}

// GetDepositedQsr retrieves the amount of QSR deposited for a Pillar.
//
// Pillar operators can deposit QSR to increase their momentum rewards weight.
// Deposited QSR can be withdrawn after meeting the lock period requirement.
//
// Parameters:
//   - address: Pillar owner address
//
// Returns deposited QSR amount or an error.
//
// Example:
//
//	deposited, err := client.PillarApi.GetDepositedQsr(pillarAddress)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Deposited QSR: %s\n", deposited)
func (pa *PillarApi) GetDepositedQsr(address types.Address) (*big.Int, error) {
	var ans string
	if err := pa.client.Call(ans, "embedded.pillar.getDepositedQsr", address.String()); err != nil {
		return nil, err
	}
	return common.StringToBigInt(ans), nil
}

// GetQsrRegistrationCost retrieves the current QSR cost for Pillar registration.
//
// The cost may vary based on network parameters. Check this before attempting
// to register a Pillar to ensure sufficient QSR balance.
//
// Returns QSR registration cost or an error.
//
// Example:
//
//	cost, err := client.PillarApi.GetQsrRegistrationCost()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Registration cost: %s QSR\n", cost)
func (pa *PillarApi) GetQsrRegistrationCost() (*big.Int, error) {
	var ans string
	if err := pa.client.Call(ans, "embedded.pillar.getQsrRegistrationCost"); err != nil {
		return nil, err
	}
	return common.StringToBigInt(ans), nil
}

func (pa *PillarApi) GetUncollectedReward(address types.Address) (*UncollectedReward, error) {
	ans := new(UncollectedReward)
	if err := pa.client.Call(ans, "embedded.pillar.getUncollectedReward", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PillarApi) GetFrontierRewardByPage(address types.Address, pageIndex, pageSize uint32) (*RewardHistoryList, error) {
	ans := new(RewardHistoryList)
	if err := pa.client.Call(ans, "embedded.pillar.getFrontierRewardByPage", address.String(), pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PillarApi) GetAll(pageIndex, pageSize uint32) (*PillarInfoList, error) {
	ans := new(PillarInfoList)
	if err := pa.client.Call(ans, "embedded.pillar.getAll", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PillarApi) GetByOwner(address types.Address) ([]*PillarInfo, error) {
	var ans []*PillarInfo
	if err := pa.client.Call(&ans, "embedded.pillar.getByOwner", address.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PillarApi) GetByName(name string) (*PillarInfo, error) {
	ans := new(PillarInfo)
	if err := pa.client.Call(ans, "embedded.pillar.getByName", name); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PillarApi) CheckNameAvailability(name string) (*bool, error) {
	ans := new(bool)
	if err := pa.client.Call(ans, "embedded.pillar.checkNameAvailability", name); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PillarApi) GetDelegatedPillar(address types.Address) (*DelegationInfo, error) {
	ans := new(DelegationInfo)
	if err := pa.client.Call(ans, "embedded.pillar.getDelegatedPillar", address); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PillarApi) GetPillarEpochHistory(pillarName string, pageIndex, pageSize uint32) (*PillarEpochHistoryList, error) {
	ans := new(PillarEpochHistoryList)
	if err := pa.client.Call(ans, "embedded.pillar.getPillarEpochHistory", pillarName, pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (pa *PillarApi) GetPillarsHistoryByEpoch(epoch uint64, pageIndex, pageSize uint32) (*PillarEpochHistoryList, error) {
	ans := new(PillarEpochHistoryList)
	if err := pa.client.Call(ans, "embedded.pillar.getPillarsHistoryByEpoch", epoch, pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// Contract calls

// Register creates a transaction template to register a new Pillar.
//
// Pillars are the backbone consensus nodes of Zenon Network. Running a Pillar requires:
//   - 15,000 ZNN collateral (locked, returned upon revocation)
//   - Variable QSR cost (check GetQsrRegistrationCost)
//   - Dedicated infrastructure (node + block producer)
//   - Unique pillar name (check CheckNameAvailability)
//
// Parameters:
//   - name: Unique pillar name (3-40 characters)
//   - producerAddress: Address that will produce blocks
//   - rewardAddress: Address that receives pillar rewards
//   - blockProducingPercentage: % of block rewards kept by pillar (0-100)
//   - delegationPercentage: % of delegation rewards kept by pillar (0-100)
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example:
//
//	template := client.PillarApi.Register(
//	    "MyPillar",
//	    producerAddr,
//	    rewardAddr,
//	    0,  // Give all block rewards to delegators
//	    50, // Keep 50% of delegation rewards
//	)
//
// Note: Requires 15,000 ZNN + QSR cost. Pillar name cannot be changed after registration.
func (pa *PillarApi) Register(name string, producerAddress, rewardAddress types.Address, blockProducingPercentage, delegationPercentage uint8) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        constants.PillarStakeAmount,
		Data: definition.ABIPillars.PackMethodPanic(
			definition.RegisterMethodName,
			name,
			producerAddress,
			rewardAddress,
			blockProducingPercentage,
			delegationPercentage,
		),
	}
}

func (pa *PillarApi) UpdatePillar(name string, producerAddress, rewardAddress types.Address, blockProducingPercentage, delegationPercentage uint8) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIPillars.PackMethodPanic(
			definition.UpdatePillarMethodName,
			name,
			producerAddress,
			rewardAddress,
			blockProducingPercentage,
			delegationPercentage,
		),
	}
}

func (pa *PillarApi) Revoke() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABIPillars.PackMethodPanic(definition.RevokeMethodName),
	}
}

func (pa *PillarApi) RegisterLegacy(name string, producerAddress, rewardAddress types.Address, blockProducingPercentage, delegationPercentage uint8, publicKey, signature string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        constants.PillarStakeAmount,
		Data: definition.ABIPillars.PackMethodPanic(
			definition.LegacyRegisterMethodName,
			name,
			producerAddress,
			rewardAddress,
			blockProducingPercentage,
			delegationPercentage,
			publicKey,
			signature,
		),
	}
}

// Delegate creates a transaction template to delegate your stake weight to a Pillar.
//
// Delegation allows ZNN holders to support Pillars and earn rewards without running
// infrastructure. Your ZNN remains in your wallet - only voting weight is delegated.
//
// Delegation benefits:
//   - Earn rewards from pillar's delegation percentage
//   - Support network decentralization
//   - No ZNN lockup required
//   - Can change delegation anytime
//
// Parameters:
//   - name: Name of the Pillar to delegate to
//
// Returns an unsigned AccountBlock template ready for processing.
//
// Example:
//
//	template := client.PillarApi.Delegate("MyFavoritePillar")
//	// Sign and publish transaction
//
// Note: Only one Pillar can be delegated to at a time. Delegating to a new Pillar
// automatically undelegates from the previous one.
func (pa *PillarApi) Delegate(name string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABIPillars.PackMethodPanic(definition.DelegateMethodName, name),
	}
}

func (pa *PillarApi) Undelegate() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABIPillars.PackMethodPanic(definition.UndelegateMethodName),
	}
}

func (pa *PillarApi) DepositQsr(amount *big.Int) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.QsrTokenStandard,
		Amount:        amount,
		Data:          definition.ABIPillars.PackMethodPanic(definition.DepositQsrMethodName),
	}
}

func (pa *PillarApi) WithdrawQsr() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABIPillars.PackMethodPanic(definition.WithdrawQsrMethodName),
	}
}

func (pa *PillarApi) CollectReward() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.PillarContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABIPillars.PackMethodPanic(definition.CollectRewardMethodName),
	}
}
