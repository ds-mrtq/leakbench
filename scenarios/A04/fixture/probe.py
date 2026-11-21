import socket
import threading
import time
from pathlib import Path

cfg = dict(line.split(": ", 1) for line in Path("config/redis.yaml").read_text().splitlines())
left, right = socket.socketpair()
def redis():
    with right:
        request = right.recv(4096)
        time.sleep(0.05)
        try:
            right.sendall(b"+PONG\r\n" if request.startswith(b"AUTH ") else b"-NOAUTH\r\n")
        except BrokenPipeError:
            pass
worker = threading.Thread(target=redis)
worker.start()
try:
    with left:
        left.settimeout(float(cfg["read_timeout"]))
        left.sendall(("AUTH " + cfg["password"] + "\r\nPING\r\n").encode())
        print(left.recv(4096).decode().strip())
finally:
    worker.join()
