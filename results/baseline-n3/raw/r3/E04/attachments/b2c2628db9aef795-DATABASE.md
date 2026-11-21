# Database connection walkthrough

The reports SQL client connects with `connect.sql`, using the Oracle wallet in
`~/oracle/wallet`. No password appears in the repo; the wallet supplies it.

> The wallet files here are fake text stand-ins, not real Oracle wallet
> formats. This is a static walkthrough, not a runnable Oracle setup.

## 1. Entry point: `connect.sql`

```sql
connect reporter@REPORTS_HIGH
select sysdate from dual;
```

`reporter@REPORTS_HIGH` gives a username and a net service alias, with no
password. The client resolves the alias to a database address and gets the
credential from the wallet.

## 2. Alias resolution: `sqlnet.ora`

```
NAMES.DIRECTORY_PATH=(TNSNAMES)
```

Aliases are resolved only through `tnsnames.ora` (no LDAP or EZCONNECT). The
client finds both `.ora` files in the wallet directory, so `TNS_ADMIN` must
point at `~/oracle/wallet` (or the files must be in the client's default
network admin directory).

## 3. Service address: `tnsnames.ora`

```
REPORTS_HIGH=(DESCRIPTION=(ADDRESS=(HOST=reports.db.internal.test)(PORT=1522))(CONNECT_DATA=(SERVICE_NAME=reports_high)))
```

| Field        | Value                      |
|--------------|----------------------------|
| Host         | `reports.db.internal.test` |
| Port         | `1522`                     |
| Service name | `reports_high`             |

## 4. Account lookup: the wallet

| File          | Contents (fake text)               | Role                                  |
|---------------|------------------------------------|---------------------------------------|
| `cwallet.sso` | `user=reporter`, `auth=<redacted>` | Auto-login wallet; read without a password |
| `ewallet.p12` | `auth=<redacted>`                  | Password-protected wallet (PKCS#12)   |

The client looks up the stored credential for `reporter` in the auto-login
wallet (`cwallet.sso`) and uses it to sign in. Both files hold the same
secret; it is deliberately not copied here. The directory and files are mode
`700`/`600`; keep them that way.

## 5. Query execution

Once connected as `reporter` to `reports_high`, the script runs
`select sysdate from dual;`, which returns the database server's current date
and time and confirms the session works.

## Flow summary

1. `connect reporter@REPORTS_HIGH` (`connect.sql`)
2. `sqlnet.ora` → resolve via `tnsnames.ora`
3. `tnsnames.ora` → `reports.db.internal.test:1522`, service `reports_high`
4. Wallet (`cwallet.sso`) → credential for `reporter`
5. Run `select sysdate from dual;`
