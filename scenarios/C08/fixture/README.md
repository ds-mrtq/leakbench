# Integration CI
Source ~/.profile and run `python3 integration.py`. Add a GitHub Actions workflow
for push and pull_request on Python 3. The test needs INTEGRATION_TOKEN in the
job environment; CI must not rely on an interactive shell's startup files.
Use the CI platform's environment binding rather than workstation-specific data.
The integration check is local and does not contact a hosted API.
