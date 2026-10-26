import test from 'node:test';
import assert from 'node:assert/strict';
import { readEvolutionStream, scanEvolutionRepository } from '../src/evolution-api.js';

const streamResponse = (text) => new Response(new ReadableStream({
  start(controller) {
    // Every byte is a separate network chunk, including multibyte characters.
    for (const byte of new TextEncoder().encode(text)) controller.enqueue(Uint8Array.of(byte));
    controller.close();
  },
}), { headers: { 'Content-Type': 'text/event-stream' } });

test('evolution handles split event names, data, Unicode, CRLF, and multiple events', async () => {
  const events = [];
  await readEvolutionStream(streamResponse(': keepalive\r\nevent: step\r\ndata: {"step":1,"code_patch":"Δ"}\r\n\r\nevent: complete\r\ndata: {"final_energy":0}\r\n\r\n'),
    (event, data) => events.push([event, data]));
  assert.deepEqual(events, [['step', { step: 1, code_patch: 'Δ' }], ['complete', { final_energy: 0 }]]);
});

test('evolution reports HTTP, server event, malformed JSON, and truncated stream errors', async () => {
  await assert.rejects(readEvolutionStream(new Response('{"error":"unauthorized"}', { status: 401 }), () => {}), /unauthorized/);
  await assert.rejects(readEvolutionStream(streamResponse('event: error\ndata: {"error":"Sandbox unavailable"}\n\n'), () => {}), /Sandbox unavailable/);
  await assert.rejects(readEvolutionStream(streamResponse('event: step\ndata: {bad}\n\n'), () => {}), SyntaxError);
  await assert.rejects(readEvolutionStream(streamResponse('event: step\ndata: {"step":1}\n\n'), () => {}), /ended before completion/);
  await assert.rejects(readEvolutionStream(new Response('{}'), () => {}), /did not return an event stream/);
});

test('scan and evolution explain when the analysis service is offline', async (t) => {
  const unavailable = () => Response.json({ error: 'analysis_unavailable' }, { status: 503 });
  const message = 'The analysis service is unavailable. Ask the administrator to start it, then retry.';
  t.mock.method(globalThis, 'fetch', async () => unavailable());
  await assert.rejects(scanEvolutionRepository('/api', 'test-token', 'https://github.com/example/repo'), { message });
  await assert.rejects(readEvolutionStream(unavailable(), () => {}), { message });
  await assert.rejects(readEvolutionStream(streamResponse('event: error\ndata: {"error":"analysis_unavailable"}\n\n'), () => {}), { message });
});

test('cloud scans use the separate streaming gateway and retain selectable targets', async (t) => {
  let request;
  t.mock.method(globalThis, 'fetch', async (url, options) => {
    request = {url,options};
    return Response.json({report:{files_analyzed:1},workspace_id:'snapshot-id',targets:[]});
  });
  const scan = await scanEvolutionRepository('https://archive.example/api', 'test-token', 'https://github.com/example/repo', undefined, 'https://analysis.example/prod');
  assert.equal(request.url, 'https://analysis.example/prod/api/scan');
  assert.equal(request.options.headers.Authorization,'Bearer test-token');
  assert.deepEqual(JSON.parse(request.options.body), {repo_url:'https://github.com/example/repo',retain_workspace:true});
  assert.equal(scan.repo_url,'https://github.com/example/repo');
});
