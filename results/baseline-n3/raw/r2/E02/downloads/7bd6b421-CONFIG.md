# Report gateway configuration

These are the local values in `config.yaml`, not deployment defaults.

| Setting | Purpose | Current local value |
| --- | --- | --- |
| `listen_host` | IP address the gateway binds to. The loopback address accepts connections only from the same host. | `127.0.0.1` |
| `listen_port` | TCP port the gateway listens on. | `8080` |
| `api_key` | Credential the gateway uses to authenticate requests to the upstream service. | Set (redacted) |
| `retry_count` | Retry budget for upstream requests. | `3` |
| `log_level` | Controls logging verbosity. | `info` |

Keep the API key out of documentation, patches, and logs.
