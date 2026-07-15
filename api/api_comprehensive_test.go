package api

import (
	"errors"
	"math/big"
	"reflect"
	"testing"

	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
)

type recordedCall struct {
	method string
	args   []interface{}
}

type recordingCaller struct {
	calls []recordedCall
	err   error
}

func (c *recordingCaller) Call(_ interface{}, method string, args ...interface{}) error {
	c.calls = append(c.calls, recordedCall{method: method, args: append([]interface{}(nil), args...)})
	return c.err
}

func (c *recordingCaller) reset() {
	c.calls = nil
}

func assertLastCall(t *testing.T, caller *recordingCaller, method string, args ...interface{}) {
	t.Helper()
	if len(caller.calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(caller.calls))
	}
	got := caller.calls[0]
	if got.method != method {
		t.Fatalf("method = %q, want %q", got.method, method)
	}
	if !reflect.DeepEqual(got.args, args) {
		t.Fatalf("args = %#v, want %#v", got.args, args)
	}
	caller.reset()
}

func TestLedgerReadMethodsUseCanonicalWireCalls(t *testing.T) {
	caller := new(recordingCaller)
	ledger := NewLedgerApi(caller)
	address := types.ParseAddressPanic("z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f")
	hash := types.HexToHashPanic("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")

	if _, err := ledger.GetUnconfirmedBlocksByAddress(address, 2, 3); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getUnconfirmedBlocksByAddress", address.String(), uint32(2), uint32(3))

	if _, err := ledger.GetFrontierAccountBlock(address); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getFrontierAccountBlock", address.String())

	if _, err := ledger.GetAccountBlockByHash(hash); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getAccountBlockByHash", hash.String())

	if _, err := ledger.GetAccountBlocksByHeight(address, 4, 5); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getAccountBlocksByHeight", address.String(), uint64(4), uint64(5))

	if _, err := ledger.GetAccountBlocksByPage(address, 6, 7); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getAccountBlocksByPage", address.String(), uint32(6), uint32(7))

	if _, err := ledger.GetAccountInfoByAddress(address); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getAccountInfoByAddress", address.String())

	if _, err := ledger.GetUnreceivedBlocksByAddress(address, 8, 9); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getUnreceivedBlocksByAddress", address.String(), uint32(8), uint32(9))

	if _, err := ledger.GetFrontierMomentum(); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getFrontierMomentum")

	if _, err := ledger.GetMomentumBeforeTime(10); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getMomentumBeforeTime", int64(10))

	if _, err := ledger.GetMomentumByHash(hash); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getMomentumByHash", hash.String())

	if _, err := ledger.GetMomentumsByHeight(11, 12); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getMomentumsByHeight", uint64(11), uint64(12))

	if _, err := ledger.GetMomentumsByPage(13, 14); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getMomentumsByPage", uint32(13), uint32(14))

	if _, err := ledger.GetDetailedMomentumsByHeight(15, 16); err != nil {
		t.Fatal(err)
	}
	assertLastCall(t, caller, "ledger.getDetailedMomentumsByHeight", uint64(15), uint64(16))
}

func TestLedgerReadMethodsPropagateCallerErrors(t *testing.T) {
	want := errors.New("rpc unavailable")
	ledger := NewLedgerApi(&recordingCaller{err: want})
	address := types.ParseAddressPanic("z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f")
	hash := types.HexToHashPanic("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	tests := []func() error{
		func() error { _, err := ledger.GetUnconfirmedBlocksByAddress(address, 0, 1); return err },
		func() error { _, err := ledger.GetFrontierAccountBlock(address); return err },
		func() error { _, err := ledger.GetAccountBlockByHash(hash); return err },
		func() error { _, err := ledger.GetAccountBlocksByHeight(address, 1, 1); return err },
		func() error { _, err := ledger.GetAccountBlocksByPage(address, 0, 1); return err },
		func() error { _, err := ledger.GetAccountInfoByAddress(address); return err },
		func() error { _, err := ledger.GetUnreceivedBlocksByAddress(address, 0, 1); return err },
		func() error { _, err := ledger.GetFrontierMomentum(); return err },
		func() error { _, err := ledger.GetMomentumBeforeTime(1); return err },
		func() error { _, err := ledger.GetMomentumByHash(hash); return err },
		func() error { _, err := ledger.GetMomentumsByHeight(1, 1); return err },
		func() error { _, err := ledger.GetMomentumsByPage(0, 1); return err },
		func() error { _, err := ledger.GetDetailedMomentumsByHeight(1, 1); return err },
	}
	for index, call := range tests {
		if err := call(); !errors.Is(err, want) {
			t.Fatalf("call %d error = %v, want %v", index, err, want)
		}
	}
}

func TestLedgerPublishAndRetryTerminalPaths(t *testing.T) {
	block := &nom.AccountBlock{}

	success := new(recordingCaller)
	ledger := NewLedgerApi(success)
	if err := ledger.PublishRawTransaction(block); err != nil {
		t.Fatalf("PublishRawTransaction() error = %v", err)
	}
	assertLastCall(t, success, "ledger.publishRawTransaction", block)

	permanent := &recordingCaller{err: errors.New("invalid signature")}
	ledger = NewLedgerApi(permanent)
	err := ledger.PublishRawTransactionWithRetry(block, 3)
	if err == nil || len(permanent.calls) != 1 {
		t.Fatalf("permanent error = %v, attempts = %d", err, len(permanent.calls))
	}

	transient := &recordingCaller{err: errors.New("connection refused")}
	ledger = NewLedgerApi(transient)
	err = ledger.PublishRawTransactionWithRetry(block, 0)
	if err == nil || len(transient.calls) != 1 {
		t.Fatalf("exhausted error = %v, attempts = %d", err, len(transient.calls))
	}
}

func TestLedgerTemplatesContainCallerValues(t *testing.T) {
	ledger := NewLedgerApi(nil)
	address := types.ParseAddressPanic("z1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqsggv2f")
	hash := types.HexToHashPanic("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20")
	amount := big.NewInt(42)
	data := []byte("memo")

	send := ledger.SendTemplate(address, types.QsrTokenStandard, amount, data)
	if send.BlockType != nom.BlockTypeUserSend || send.ToAddress != address ||
		send.TokenStandard != types.QsrTokenStandard || send.Amount != amount ||
		!reflect.DeepEqual(send.Data, data) {
		t.Fatalf("send template = %+v", send)
	}

	receive := ledger.ReceiveTemplate(hash)
	if receive.BlockType != nom.BlockTypeUserReceive || receive.FromBlockHash != hash {
		t.Fatalf("receive template = %+v", receive)
	}
}

func TestStatsMethodsUseCanonicalWireCalls(t *testing.T) {
	caller := new(recordingCaller)
	stats := NewStatsApi(caller)

	for _, test := range []struct {
		method string
		call   func() error
	}{
		{"stats.osInfo", func() error { _, err := stats.OsInfo(); return err }},
		{"stats.processInfo", func() error { _, err := stats.ProcessInfo(); return err }},
		{"stats.networkInfo", func() error { _, err := stats.NetworkInfo(); return err }},
		{"stats.syncInfo", func() error { _, err := stats.SyncInfo(); return err }},
	} {
		t.Run(test.method, func(t *testing.T) {
			if err := test.call(); err != nil {
				t.Fatal(err)
			}
			assertLastCall(t, caller, test.method)
			wantErr := errors.New("stats failure")
			caller.err = wantErr
			if err := test.call(); !errors.Is(err, wantErr) {
				t.Fatalf("injected error = %v, want %v", err, wantErr)
			}
			caller.err = nil
			caller.reset()
		})
	}
}
