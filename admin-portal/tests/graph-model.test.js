import test from 'node:test';
import assert from 'node:assert/strict';
import { repositoryFiles, buildGraphModel } from '../../graph-visualizer/model.mjs';
import { routeGraphEdges } from '../../graph-visualizer/routes.mjs';

function assertClearConnections({nodes, links}) {
  const onBorder = (node, [x, y]) => x >= node.x && x <= node.x + node.width && y >= node.y && y <= node.y + node.height
    && (x === node.x || x === node.x + node.width || y === node.y || y === node.y + node.height);
  for (const link of links) {
    assert.ok(onBorder(link.source, link.points[0]), 'Connection must start on its source card');
    assert.ok(onBorder(link.target, link.points.at(-1)), 'Arrowhead must end on its destination card');
    for (let i = 1; i < link.points.length; i++) {
      const [a, b] = [link.points[i - 1], link.points[i]];
      assert.ok(a[0] === b[0] || a[1] === b[1], 'Connection should use clear horizontal and vertical channels');
      assert.notDeepEqual(a, b, 'Arrow direction must have a nonzero final segment');
      for (const node of nodes) {
        const crosses = a[1] === b[1]
          ? a[1] > node.y && a[1] < node.y + node.height && Math.max(a[0], b[0]) > node.x && Math.min(a[0], b[0]) < node.x + node.width
          : a[0] > node.x && a[0] < node.x + node.width && Math.max(a[1], b[1]) > node.y && Math.min(a[1], b[1]) < node.y + node.height;
        assert.equal(crosses, false, `${link.source.id} → ${link.target.id} crosses card ${node.id}`);
      }
    }
  }
}

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
  assertClearConnections(graph);
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
  assertClearConnections(graph);
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
  assertClearConnections(graph);
  const folder = buildGraphModel(files, {groupFolders:true, folderPath:'api', relationships});
  assert.deepEqual(folder.nodes.map(node => node.label), ['api/main.go']);
});

test('arrows route around intervening cards and stay clear in both directions', () => {
  const nodes = [
    {id:'source', x:30, y:30, width:100, height:80},
    {id:'obstacle', x:180, y:30, width:100, height:80},
    {id:'target', x:330, y:30, width:100, height:80},
  ];
  const links = [
    {source:nodes[0], target:nodes[2], kind:'imports'},
    {source:nodes[2], target:nodes[0], kind:'imports'},
    {source:nodes[0], target:nodes[0], kind:'calls'},
  ];
  const routed = routeGraphEdges(nodes, links);
  assertClearConnections({nodes, links:routed});
  assert.ok(routed[0].points.some(([, y]) => y < 30 || y > 110), 'Connection must go above or below the intervening card');
  assert.notDeepEqual(routed[0].points, [...routed[1].points].reverse(), 'Opposite arrows need distinguishable routes');
  assert.ok(routed[2].points.length >= 4, 'Recursive calls need a visible loop outside the card');
  assert.deepEqual(routed, routeGraphEdges(nodes, links), 'Routing must stay stable when the graph redraws');
});

test('membership and call arrows avoid other functions across multiple rows and files', () => {
  const files = repositoryFiles({file_reports:Array.from({length:4}, (_, file) => ({
    file_path:`src/file${file}.go`, language:'Go', hotspots:Array.from({length:8}, (_, fn) => ({
      function_name:`fn${fn}`, byte_range:[fn * 10, fn * 10 + 9], line_number:fn + 1,
    })),
  }))});
  const relationships = files.flatMap((file, i) => file.hotspots.map((fn, j) => ({
    kind:'calls', source_file:file.path, source_start_byte:fn.byte_range[0],
    target_file:files[(i + 1) % files.length].path, target_start_byte:((j + 3) % 8) * 10,
  })));
  const graph = buildGraphModel(files, {showFunctions:true, relationships});
  assert.equal(graph.links.filter(link => link.kind === 'calls').length, 32);
  assert.equal(graph.links.filter(link => link.kind === 'contains').length, 32);
  assertClearConnections(graph);
});
