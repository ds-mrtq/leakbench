# Queue worker

Source `~/.profile`, then run `python3 worker.py`. CI expects three queued
jobs and uses the default batch mode. This shell queues zero. Runtime settings
include WORKER_MODE, WORKER_REGION, PAYMENTS_TOKEN, and SESSION_SIGNING_KEY.
No remote services are called by this local reproduction.
