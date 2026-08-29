import fs from 'node:fs';
import http from 'node:http';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const root = path.resolve(process.argv[2]);
const port = Number(process.env.PORT || 4177);
const mime = {
  '.css': 'text/css; charset=utf-8',
  '.html': 'text/html; charset=utf-8',
  '.js': 'text/javascript; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.png': 'image/png',
  '.svg': 'image/svg+xml',
  '.webmanifest': 'application/manifest+json; charset=utf-8',
};

const safePath = (urlPath) => {
  const decoded = decodeURIComponent(urlPath.split('?')[0]);
  const relative = decoded.replace(/^\/+/, '');
  const candidate = path.resolve(root, relative);
  return candidate.startsWith(root + path.sep) || candidate === root ? candidate : null;
};

const server = http.createServer((request, response) => {
  const candidate = safePath(request.url || '/');
  if (!candidate) {
    response.writeHead(400);
    response.end('bad path');
    return;
  }
  const candidates = [
    candidate,
    path.join(candidate, 'index.html'),
    path.join(root, 'index.html'),
  ];
  const file = candidates.find((item) => fs.existsSync(item) && fs.statSync(item).isFile());
  if (!file) {
    response.writeHead(404);
    response.end('not found');
    return;
  }
  response.writeHead(200, { 'Content-Type': mime[path.extname(file)] || 'application/octet-stream' });
  fs.createReadStream(file).pipe(response);
});

server.listen(port, '127.0.0.1');
