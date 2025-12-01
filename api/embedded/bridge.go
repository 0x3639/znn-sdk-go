package embedded

import (
	"math/big"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type BridgeApi struct {
	client *server.Client
}

func NewBridgeApi(client *server.Client) *BridgeApi {
	return &BridgeApi{
		client: client,
	}
}

func (ba *BridgeApi) GetBridgeInfo() (*BridgeInfo, error) {
	ans := new(BridgeInfo)
	if err := ba.client.Call(ans, "embedded.bridge.getBridgeInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetOrchestratorInfo() (*OrchestratorInfo, error) {
	ans := new(OrchestratorInfo)
	if err := ba.client.Call(ans, "embedded.bridge.getOrchestratorInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetTimeChallengesInfo() (*TimeChallengesList, error) {
	ans := new(TimeChallengesList)
	if err := ba.client.Call(ans, "embedded.bridge.getTimeChallengesInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetSecurityInfo() (*SecurityInfo, error) {
	ans := new(SecurityInfo)
	if err := ba.client.Call(ans, "embedded.bridge.getSecurityInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetNetworkInfo(networkClass, chainId uint32) (*BridgeNetworkInfo, error) {
	ans := new(BridgeNetworkInfo)
	if err := ba.client.Call(ans, "embedded.bridge.getNetworkInfo", networkClass, chainId); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetWrapTokenRequestById(id types.Hash) (*WrapTokenRequest, error) {
	ans := new(WrapTokenRequest)
	if err := ba.client.Call(ans, "embedded.bridge.getWrapTokenRequestById", id.String()); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetAllWrapTokenRequests(pageIndex, pageSize uint32) (*WrapTokenRequestList, error) {
	ans := new(WrapTokenRequestList)
	if err := ba.client.Call(ans, "embedded.bridge.getAllWrapTokenRequests", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetAllWrapTokenRequestsByToAddress(toAddress string, pageIndex, pageSize uint32) (*WrapTokenRequestList, error) {
	ans := new(WrapTokenRequestList)
	if err := ba.client.Call(ans, "embedded.bridge.getAllWrapTokenRequestsByToAddress", toAddress, pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetAllWrapTokenRequestsByToAddressNetworkClassAndChainId(toAddress string, networkClass, pageIndex, pageSize uint32) (*WrapTokenRequestList, error) {
	ans := new(WrapTokenRequestList)
	if err := ba.client.Call(ans, "embedded.bridge.getAllWrapTokenRequestsByToAddressNetworkClassAndChainId", toAddress, networkClass, pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetAllNetworks(pageIndex, pageSize uint32) (*BridgeNetworkInfoList, error) {
	ans := new(BridgeNetworkInfoList)
	if err := ba.client.Call(ans, "embedded.bridge.getAllNetworks", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetAllUnsignedWrapTokenRequests(pageIndex, pageSize uint32) (*WrapTokenRequestList, error) {
	ans := new(WrapTokenRequestList)
	if err := ba.client.Call(ans, "embedded.bridge.getAllUnsignedWrapTokenRequests", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetUnwrapTokenRequestByHashAndLog(txHash types.Hash, logIndex uint32) (*UnwrapTokenRequest, error) {
	ans := new(UnwrapTokenRequest)
	if err := ba.client.Call(ans, "embedded.bridge.getUnwrapTokenRequestByHashAndLog", txHash, logIndex); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetAllUnwrapTokenRequests(pageIndex, pageSize uint32) (*UnwrapTokenRequestList, error) {
	ans := new(UnwrapTokenRequestList)
	if err := ba.client.Call(ans, "embedded.bridge.getAllUnwrapTokenRequests", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) GetAllUnwrapTokenRequestsByToAddress(toAddress string, pageIndex, pageSize uint32) (*UnwrapTokenRequestList, error) {
	ans := new(UnwrapTokenRequestList)
	if err := ba.client.Call(ans, "embedded.bridge.getAllUnwrapTokenRequestsByToAddress", toAddress, pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

func (ba *BridgeApi) WrapToken(networkClass uint32, chainId uint32, toAddress string, amount *big.Int, tokenStandard types.ZenonTokenStandard) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: tokenStandard,
		Amount:        amount,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.WrapTokenMethodName,
			networkClass,
			chainId,
			toAddress,
		),
	}
}

func (ba *BridgeApi) UpdateWrapRequest(id types.Hash, signature string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.UpdateWrapRequestMethodName,
			id,
			signature,
		),
	}
}

func (ba *BridgeApi) UnwrapToken(networkClass uint32, chainId uint32, tokenAddress string, txHash types.Hash, logIndex uint32, amount *big.Int, toAddress types.Address, signature string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.UnwrapTokenMethodName,
			networkClass,
			chainId,
			txHash,
			logIndex,
			toAddress,
			tokenAddress,
			amount,
			signature,
		),
	}
}

func (ba *BridgeApi) Redeem(hash types.Hash, logIndex uint32) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.RedeemUnwrapMethodName,
			hash,
			logIndex,
		),
	}
}

func (ba *BridgeApi) Halt(signature string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.HaltMethodName,
			signature,
		),
	}
}

func (ba *BridgeApi) Emergency() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.EmergencyMethodName,
		),
	}
}

func (ba *BridgeApi) Unhalt() *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.UnhaltMethodName,
		),
	}
}

func (ba *BridgeApi) SetAllowKeygen(allowKeyGen bool) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.SetAllowKeygenMethodName,
			allowKeyGen,
		),
	}
}

