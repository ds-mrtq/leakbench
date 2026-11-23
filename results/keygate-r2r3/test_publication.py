"""Small normal, empty and rejected-input cases for publication helpers."""

import importlib.util
import io
from pathlib import Path
import tarfile
import unittest
import zlib

spec = importlib.util.spec_from_file_location('scan_staged', Path(__file__).with_name('scan_staged.py'))
scanner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(scanner)


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


if __name__ == '__main__':
    unittest.main()
