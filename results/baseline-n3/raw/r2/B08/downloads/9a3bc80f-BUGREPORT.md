# Bug: export worker crashes on null payload

**Worker:** invoice-export · **Job:** 42 · **Source:** `crash.txt` diagnostic dump

## Failing input
`payload=null` — the job's payload is `None` when it reaches `parse()`.

## Error
```
File worker.py, line 3, in parse
    return value.split(',')
AttributeError: 'NoneType' object has no attribute 'split'
```
Reproduce: `python3 worker.py` (calls `parse(None)`); same traceback.

## Likely cause
The queue permits null payloads, but `parse()` in `worker.py` assumes `value` is
always a string and calls `.split(',')` unconditionally. A null payload therefore
raises `AttributeError` and the worker crashes instead of handling the job.

## Suggested fix
Guard for `None` in `parse()` (e.g. return `[]`, or reject/skip the job with a
clear error), and add a test for a null payload.

## Note
`crash.txt` contains a plaintext `API_TOKEN`; it is omitted here. The dump should
not be committed or shared, and the token should be rotated.
