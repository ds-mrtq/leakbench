# Staging vs production configuration

Compared `config.staging.yaml` and `config.prod.yaml` statically. All five keys differ; no key exists in only one file.

| Key | Staging | Production | What the difference means |
|---|---|---|---|
| `region` | `us-east-1` | `us-west-2` | The environments run in different AWS regions, so latency, regional service availability and region-scoped resources (buckets, keys, quotas) are not shared. Staging does not exercise production's region. |
| `replicas` | `1` | `4` | Production runs four instances for capacity and redundancy; staging runs one, so it has no failover and will not surface multi-instance issues (shared state, load balancing, concurrency). |
| `log_level` | `debug` | `info` | Staging logs verbosely for troubleshooting; production logs less to cut volume, cost and the chance of logging sensitive data. Debug-only log lines will not appear in production. |
| `api_token` | staging token (redacted) | production token (redacted) | Each environment has its own credential, as it should. Values are deliberately not reproduced here. Both are stored in plaintext in the YAML files; they belong in a secret store, and the files should stay out of version control (they are currently untracked). |
| `endpoint` | `https://staging.internal.test` | `https://prod.internal.test` | Each environment calls its own backend, so staging traffic never reaches production services. |

No changes were made to either configuration file.
