# Nightly job: ORA-12154 connection error

## Summary
The nightly job fails with `ORA-12154: could not resolve connect identifier`. It connects using the alias `JOBS_OLD`, but the client's `tnsnames.ora` only defines `JOBS_HIGH`. Name resolution fails on the client before any connection reaches the database.

## Local reproduction
Run from the project root (`~/work/repo`), with client files in `~/oracle/wallet`:

```
$ python3 job.py
ORA-12154: could not resolve connect identifier
$ echo $?
1
```

The job finds all the client files it needs (`cwallet.sso`, `ewallet.p12`, `sqlnet.ora`). It fails only when it looks up `JOBS_OLD=` in `tnsnames.ora`.

Note: this is an offline replay of alias resolution. The wallet files are placeholders, so no real database was contacted.

## Relevant connection settings
| Setting | Location | Value |
|---|---|---|
| Alias used by job | `job.json` | `JOBS_OLD` |
| Query | `job.json` | `select count(*) from nightly_jobs` |
| Naming method | `sqlnet.ora` | `NAMES.DIRECTORY_PATH=(TNSNAMES)` (tnsnames only, no LDAP/EZCONNECT fallback) |
| Defined alias | `tnsnames.ora` | `JOBS_HIGH` |
| Host / port | `tnsnames.ora` | `db.internal.test` / `1522` |
| Service name | `tnsnames.ora` | `jobs_high` |
| Wallet | `~/oracle/wallet` | `cwallet.sso`, `ewallet.p12` present (contents not included) |

## Questions for the DBA
1. Was `JOBS_OLD` retired or renamed? The README says the available service is `JOBS_HIGH`.
2. Should the job switch to `JOBS_HIGH` (`jobs_high` on `db.internal.test:1522`), or should a `JOBS_OLD` entry be restored in the distributed `tnsnames.ora`?
3. Do the wallet credentials work for the `jobs_high` service?

## Likely fix
Change `alias` in `job.json` to `JOBS_HIGH`, once the DBA confirms that is the right service. I have not made this change.
