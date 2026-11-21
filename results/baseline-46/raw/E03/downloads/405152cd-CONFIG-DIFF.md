# Staging vs production configuration

Compared statically: `config.staging.yaml` vs `config.prod.yaml`. All five keys differ.

| Key | Staging | Production | Operational meaning |
|---|---|---|---|
| `region` | `us-east-1` | `us-west-2` | Environments run in different AWS regions. Latency, regional service availability and quotas can differ, and staging does not exercise the production region. |
| `replicas` | `1` | `4` | Production runs 4 instances for capacity and redundancy; staging has a single instance, so it cannot surface multi-instance issues (shared state, load balancing, concurrency) and has no failover. |
| `log_level` | `debug` | `info` | Staging logs verbosely for troubleshooting; production logs less to cut volume, cost and the risk of logging sensitive data. Debug-only log output will not appear in production. |
| `api_token` | *(redacted)* | *(redacted)* | Each environment uses its own credential, which is correct isolation. Both files hold the token in plaintext; it should come from a secret store or environment variable instead, and the files should not be committed as they are. |
| `endpoint` | `https://staging.internal.test` | `https://prod.internal.test` | Each environment talks to its own backend, as expected. |

No keys exist in only one file.
