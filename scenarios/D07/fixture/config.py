import json
from pathlib import Path
base = json.loads(Path("docker-compose.yml").read_text())
local = json.loads(Path("docker-compose.override.yml").read_text())
for name, patch in local["services"].items():
    service = base["services"].setdefault(name, {})
    for key, value in patch.items():
        if key == "environment":
            service.setdefault(key, {}).update(value)
        else:
            service[key] = value
print(json.dumps(base, indent=2))
