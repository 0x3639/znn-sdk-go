package modelwire

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// CodeRabbit finding: dependency decoders that dereference nil pointers must
// surface as errors, not process-terminating panics.
func TestRoundTrip_ContainsDecoderPanics(t *testing.T) {
	cases := map[string]string{
		"BalanceInfoListItem": `{}`,
		"AccountBlockList":    `{"list":[null],"count":1,"more":false}`,
	}
	for model, input := range cases {
		_, err := RoundTrip(model, json.RawMessage(input))
		if err == nil {
			t.Errorf("%s: expected an error", model)
		}
	}
}

func TestRoundTrip_RejectsOversizedInput(t *testing.T) {
	input := `{"count":1,"list":[` + strings.Repeat(`{"votes":{}},`, MaxInputSize/12) + `{"votes":{}}]}`
	_, err := RoundTrip("ProjectList", json.RawMessage(input))
	if !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("expected ErrInputTooLarge, got %v", err)
	}
	// Codex finding: a byte-limit-compliant input with excessive cardinality
	// must also be rejected before model expansion.
	small := `{"count":1,"list":[` + strings.Repeat(`{"votes":{}},`, MaxInputNodes) + `{"votes":{}}]}`
	if len(small) > MaxInputSize {
		t.Fatalf("test fixture exceeds MaxInputSize: %d", len(small))
	}
	_, err = RoundTrip("ProjectList", json.RawMessage(small))
	if !errors.Is(err, ErrInputTooLarge) {
		t.Fatalf("expected ErrInputTooLarge for %d nodes, got %v", MaxInputNodes*2, err)
	}
}
