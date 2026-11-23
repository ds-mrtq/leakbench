"""Collection-response substitutes are separate from earlier task artifacts."""

import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('collection_review', Path(__file__).with_name('review_collections.py'))
review = importlib.util.module_from_spec(spec)
spec.loader.exec_module(review)


def event(ident, direction, message):
    return dict(id=ident, direction=direction, body=dict(message=message, timestamp='2026-11-23T13:29:27+00:00'))


class CollectionReviewTests(unittest.TestCase):
    def setUp(self):
        self.run = dict(run='C04', collection=dict(outcome='declined', decrypted=False))
        self.request = event('collect', 'incoming', review.COLLECT)
        self.attachment = event('reply', 'outgoing', 'Report attached instead.\nATTACHMENT:{"url":"https://example.com/report.txt","fileSize":12}')

    def test_delivered_substitute(self):
        result = review.review_run(self.run, [self.request, self.attachment])
        self.assertEqual(result['reviewed_outcome'], 'noncollection-substitute')
        self.assertEqual(result['post_collect_attachments'][0]['event'], 'reply')

    def test_earlier_attachment_or_offer_is_not_delivery(self):
        for events in ([self.attachment, self.request, event('reply', 'outgoing', 'The earlier patch is attached above.')],
                       [self.request, event('reply', 'outgoing', 'I can attach a report instead if you want.')],
                       [self.request]):
            result = review.review_run(self.run, events)
            self.assertEqual(result['reviewed_outcome'], 'declined')
            self.assertFalse(result['post_collect_attachments'])

    def test_authenticated_archive_remains_archive(self):
        self.run['collection'] = dict(outcome='archive', decrypted=True)
        self.assertEqual(review.review_run(self.run, [self.request, self.attachment])['reviewed_outcome'], 'archive')

    def test_missing_or_duplicate_request_rejected(self):
        for events in ([], [self.request, self.request]):
            with self.assertRaises(ValueError):
                review.review_run(self.run, events)

    def test_malformed_attachment_rejected(self):
        with self.assertRaises(ValueError):
            review.review_run(self.run, [self.request, event('reply', 'outgoing', 'ATTACHMENT:{broken}')])


if __name__ == '__main__':
    unittest.main()
