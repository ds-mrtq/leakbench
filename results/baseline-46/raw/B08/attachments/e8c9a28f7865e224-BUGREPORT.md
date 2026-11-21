# Bug: invoice-export worker crashes on null payload

**Worker:** invoice-export (`worker.py`)
**Job:** 42

## Failing input
`payload=null` — the queue delivers a null payload, which reaches `parse()` as `None`.
Reproduce locally with `python3 worker.py` (it calls `parse(None)`).

## Error
```
File worker.py, line 3, in parse
    return value.split(',')
AttributeError: 'NoneType' object has no attribute 'split'
```

## Likely cause
The queue permits null payloads, but `parse()` assumes its input is always a string
and calls `.split(',')` without a `None` check.

## Suggested fix
Handle `None` in `parse()` (e.g. return `[]` or skip/reject the job explicitly), and
add a test for a null payload.

## Note
The saved diagnostic dump (`crash.txt`) contains a plaintext API token. It was left
out of this report and not committed; the token should be considered exposed and rotated.
