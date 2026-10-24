package rpc_client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// Default transport bounds applied by NewRpcClient.
const (
	// DefaultMaxHTTPResponseBytes bounds the body of every HTTP(S) JSON-RPC
	// response (successful or error) before it is decoded or buffered (CWE-400).
	DefaultMaxHTTPResponseBytes int64 = 16 << 20

	// DefaultMaxSubscriptionMessageBytes bounds each WebSocket frame read on
	// the dedicated subscription connection created by RpcClient.Subscribe.
	// This matches the limit go-zenon applies to its own request/response
	// WebSocket.
	DefaultMaxSubscriptionMessageBytes int64 = 15 << 20

	// DefaultHTTPTimeout is the end-to-end timeout for one HTTP JSON-RPC call.
	DefaultHTTPTimeout = 60 * time.Second
)

// ErrDestinationNotAllowed is returned (wrapped) by a DialPolicy that rejects
// a destination, and therefore by dials that hit such a policy.
var ErrDestinationNotAllowed = errors.New("rpc destination not allowed by dial policy")

// ErrResponseTooLarge is returned (wrapped) when an HTTP JSON-RPC response body
// exceeds ClientOptions.MaxHTTPResponseBytes.
var ErrResponseTooLarge = errors.New("rpc response exceeds maximum size")

// DialPolicy authorizes an outbound connection just before it is made.
//
// It is invoked with the network ("tcp4"/"tcp6") and the resolved
// "ip:port" address for every connection the client opens: the initial
// dial, every reconnect, every subscription socket, and any redirect the HTTP
// transport follows. Because it sees resolved addresses rather than the URL,
// DNS tricks cannot bypass it. Return nil to allow the connection or an error
// (conventionally wrapping ErrDestinationNotAllowed) to refuse it.
//
// The default is nil (allow everything), because the canonical Zenon node
// endpoint is ws://127.0.0.1:35998 and most SDK users intentionally talk to a
// local or private node. Applications that accept node URLs from untrusted
// users should set RejectNonPublicDestinations (or a stricter policy) to
// prevent server-side request forgery (CWE-918).
type DialPolicy func(network, address string) error

// RejectNonPublicDestinations is a DialPolicy that refuses loopback, private
// (RFC 1918 / ULA), link-local, multicast, unspecified, and CGNAT addresses.
//
// Example:
//
//	opts := rpc_client.DefaultClientOptions()
//	opts.DialPolicy = rpc_client.RejectNonPublicDestinations
//	client, err := rpc_client.NewRpcClientWithOptions(userSuppliedURL, opts)
func RejectNonPublicDestinations(network, address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%w: %q: %s", ErrDestinationNotAllowed, address, err.Error())
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("%w: %q is not an IP address", ErrDestinationNotAllowed, host)
	}
	if !isPublicIP(ip) {
		return fmt.Errorf("%w: %s is not a public address", ErrDestinationNotAllowed, ip)
	}
	return nil
}

var cgnat = net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() ||
		ip.IsInterfaceLocalMulticast() {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		if cgnat.Contains(v4) {
			return false
		}
		// 0.0.0.0/8 "this network" and 255.255.255.255 broadcast
		if v4[0] == 0 || v4.Equal(net.IPv4bcast) {
			return false
		}
	}
	return true
}

// policyGuard records the most recent DialPolicy rejection so callers can
// surface it even when an intermediate library (for example go-zenon's
// WebSocket handshake error) does not implement errors.Unwrap.
type policyGuard struct {
	policy DialPolicy
	mu     sync.Mutex
	last   error
}

func (g *policyGuard) control(network, address string, _ syscall.RawConn) error {
	err := g.policy(network, address)
	if err != nil {
		g.mu.Lock()
		g.last = err
		g.mu.Unlock()
	}
	return err
}

