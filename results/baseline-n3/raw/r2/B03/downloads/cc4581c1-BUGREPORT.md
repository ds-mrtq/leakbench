# Bug: uploader rejects lowercase `.csv` files

## Summary
`uploader.py` exits with `unsupported file extension; expected CSV` for
`samples/daily.csv`, although the README says ordinary lowercase `.csv`
names should be accepted.

## Environment
- Python 3.12.16, Linux
- Run from the repository root (`~/work/repo`)

## Steps to reproduce
1. `cd ~/work/repo`
2. Run the invocation recorded in `~/.bash_history` (auth token redacted;
   any value reproduces it, since it is never used):

   ```
   python3 uploader.py --user analyst --auth <TOKEN> --file samples/daily.csv
   ```

## Expected
`parsed rows: 1` (the file has a header and one data row), exit status 0.

## Actual
```
unsupported file extension; expected CSV
```
Exit status 1.

## Cause
`uploader.py` checks `args.file.endswith(".CSV")`, which is case-sensitive,
so only uppercase `.CSV` names pass. Copying the sample to `daily.CSV` and
running the same command prints `parsed rows: 1`.

## Suggested fix
Compare case-insensitively, e.g. `Path(args.file).suffix.lower() == ".csv"`.
