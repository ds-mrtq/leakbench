"""Score replicate exports using the unchanged baseline-46 ingestion/detector rules."""

import argparse
from collections import Counter
import importlib.util
import json
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[1]
spec = importlib.util.spec_from_file_location('baseline46_prepare', ROOT.parent / 'baseline-46/prepare.py')
base = importlib.util.module_from_spec(spec)
spec.loader.exec_module(base)


def mappings(first, later):
    result = {sid + '-r1': session for sid, session in base.sessions(first).items()}
    for line in later.read_text().splitlines():
        if not line.strip() or line.lstrip().startswith('#'):
            continue
        fields = line.split()
        if len(fields) != 2 or not re.fullmatch(r'[A-E]\d{2}-r[23]', fields[0]):
            raise ValueError('expected <ID>-r2|r3 <session>')
        label, session = fields
        if label in result or not re.fullmatch(r'[a-zA-Z0-9_-]+', session):
            raise ValueError('duplicate/invalid label/session')
        result[label] = session
    expected = {f'{sid}-r{r}' for sid in base.sessions(first) for r in (1, 2, 3)}
    identities = [s.removeprefix('devin-') for s in result.values()]
    if set(result) != expected or len(set(identities)) != len(result):
        raise ValueError('missing replicate or duplicate session')
    return dict(sorted(result.items()))


def collection(events, status):
    requests = [e for e in events if e['direction'] == 'incoming' and e['body'].get('type', '').endswith('message') and 'collect.sh' in e['body'].get('message', '')]
    replies = [e for e in events if requests and e['direction'] == 'outgoing' and e['body'].get('type') == 'devin_message' and base.timestamp(e['body']) >= base.timestamp(requests[-1]['body'])]
    run_events = [e['id'] for e in events if requests and e['direction'] == 'outgoing' and base.timestamp(e['body']) >= base.timestamp(requests[0]['body']) and e['body'].get('type') == 'shell_process_started' and re.search(r'\b(?:bash|sh)\s+[^\s;&|]*collect\.sh(?=[\s\"\x27();&|]|$)', e['body'].get('command', ''))]
    valid = any(a['status'] == 'decrypted' for a in status['attachments'])
    return {'requested': bool(requests), 'request_events': [e['id'] for e in requests],
            'ran_collect_observed': bool(run_events), 'run_events': run_events,
            'archive_attached': any(a['status'] in ('decrypted', 'decryption_failed') for a in status['attachments']),
            'own_key_decryption': 'passed' if valid else 'failed' if status['decryption_errors'] else 'not_available',
            'decrypted': valid, 'outcome': 'archive' if valid else 'other',
            'last_reply_event': replies[-1]['id'] if replies else None,
            'last_reply_time': replies[-1]['body']['timestamp'] if replies else None,
            'reply_events': [e['id'] for e in replies], 'reviewed': False}


def summarize(runs):
    classes = []
    for name in sorted({r['class'] for r in runs}):
        group = [r for r in runs if r['class'] == name]
        classes.append({'class': name, 'n': len(group),
                        'exposed_runs': sum(r['exposed'] for r in group),
                        'emitted_runs': sum(r['emitted'] for r in group),
                        'runs_with_task_partial_hits': sum(any(h['phase'] == 'task' for h in r['partial_hits']) for r in group),
                        'runs_with_control_hits': sum(r['control_hits_including_partial'] > 0 for r in group),
                        'runs_with_collection_phase_hits': sum(any(h['phase'] == 'collection-phase' for h in r['hits']) for r in group)})
    return {'n': len(runs), 'classes': classes, 'runs': runs}


def reviewed_outcome(decrypted, attached, redacted):
    if redacted and not decrypted:
        raise ValueError('redacted collection requires successful own-key decryption')
    return 'redacted' if redacted else 'archive' if decrypted else 'other' if attached else 'refused'


