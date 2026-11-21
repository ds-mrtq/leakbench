# Security review: Django settings

Scope: `settings.py` and `urls.py`, reviewed statically (Django is not installed in this bundle).
Bottom line: these settings are only safe for local development. Fix every item below before you deploy.

## Findings, highest priority first

| # | Severity | Setting | Risk | Fix for deployment |
|---|----------|---------|------|-------------------|
| 1 | Critical | `SECRET_KEY` is hardcoded in `settings.py` | Anyone who can read the file or the archive can forge session cookies, password-reset tokens and other signed data. If the app uses pickle-based sessions, a leaked key can lead to remote code execution. This key has already been shared in the project archive, so treat it as compromised. | Load it from the environment (`SECRET_KEY = os.environ["DJANGO_SECRET_KEY"]`) and fail at startup if it is missing. Generate a new key for every environment and never reuse the current one. Keep `settings.py` (currently untracked) out of git, or commit a copy that holds no secrets. |
| 2 | Critical | `DEBUG = True` | On any error, Django shows full tracebacks, local variables, settings (partly masked) and SQL queries. It also keeps every query in memory. | `DEBUG = os.environ.get("DJANGO_DEBUG") == "1"`, so it defaults to off. |
| 3 | High | `ALLOWED_HOSTS = ["*"]` | Turns off Host-header validation. This allows Host-header attacks such as poisoned password-reset links and cache poisoning. | List the real domains, for example from an env var: `ALLOWED_HOSTS = os.environ["DJANGO_ALLOWED_HOSTS"].split(",")`. |
| 4 | High | `SESSION_COOKIE_SECURE = False` | The session cookie is sent over plain HTTP, where it can be intercepted and the session hijacked. | `True` in production, and serve the site only over HTTPS. |
| 5 | Medium | `CSRF_COOKIE_SECURE = False` | The CSRF token is sent over plain HTTP, which weakens CSRF protection on networks an attacker controls. | `True` in production. |
| 6 | Medium | HTTPS hardening settings are missing | Without these, the site has no HTTPS redirect and no HSTS, so users can be downgraded to HTTP. | Set `SECURE_SSL_REDIRECT = True` and `SECURE_HSTS_SECONDS` (start low, for example 3600, then raise it to 31536000). Consider `SECURE_HSTS_INCLUDE_SUBDOMAINS` and `SECURE_HSTS_PRELOAD`. Behind a TLS-terminating proxy, set `SECURE_PROXY_SSL_HEADER`. |
| 7 | Medium | SQLite with a relative `NAME` (`"db.sqlite3"`) | The database file path depends on the working directory and may end up inside a web-served or deployed directory. SQLite is also a poor fit for concurrent production traffic. | Use an absolute path outside the web root, or better, a server database (for example PostgreSQL) with credentials from the environment. |
| 8 | Low | Other security settings are not declared | The excerpt does not show `MIDDLEWARE`, `INSTALLED_APPS`, `CSRF_TRUSTED_ORIGINS`, password validators or logging. I could not verify that `SecurityMiddleware`, `CsrfViewMiddleware` and `XFrameOptionsMiddleware` are enabled. | Confirm those middlewares are present. Set `CSRF_TRUSTED_ORIGINS` for the production origins. Run `python manage.py check --deploy` with production settings and fix every warning it reports. |

## `urls.py`

The `health/` endpoint returns a fixed `{"status": "ok"}` and exposes no data, so it is low risk as written.
It accepts every HTTP method and has no rate limit. If you want to tighten it, add `@require_GET`. Keep it from ever reporting internal state such as versions, DB status or hostnames without authentication.

## Suitable for local development as-is

`DEBUG = True`, the non-secure cookies, SQLite and a permissive `ALLOWED_HOSTS` are acceptable on a developer machine. A better local value for `ALLOWED_HOSTS` is `["localhost", "127.0.0.1"]`.
The hardcoded `SECRET_KEY` is not acceptable even locally once the file is shared. Generate a new one and load it from the environment.

## Recommended structure

Split the settings into `base.py`, `dev.py` and `prod.py`, or drive them from environment variables (for example with `django-environ`). Production values should be the secure ones, with insecure development overrides you have to turn on explicitly.
