package embedded

import (
	"github.com/zenon-network/go-zenon/rpc/api/embedded"
	"github.com/zenon-network/go-zenon/rpc/server"
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
func (sa *SporkApi) GetAll(pageIndex, pageSize uint32) (*embedded.SporkList, error) {
	ans := new(embedded.SporkList)
	if err := sa.client.Call(ans, "embedded.spork.getAll", pageIndex, pageSize); err != nil {
		return nil, err
	}
	return ans, nil
}
