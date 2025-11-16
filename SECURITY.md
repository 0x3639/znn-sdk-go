# Security

This document describes the security features and considerations for the Zenon Go SDK.

## TLS/Certificate Validation

### WebSocket Connection Security

The SDK supports both encrypted (`wss://`) and unencrypted (`ws://`) WebSocket connections to Zenon nodes.

**TLS Certificate Validation: ENABLED BY DEFAULT**

When connecting to nodes via `wss://` (WebSocket Secure), the SDK performs full TLS certificate validation:

- ✅ Certificate chain verification against system CA certificates
- ✅ Hostname verification (SNI) to prevent man-in-the-middle attacks
- ✅ Certificate expiry validation
- ✅ Uses Go's standard `crypto/tls` package with industry-standard security

**Example:**
```go
// Secure connection with full certificate validation
client, err := rpc_client.NewRpcClient("wss://my.hc1node.com:35998")
```

### Recommended Practices

**Production Environments:**
- ✅ **ALWAYS use `wss://`** for connections over public networks
- ✅ Ensure node certificates are issued by trusted CAs
- ✅ Verify certificate expiration dates

**Development/Local Environments:**
- For local testing, `ws://127.0.0.1:35998` (unencrypted) is acceptable
- For remote testing with self-signed certificates, see Limitations below

### Current Limitations

**No TLS Configuration Options (v0.x)**

The SDK does not currently expose TLS configuration options. This means:

- ❌ Cannot connect to nodes with self-signed certificates
- ❌ Cannot use custom CA certificate bundles
- ❌ Cannot configure mutual TLS (client certificates)
- ❌ Cannot disable certificate validation (even for testing)

**Why This Limitation Exists:**

The underlying `github.com/zenon-network/go-zenon/rpc/server` dependency does not expose TLS configuration through its `Dial()` function. This is intentional for security-by-default design.

**Workarounds:**

1. **Use Valid Certificates:** Obtain properly signed certificates from Let's Encrypt or other CAs
2. **Local Connections:** Use `ws://` for localhost testing (unencrypted but acceptable for local)
3. **Trusted Networks:** Deploy nodes with valid certificates even in private networks

**Future Enhancement:**

We are tracking a feature request to add TLS configuration support for advanced use cases:
- Custom CA certificates for private deployments
- Client certificate support for mutual TLS
- Configurable TLS versions and cipher suites

If you need these features, please open an issue describing your use case.

## Private Key Security

### Memory Protection

**KeyPair.Destroy() Method**

The SDK provides a `Destroy()` method to securely zero sensitive key material from memory:

```go
kp, err := wallet.NewKeyPairFromSeed(seed)
if err != nil {
    return err
}
defer kp.Destroy()  // Ensures cleanup even on panic

// Use keypair for signing...
signature, err := kp.Sign(message)
```

**Security Properties:**
- ✅ Zeros private key bytes before releasing memory
- ✅ Zeros public key bytes (defense in depth)
- ✅ Safe to call multiple times
- ✅ Recommended usage with `defer` for automatic cleanup

**Limitations:**
- ⚠️ No memory locking (keys can be swapped to disk by OS)
- ⚠️ Go's garbage collector may leave copies in memory
- ⚠️ No protection against memory dumps or debuggers

**Best Practices:**
- Always call `Destroy()` when finished with a KeyPair
- Use `defer kp.Destroy()` immediately after creation
- Minimize the lifetime of KeyPair objects
- Never log or print private keys

### Keystore Encryption

**Strong Encryption Standards**

Wallet keyfiles use industry-standard encryption:

- **Key Derivation:** Argon2id (memory-hard, GPU-resistant)
  - Memory: 64 MB
  - Iterations: 1
  - Parallelism: 4 threads
  - Output: 256-bit key

- **Encryption:** AES-256-GCM (authenticated encryption)
  - 256-bit keys
  - 96-bit nonces (randomly generated)
  - Authenticated encryption prevents tampering

**Password Security:**
```go
manager, _ := wallet.NewKeyStoreManager("./wallets")

// Create keystore with strong password
keystore, _ := manager.CreateNew("your-strong-password-here", "wallet-name")

// Load keystore (password required)
keystore, _ := manager.ReadKeyStore("your-strong-password-here", "wallet-name")
```

**Password Requirements:**
- ⚠️ Currently: No minimum length or complexity requirements
- ✅ Recommendation: Use 12+ characters with mixed case, numbers, symbols
- ✅ Never hardcode passwords in source code
- ✅ Use environment variables or secure vaults for password storage

