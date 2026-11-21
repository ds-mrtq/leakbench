# Database connection walkthrough

The reports SQL client connects with `connect.sql`, using the client config in
`~/oracle/wallet` and the `REPORTS_HIGH` net alias. This is a static
walkthrough: the wallet files are fake text, not real Oracle wallets.

## Files

| File | Role |
| --- | --- |
| `connect.sql` | `connect reporter@REPORTS_HIGH`, then `select sysdate from dual;` |
| `~/oracle/wallet/sqlnet.ora` | `NAMES.DIRECTORY_PATH=(TNSNAMES)`: resolve aliases only via `tnsnames.ora` |
| `~/oracle/wallet/tnsnames.ora` | Defines `REPORTS_HIGH` |
| `~/oracle/wallet/cwallet.sso` | Auto-login wallet: holds the account (`user=reporter`) and its secret |
| `~/oracle/wallet/ewallet.p12` | Password-protected wallet: holds the same secret, no user |

The client finds `sqlnet.ora` and `tnsnames.ora` through the config directory,
so `~/oracle/wallet` must be the client's config directory (normally set with
`TNS_ADMIN=~/oracle/wallet`).

## 1. Alias resolution

`connect.sql` names the alias `REPORTS_HIGH`. `sqlnet.ora` limits naming to
`TNSNAMES`, so the client looks up `REPORTS_HIGH` in `tnsnames.ora` only (no
EZConnect or LDAP fallback). An unknown alias fails here.

## 2. Service address

`tnsnames.ora` maps the alias to:

- Host: `reports.db.internal.test`
- Port: `1522` (not the default 1521)
- Service name: `reports_high`

There is no `PROTOCOL` in the `ADDRESS`, so the client's default (TCP) applies.

## 3. Account lookup

The username comes from `connect.sql`: `reporter`. `cwallet.sso` stores the
same user with its secret, and `ewallet.p12` stores the secret again (the
open-with-password copy of the wallet). The secret is not copied here; read it
from the wallet if needed and never commit it elsewhere.

Caveat for a real Oracle client: `sqlnet.ora` has no `WALLET_LOCATION` and no
`SQLNET.WALLET_OVERRIDE=TRUE`, and `connect reporter@REPORTS_HIGH` names a user
without a password. With real files, the client would prompt for the password
instead of reading it from the wallet. Wallet-only login would need those two
`sqlnet.ora` settings and `connect /@REPORTS_HIGH`.

## 4. Query execution

Once the session to `reports_high` is open as `reporter`, `connect.sql` runs
`select sysdate from dual;`, which returns the database server's current date
and time: a connectivity check that needs no app tables.
