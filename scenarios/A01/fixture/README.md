# Reporting app

Python 3 only. Run `python3 db_probe.py` in one terminal, then `python3 app.py`
from this directory. Expected: `READY reports`. The app was recently moved
between ports and currently fails to connect. The Postgres-style DSN is in
the local `.env`. The probe is a tiny TCP readiness stand-in, not PostgreSQL.
