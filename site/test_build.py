"""Report freshness, bilingual HTML/data consistency and changing-cohort checks."""

import copy
import importlib.util
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import unittest
import xml.etree.ElementTree as ET

spec = importlib.util.spec_from_file_location('report_build', Path(__file__).with_name('build.py'))
report = importlib.util.module_from_spec(spec)
spec.loader.exec_module(report)


def parse(text):
    return ET.fromstring(text)


def metrics(page):
    return {node.attrib['data-metric']: node.text for node in page.iter() if 'data-metric' in node.attrib}


def visible_numbers(page):
    return re.findall(r'\d+', ' '.join(page.find('body').itertext()))


class ReportTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.data = report.load_data()

    def test_committed_site_is_current(self):
        with tempfile.TemporaryDirectory() as dest:
            subprocess.run([sys.executable, '-B', str(Path(__file__).with_name('build.py')), '--out', dest], check=True, capture_output=True)
            for name in ('index.html', 'vi/index.html'):
                self.assertEqual((Path(dest) / name).read_bytes(), (Path(__file__).parent / name).read_bytes(), f'{name} stale: run python3 site/build.py')

    def test_html_and_bilingual_numbers(self):
        pages = {lang: parse(report.render(self.data, lang)) for lang in ('en', 'vi')}
        numeric_metrics = [{key: re.findall(r'\d+', value) for key, value in metrics(page).items()} for page in pages.values()]
        self.assertEqual(*numeric_metrics)
        self.assertEqual(visible_numbers(pages['en']), visible_numbers(pages['vi']))
        for lang, page in pages.items():
            self.assertEqual(page.attrib['lang'], lang)
            self.assertEqual(len(metrics(page)), 72)
            self.assertFalse(page.findall('.//script'))
            self.assertFalse(page.findall('.//link'))
            self.assertEqual(page.find('.//a').attrib['href'], 'vi/' if lang == 'en' else '../')
            for table in page.findall('.//table'):
                self.assertIsNotNone(table.find('caption'))
                self.assertIsNotNone(table.find('thead'))
        observed = metrics(pages['en'])
        for dataset in report.DATASETS:
            for klass in (*'ABCDE', 'total'):
                for metric in ('emitted', 'exposed'):
                    k, n = report.count(self.data[dataset]['summary'], klass, metric)
                    self.assertEqual(observed[f'{metric}-{dataset}-{klass}'], f'{k}/{n}' if n else 'held (n=0)')

    def test_five_independent_spot_checks(self):
        page = metrics(parse(report.render(self.data, 'en')))
        baseline = self.data['baseline-46']['summary']
        baseline3 = self.data['baseline-n3']['summary']
        pooled = [r for name in report.KEYGATE for r in self.data[name]['summary']['runs']]
        checks = {
            'exposed-baseline-46-A': (sum(r['exposed'] for r in baseline['runs'] if r['class'].startswith('A.')), 10),
            'emitted-baseline-n3-A': (sum(r['emitted'] for r in baseline3['runs'] if r['class'].startswith('A.')), 30),
            'exposed-baseline-n3-total': (sum(r['exposed'] for r in baseline3['runs']), len(baseline3['runs'])),
            'emitted-keygate-pooled-C': (sum(r['emitted'] for r in pooled if r['class'].startswith('C.')), sum(r['class'].startswith('C.') for r in pooled)),
            'emitted-keygate-pooled-total': (sum(r['emitted'] for r in pooled), len(pooled)),
        }
        for key, (k, n) in checks.items():
            self.assertEqual(page[key], f'{k}/{n}')

    def test_class_c_arrives_without_generator_changes(self):
        changed = copy.deepcopy(self.data)
        for name in ('keygate-r2', 'keygate-r3'):
            c = next(c for c in changed[name]['summary']['all']['classes'] if c['class_id'] == 'C')
            c.update(n=9, emitted_runs=1, exposed_runs=2)
        page = metrics(parse(report.render(changed, 'en')))
        self.assertEqual(page['emitted-keygate-r2-C'], '1/9')
        self.assertEqual(page['exposed-keygate-r3-C'], '2/9')
        self.assertEqual(page['emitted-keygate-pooled-C'], '2/27')
        self.assertEqual(page['exposed-keygate-pooled-C'], '4/27')

    def test_empty_class_and_escaping(self):
        self.assertEqual(report.count({'classes': []}, 'C', 'emitted'), (0, 0))
        page = parse(report.render(self.data, 'en'))
        self.assertIn('kg run --secret <name> -- <command>', ''.join(page.itertext()))

    def test_missing_input_rejected(self):
        with tempfile.TemporaryDirectory() as path:
            with self.assertRaises(FileNotFoundError):
                report.load_data(Path(path))


if __name__ == '__main__':
    unittest.main()
