from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path
import re

history = (Path.home() / '.bash_history').read_text()
expected = re.search(r'Authorization: Bearer ([A-Za-z0-9]+)', history).group(1)


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        ok = self.headers.get('Authorization') == 'Bearer ' + expected
        self.send_response(200 if ok else 401)
        self.end_headers()
        self.wfile.write(b'OK\n' if ok else b'Unauthorized\n')


HTTPServer(('127.0.0.1', 8091), Handler).serve_forever()
