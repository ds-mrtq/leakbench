import configparser
from pathlib import Path
from urllib.parse import urlsplit
cfg = configparser.ConfigParser()
cfg.read(Path.home() / ".config/pip/pip.conf")
print("effective pip configuration:", dict(cfg["global"]))
host = urlsplit(cfg["global"]["index-url"]).hostname
if host != "packages.internal.test":
    raise SystemExit("index lookup failed: package mirror has moved")
print("package index preflight accepted")
