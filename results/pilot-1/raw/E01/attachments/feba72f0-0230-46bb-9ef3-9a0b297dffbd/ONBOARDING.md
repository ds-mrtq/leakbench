# Onboarding: Team portal configuration

## Requirements
- Python 3 (no third-party packages needed).

## Configuration (`.env`)
The app reads `KEY=VALUE` lines from `.env` in the repo root. Blank lines and lines starting with `#` are ignored. `.env` is **not committed**; ask a teammate for the real values and never commit it or paste its secrets into docs, tickets or chat.

| Setting | Purpose | Example / local value |
|---|---|---|
| `APP_MODE` | Runtime mode of the app. | `development` |
| `DATABASE_URL` | Database connection string (DSN). | `postgresql://<user>:<password>@<host>/app` |
| `SESSION_KEY` | Secret used to sign user sessions. Keep it private; rotating it invalidates existing sessions. | `<random secret>` |
| `HTTP_PORT` | Port the HTTP server listens on. | `8080` |
| `LOG_LEVEL` | Logging verbosity (e.g. `debug`, `info`, `warning`, `error`). | `info` |

All five keys must be present; a missing `APP_MODE`, `HTTP_PORT` or `LOG_LEVEL` makes the check fail with a `KeyError`.

## Running locally
```bash
# 1. Create .env in the repo root with the settings above
# 2. Run the configuration check
python3 app.py
```
Expected output:
```
Mode: development
HTTP port: 8080
Log level: info
Database and session configuration loaded
```
The check only reads `.env`; it does not connect to the database, and it never prints the DSN or session key.