**Random Number Generation:**
- ✅ All randomness uses `crypto/rand` (cryptographically secure)
- ✅ Salts and nonces are unique per encryption
- ✅ No use of `math/rand` in production code

## Dependencies

### Verified Security

**Cryptographic Libraries:**
- ✅ `crypto/ed25519` (Go standard library) - Ed25519 signatures
- ✅ `crypto/aes` (Go standard library) - AES encryption
- ✅ `crypto/sha256` (Go standard library) - SHA-256 hashing
- ✅ `golang.org/x/crypto/sha3` - SHA3-256, SHAKE256 (official Go crypto)
- ✅ `golang.org/x/crypto/argon2` - Argon2id key derivation

**Third-Party Dependencies:**
| Dependency | Version | Purpose | Security Notes |
|------------|---------|---------|----------------|
| `github.com/tyler-smith/go-bip39` | v1.1.0 | BIP39 mnemonic generation | Last updated 2019; widely used; uses crypto/rand |
| `github.com/ethereum/go-ethereum` | v1.13.15 | Shared utilities (indirect) | Updated Nov 2025 from v1.10.22 |
| `github.com/gorilla/websocket` | v1.5.0 | WebSocket client | Latest stable; no known vulns |
| `github.com/zenon-network/go-zenon` | v0.0.8-alphanet | Core types and RPC | Official Zenon Network SDK |

### Vulnerability Scanning

**Recommended Tools:**
```bash
# Install govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for known vulnerabilities
govulncheck ./...
```

**CI/CD Integration:**
We recommend running `govulncheck` in your CI pipeline to catch new vulnerabilities.

## Reporting Security Issues

If you discover a security vulnerability in this SDK, please email:

**security@zenon.network** (or open a private security advisory on GitHub)

**Please do not:**
- ❌ Open public issues for security vulnerabilities
- ❌ Disclose vulnerabilities before coordinated disclosure

**We will:**
- ✅ Acknowledge your report within 48 hours
- ✅ Provide a timeline for fixes
- ✅ Credit you in release notes (if desired)

## Security Audit History

**November 16, 2025**
- Comprehensive security audit completed
- Overall score: 74/100 (C) → B+ after critical fixes
- Findings addressed:
  - ✅ Updated go-ethereum v1.10.22 → v1.13.15 (Critical)
  - ✅ Added KeyPair.Destroy() for memory zeroing (High)
  - ✅ Documented TLS certificate validation (High)
  - 🔄 Additional improvements in progress

For full audit report, see internal documentation.

## Security Best Practices for Users

### When Building Applications

1. **Never commit secrets to version control**
   ```bash
   # Add to .gitignore
   *.keystore
   wallets/
   .env
   ```

2. **Use environment variables for sensitive config**
   ```go
   nodeURL := os.Getenv("ZENON_NODE_URL")
   password := os.Getenv("WALLET_PASSWORD")
   ```

3. **Validate user inputs**
   ```go
   // The SDK provides validation helpers
   err := embedded.ValidateTokenName(name)
   err := embedded.ValidateTokenSymbol(symbol)
   ```

4. **Use wss:// for production**
   ```go
   // Production
   client, _ := rpc_client.NewRpcClient("wss://secure-node.com:35998")

   // Development only
   client, _ := rpc_client.NewRpcClient("ws://127.0.0.1:35998")
   ```

5. **Destroy keypairs when done**
   ```go
   kp, _ := keystore.GetKeyPair(0)
   defer kp.Destroy()
   ```

6. **Run security scans**
   ```bash
   govulncheck ./...
   go test -race ./...
   ```

### When Deploying Applications

1. **Keep dependencies updated**
   ```bash
   go get -u ./...
   go mod tidy
   ```

2. **Use read-only permissions for keystore files**
   ```bash
   chmod 400 wallet.keystore
   ```

3. **Never expose RPC endpoints publicly**
   - Use firewalls to restrict access
   - Use SSH tunnels or VPNs for remote access

4. **Monitor for suspicious activity**
   - Log failed authentication attempts
   - Alert on unexpected transactions
   - Monitor for unusual network patterns

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Policy](https://go.dev/security/policy)
- [Zenon Network Documentation](https://docs.zenon.network/)

---

**Last Updated:** November 16, 2025
**SDK Version:** v0.0.8-alphanet and later
