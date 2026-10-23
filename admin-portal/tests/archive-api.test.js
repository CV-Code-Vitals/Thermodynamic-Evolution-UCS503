import test from 'node:test';
import assert from 'node:assert/strict';
import { createArchiveApi } from '../src/archive-api.js';

const jsonResponse = (status, body) => new Response(JSON.stringify(body), {
  status,
  headers: { 'Content-Type': 'application/json' },
});

test('login accepts only an authenticated response with a bearer token', async () => {
  const api = createArchiveApi('https://api.example.com/api', async (url, options) => {
    assert.equal(url, 'https://api.example.com/api/login');
    assert.equal(options.method, 'POST');
    assert.equal(options.headers['Content-Type'], 'application/json');
    assert.equal(options.credentials, 'omit');
    assert.deepEqual(JSON.parse(options.body), { passkey: 'correct-passkey' });
    return jsonResponse(200, { status: 'authenticated', token: 'opaque-token' });
  });

  assert.deepEqual(await api.login('correct-passkey'), {
    status: 'authenticated',
    token: 'opaque-token',
  });

  const invalidApi = createArchiveApi('/api', async () => jsonResponse(200, {
    status: 'authenticated',
  }));
  await assert.rejects(invalidApi.login('passkey'), /invalid response/i);
});

test('upload posts the presigned form before completing with bearer authentication', async () => {
  const calls = [];
  const api = createArchiveApi('https://api.example.com/api', async (url, options) => {
    calls.push({ url, options });
    if (url.endsWith('/upload')) {
      return jsonResponse(201, {
        uploadId: 'upload-123',
        url: 'https://s3.example.com/presigned',
        fields: { key: 'pending/upload-123', policy: 'signed-policy' },
      });
    }
    if (url === 'https://s3.example.com/presigned') {
      return new Response(null, { status: 204 });
    }
    return jsonResponse(201, { id: 'archive-123', state: 'available' });
  });
  const file = new Blob(['%PDF-1.7'], { type: 'application/pdf' });
  file.name = 'report.pdf';
  const metadata = {
    originalName: file.name,
    contentType: file.type,
    size: file.size,
    title: 'Report',
    version: 'v1',
    date: '2026-10-22',
    summary: 'Initial upload',
  };

  const uploadId = await api.upload('memory-token', metadata, file);

  assert.equal(uploadId, 'upload-123');
  assert.equal(calls.length, 2);
  assert.equal(calls[0].url, 'https://api.example.com/api/upload');
  assert.equal(calls[0].options.headers.Authorization, 'Bearer memory-token');
  assert.deepEqual(JSON.parse(calls[0].options.body), metadata);
  assert.equal(calls[1].url, 'https://s3.example.com/presigned');
  assert.equal(calls[1].options.headers?.Authorization, undefined);
  assert.equal(calls[1].options.body.get('key'), 'pending/upload-123');
  assert.equal(calls[1].options.body.get('policy'), 'signed-policy');
  assert.equal(calls[1].options.body.get('file').name, 'report.pdf');
  const result = await api.completeUpload('memory-token', uploadId);
  assert.deepEqual(result, { id: 'archive-123', state: 'available' });
  assert.equal(calls[2].url, 'https://api.example.com/api/upload/complete');
  assert.equal(calls[2].options.headers.Authorization, 'Bearer memory-token');
  assert.deepEqual(JSON.parse(calls[2].options.body), { uploadId: 'upload-123' });
});

test('failed S3 form upload never finalizes the archive record', async () => {
  const calls = [];
  const api = createArchiveApi('/api', async (url) => {
    calls.push(url);
    if (url.endsWith('/upload')) {
      return jsonResponse(201, {
        uploadId: 'upload-123',
        url: 'https://s3.example.com/presigned',
        fields: {},
      });
    }
    return new Response('S3 rejected upload', { status: 403 });
  });
  const file = new Blob(['%PDF-1.7'], { type: 'application/pdf' });
  file.name = 'report.pdf';

  await assert.rejects(api.upload('token', {
    originalName: file.name,
    contentType: file.type,
    size: file.size,
  }, file), /S3 rejected upload/);
  assert.deepEqual(calls, ['/api/upload', 'https://s3.example.com/presigned']);
});

test('local upload returns its ID so finalization can be retried separately', async () => {
  const calls = [];
  const api = createArchiveApi('/api', async (url, options) => {
    calls.push({ url, options });
    if (url === '/api/upload') {
      return jsonResponse(201, {
        uploadId: 'local-upload',
        url: '/api/local-upload?uploadId=local-upload',
      });
    }
    if (url.includes('/local-upload?')) {
      return new Response(null, { status: 204 });
    }
    return jsonResponse(201, { id: 'local-upload', state: 'available' });
  });
  const file = new Blob(['%PDF-1.7'], { type: 'application/pdf' });
  file.name = 'report.pdf';

  const uploadId = await api.upload('memory-token', {
    originalName: file.name,
    contentType: file.type,
    size: file.size,
  }, file);

  assert.equal(uploadId, 'local-upload');
  assert.equal(calls.length, 2);
  await api.completeUpload('memory-token', uploadId);
  assert.equal(calls[1].options.headers.Authorization, 'Bearer memory-token');
  assert.equal(calls[1].options.headers['Content-Type'], 'application/pdf');
  assert.equal(calls[1].options.body, file);
  assert.equal(calls[2].url, '/api/upload/complete');
});

