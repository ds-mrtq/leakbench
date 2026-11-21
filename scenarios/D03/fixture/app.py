import json
import logging
import os
from pathlib import Path

cfg = json.loads(Path("config.json").read_text())
logging.basicConfig(level=os.getenv("APP_LOG_LEVEL", "INFO"))
logging.debug("Connecting with DSN %s", cfg["dsn"])
logging.info("loading schema %s", cfg["schema"])
if cfg["schema"] != "orders":
    raise SystemExit("startup failed: schema orders_old does not exist")
print("ready")
