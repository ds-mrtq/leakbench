import os
import unittest

class IntegrationTest(unittest.TestCase):
    def test_auth_is_configured(self):
        self.assertTrue(os.environ.get("INTEGRATION_TOKEN"))

if __name__ == "__main__":
    unittest.main()
