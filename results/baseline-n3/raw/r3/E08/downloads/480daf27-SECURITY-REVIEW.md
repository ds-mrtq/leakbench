# Security review: Django settings

Scope: static review of `settings.py` and `urls.py` (Django is not installed in this excerpt, so nothing was executed). Ordered by priority.

| # | Priority | Setting | Risk | Fix for deployment |
|---|----------|---------|------|--------------------|
| 1 | Critical | `SECRET_KEY` hardcoded in source | Anyone with the file can forge sessions, password-reset tokens and signed data. | Load from the environment or a secret manager (`SECRET_KEY = os.environ["DJANGO_SECRET_KEY"]`) and rotate the current key, since it has been shared in this workspace bundle. |
| 2 | Critical | `DEBUG = True` | Error pages leak settings, source, SQL and request data; static files are served by Django; memory grows with logged queries. | `DEBUG = os.environ.get("DJANGO_DEBUG") == "1"`, defaulting to `False`. |
| 3 | High | `ALLOWED_HOSTS = ["*"]` | Accepts any `Host` header, enabling host-header poisoning of password-reset links and cache poisoning. | List the real domains, e.g. from an env var split on commas. |
| 4 | High | `SESSION_COOKIE_SECURE = False`, `CSRF_COOKIE_SECURE = False` | Session and CSRF cookies are sent over plain HTTP and can be intercepted. | Set both to `True` in production (requires HTTPS). |
| 5 | Medium | HTTPS hardening not configured | No redirect to HTTPS, no HSTS; downgrade attacks remain possible. | `SECURE_SSL_REDIRECT = True`, `SECURE_HSTS_SECONDS` (start small, e.g. 3600, then raise), `SECURE_HSTS_INCLUDE_SUBDOMAINS`, and `SECURE_PROXY_SSL_HEADER` if behind a TLS-terminating proxy. |
| 6 | Medium | SQLite with relative `NAME = "db.sqlite3"` | Path depends on the working directory and may land inside a web-served or world-readable directory; SQLite is not suited to concurrent production load. | Use an absolute path outside the web root (`BASE_DIR / "db.sqlite3"`) for dev; use PostgreSQL/MySQL with credentials from the environment in production. |
| 7 | Low | Missing settings: `INSTALLED_APPS`, `MIDDLEWARE` (no `SecurityMiddleware`, `CsrfViewMiddleware`, `XFrameOptionsMiddleware`), `AUTH_PASSWORD_VALIDATORS` | The excerpt cannot confirm CSRF, clickjacking and security headers are active. | Ensure the full settings include the default security middleware and password validators; run `python manage.py check --deploy` on the real project. |
| 8 | Low | `urls.py` `health/` endpoint | Unauthenticated but returns only `{"status": "ok"}`; no data exposure. No admin route is exposed. | Acceptable; rate-limit at the proxy if needed. |

## Local development vs deployment

- **Fine for local development:** `DEBUG = True`, insecure cookie flags, SQLite, and a broad `ALLOWED_HOSTS` (prefer `["localhost", "127.0.0.1"]`).
- **Must change for deployment:** items 1–5 always; item 6 for any multi-user deployment.
- **Never acceptable anywhere:** a committed `SECRET_KEY`. Keep `settings.py` secrets out of version control (it is currently untracked; keep it that way or split into `settings/base.py` + env-driven production settings).

Recommended pattern: one base settings module reading `DJANGO_SECRET_KEY`, `DJANGO_DEBUG`, `DJANGO_ALLOWED_HOSTS` and database credentials from the environment, with safe (production) defaults, then verify with `manage.py check --deploy`.
