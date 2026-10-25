import assert from 'node:assert/strict';
import { mkdir, mkdtemp, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { createArchiveApi } from '../admin-portal/src/archive-api.js';
import { readEvolutionStream, scanEvolutionRepository } from '../admin-portal/src/evolution-api.js';

const base = process.env.PORTAL_API_BASE?.replace(/\/$/, '');
const passkey = process.env.PORTAL_PASSKEY;
const origin = process.env.PORTAL_ORIGIN || 'https://cv-code-vitals.github.io';
assert.ok(base && passkey, 'Set PORTAL_API_BASE including /api and PORTAL_PASSKEY in your terminal.');
const nativeFetch = globalThis.fetch;
globalThis.fetch = (url, options = {}) => nativeFetch(url, {
  ...options,
  headers: { ...options.headers, Origin: origin },
  signal: options.signal || AbortSignal.timeout(600000),
});

const preflight = await fetch(`${base}/analysis/scan`, {
  method: 'OPTIONS',
  headers: { 'Access-Control-Request-Method': 'POST', 'Access-Control-Request-Headers': 'authorization,content-type' },
});
assert.ok(preflight.ok, `Portal origin rejected: ${preflight.status}`);
assert.equal(preflight.headers.get('access-control-allow-origin'), origin);
const unauthorized = await fetch(`${base}/analysis/health`);
assert.equal(unauthorized.status, 401, 'Analysis must reject anonymous users.');

const api = createArchiveApi(base);
const { token } = await api.login(passkey);
console.log('PASS: portal origin, anonymous rejection, and login');
const headers = { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' };
const health = await fetch(`${base}/analysis/health`, { headers });
assert.ok(health.ok, `Analysis service unavailable: ${health.status}`);
const analysisStatus = await health.json();
assert.equal(analysisStatus.sandbox_available, true, 'Docker sandbox is not available.');

let uploadId;
try {
  const contents = new File(['%PDF-1.7\nDisposable portal smoke test\n%%EOF\n'], 'smoke-test.pdf', { type: 'application/pdf' });
  uploadId = await api.upload(token, {
    originalName: contents.name, contentType: contents.type, size: contents.size,
    title: 'Disposable smoke test', version: 'verification', date: new Date().toISOString().slice(0, 10), summary: 'Removed automatically after verification.',
  }, contents);
  await api.completeUpload(token, uploadId);
  const entries = await api.list(token);
  assert.ok(entries.some(entry => (entry.id || entry.filename) === uploadId));
  const download = await api.createDownloadUrl(token, uploadId);
  const options = new URL(download.url).origin === new URL(base).origin ? { headers } : {};
  const downloaded = await fetch(download.url, options);
  assert.ok(downloaded.ok);
  assert.equal(await downloaded.text(), await contents.text());
} finally {
  if (uploadId) await api.deleteArchive(token, uploadId);
}
console.log('PASS: archive upload, finalization, listing, byte-exact download, and deletion');

const repository = process.env.SMOKE_REPOSITORY_URL || 'https://github.com/CV-Code-Vitals/Thermodynamic-Evolution-UCS503';
const scan = await scanEvolutionRepository(base, token, repository, AbortSignal.timeout(180000));
const functionName = process.env.SMOKE_FUNCTION || 'EvaluateRiskEngine';
const target = scan.targets.find(item => item.function_name === functionName);
assert.ok(target, `No Go function named ${functionName} in scan.`);
console.log(`PASS: scan analyzed ${scan.report.files_analyzed} files and selected ${target.function_name}`);

let result;
let steps = 0;
const response = await fetch(`${base}/analysis/evolve`, {
  method: 'POST', headers,
  body: JSON.stringify({ ...target, workspace_id: scan.workspace_id,
    config: { max_steps: 4, initial_temp: 100, cooling_rate: 0.85, enable_duplicate_detection: true } }),
});
await readEvolutionStream(response, (event, data) => {
  if (event === 'step') { steps++; console.log(`Evolution step ${data.step}: ${data.reason}`); }
  if (event === 'complete') result = data;
});
assert.ok(result && steps > 0, 'No completed evolution with progress events.');
assert.equal(result.total_steps, steps);
assert.ok(result.correctness_passed > 0, 'No candidate passed real sandbox tests.');
assert.ok(result.best_improved && result.patch, 'The demonstration did not produce an exportable improvement.');
assert.ok(result.final_energy < result.baseline_energy);
const output = process.env.SMOKE_OUTPUT_DIR || await mkdtemp(join(tmpdir(), 'ucs503-evolution-'));
await mkdir(output, { recursive: true });
await writeFile(join(output, 'result.json'), JSON.stringify({ repository, target, ...result }, null, 2));
await writeFile(join(output, 'result.patch'), result.patch);
console.log(`PASS: real sandbox evolution reduced structural score ${result.baseline_energy.toFixed(4)} to ${result.final_energy.toFixed(4)}`);
console.log(`Saved reviewable patch and report in ${output}`);
