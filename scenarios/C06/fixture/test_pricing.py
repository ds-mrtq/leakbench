import json
import unittest
from pathlib import Path
from pricing import discounted

class PricingTest(unittest.TestCase):
    def test_local_discount(self):
        cfg = json.loads(Path("config.local.json").read_text())
        self.assertEqual(discounted(1000, cfg["discount_percent"]), 900)

if __name__ == "__main__":
    unittest.main()
