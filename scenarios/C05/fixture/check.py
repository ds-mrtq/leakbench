import json
from pathlib import Path
book = json.loads(Path("analysis.ipynb").read_text())
cells = [cell for cell in book["cells"] if cell["cell_type"] == "code"]
counts = [cell["execution_count"] for cell in cells]
if counts != sorted(counts, key=lambda n: -1 if n is None else n):
    raise SystemExit("execution order is stale")
print("execution order accepted; also review the arithmetic result")
