package embedded

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/constants"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

// =============================================================================
// AcceleratorApi Tests
// =============================================================================

func TestNewAcceleratorApi(t *testing.T) {
	api := NewAcceleratorApi(nil)
	if api == nil {
		t.Fatal("NewAcceleratorApi returned nil")
	}
}

func TestAcceleratorApi_CreateProject(t *testing.T) {
	api := NewAcceleratorApi(nil)
	znnNeeded := big.NewInt(5000 * 100000000)
	qsrNeeded := big.NewInt(50000 * 100000000)

	block := api.CreateProject("MyProject", "Description", "https://example.com", znnNeeded, qsrNeeded)
	if block == nil {
		t.Fatal("CreateProject returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.AcceleratorContract {
		t.Errorf("ToAddress = %s, want AcceleratorContract", block.ToAddress)
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard)
	}
	if block.Amount.Cmp(constants.ProjectCreationAmount) != 0 {
		t.Errorf("Amount = %v, want %v", block.Amount, constants.ProjectCreationAmount)
	}
	expected := definition.ABIAccelerator.PackMethodPanic(
		definition.CreateProjectMethodName, "MyProject", "Description", "https://example.com", znnNeeded, qsrNeeded,
	)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for CreateProject")
	}
}

func TestAcceleratorApi_AddPhase(t *testing.T) {
	api := NewAcceleratorApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	znnNeeded := big.NewInt(1000 * 100000000)
	qsrNeeded := big.NewInt(10000 * 100000000)

	block := api.AddPhase(id, "Phase1", "Phase desc", "https://phase.com", znnNeeded, qsrNeeded)
	if block == nil {
		t.Fatal("AddPhase returned nil")
	}
	if block.ToAddress != types.AcceleratorContract {
		t.Errorf("ToAddress = %s, want AcceleratorContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABIAccelerator.PackMethodPanic(
		definition.AddPhaseMethodName, id, "Phase1", "Phase desc", "https://phase.com", znnNeeded, qsrNeeded,
	)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for AddPhase")
	}
}

func TestAcceleratorApi_UpdatePhase(t *testing.T) {
	api := NewAcceleratorApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	znnNeeded := big.NewInt(500 * 100000000)
	qsrNeeded := big.NewInt(5000 * 100000000)

	block := api.UpdatePhase(id, "UpdatedPhase", "Updated desc", "https://updated.com", znnNeeded, qsrNeeded)
	if block == nil {
		t.Fatal("UpdatePhase returned nil")
	}
	if block.ToAddress != types.AcceleratorContract {
		t.Errorf("ToAddress = %s, want AcceleratorContract", block.ToAddress)
	}
	expected := definition.ABIAccelerator.PackMethodPanic(
		definition.UpdatePhaseMethodName, id, "UpdatedPhase", "Updated desc", "https://updated.com", znnNeeded, qsrNeeded,
	)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for UpdatePhase")
	}
}

func TestAcceleratorApi_Donate(t *testing.T) {
	api := NewAcceleratorApi(nil)
	amount := big.NewInt(100 * 100000000)

	block := api.Donate(amount, types.ZnnTokenStandard)
	if block == nil {
		t.Fatal("Donate returned nil")
	}
	if block.ToAddress != types.AcceleratorContract {
		t.Errorf("ToAddress = %s, want AcceleratorContract", block.ToAddress)
	}
	if block.Amount.Cmp(amount) != 0 {
		t.Errorf("Amount = %v, want %v", block.Amount, amount)
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard)
	}
}

func TestAcceleratorApi_VoteByName(t *testing.T) {
	api := NewAcceleratorApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	block := api.VoteByName(id, "MyPillar", 1)
	if block == nil {
		t.Fatal("VoteByName returned nil")
	}
	if block.ToAddress != types.AcceleratorContract {
		t.Errorf("ToAddress = %s, want AcceleratorContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABIAccelerator.PackMethodPanic(definition.VoteByNameMethodName, id, "MyPillar", uint8(1))
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for VoteByName")
	}
}

func TestAcceleratorApi_VoteByProducerAddress(t *testing.T) {
	api := NewAcceleratorApi(nil)
	id := types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

	block := api.VoteByProducerAddress(id, 2)
	if block == nil {
		t.Fatal("VoteByProducerAddress returned nil")
	}
	if block.ToAddress != types.AcceleratorContract {
		t.Errorf("ToAddress = %s, want AcceleratorContract", block.ToAddress)
	}
	expected := definition.ABIAccelerator.PackMethodPanic(definition.VoteByProdAddressMethodName, id, uint8(2))
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for VoteByProducerAddress")
	}
}

// =============================================================================
// SentinelApi Tests
// =============================================================================

func TestNewSentinelApi(t *testing.T) {
	api := NewSentinelApi(nil)
	if api == nil {
		t.Fatal("NewSentinelApi returned nil")
	}
}

