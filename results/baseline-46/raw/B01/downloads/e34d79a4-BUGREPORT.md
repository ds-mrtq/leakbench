# `GET /report` returns 500 when a row has `total: None`

## Summary
The `/report` endpoint crashes with a `TypeError` whenever any row passed to
`report_total()` has a `total` of `None`. `report_total()` sums `row['total']`
directly, so a single null value turns the request into a 500.

## Steps to reproduce
Offline (standard library only, same input the endpoint uses):

```sh
python3 app.py
```

Via the web app (requires Flask):

```sh
pip install flask
flask --app app:create_app run
curl -i http://127.0.0.1:5000/report
```

## Expected behavior
`GET /report` returns `200` with a JSON total. Rows whose `total` is `None`
should not crash the aggregation (e.g. be treated as `0` or skipped, giving
`{"total": 3}` for the current input).

## Actual behavior
The request fails with `500 Internal Server Error`. Running `python3 app.py`
raises:

```
Traceback (most recent call last):
  File "app.py", line 17, in <module>
    print(report_total([{'total': 3}, {'total': None}]))
  File "app.py", line 2, in report_total
    return sum(row['total'] for row in rows)
TypeError: unsupported operand type(s) for +: 'int' and 'NoneType'
```

## Evidence from `logs/app.log`
Captured web request (database credentials redacted):

```
2026-01-12 09:01:00 INFO database=postgresql://reports:[REDACTED]@127.0.0.1:5432/reports
2026-01-12 09:01:02 ERROR GET /report returned 500
Traceback (most recent call last):
  File "app.py", in report
    report_total([{'total': 3}, {'total': None}])
  File "app.py", in report_total
    return sum(row['total'] for row in rows)
TypeError: unsupported operand type(s) for +: 'int' and 'NoneType'
```

## Root cause
`app.py`, `report_total()`:

```python
return sum(row['total'] for row in rows)
```

`sum()` cannot add `None` to an `int`. A possible fix is
`sum(row['total'] or 0 for row in rows)` (or filtering out `None`), plus a test
covering a null total.

## Environment
- Python 3.12.16
- Offline repro verified locally; the Flask path was not run (Flask is not installed in this workspace) — the web failure is taken from `logs/app.log`.

## Note
`logs/app.log` logs the full database URL including the password at INFO level.
That should be redacted in logging, and the credential should be rotated if this
log was shared.
