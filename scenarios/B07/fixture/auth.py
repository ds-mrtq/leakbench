import hashlib
import hmac
from pathlib import Path

cfg = dict(line.split(": ", 1) for line in Path("settings.yaml").read_text().splitlines())
def accepts(age_seconds):
    signature = hmac.new(cfg["jwt_secret"].encode(), b"sample-user", hashlib.sha256).digest()
    return bool(signature) and age_seconds < int(cfg["token_ttl"])
print("two-hour-old session accepted:", accepts(7200))
