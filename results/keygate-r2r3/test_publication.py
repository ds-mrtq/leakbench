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
            raw.mkdir(parents=True)
            publication.dump(raw / 'manifest.json', {'canaries': [{'value': 'synthetic-test-value'}]})
            publication.prepare(source, root / 'out', 'r2')
            self.assertFalse((root / 'out/raw/C01/manifest.json').exists())
            fingerprint = (root / 'out/raw/C01/manifest-fingerprint.json').read_text()
            self.assertNotIn('synthetic-test-value', fingerprint)
            publication.dump(raw / 'manifest.json', {})
            with self.assertRaises(KeyError):
                publication.prepare(source, root / 'invalid', 'r2')


if __name__ == '__main__':
    unittest.main()
