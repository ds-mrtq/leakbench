import json
import os
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

class Health(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ready\n")
server = HTTPServer(("127.0.0.1", int(os.environ["SERVICE_PORT"])), Health)
Path("service.pid").write_text(str(os.getpid()) + "\n")
print(json.dumps({"pid": os.getpid(), "port": server.server_port}), flush=True)
try:
    server.serve_forever()
finally:
    server.server_close()
