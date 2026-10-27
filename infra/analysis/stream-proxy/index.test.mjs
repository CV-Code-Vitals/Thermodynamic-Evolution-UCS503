import assert from 'node:assert/strict';
import { Writable } from 'node:stream';
import test from 'node:test';
import { createProxyHandler } from './index.mjs';

const snapshotID = '0123456789abcdef0123456789abcdef';

function event({
  path = '/api/scan',
  body = { repo_url: 'https://github.com/octocat/repo' },
  headers = {},
  sourceIp = '8.8.8.8',
  authorizer = {},
  rawBody,
  isBase64Encoded = false,
} = {}) {
  return {
    path,
    httpMethod: 'POST',
    body: rawBody ?? JSON.stringify(body),
    isBase64Encoded,
    headers,
    requestContext: {
      identity: { sourceIp },
      authorizer,
    },
  };
}

function createResponseStream() {
  const chunks = [];
  const output = new Writable({
    write(chunk, _encoding, callback) {
      chunks.push(Buffer.from(chunk));
      callback();
    },
  });
  output.metadata = undefined;
  return {
    output,
    body: () => Buffer.concat(chunks).toString(),
  };
}

function dependencies(overrides = {}) {
  return {
    analysisBaseURL: 'http://127.0.0.1:8080',
    frontendOrigin: 'https://zylus08.github.io',
    stager: async () => ({ snapshot_id: snapshotID }),
    getScanToken: async () => 'ssm-internal-token',
    fetchImpl: async () => new Response('{"ok":true}', { status: 200 }),
    responseStreamFactory: (stream, metadata) => {
      stream.metadata = metadata;
      return stream;
    },
    logger: () => {},
    ...overrides,
  };
}

async function invoke(proxy, requestEvent) {
  const stream = createResponseStream();
  await proxy(requestEvent, stream.output, {});
  return { metadata: stream.output.metadata, body: stream.body() };
}

test('rejects oversized and malformed request bodies', async () => {
  const proxy = createProxyHandler(dependencies({ maxRequestBytes: 64 }));
  const oversized = await invoke(proxy, event({ rawBody: 'x'.repeat(65) }));
  assert.equal(oversized.metadata.statusCode, 413);

  const malformed = await invoke(proxy, event({ rawBody: '{"repo_url":' }));
  assert.equal(malformed.metadata.statusCode, 400);
});

test('overwrites client-supplied internal headers and uses gateway identity', async () => {
  let upstreamRequest;
  const proxy = createProxyHandler(dependencies({
    fetchImpl: async (url, options) => {
      upstreamRequest = { url, options };
      return new Response('{"ok":true}', { status: 200 });
    },
  }));
  const response = await invoke(proxy, event({
    headers: {
      authorization: 'Bearer browser-secret',
      'x-api-token': 'attacker-token',
      'x-trusted-client-ip': '203.0.113.7',
      'x-forwarded-for': '203.0.113.7',
    },
    sourceIp: '8.8.8.8',
  }));

  assert.equal(response.metadata.statusCode, 200);
  assert.equal(response.metadata.headers['Access-Control-Allow-Origin'], 'https://zylus08.github.io');
  assert.equal(response.metadata.headers['Access-Control-Allow-Headers'], 'Authorization,Content-Type');
  assert.equal(upstreamRequest.options.headers['X-API-Token'], 'ssm-internal-token');
  assert.equal(upstreamRequest.options.headers['X-Trusted-Client-IP'], '8.8.8.8');
  assert.equal(upstreamRequest.options.headers.Authorization, undefined);
  assert.equal(upstreamRequest.options.headers['X-Forwarded-For'], undefined);
  assert.deepEqual(JSON.parse(upstreamRequest.options.body), { snapshot_id: snapshotID });
});

test('uses authorizer token for Evolution and SSM token for anonymous scan', async () => {
  let scanTokenReads = 0;
  const upstreamRequests = [];
  const proxy = createProxyHandler(dependencies({
    getScanToken: async () => {
      scanTokenReads += 1;
      return 'ssm-internal-token';
    },
    fetchImpl: async (url, options) => {
      upstreamRequests.push({ url, options });
      return new Response('event: complete\ndata: {}\n\n', {
        status: 200,
        headers: { 'content-type': 'text/event-stream' },
      });
    },
  }));

  await invoke(proxy, event());
  await invoke(proxy, event({
    path: '/api/evolve',
    body: {
      repo_url: 'https://github.com/octocat/repo',
      file_path: 'main.go',
      function_name: 'main',
    },
    authorizer: { analysisApiToken: 'admin-authorizer-token' },
  }));

  assert.equal(scanTokenReads, 1);
  assert.equal(upstreamRequests[0].options.headers['X-API-Token'], 'ssm-internal-token');
  assert.equal(upstreamRequests[1].options.headers['X-API-Token'], 'admin-authorizer-token');
  assert.equal(upstreamRequests[1].options.headers['X-Trusted-Client-IP'], undefined);
  assert.deepEqual(JSON.parse(upstreamRequests[1].options.body), {
    snapshot_id: snapshotID,
    file_path: 'main.go',
    function_name: 'main',
  });
});

