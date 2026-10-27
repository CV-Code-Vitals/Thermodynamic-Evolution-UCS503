import { repositoryFiles, buildGraphModel } from './model.mjs';
import scanConfig from './config.mjs';
import { scanPublicRepository } from './scan.mjs';
import { scoreBand, energyBreakdown, summarizeRepository } from './measurements.mjs';

const $ = id => document.getElementById(id);
const canvas = $('graph'), ctx = canvas.getContext('2d'), stage = $('stage');
let files = [], relationships = [], model = buildGraphModel([]), focusPath = '', folderPath = '', selected = null;
let camera = { x: 0, y: 0, zoom: 1 }, width = 1, height = 1, pending = false, pointer = null;
const number = value => Number(value || 0).toFixed(2);
const nodeColors = new WeakMap();
const color = node => {
  if (!nodeColors.has(node)) nodeColors.set(node, scoreBand(node.type === 'function' ? node.score
    : summarizeRepository(node.type === 'folder' ? node.files : [node.file]).peak).color);
  return nodeColors.get(node);
};

function scheduleDraw() {
  if (pending) return;
  pending = true;
  requestAnimationFrame(() => { pending = false; draw(); });
}
function fit() {
  camera.zoom = Math.min(1.15, Math.max(0.02, Math.min((width - 36) / model.width, (height - 36) / model.height)));
  camera.x = (width - model.width * camera.zoom) / 2;
  camera.y = (height - model.height * camera.zoom) / 2;
  scheduleDraw();
}
function readableView() {
  fit();
  if (model.nodes.length && camera.zoom < 0.7) {
    camera = {zoom:0.7, x:0, y:0};
    scheduleDraw();
  }
}
function zoomAt(factor, x = width / 2, y = height / 2) {
  const next = Math.max(0.02, Math.min(3, camera.zoom * factor));
  camera.x = x - (x - camera.x) * next / camera.zoom;
  camera.y = y - (y - camera.y) * next / camera.zoom;
  camera.zoom = next;
  scheduleDraw();
}
function draw() {
  const ratio = Math.min(window.devicePixelRatio || 1, 2);
  ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
  ctx.clearRect(0, 0, width, height);
  ctx.translate(camera.x, camera.y);
  ctx.scale(camera.zoom, camera.zoom);
  const connected = new Set(selected ? [selected.id] : []);
  const arrowheads = [];
  const active = link => !selected || link.source.id === selected.id || link.target.id === selected.id;
  if (selected) for (const link of model.links) if (active(link)) {
    connected.add(link.source.id); connected.add(link.target.id);
  }
  ctx.lineJoin = 'round'; ctx.lineCap = 'round';
  for (const foreground of selected ? [false, true] : [true]) for (const link of model.links) {
    if (active(link) !== foreground) continue;
    const {kind, points} = link;
    const ink = kind === 'imports' ? '#ffd700' : kind === 'calls' ? '#f59e0b' : '#999999';
    ctx.globalAlpha = foreground ? (kind === 'contains' && !selected ? .6 : 1) : .1;
    ctx.lineWidth = (selected && foreground ? 2.6 : kind === 'contains' ? 1.3 : 1.8) / camera.zoom;
    ctx.setLineDash(kind === 'calls' ? [6 / camera.zoom, 4 / camera.zoom] : []);
    ctx.beginPath(); ctx.moveTo(...points[0]);
    for (const point of points.slice(1)) ctx.lineTo(...point);
    if (selected && foreground) {
      ctx.strokeStyle = '#050505'; ctx.lineWidth += 4 / camera.zoom; ctx.stroke(); ctx.lineWidth -= 4 / camera.zoom;
    }
    ctx.strokeStyle = ink; ctx.stroke(); ctx.setLineDash([]);
    arrowheads.push({points, kind, ink, opacity:ctx.globalAlpha});
    ctx.fillStyle = ink;
    ctx.beginPath(); ctx.arc(...points[0], 2.3 / camera.zoom, 0, Math.PI * 2); ctx.fill();
  }
  ctx.globalAlpha = 1;
  for (const node of model.nodes) {
    const sx = node.x * camera.zoom + camera.x, sy = node.y * camera.zoom + camera.y;
    if (sx > width || sy > height || sx + node.width * camera.zoom < 0 || sy + node.height * camera.zoom < 0) continue;
    ctx.globalAlpha = selected && !connected.has(node.id) ? .4 : 1;
    ctx.fillStyle = selected?.id === node.id ? '#24200a' : '#111111';
    ctx.strokeStyle = selected?.id === node.id ? '#ffd700' : connected.has(node.id) ? '#cca800' : '#333333';
    ctx.lineWidth = (selected?.id === node.id ? 2 : 1) / camera.zoom;
    ctx.beginPath(); ctx.rect(node.x, node.y, node.width, node.height); ctx.fill(); ctx.stroke();
    ctx.fillStyle = color(node); ctx.fillRect(node.x, node.y + 9, 3, node.height - 18);
    if (camera.zoom < 0.16) continue;
    ctx.save(); ctx.beginPath(); ctx.rect(node.x + 12, node.y, node.width - 24, node.height); ctx.clip();
    ctx.textBaseline = 'top';
    ctx.font = `${node.type !== 'function' ? '600' : '500'} 15px Rajdhani, sans-serif`;
    ctx.fillStyle = '#e0e0e0';
    ctx.fillText(node.type === 'folder' ? `${node.label}/` : node.type === 'file' ? node.file.path.split('/').pop() : node.label, node.x + 13, node.y + 12);
    ctx.font = '11px "Share Tech Mono", monospace'; ctx.fillStyle = '#999999';
    ctx.fillText(node.type === 'folder' ? `${node.files.length} files · ${node.functionCount} functions` : node.type === 'file' ? node.file.path : `Line ${node.fn.line_number} · score ${number(node.score)}`, node.x + 13, node.y + 33);
    if (node.type === 'folder') { ctx.fillText('Select to open folder', node.x + 13, node.y + 57); }
    else if (node.type === 'file') {
      ctx.fillStyle = color(node);
      ctx.fillText(node.file.analysis_mode === 'text' ? `${node.file.language} · Function metrics unavailable`
        : `${node.file.language} · ${node.file.hotspots.length} functions · ${node.file.lines_scanned} code lines`, node.x + 13, node.y + 57);
    }
    ctx.restore();
  }
  // Draw markers last so card fills and borders cannot cover their tips.
  for (const {points, kind, ink, opacity} of arrowheads) {
    const end = points.at(-1), before = points.at(-2);
    const angle = Math.atan2(end[1] - before[1], end[0] - before[0]);
    const gap = Math.min(4 / camera.zoom, Math.hypot(end[0] - before[0], end[1] - before[1]) / 3);
    const x = end[0] - Math.cos(angle) * gap, y = end[1] - Math.sin(angle) * gap;
    const size = Math.min(16, (kind === 'contains' ? 9 : 12) / camera.zoom);
    ctx.globalAlpha = opacity;
    ctx.fillStyle = ink; ctx.strokeStyle = '#050505'; ctx.lineWidth = 2 / camera.zoom;
    ctx.beginPath(); ctx.moveTo(x, y);
    ctx.lineTo(x - Math.cos(angle) * size + Math.sin(angle) * size * .5, y - Math.sin(angle) * size - Math.cos(angle) * size * .5);
    ctx.lineTo(x - Math.cos(angle) * size - Math.sin(angle) * size * .5, y - Math.sin(angle) * size + Math.cos(angle) * size * .5);
    ctx.closePath(); ctx.stroke(); ctx.fill();
  }
  ctx.globalAlpha = 1;
}
function element(tag, text, parent, className) {
  const el = document.createElement(tag);
  if (text !== undefined) el.textContent = text;
  if (className) el.className = className;
  if (parent) parent.append(el);
  return el;
}
function centerNode(node) {
  selected = node;
  camera.zoom = Math.max(camera.zoom, 0.85);
  camera.x = width / 2 - (node.x + node.width / 2) * camera.zoom;
  camera.y = height / 2 - (node.y + node.height / 2) * camera.zoom;
  details(node);
  scheduleDraw();
}
function details(node) {
  const panel = $('details'); panel.replaceChildren();
  if (!node) return;
  if (node.type === 'folder') return;
  const file = node.file;
  element('h2', node.label, panel);
  element('p', file.path, panel);
  const mode = file.analysis_mode === 'text' ? 'Text inventory: language and line count. Function metrics are not available for this language yet.'
    : file.parse_errors ? 'Partial AST analysis: the parser found syntax errors. Some functions may be missing.' : 'Function analysis';
  element('p', `${file.language} · ${file.lines_scanned} code lines`, panel, 'badge');
  element('p', mode, panel);
  const resolved = relationships.filter(edge => edge.kind === (node.type === 'function' ? 'calls' : 'imports'));
  const outgoing = resolved.filter(edge => edge.source_file === file.path && (node.type !== 'function' || edge.source_start_byte === node.fn.byte_range?.[0])).length;
  const incoming = resolved.filter(edge => edge.target_file === file.path && (node.type !== 'function' || edge.target_start_byte === node.fn.byte_range?.[0])).length;
  element('p', `Resolved ${node.type === 'function' ? 'calls' : 'imports'}: ${incoming} incoming · ${outgoing} outgoing`, panel);
  if (node.type === 'function') {
    const metrics = node.fn.metrics || {};
    const list = element('dl', undefined, panel);
    for (const [label, value] of [['Structural score', number(node.score)], ['Line', node.fn.line_number],
      ['Code heat', scoreBand(node.score).label], ['Branches + base path', metrics.cyclomatic_complexity], ['Nesting depth', metrics.max_nesting_depth], ['AST nodes', metrics.ast_node_count], ['Concurrency primitives', metrics.concurrency_primitives]]) {
      element('dt', label, list); element('dd', String(value ?? '—'), list);
    }
    element('h3', 'Score contributions', panel);
    const breakdown = element('dl', undefined, panel);
    for (const part of energyBreakdown(metrics)) {
      element('dt', `${part.label} × ${part.weight}${part.logarithmic ? ' (log₂)' : ''}`, breakdown);
      element('dd', part.contribution === null ? 'Unavailable' : number(part.contribution), breakdown);
    }
    element('pre', node.fn.source_snippet || '', panel);
    element('p', node.fn.description || '', panel);
  } else {
    if (file.imports?.length) element('pre', file.imports.map(item => item.path).join('\n'), panel);
    element('p', file.analysis_mode === 'text' ? 'Function metrics unavailable'
      : `${file.hotspots.length} functions · total energy ${number(file.total_entropy)}`, panel);
    for (const fn of file.hotspots) {
      const button = element('button', `${fn.function_name} · ${number(fn.entropy_score)} · L${fn.line_number}`, panel);
      button.type = 'button';
      button.onclick = () => {
        if (focusPath !== file.path) { focusPath = file.path; folderPath = ''; rebuild(); }
        const target = model.nodes.find(item => item.id === fn.id);
        if (target) centerNode(target);
      };
    }
  }
  const connections = model.links.filter(link => link.kind !== 'contains' && (link.source.id === node.id || link.target.id === node.id));
  if (connections.length) element('h3', 'Connections', panel);
  for (const link of connections) {
    const outgoing = link.source.id === node.id;
    const other = outgoing ? link.target : link.source;
    const action = link.kind === 'imports' ? (outgoing ? 'Imports' : 'Imported by') : (outgoing ? 'Calls' : 'Called by');
    const label = other.type === 'function' ? `${other.label} · ${other.file.path}` : other.label;
    const button = element('button', `${action}: ${label}`, panel); button.type = 'button';
    button.onclick = () => centerNode(other);
  }
}
function measurements() {
  const summary = summarizeRepository(files);
  $('measurements').hidden = false;
  $('measurement-cards').replaceChildren();
  for (const [label, value, note] of [
    ['Total structural energy', summary.total, 'Sum across measured functions'],
    ['Average function energy', summary.average, `${summary.functions.length} measured functions`],
    ['Peak function energy', summary.peak, `${summary.highCount} functions at high or critical heat`],
    ['Code heat', scoreBand(summary.average).label, 'Based on average function energy'],
  ]) {
    const card = element('div', undefined, $('measurement-cards'), 'measurement-card');
    element('span', label, card);
    const output = element('strong', typeof value === 'number' ? number(value) : value ?? 'Unavailable', card);
    if (label === 'Code heat') output.style.color = scoreBand(summary.average).color;
    element('small', note, card);
  }
  $('coverage').textContent = `Analysis coverage: ${summary.astFiles} of ${summary.fileCount} reported files have AST analysis · ${summary.textFiles} text-only · ${summary.partialFiles} with parsing problems. Text-only files have no function score. Counts exclude files skipped by the scanner. Resolved relationships: ${relationships.filter(edge => edge.kind === 'imports').length} imports · ${relationships.filter(edge => edge.kind === 'calls').length} calls.`;
  const rows = $('hotspot-rows'); rows.replaceChildren();
  for (const {file, fn, score} of summary.functions.slice(0, 10)) {
    const row = element('tr', undefined, rows);
    const cell = element('td', undefined, row);
    const button = element('button', fn.function_name, cell); button.type = 'button';
    button.onclick = () => {
      $('search').value = ''; $('language').value = ''; focusPath = file.path; folderPath = ''; rebuild();
      const target = model.nodes.find(node => node.id === fn.id); if (target) centerNode(target);
      $('stage').scrollIntoView({block:'center', behavior:'smooth'});
    };
    element('td', `${file.path}:${fn.line_number}`, row);
    element('td', number(score), row);
    const heat = element('td', scoreBand(score).label, row); heat.style.color = scoreBand(score).color;
    element('td', fn.vulnerability_type || 'Structural complexity', row);
  }
  $('hotspot-empty').hidden = summary.functions.length > 0;
}
function rebuild() {
  const options = {folderPath, groupFolders: $('folders').checked, relationships, showImports: $('imports').checked, query: $('search').value, language: $('language').value, focusPath, showFunctions: $('functions').checked, sort: $('sort').value};
  model = buildGraphModel(files, options);
  if (focusPath && !model.fileCount) { focusPath = ''; model = buildGraphModel(files, {...options, focusPath}); }
  selected = selected && model.nodes.find(node => node.id === selected.id);
  $('summary').textContent = `${model.fileCount} of ${files.length} files · ${model.functionCount} functions · ${model.links.filter(edge => edge.kind === 'imports').length} imports${focusPath ? ' · file focus' : folderPath ? ` · ${folderPath}/` : ''}`;
  $('file-count').textContent = `${model.filtered.length} matching files`;
  $('empty').hidden = model.nodes.length > 0;
  $('empty').textContent = files.length ? 'No files match these filters.' : 'Scan a repository to load its files and functions.';
  $('files').replaceChildren();
  const fragment = document.createDocumentFragment();
  for (const file of model.filtered) {
    const button = element('button', file.path, fragment);
    button.type = 'button'; button.setAttribute('aria-pressed', String(file.path === focusPath));
    button.onclick = () => {
      focusPath = file.path; folderPath = ''; rebuild(); selected = model.nodes.find(node => node.type === 'file' && node.file.path === focusPath); details(selected); scheduleDraw();
    };
  }
  $('files').append(fragment);
  details(selected);
  readableView();
}
function loadReport(report) {
  files = repositoryFiles(report); relationships = report.relationships || []; focusPath = ''; folderPath = ''; selected = null;
  const languages = [...new Set(files.map(file => file.language))].sort();
  $('language').replaceChildren();
  const all = element('option', 'All languages', $('language')); all.value = '';
  for (const name of languages) {
    const option = element('option', `${name} (${files.filter(file => file.language === name).length})`, $('language')); option.value = name;
  }
  measurements(); rebuild();
}
window.addEventListener('message', event => {
  if (event.origin !== location.origin || event.source !== parent || event.data?.type !== 'ucs503:report') return;
  if (!Array.isArray(event.data.report?.file_reports)) return;
  loadReport(event.data.report);
});
const embedded = new URLSearchParams(location.search).get('embedded') === 'true';
$('public-tools').hidden = embedded;
$('public-nav').hidden = embedded;
if (!embedded) {
  const root = new URL('../', location.href);
  if (root.pathname.endsWith('/admin-portal/')) {
    $('home-link').href = new URL('../', root);
    $('admin-link').href = new URL('#dashboard', root);
  }
  let controller;
  $('scan-form').onsubmit = async event => {
    event.preventDefault();
    controller = new AbortController();
    $('scan-button').disabled = true;
    $('cancel-scan').hidden = false;
    $('scan-error').textContent = '';
    $('scan-status').textContent = 'Scanning repository…';
    try {
      const report = await scanPublicRepository(scanConfig, $('repository-url').value, controller.signal);
      $('search').value = '';
      loadReport(report);
      $('scan-status').textContent = `Scan complete: ${files.length} files. Select a file or folder below to explore.`;
    } catch (error) {
      $('scan-status').textContent = error.name === 'AbortError' ? 'Scan cancelled.' : '';
      if (error.name !== 'AbortError') $('scan-error').textContent = error.message || 'Repository scan failed. Please try again.';
    } finally {
      $('scan-button').disabled = false;
      $('cancel-scan').hidden = true;
    }
  };
  $('cancel-scan').onclick = () => controller?.abort();
  window.addEventListener('pagehide', () => controller?.abort());
}
new ResizeObserver(() => {
  width = stage.clientWidth; height = stage.clientHeight;
  const ratio = Math.min(window.devicePixelRatio || 1, 2);
  canvas.width = Math.round(width * ratio); canvas.height = Math.round(height * ratio);
  canvas.style.width = `${width}px`; canvas.style.height = `${height}px`;
  readableView();
}).observe(stage);
canvas.addEventListener('pointerdown', event => {
  if (event.button !== 0) return;
  canvas.setPointerCapture(event.pointerId);
  pointer = { id:event.pointerId, x:event.clientX, y:event.clientY, cx:camera.x, cy:camera.y, moved:false };
});
canvas.addEventListener('pointermove', event => {
  if (!pointer || pointer.id !== event.pointerId) return;
  const dx = event.clientX - pointer.x, dy = event.clientY - pointer.y;
  pointer.moved ||= Math.hypot(dx, dy) > 4;
  if (pointer.moved) { camera.x = pointer.cx + dx; camera.y = pointer.cy + dy; scheduleDraw(); }
});
canvas.addEventListener('pointerup', event => {
  if (!pointer || pointer.id !== event.pointerId) return;
  if (!pointer.moved) {
    const rect = canvas.getBoundingClientRect();
    const x = (event.clientX - rect.left - camera.x) / camera.zoom, y = (event.clientY - rect.top - camera.y) / camera.zoom;
    const node = model.nodes.findLast(item => x >= item.x && y >= item.y && x <= item.x + item.width && y <= item.y + item.height);
    if (node?.type === 'folder') { folderPath = node.label; focusPath = ''; selected = null; rebuild(); }
    else if (node?.type === 'file') { focusPath = node.file.path; folderPath = ''; rebuild(); selected = model.nodes.find(item => item.id === node.id); details(selected); }
    else { selected = node || null; details(selected); }
    scheduleDraw();
  }
  pointer = null;
});
canvas.addEventListener('pointercancel', () => { pointer = null; });
canvas.addEventListener('wheel', event => {
  event.preventDefault(); const rect = canvas.getBoundingClientRect();
  zoomAt(Math.exp(-Math.max(-100, Math.min(100, event.deltaY)) * 0.003), event.clientX - rect.left, event.clientY - rect.top);
}, {passive:false});
canvas.addEventListener('keydown', event => {
  if (['+', '=', '-', 'f', 'F', 'Escape'].includes(event.key)) event.preventDefault();
  if (event.key === '+' || event.key === '=') zoomAt(1.25);
  else if (event.key === '-') zoomAt(0.8);
  else if (event.key.toLowerCase() === 'f') fit();
  else if (event.key === 'Escape') { focusPath = ''; folderPath = ''; selected = null; rebuild(); }
});
$('zoom-in').onclick = () => zoomAt(1.25);
$('zoom-out').onclick = () => zoomAt(0.8);
$('fit').onclick = fit;
$('overview').onclick = () => { focusPath = ''; folderPath = ''; selected = null; rebuild(); };
let searchTimer;
document.fonts.ready.then(scheduleDraw);
$('search').oninput = () => { clearTimeout(searchTimer); searchTimer = setTimeout(rebuild, 120); };
$('language').onchange = $('sort').onchange = $('functions').onchange = $('imports').onchange = $('folders').onchange = rebuild;
