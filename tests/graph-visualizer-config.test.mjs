import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const graphHTML = await readFile(new URL('../graph-visualizer/graph.mjs', import.meta.url), 'utf8');
const portal = await readFile(new URL('../admin-portal/src/components/AdminPortal.jsx', import.meta.url), 'utf8');
const workflow = await readFile(new URL('../.github/workflows/deploy-frontend.yml', import.meta.url), 'utf8');
const publicPage = await readFile(new URL('../graph-visualizer/index.html', import.meta.url), 'utf8');
const landingPage = await readFile(new URL('../index.html', import.meta.url), 'utf8');

test('standalone graph offers public scanning with the configured analysis gateway', () => {
  assert.doesNotMatch(graphHTML, /location\.replace/);
  assert.match(publicPage, /id="scan-form"/);
  assert.match(landingPage, /href="graph-visualizer\/"/);
  assert.match(graphHTML, /scanPublicRepository\(scanConfig/);
  assert.doesNotMatch(graphHTML, /CUSTOM_SCAN_API|scanClientSide|api\.github\.com|raw\.githubusercontent\.com/);
  assert.match(portal, /VITE_ANALYSIS_API_BASE/);
  assert.match(workflow, /VITE_ANALYSIS_API_BASE:.*vars\.UCS503_ANALYSIS_API_BASE/);
  assert.match(workflow, /cp -R admin-portal\/dist\/graph-visualizer\/\. _site\/graph-visualizer\//);
  assert.doesNotMatch(workflow, /secrets\.(?:UCS503_ANALYSIS_API_BASE|ANALYSIS_API_TOKEN)/);
});
