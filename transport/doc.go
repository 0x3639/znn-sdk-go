// Package transport provides normalized JSON-RPC request, error, and
// subscription types shared by the Zenon SDK HTTP and WebSocket lifecycles.
//
// Requests always use JSON-RPC 2.0 positional parameters. RPC failures preserve
// their numeric code, message, optional data, method, and original parameter
// list in [RPCError]. WebSocket notifications normalize to [SubscriptionEvent]
// values containing both the opaque subscription ID and decoded updates.
//
// Most callers use these types through rpc_client.RpcClient. The standalone
// helpers are useful for adapters, diagnostics, and custom transports.
package transport