test('preserves streamed SSE frames and upstream status', async () => {
  const frames = ['event: step\ndata: {"step":1}\n\n', 'event: complete\ndata: {}\n\n'];
  const proxy = createProxyHandler(dependencies({
    fetchImpl: async () => new Response(new ReadableStream({
      start(controller) {
        for (const frame of frames) controller.enqueue(new TextEncoder().encode(frame));
        controller.close();
      },
    }), {
      status: 206,
      headers: {
        'content-type': 'text/event-stream',
        'cache-control': 'no-cache',
      },
    }),
  }));
  const response = await invoke(proxy, event({
    path: '/api/evolve',
    body: { repo_url: 'https://github.com/octocat/repo', file_path: 'main.go' },
    authorizer: { analysisApiToken: 'admin-token' },
  }));

  assert.equal(response.metadata.statusCode, 206);
  assert.equal(response.metadata.headers['Content-Type'], 'text/event-stream');
  assert.equal(response.body, frames.join(''));
});

test('cancels upstream when the client disconnects', async () => {
  let upstreamSignal;
  let fetchStarted;
  const started = new Promise((resolve) => { fetchStarted = resolve; });
  const proxy = createProxyHandler(dependencies({
    fetchImpl: async (_url, options) => {
      upstreamSignal = options.signal;
      fetchStarted();
      return new Promise((_resolve, reject) => {
        upstreamSignal.addEventListener('abort', () => reject(upstreamSignal.reason), { once: true });
      });
    },
  }));
  const stream = createResponseStream();
  const pending = proxy(event({
    path: '/api/evolve',
    body: { repo_url: 'https://github.com/octocat/repo', file_path: 'main.go' },
    authorizer: { analysisApiToken: 'admin-token' },
  }), stream.output, {});

  await started;
  stream.output.destroy();
  await pending;
  assert.equal(upstreamSignal.aborted, true);
});

test('does not log request bodies or credentials', async () => {
  const logEntries = [];
  const proxy = createProxyHandler(dependencies({
    logger: (...args) => logEntries.push(args.join(' ')),
    getScanToken: async () => 'internal-secret',
  }));
  await invoke(proxy, event({
    rawBody: JSON.stringify({ repo_url: 'https://github.com/octocat/private-name' }),
    headers: { authorization: 'Bearer user-secret', 'x-api-token': 'internal-secret' },
  }));

  const logs = logEntries.join('\n');
  assert.equal(logs.includes('private-name'), false);
  assert.equal(logs.includes('user-secret'), false);
  assert.equal(logs.includes('internal-secret'), false);
});

test('does not call analysis host when source staging fails', async () => {
  let upstreamCalls = 0;
  const proxy = createProxyHandler(dependencies({
    stager: async () => { throw new Error('stager unavailable'); },
    fetchImpl: async () => {
      upstreamCalls += 1;
      return new Response('{}');
    },
  }));
  const response = await invoke(proxy, event());

  assert.equal(response.metadata.statusCode, 502);
  assert.equal(upstreamCalls, 0);
});

test('forwards the portal scan contract and source fingerprint through staging', async () => {
  const bodies = [];
  const proxy = createProxyHandler(dependencies({
    fetchImpl: async (_, options) => {
      bodies.push(JSON.parse(options.body));
      return new Response('{}', {status:200});
    },
  }));
  await invoke(proxy, event({body:{repo_url:'https://github.com/octocat/repo',retain_workspace:true}}));
  await invoke(proxy, event({path:'/api/evolve',body:{repo_url:'https://github.com/octocat/repo',file_path:'main.go',source_hash:'a'.repeat(64)},authorizer:{analysisApiToken:'admin-token'}}));
  assert.deepEqual(bodies[0], {snapshot_id:snapshotID,retain_workspace:true});
  assert.deepEqual(bodies[1], {snapshot_id:snapshotID,file_path:'main.go',source_hash:'a'.repeat(64)});
});
