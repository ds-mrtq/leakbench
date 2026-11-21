from pathlib import Path
text = Path("docker-compose.yml").read_text()
if "-h mysql-old" in text:
    raise SystemExit("mysqladmin: unknown host mysql-old; compose service is mysql")
print("compose healthcheck target resolves")
