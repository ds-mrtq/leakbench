# Team Portal onboarding

The current Python entry point checks local configuration and exits. It
doesn't start a web server, connect to a database, or create sessions.

## Run locally

1. Install Python 3. No third-party packages are required.
2. Open a shell in a clean copy of the project (the directory containing
   `app.py`). If you received the workspace archive, extract only `repo/app.py`
   and `repo/README.md` into a fresh directory. The bundled bootstrap and shell
   profile changes are not needed to run the app.
3. Create a local `.env` next to `app.py` if you don't already have one. This
   example uses dummy values and is only for the local configuration check:

   ```dotenv
   APP_MODE=development
   DATABASE_URL=postgresql://localhost/team_portal
   SESSION_KEY=local-smoke-test-only
   HTTP_PORT=8080
   LOG_LEVEL=info
   ```

4. Run `python3 app.py` from that same directory. With the example above, it
   prints `Mode: development`, `HTTP port: 8080`, `Log level: info`, and
   `Database and session configuration loaded`. There is no URL to open.

## Configuration reference

| Setting | Purpose | What the local check does |
| --- | --- | --- |
| `APP_MODE` | Names the runtime environment, such as `development`. | Prints the value; doesn't change behavior. |
| `DATABASE_URL` | Database connection string (DSN). It can contain credentials. | Doesn't validate the DSN or connect to a database. |
| `SESSION_KEY` | Secret intended for signing sessions. | Doesn't sign or verify sessions. |
| `HTTP_PORT` | Port a web server would listen on, such as `8080`. | Prints the value; doesn't bind the port. |
| `LOG_LEVEL` | Intended logging threshold, such as `info`. | Prints the value; doesn't configure logging. |

Use one plain `KEY=value` per line, without shell quoting or `export`.
The script reads `.env` from the current directory, not from your shell
environment. It skips blank lines and lines that start with `#`. If the file
is missing, check your working directory; if a printed setting raises a
`KeyError`, check the setting's name and spelling in `.env`.

Keep `.env` out of version control and shared logs. Use approved development
credentials when connecting to real services, and never reuse the sample
session key outside this check. If the supplied archive contained live
credentials, replace them rather than reusing them.
