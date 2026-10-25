import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import vm from 'node:vm';

const html = readFileSync(new URL('../../graph-visualizer/index.html', import.meta.url), 'utf8');
const start = html.indexOf('function buildData()');
const end = html.indexOf('\nconst canvas =', start);
assert.ok(start >= 0 && end > start, 'Graph data builder must remain independently testable.');

test('graph preserves files with identical basenames and methods with identical names', () => {
  const context = vm.createContext({
    REPORT: { file_reports: [
      { file_path: 'api/main.go', total_entropy: 20, hotspots: [
        { function_name: 'Run', line_number: 10, entropy_score: 10 },
        { function_name: 'Run', line_number: 20, entropy_score: 10 },
      ] },
      { file_path: 'backend/main.go', total_entropy: 10, hotspots: [
        { function_name: 'Run', line_number: 10, entropy_score: 10 },
      ] },
    ] },
    allNodesMap: new Map(), snapshots: [], allNodes: [], PHASES: [], targetPhaseCount: 5,
    nodeRadius: () => 1,
    document: { getElementById: () => null },
  });
  vm.runInContext(`${html.slice(start, end)}\nbuildData();`, context);
  const complete = context.snapshots.at(-1);
  assert.equal(complete.nodes.length, 5);
  assert.equal(complete.nodes.filter(node => node.type === 'file').length, 2);
  assert.equal(complete.nodes.filter(node => node.type === 'hotspot').length, 3);
  assert.equal(new Set(complete.nodes.map(node => node.id)).size, 5);
  assert.equal(context.PHASES.at(-1).label, 'Complete repository scan');
});
