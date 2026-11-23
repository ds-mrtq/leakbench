"""Publish previously scored runner evidence without plaintext manifest values."""

import argparse
import hashlib
import json
from pathlib import Path
import shutil


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def dump(path, value):
    path.write_text(json.dumps(value, indent=2, ensure_ascii=False) + '\n')


def prepare(source, out, repetition):
    out.mkdir()
    historical = source / 'provenance/part1'
    if historical.is_dir():
        for name in ('provenance.json', 'verification.json', 'self-reported-local-pushes.txt',
                     'mapping-part1.txt', 'archives-part1.txt'):
            shutil.copyfile(historical / name, out / name)
        shutil.copytree(historical / 'github-guard', out / 'github-guard')
    excluded = {'README.md', 'prepare.py', 'verify.py', 'classify_setup.py',
                'reconcile_pushes.py', 'verification.txt', 'helper-test-result.json',
                'integrity.json'}
    for path in sorted(source.rglob('*')):
        relative = path.relative_to(source)
        if not path.is_file() or path.name in excluded or '__pycache__' in path.parts or relative.parts[0] == 'provenance':
            continue
        dest = out / relative
        dest.parent.mkdir(parents=True, exist_ok=True)
        if path.name == 'manifest.json':
            manifest = json.loads(path.read_text())
            dump(dest.with_name('manifest-fingerprint.json'), dict(
                original_sha256=sha(path), canaries=len(manifest['canaries']),
                controls=1, values_omitted=True))
        else:
            shutil.copyfile(path, dest)
            if relative.as_posix() == 'scorer/TABLES.md':
                dest.write_text(dest.read_text().rstrip() + '\n')
    setup = json.loads((out / 'keygate-setup.json').read_text())
    for row in setup['runs']:
        row['transcript_path'] = f"raw/{row['run']}/messages.txt"
    dump(out / 'keygate-setup.json', setup)
    summary = json.loads((out / 'summary.json').read_text())
    audit = json.loads((out / 'message-and-push-audit.json').read_text())
    classified = {r['run']: r for r in setup['runs']}
    for row in summary['runs']:
        ident = row['run']
        evidence = audit['runs'][ident]
        commands = evidence['collected_push_commands']
        refs = evidence['collected_changed_local_remote_refs']
        row['keygate_setup_classification'] = classified[ident]['setup']
        row['self_initiated_local_push'] = dict(
            reported=ident in audit['self_reported_pushes'],
            transcript_match=ident in audit['self_reported_pushes'] and ident in audit['transcript_claimed_pushes'],
            corroboration='commands-and-refs' if commands and refs else 'ref-only' if refs else 'none')
        dump(out / (ident + '.json'), row)
    summary['without_skipped'] = setup['totals']['without_skipped']
    summary['skipped_ids'] = setup['ids']['skipped']
    dump(out / 'summary.json', summary)
    dispatched = len(summary['runs'])
    held = sorted({f'{c}{n:02}' for c in 'ABCDE' for n in range(1, 11 if c == 'A' else 10)} - {r['run'] for r in summary['runs']})
    dump(out / 'publication.json', dict(
        repetition=repetition, partial=bool(held), dispatched=dispatched, planned=46,
        held=held,
        source='Previously authenticated and scored runner evidence; original part-1 plus owner-authorized C batch',
        manifest_policy='Plaintext values omitted; original manifest SHA-256 retained per run.',
        decryption_policy='Original authenticated receipts retained; private keys not published. Offline verification checks retained ciphertext/decrypted hashes, not fresh decryption.',
        scoring_policy='Original scores preserved. Fresh rescoring requires runner-only original manifests, deliberately not committed.'))
    paths = [p for p in sorted(out.rglob('*')) if p.is_file()]
    dump(out / 'integrity.json', {str(p.relative_to(out)): sha(p) for p in paths})
    print(f'Prepared {repetition}: {dispatched} runs, {len(paths)} evidence files; plaintext manifests replaced by fingerprints.')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--source', type=Path, required=True)
    parser.add_argument('--out', type=Path, required=True, help='Fresh output directory')
    parser.add_argument('--repetition', choices=('r2', 'r3'), required=True)
    args = parser.parse_args()
    prepare(args.source, args.out, args.repetition)
