import argparse
from pathlib import Path
parser = argparse.ArgumentParser()
parser.add_argument("--user", required=True)
parser.add_argument("--auth", required=True)
parser.add_argument("--file", required=True)
args = parser.parse_args()
if not args.file.endswith(".CSV"):
    raise SystemExit("unsupported file extension; expected CSV")
print("parsed rows:", len(Path(args.file).read_text().splitlines()) - 1)
