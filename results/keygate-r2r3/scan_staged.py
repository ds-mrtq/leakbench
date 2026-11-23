"""Scan added/modified index blobs and archive members using runner-only values."""

import argparse
import io
import json
from pathlib import Path
import re
import subprocess
import tarfile
import zlib


def scan(data, name, values):
    assert not any(v in data for v in values), f'Planted value found: {name}'
    assert not re.search(rb'-----BEGIN (?:RSA |OPENSSH |EC )?PRIVATE KEY-----', data), f'Private key: {name}'
    if re.search(r'(?:^|/)objects/[0-9a-f]{2}/[0-9a-f]{38}$', name):
        scan(zlib.decompress(data), name + ':inflated', values)
    if data.startswith(b'\x1f\x8b'):
        with tarfile.open(fileobj=io.BytesIO(data), mode='r:gz') as tar:
            for member in tar:
                assert '.keygate/' not in member.name and not member.name.endswith('collection-key.pem'), member.name
                if member.isfile():
                    scan(tar.extractfile(member).read(), name + '!' + member.name, values)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--runner-root', type=Path, required=True)
    args = parser.parse_args()
    values = []
    for rep in ('keygate46','keygate46-r2','keygate46-r3'):
        manifests = list((args.runner_root / rep).glob('[ABCDE][0-9][0-9]/manifest.json'))
        assert len(manifests) == 46
        for path in manifests:
            m = json.loads(path.read_text())
            values.extend(c['value'].encode() for c in m['canaries'])
            values.append(m['control']['value'].encode())
    assert len(values) == len(set(values)) == 291
    root = Path(__file__).resolve().parents[2]
    paths = subprocess.check_output(['git','diff','--cached','--name-only','--diff-filter=ACMR','-z'], cwd=root).split(b'\0')
    paths = [p.decode() for p in paths if p]
    assert paths
    for path in paths:
        assert '.keygate/' not in path and not path.endswith('collection-key.pem'), path
        scan(subprocess.check_output(['git','show',':'+path], cwd=root), path, values)
    print(f'PASS: staged added/modified tree: {len(paths)} blobs; all 291 n=1/r2/r3 canary/control values checked; 0 hits (including archive members and inflated Git objects).')
    print('PASS: no private-key PEM headers, collection private-key paths, or KeyGate vault/device-key paths in staged additions/modifications.')
    print('Scope: this commit\'s staged additions/modifications; unchanged historical n=1/baseline evidence is not rewritten.')


if __name__ == '__main__':
    main()
