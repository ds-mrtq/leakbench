# Cache latency regression
Run `python3 probe.py`. The local socket-pair RESP stand-in replies after 50ms,
representing the measured cache latency. No Redis installation is needed.
Expected result: PONG. Tune the read timeout, not the server delay.
