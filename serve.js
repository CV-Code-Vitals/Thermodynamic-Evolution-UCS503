const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = process.env.PORT || 8000;
const ROOT_DIR = path.resolve(__dirname);
const ADMIN_SOURCE = path.join(ROOT_DIR, 'admin-portal');
const ADMIN_DIST = path.join(ADMIN_SOURCE, 'dist');

function getAdminRoot() {
  return fs.existsSync(ADMIN_DIST) ? ADMIN_DIST : ADMIN_SOURCE;
}

const mimeTypes = {
  '.html': 'text/html',
  '.css': 'text/css',
  '.js': 'application/javascript',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.ico': 'image/x-icon',
  '.woff2': 'font/woff2',
  '.woff': 'font/woff',
  '.ttf': 'font/ttf',
};

function getFilePath(urlPath) {
  const adminRoot = getAdminRoot();

  if (urlPath === '/admin-portal') {
    return { redirect: '/admin-portal/' };
  }

  if (urlPath.startsWith('/admin-portal/')) {
    const subPath = urlPath.replace('/admin-portal/', '');
    const safePath = subPath ? path.normalize(subPath) : 'index.html';
    const targetPath = path.join(adminRoot, safePath);
    return { file: targetPath };
  }

  const safePath = urlPath === '/' ? 'index.html' : path.normalize(urlPath.replace(/^\//, ''));
  return { file: path.join(ROOT_DIR, safePath) };
}

function sendFile(res, filePath) {
  const ext = path.extname(filePath).toLowerCase();
  const contentType = mimeTypes[ext] || 'application/octet-stream';
  const relativePath = path.relative(ROOT_DIR, filePath);
  const isLegacyPage = relativePath === 'index.html' || relativePath.startsWith(`graph-visualizer${path.sep}`);
  const contentSecurityPolicy = isLegacyPage
    ? "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; script-src 'self' 'unsafe-inline'; connect-src 'self' http://localhost:8080 http://127.0.0.1:8080 https://api.github.com https://raw.githubusercontent.com; base-uri 'self'; frame-ancestors 'none'"
    : "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com data:; script-src 'self'; base-uri 'self'; frame-ancestors 'none'";

  fs.readFile(filePath, (err, content) => {
    if (err) {
      res.writeHead(404, { 'Content-Type': 'text/plain' });
      res.end('404 Not Found');
      return;
    }

    res.writeHead(200, {
      'Content-Type': contentType,
      'X-Content-Type-Options': 'nosniff',
      'Content-Security-Policy': contentSecurityPolicy,
    });
    res.end(content);
  });
}

const server = http.createServer((req, res) => {
  // Reverse proxy /api/* traffic to the Go backend on port 8080
  if (req.url.startsWith('/api/') || req.url === '/api') {
    const backendPort = process.env.BACKEND_PORT || 8080;
    const proxyReq = http.request({
      hostname: '127.0.0.1',
      port: backendPort,
      path: req.url,
      method: req.method,
      headers: {
        ...req.headers,
        host: `127.0.0.1:${backendPort}`,
      },
    }, (proxyRes) => {
      res.writeHead(proxyRes.statusCode, proxyRes.headers);
      proxyRes.pipe(res, { end: true });
    });

    proxyReq.on('error', (err) => {
      console.error(`[PROXY ERROR] Could not reach backend on port ${backendPort}:`, err.message);
      res.writeHead(502, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'bad_gateway', message: `Cannot connect to Go backend on port ${backendPort}` }));
    });

    req.pipe(proxyReq, { end: true });
    return;
  }

  let requestUrl;
  try {
    requestUrl = decodeURIComponent(req.url.split('?')[0]);
  } catch {
    res.writeHead(400, { 'Content-Type': 'text/plain' });
    res.end('400 Bad Request');
    return;
  }
  if (requestUrl.split('/').some((segment) => segment.startsWith('.') && segment.length > 1)) {
    res.writeHead(404, { 'Content-Type': 'text/plain' });
    res.end('404 Not Found');
    return;
  }
  const { redirect, file } = getFilePath(requestUrl);

  if (redirect) {
    res.writeHead(302, { Location: redirect });
    res.end();
    return;
  }

  const resolvedPath = file && path.resolve(file);
  const allowedRoot = requestUrl.startsWith('/admin-portal/') ? getAdminRoot() : ROOT_DIR;
  if (!resolvedPath || (resolvedPath !== allowedRoot && !resolvedPath.startsWith(allowedRoot + path.sep))) {
    res.writeHead(400, { 'Content-Type': 'text/plain' });
    res.end('400 Bad Request');
    return;
  }

  fs.stat(resolvedPath, (err, stats) => {
    if (err) {
      res.writeHead(404, { 'Content-Type': 'text/plain' });
      res.end('404 Not Found');
      return;
    }

    if (stats.isDirectory()) {
      sendFile(res, path.join(resolvedPath, 'index.html'));
    } else {
      sendFile(res, resolvedPath);
    }
  });
});

server.listen(PORT, () => {
  console.log(`Static server running at http://localhost:${PORT}`);
  console.log(`Root site: http://localhost:${PORT}/`);
  console.log(`Admin portal: http://localhost:${PORT}/admin-portal/`);
});
