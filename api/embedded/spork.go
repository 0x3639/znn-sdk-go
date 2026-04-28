package embedded

import (
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/rpc/server"
	"github.com/zenon-network/go-zenon/vm/embedded/definition"
)

type SporkApi struct {
	client *server.Client
}

func NewSporkApi(client *server.Client) *SporkApi {
	return &SporkApi{
		client: client,
	}
}

// GetAll retrieves a paginated list of all sporks.
//
// Sporks are protocol activation mechanisms that enable/disable features across
// the network in a coordinated manner. They allow safe protocol upgrades without
// hard forks.
//
// Parameters:
//   - pageIndex: Page number (0-indexed)
//   - pageSize: Number of sporks per page
//
// Returns spork list or an error.
//
// Example:
//
//	sporks, err := client.SporkApi.GetAll(0, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, spork := range sporks.List {
//	    fmt.Printf("Spork: %s, Active: %t\n", spork.Name, spork.Activated)
//	}
func (sa *SporkApi) GetAll(pageIndex, pageSize uint32) (*SporkList, error) {
	ans := new(SporkList)
	if err := sa.client.Call(ans, "embedded.spork.getAll", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}

// CreateSpork creates a transaction template that proposes a new spork.
//
// Sporks are protocol-feature flags governed by the spork administrator. After
// creation, a spork must be activated via ActivateSpork before its enforcement
// height takes effect across the network.
//
// Parameters:
//   - name: Human-readable spork name
//   - description: Description of the protocol change the spork enables
//
// Returns an unsigned AccountBlock template ready for signing and publishing.
//
// Reference: znn_sdk_dart/lib/src/api/embedded/spork.dart:23-26
func (sa *SporkApi) CreateSpork(name, description string) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.SporkContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABISpork.PackMethodPanic(definition.SporkCreateMethodName, name, description),
	}
}

// ActivateSpork creates a transaction template that activates an existing spork.
//
// Parameters:
//   - id: Hash identifier of the spork to activate (returned by GetAll)
//
// Returns an unsigned AccountBlock template ready for signing and publishing.
//
// Reference: znn_sdk_dart/lib/src/api/embedded/spork.dart:28-31
func (sa *SporkApi) ActivateSpork(id types.Hash) *nom.AccountBlock {
	return &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     types.SporkContract,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        common.Big0,
		Data:          definition.ABISpork.PackMethodPanic(definition.SporkActivateMethodName, id),
	}
}
