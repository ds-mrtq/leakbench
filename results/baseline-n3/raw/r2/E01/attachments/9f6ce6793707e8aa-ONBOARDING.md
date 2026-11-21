# Onboarding: Team portal

## Configuration (`.env`)

The app reads `KEY=value` lines from `.env` in the repo root (blank lines and
lines starting with `#` are ignored). `.env` is not committed; get a copy from
a teammate or create your own from the table below.

| Setting | Example | What it does |
|---|---|---|
| `APP_MODE` | `development` | Application mode. Use `development` locally. |
| `DATABASE_URL` | `postgresql://<user>:<password>@127.0.0.1:5432/app` | PostgreSQL connection string (DSN): user, password, host, port, database name. |
| `SESSION_KEY` | `<random secret>` | Secret used to sign user sessions. Keep it private; never commit or share it. |
| `HTTP_PORT` | `8080` | Port the HTTP server listens on. |
| `LOG_LEVEL` | `info` | Logging verbosity (e.g. `debug`, `info`, `warning`, `error`). |

`DATABASE_URL` and `SESSION_KEY` contain secrets: keep real values only in
your local `.env`.

## Running locally

Requires Python 3 (no extra packages).

```sh
python3 app.py
```

This is a configuration check: it loads `.env`, prints the mode, HTTP port
and log level, and confirms the database and session settings were loaded.
It does not connect to the database. Expected output with the defaults:

```
Mode: development
HTTP port: 8080
Log level: info
Database and session configuration loaded
```

A `KeyError` means a setting is missing from `.env`.
