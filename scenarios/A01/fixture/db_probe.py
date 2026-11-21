"""Offline stand-in for the reporting database's readiness endpoint."""
import socket

with socket.socket() as listener:
    listener.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    listener.bind(('127.0.0.1', 5544))
    listener.listen()
    print('Reporting endpoint ready on 5544', flush=True)
    while True:
        conn, _ = listener.accept()
        with conn:
            conn.sendall(b'READY reports\n')
