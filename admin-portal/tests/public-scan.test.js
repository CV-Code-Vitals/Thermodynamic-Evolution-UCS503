import assert from 'node:assert/strict';
import test from 'node:test';
import { scanPublicRepository } from '../../graph-visualizer/scan.mjs';

const report = { file_reports: [{ file_path: 'main.go', language: 'Go' }], files_analyzed: 1 };

test('public scanning works without credentials through the local proxy or cloud gateway', async (t) => {
  for (const [config, endpoint, response] of [
    [{ apiBase: '/api', analysisApiBase: '' }, '/api/analysis/scan', report],
    [{ apiBase: 'https://archive.example/api', analysisApiBase: 'https://analysis.example/prod/' }, 'https://analysis.example/prod/api/scan', { report }],
  ]) {
    const signal = new AbortController().signal;
    t.mock.method(globalThis, 'fetch', async (url, options) => {
      assert.equal(url, endpoint);
      assert.equal(options.method, 'POST');
      assert.equal(options.credentials, 'omit');
      assert.equal(options.headers.Authorization, undefined);
      assert.equal(options.headers['X-API-Token'], undefined);
      assert.deepEqual(JSON.parse(options.body), { repo_url: 'https://github.com/example/project' });
      assert.equal(options.signal, signal);
      return Response.json(response);
    });
    assert.deepEqual(await scanPublicRepository(config, ' https://github.com/example/project ', signal), report);
    t.mock.restoreAll();
  }
});

test('public scanning reports unavailable services, rate limits, and invalid reports', async (t) => {
  const config = { apiBase: '/api' };
  for (const [body, status, message] of [
    [{ error: 'analysis_not_configured' }, 503, /not configured/],
    [{ error: 'analysis_unavailable' }, 502, /unavailable/],
    [{ error: 'scan_rate_limited' }, 429, /wait a minute/],
    [{ error: 'scan_capacity_reached' }, 429, /busy/],
    [{ error: 'invalid_repository' }, 400, /Enter a public GitHub repository URL/],
    [{}, 200, /did not return a repository report/],
  ]) {
    t.mock.method(globalThis, 'fetch', async () => Response.json(body, { status }));
    await assert.rejects(scanPublicRepository(config, 'https://github.com/example/project'), message);
    t.mock.restoreAll();
  }
});
