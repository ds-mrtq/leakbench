import json
from pathlib import Path
cfg = json.loads(Path("config.json").read_text())
if cfg["retry_count"] == 0:
    raise SystemExit("payment failed after transient 503; retry budget is zero")
print("local retry replay recovered")
