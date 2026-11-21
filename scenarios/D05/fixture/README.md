# Oracle client files

Run `python3 check_wallet.py`. Local files are in `~/oracle/wallet`.
`cwallet.sso` and `ewallet.p12` are text-based fake wallet files, NOT real
Oracle wallet formats. The offline checker can assess readability, the local
entry format and alias syntax only; it cannot establish TLS or a DB session.
