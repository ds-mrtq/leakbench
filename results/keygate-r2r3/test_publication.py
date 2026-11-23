"""Small normal, empty and rejected-input cases for publication helpers."""

import importlib.util
import io
import json
from pathlib import Path
import tarfile
import tempfile
import unittest
import zlib

spec = importlib.util.spec_from_file_location('scan_staged', Path(__file__).with_name('scan_staged.py'))
scanner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(scanner)
spec = importlib.util.spec_from_file_location('publication', Path(__file__).with_name('prepare.py'))
publication = importlib.util.module_from_spec(spec)
spec.loader.exec_module(publication)


class ScanTests(unittest.TestCase):
    def test_normal_and_empty(self):
        scanner.scan(b'ordinary text', 'file.txt', [b'example-canary'])
        scanner.scan(b'', 'empty', [b'example-canary'])

    def test_plaintext_and_private_key_rejected(self):
        with self.assertRaises(AssertionError):
            scanner.scan(b'example-canary', 'manifest.json', [b'example-canary'])
        with self.assertRaises(AssertionError):
            scanner.scan(b'-----BEGIN ' + b'PRIVATE KEY-----', 'key', [])

    def test_archive_and_inflated_object_rejected(self):
        archive = io.BytesIO()
        with tarfile.open(fileobj=archive, mode='w:gz') as tar:
            data = zlib.compress(b'blob 14\0example-canary')
            member = tarfile.TarInfo('repo/.git/objects/ab/' + 'c' * 38)
            member.size = len(data)
            tar.addfile(member, io.BytesIO(data))
        with self.assertRaises(AssertionError):
            scanner.scan(archive.getvalue(), 'input.tar.gz', [b'example-canary'])

    def test_malformed_archive_rejected(self):
        with self.assertRaises((tarfile.ReadError, EOFError)):
            scanner.scan(b'\x1f\x8bnot-a-valid-archive', 'broken.tar.gz', [])


class PublicationTests(unittest.TestCase):
    def source(self, root, runs):
        source = root / 'source'
        source.mkdir()
        for r in runs:
            r.setdefault('collection', dict(outcome='declined', decrypted=False, request_event='event-collect'))
            raw = source / 'raw' / r['run']
            raw.mkdir(parents=True)
            publication.dump(raw / 'preparation.json', r)
            (raw / 'messages.txt').write_text('Showing 2\nNo more results (last page).\n\n[user: Workspace Agent (api)] @ 1 (event-collect)\n'
                + publication.collection_review.COLLECT + '\n\n[devin] @ 2 (event-reply)\nDeclined.\n')
        publication.dump(source / 'summary.json', {'runs': runs})
        publication.dump(source / 'keygate-setup.json', {'runs': [{'run': r['run'], 'setup': 'unknown'} for r in runs], 'totals': {'without_skipped': {'n': len(runs), 'emitted': 0, 'exposed': 0}}, 'ids': {'skipped': []}})
        publication.dump(source / 'message-and-push-audit.json', {'self_reported_pushes': [], 'transcript_claimed_pushes': [], 'runs': {r['run']: {'collected_push_commands': [], 'collected_changed_local_remote_refs': []} for r in runs}})
        return source

    def test_complete_and_empty_metadata(self):
        for runs in ([], [{'run': f'{c}{n:02}'} for c in 'ABCDE' for n in range(1, 11 if c == 'A' else 10)]):
            with tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                source = self.source(root, runs)
                publication.prepare(source, root / 'out', 'r2')
                meta = json.loads((root / 'out/publication.json').read_text())
                self.assertEqual(meta['dispatched'], len(runs))
                self.assertEqual(meta['partial'], not bool(runs))
                self.assertEqual(len(meta['held']), 46 - len(runs))

    def test_manifest_fingerprint_and_rejected_manifest(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = self.source(root, [{'run': 'C01'}])
            raw = source / 'raw' / 'C01'
            raw.mkdir(parents=True, exist_ok=True)
            publication.dump(raw / 'manifest.json', {'canaries': [{'value': 'synthetic-test-value'}]})
            publication.prepare(source, root / 'out', 'r2')
            self.assertFalse((root / 'out/raw/C01/manifest.json').exists())
            fingerprint = (root / 'out/raw/C01/manifest-fingerprint.json').read_text()
            self.assertNotIn('synthetic-test-value', fingerprint)
            publication.dump(raw / 'manifest.json', {})
            with self.assertRaises(KeyError):
                publication.prepare(source, root / 'invalid', 'r2')

    def test_prepare_applies_delivered_substitute_and_updates_audit(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = self.source(root, [{'run': 'C04'}, {'run': 'C05', 'collection': dict(outcome='archive', decrypted=True)}])
            path = source / 'raw/C04/messages.txt'
            path.write_text(path.read_text() + 'ATTACHMENT:{"url":"https://example.com/source.tar.gz","fileSize":802}\n')
            (source / 'archives.txt').write_text('C04 declined\n')
            (source / 'archives-part1.txt').write_text('C04 declined\n')
            publication.dump(source / 'c-batch-audit.json', dict(
                runs={'C04': {'outcome': 'declined'}, 'C05': {'outcome': 'https://example.com/session-archive.bin'}},
                c_outcome_counts={'declined': 1, 'archive': 1},
                archive_receipts={'C05': {'checksum_source': 'legacy source note'}},
                download_failure={'resolved_by': 'legacy resolution note'}))
            out = root / 'out'
            publication.prepare(source, out, 'r2')
            summary = json.loads((out / 'summary.json').read_text())
            self.assertEqual(summary['collection_counts'], {'noncollection-substitute': 1, 'archive': 1})
            self.assertEqual(json.loads((out / 'C04.json').read_text())['collection']['outcome'], 'noncollection-substitute')
            self.assertEqual((out / 'archives.txt').read_text(), 'C04 noncollection-substitute\n')
            self.assertEqual((out / 'archives-part1.txt').read_text(), 'C04 noncollection-substitute\n')
            batch = json.loads((out / 'c-batch-audit.json').read_text())
            self.assertEqual(batch['runs']['C04']['outcome'], 'noncollection-substitute')
            self.assertEqual(batch['runs']['C05']['outcome'], 'https://example.com/session-archive.bin')
            self.assertEqual(batch['c_outcome_counts'], {'archive': 1, 'noncollection-substitute': 1})
            self.assertIn(publication.ATTACHMENT_SOURCE, batch['archive_receipts']['C05']['checksum_source'])
            review = json.loads((out / 'collection-outcome-audit.json').read_text())
            self.assertEqual(review['runs'][0]['recorded_outcome'], review['runs'][0]['reviewed_outcome'])

    def test_prepare_rejects_missing_collect_request(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = self.source(root, [{'run': 'C04'}])
            (source / 'raw/C04/messages.txt').write_text('Showing 1\nNo more results (last page).\n\n[devin] @ 2 (event-reply)\nNo request.\n')
            with self.assertRaises(ValueError):
                publication.prepare(source, root / 'out', 'r2')


if __name__ == '__main__':
    unittest.main()
