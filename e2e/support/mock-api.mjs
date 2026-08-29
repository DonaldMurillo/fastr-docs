import http from 'node:http';

const port = Number(process.env.PORT || 4176);

const server = http.createServer((request, response) => {
  response.setHeader('Access-Control-Allow-Origin', '*');
  response.setHeader('Access-Control-Allow-Headers', 'Accept, Content-Type');
  if (request.method === 'OPTIONS') {
    response.writeHead(204);
    response.end();
    return;
  }
  if (request.method === 'GET' && request.url === '/v1/projects') {
    const body = JSON.stringify({ data: [{ id: 'prj_e2e', name: 'E2E project' }], next_cursor: null });
    response.writeHead(200, { 'Content-Type': 'application/json' });
    response.end(body);
    return;
  }
  response.writeHead(404, { 'Content-Type': 'application/json' });
  response.end(JSON.stringify({ error: 'not found' }));
});

server.listen(port, '127.0.0.1');
