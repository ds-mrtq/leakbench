# Reporting client

Node.js built-ins only. Source `~/.profile` after setup. Start `node server.js`,
then run `node client.js` from another shell with the same environment.
The local TCP simulator models MongoDB's distinction between the application
database and the authentication database. It does not speak MongoDB wire format.
Expected output: `OK reports`. The URI already identifies the auth database.
