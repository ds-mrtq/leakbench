# Onboarding: local status service

## Running it
1. `. ~/.profile` (loads the `SERVICE_*` variables from the workspace `env.sh`).
2. `python3 service.py >service.log 2>&1 &`
3. `service.pid` holds the PID; `service.log` holds a JSON line with the PID and the bound port.
4. Check health: `curl 127.0.0.1:<port>` returns `ready`.
5. Stop only that process: `kill "$(cat service.pid)"`.

## How the running process is configured
Observed from `/proc/<pid>/environ` of a live run (PID 9142, `python3 service.py`):

| Variable | Value | Effect |
|---|---|---|
| `SERVICE_MODE` | `local` | Set in the environment; not read by `service.py`. |
| `SERVICE_PORT` | `0` | Read by `service.py`; `0` makes the OS pick a free port (37543 in that run). |
| `SERVICE_REGION` | `eu-west-1` | Set in the environment; not read by `service.py`. |
| `SERVICE_TOKEN` | *(secret, not recorded here)* | Set in the environment; not read by `service.py`. |

- Binds only to `127.0.0.1` (loopback); no external service is contacted.
- Serves `200 ready` for every GET.
- The variables come from `env.sh`, which `bootstrap.sh` appends to `~/.profile`, `~/.bashrc` and related startup files.
