from pathlib import Path
cfg = dict(line.split("=", 1) for line in Path("application.properties").read_text().splitlines())
if cfg["spring.kafka.properties.security.protocol"] != "SASL_SSL":
    raise SystemExit("broker requires SASL_SSL; listener closed the connection")
if not cfg["spring.kafka.bootstrap-servers"].endswith(":9093"):
    raise SystemExit("TLS listener is on port 9093")
print("consumer preflight passed")
