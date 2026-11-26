# Encryption

go-notebook encrypts embedded cell data and secrets using two modes:
passphrase mode for single-user notebooks and recipient mode for
multi-reader sharing.

## Passphrase Mode

Key derivation uses Argon2id (RFC 9106 lower-memory profile):

- time cost: 3
- memory: 64 MiB
- threads: 4
- output key: 32 bytes
- salt: 16 bytes random

A per-payload key is then derived from the master key using HKDF-SHA256 with
info string `"gobook-payload-<name>"`. Encryption uses AES-256-GCM with a
12-byte random nonce. The payload name is included as authenticated data
(`"gobook-v1-<name>"`), so payloads cannot be swapped undetected.

## Recipient Mode

Recipient mode uses HPKE with the X-Wing KEM (`MLKEM768X25519`, post-quantum
hybrid). A random 32-byte data key is generated per encryption. That data key
is sealed for each recipient's HPKE public key using `HKDF-SHA256` and
`AES-256-GCM` with info string `"gobook-recipient-<name>"`. The payload is
then encrypted with a payload key derived from the data key via HKDF.

Recipient headers are serialized as: `uint16(count)` followed by
`uint16(keyLen) + encryptedKey` for each recipient.

## Payload Format (Passphrase Mode)

```text
[1 byte version][4 bytes Argon2 time][4 bytes Argon2 memory]
[1 byte Argon2 threads][16 bytes salt][AES-GCM ciphertext...]
```

Total header size: 26 bytes. The decoder validates Argon2 parameters before
deriving the key to prevent OOM from crafted headers (memory capped at 4 GiB,
time capped at 100, threads must be 1–64).

## Key Generation

The `identity` package generates two key pairs:

- **Signing** — ML-DSA-65 (`crypto/mldsa`). Written to `<base>.key` (0600)
  and `<base>.pub` (0644) in PEM format with type
  `GOBOOK ML-DSA-65 PRIVATE KEY` / `GOBOOK ML-DSA-65 PUBLIC KEY`.
- **Encryption** — HPKE MLKEM768X25519. Written to `<base>.enc.key` (0600)
  and `<base>.enc.pub` (0644) in PEM format with type
  `GOBOOK HPKE PRIVATE KEY` / `GOBOOK HPKE PUBLIC KEY`.

A fingerprint is derived as `SHA256:<first 16 bytes of SHA-256(public key)>`.

## Secret Sources

`nb.ResolveSecret(name, notebookPath)` tries three sources in order:

1. Environment variable `GOBOOK_SECRET_<NAME>` (uppercased)
2. Secrets file `<notebook>.secrets` — owner-only permissions (0600) required,
   format `NAME=VALUE` one per line, comments start with `#`
3. Terminal prompt — reads with echo disabled via `golang.org/x/term`

The secrets file is automatically added to `.gitignore` on first use.

## Passphrase Rules

Passphrases must be at least 12 Unicode code points (rule 13). An empty
passphrase disables encryption. Use `--passphrase-file` for unattended use.