func TestSentinelApi_Register(t *testing.T) {
	api := NewSentinelApi(nil)

	block := api.Register()
	if block == nil {
		t.Fatal("Register returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.SentinelContract {
		t.Errorf("ToAddress = %s, want SentinelContract", block.ToAddress)
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard)
	}
	if block.Amount.Cmp(constants.SentinelZnnRegisterAmount) != 0 {
		t.Errorf("Amount = %v, want %v", block.Amount, constants.SentinelZnnRegisterAmount)
	}
	expected := definition.ABISentinel.PackMethodPanic(definition.RegisterSentinelMethodName)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for Register")
	}
}

func TestSentinelApi_Revoke(t *testing.T) {
	api := NewSentinelApi(nil)

	block := api.Revoke()
	if block == nil {
		t.Fatal("Revoke returned nil")
	}
	if block.ToAddress != types.SentinelContract {
		t.Errorf("ToAddress = %s, want SentinelContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABISentinel.PackMethodPanic(definition.RevokeSentinelMethodName)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for Revoke")
	}
}

func TestSentinelApi_DepositQsr(t *testing.T) {
	api := NewSentinelApi(nil)
	amount := big.NewInt(50000 * 100000000)

	block := api.DepositQsr(amount)
	if block == nil {
		t.Fatal("DepositQsr returned nil")
	}
	if block.ToAddress != types.SentinelContract {
		t.Errorf("ToAddress = %s, want SentinelContract", block.ToAddress)
	}
	if block.TokenStandard != types.QsrTokenStandard {
		t.Errorf("TokenStandard = %s, want QSR", block.TokenStandard)
	}
	if block.Amount.Cmp(amount) != 0 {
		t.Errorf("Amount = %v, want %v", block.Amount, amount)
	}
}

func TestSentinelApi_WithdrawQsr(t *testing.T) {
	api := NewSentinelApi(nil)

	block := api.WithdrawQsr()
	if block == nil {
		t.Fatal("WithdrawQsr returned nil")
	}
	if block.ToAddress != types.SentinelContract {
		t.Errorf("ToAddress = %s, want SentinelContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
}

func TestSentinelApi_CollectReward(t *testing.T) {
	api := NewSentinelApi(nil)

	block := api.CollectReward()
	if block == nil {
		t.Fatal("CollectReward returned nil")
	}
	if block.ToAddress != types.SentinelContract {
		t.Errorf("ToAddress = %s, want SentinelContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABISentinel.PackMethodPanic(definition.CollectRewardMethodName)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for CollectReward")
	}
}

// =============================================================================
// StakeApi Tests
// =============================================================================

func TestNewStakeApi(t *testing.T) {
	api := NewStakeApi(nil)
	if api == nil {
		t.Fatal("NewStakeApi returned nil")
	}
}

func TestStakeApi_Stake(t *testing.T) {
	api := NewStakeApi(nil)
	amount := big.NewInt(100 * 100000000)
	const duration = int64(2592000) // 1 month

	block := api.Stake(duration, amount)
	if block == nil {
		t.Fatal("Stake returned nil")
	}
	if block.BlockType != nom.BlockTypeUserSend {
		t.Errorf("BlockType = %d, want %d", block.BlockType, nom.BlockTypeUserSend)
	}
	if block.ToAddress != types.StakeContract {
		t.Errorf("ToAddress = %s, want StakeContract", block.ToAddress)
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN", block.TokenStandard)
	}
	if block.Amount.Cmp(amount) != 0 {
		t.Errorf("Amount = %v, want %v", block.Amount, amount)
	}
	expected := definition.ABIStake.PackMethodPanic(definition.StakeMethodName, duration)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for Stake")
	}
}

func TestStakeApi_Cancel(t *testing.T) {
	api := NewStakeApi(nil)
	id := types.HexToHashPanic("aaaa000000000000000000000000000000000000000000000000000000000001")

	block := api.Cancel(id)
	if block == nil {
		t.Fatal("Cancel returned nil")
	}
	if block.ToAddress != types.StakeContract {
		t.Errorf("ToAddress = %s, want StakeContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABIStake.PackMethodPanic(definition.CancelStakeMethodName, id)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for Cancel")
	}
}

func TestStakeApi_CollectReward(t *testing.T) {
	api := NewStakeApi(nil)

	block := api.CollectReward()
	if block == nil {
		t.Fatal("CollectReward returned nil")
	}
	if block.ToAddress != types.StakeContract {
		t.Errorf("ToAddress = %s, want StakeContract", block.ToAddress)
	}
	if block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want 0", block.Amount)
	}
	expected := definition.ABIStake.PackMethodPanic(definition.CollectRewardMethodName)
	if !bytes.Equal(block.Data, expected) {
		t.Errorf("Data mismatch for CollectReward")
	}
}
