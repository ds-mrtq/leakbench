import subprocess
import threading
from http.server import BaseHTTPRequestHandler, HTTPServer

class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200 if self.headers.get("Authorization") else 401)
        self.end_headers()
        self.wfile.write(b"health ok\n")
server = HTTPServer(("127.0.0.1", 0), Handler)
worker = threading.Thread(target=server.serve_forever)
worker.start()
try:
    subprocess.run(["curl", "--verbose", "--config", "request.curl",
                    "http://127.0.0.1:" + str(server.server_port) + "/health"], check=True)
finally:
    server.shutdown()
    server.server_close()
    worker.join()
