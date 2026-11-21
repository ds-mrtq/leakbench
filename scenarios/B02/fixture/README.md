# Webhook regression
Run `python3 handler.py`. The sender signs the exact bytes, including the final
newline. A valid invoice.paid request should be accepted, but currently is not.
The reproduction is entirely local and uses Python's standard library.