func (ba *BridgeApi) ChangeTssECDSAPubKey(pubKey, signature, newSignature string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.ChangeTssECDSAPubKeyMethodName,
			pubKey,
			signature,
			newSignature,
		),
	}
}

func (ba *BridgeApi) ChangeAdministrator(administrator types.Address) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.ChangeAdministratorMethodName,
			administrator,
		),
	}
}

func (ba *BridgeApi) AddNetwork(networkClass uint32, chainId uint32, name, contractAddress, metadata string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.SetNetworkMethodName,
			networkClass,
			chainId,
			name,
			contractAddress,
			metadata,
		),
	}
}

func (ba *BridgeApi) RemoveNetwork(networkClass uint32, chainId uint32) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.RemoveNetworkMethodName,
			networkClass,
			chainId,
		),
	}
}

func (ba *BridgeApi) SetTokenPair(networkClass uint32, chainId uint32, tokenStandard types.ZenonTokenStandard, tokenAddress string, bridgeable, redeemable, owned bool, minAmount *big.Int, fee, redeemDelay uint32, metadata string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.SetTokenPairMethod,
			networkClass,
			chainId,
			tokenStandard,
			tokenAddress,
			bridgeable,
			redeemable,
			owned,
			minAmount,
			fee,
			redeemDelay,
			metadata,
		),
	}
}

func (ba *BridgeApi) RemoveTokenPair(networkClass uint32, chainId uint32, tokenStandard types.ZenonTokenStandard, tokenAddress string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.RemoveTokenPairMethodName,
			networkClass,
			chainId,
			tokenStandard,
			tokenAddress,
		),
	}
}

func (ba *BridgeApi) SetNetworkMetadata(networkClass uint32, chainId uint32, metadata string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.SetNetworkMetadataMethodName,
			networkClass,
			chainId,
			metadata,
		),
	}
}

func (ba *BridgeApi) SetOrchestratorInfo(windowSize uint64, keyGenThreshold, confirmationsToFinality, estimatedMomentumTime uint32) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.SetOrchestratorInfoMethodName,
			windowSize,
			keyGenThreshold,
			confirmationsToFinality,
			estimatedMomentumTime,
		),
	}
}

func (ba *BridgeApi) NominateGuardians(guardians []types.Address) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.NominateGuardiansMethodName,
			guardians,
		),
	}
}

func (ba *BridgeApi) SetBridgeMetadata(metadata string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.BridgeContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data: definition.ABIBridge.PackMethodPanic(
			definition.SetBridgeMetadataMethodName,
			metadata,
		),
	}
}
