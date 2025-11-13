package rpc_client

import (
	"github.com/MoonBaZZe/znn-sdk-go/api"
	"github.com/MoonBaZZe/znn-sdk-go/api/embedded"

	"github.com/zenon-network/go-zenon/rpc/server"
)

type RpcClient struct {
	client *server.Client

	// Embedded
	AcceleratorApi *embedded.AcceleratorApi
	PillarApi      *embedded.PillarApi
	PlasmaApi      *embedded.PlasmaApi
	SentinelApi    *embedded.SentinelApi
	SporkApi       *embedded.SporkApi
	StakeApi       *embedded.StakeApi
	SwapApi        *embedded.SwapApi
	TokenApi       *embedded.TokenApi
	BridgeApi      *embedded.BridgeApi
	LiquidityApi   *embedded.LiquidityApi
	HtlcApi        *embedded.HtlcApi

	LedgerApi     *api.LedgerApi
	StatsApi      *api.StatsApi
	SubscriberApi *api.SubscriberApi
}

func NewRpcClient(url string) (*RpcClient, error) {
	newClient, err := server.Dial(url)
	if err != nil {
		return nil, err
	}
	return &RpcClient{
		client:         newClient,
		AcceleratorApi: embedded.NewAcceleratorApi(newClient),
		BridgeApi:      embedded.NewBridgeApi(newClient),
		PillarApi:      embedded.NewPillarApi(newClient),
		PlasmaApi:      embedded.NewPlasmaApi(newClient),
		SentinelApi:    embedded.NewSentinelApi(newClient),
		SporkApi:       embedded.NewSporkApi(newClient),
		StakeApi:       embedded.NewStakeApi(newClient),
		SwapApi:        embedded.NewSwapApi(newClient),
		TokenApi:       embedded.NewTokenApi(newClient),
		LiquidityApi:   embedded.NewLiquidityApi(newClient),
		HtlcApi:        embedded.NewHtlcApi(newClient),
		LedgerApi:      api.NewLedgerApi(newClient),
		StatsApi:       api.NewStatsApi(newClient),
		SubscriberApi:  api.NewSubscriberApi(newClient),
	}, nil
}

func (c *RpcClient) Stop() {
	if c.client != nil {
		c.client.Close()
	}
}
