# 04 — ignored errors fail *open*

## The trap

In Go, errors are ordinary return values. Many security-critical functions signal "this failed" **only** through that error — the other return value (or the absence of one) is not, by itself, a verdict. Drop the error and the check silently *passes*.

`crypto/rsa`'s signature verification is the clean example: `VerifyPKCS1v15` returns `nil` for a valid signature and a non-nil `error` for an invalid one. It has no boolean. Called as a bare statement, the error is discarded — and a forged message is treated as authentic.

```go
// FOOTGUN: the only signal of a bad signature is discarded → fail OPEN
rsa.VerifyPKCS1v15(&pub, crypto.SHA256, tampered[:], sig)
process(tampered) // proceeds as if the signature checked out
```

## The correct shape

The error *is* the check. Handle it, and an invalid signature stops execution.

```go
if err := rsa.VerifyPKCS1v15(&pub, crypto.SHA256, tampered[:], sig); err != nil {
    return err // reject — do NOT proceed
}
process(tampered)
```

Run the reproducer to see both paths:

```
$ go run ./04-error-fail-open
FAIL-OPEN   → processed: transfer £9000 to mallory
FAIL-CLOSED → rejected: crypto/rsa: verification error
```

## Why Go differs from other languages

This is the sharpest security contrast between Go and exception-based languages.

| | Verification fails, result ignored |
|---|---|
| **Python** (`cryptography`) | `verify()` **raises** `InvalidSignature`; ignoring it **crashes** the request — fail *closed*, loudly. |
| **Java** (`Signature.verify`) | returns `boolean`; ignoring it is a type error waiting to happen, and the common shape *forces* you to read the result. |
| **Go** | returns `error`; you may call the function as a statement and **drop the error entirely** — execution continues as if valid. Fail *open*, silently. |

Same mistake, opposite blast radius. Go's error-as-value model is what lets an unhandled failure become a silent authorization bypass rather than a crash.

## What breaks

`VerifyPKCS1v15` returns `nil` for success and — critically — the caller who ignores the error never learns the signature was forged. There is no panic, no log, no type mismatch. The bug ships and passes every "happy path" test, because the happy path *also* ignores the error and *also* proceeds.

The same footgun applies to any function whose failure is error-only:

- `rsa.VerifyPKCS1v15`, `rsa.VerifyPSS`, `ecdsa`/`ed25519` verify wrappers
- `(*x509.Certificate).CheckSignature`, `cert.Verify`
- `hmac`/AEAD open (`cipher.AEAD.Open` returns `(plaintext, error)` — using the plaintext without checking the error uses *unauthenticated* data)
- `json.Unmarshal`, `strconv.Atoi`, … — where the ignored error leaves a zero value that downstream logic then trusts

## How to spot it in review

- A `Verify*` / `Check*` / `Open` / `Decrypt*` call as a **bare statement** (no `if err :=` around it, no assignment).
- `foo, _ := verify(...)` or `_ = verify(...)` — the `_` on a security function is a red flag, not a convenience.
- Any security decision that reads a *value* returned alongside an error without first checking the error (the value is only meaningful when `err == nil`).

## The fix / the habit

> If a function can tell you a security check failed, it tells you through its `error`. Never call such a function as a bare statement, and never `_` its error. `err == nil` is the authorization; the other return value is only valid once you've confirmed it.

Tooling: `errcheck` and `staticcheck` (SA4006/unchecked errors) flag discarded errors; `gosec` flags several of the crypto-specific shapes. Plain `go vet` does **not** catch a discarded error, so add one of these to CI.

## References

- [`rsa.VerifyPKCS1v15`](https://pkg.go.dev/crypto/rsa#VerifyPKCS1v15) — returns `nil` on success, `error` on failure; there is no boolean
- [`cipher.AEAD`](https://pkg.go.dev/crypto/cipher#AEAD) — `Open` returns `(plaintext, error)`; the plaintext is unauthenticated until the error is checked
- [errcheck](https://github.com/kisielk/errcheck) · [staticcheck](https://staticcheck.dev/) — catch discarded errors that `go vet` misses
