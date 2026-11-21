# Bug: valid webhook signatures are rejected

## Summary
Signature verification in `handler.py` computes the HMAC-SHA256 over `body.strip()`
instead of the raw request bytes. The sender signs the exact body, including the
trailing newline, so the two digests never match and every valid `invoice.paid`
request is rejected.

## Root cause
`handler.py`, line 8:

```python
received = hmac.new(key, body.strip(), hashlib.sha256).digest()
```

`.strip()` removes the final `\n`, so the verifier hashes different bytes than the
sender signed. HMAC verification must run on the raw, unmodified body.

## Local reproduction
Uses only the Python standard library and the key in `.env` (`WEBHOOK_KEY=...`).

```sh
cd repo
python3 handler.py
```

Actual output:

```
valid signed request: False
```

## Expected behavior
A request whose signature was computed over the exact body bytes
(`{"event":"invoice.paid"}\n`) must verify:

```
valid signed request: True
```

## Suggested fix
Hash the raw body without normalizing it:

```python
received = hmac.new(key, body, hashlib.sha256).digest()
```

Keep `hmac.compare_digest` for the comparison. Any parsing or whitespace handling
should happen only after the signature has been verified.
