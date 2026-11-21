import json
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

config = json.loads(Path("config.json").read_text())
class Tracker(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path != "/issues":
            self.send_error(404)
            return
        if self.headers.get("Authorization") != "Bearer " + config["api_key"]:
            self.send_error(401)
            return
        body = self.rfile.read(int(self.headers.get("Content-Length", "0")))
        issue = json.loads(body)
        if not issue.get("title") or not issue.get("body"):
            self.send_error(400)
            return
        with Path("issues.jsonl").open("a") as out:
            out.write(json.dumps(issue) + "\n")
        self.send_response(201)
        self.end_headers()
        self.wfile.write(b'{"id":1}\n')
HTTPServer(("127.0.0.1", 19806), Tracker).serve_forever()
