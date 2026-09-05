import http from 'node:http';

const port = Number(process.env.PORT || 4176);

const server = http.createServer((request, response) => {
  const requestURL = new URL(request.url, `http://${request.headers.host}`);
  response.setHeader('Access-Control-Allow-Origin', '*');
  response.setHeader('Access-Control-Allow-Headers', 'Accept, Content-Type');
  response.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  if (request.method === 'OPTIONS') {
    response.writeHead(204);
    response.end();
    return;
  }
  if (request.method === 'GET' && requestURL.pathname === '/v1/projects') {
    const body = JSON.stringify({ data: [{ id: 'prj_e2e', name: 'E2E project' }], next_cursor: null });
    response.writeHead(200, { 'Content-Type': 'application/json' });
    response.end(body);
    return;
  }
  if (request.method === 'GET' && requestURL.pathname === '/v1/projects/prj_e2e') {
    response.writeHead(200, { 'Content-Type': 'application/json' });
    response.end(JSON.stringify({ id: 'prj_e2e', name: 'E2E project detail' }));
    return;
  }
  if (request.method === 'POST' && request.url === '/v1/projects') {
    let body = '';
    request.on('data', chunk => { body += chunk; });
    request.on('end', () => {
      response.writeHead(201, { 'Content-Type': 'application/json' });
      response.end(JSON.stringify({ received: JSON.parse(body || '{}') }));
    });
    return;
  }
  response.writeHead(404, { 'Content-Type': 'application/json' });
  response.end(JSON.stringify({ error: 'not found' }));
});

server.listen(port, '127.0.0.1');