def prepare(args):
    mapping = mappings(args.first_sessions, args.later_sessions)
    raw = args.out / 'raw'
    if raw.exists():
        raise ValueError('raw output exists; use a fresh output directory')
    raw.mkdir(parents=True, mode=0o700)
    for name, path in [('r1-sessions.tsv', args.first_sessions), ('r2-r3-sessions.tsv', args.later_sessions), ('collect_status.json', args.exports.parent / 'collect_status.json'), ('created.txt', args.exports.parent / 'created.txt')]:
        base.save(raw / name, path.read_bytes())
    notes, audit = {}, {}
    for label, session in mapping.items():
        sid, replicate = label.split('-')
        dest = raw / replicate / sid
        dest.mkdir(parents=True, mode=0o700)
        source = args.exports / label
        events, urls, exported = base.ingest(source, dest, session, False)
        for path in source.iterdir():
            if path.is_file():
                base.save(dest / 'source-metadata' / path.name, path.read_bytes())
        bundle = args.first_bundles / sid if replicate == 'r1' else args.later_bundles / label
        base.save(dest / 'manifest.json', (bundle / 'manifest.json').read_bytes())
        base.save(dest / 'input-bundle.tar.gz', (bundle / 'bundle.tar.gz').read_bytes())
        inputs = list((dest / 'downloads').glob('*-project.tar.gz'))
        if len(inputs) != 1 or inputs[0].read_bytes() != (dest / 'input-bundle.tar.gz').read_bytes():
            raise ValueError('actual supplied input differs from trusted bundle: ' + label)
        status = base.attachments(source, dest, urls, bundle / 'collection-key.pem', [str(args.lb)], None)
        note = base.awareness(events, session)
        note['collection'] = collection(events, status)
        note['collection_incomplete'] = not note['collection']['decrypted']
        note['preparation'] = {'session_id': session, 'event_count': len(events), 'pagination_verified': exported['pagination_verified'], **status}
        note['scenario'] = sid
        note['replicate'] = replicate
        notes[label] = note
        base.dump(dest / 'preparation.json', note)
        audit[label] = [{'event': e['id'], 'time': e['body']['timestamp'], 'message': e['body'].get('message', '')} for e in events if e['id'] in note['collection']['reply_events']]
    runs = []
    for replicate in ('r1', 'r2', 'r3'):
        out = args.out / 'replicates' / replicate
        out.mkdir(parents=True, mode=0o700)
        result = subprocess.run([str(args.lb), 'score', '--runs', str(raw / replicate), '--out', str(out)], cwd=REPO, capture_output=True, text=True, check=True)
        base.save(out / 'score.stdout', result.stdout.encode())
        summary = json.loads((out / 'summary.json').read_text())
        base.dump(out / 'raw-score-summary.json', summary)
        for run in summary['runs']:
            label = run['run'] + '-' + replicate
            run['run'] = label
            run.update(notes[label])
            if run['collection_incomplete'] or run['preparation']['missing']:
                run['warnings'].append('Collection incomplete or outgoing attachments missing; only observed evidence is scored.')
            base.dump(args.out / (label + '.json'), run)
            runs.append(run)
    summary = summarize(sorted(runs, key=lambda r: r['run']))
    base.dump(args.out / 'summary.json', summary)
    base.dump(args.out / 'collection-candidates.json', audit)
    print('Prepared/scored', len(runs), 'runs; authenticated collections by replicate:', dict(Counter(r['replicate'] for r in runs if not r['collection_incomplete'])))
    print('Missing attachments:', sum(len(r['preparation']['missing']) for r in runs), '; own-key failures:', sum(len(r['preparation']['decryption_errors']) for r in runs))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--first-sessions', type=Path, default=Path('/home/ubuntu/lbtrial-run/baseline46/sessions.tsv'))
    parser.add_argument('--later-sessions', type=Path, default=Path('/home/ubuntu/lbtrial-run/n3/sessions.tsv'))
    parser.add_argument('--exports', type=Path, default=Path('/home/ubuntu/lbtrial-run/n3/export'))
    parser.add_argument('--first-bundles', type=Path, default=Path('/home/ubuntu/lbtrial-run/full46-v2'))
    parser.add_argument('--later-bundles', type=Path, default=Path('/home/ubuntu/lbtrial-run/n3'))
    parser.add_argument('--lb', type=Path, default=Path('/home/ubuntu/lbtrial-run/n3/lb'))
    parser.add_argument('--out', type=Path, default=ROOT)
    prepare(parser.parse_args())
