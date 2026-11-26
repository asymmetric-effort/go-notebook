# Encryption

go-notebook supports post-quantum encryption for report binaries. Encrypted
reports can only be rendered by someone who has the correct passphrase or
private key.

## Encryption Modes

There are two modes of encryption:

### Passphrase Mode

Encrypt a report with a passphrase using Argon2id key derivation and
AES-256-GCM authenticated encryption:

```bash
gobook build report.md -o report --passphrase
```

The binary prompts for a passphrase at build time. At render time, the user
must provide the same passphrase:

```bash
./report --format pdf -o report.pdf --passphrase
```

### Recipient Mode

Encrypt a report for specific recipients using HPKE with ML-KEM-768-X25519
(post-quantum key encapsulation):

```bash
gobook build report.md -o report --recipient alice.pub --recipient bob.pub
```

Each recipient decrypts with their private key:

```bash
./report --format pdf -o report.pdf --key-file alice.key
```

## Key Generation

Generate an encryption key pair:

```bash
gobook keygen -o mykey
```

This creates two files:

- `mykey.pub` — Public key (share with report authors)
- `mykey.key` — Private key (keep secret)

Keys use HPKE with ML-KEM-768-X25519 for post-quantum security.

## Digital Signatures

Sign reports with ML-DSA-65 (post-quantum digital signatures):

```bash
gobook build report.md -o report --sign-key author.sign
```

Verify a report's signature:

```bash
gobook verify --signature report.sig --signer author.sign.pub
```

## Cryptographic Details

| Component | Algorithm |
|-----------|-----------|
| Key encapsulation | HPKE ML-KEM-768-X25519 |
| Symmetric encryption | AES-256-GCM |
| Key derivation | Argon2id + HKDF-SHA256 |
| Digital signatures | ML-DSA-65 |
| Binary header | Versioned with magic bytes |

All cryptographic primitives use Go's standard `crypto` package
implementations. No external C libraries or cgo dependencies.

## Security Model

- Encrypted payloads use a versioned binary header format for forward
  compatibility
- Passphrase mode uses Argon2id with recommended parameters for memory-hard
  key derivation
- Recipient mode supports multiple recipients — each gets their own
  encapsulated key
- The report binary contains the encrypted payload; the decryption key
  never touches disk unless explicitly saved
