import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

const graphHTML = await readFile(new URL('../graph-visualizer/index.html', import.meta.url), 'utf8');
const workflow = await readFile(new URL('../.github/workflows/deploy-frontend.yml', import.meta.url), 'utf8');

test('TestGraphVisualizer_UsesConfiguredAnalysisAPIWithoutFallback', () => {
  assert.ok(/window\.UCS503_ANALYSIS_API_BASE/.test(graphHTML), 'scanner reads the public window configuration');
  assert.ok(/config\.js/.test(graphHTML), 'scanner loads its deployment configuration');
  assert.ok(/analysis API.*not configured/i.test(graphHTML), 'missing configuration is visible');
  assert.ok(/\/api\/scan/.test(graphHTML), 'scanner calls the anonymous scan route');
  assert.ok(!/CUSTOM_SCAN_API|scanClientSide|api\.github\.com|raw\.githubusercontent\.com/.test(graphHTML), 'scanner has no custom or client-side fallback');
  assert.ok(/vars\.UCS503_ANALYSIS_API_BASE/.test(workflow), 'Pages build reads the public endpoint variable');
  assert.ok(/_site\/graph-visualizer\/config\.js/.test(workflow), 'Pages build emits the graph config script');
  assert.ok(!/secrets\.(?:UCS503_ANALYSIS_API_BASE|ANALYSIS_API_TOKEN)/.test(workflow), 'workflow does not place credentials in frontend config');
});
