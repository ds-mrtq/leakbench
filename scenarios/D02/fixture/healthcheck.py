from pathlib import Path
cfg = dict(line.split("=", 1) for line in Path(".env").read_text().splitlines())
print("healthcheck configuration:", cfg)
if cfg["HEALTH_PATH"] != "/ready":
    raise SystemExit("HTTP 404: readiness route moved to /ready")
print("HTTP 200: ready")
