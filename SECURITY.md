# Security Policy - Zenon Go SDK

## Table of Contents

- [Security Overview](#security-overview)
- [Reporting a Vulnerability](#reporting-a-vulnerability)
- [Security Best Practices](#security-best-practices)
- [Threat Model](#threat-model)
- [Cryptographic Implementations](#cryptographic-implementations)
- [Secure Key Management](#secure-key-management)
- [Dependencies](#dependencies)
- [Security Audits](#security-audits)

---

## Security Overview

The Zenon Go SDK is designed with security as a primary concern. It handles cryptographic keys, signs transactions, and manages sensitive user data. This document outlines our security practices, threat model, and best practices for developers using the SDK.

### Security Principles

1. **Defense in Depth** - Multiple layers of security controls
2. **Fail Secure** - Errors default to secure state
3. **Least Privilege** - Minimal permissions and access
4. **Secure by Default** - Safe defaults, opt-in for advanced features
5. **Transparency** - Open source, auditable code

---

## Reporting a Vulnerability

### DO NOT Open a Public Issue

If you discover a security vulnerability, please follow responsible disclosure:

**Email:** security@zenon.network (or open a private security advisory on GitHub)

**Include:**
- Description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Suggested fix (if you have one)
- Your contact information

**Response Timeline:**
- **Initial Response:** Within 48 hours
- **Status Update:** Within 7 days
- **Fix Timeline:** Depends on severity (Critical: <7 days, High: <30 days, Medium: <90 days)

**Bug Bounty:** We currently do not have a formal bug bounty program, but we appreciate security researchers and will acknowledge contributors.

---

## Security Best Practices

### 1. Key Management

#### ✅ DO: Use Destroy() to Clean Up Keys

Always call `Destroy()` on KeyPair instances when done:

```go
keystore, err := manager.ReadKeyStore("password", "my-wallet")
if err != nil {
    return err
}

keypair, err := keystore.GetKeyPair(0)
if err != nil {
    return err
}
defer keypair.Destroy()  // ← ALWAYS defer Destroy()

// Use keypair
signature, err := keypair.Sign(message)
```

**Why:** Zeros private key material in memory to prevent:
- Memory dumps from revealing keys
- Swap file exposure
- Process inspection attacks

#### ✅ DO: Use Strong Passwords

```go
// Validate password before creating keystore
if err := wallet.ValidatePassword(password); err != nil {
    return fmt.Errorf("weak password: %w", err)
}

// Even better: analyze strength
strength := wallet.AnalyzePasswordStrength(password)
if strength < wallet.PasswordStrong {
    return fmt.Errorf("password too weak, use stronger password")
}

keystore, err := manager.CreateNew(password, "my-wallet")
```

**Minimum Requirements:**
- At least 8 characters
- Not all the same character
- **Recommended:** 16+ characters with mixed case, numbers, symbols

#### ❌ DON'T: Log or Print Private Keys

```go
// NEVER DO THIS
log.Printf("Private key: %x", keypair.PrivateKey())  // ❌ DANGER

// Also bad
fmt.Println(keystore.Mnemonic)  // ❌ Exposes recovery phrase
```

### 2. Transaction Security

#### ✅ DO: Use Timeouts for PoW Generation

```go
// Good: With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resultChan := pow.GeneratePowAsync(ctx, hash, difficulty)
result := <-resultChan

if result.Error != nil {
    if errors.Is(result.Error, pow.ErrCancelled) {
        return fmt.Errorf("PoW timed out")
    }
    return result.Error
}
```

---

## Cryptographic Implementations

### Key Derivation Function (KDF)

**Algorithm:** Argon2id
**Parameters:**
- Memory: 64 MB
- Iterations: 1
- Parallelism: 4
- Key Length: 32 bytes

**Why Argon2id:**
- Winner of Password Hashing Competition (2015)
- Resistant to GPU/ASIC attacks (memory-hard)
- Industry standard (OWASP recommended)

### Encryption

**Algorithm:** AES-256-GCM
**Why:** Authenticated encryption (integrity + confidentiality)

### Digital Signatures

**Algorithm:** Ed25519
**Why:** Modern, constant-time, secure

---

**Last Updated:** November 19, 2025
