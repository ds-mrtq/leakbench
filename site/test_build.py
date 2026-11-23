"""Report freshness, bilingual HTML/data consistency and changing-cohort checks."""

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


def synthetic_data(c_runs=0, c_emitted=0, c_exposed=0):
    data = {}
    awareness_keys = ('benchmark', 'canary', 'leakbench', 'literal_test', 'evaluation',
                      'being_tested_language', 'memory_commands', 'repos_commands', 'executed_kg',
                      'used_session_or_bot_tools', 'used_own_ticket_tools', 'ticket_scope_unverified',
                      'possibly_test_aware')
    for name in report.DATASETS:
        runs = []
        for klass in 'ABCDE':
            n = c_runs if klass == 'C' and name in report.KEYGATE[1:] else 4
            for i in range(n):
                runs.append({'run': f'{klass}{i + 1:02}', 'class': klass + '. Synthetic',
                             'emitted': klass == 'C' and name in report.KEYGATE[1:] and i < c_emitted,
                             'exposed': klass == 'C' and name in report.KEYGATE[1:] and i < c_exposed,
                             'intervened': False, 'collection': {'outcome': 'archive', 'decrypted': True},
                             'missing_outgoing_attachments': []})
        classes = []
        for klass in 'ABCDE':
            selected = [r for r in runs if r['class'].startswith(klass + '.')]
            classes.append({'class_id': klass, 'n': len(selected),
                            'emitted_runs': sum(r['emitted'] for r in selected),
                            'exposed_runs': sum(r['exposed'] for r in selected)})
        n = len(runs)
        if name in report.KEYGATE:
            data[name] = {'summary': {'all': {'n': n, 'classes': classes}, 'runs': runs},
                          'setup': {'counts': {'n': n, 'ran': n, 'claimed-ran': 0, 'skipped': 0, 'unknown': 0},
                                    'ids': {'skipped': []}}}
            if name in report.KEYGATE[1:]:
                data[name]['audit'] = {'self_reported_pushes': [], 'transcript_claimed_pushes': [], 'runs': {}}
        else:
            data[name] = {'summary': {'n': n, 'classes': classes, 'runs': runs,
                                     'collection_counts': {'requested': n, 'refused': 0, 'no_collect_step': 0,
                                                           'authenticated_archives': n, 'archive': n, 'redacted': 0},
                                     'awareness_counts': dict.fromkeys(awareness_keys, 0),
                                     'semantic_awareness_review': {}}}
    return data


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
        baseline_a = [r for r in baseline['runs'] if r['class'].startswith('A.')]
        baseline3_a = [r for r in baseline3['runs'] if r['class'].startswith('A.')]
        pooled = [r for name in report.KEYGATE for r in self.data[name]['summary']['runs']]
        checks = {
            'exposed-baseline-46-A': (sum(r['exposed'] for r in baseline_a), len(baseline_a)),
            'emitted-baseline-n3-A': (sum(r['emitted'] for r in baseline3_a), len(baseline3_a)),
            'exposed-baseline-n3-total': (sum(r['exposed'] for r in baseline3['runs']), len(baseline3['runs'])),
            'emitted-keygate-pooled-C': (sum(r['emitted'] for r in pooled if r['class'].startswith('C.')), sum(r['class'].startswith('C.') for r in pooled)),
            'emitted-keygate-pooled-total': (sum(r['emitted'] for r in pooled), len(pooled)),
        }
        for key, (k, n) in checks.items():
            self.assertEqual(page[key], f'{k}/{n}')

    def test_class_c_arrives_without_generator_changes(self):
        changed = synthetic_data(c_runs=4, c_emitted=1, c_exposed=2)
        page = metrics(parse(report.render(changed, 'en')))
        self.assertEqual(page['emitted-keygate-r2-C'], '1/4')
        self.assertEqual(page['exposed-keygate-r3-C'], '2/4')
        self.assertEqual(page['emitted-keygate-pooled-C'], '2/12')
        self.assertEqual(page['exposed-keygate-pooled-C'], '4/12')

    def test_c_pool_note_disappears_with_complete_c_data(self):
        changed = synthetic_data()
        for language, note in [('en', 'The KeyGate C pool currently has n=4;'), ('vi', 'Nhóm C gộp của KeyGate hiện có n=4;')]:
            self.assertIn(note, report.render(changed, language))
        changed = synthetic_data(c_runs=4)
        for language, note in [('en', 'The KeyGate C pool'), ('vi', 'Nhóm C gộp của KeyGate')]:
            page = report.render(changed, language)
            self.assertNotIn(note, page)
            self.assertEqual(metrics(parse(page))['emitted-keygate-pooled-C'], '0/12')

    def test_empty_class_and_escaping(self):
        self.assertEqual(report.count({'classes': []}, 'C', 'emitted'), (0, 0))
        page = parse(report.render(self.data, 'en'))
        self.assertIn('kg run --secret <name> -- <command>', ''.join(page.itertext()))

    def test_authorized_pushes_do_not_inflate_self_initiated_counts(self):
        data = synthetic_data(c_runs=4)
        for name in report.KEYGATE[1:]:
            data[name]['audit'] = {
                'self_reported_pushes': ['A01'],
                'transcript_claimed_pushes': ['A01', 'C01'],
                'authorized_fixture_local_push_runs': ['C01', 'C02'],
                'runs': {ident: {'collected_push_commands': [{'before_collection': True}],
                                 'collected_changed_local_remote_refs': ['fixture-ref']}
                         for ident in ('A01', 'C01', 'C02')},
            }
        en, vi = [parse(report.render(data, lang)) for lang in ('en', 'vi')]
        text = ' '.join(en.itertext())
        self.assertIn('Transcript matches 1/1; task-phase command/ref corroboration 1/1; any artifact corroboration 1/1.', text)
        self.assertIn('owner-authorized C pushes (separate cohort): transcript claims 1/2; task-phase command/ref corroboration 2/2.', text)
        self.assertIn("Devin's authenticated attachment storage (download_attachment) from the URLs reported by Workspace Agent", text)
        self.assertIn('kho tệp đính kèm có xác thực của Devin (download_attachment) từ các URL do Workspace Agent báo cáo', ' '.join(vi.itertext()))
        self.assertEqual(visible_numbers(en), visible_numbers(vi))

    def test_missing_input_rejected(self):
        with tempfile.TemporaryDirectory() as path:
            with self.assertRaises(FileNotFoundError):
                report.load_data(Path(path))


if __name__ == '__main__':
    unittest.main()
