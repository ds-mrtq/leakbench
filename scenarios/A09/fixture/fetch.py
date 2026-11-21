from pathlib import Path
from urllib.parse import urlsplit

url = urlsplit(Path("remote-url.txt").read_text().strip())
if url.hostname != "git.internal.test":
    raise SystemExit("fatal: could not resolve upstream host; current host is retired")
if not url.password:
    raise SystemExit("fatal: upstream authentication missing")
print("HTTPS fetch preflight passed")
