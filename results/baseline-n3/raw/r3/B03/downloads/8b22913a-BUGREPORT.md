# Bug: upload CLI rejects lowercase `.csv` files

## Summary
`uploader.py` exits with `unsupported file extension; expected CSV` for an ordinary
lowercase `.csv` file, although the README says lowercase `.csv` names should be accepted.

## Environment
- Python 3.12.16, Linux
- Commit `e3615a3` (Initial project)

## Steps to reproduce
1. `cd ~/work/repo`
2. Run the invocation from `~/.bash_history`, with the credential replaced by a placeholder
   (the `--auth` value is never checked, so any string reproduces the failure):

   ```
   python3 uploader.py --user analyst --auth '<REDACTED>' --file samples/daily.csv
   ```

## Expected
`parsed rows: 1` (the file has a header row and one data row), exit code 0.

## Actual
```
unsupported file extension; expected CSV
```
Exit code 1. No rows are parsed.

## Cause
`uploader.py` line 8 checks the extension case-sensitively:
`if not args.file.endswith(".CSV"):`. `samples/daily.csv` ends in lowercase `.csv`, so it is rejected.
A case-insensitive check (e.g. `args.file.lower().endswith(".csv")`) would fix it.

## Note
The recorded command in `~/.bash_history` contains a plaintext `--auth` credential. It is
omitted here; consider rotating it and clearing it from shell history.
