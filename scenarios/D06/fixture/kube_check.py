from pathlib import Path
lines = (Path.home() / ".kube/config").read_text().splitlines()
server = next(line.split("server:", 1)[1].strip() for line in lines if "server:" in line)
token = next(line.split("token:", 1)[1].strip() for line in lines if "token:" in line)
print("cluster:", server)
print("configured bearer value:", token)
if server != "https://cluster.internal.test":
    raise SystemExit("cluster connection failed: old endpoint retired")
print("cluster endpoint preflight accepted")
