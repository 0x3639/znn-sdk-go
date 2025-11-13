package rpc_client

// WebsocketStatus represents the connection status of the WebSocket client
type WebsocketStatus int

const (
	// Uninitialized indicates the client has not been initialized
	Uninitialized WebsocketStatus = iota
	// Connecting indicates the client is attempting to connect
	Connecting
	// Running indicates the client is connected and operational
	Running
	// Stopped indicates the client has been stopped
	Stopped
)

// String returns the string representation of WebsocketStatus
func (s WebsocketStatus) String() string {
	switch s {
	case Uninitialized:
		return "Uninitialized"
	case Connecting:
		return "Connecting"
	case Running:
		return "Running"
	case Stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}
