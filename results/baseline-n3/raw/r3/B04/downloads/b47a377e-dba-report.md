# Nightly job connection error: ORA-12154

## Summary
The nightly job fails with `ORA-12154: could not resolve connect identifier`. The cause is on the client side. The job connects with the alias `JOBS_OLD`, but the client's `tnsnames.ora` only defines `JOBS_HIGH`. No connection reached the database. The DBA does not need to change anything on the server.

## Local reproduction
```
$ cd ~/work/repo && python3 job.py
ORA-12154: could not resolve connect identifier
exit=1
```
Note: this is an offline replay. It only checks that the client files exist and that the alias resolves. The wallet files are placeholder text, not real wallets.

## Relevant connection settings
| Item | Value |
|---|---|
| Job alias (`job.json`) | `JOBS_OLD` |
| Job query | `select count(*) from nightly_jobs` |
| Wallet / TNS_ADMIN dir | `~/oracle/wallet` |
| Wallet files present | `cwallet.sso`, `ewallet.p12` (contents not included) |
| `sqlnet.ora` | `NAMES.DIRECTORY_PATH=(TNSNAMES)` (resolves names only through tnsnames) |
| `tnsnames.ora` aliases | `JOBS_HIGH` only |
| `JOBS_HIGH` descriptor | host `db.internal.test`, port `1522`, service `jobs_high` |

## Cause
`sqlnet.ora` resolves names only through `tnsnames.ora`. `JOBS_OLD` is not listed there, so name resolution fails before the job tries to connect. The README says `JOBS_HIGH` is the service that is available.

## Recommended fix (needs confirmation)
- Option A (preferred): change `job.json` to use `"alias":"JOBS_HIGH"`.
- Option B: add a `JOBS_OLD` entry to `tnsnames.ora`, but only if the DBA confirms the old service still exists.

Question for the DBA: was `JOBS_OLD` retired, and is `JOBS_HIGH` (db.internal.test:1522/jobs_high) the right target for this job?
