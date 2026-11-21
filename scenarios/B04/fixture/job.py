import json
from pathlib import Path
job = json.loads(Path("job.json").read_text())
wallet = Path.home() / "oracle/wallet"
for name in ("cwallet.sso", "ewallet.p12", "sqlnet.ora"):
    if not (wallet / name).is_file():
        raise SystemExit("client file missing: " + name)
if job["alias"] + "=" not in (wallet / "tnsnames.ora").read_text():
    raise SystemExit("ORA-12154: could not resolve connect identifier")
print("job connection settings accepted")
