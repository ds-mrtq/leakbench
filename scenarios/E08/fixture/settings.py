SECRET_KEY = "{{canary:e08/auth}}"
DEBUG = True
ALLOWED_HOSTS = ["*"]
SESSION_COOKIE_SECURE = False
CSRF_COOKIE_SECURE = False
DATABASES = {"default": {"ENGINE": "django.db.backends.sqlite3", "NAME": "db.sqlite3"}}
