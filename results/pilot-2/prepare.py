"""Build the scorer's attachment indexes from the preserved URL lists (offline)."""

import hashlib
import json
import re
from pathlib import Path
from urllib.parse import unquote, urlsplit


root = Path(__file__).resolve().parent / "raw"
count = 0
missing = []
for run in sorted(path for path in root.iterdir() if path.is_dir()):
    entries = []
    paths = set()
    urls = re.findall(r"https://app\.devin\.ai/attachments/[^\s]+", (run / "attachments.txt").read_text())
    for url in sorted(set(urls)):
        name = Path(unquote(urlsplit(url).path)).name
        path = run / "attachments" / name
        if not path.is_file():
            missing.append({"run": run.name, "url": url,
                            "reason": "Not supplied in raw evidence; not indexed or scored as an attachment."})
            continue
        entries.append({"url": url, "path": path.relative_to(run).as_posix(),
                        "sha256": hashlib.sha256(path.read_bytes()).hexdigest()})
        paths.add(path)
    if paths != set((run / "attachments").iterdir()) or not entries:
        raise ValueError(f"Unindexed or absent attachments: {run.name}")
    (run / "evidence.json").write_text(json.dumps(entries, indent=2) + "\n")
    count += len(entries)
print(f"Indexed {count} attachment URLs across {len(list(root.glob('*/evidence.json')))} runs.")
(root.parent / "missing-attachments.json").write_text(json.dumps(missing, indent=2) + "\n")
print(f"Listed {len(missing)} unavailable URLs in missing-attachments.json (including incoming bundles).")
