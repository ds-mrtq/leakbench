import unittest

from prepare import messages, summarize


class PreparationTests(unittest.TestCase):
    def test_message_export(self):
        text = 'Showing 1\nNo more results (last page).\n[devin] @ 1795278050 (event-one)\nbody\n'
        event = messages(text)[0]
        self.assertEqual(event['id'], 'event-one')
        self.assertEqual(event['direction'], 'outgoing')
        self.assertEqual(event['body']['message'], 'body')

    def test_empty_and_incomplete_exports_rejected(self):
        for text in ('', 'Showing 1\n[devin] @ 1 (event-one)\nbody',
                     'Showing 2\nNo more results (last page).\n[devin] @ 1 (event-one)\nbody'):
            with self.assertRaises(AssertionError):
                messages(text)

    def test_summary_empty(self):
        result = summarize([])
        self.assertEqual(result['n'], 0)
        self.assertEqual(sum(c['n'] for c in result['classes']), 0)

    def test_summary_counts(self):
        result = summarize([dict(run='D08', exposed=True, emitted=False),
                            dict(run='A01', exposed=True, emitted=True)])
        self.assertEqual((result['n'], result['exposed_runs'], result['emitted_runs']), (2, 2, 1))
        self.assertEqual(result['classes'][3]['exposed_runs'], 1)


if __name__ == '__main__':
    unittest.main()
