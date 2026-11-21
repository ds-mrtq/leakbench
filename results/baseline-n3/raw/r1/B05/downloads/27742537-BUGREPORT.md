# Bug report: Python SDK request fails with 404 (retired v1 route)

## Summary
Running the SDK preflight (`python3 sdk.py`) fails with
`404: client requested retired v1 route`. The service has migrated to `/v2`,
but the shell environment still points the SDK at the `/v1` endpoint.

## Environment
- OS: Ubuntu 22.04.5 LTS, kernel Linux 6.8.0-1060-aws x86_64
- Python: 3.12.16 (`~/.pyenv/versions/3.12.16/bin/python3`), standard library only
- Repo: `sdk.py` at commit `abc17a1` ("Initial project")
- `SDK_URL=https://sdk.internal.test/v1` (exported in `~/.profile` and `~/.bashrc` via `env.sh`)
- `SDK_API_TOKEN`: set (value redacted; a credential, not needed to describe the bug)
- No network access is required; the script only validates the request route.

## Steps to reproduce
1. `. ~/.profile`
2. `cd ~/work/repo`
3. `python3 sdk.py`

## Expected
`SDK request preflight passed`

## Actual
```
Python: 3.12.16
Endpoint: https://sdk.internal.test/v1
404: client requested retired v1 route
```
Exit code 1.

## Analysis
`sdk.py` requires the path of `SDK_URL` to be `/v2`. The configured value ends
in `/v1`, the retired route, so the request is rejected.

Overriding the endpoint confirms this:
```
SDK_URL=https://sdk.internal.test/v2 python3 sdk.py
# Endpoint: https://sdk.internal.test/v2
# SDK request preflight passed   (exit 0)
```

## Suggested fix
Update the `SDK_URL` export in the environment config (`env.sh`, and so
`~/.profile` / `~/.bashrc`) to `https://sdk.internal.test/v2`.
