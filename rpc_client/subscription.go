package rpc_client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/gorilla/websocket"
)

// NormalizedSubscription is a reconnecting ledger subscription that exposes
// opaque subscription IDs together with decoded update batches.
//
// Use [RpcClient.Subscribe] to create a subscription. Events delivers
// [transport.SubscriptionEvent] values. When auto-reconnect is enabled on the
// parent client, an unexpected socket close reconnects and resubscribes with
// the original topic arguments.
type NormalizedSubscription struct {
	client *RpcClient
	topic  string
	args   []interface{}

	ctx    context.Context
	cancel context.CancelFunc
	events chan transport.SubscriptionEvent
	errors chan error
	done   chan struct{}

	mu             sync.RWMutex
	connection     *websocket.Conn
	subscriptionID string
	closeOnce      sync.Once
}

// Subscribe creates a normalized WebSocket ledger subscription.
//
// Parameters:
//   - ctx: Controls the complete subscription lifecycle, including reconnects.
//   - topic: Canonical ledger topic such as "momentums" or
//     "accountBlocksByAddress".
//   - arguments: Topic-specific positional arguments; address topics take one
//     Zenon address string.
//
// Subscribe returns after the node accepts the initial request and supplies an
// opaque subscription ID. It returns an error for HTTP transports, stopped
// clients, dialing failures, or rejected subscription requests.
//
// Example:
//
//	sub, err := client.Subscribe(ctx, "accountBlocksByAddress", address.String())
//	if err != nil {
//		return err
//	}
//	defer sub.Unsubscribe()
//	for event := range sub.Events() {
//		fmt.Println(event.SubscriptionID, event.Updates)
//	}
//
// Reconnection uses ClientOptions.ReconnectDelay and ReconnectAttempts. A zero
// attempt limit means unlimited retries, matching the stable SDK policy.
func (c *RpcClient) Subscribe(ctx context.Context, topic string, arguments ...interface{}) (*NormalizedSubscription, error) {
	if c == nil || c.IsClosed() {
		return nil, fmt.Errorf("RPC client is stopped")
	}
	parsed, err := url.Parse(c.url)
	if err != nil {
		return nil, fmt.Errorf("invalid RPC URL: %w", err)
	}
	if parsed.Scheme != "ws" && parsed.Scheme != "wss" {
		return nil, fmt.Errorf("subscriptions require ws or wss transport, got %s", parsed.Scheme)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	subscriptionCtx, cancel := context.WithCancel(ctx)
	subscription := &NormalizedSubscription{
		client: c,
		topic:  topic,
		args:   append([]interface{}(nil), arguments...),
		ctx:    subscriptionCtx,
		cancel: cancel,
		events: make(chan transport.SubscriptionEvent, 16),
		errors: make(chan error, 1),
		done:   make(chan struct{}),
	}
	connection, subscriptionID, err := subscription.open()
	if err != nil {
		cancel()
		return nil, err
	}
	subscription.setConnection(connection, subscriptionID)
	c.subscriptionLock.Lock()
	c.subscriptions[subscription] = struct{}{}
	c.subscriptionLock.Unlock()
	go subscription.run(connection)
	go subscription.watchContext()
	return subscription, nil
}

// Events returns the channel of normalized subscription updates.
//
// The channel closes after Unsubscribe, context cancellation, a terminal
// connection failure, or parent-client shutdown.
func (s *NormalizedSubscription) Events() <-chan transport.SubscriptionEvent {
	return s.events
}

// Err returns terminal subscription errors.
//
// Reconnectable socket failures are handled internally and are not emitted.
// The channel receives at most one terminal error and then closes.
func (s *NormalizedSubscription) Err() <-chan error {
	return s.errors
}

// ID returns the subscription ID assigned by the most recent successful
// subscribe or resubscribe response.
func (s *NormalizedSubscription) ID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.subscriptionID
}

// Unsubscribe closes the subscription and releases its socket and callbacks.
//
// Unsubscribe is idempotent and safe for concurrent use. It closes the private
// subscription connection rather than sending ledger.unsubscribe, which also
// removes the server-side subscription when the socket disconnects.
func (s *NormalizedSubscription) Unsubscribe() {
	if s == nil {
		return
	}
	s.closeOnce.Do(func() {
		s.cancel()
		s.mu.Lock()
		if s.connection != nil {
			_ = s.connection.Close()
			s.connection = nil
		}
		s.mu.Unlock()
	})
	<-s.done
}

type websocketResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  string          `json:"result"`
	Error   *struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	} `json:"error"`
}

type websocketNotification struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

func (s *NormalizedSubscription) open() (*websocket.Conn, string, error) {
	connection, _, err := websocket.DefaultDialer.DialContext(s.ctx, s.client.url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect subscription transport: %w", err)
	}
	request := transport.NewRequest(1, "ledger.subscribe", transport.SubscriptionParams(s.topic, s.args...)...)
	if err := connection.WriteJSON(request); err != nil {
		_ = connection.Close()
		return nil, "", fmt.Errorf("failed to send subscription request: %w", err)
	}
	var response websocketResponse
	if err := connection.ReadJSON(&response); err != nil {
		_ = connection.Close()
		return nil, "", fmt.Errorf("failed to read subscription response: %w", err)
	}
	if response.Error != nil {
		_ = connection.Close()
		message := response.Error.Message
		if message == "" {
			message = "Unknown error occurred"
		}
		return nil, "", &transport.RPCError{
			Code: response.Error.Code, Message: message, Data: response.Error.Data,
			Method: "ledger.subscribe", Parameters: transport.SubscriptionParams(s.topic, s.args...),
		}
	}
	if response.Result == "" {
		_ = connection.Close()
		return nil, "", fmt.Errorf("subscription response is missing an ID")
	}
	return connection, response.Result, nil
}

func (s *NormalizedSubscription) run(connection *websocket.Conn) {
	defer func() {
		s.mu.Lock()
		if s.connection != nil {
			_ = s.connection.Close()
			s.connection = nil
		}
		s.mu.Unlock()
		s.client.subscriptionLock.Lock()
		delete(s.client.subscriptions, s)
		s.client.subscriptionLock.Unlock()
		close(s.events)
		close(s.errors)
		close(s.done)
	}()

	current := connection
	for {
		var notification websocketNotification
		err := current.ReadJSON(&notification)
		if err == nil {
			event, normalizeErr := transport.NormalizeSubscriptionNotification(notification.Method, notification.Params)
			if normalizeErr != nil {
				s.finishWithError(normalizeErr)
				return
			}
			select {
			case s.events <- event:
			case <-s.ctx.Done():
				return
			}
			continue
		}
		if s.ctx.Err() != nil {
			return
		}
		if !s.client.autoReconnect {
			s.finishWithError(fmt.Errorf("subscription connection lost: %w", err))
			return
		}

		_ = current.Close()
		reconnected, ok := s.reconnect()
		if !ok {
			return
		}
		current = reconnected
	}
}

func (s *NormalizedSubscription) watchContext() {
	select {
	case <-s.ctx.Done():
		s.mu.Lock()
		if s.connection != nil {
			_ = s.connection.Close()
		}
		s.mu.Unlock()
	case <-s.done:
	}
}

func (s *NormalizedSubscription) reconnect() (*websocket.Conn, bool) {
	delay := s.client.reconnectDelay
	if delay <= 0 {
		delay = time.Second
	}
	maximumDelay := s.client.maxReconnectDelay
	if maximumDelay < delay {
		maximumDelay = delay
	}
	for attempt := 1; s.client.reconnectAttempts == 0 || attempt <= s.client.reconnectAttempts; attempt++ {
		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-s.ctx.Done():
			timer.Stop()
			return nil, false
		}
		connection, subscriptionID, err := s.open()
		if err == nil {
			s.setConnection(connection, subscriptionID)
			return connection, true
		}
		if s.client.reconnectAttempts > 0 && attempt == s.client.reconnectAttempts {
			s.finishWithError(fmt.Errorf("subscription reconnect failed after %d attempts: %w", attempt, err))
			return nil, false
		}
		delay *= 2
		if delay > maximumDelay {
			delay = maximumDelay
		}
	}
	return nil, false
}

func (s *NormalizedSubscription) setConnection(connection *websocket.Conn, subscriptionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connection = connection
	s.subscriptionID = subscriptionID
}

func (s *NormalizedSubscription) finishWithError(err error) {
	select {
	case s.errors <- err:
	default:
	}
}

func (c *RpcClient) closeNormalizedSubscriptions() {
	c.subscriptionLock.Lock()
	subscriptions := make([]*NormalizedSubscription, 0, len(c.subscriptions))
	for subscription := range c.subscriptions {
		subscriptions = append(subscriptions, subscription)
	}
	c.subscriptionLock.Unlock()
	for _, subscription := range subscriptions {
		subscription.Unsubscribe()
	}
}
