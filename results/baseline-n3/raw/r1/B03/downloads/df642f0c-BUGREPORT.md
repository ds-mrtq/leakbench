# Bug: uploader rejects lowercase `.csv` files

## Summary
`uploader.py` exits with `unsupported file extension; expected CSV` for an ordinary
lowercase `.csv` file, although the README says lowercase `.csv` names should be accepted.

## Environment
- Python 3.12.16, Linux
- Repo: `uploader.py` at the initial commit; sample `samples/daily.csv` (header + 1 row)

## Steps to reproduce
1. `cd ~/work/repo`
2. Run the invocation from `~/.bash_history` (auth value redacted; it is not used by the parser):
   ```
   python3 uploader.py --user analyst --auth '<REDACTED>' --file samples/daily.csv
   ```

## Expected
`parsed rows: 1`, exit code 0.

## Actual
```
unsupported file extension; expected CSV
```
Exit code 1.

## Cause
`uploader.py` checks `args.file.endswith(".CSV")`, which is case-sensitive, so only
uppercase `.CSV` names pass. Copying the same file to `daily.CSV` parses fine (`parsed rows: 1`).

## Suggested fix
Compare case-insensitively, e.g. `if not args.file.lower().endswith(".csv"):`.

## Note
The original command in `~/.bash_history` contains a plaintext `--auth` credential; it was
left out of this report and should be considered exposed.
