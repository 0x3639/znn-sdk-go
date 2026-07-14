package api

import (
	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/zenon-network/go-zenon/protocol"
	"github.com/zenon-network/go-zenon/rpc/api"
)

type StatsApi struct {
	client transport.Caller
}

func NewStatsApi(client transport.Caller) *StatsApi {
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
