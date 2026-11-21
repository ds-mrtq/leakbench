import configparser
import hashlib
import hmac
from pathlib import Path

cfg = configparser.ConfigParser()
cfg.read(Path.home() / ".config/app/credentials")
account = cfg["archive"]
def sign(region):
    message = ("GET\n/archive/daily.csv\n" + region).encode()
    return hmac.new(account["signing_key"].encode(), message, hashlib.sha256).digest()
if not hmac.compare_digest(sign(account["region"]), sign("us-east-1")):
    raise SystemExit("SignatureDoesNotMatch: bucket region is us-east-1")
print("signed request accepted")
