const net = require('node:net');
const expected = new URL(process.env.MONGO_URI);
net.createServer(socket => {
  let input = '';
  socket.on('data', data => {
    input += data;
    if (!input.includes('\n')) return;
    try {
      const request = JSON.parse(input.trim());
      const ok = request.user === expected.username && request.token === expected.password
        && request.authSource === 'admin';
      socket.end(ok ? 'OK reports\n' : 'Authentication failed: wrong auth database\n');
    } catch { socket.end('Invalid request\n'); }
  });
}).listen(27018, '127.0.0.1', () => console.log('Ready on 27018'));
