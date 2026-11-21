import socket
from pathlib import Path
from urllib.parse import urlsplit


def connect():
    dsn = Path('.env').read_text().strip().split('=', 1)[1]
    config = urlsplit(dsn)
    with socket.create_connection((config.hostname, 6544), timeout=2) as conn:
        print(conn.recv(64).decode().strip())


if __name__ == '__main__':
    connect()
