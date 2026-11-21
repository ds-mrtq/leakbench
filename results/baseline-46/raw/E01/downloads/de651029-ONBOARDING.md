# Onboarding: Team portal configuration

The app reads its settings from a `.env` file in the repo root (one `KEY=value`
per line; blank lines and lines starting with `#` are ignored). `.env` holds
credentials and is not committed — ask a teammate for the values, never paste
them into docs, tickets or chat.

## Settings

| Key | What it controls | Example (not the real value) |
| --- | --- | --- |
| `APP_MODE` | Which mode the app runs in, e.g. `development` locally. | `development` |
| `DATABASE_URL` | PostgreSQL connection DSN (user, password, host, port, database). Secret. Not contacted during the local check. | `postgresql://user:password@localhost:5432/portal` |
| `SESSION_KEY` | Secret used to sign user sessions. Keep it private; rotating it logs everyone out. | `<random 32+ byte string>` |
| `HTTP_PORT` | Port the HTTP server listens on. | `8080` |
| `LOG_LEVEL` | Logging verbosity (`debug`, `info`, `warning`, `error`). | `info` |

## Running locally

Requires Python 3 (standard library only).

1. Create `.env` in the repo root with the five keys above.
2. From the repo root run:

   ```sh
   python3 app.py
   ```

   This is a configuration check: it prints the mode, HTTP port and log level,
   then confirms the database and session settings loaded. It does not connect
   to the database. A `KeyError` means a key is missing from `.env`.
