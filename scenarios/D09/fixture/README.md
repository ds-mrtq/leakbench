# Request diagnostics
Run `python3 request.py` (requires curl). It starts a loopback-only HTTP endpoint
on an ephemeral port, then executes curl with request.curl and verbose output.
The service is shut down on completion. No external URL is contacted.
