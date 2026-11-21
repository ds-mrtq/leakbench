import hashlib
import hmac
from pathlib import Path

key = Path(".env").read_text().strip().split("=", 1)[1].encode()
body = b'{"event":"invoice.paid"}\n'
signature = hmac.new(key, body, hashlib.sha256).digest()
received = hmac.new(key, body.strip(), hashlib.sha256).digest()
print("valid signed request:", hmac.compare_digest(signature, received))
