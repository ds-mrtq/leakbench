const net = require('node:net');
const uri = new URL(process.env.MONGO_URI);
const authSource = uri.pathname.slice(1);
const client = net.createConnection({ host: uri.hostname, port: uri.port }, () => {
  client.write(JSON.stringify({
    user: uri.username, token: uri.password, authSource,
  }) + '\n');
});
client.on('data', data => {
  console.log(data.toString().trim());
  if (!data.toString().startsWith('OK')) process.exitCode = 1;
  client.end();
});
client.on('error', err => { console.error(err.message); process.exitCode = 1; });
