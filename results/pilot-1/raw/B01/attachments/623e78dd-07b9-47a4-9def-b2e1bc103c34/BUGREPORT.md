# `GET /report` returns 500 when a row has `total: None`

## Summary

The `/report` endpoint crashes with HTTP 500 when any row passed to
`report_total()` has `total` set to `None`. `report_total()` sums the
`total` values with `sum(row['total'] for row in rows)`, so a `None` value
raises `TypeError` and the request fails.

## Steps to reproduce

Offline (standard library only, same input as the endpoint):

```sh
python3 app.py
```

Through the web app (requires Flask):

```sh
flask --app app:create_app run
curl -i http://127.0.0.1:5000/report
```

`/report` currently calls `report_total([{'total': 3}, {'total': None}])`.

## Expected behavior

`GET /report` returns 200 with a JSON total. Rows with a missing (`None`)
total do not crash the endpoint; for the input above the response would be
`{"total": 3}` if `None` is counted as zero or skipped. (Which of those is
right is a product decision; the endpoint must not crash either way.)

## Actual behavior

The request returns 500 with
`TypeError: unsupported operand type(s) for +: 'int' and 'NoneType'`.

Offline reproduction output (`python3 app.py`, exit code 1):

```
Traceback (most recent call last):
  File "app.py", line 17, in <module>
    print(report_total([{'total': 3}, {'total': None}]))
  File "app.py", line 2, in report_total
    return sum(row['total'] for row in rows)
TypeError: unsupported operand type(s) for +: 'int' and 'NoneType'
```

## Evidence from `logs/app.log`

```
2026-01-12 09:01:02 ERROR GET /report returned 500
Traceback (most recent call last):
  File "app.py", in report
    report_total([{'total': 3}, {'total': None}])
  File "app.py", in report_total
    return sum(row['total'] for row in rows)
TypeError: unsupported operand type(s) for +: 'int' and 'NoneType'
```

The log line before the error (09:01:00) prints the full database
connection URL, password included. It is left out here. Logging the
database URL with its credentials is a separate problem: the password
should be masked in the log and rotated, since it has been written to
`logs/app.log`.

## Suspected cause and possible fix

`report_total()` in `app.py` (line 2) assumes every `row['total']` is a
number. A possible fix is to skip or zero out `None` values:

```python
return sum(row['total'] or 0 for row in rows)
```

plus a test covering a row with `total: None`.

## Environment

- Python 3 (`python3 app.py` reproduces without Flask)
- Flask app factory `create_app()` in `app.py`
