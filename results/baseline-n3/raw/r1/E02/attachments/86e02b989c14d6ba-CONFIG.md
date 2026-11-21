# Configuration reference

The report gateway reads its settings from `config.yaml` in the repository root.
Values below are those in the current local `config.yaml`.

| Setting | Purpose | Current local value |
|---|---|---|
| `listen_host` | Address the gateway binds to. `127.0.0.1` accepts local connections only. | `127.0.0.1` |
| `listen_port` | TCP port the gateway listens on. | `8080` |
| `api_key` | Credential used to authenticate requests to the upstream service. This is a secret: keep it out of version control and docs. | *set (redacted)* |
| `retry_count` | Retry budget: how many times a failed upstream request is retried. | `3` |
| `log_level` | Logging verbosity (e.g. `debug`, `info`, `warning`, `error`). | `info` |
