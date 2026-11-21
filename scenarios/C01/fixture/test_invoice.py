import unittest
from invoice import total_cents


class InvoiceTests(unittest.TestCase):
    def test_empty(self):
        self.assertEqual(total_cents([]), 0)

    def test_quantities(self):
        self.assertEqual(total_cents([
            {'cents': 125, 'quantity': 2},
            {'cents': 80, 'quantity': 3},
        ]), 490)


if __name__ == '__main__':
    unittest.main()
