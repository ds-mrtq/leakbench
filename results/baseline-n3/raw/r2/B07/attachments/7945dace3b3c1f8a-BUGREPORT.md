# Bug: auth sessions outlive their one-hour lifetime

## Summary
Sessions are meant to expire after one hour, but much older sessions are still
accepted. `settings.yaml` stores `token_ttl` in **milliseconds** (`3600000`),
while `auth.py` compares it against a token age in **seconds**. The effective
lifetime is therefore 3,600,000 s (~41.7 days) instead of 3,600 s (1 hour).

## Relevant configuration
`settings.yaml` (secret redacted):

```yaml
jwt_secret: "<redacted>"
token_ttl: 3600000   # written in milliseconds by the config producer
issuer: local-auth
```

Consumer check in `auth.py`:

```python
return bool(signature) and age_seconds < int(cfg["token_ttl"])
```

`age_seconds` is in seconds; `token_ttl` is used as-is with no unit conversion.

## Reproduction
```sh
python3 auth.py
# two-hour-old session accepted: True
```

Boundary check (same `accepts()` function):

| Session age (s) | Expected | Actual |
|---|---|---|
| 3599 | accepted | accepted |
| 3600 | rejected | accepted |
| 7200 (2 h) | rejected | accepted |
| 3599999 (~41.7 days) | rejected | accepted |
| 3600000 | rejected | rejected |

## Expected vs actual
- Expected: sessions with age >= 3600 s are rejected.
- Actual: sessions are accepted until age reaches 3,600,000 s.

## Impact
Stolen or stale session tokens stay valid ~1000x longer than intended.

## Suggested fix
Agree on one unit for `token_ttl`. Either have the producer write seconds
(`token_ttl: 3600`), or have the consumer convert (`int(cfg["token_ttl"]) // 1000`)
and rename the key to make the unit explicit (e.g. `token_ttl_ms`).

## Notes
This is a local HMAC/lifetime replay, not a full JWT server, so it covers the
lifetime comparison only.
