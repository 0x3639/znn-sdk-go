package api

import (
	"context"
	"testing"

	"github.com/zenon-network/go-zenon/common/types"
)

// Regression test for the post-Stop panic: RpcClient.Stop() calls
// SubscriberApi.SetClient(nil) so the stable API object "fails cleanly", but
// the To* methods dereferenced the nil client inside server.Client.Subscribe
// and panicked. A subscribe that races Stop() must return an error, never
// panic — callers (e.g. a wallet's auto-receive racing a node disconnect)
// snapshot the client before subscribing and cannot exclude that interleaving.
func TestSubscriberMethodsReturnErrorAfterStop(t *testing.T) {
	sa := NewSubscriberApi(nil) // the state Stop()/SetClient(nil) leaves behind
	ctx := context.Background()
	addr := types.ZeroAddress

	cases := []struct {
		name string
		call func() error
	}{
		{"ToMomentums", func() error {
			_, _, err := sa.ToMomentums(ctx)
			return err
		}},
		{"ToAllAccountBlocks", func() error {
			_, _, err := sa.ToAllAccountBlocks(ctx)
			return err
		}},
		{"ToAccountBlocksByAddress", func() error {
			_, _, err := sa.ToAccountBlocksByAddress(ctx, addr)
			return err
		}},
		{"ToUnreceivedAccountBlocksByAddress", func() error {
			_, _, err := sa.ToUnreceivedAccountBlocksByAddress(ctx, addr)
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("%s panicked with nil client: %v", tc.name, r)
				}
			}()
			if err := tc.call(); err == nil {
				t.Fatalf("%s returned nil error with nil client", tc.name)
			}
		})
	}
}
