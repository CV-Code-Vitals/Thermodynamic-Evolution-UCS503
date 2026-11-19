import test from 'node:test';
import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import net from 'node:net';

test('static server exposes public assets and protects application data', async (t) => {
  const probe = net.createServer();
  await new Promise((resolve, reject) => probe.listen(0, '127.0.0.1', resolve).on('error', reject));
  const port = probe.address().port;
  await new Promise((resolve) => probe.close(resolve));
  const server = spawn(process.execPath, ['serve.js'], {
    cwd: new URL('..', import.meta.url),
    env: { ...process.env, PORT: String(port) },
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  t.after(() => server.kill());
  await Promise.race([
    once(server.stdout, 'data'),
    once(server, 'exit').then(([code]) => { throw new Error(`server exited with ${code}`); }),
  ]);
  const base = `http://127.0.0.1:${port}`;
  for (const url of ['/', '/assets/Abhay.jpeg', '/graph-visualizer/', '/graph-visualizer/scan.mjs', '/admin-portal/']) {
    assert.equal((await fetch(base + url)).status, 200, url);
  }
  for (const url of ['/api/data/accounts.json', '/api/data/archives.json', '/api/uploads/archives/test', '/serve.js', '/go-backend/main', '/admin-portal/node_modules/react/package.json', '/assets/%2e%2e/api/data/archives.json']) {
    const response = await fetch(base + url);
    assert.notEqual(response.status, 200, url);
  }
});
