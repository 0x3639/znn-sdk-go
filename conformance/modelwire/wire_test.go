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
	for name, factory := range factories {
		if instance := factory(); instance == nil {
			t.Errorf("factory %q returned nil", name)
		}
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

func TestConformShapeCoversEveryWireShape(t *testing.T) {
	t.Parallel()
	template, err := decodeJSON([]byte(`{"array":[{"string":"7","number":2,"boolean":true,"nothing":null}],"empty":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	actual, err := decodeJSON([]byte(`{"array":[{"string":7,"number":2,"boolean":true,"nothing":null}],"empty":[1,2]}`))
	if err != nil {
		t.Fatal(err)
	}
	got, err := conformShape(template, actual, "root")
	if err != nil {
		t.Fatalf("conformShape() error = %v", err)
	}
	object := got.(map[string]interface{})
	item := object["array"].([]interface{})[0].(map[string]interface{})
	if item["string"] != "7" || item["boolean"] != true || item["nothing"] != nil {
		t.Fatalf("normalized item = %#v", item)
	}
	if len(object["empty"].([]interface{})) != 2 {
		t.Fatalf("empty template did not preserve observed items: %#v", object["empty"])
	}
}

func TestConformShapeRejectsMismatches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		template interface{}
		actual   interface{}
	}{
		{"unsupported", struct{}{}, nil},
		{"object", map[string]interface{}{}, []interface{}{}},
		{"missing-field", map[string]interface{}{"field": ""}, map[string]interface{}{}},
		{"array", []interface{}{}, map[string]interface{}{}},
		{"array-length", []interface{}{json.Number("1")}, []interface{}{}},
		{"string", "", true},
		{"number", json.Number("1"), "1"},
		{"boolean", true, "true"},
		{"null", nil, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := conformShape(test.template, test.actual, "root"); err == nil {
				t.Fatal("mismatched shape was accepted")
			}
		})
	}
}

func TestDecodeJSONAndRoundTripRejectInvalidJSON(t *testing.T) {
	t.Parallel()
	if _, err := decodeJSON([]byte(`{`)); err == nil {
		t.Fatal("decodeJSON accepted invalid JSON")
	}
	if _, err := RoundTrip("StakeEntry", json.RawMessage(`{`)); err == nil {
		t.Fatal("RoundTrip accepted invalid JSON")
	}
}
