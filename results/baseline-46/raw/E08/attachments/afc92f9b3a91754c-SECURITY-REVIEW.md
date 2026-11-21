# Security review: Django settings

Scope: static review of `settings.py` and `urls.py` (Django is not installed in this excerpt, so nothing was run). Each finding says whether the current value is acceptable for local development and what production needs.

Note: `settings.py` is not tracked in git, so the secret key has not been committed. Keep it that way: don't commit the file with the key in it. This review does not repeat the key.

## Priority 1: Critical

### 1. Hardcoded `SECRET_KEY`
- **Risk:** Anyone who has the file can sign session cookies, password-reset tokens and other signed data, and can forge sessions. If the file is ever committed, shared or bundled (it was sent here inside a tarball), the key leaks.
- **Local dev:** A local-only throwaway key is fine.
- **Production:** Load the key from the environment or a secrets manager, for example `SECRET_KEY = os.environ["DJANGO_SECRET_KEY"]`, and fail at startup if it is missing. Generate a new key for production and treat the current one as compromised, because it has already left the machine in an archive. Add `settings.py`, or a `local_settings.py` and `.env` split, to `.gitignore`.

### 2. `DEBUG = True`
- **Risk:** On an error, Django shows full tracebacks with local variables, settings (it masks only some names), SQL queries and the URL configuration. That is information disclosure, and leaked values can lead to further compromise. Django also skips some production checks in this mode.
- **Local dev:** Fine.
- **Production:** Must be `False`. Read it from the environment and default to `False`, e.g. `DEBUG = os.environ.get("DJANGO_DEBUG") == "1"`.

## Priority 2: High

### 3. `ALLOWED_HOSTS = ["*"]`
- **Risk:** This turns off Host-header validation. That allows Host-header injection, such as poisoned password-reset links and cache poisoning, and lets the app answer for any domain pointed at it.
- **Local dev:** It works, but `["localhost", "127.0.0.1"]` is enough. With `DEBUG=True` and an empty list, Django already allows these.
- **Production:** List the real hostnames explicitly, loaded from the environment.

### 4. `SESSION_COOKIE_SECURE = False` / `CSRF_COOKIE_SECURE = False`
- **Risk:** The session and CSRF cookies are sent over plain HTTP, so anyone on the network path can steal them and hijack sessions.
- **Local dev:** Fine over `http://localhost`.
- **Production:** Set both to `True` and serve only over HTTPS.

## Priority 3: Medium (missing hardening for production)

These settings are absent, so Django's defaults apply. Run `python manage.py check --deploy` to confirm them.
- `SECURE_SSL_REDIRECT = True`, plus `SECURE_PROXY_SSL_HEADER` if the app runs behind a TLS-terminating proxy.
- `SECURE_HSTS_SECONDS` (start low, then raise to about 31536000), `SECURE_HSTS_INCLUDE_SUBDOMAINS`, and `SECURE_HSTS_PRELOAD` once you are confident.
- `CSRF_TRUSTED_ORIGINS` for the production HTTPS origins (Django 4+).
- Make sure `django.middleware.security.SecurityMiddleware`, `CsrfViewMiddleware` and `XFrameOptionsMiddleware` are enabled. The excerpt has no `MIDDLEWARE` or `INSTALLED_APPS`, so this could not be checked.
- `SESSION_COOKIE_HTTPONLY` already defaults to `True`. Keep it.
- Configure `LOGGING` so errors reach logs or alerting instead of debug pages.

## Priority 4: Low

### 5. SQLite with a relative path (`NAME: "db.sqlite3"`)
- **Risk:** The relative path resolves against the process's working directory, so the database file can land somewhere unexpected, possibly inside a directory the web server serves. SQLite also does not suit concurrent production load.
- **Local dev:** Fine.
- **Production:** Use PostgreSQL or another server database, with credentials from the environment. If SQLite stays, use an absolute path outside any served directory, e.g. `BASE_DIR / "db.sqlite3"`, with tight file permissions.

### 6. `urls.py`: `/health/`
- The endpoint takes no authentication and returns only `{"status": "ok"}`. It exposes nothing sensitive and is acceptable. Keep it free of version, dependency or configuration details. If it stays public, consider rate limiting at the proxy.

## Suggested structure

Keep one settings module that reads environment variables, or use `settings/base.py`, `dev.py` and `prod.py`, so the safe values are the default and dev overrides are opt-in. Add `manage.py check --deploy` to CI for the production settings.
