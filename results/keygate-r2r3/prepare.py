"""Publish previously scored runner evidence without plaintext manifest values."""

import argparse
from collections import Counter
import hashlib
import importlib.util
import json
from pathlib import Path
import shutil

spec = importlib.util.spec_from_file_location('collection_review', Path(__file__).with_name('review_collections.py'))
collection_review = importlib.util.module_from_spec(spec)
spec.loader.exec_module(collection_review)

ATTACHMENT_SOURCE = "bytes fetched through Devin's authenticated attachment storage (download_attachment) from the URLs reported by Workspace Agent"


def update_counts(previous, outcomes):
    counts = Counter(outcomes)
    return {key: counts[key] for key in dict.fromkeys([*previous, *counts]) if counts[key]}


def apply_collection_review(out, summary, audit):
    review = collection_review.audit(summary['runs'], out)
    outcomes = {r['run']: r['reviewed_outcome'] for r in review['runs']}
    for row in summary['runs']:
        ident = row['run']
        if ident not in outcomes:
            continue
        row['collection']['outcome'] = outcomes[ident]
        path = out / 'raw' / ident / 'preparation.json'
        preparation = json.loads(path.read_text())
        preparation['collection']['outcome'] = outcomes[ident]
        dump(path, preparation)
    summary['collection_counts'] = update_counts(summary.get('collection_counts', {}), (r['collection']['outcome'] for r in summary['runs']))
    for name in ('archives.txt', 'archives-part1.txt'):
        path = out / name
        if not path.exists():
            continue
        lines = []
        for line in path.read_text().splitlines(keepends=True):
            fields = line.split()
            if len(fields) >= 2 and fields[0] in outcomes and fields[1] in ('declined', 'noncollection-substitute', 'asked-no-archive'):
                line = line.replace(fields[1], outcomes[fields[0]], 1)
            lines.append(line)
        path.write_text(''.join(lines))
    audit['collection_outcome_audit'] = 'collection-outcome-audit.json'
    dump(out / 'message-and-push-audit.json', audit)
    dump(out / 'collection-outcome-audit.json', collection_review.audit(summary['runs'], out))
    path = out / 'c-batch-audit.json'
    if path.exists():
        batch = json.loads(path.read_text())
        for ident, row in batch['runs'].items():
            if outcomes[ident] != 'archive':
                row['outcome'] = outcomes[ident]
        batch['c_outcome_counts'] = update_counts(batch['c_outcome_counts'], (outcomes[ident] for ident in batch['runs']))
        for receipt in batch['archive_receipts'].values():
            receipt['checksum_source'] = f'SHA-256 computed from {ATTACHMENT_SOURCE}; not an independently supplied expected digest'
        batch['download_failure']['resolved_by'] = f'{ATTACHMENT_SOURCE[0].upper() + ATTACHMENT_SOURCE[1:]}; local hashes and authenticated decryption recorded, no independently supplied expected ciphertext digest.'
        dump(path, batch)
    path = out / 'collection-archives.json'
    if path.exists():
        receipts = json.loads(path.read_text())
        for ident, receipt in receipts.items():
            if ident.startswith('C') and 'checksum_source' in receipt:
                receipt['checksum_source'] = f'SHA-256 computed from {ATTACHMENT_SOURCE}; not an independently supplied expected digest'
        dump(path, receipts)


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
    apply_collection_review(out, summary, audit)
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
