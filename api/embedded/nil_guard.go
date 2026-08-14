package embedded

import (
	"encoding/json"
	"fmt"
)

// requireNoNilEntries returns an error if any element of a pointer slice
// decoded from a node response is nil.
//
// JSON null elements (e.g. "list":[null]) decode to nil pointers without a
// decode error, which would then crash consumers that iterate the slice and
// dereference each entry. A hostile or compromised node can send such a
// payload, so list decoders reject it here (CWE-476). The field argument
// names the slice for the error message.
func requireNoNilEntries[T any](field string, list []*T) error {
	for i, entry := range list {
		if entry == nil {
			return fmt.Errorf("nil %s entry at index %d in node response", field, i)
		}
	}
	return nil
}

// The paginated list types below carry only a Count and a []*T List and rely
// on the default struct decoding, which silently accepts JSON null list
// elements. Each UnmarshalJSON decodes through an alias (to avoid recursion)
// and then rejects nil entries, closing the crash vector uniformly.

func (l *ProjectList) UnmarshalJSON(data []byte) error {
	type alias ProjectList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("project", aux.List); err != nil {
		return err
	}
	*l = ProjectList(aux)
	return nil
}

func (l *SentinelInfoList) UnmarshalJSON(data []byte) error {
	type alias SentinelInfoList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("sentinel", aux.List); err != nil {
		return err
	}
	*l = SentinelInfoList(aux)
	return nil
}

func (l *RewardHistoryList) UnmarshalJSON(data []byte) error {
	type alias RewardHistoryList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("reward history", aux.List); err != nil {
		return err
	}
	*l = RewardHistoryList(aux)
	return nil
}

func (l *TimeChallengesList) UnmarshalJSON(data []byte) error {
	type alias TimeChallengesList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("time challenge", aux.List); err != nil {
		return err
	}
	*l = TimeChallengesList(aux)
	return nil
}

func (l *TokenList) UnmarshalJSON(data []byte) error {
	type alias TokenList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("token", aux.List); err != nil {
		return err
	}
	*l = TokenList(aux)
	return nil
}

func (l *PillarInfoList) UnmarshalJSON(data []byte) error {
	type alias PillarInfoList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("pillar", aux.List); err != nil {
		return err
	}
	*l = PillarInfoList(aux)
	return nil
}

func (l *PillarEpochHistoryList) UnmarshalJSON(data []byte) error {
	type alias PillarEpochHistoryList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("pillar epoch history", aux.List); err != nil {
		return err
	}
	*l = PillarEpochHistoryList(aux)
	return nil
}

func (l *SporkList) UnmarshalJSON(data []byte) error {
	type alias SporkList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("spork", aux.List); err != nil {
		return err
	}
	*l = SporkList(aux)
	return nil
}

func (l *BridgeNetworkInfoList) UnmarshalJSON(data []byte) error {
	type alias BridgeNetworkInfoList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("bridge network", aux.List); err != nil {
		return err
	}
	*l = BridgeNetworkInfoList(aux)
	return nil
}

func (l *WrapTokenRequestList) UnmarshalJSON(data []byte) error {
	type alias WrapTokenRequestList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("wrap token request", aux.List); err != nil {
		return err
	}
	*l = WrapTokenRequestList(aux)
	return nil
}

func (l *UnwrapTokenRequestList) UnmarshalJSON(data []byte) error {
	type alias UnwrapTokenRequestList
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if err := requireNoNilEntries("unwrap token request", aux.List); err != nil {
		return err
	}
	*l = UnwrapTokenRequestList(aux)
	return nil
}
