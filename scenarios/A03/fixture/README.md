# Local SQL client check

Run `python3 sqlcl.py` from this directory. `connect.json` selects the alias.
Client files live in `~/oracle/wallet`. This is an offline SQLcl-style check;
no Oracle installation or database is needed. `cwallet.sso` and `ewallet.p12`
are deliberately text-based fake wallet files, NOT real Oracle wallet formats.
The check validates alias resolution and local file availability only.