test('finalization errors are reported rather than returning upload success', async () => {
  const api = createArchiveApi('/api', async (url) => {
    if (url.endsWith('/upload')) {
      return jsonResponse(201, {
        uploadId: 'upload-123',
        url: 'https://s3.example.com/presigned',
        fields: {},
      });
    }
    if (url === 'https://s3.example.com/presigned') {
      return new Response(null, { status: 204 });
    }
    return jsonResponse(502, { error: 'storage_unavailable' });
  });
  const file = new Blob(['%PDF-1.7'], { type: 'application/pdf' });
  file.name = 'report.pdf';

  const uploadId = await api.upload('token', {
    originalName: file.name,
    contentType: file.type,
    size: file.size,
  }, file);
  await assert.rejects(api.completeUpload('token', uploadId), /storage_unavailable/);
});

test('completion can be retried with the original upload ID without starting another upload', async () => {
  const calls = [];
  let completionAttempts = 0;
  const api = createArchiveApi('/api', async (url, options) => {
    calls.push({ url, options });
    if (url === '/api/upload') {
      return jsonResponse(201, {
        uploadId: 'original-upload-id',
        url: 'https://s3.example.com/presigned',
        fields: {},
      });
    }
    if (url === 'https://s3.example.com/presigned') {
      return new Response(null, { status: 204 });
    }
    if (url === '/api/upload/complete') {
      completionAttempts += 1;
      if (completionAttempts === 1) {
        return jsonResponse(502, { error: 'storage_unavailable' });
      }
      return jsonResponse(201, { id: 'archive-123', state: 'available' });
    }
    return jsonResponse(404, { error: 'unexpected_request' });
  });

  const file = new Blob(['%PDF-1.7'], { type: 'application/pdf' });
  file.name = 'report.pdf';
  const uploadId = await api.upload('token', {
    originalName: file.name,
    contentType: file.type,
    size: file.size,
  }, file);

  await assert.rejects(api.completeUpload('token', uploadId), /storage_unavailable/);
  assert.deepEqual(await api.completeUpload('token', uploadId), {
    id: 'archive-123',
    state: 'available',
  });
  assert.equal(calls.filter(({ url }) => url === '/api/upload').length, 1);
  assert.equal(calls.filter(({ url }) => url === 'https://s3.example.com/presigned').length, 1);
  assert.deepEqual(calls.map(({ url }) => url), [
    '/api/upload',
    'https://s3.example.com/presigned',
    '/api/upload/complete',
    '/api/upload/complete',
  ]);
  assert.deepEqual(calls
    .filter(({ url }) => url === '/api/upload/complete')
    .map(({ options }) => JSON.parse(options.body).uploadId), [
    'original-upload-id',
    'original-upload-id',
  ]);
});

test('upload returns its ID after transfer without attempting finalization', async () => {
  const calls = [];
  const api = createArchiveApi('/api', async (url) => {
    calls.push(url);
    if (url === '/api/upload') {
      return jsonResponse(201, {
        uploadId: 'same-upload-id',
        url: 'https://s3.example.com/presigned',
        fields: {},
      });
    }
    return new Response(null, { status: 204 });
  });
  const file = new Blob(['%PDF-1.7'], { type: 'application/pdf' });
  file.name = 'report.pdf';

  const uploadId = await api.upload('token', {
    originalName: file.name,
    contentType: file.type,
    size: file.size,
  }, file);

  assert.equal(uploadId, 'same-upload-id');
  assert.deepEqual(calls, ['/api/upload', 'https://s3.example.com/presigned']);
});

test('completion errors distinguish an unavailable original upload ID', async () => {
  const api = createArchiveApi('/api', async () =>
    jsonResponse(404, { error: 'archive_not_found' }));

  await assert.rejects(api.completeUpload('token', 'missing-upload-id'), (error) => {
    assert.match(error.message, /archive_not_found/);
    assert.equal(error.code, 'archive_not_found');
    assert.equal(error.status, 404);
    return true;
  });
});

test('list, download URL creation, and deletion use the authenticated archive contract', async () => {
  const calls = [];
  const api = createArchiveApi('/api/', async (url, options) => {
    calls.push({ url, options });
    if (url.endsWith('/deliverables')) {
      return jsonResponse(200, [{ id: 'id 1', state: 'available' }]);
    }
    if (url.includes('/download?')) {
      return jsonResponse(200, {
        url: 'https://s3.example.com/download?signature=temporary',
        expiresAt: '2026-10-22T12:05:00Z',
      });
    }
    return new Response(null, { status: 204 });
  });

  assert.deepEqual(await api.list('memory-token'), [{ id: 'id 1', state: 'available' }]);
  assert.equal((await api.createDownloadUrl('memory-token', 'id 1')).url,
    'https://s3.example.com/download?signature=temporary');
  assert.equal(await api.deleteArchive('memory-token', 'id 1'), undefined);
  assert.deepEqual(calls.map(({ url }) => url), [
    '/api/deliverables',
    '/api/download?filename=id%201',
    '/api/archive?filename=id%201',
  ]);
  for (const { options } of calls) {
    assert.equal(options.headers.Authorization, 'Bearer memory-token');
    assert.equal(options.credentials, 'omit');
  }
  assert.equal(calls[2].options.method, 'DELETE');
});

test('archive API rejects malformed successful responses and reports server errors', async () => {
  const emptyArchive = createArchiveApi('/api', async () => jsonResponse(200, []));
  assert.deepEqual(await emptyArchive.list('token'), []);

  const malformedList = createArchiveApi('/api', async () => jsonResponse(200, {}));
  await assert.rejects(malformedList.list('token'), /invalid .* response/i);

  const unavailable = createArchiveApi('/api', async () => jsonResponse(502, {
    error: 'storage_unavailable',
  }));
  await assert.rejects(unavailable.list('token'), /storage_unavailable/);
});
