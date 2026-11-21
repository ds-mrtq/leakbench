# Nightly job connection error: ORA-12154

## Summary
The nightly job fails before connecting with `ORA-12154: could not resolve connect identifier`. The job is configured for alias `JOBS_OLD`, and the client's `tnsnames.ora` only defines `JOBS_HIGH`. Name resolution fails on the client, so no request reaches the database server.

## Local reproduction
Environment: Python 3.12, repo at `~/work/repo`, client files in `~/oracle/wallet`. This is an offline replay of alias resolution only: no real Oracle client or network connection is involved, and the wallet files are placeholder text.

```
$ cd ~/work/repo && python3 job.py
ORA-12154: could not resolve connect identifier
(exit code 1)
```

Control test: the same `job.py` run with the alias set to `JOBS_HIGH` in a scratch copy of `job.json` (the repo was not changed):
```
job connection settings accepted
(exit code 0)
```

## Relevant connection settings
| Setting | Value |
|---|---|
| Job alias (`job.json`) | `JOBS_OLD` |
| Job query | `select count(*) from nightly_jobs` |
| `sqlnet.ora` | `NAMES.DIRECTORY_PATH=(TNSNAMES)`: resolution uses tnsnames only, with no LDAP/EZCONNECT fallback |
| `tnsnames.ora` aliases | `JOBS_HIGH` only |
| `JOBS_HIGH` descriptor | `HOST=db.internal.test`, `PORT=1522`, `SERVICE_NAME=jobs_high` |
| Wallet files | `cwallet.sso` and `ewallet.p12` are present. Their contents are credentials and are left out of this report. |

## Questions for the DBA
1. Was `JOBS_OLD` retired or renamed? If so, is `JOBS_HIGH` (`jobs_high` on `db.internal.test:1522`) the right replacement for this job?
2. If `JOBS_OLD` should still exist, please provide its descriptor so it can be added to `tnsnames.ora`.

Proposed fix once confirmed: change `alias` in `job.json` to `JOBS_HIGH`. No change has been made yet.
