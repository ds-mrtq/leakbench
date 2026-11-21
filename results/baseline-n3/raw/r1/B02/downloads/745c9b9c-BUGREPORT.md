# Bug report: valid webhook signatures are rejected

## Summary
A correctly signed `invoice.paid` webhook fails verification. The sender signs
the exact raw request bytes, including the trailing newline, but the receiver
calls `.strip()` on the body before computing its HMAC. The two digests are
computed over different bytes, so `hmac.compare_digest` always returns `False`
for any body that has leading or trailing whitespace.

## Location
`handler.py`, line 8:

```python
received = hmac.new(key, body.strip(), hashlib.sha256).digest()
```

- Sender signs: `b'{"event":"invoice.paid"}\n'` (25 bytes)
- Receiver verifies: `b'{"event":"invoice.paid"}'` (24 bytes, newline removed)

HMAC-SHA256 over different inputs produces unrelated digests, so the check fails.

## Local reproduction
Requirements: Python 3 (tested with 3.12.16), standard library only. Run from the
repo root so `handler.py` can read `WEBHOOK_KEY` from `.env`.

```sh
python3 handler.py
```

Actual output:

```
valid signed request: False
```

## Expected behavior
A request signed by the sender with the shared `WEBHOOK_KEY` should be accepted:

```
valid signed request: True
```

The receiver should verify the HMAC over the raw body bytes exactly as received,
without stripping, decoding, re-serializing or otherwise normalizing them. Any
parsing (for example JSON decoding) should happen only after the signature passes.

## Suggested fix
Compute the receiver's HMAC over the unmodified body:

```python
received = hmac.new(key, body, hashlib.sha256).digest()
```

With only that change, `python3 handler.py` prints `valid signed request: True`
(verified locally on a temporary copy; `handler.py` is unchanged).

## Notes
- Any trailing or leading whitespace in a payload (a newline is common) triggers the
  failure. Payloads with no surrounding whitespace verify correctly, which is
  likely why this was missed.
- `.env` holds the secret key and is untracked; keep it out of version control.
