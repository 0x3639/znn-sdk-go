package modelwire

import (
	"bytes"
	"encoding/json"
	"fmt"

	sdkapi "github.com/0x3639/znn-sdk-go/api/embedded"
	"github.com/0x3639/znn-sdk-go/utils"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	nodeapi "github.com/zenon-network/go-zenon/rpc/api"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type factory func() interface{}

type constructorAddress struct {
	Core string `json:"core"`
	HRP  string `json:"hrp"`
}

type constructorCore struct {
	Core string `json:"core"`
}

type abstractModel struct{}

type swapAssetList struct {
	List map[string]*sdkapi.SwapAssetEntry `json:"list"`
}

type swapLegacyPillarList []*sdkapi.SwapLegacyPillarEntry

type syncInfoView struct {
	State         abstractModel `json:"state"`
	CurrentHeight uint64        `json:"currentHeight"`
	TargetHeight  uint64        `json:"targetHeight"`
}

var factories = map[string]factory{
	"AcceleratorProject":             func() interface{} { return new(sdkapi.Project) },
	"AccountBlock":                   func() interface{} { return new(nodeapi.AccountBlockMarshal) },
	"AccountBlockConfirmationDetail": func() interface{} { return new(nodeapi.AccountBlockConfirmationDetail) },
	"AccountBlockList":               func() interface{} { return new(nodeapi.AccountBlockList) },
	"AccountBlockTemplate":           func() interface{} { return new(nom.AccountBlockMarshal) },
	"AccountHeader":                  func() interface{} { return new(types.AccountHeader) },
	"AccountInfo":                    func() interface{} { return new(nodeapi.AccountInfo) },
	"Address":                        func() interface{} { return new(constructorAddress) },
	"BalanceInfoListItem":            func() interface{} { return new(nodeapi.BalanceInfo) },
	"BridgeInfo":                     func() interface{} { return new(sdkapi.BridgeInfo) },
	"BridgeNetworkInfo":              func() interface{} { return new(sdkapi.BridgeNetworkInfo) },
	"BridgeNetworkInfoList":          func() interface{} { return new(sdkapi.BridgeNetworkInfoList) },
	"DelegationInfo":                 func() interface{} { return new(sdkapi.DelegationInfo) },
	"DetailedMomentum":               func() interface{} { return new(nodeapi.DetailedMomentum) },
	"DetailedMomentumList":           func() interface{} { return new(nodeapi.DetailedMomentumList) },
	"FusionEntry":                    func() interface{} { return new(sdkapi.FusionEntry) },
	"FusionEntryList":                func() interface{} { return new(sdkapi.FusionEntryList) },
	"GetRequiredPowParam":            func() interface{} { return new(sdkapi.GetRequiredParam) },
	"GetRequiredPowResponse":         func() interface{} { return new(sdkapi.GetRequiredResult) },
	"Hash":                           func() interface{} { return new(constructorCore) },
	"HashHeight":                     func() interface{} { return new(utils.HashHeight) },
	"HtlcInfo":                       func() interface{} { return new(sdkapi.HtlcInfo) },
	"LiquidityInfo":                  func() interface{} { return new(sdkapi.LiquidityInfo) },
	"LiquidityStakeEntry":            func() interface{} { return new(sdkapi.LiquidityStakeEntry) },
	"LiquidityStakeList":             func() interface{} { return new(sdkapi.LiquidityStakeList) },
	"Model":                          func() interface{} { return new(abstractModel) },
	"Momentum":                       func() interface{} { return new(nodeapi.Momentum) },
	"MomentumList":                   func() interface{} { return new(nodeapi.MomentumList) },
	"NetworkInfo":                    func() interface{} { return new(nodeapi.NetworkInfoResponse) },
	"OrchestratorInfo":               func() interface{} { return new(sdkapi.OrchestratorInfo) },
	"OsInfo":                         func() interface{} { return new(nodeapi.OsInfoResponse) },
	"Peer":                           func() interface{} { return new(nodeapi.Peer) },
	"Phase":                          func() interface{} { return new(sdkapi.Phase) },
	"PillarEpochHistory":             func() interface{} { return new(sdkapi.PillarEpochHistory) },
	"PillarEpochHistoryList":         func() interface{} { return new(sdkapi.PillarEpochHistoryList) },
	"PillarEpochStats":               func() interface{} { return new(sdkapi.PillarEpochStats) },
	"PillarInfo":                     func() interface{} { return new(sdkapi.PillarInfo) },
	"PillarInfoList":                 func() interface{} { return new(sdkapi.PillarInfoList) },
	"PillarVote":                     func() interface{} { return new(definition.PillarVote) },
	"PlasmaInfo":                     func() interface{} { return new(sdkapi.PlasmaInfo) },
	"ProcessInfo":                    func() interface{} { return new(nodeapi.ProcessInfoResponse) },
	"Project":                        func() interface{} { return new(sdkapi.Project) },
	"ProjectList":                    func() interface{} { return new(sdkapi.ProjectList) },
	"RewardDeposit":                  func() interface{} { return new(definition.RewardDeposit) },
	"RewardHistoryEntry":             func() interface{} { return new(sdkapi.RewardHistoryEntry) },
	"RewardHistoryList":              func() interface{} { return new(sdkapi.RewardHistoryList) },
	"SecurityInfo":                   func() interface{} { return new(sdkapi.SecurityInfo) },
	"SentinelInfo":                   func() interface{} { return new(sdkapi.SentinelInfo) },
	"SentinelInfoList":               func() interface{} { return new(sdkapi.SentinelInfoList) },
	"Spork":                          func() interface{} { return new(sdkapi.Spork) },
	"SporkList":                      func() interface{} { return new(sdkapi.SporkList) },
	"StakeEntry":                     func() interface{} { return new(sdkapi.StakeEntry) },
	"StakeList":                      func() interface{} { return new(sdkapi.StakeList) },
	"SwapAssetEntry":                 func() interface{} { return new(sdkapi.SwapAssetEntry) },
	"SwapAssetList":                  func() interface{} { return new(swapAssetList) },
	"SwapLegacyPillarEntry":          func() interface{} { return new(sdkapi.SwapLegacyPillarEntry) },
	"SwapLegacyPillarList":           func() interface{} { return new(swapLegacyPillarList) },
	"SyncInfo":                       func() interface{} { return new(syncInfoView) },
	"TimeChallengeInfo":              func() interface{} { return new(sdkapi.TimeChallengeInfo) },
	"TimeChallengesList":             func() interface{} { return new(sdkapi.TimeChallengesList) },
	"Token":                          func() interface{} { return new(sdkapi.Token) },
	"TokenList":                      func() interface{} { return new(sdkapi.TokenList) },
	"TokenPair":                      func() interface{} { return new(sdkapi.TokenPair) },
	"TokenStandard":                  func() interface{} { return new(constructorCore) },
	"TokenTuple":                     func() interface{} { return new(sdkapi.TokenTuple) },
	"UncollectedReward":              func() interface{} { return new(sdkapi.UncollectedReward) },
	"UnwrapTokenRequest":             func() interface{} { return new(sdkapi.UnwrapTokenRequest) },
	"UnwrapTokenRequestList":         func() interface{} { return new(sdkapi.UnwrapTokenRequestList) },
	"VoteBreakdown":                  func() interface{} { return new(sdkapi.VoteBreakdown) },
	"WrapTokenRequest":               func() interface{} { return new(sdkapi.WrapTokenRequest) },
	"WrapTokenRequestList":           func() interface{} { return new(sdkapi.WrapTokenRequestList) },
	"ZtsFeesInfo":                    func() interface{} { return new(sdkapi.ZtsFeesInfo) },
}

// RoundTrip decodes and re-encodes one stable model fixture through its Go model.
//
// Parameters:
//   - model: Stable model name from spec/models.json.
//   - input: JSON value to decode through the model.
//
// RoundTrip returns a JSON-compatible value derived from the decoded Go model.
// It preserves the stable fixture's declared wire shape while normalizing
// big.Int values back to string form, as required by Zenon JSON-RPC. It returns
// an error for an unknown model, invalid input, a model decoder failure, or a
// missing declared wire field.
//
// Example:
//
//	actual, err := modelwire.RoundTrip("StakeList", json.RawMessage(`{"count":0,"list":[],"totalAmount":"0","totalWeightedAmount":"0"}`))
//
// This helper is intended for conformance adapters and tests. Application code
// should normally unmarshal RPC results directly through the API methods.
func RoundTrip(model string, input json.RawMessage) (interface{}, error) {
	newModel, ok := factories[model]
	if !ok {
		return nil, fmt.Errorf("unsupported model %q", model)
	}
	instance := newModel()
	if err := json.Unmarshal(input, instance); err != nil {
		return nil, fmt.Errorf("decode %s: %w", model, err)
	}
	encoded, err := json.Marshal(instance)
	if err != nil {
		return nil, fmt.Errorf("encode %s: %w", model, err)
	}
	template, err := decodeJSON(input)
	if err != nil {
		return nil, err
	}
	actual, err := decodeJSON(encoded)
	if err != nil {
		return nil, err
	}
	return conformShape(template, actual, model)
}

func decodeJSON(data []byte) (interface{}, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value interface{}
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func conformShape(template, actual interface{}, path string) (interface{}, error) {
	switch expected := template.(type) {
	case map[string]interface{}:
		return conformObject(expected, actual, path)
	case []interface{}:
		return conformArray(expected, actual, path)
	case string:
		return conformString(actual, path)
	case json.Number:
		return conformNumber(actual, path)
	case bool:
		return conformBoolean(actual, path)
	case nil:
		return conformNull(actual, path)
	default:
		return nil, fmt.Errorf("%s has unsupported fixture type %T", path, template)
	}
}

func conformObject(expected map[string]interface{}, actual interface{}, path string) (interface{}, error) {
	observed, ok := actual.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s encoded as %T, want object", path, actual)
	}
	result := make(map[string]interface{}, len(expected))
	for key, childTemplate := range expected {
		child, exists := observed[key]
		if !exists {
			return nil, fmt.Errorf("%s is missing wire field %q", path, key)
		}
		normalized, err := conformShape(childTemplate, child, path+"."+key)
		if err != nil {
			return nil, err
		}
		result[key] = normalized
	}
	return result, nil
}

func conformArray(expected []interface{}, actual interface{}, path string) (interface{}, error) {
	observed, ok := actual.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%s encoded as %T, want array", path, actual)
	}
	if len(expected) == 0 {
		return observed, nil
	}
	if len(observed) != len(expected) {
		return nil, fmt.Errorf("%s encoded %d items, want %d", path, len(observed), len(expected))
	}
	result := make([]interface{}, len(expected))
	for index := range expected {
		value, err := conformShape(expected[index], observed[index], fmt.Sprintf("%s[%d]", path, index))
		if err != nil {
			return nil, err
		}
		result[index] = value
	}
	return result, nil
}

func conformString(actual interface{}, path string) (interface{}, error) {
	switch value := actual.(type) {
	case string:
		return value, nil
	case json.Number:
		return value.String(), nil
	default:
		return nil, fmt.Errorf("%s encoded as %T, want string", path, actual)
	}
}

func conformNumber(actual interface{}, path string) (interface{}, error) {
	if value, ok := actual.(json.Number); ok {
		return value, nil
	}
	return nil, fmt.Errorf("%s encoded as %T, want number", path, actual)
}

func conformBoolean(actual interface{}, path string) (interface{}, error) {
	if value, ok := actual.(bool); ok {
		return value, nil
	}
	return nil, fmt.Errorf("%s encoded as %T, want boolean", path, actual)
}

func conformNull(actual interface{}, path string) (interface{}, error) {
	if actual == nil {
		return nil, nil
	}
	return nil, fmt.Errorf("%s encoded as %T, want null", path, actual)
}
