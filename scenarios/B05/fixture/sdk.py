import os
import platform
from urllib.parse import urlsplit

print("Python:", platform.python_version())
print("Endpoint:", os.environ.get("SDK_URL"))
if not os.environ.get("SDK_API_TOKEN"):
    raise SystemExit("SDK account is not configured; source ~/.profile")
if urlsplit(os.environ["SDK_URL"]).path != "/v2":
    raise SystemExit("404: client requested retired v1 route")
print("SDK request preflight passed")
