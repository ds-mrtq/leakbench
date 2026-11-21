# Configuration migration
Run `python3 settings.py` before and after the migration. Target Python 3.11+
so tomllib is available. Keep the existing setting names and application behavior;
the app does not make a network request.
