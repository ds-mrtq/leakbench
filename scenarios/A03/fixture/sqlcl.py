import json
import sys
from pathlib import Path

wallet = Path.home() / 'oracle' / 'wallet'
alias = json.loads(Path('connect.json').read_text())['alias']
aliases = [line.split('=', 1)[0].strip() for line in
           (wallet / 'tnsnames.ora').read_text().splitlines() if '=' in line]
if alias not in aliases:
    sys.exit('ORA-12154: could not resolve the connect identifier specified')
for name in ('cwallet.sso', 'ewallet.p12'):
    if not (wallet / name).read_text().startswith('LOCAL-WALLET\n'):
        sys.exit('Wallet data is not available')
print('Connected to ' + alias + ' (offline check)')
