import test from 'node:test';
import assert from 'node:assert/strict';
import { repositoryFiles, buildGraphModel } from '../../graph-visualizer/model.mjs';

const sample = { scanned_directory: '/scan/repo', file_reports: [
  { file_path: '/scan/repo/api/main.go', language: 'Go', total_entropy: 20, hotspots: [
    { function_name: 'Run', line_number: 10, entropy_score: 10 },
    { function_name: 'Run', line_number: 20, entropy_score: 10 },
  ] },
  { file_path: '/scan/repo/backend/main.go', language: 'Go', total_entropy: 10, hotspots: [
    { function_name: 'Run', line_number: 10, entropy_score: 10 },
  ] },
  { file_path: '/scan/repo/ui/Card.tsx', language: 'TSX', total_entropy: 5, hotspots: [
    { function_name: 'Card', line_number: 2, entropy_score: 5 },
  ] },
  { file_path: '/scan/repo/main.novel', language: 'novel (text)', analysis_mode: 'text', hotspots: [] },
] };

test('graph preserves full paths and same-named functions at distinct source locations', () => {
  const graph = buildGraphModel(repositoryFiles(sample), {showFunctions:true});
  assert.equal(graph.nodes.length, 8);
  assert.equal(graph.links.length, 4);
  assert.equal(new Set(graph.nodes.map(node => node.id)).size, 8);
  assert.deepEqual(graph.filtered.map(file => file.path), ['api/main.go', 'backend/main.go', 'main.novel', 'ui/Card.tsx']);
});

test('overview collapses functions; search, language filters and focus show the relevant graph', () => {
  const files = repositoryFiles(sample);
  assert.equal(buildGraphModel(files).nodes.length, 4);
  const match = buildGraphModel(files, {query:'card', language:'TSX', focusPath:'ui/Card.tsx'});
  assert.deepEqual(match.nodes.map(node => node.label), ['ui/Card.tsx', 'Card']);
  assert.equal(match.links[0].source.id, 'file:ui/Card.tsx');
  assert.equal(buildGraphModel(files, {query:'missing'}).nodes.length, 0);
  assert.equal(buildGraphModel(files, {language:'novel (text)'}).nodes[0].file.analysis_mode, 'text');
});

test('large graphs have deterministic non-overlapping cards and valid membership links', () => {
  const files = repositoryFiles({file_reports:Array.from({length:60}, (_, file) => ({
    file_path:`dir${file}/main.go`, language:'Go', hotspots:Array.from({length:file % 20 + 1}, (_, fn) => ({
      function_name:'Run', line_number:fn + 1, entropy_score:fn,
    })),
  }))});
  const graph = buildGraphModel(files, {showFunctions:true});
  assert.deepEqual(graph, buildGraphModel(files, {showFunctions:true}));
  const ids = new Set(graph.nodes.map(node => node.id));
  for (const link of graph.links) {
    assert.ok(ids.has(link.source.id) && ids.has(link.target.id));
    assert.equal(link.source.file.path, link.target.file.path);
  }
  for (let i = 0; i < graph.nodes.length; i++) {
    const a = graph.nodes[i];
    assert.ok(a.x >= 0 && a.y >= 0 && a.x + a.width <= graph.width && a.y + a.height <= graph.height);
    for (const b of graph.nodes.slice(i + 1)) {
      assert.ok(a.x + a.width <= b.x || b.x + b.width <= a.x || a.y + a.height <= b.y || b.y + b.height <= a.y);
    }
  }
});

test('file focus includes connected imports and resolves calls by byte location', () => {
  const files = repositoryFiles({file_reports:[
    {file_path:'a.ts', language:'TypeScript', hotspots:[
      {function_name:'caller', byte_range:[0,20]}, {function_name:'callee', byte_range:[21,40]},
    ]},
    {file_path:'b.ts', language:'TypeScript', hotspots:[]},
    {file_path:'unrelated.ts', language:'TypeScript', hotspots:[]},
  ]});
  const relationships = [
    {kind:'imports', source_file:'a.ts', target_file:'b.ts'},
    {kind:'calls', source_file:'a.ts', target_file:'a.ts', source_start_byte:0, target_start_byte:21},
  ];
  const graph = buildGraphModel(files, {focusPath:'a.ts', relationships});
  assert.equal(graph.fileCount, 2);
  assert.deepEqual(graph.links.map(edge => edge.kind), ['contains', 'contains', 'imports', 'calls']);
  assert.equal(graph.links.at(-1).source.fn.function_name, 'caller');
  assert.equal(graph.links.at(-1).target.fn.function_name, 'callee');
  assert.equal(buildGraphModel(files, {focusPath:'a.ts', relationships, showImports:false}).fileCount, 1);
});

test('folder overview preserves file counts and aggregates cross-folder imports', () => {
  const files = repositoryFiles(sample);
  const relationships = [
    {kind:'imports', source_file:'api/main.go', target_file:'backend/main.go'},
    {kind:'imports', source_file:'api/main.go', target_file:'ui/Card.tsx'},
  ];
  const graph = buildGraphModel(files, {groupFolders:true, relationships});
  assert.equal(graph.fileCount, 4);
  assert.equal(graph.nodes.length, 4);
  assert.ok(graph.nodes.every(node => node.type === 'folder'));
  assert.equal(graph.links.length, 2);
  const folder = buildGraphModel(files, {groupFolders:true, folderPath:'api', relationships});
  assert.deepEqual(folder.nodes.map(node => node.label), ['api/main.go']);
});
