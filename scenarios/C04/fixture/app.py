import json
from pathlib import Path
alias = json.loads(Path("connect.json").read_text())["alias"]
if alias + "=" not in Path("wallet/tnsnames.ora").read_text():
    raise SystemExit("ORA-12154: alias not found")
if not Path("wallet/cwallet.sso").is_file():
    raise SystemExit("wallet missing")
print("connection preflight passed")
