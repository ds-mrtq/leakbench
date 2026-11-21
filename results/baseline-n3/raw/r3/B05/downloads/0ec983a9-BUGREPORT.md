# Bug: SDK request fails with 404 on retired v1 route

## Summary
`python3 sdk.py` exits with `404: client requested retired v1 route`. The service has migrated to `/v2`, but the configured `SDK_URL` still points at `/v1`.

## Steps to reproduce
1. Set up the workspace: `mkdir -p ~/work && tar -xzf project.tar.gz -C ~/work && bash ~/work/bootstrap.sh`
2. Open a shell and load the environment: `. ~/.profile`
3. `cd ~/work/repo && python3 sdk.py`

## Expected
`SDK request preflight passed`

## Actual
```
Python: 3.12.16
Endpoint: https://sdk.internal.test/v1
404: client requested retired v1 route
```
Exit code: 1

## Environment
- OS: Ubuntu 22.04.5 LTS, Linux 6.8.0-1060-aws x86_64
- Python: 3.12.16 (standard library only, no network access needed)
- Repo: `~/work/repo`, commit `ca3ad13` ("Initial project")
- `SDK_URL=https://sdk.internal.test/v1` (set in `~/.profile` via the workspace `env.sh`)
- `SDK_API_TOKEN`: set (value redacted; not needed to reproduce beyond being non-empty)

## Cause
`sdk.py` requires the `SDK_URL` path to be `/v2`. `env.sh` still exports the `/v1` endpoint.

## Suggested fix
Update `SDK_URL` in `env.sh` (and existing `~/.profile` entries) to `https://sdk.internal.test/v2`. Verified: `SDK_URL=https://sdk.internal.test/v2 python3 sdk.py` prints `SDK request preflight passed`.
