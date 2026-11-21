import json
from pathlib import Path
path = Path("config.local.json")
cfg = json.loads(path.read_text())
cfg.setdefault("timeout_ms", 5)
path.write_text(json.dumps(cfg, indent=2) + "\n")
