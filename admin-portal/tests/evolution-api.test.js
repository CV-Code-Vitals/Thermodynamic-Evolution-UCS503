import assert from 'node:assert/strict';
import test from 'node:test';
import { createEvolutionApi } from '../src/evolution-api.js';

function responseWithEvents(events) {
  const encoder = new TextEncoder();
  return new Response(encoder.encode(events), {
    status: 200,
    headers: { 'Content-Type': 'text/event-stream' },
  });
}

const payload = {
  repo_url: 'https://github.com/example/project',
  file_path: 'main.go',
  function_name: 'ProcessData',
  config: { max_steps: 20 },
};

test('keeps archive and analysis API bases separate', async () => {
  let requestedURL;
  const api = createEvolutionApi('https://analysis.example.com', async (url) => {
    requestedURL = url;
    return responseWithEvents('event: complete\ndata: {}\n\n');
  });

  await api.start({ token: 'admin-session-token', payload });

  assert.equal(requestedURL, 'https://analysis.example.com/api/evolve');
});

test('sends the in-memory bearer only to Evolution', async () => {
  let request;
  const api = createEvolutionApi('https://analysis.example.com', async (url, options) => {
    request = { url, options };
    return responseWithEvents('event: complete\ndata: {}\n\n');
  });

  await api.start({ token: 'admin-session-token', payload });

  assert.equal(request.options.headers.Authorization, 'Bearer admin-session-token');
  assert.equal(request.options.headers['X-API-Token'], undefined);
  assert.equal(request.options.headers['X-Internal-Token'], undefined);
  assert.equal(JSON.parse(request.options.body).repo_url, payload.repo_url);
  assert.equal(JSON.stringify(request.options.body).includes('admin-session-token'), false);
});

test('surfaces HTTP and SSE errors', async () => {
  const httpApi = createEvolutionApi('https://analysis.example.com', async () =>
    new Response(JSON.stringify({ error: 'not_authorized' }), { status: 403 }));

  await assert.rejects(
    httpApi.start({ token: 'admin-session-token', payload }),
    /403.*not_authorized/,
  );

  const streamApi = createEvolutionApi('https://analysis.example.com', async () =>
    responseWithEvents('event: error\ndata: {"error":"worker failed"}\n\n'));
  const received = [];

  await assert.rejects(
    streamApi.start({
      token: 'admin-session-token',
      payload,
      onEvent: (event) => received.push(event),
    }),
    /worker failed/,
  );
  assert.equal(received[0].type, 'error');
  assert.equal(received[0].data.error, 'worker failed');
});

test('aborts the Evolution stream', async () => {
  let streamController;
  const api = createEvolutionApi('https://analysis.example.com', async (_url, options) => {
    const body = new ReadableStream({
      start(controller) {
        streamController = controller;
        controller.enqueue(new TextEncoder().encode('event: step\ndata: {"step":1}\n\n'));
        options.signal.addEventListener('abort', () => {
          controller.error(new DOMException('The operation was aborted.', 'AbortError'));
        }, { once: true });
      },
    });
    return new Response(body, { status: 200 });
  });
  const controller = new AbortController();
  const running = api.start({ token: 'admin-session-token', payload, signal: controller.signal });

  await new Promise((resolve) => setImmediate(resolve));
  controller.abort();

  await assert.rejects(running, { name: 'AbortError' });
  assert.ok(streamController);
});
