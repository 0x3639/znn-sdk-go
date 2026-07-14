package modelwire

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStableModelRegistryIsComplete(t *testing.T) {
	t.Parallel()
	if got, want := len(factories), 72; got != want {
		t.Fatalf("registered models = %d, want %d", got, want)
	}
}

func TestRoundTripUsesSDKModelValues(t *testing.T) {
	t.Parallel()
	input := json.RawMessage(`{"amount":"7","expirationTimestamp":3,"id":"0000000000000000000000000000000000000000000000000000000000000000","startTimestamp":2,"weightedAmount":"9","address":"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f"}`)
	actual, err := RoundTrip("StakeEntry", input)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	encoded, err := json.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}
	var got, want interface{}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(input, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip = %s, want %s", encoded, input)
	}
}

func TestRoundTripConstructorViewAndUnknownModel(t *testing.T) {
	t.Parallel()
	actual, err := RoundTrip("Address", json.RawMessage(`{"core":"fixture","hrp":"fixture"}`))
	if err != nil {
		t.Fatalf("constructor view: %v", err)
	}
	if got := actual.(map[string]interface{})["core"]; got != "fixture" {
		t.Fatalf("core = %v", got)
	}
	if _, err := RoundTrip("NotAModel", json.RawMessage(`{}`)); err == nil {
		t.Fatal("unknown model was accepted")
	}
}

func TestRoundTripPreservesEmptyByteString(t *testing.T) {
	t.Parallel()
	input := json.RawMessage(`{"amount":"1","expirationTime":1,"hashLock":"","hashLocked":"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f","hashType":1,"id":"0000000000000000000000000000000000000000000000000000000000000000","keyMaxSize":1,"timeLocked":"z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f","tokenStandard":"zts1qqqqqqqqqqqqqqqqtq587y"}`)
	actual, err := RoundTrip("HtlcInfo", input)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	if got := actual.(map[string]interface{})["hashLock"]; got != "" {
		t.Fatalf("hashLock = %#v, want empty string", got)
	}
}
