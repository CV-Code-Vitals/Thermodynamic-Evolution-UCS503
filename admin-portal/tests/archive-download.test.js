import test from 'node:test';
import assert from 'node:assert/strict';
import { fetchArchiveDownload } from '../src/archive-download.js';

test('downloads signed archive bytes without sending the API bearer token to S3', async () => {
  const expectedBlob = new Blob(['archive contents']);
  let requestedUrl;
  let requestOptions;

  const result = await fetchArchiveDownload(
    'https://api.example.com/api',
    'https://s3.example.com/report.pdf?signature=temporary',
    'memory-token',
    async (url, options) => {
    requestedUrl = url;
    requestOptions = options;
    return {
      ok: true,
      blob: async () => expectedBlob,
    };
    },
  );

  assert.equal(requestedUrl, 'https://s3.example.com/report.pdf?signature=temporary');
  assert.equal(requestOptions.headers.Authorization, undefined);
  assert.equal(requestOptions.credentials, 'omit');
  assert.equal(result, expectedBlob);
});

test('sends the bearer token only when downloading through the local API', async () => {
  let requestOptions;
  await fetchArchiveDownload('/api', '/api/local-download?key=archives%2Fid', 'memory-token', async (_url, options) => {
    requestOptions = options;
    return { ok: true, blob: async () => new Blob(['local file']) };
  });
  assert.equal(requestOptions.headers.Authorization, 'Bearer memory-token');
  assert.equal(requestOptions.credentials, 'omit');
});

test('rejects signed download failures with the API error message', async () => {
  await assert.rejects(
    fetchArchiveDownload('/api', 'https://s3.example.com/missing.pdf', 'token', async () => ({
      ok: false,
      status: 404,
      text: async () => 'Archive not found',
    })),
    /Archive not found/,
  );
});
