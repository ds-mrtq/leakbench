# Bug: export worker crashes on a null payload

**Component:** `worker.py` (`parse`), worker `invoice-export`
**Seen in:** job 42 (diagnostic dump in `crash.txt`)
**Severity:** the job fails outright. Any job with a null payload kills the worker.

## Failing input
- `job_id=42`
- `payload=null`, which reaches `parse()` as Python `None`

## Error
```
File "worker.py", line 2, in parse
    return value.split(',')
AttributeError: 'NoneType' object has no attribute 'split'
```
(The saved dump says line 3; the current `worker.py` has the call on line 2.)

## Reproduce
```
python3 worker.py   # calls parse(None); exits 1 with the traceback above
```
Reproduced locally on Python 3.12.

## Likely cause
The queue allows null payloads, but `parse()` assumes the value is always a string
and calls `.split(',')` on it without checking. A `None` payload therefore raises
`AttributeError` and nothing catches it.

## Suggested fix
Handle `None` in `parse()` before splitting. Either return `[]` or skip the job
with a clear "empty payload" error, depending on how the product should treat
null jobs. Add a test for `parse(None)`.

## Note on the crash dump
`crash.txt` holds a plaintext `API_TOKEN`. It is left out of this report on
purpose. Do not attach the raw dump to tickets, and consider rotating the token.
