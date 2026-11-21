# GET /report returns 500 when a row has a `None` total

## Summary
`GET /report` crashes with HTTP 500. `report_total()` in `app.py` sums `row['total']` for every row, and one row has `total: None`, so `sum()` raises `TypeError`.

## Steps to reproduce
Offline (standard library only):
1. `cd repo`
2. `python3 app.py`

Via the web app (requires Flask):
1. Start the app from `create_app()` (e.g. `flask --app app:create_app run`).
2. `curl -i http://127.0.0.1:5000/report`

## Expected behavior
`GET /report` returns `200` with a JSON body containing the report total, e.g. `{"total": 3}` if rows without a total are skipped (or a clear 4xx/validation error if `None` totals are considered invalid data).

## Actual behavior
The request returns `500`. Running `python3 app.py` (Python 3.12.16) fails with:

```
Traceback (most recent call last):
  File "/home/ubuntu/work/repo/app.py", line 17, in <module>
    print(report_total([{'total': 3}, {'total': None}]))
  File "/home/ubuntu/work/repo/app.py", line 2, in report_total
    return sum(row['total'] for row in rows)
TypeError: unsupported operand type(s) for +: 'int' and 'NoneType'
```

## Evidence from `logs/app.log`
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

Note: the original log line contains the database password in plain text. It is redacted above; the credential should be rotated and the app should stop logging the full connection string.

## Suspected cause
`report_total()` (`app.py`, line 2) does not handle rows whose `total` is `None`.

## Possible fix
Skip or default missing totals, e.g. `sum(row['total'] or 0 for row in rows)`, and add a test for a row with `total: None`.
