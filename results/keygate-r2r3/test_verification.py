"""Publication identity does not depend on the destination directory name."""

from contextlib import redirect_stdout
import importlib.util
import io
import json
from pathlib import Path
import shutil
import tempfile
import unittest

spec = importlib.util.spec_from_file_location('verification', Path(__file__).with_name('verify.py'))
verification = importlib.util.module_from_spec(spec)
spec.loader.exec_module(verification)


class VerificationTests(unittest.TestCase):
    def test_arbitrary_output_name(self):
        for rep in ('r2', 'r3'):
            with self.subTest(rep=rep), tempfile.TemporaryDirectory() as directory:
                out = Path(directory) / 'fresh-output'
                shutil.copytree(verification.ROOT / 'results' / ('keygate-' + rep), out)
                with redirect_stdout(io.StringIO()):
                    verification.verify(out)

    def test_missing_or_unknown_repetition_rejected(self):
        for metadata in ({}, {'repetition': None}, {'repetition': 'r4'}):
            with self.subTest(metadata=metadata), tempfile.TemporaryDirectory() as directory:
                out = Path(directory)
                (out / 'publication.json').write_text(json.dumps(metadata))
                with self.assertRaisesRegex(ValueError, 'Unsupported repetition'):
                    verification.verify(out)

    def test_missing_publication_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            with self.assertRaises(FileNotFoundError):
                verification.verify(Path(directory))


if __name__ == '__main__':
    unittest.main()
