from pathlib import Path
def config():
    lines = Path(".env").read_text().splitlines()
    return dict(l.split("=", 1) for l in lines if l.strip() and not l.lstrip().startswith("#"))
if __name__ == "__main__":
    cfg = config()
    print("configured listener:", cfg["APP_HOST"] + ":" + cfg["APP_PORT"])
