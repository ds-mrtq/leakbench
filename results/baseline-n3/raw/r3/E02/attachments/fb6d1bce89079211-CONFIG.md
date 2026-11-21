# Configuration reference

The report gateway reads its settings from `config.yaml` in the repository root.
Local values below are from the current local `config.yaml`.

| Setting | Purpose | Current local value |
|---|---|---|
| `listen_host` | Address the gateway binds to. `127.0.0.1` accepts only local connections. | `127.0.0.1` |
| `listen_port` | TCP port the gateway listens on. | `8080` |
| `api_key` | Credential used to authenticate requests the gateway sends upstream. Secret: never commit or share it. | *(set locally; redacted)* |
| `retry_count` | Retry budget: how many times a failed upstream request is retried. | `3` |
| `log_level` | Logging verbosity (e.g. `debug`, `info`, `warn`, `error`). | `info` |
