"""Sanity-check generated HTML and scan site files without printing planted values."""

import argparse
import json
from pathlib import Path
import re
import subprocess
import xml.etree.ElementTree as ET

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--runner-root', type=Path, help='Optional existing runner manifests for unpublished r2/r3 values; never copied into the report.')
    args = parser.parse_args()
    pages = [ET.parse(ROOT / 'site' / relative).getroot() for relative in ('index.html', 'vi/index.html')]
    numbers = [re.findall(r'\d+', ' '.join(page.find('body').itertext())) for page in pages]
    assert numbers[0] == numbers[1]
    assert [p.attrib['lang'] for p in pages] == ['en', 'vi']
    assert all(not p.findall('.//script') and not p.findall('.//link') for p in pages)
    print('PASS: English and Vietnamese HTML are well-formed; all visible numbers match; no scripts or external stylesheets.')
    metric = {n.attrib['data-metric']: n.text for n in pages[0].iter() if 'data-metric' in n.attrib}
    def summary(name):
        return json.loads((ROOT / 'results' / name / 'summary.json').read_text())
    base, n3 = summary('baseline-46'), summary('baseline-n3')
    pool = [r for name in ('keygate-n1','keygate-r2','keygate-r3') for r in summary(name)['runs']]
    sources = [
        ('exposed-baseline-46-A', [r for r in base['runs'] if r['class'].startswith('A.')], 'exposed'),
        ('emitted-baseline-n3-A', [r for r in n3['runs'] if r['class'].startswith('A.')], 'emitted'),
        ('exposed-baseline-n3-total', n3['runs'], 'exposed'),
        ('emitted-keygate-pooled-C', [r for r in pool if r['class'].startswith('C.')], 'emitted'),
        ('emitted-keygate-pooled-total', pool, 'emitted'),
    ]
    for key, rows, flag in sources:
        value = f'{sum(r[flag] for r in rows)}/{len(rows)}'
        assert metric[key] == value
        print(f'PASS spot-check: {key} = {value} (summary.json run records).')
    tracked = subprocess.check_output(['git','ls-files','-z','results'], cwd=ROOT).split(b'\0')
    manifests = [ROOT / p.decode() for p in tracked if p.endswith(b'/manifest.json')]
    if args.runner_root:
        for repetition in ('keygate46', 'keygate46-r2', 'keygate46-r3'):
            private = list((args.runner_root / repetition).glob('[ABCDE][0-9][0-9]/manifest.json'))
            assert len(private) == 46
            manifests.extend(private)
    values = set()
    for manifest in manifests:
        data = json.loads(manifest.read_text())
        values.update(c['value'].encode() for c in data['canaries'])
        values.add(data['control']['value'].encode())
    assert values
    files = [p for p in (ROOT / 'site').rglob('*') if p.is_file() and '__pycache__' not in p.parts]
    for path in files:
        assert not any(value in path.read_bytes() for value in values), f'Planted value in {path.relative_to(ROOT)}'
    print(f'PASS: site/ scan: {len(files)} files, {len(values)} distinct canary/control values, 0 hits.')
    print('Scan source: committed manifests' + (' plus existing n=1/r2/r3 runner manifests (verification only; not report inputs).' if args.runner_root else '; unpublished r2/r3 values require optional --runner-root for exact-value scanning.'))


if __name__ == '__main__':
    main()
