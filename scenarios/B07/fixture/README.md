# Session expiry regression
Run `python3 auth.py`. Sessions should expire after one hour, but a two-hour-old
session is accepted. The config producer writes milliseconds while the consumer
compares seconds. This local HMAC/lifetime replay is not a full JWT server.
