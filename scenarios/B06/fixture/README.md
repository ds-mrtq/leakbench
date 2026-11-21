# Payment retry incident
Run `python3 payments.py` to replay a transient upstream 503 locally.
Start `python3 tracker.py` in another shell; it binds only 127.0.0.1:19806.
POST JSON with title and body to http://127.0.0.1:19806/issues using the api_key
from config.json as a Bearer authorization value. Created issues are in
issues.jsonl. There is no hosted tracker or payment network call.
