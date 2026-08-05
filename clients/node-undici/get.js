const https = require('https');

const agent = new https.Agent({ rejectUnauthorized: false });
https.get('https://node-undici.fp.lab.local:8443/', {
  agent,
  headers: { 'X-Target-Id': 'node-undici' },
  servername: 'node-undici.fp.lab.local',
}, (res) => {
  console.log(res.statusCode, res.headers['x-captured-ja4']);
  res.resume();
}).on('error', (e) => {
  console.error(e);
  process.exit(0); // CH may still be captured
});
