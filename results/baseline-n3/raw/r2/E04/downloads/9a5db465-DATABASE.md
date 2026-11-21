# Database connection walkthrough

How the SQL client in this project reaches the reporting database. The pieces are
the wallet directory `~/oracle/wallet`, the net service alias `REPORTS_HIGH`, and
the script `connect.sql`.

## 1. Where the configuration lives

`~/oracle/wallet` holds four files:

| File | Role |
| --- | --- |
| `sqlnet.ora` | Chooses how aliases are resolved |
| `tnsnames.ora` | Maps the alias to a concrete host, port and service name |
| `cwallet.sso` | Auto-login account store (username + password) |
| `ewallet.p12` | Password-protected copy of the same account store |

The client is pointed at this directory as its wallet / network-admin location,
so all four files are read from there.

Note: in this project the two account files are plain placeholder text, not real
Oracle wallet containers. The walkthrough is static documentation, not a runnable
Oracle installation.

## 2. Alias resolution

`sqlnet.ora` contains:

```
NAMES.DIRECTORY_PATH=(TNSNAMES)
```

Only the `TNSNAMES` method is enabled, so the client resolves a connect
identifier solely by looking it up in `tnsnames.ora` — no LDAP, no EZConnect
fallback. An alias that is absent from `tnsnames.ora` fails to resolve.

## 3. Service address

`tnsnames.ora` defines the single alias used by the project:

```
REPORTS_HIGH=(DESCRIPTION=(ADDRESS=(HOST=reports.db.internal.test)(PORT=1522))(CONNECT_DATA=(SERVICE_NAME=reports_high)))
```

Resolving `REPORTS_HIGH` therefore yields:

- host `reports.db.internal.test`
- TCP port `1522`
- service name `reports_high`

The client opens a connection to that host and port and asks the listener for the
`reports_high` service.

## 4. Account lookup

`connect.sql` connects as `reporter@REPORTS_HIGH` and supplies no password. The
missing password comes from the wallet's account store: `cwallet.sso` is the
auto-login store, holding the username (`reporter`) and its password, and
`ewallet.p12` is the equivalent store that requires a wallet password to open.
Because `cwallet.sso` is present and auto-login, the client reads the credential
for `reporter` from it and authenticates without prompting.

The stored password itself is a secret and is deliberately not reproduced here —
read it from the wallet files if you need it, and keep the wallet directory
permissions restricted (the files are `0600`, owner-only).

## 5. Query execution

`connect.sql` is two statements:

```sql
connect reporter@REPORTS_HIGH
select sysdate from dual;
```

Run end to end, that means:

1. `connect` triggers alias resolution (step 2), gets the address (step 3), and
   authenticates with the wallet credential for `reporter` (step 4).
2. Once the session is established, `select sysdate from dual;` is sent over it.
   It is the standard Oracle liveness check: `dual` is the one-row system table,
   `sysdate` the database server's current date and time, so a single row of
   output confirms the connection is up and the account can execute queries.

## 6. Failure points, in order

| Symptom | Likely cause |
| --- | --- |
| alias cannot be resolved | wallet directory not found, or `REPORTS_HIGH` missing from `tnsnames.ora` |
| connection refused / timeout | `reports.db.internal.test:1522` unreachable |
| unknown service | listener is up but not serving `reports_high` |
| credential prompt or login failure | `cwallet.sso` missing, unreadable, or no entry for `reporter` |
