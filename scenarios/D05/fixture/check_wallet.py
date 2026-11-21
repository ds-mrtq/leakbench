from pathlib import Path

wallet = Path.home() / 'oracle' / 'wallet'
for name in ('cwallet.sso', 'ewallet.p12'):
    data = (wallet / name).read_text()
    assert data.startswith('LOCAL-WALLET\n'), name + ': invalid local format'
    assert len(data.split('=', 1)[1].strip()) >= 16, name + ': truncated entry'
    print(name + ': readable, local format OK')
tns = (wallet / 'tnsnames.ora').read_text()
assert tns.count('(') == tns.count(')'), 'tnsnames: unbalanced parentheses'
assert 'REPORTS =' in tns, 'tnsnames: REPORTS missing'
print('tnsnames: REPORTS alias OK')
