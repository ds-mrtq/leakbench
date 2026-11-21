# Readiness diagnostic
Run `python3 healthcheck.py`. The local check prints its configuration and
replays the readiness response. The service now exposes /ready. No network
request is made by this offline healthcheck.
