# Bug report: Python SDK request fails with 404 (retired v1 route)

## Summary
The SDK preflight fails with `404: client requested retired v1 route`. The service has migrated to `/v2`, but the configured endpoint (`SDK_URL`, set in `~/.profile` via `env.sh`) still points at the `/v1` route.

## Steps to reproduce
1. Set up the workspace: `mkdir -p ~/work && tar -xzf project.tar.gz -C ~/work && bash ~/work/bootstrap.sh`
2. Open a shell and load the environment: `. ~/.profile`
3. `cd ~/work/repo && python3 sdk.py`

## Expected
`SDK request preflight passed` (the request goes to the `/v2` route).

## Actual
```
Python: 3.12.16
Endpoint: https://sdk.internal.test/v1
404: client requested retired v1 route
```
The process exits with status 1.

## Environment
- OS: Ubuntu 22.04.5 LTS, kernel Linux 6.8.0-1060-aws x86_64
- Python: 3.12.16 (standard library only; no network access needed)
- Repo commit: `ca3ad13` (Initial project)
- `SDK_URL=https://sdk.internal.test/v1` (from `~/.profile` / `env.sh`)
- `SDK_API_TOKEN`: set (value redacted; it is a credential and is not needed to reproduce, any non-empty value works)

## Likely cause / fix
`sdk.py` requires the `SDK_URL` path to be exactly `/v2`. The environment still exports the old `/v1` URL. Updating it to `https://sdk.internal.test/v2` should fix it; this is not verified against the real service.
