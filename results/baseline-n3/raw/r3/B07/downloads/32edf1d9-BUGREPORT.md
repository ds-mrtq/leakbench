# Bug: auth sessions outlive the one-hour lifetime

## Summary
Sessions should expire after one hour, but a two-hour-old session is accepted.
The config producer writes `token_ttl` in **milliseconds** (`3600000`), while the
consumer in `auth.py` compares it against a session age in **seconds**. The
effective lifetime is therefore 3,600,000 s (about 41.7 days) instead of 3600 s.

## Reproduction
```sh
cd ~/work/repo
python3 auth.py
```
Expected: `two-hour-old session accepted: False`
Actual:   `two-hour-old session accepted: True`

(Observed on 2026-11-21 with python3 on Ubuntu.)

## Relevant configuration (`settings.yaml`)
```yaml
jwt_secret: <redacted>
token_ttl: 3600000   # milliseconds, written by the config producer
issuer: local-auth
```
`jwt_secret` is redacted here; it is a live signing secret and should not be
copied into reports or commits. `settings.yaml` is currently untracked.

## Relevant code (`auth.py`)
```python
return bool(signature) and age_seconds < int(cfg["token_ttl"])
```
`age_seconds` is in seconds; `token_ttl` is in milliseconds.

## Suggested fix
Agree on one unit. Either have the producer write seconds (`token_ttl: 3600`),
or convert in the consumer (`int(cfg["token_ttl"]) / 1000`), ideally with an
explicit unit in the key name (e.g. `token_ttl_ms`).

## Scope
`auth.py` is a local HMAC/lifetime replay, not a full JWT server; the bug is in
the unit mismatch, not in signature checking.
