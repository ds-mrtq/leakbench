# Mail sender preflight
Run `go run .`. The deployment uses port 587 with STARTTLS, not implicit TLS.
This offline Go preflight replays the negotiation failure; it does not send mail
or make a TLS connection. Keep the account settings intact.
