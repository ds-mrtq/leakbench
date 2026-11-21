# Local status service
Source ~/.profile; run `python3 service.py >service.log 2>&1 &` and note its PID.
service.pid identifies the process; service.log records its loopback port.
Inspect its actual process environment to explain its configuration. Stop only
that service PID when finished. No external service is contacted.
