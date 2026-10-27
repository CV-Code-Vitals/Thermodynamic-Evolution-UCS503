import test from 'node:test';
import assert from 'node:assert/strict';
import {createAccountClient} from '../../graph-visualizer/account.mjs';

test('personal account operations use a separate bearer and never send browser cookies', async () => {
  const calls = [];
  const client = createAccountClient('https://api.example/api/', async (url, options) => {
    calls.push({url, ...options});
    return new Response(JSON.stringify({ok:true}), {status:200});
  });
  await client.config();
  await client.exchange('one-time-ticket', 'browser-proof');
  await client.me('personal-session');
  await client.save('personal-session', 'https://github.com/owner/repo');
  await client.repositories('personal-session');
  await client.remove('personal-session', 'repo/id');
  await client.logout('personal-session');
  assert.equal(client.signInURL, 'https://api.example/api/auth/github/start');
  assert.deepEqual(calls.map(call => call.url), ['config','session','me','repositories','repositories','repositories/repo%2Fid','logout'].map(path => `https://api.example/api/account/${path}`));
  assert.ok(calls.every(call => call.credentials === 'omit'));
  assert.equal(calls[0].headers.Authorization, undefined);
  assert.equal(calls[1].headers.Authorization, undefined);
  assert.equal(calls[1].body, '{"ticket":"one-time-ticket","verifier":"browser-proof"}');
  assert.ok(calls.slice(2).every(call => call.headers.Authorization === 'Bearer personal-session'));
  assert.equal(calls[3].body, '{"url":"https://github.com/owner/repo"}');
  assert.equal(calls[5].method, 'DELETE');
  assert.equal(calls[6].method, 'POST');
});

test('personal account client reports expired sessions and repository limits', async () => {
  const expired = createAccountClient('/api', async () => new Response('{"error":"user_signin_required"}', {status:401}));
  await assert.rejects(expired.repositories('expired'), error => error.status === 401 && /session expired/.test(error.message));
  const limited = createAccountClient('/api', async () => new Response('{"error":"saved_repository_limit"}', {status:409}));
  await assert.rejects(limited.save('session','url'), /50 repositories/);
  const removed = createAccountClient('/api', async () => new Response(null, {status:204}));
  assert.equal(await removed.remove('session', 'id'), null);
});
