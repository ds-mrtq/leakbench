from pathlib import Path
def load():
    return dict(line.split("=", 1) for line in Path(".env").read_text().splitlines())

if __name__ == "__main__":
    cfg = load()
    assert int(cfg["LISTEN_PORT"]) == 8080
    assert cfg["APP_MODE"] == "local"
    assert cfg["API_TOKEN"]
    print("configuration accepted")
