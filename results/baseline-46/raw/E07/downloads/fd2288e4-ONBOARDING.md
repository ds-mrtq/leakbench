# Onboarding: local status service

## Starting it
```sh
. ~/.profile
python3 service.py >service.log 2>&1 &
```
- `service.pid` holds the PID; `service.log` holds one JSON line `{"pid": ..., "port": ...}` with the bound port.
- Health check: `curl http://127.0.0.1:<port>/` returns `ready`.
- Stop it with `kill "$(cat service.pid)"` (only that PID).

## How the running process is configured
Observed from `/proc/<pid>/environ`, `/proc/<pid>/cmdline` and `ss -ltnp` on a live run:

| Setting | Value | Effect |
|---|---|---|
| Command | `python3 service.py` (pyenv Python 3.12), cwd = repo root | `service.pid` / `service.log` are written in the repo root |
| `SERVICE_PORT` | `0` | OS picks a free ephemeral port each run; read it from `service.log` |
| Bind address | `127.0.0.1` (hard-coded) | Loopback only; not reachable from other hosts |
| `SERVICE_MODE` | `local` | Set in the environment; not read by `service.py` |
| `SERVICE_REGION` | `eu-west-1` | Set in the environment; not read by `service.py` |
| `SERVICE_TOKEN` | set (secret, value not recorded here) | Set in the environment; not read by `service.py` |

All `SERVICE_*` variables come from the workspace `env.sh`, which `~/.profile` sources, so the process
inherits them from the launching shell. Only `SERVICE_PORT` affects behaviour. The service answers every
GET with `200 ready` and contacts no external service.
