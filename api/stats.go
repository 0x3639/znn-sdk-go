package api

import (
	"github.com/zenon-network/go-zenon/protocol"
	"github.com/zenon-network/go-zenon/rpc/api"
	"github.com/zenon-network/go-zenon/rpc/server"
)

type StatsApi struct {
	client *server.Client
}

func NewStatsApi(client *server.Client) *StatsApi {
	return &StatsApi{
		client: client,
	}
}

func (sa *StatsApi) OsInfo() (*api.OsInfoResponse, error) {
	ans := new(api.OsInfoResponse)
	if err := sa.client.Call(ans, "stats.osInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *StatsApi) ProcessInfo() (*api.ProcessInfoResponse, error) {
	ans := new(api.ProcessInfoResponse)
	if err := sa.client.Call(ans, "stats.processInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *StatsApi) NetworkInfo() (*api.NetworkInfoResponse, error) {
	ans := new(api.NetworkInfoResponse)
	if err := sa.client.Call(ans, "stats.networkInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}

func (sa *StatsApi) SyncInfo() (*protocol.SyncInfo, error) {
	ans := new(protocol.SyncInfo)
	if err := sa.client.Call(ans, "stats.syncInfo"); err != nil {
		return nil, err
	}
	return ans, nil
}