// lastError returns and clears the most recent rejection.
func (g *policyGuard) lastError() error {
	if g == nil {
		return nil
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	err := g.last
	g.last = nil
	return err
}

// newNetDialer returns a net.Dialer that consults policy for every resolved
// destination, plus the guard that records rejections. A nil policy allows
// everything and yields a nil guard.
func newNetDialer(policy DialPolicy) (*net.Dialer, *policyGuard) {
	d := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	if policy == nil {
		return d, nil
	}
	guard := &policyGuard{policy: policy}
	d.Control = guard.control
	return d, guard
}

// wrapDialError attaches the recorded policy rejection (if any) to err so
// errors.Is(err, ErrDestinationNotAllowed) works regardless of how the
// transport library wrapped the underlying dial failure.
func wrapDialError(err error, guard *policyGuard) error {
	if err == nil {
		return nil
	}
	if policyErr := guard.lastError(); policyErr != nil && !errors.Is(err, ErrDestinationNotAllowed) {
		return fmt.Errorf("%w: %s", policyErr, err.Error())
	}
	return err
}

// proxyFor returns the proxy resolver to use with the given policy. When a
// DialPolicy is active, proxies are disabled: with a proxy the TCP dial (and
// therefore the policy) would see the proxy's address while the proxy itself
// fetched or CONNECTed to the real, possibly internal, destination. Without a
// policy the environment proxy settings are honored as before.
func proxyFor(policy DialPolicy) func(*http.Request) (*url.URL, error) {
	if policy != nil {
		return nil
	}
	return http.ProxyFromEnvironment
}

// newWebsocketDialer returns a gorilla dialer whose TCP connections go through
// the dial policy.
func newWebsocketDialer(policy DialPolicy) (websocket.Dialer, *policyGuard) {
	netDialer, guard := newNetDialer(policy)
	return websocket.Dialer{
		Proxy:            proxyFor(policy),
		HandshakeTimeout: 45 * time.Second,
		NetDialContext:   netDialer.DialContext,
	}, guard
}

// newHTTPClient returns an http.Client that applies the dial policy, a finite
// timeout, and a response-body size limit to every JSON-RPC call.
func newHTTPClient(policy DialPolicy, maxResponseBytes int64, timeout time.Duration) *http.Client {
	netDialer, _ := newNetDialer(policy)
	transport := &http.Transport{
		Proxy:                 proxyFor(policy),
		DialContext:           netDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: &boundedBodyTransport{base: transport, limit: maxResponseBytes},
	}
}

// boundedBodyTransport wraps every response body in a reader that fails with
// ErrResponseTooLarge once more than limit bytes have been read. go-zenon's
// HTTP codec decodes successful bodies with json.Decoder and buffers non-2xx
// bodies in full; bounding the body here covers both paths.
type boundedBodyTransport struct {
	base  http.RoundTripper
	limit int64
}

func (t *boundedBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if t.limit > 0 {
		if resp.ContentLength > t.limit {
			// #nosec G104 -- already failing; nothing to do with a close error
			resp.Body.Close() //nolint:errcheck
			return nil, fmt.Errorf("%w: Content-Length %d > %d", ErrResponseTooLarge, resp.ContentLength, t.limit)
		}
		if !followsRedirect(req, resp) {
			// go-zenon's HTTP codec buffers non-2xx bodies in full, discards
			// the read error, and does not close the body on that path. An
			// oversized chunked error response would therefore surface as a
			// generic HTTP error while the connection lingered until the
			// timeout. Buffer the body here, bounded, so the size-limit
			// failure is preserved and the connection is released now.
			body, err := io.ReadAll(&limitedBody{ReadCloser: resp.Body, remaining: t.limit})
			// #nosec G104 -- body fully consumed or already failing
			resp.Body.Close() //nolint:errcheck
			if err != nil {
				return nil, fmt.Errorf("reading %s response body: %w", resp.Status, err)
			}
			resp.Body = io.NopCloser(bytes.NewReader(body))
			return resp, nil
		}
		resp.Body = &limitedBody{ReadCloser: resp.Body, remaining: t.limit}
	}
	return resp, nil
}

// followsRedirect reports whether net/http will follow resp for req instead
// of returning it to the caller, mirroring its redirectBehavior: 301/302/303
// are followed (as GET), 307/308 only when the request body can be replayed,
// and any other status never. A 2xx is reported as "followed" too, meaning
// the response is handed on without pre-buffering. net/http closes a
// followed redirect's body after draining at most a couple of KiB for
// connection reuse rather than reading it in full, so buffering it here
// would only delay the redirect, possibly past HTTPTimeout. Everything else is a terminal
// response that go-zenon turns into an HTTPError.
func followsRedirect(req *http.Request, resp *http.Response) bool {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	if resp.Header.Get("Location") == "" {
		return false
	}
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther:
		return true
	case http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		// net/http decides replayability from the ORIGINAL request of the
		// chain (reqs[0]), not the current hop: after a 302 turned a POST
		// into a bodiless GET, a 307 is still returned unfollowed because
		// the original POST cannot be re-sent. go-zenon's POSTs carry a
		// body and set no GetBody, so for them this is always false.
		first := originalRequest(req)
		return first.GetBody != nil || first.Body == nil || first.Body == http.NoBody
	}
	return false
}

// originalRequest walks a redirect chain back to its first request. net/http
// links each hop to the previous response via Request.Response.
func originalRequest(req *http.Request) *http.Request {
	for req.Response != nil && req.Response.Request != nil && req.Response.Request != req {
		req = req.Response.Request
	}
	return req
}

type limitedBody struct {
	io.ReadCloser
	remaining int64
}

func (l *limitedBody) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if l.remaining <= 0 {
		// Exactly limit bytes have been delivered. Probe one byte so a body
		// that is precisely at the limit still reports EOF rather than an
		// oversize error.
		var probe [1]byte
		n, err := l.ReadCloser.Read(probe[:])
		if n > 0 {
			return 0, l.tooLarge()
		}
		if err == nil {
			err = io.EOF
		}
		return 0, err
	}
	// Read one byte past the limit so an exact-limit body still succeeds
	// while anything longer is detected on the next call. Compare with
	// len(p)-1 rather than remaining+1: the latter overflows when the limit
	// is math.MaxInt64 and produced a negative slice bound.
	if int64(len(p))-1 > l.remaining {
		p = p[:l.remaining+1]
	}
	n, err := l.ReadCloser.Read(p)
	l.remaining -= int64(n)
	if l.remaining < 0 {
		return n, l.tooLarge()
	}
	return n, err
}

// tooLarge closes the underlying body so the connection is released
// immediately instead of lingering until the HTTP timeout, then returns
// ErrResponseTooLarge. Closing twice is safe for net/http bodies, so the
// caller's own deferred Close is unaffected.
func (l *limitedBody) tooLarge() error {
	// #nosec G104 -- already failing; nothing to do with a close error
	l.Close() //nolint:errcheck
	return ErrResponseTooLarge
}
