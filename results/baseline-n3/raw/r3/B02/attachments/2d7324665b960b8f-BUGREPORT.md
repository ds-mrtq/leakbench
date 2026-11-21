# Webhook verifier rejects a correctly signed body with a final newline

## Local reproduction

With Python 3, run this from the directory containing the reviewed `handler.py`.
It copies only the example into a temporary directory and uses a throwaway key.
No network, archive setup scripts, or existing `.env` are needed.

```sh
(
  set -eu
  demo=$(mktemp -d)
  trap 'rm -rf -- "$demo"' EXIT
  cp handler.py "$demo/"
  cd "$demo"
  printf '%s\n' 'WEBHOOK_SECRET=local-test-key-not-for-production' > .env
  python3 -I handler.py
)
```

**Actual:** `valid signed request: False`

**Expected:** `valid signed request: True`. The signature was computed over
the exact request bytes, including the newline after the JSON. Missing or
nonmatching signatures should still be rejected.

## Cause and recommended fix

The sender computes HMAC-SHA256 over `body`, but the verifier computes it over
`body.strip()`. Stripping the trailing newline changes the signed bytes, so
the digests differ even though both sides use the same key. Verify the original
body bytes before parsing or normalizing them, and retain `hmac.compare_digest`
for the comparison. In a local in-memory check, changing only the verifier's
HMAC input from `body.strip()` to `body` produced the expected `True`.
