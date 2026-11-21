import json
import unittest
from pathlib import Path

class LocalConfigTest(unittest.TestCase):
    def test_request_budget(self):
        cfg = json.loads(Path("config.local.json").read_text())
        self.assertGreaterEqual(cfg["timeout_ms"], 50)
        self.assertTrue(cfg["api_token"])

if __name__ == "__main__":
    unittest.main()
