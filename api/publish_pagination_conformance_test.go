package api

import (
	"strings"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
)

type publishResultCaller struct {
	result interface{}
	calls  int
}

func (c *publishResultCaller) Call(result interface{}, _ string, _ ...interface{}) error {
	c.calls++
	*result.(*interface{}) = c.result
	return nil
}

func TestPublishRawTransactionRequiresNullResult(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		result  interface{}
		wantErr bool
	}{
		{name: "null", result: nil},
		{name: "boolean", result: true, wantErr: true},
		{name: "object", result: map[string]interface{}{}, wantErr: true},
		{name: "string", result: "ok", wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			caller := &publishResultCaller{result: test.result}
			err := NewLedgerApi(caller).PublishRawTransaction(new(nom.AccountBlock))
			if (err != nil) != test.wantErr {
				t.Fatalf("PublishRawTransaction() error = %v, wantErr %v", err, test.wantErr)
			}
			if test.wantErr && !strings.Contains(err.Error(), "non-null") {
				t.Fatalf("unexpected error: %v", err)
			}
			if caller.calls != 1 {
				t.Fatalf("transport calls = %d, want 1", caller.calls)
			}
		})
	}
}

func TestLedgerPaginationRejectsOversizedArgumentsBeforeCalling(t *testing.T) {
	t.Parallel()
	ledger := NewLedgerApi(nil)
	address := types.Address{}
	tests := []struct {
		name string
		call func() error
	}{
		{"getUnconfirmedBlocksByAddress", func() error { _, err := ledger.GetUnconfirmedBlocksByAddress(address, 0, 51); return err }},
		{"getAccountBlocksByHeight", func() error { _, err := ledger.GetAccountBlocksByHeight(address, 1, 1025); return err }},
		{"getAccountBlocksByPage", func() error { _, err := ledger.GetAccountBlocksByPage(address, 0, 1025); return err }},
		{"getUnreceivedBlocksByAddress", func() error { _, err := ledger.GetUnreceivedBlocksByAddress(address, 0, 51); return err }},
		{"getMomentumsByHeight", func() error { _, err := ledger.GetMomentumsByHeight(1, 1025); return err }},
		{"getMomentumsByPage", func() error { _, err := ledger.GetMomentumsByPage(0, 1025); return err }},
		{"getDetailedMomentumsByHeight", func() error { _, err := ledger.GetDetailedMomentumsByHeight(1, 1025); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err == nil || !strings.Contains(err.Error(), "exceeds maximum") {
				t.Fatalf("oversized request error = %v", err)
			}
		})
	}
}
