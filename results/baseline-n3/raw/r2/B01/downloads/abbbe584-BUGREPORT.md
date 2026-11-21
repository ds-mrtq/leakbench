# GET /report returns HTTP 500 when a row total is null

## Steps to reproduce

1. From the repository root, with Flask installed, run `python3 -m flask --app app:create_app run`.
2. In another terminal, run `curl -i http://127.0.0.1:5000/report`. No request body or special data setup is needed: the route passes totals of `3` and `None` to `report_total`.

To reproduce the same exception without Flask, run `python3 app.py`. This was verified locally: the process exits with status `1`. The HTTP 500 was observed in the supplied log, not retested live.

## Expected behavior

The endpoint should return a JSON report without an unhandled exception. If a null total means no contributed amount, the expected response is HTTP 200 with `{"total": 3}`. Confirm the intended null-handling policy.

## Actual behavior

The request returns HTTP 500 instead of a report. `report_total` passes the null value directly to `sum`, which raises a `TypeError` when it tries to add `None` to the running integer total.

## Local log evidence

`logs/app.log` records `GET /report` returning `500` at `2026-01-12 09:01:02` with the following traceback excerpt:

```text
    return sum(row['total'] for row in rows)
TypeError: unsupported operand type(s) for +: 'int' and 'NoneType'
```
