# Bug: auth sessions outlive the configured one-hour lifetime

## Summary
Sessions should expire after one hour, but a two-hour-old session is still accepted.
The config producer writes `token_ttl` in **milliseconds**, while `auth.py` compares
it against a session age in **seconds**. The effective lifetime is therefore
3,600,000 s (1000 hours, ~41.7 days) instead of 3600 s.

## Reproduction
```sh
cd repo
python3 auth.py
```
Observed (Python 3.12):
```
two-hour-old session accepted: True
```
Expected: `False` — a 7200 s session exceeds the 1 h lifetime.

## Relevant configuration
`settings.yaml` (untracked; secret redacted):
```yaml
jwt_secret: "<redacted>"
token_ttl: 3600000    # written in milliseconds by the config producer
issuer: local-auth
```

## Relevant code
`auth.py`:
```python
return bool(signature) and age_seconds < int(cfg["token_ttl"])
```
`age_seconds` is in seconds; `token_ttl` is used raw with no unit conversion.

## Suggested fix
Agree on one unit. Either have the producer write seconds (`token_ttl: 3600`), or
convert in the consumer (`int(cfg["token_ttl"]) / 1000`), ideally renaming the key
to carry its unit (e.g. `token_ttl_ms`). Add a test that a 7200 s session is rejected.

## Notes
- This is a local HMAC/lifetime replay, not a full JWT server.
- The `jwt_secret` in the local `settings.yaml` is a real-looking value in plaintext;
  it is intentionally not reproduced here and `settings.yaml` should stay out of git.
