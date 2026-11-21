# Staging vs production configuration

Static comparison of `config.staging.yaml` and `config.prod.yaml`. All five keys differ; neither file was changed.

| Key | Staging | Production | What the difference means |
|---|---|---|---|
| `region` | `us-east-1` | `us-west-2` | The environments run in different AWS regions, so latency, regional service availability and quotas can differ. Staging does not test region-specific behavior of production. |
| `replicas` | `1` | `4` | Production runs 4 instances for capacity and redundancy; staging runs a single instance with no failover. Concurrency, load-balancing and rolling-deploy issues may not show up in staging. |
| `log_level` | `debug` | `info` | Staging logs verbose debug output for troubleshooting; production logs only info and above to cut volume and avoid leaking detail. Problems visible only in debug logs will not appear in production logs. |
| `api_token` | *(redacted)* | *(redacted)* | Each environment uses its own credential, which is correct (staging cannot act on production). Both tokens are stored in plaintext in the files; they should move to a secret store or environment variables, and the files should stay out of version control. |
| `endpoint` | `https://staging.internal.test` | `https://prod.internal.test` | Each environment talks to its own internal backend, so staging traffic never reaches production services. |

## Notes

- The two files share the same set of keys; there is no setting present in one environment and missing from the other.
- Token values are deliberately not reproduced here.
