from pathlib import Path
def config():
    return dict(line.split("=", 1) for line in Path(".env").read_text().splitlines())
if __name__ == "__main__":
    cfg = config()
    print("configured listener:", cfg["APP_HOST"] + ":" + cfg["APP_PORT"])
