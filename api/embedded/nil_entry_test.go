package embedded

import (
	"encoding/json"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

// Regression tests for the nil-entry unmarshal class (findings #2, #6, #7,
// #8, #9, #11, #12, #13, #15, #16): a hostile or compromised node can send
// "list":[null] (or a null map value), which encoding/json accepts while
// leaving a nil pointer that crashes consumers following the documented
// iteration patterns. Decoding must reject these payloads with an error.

func TestListTypesRejectNullEntries(t *testing.T) {
	const nullListPayload = `{"count":1,"list":[null]}`

	cases := []struct {
		name   string
		decode func([]byte) error
	}{
		{"ProjectList", func(b []byte) error { return json.Unmarshal(b, &ProjectList{}) }},
		{"SentinelInfoList", func(b []byte) error { return json.Unmarshal(b, &SentinelInfoList{}) }},
		{"RewardHistoryList", func(b []byte) error { return json.Unmarshal(b, &RewardHistoryList{}) }},
		{"TokenList", func(b []byte) error { return json.Unmarshal(b, &TokenList{}) }},
		{"PillarInfoList", func(b []byte) error { return json.Unmarshal(b, &PillarInfoList{}) }},
		{"PillarEpochHistoryList", func(b []byte) error { return json.Unmarshal(b, &PillarEpochHistoryList{}) }},
		{"StakeList", func(b []byte) error { return json.Unmarshal(b, &StakeList{}) }},
		{"FusionEntryList", func(b []byte) error { return json.Unmarshal(b, &FusionEntryList{}) }},
		{"SporkList", func(b []byte) error { return json.Unmarshal(b, &SporkList{}) }},
		{"LiquidityStakeList", func(b []byte) error { return json.Unmarshal(b, &LiquidityStakeList{}) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.decode([]byte(nullListPayload)); err == nil {
				t.Errorf("%s accepted a null list entry", tc.name)
			}
		})
	}
}

func TestProjectRejectsNilNestedFields(t *testing.T) {
	base := `"id":"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",` +
		`"owner":"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",` +
		`"znnFundsNeeded":"0","qsrFundsNeeded":"0"`

	cases := map[string]string{
		"nil votes":       `{` + base + `,"votes":null,"phases":[]}`,
		"null phase elem": `{` + base + `,"votes":{"total":0,"yes":0,"no":0},"phases":[null]}`,
		"nil phase.phase": `{` + base + `,"votes":{"total":0,"yes":0,"no":0},"phases":[{"phase":null,"votes":{"total":0,"yes":0,"no":0}}]}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var p Project
			if err := json.Unmarshal([]byte(payload), &p); err == nil {
				t.Errorf("Project accepted %s", name)
			}
		})
	}
}

func TestPillarInfoRejectsNilCurrentStats(t *testing.T) {
	payload := `{"name":"p","ownerAddress":"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",` +
		`"producerAddress":"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",` +
		`"withdrawAddress":"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f",` +
		`"weight":"0","currentStats":null}`
	var p PillarInfo
	if err := json.Unmarshal([]byte(payload), &p); err == nil {
		t.Error("PillarInfo accepted a nil currentStats")
	}
}

// decodingCaller unmarshals a fixed payload into the result, mirroring the
// contract that transport.Caller implementations honor.
type decodingCaller struct {
	payload string
}

func (c *decodingCaller) Call(result interface{}, _ string, _ ...interface{}) error {
	return json.Unmarshal([]byte(c.payload), result)
}

// Finding #10: GetAssets passed a non-pointer map to Call, so it could never
// decode. Finding #16: null map values must be rejected.
func TestSwapGetAssetsDecodesAndRejectsNullValues(t *testing.T) {
	hash := "6900000000000000000000000000000000000000000000000000000000000001"

	ok := &decodingCaller{payload: `{"` + hash + `":{"qsr":"5","znn":"7"}}`}
	assets, err := NewSwapApi(ok).GetAssets()
	if err != nil {
		t.Fatalf("GetAssets(valid) error = %v", err)
	}
	entry, present := assets[types.HexToHashPanic(hash)]
	if !present || entry == nil || entry.Qsr.Int64() != 5 || entry.Znn.Int64() != 7 {
		t.Fatalf("GetAssets did not decode the asset map: %+v", assets)
	}

	hostile := &decodingCaller{payload: `{"` + hash + `":null}`}
	if _, err := NewSwapApi(hostile).GetAssets(); err == nil {
		t.Error("GetAssets accepted a null map value")
	}
}

// Finding #15: null legacy pillar slice entries must be rejected.
func TestSwapGetLegacyPillarsRejectsNullEntries(t *testing.T) {
	hostile := &decodingCaller{payload: `[null,{"numPillars":1,"keyIdHash":"6900000000000000000000000000000000000000000000000000000000000001"}]`}
	if _, err := NewSwapApi(hostile).GetLegacyPillars(); err == nil {
		t.Error("GetLegacyPillars accepted a null slice entry")
	}
}
