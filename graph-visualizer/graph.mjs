import { repositoryFiles, buildGraphModel } from './model.mjs';

const $ = id => document.getElementById(id);
const canvas = $('graph'), ctx = canvas.getContext('2d'), stage = $('stage');
let files = [], relationships = [], model = buildGraphModel([]), focusPath = '', folderPath = '', selected = null;
let camera = { x: 0, y: 0, zoom: 1 }, width = 1, height = 1, pending = false, pointer = null;
const number = value => Number(value || 0).toFixed(2);
const color = node => node.type === 'folder' ? '#60a5fa' : node.type === 'file' ? (node.file.analysis_mode === 'text' ? '#94a3b8' : '#a5b4fc')
  : node.score >= 10 ? '#fb7185' : node.score >= 5 ? '#fbbf24' : '#5eead4';

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
  ctx.lineWidth = 1.2 / camera.zoom;
  for (const { source, target, kind } of model.links) {
    const highlighted = !selected || selected.id === source.id || selected.id === target.id;
    ctx.globalAlpha = highlighted ? 0.75 : 0.12;
    ctx.strokeStyle = kind === 'imports' ? '#a78bfa' : kind === 'calls' ? '#fbbf24' : '#3b4b66';
    ctx.beginPath();
    if (source.id === target.id) {
      const x = source.x + source.width, y = source.y + source.height / 2;
      ctx.moveTo(x, y - 12); ctx.bezierCurveTo(x + 50, y - 45, x + 50, y + 45, x, y + 12);
    } else if (kind === 'contains') {
      const sx = source.x + source.width / 2, sy = source.y + source.height;
      const tx = target.x + target.width / 2, ty = target.y;
      ctx.moveTo(sx, sy); ctx.bezierCurveTo(sx, sy + 24, tx, ty - 24, tx, ty);
    } else {
      const right = target.x >= source.x + source.width;
      const sx = source.x + (right ? source.width : 0), sy = source.y + source.height / 2;
      const tx = target.x + (right ? 0 : target.width), ty = target.y + target.height / 2;
      const bend = Math.max(28, Math.abs(tx - sx) / 2);
      ctx.moveTo(sx, sy); ctx.bezierCurveTo(sx + (right ? bend : -bend), sy, tx + (right ? -bend : bend), ty, tx, ty);
      ctx.stroke();
      ctx.beginPath(); ctx.moveTo(tx + (right ? -7 : 7), ty - 4); ctx.lineTo(tx, ty); ctx.lineTo(tx + (right ? -7 : 7), ty + 4);
    }
    ctx.stroke();
  }
  ctx.globalAlpha = 1;
  for (const node of model.nodes) {
    const sx = node.x * camera.zoom + camera.x, sy = node.y * camera.zoom + camera.y;
    if (sx > width || sy > height || sx + node.width * camera.zoom < 0 || sy + node.height * camera.zoom < 0) continue;
    ctx.fillStyle = selected?.id === node.id ? '#263348' : '#151f2f';
    ctx.strokeStyle = selected?.id === node.id ? '#ffd700' : '#35445e';
    ctx.lineWidth = (selected?.id === node.id ? 2 : 1) / camera.zoom;
    ctx.beginPath(); ctx.roundRect(node.x, node.y, node.width, node.height, 8); ctx.fill(); ctx.stroke();
    ctx.fillStyle = color(node); ctx.fillRect(node.x, node.y + 9, 3, node.height - 18);
    if (camera.zoom < 0.16) continue;
    ctx.save(); ctx.beginPath(); ctx.rect(node.x + 12, node.y, node.width - 24, node.height); ctx.clip();
    ctx.textBaseline = 'top';
    ctx.font = `${node.type !== 'function' ? '600' : '500'} 13px system-ui`;
    ctx.fillStyle = '#e5e7eb';
    ctx.fillText(node.type === 'folder' ? `${node.label}/` : node.type === 'file' ? node.file.path.split('/').pop() : node.label, node.x + 13, node.y + 12);
    ctx.font = '11px system-ui'; ctx.fillStyle = '#a4b0c3';
    ctx.fillText(node.type === 'folder' ? `${node.files.length} files · ${node.functionCount} functions` : node.type === 'file' ? node.file.path : `Line ${node.fn.line_number} · score ${number(node.score)}`, node.x + 13, node.y + 33);
    if (node.type === 'folder') { ctx.fillText('Select to open folder', node.x + 13, node.y + 57); }
    else if (node.type === 'file') {
      ctx.fillStyle = color(node);
      ctx.fillText(`${node.file.language} · ${node.file.hotspots.length} functions · ${node.file.lines_scanned} lines`, node.x + 13, node.y + 57);
    }
    ctx.restore();
  }
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
  element('p', `${file.language} · ${file.lines_scanned} lines`, panel, 'badge');
  element('p', mode, panel);
  if (node.type === 'function') {
    const metrics = node.fn.metrics || {};
    const list = element('dl', undefined, panel);
    for (const [label, value] of [['Structural score', number(node.score)], ['Line', node.fn.line_number],
      ['Branches + base path', metrics.cyclomatic_complexity], ['Nesting depth', metrics.max_nesting_depth], ['AST nodes', metrics.ast_node_count]]) {
      element('dt', label, list); element('dd', String(value ?? '—'), list);
    }
    element('pre', node.fn.source_snippet || '', panel);
    element('p', node.fn.description || '', panel);
  } else {
    const imports = relationships.filter(edge => edge.kind === 'imports' && edge.source_file === file.path);
    if (imports.length) {
      element('p', 'Imports resolved to local files', panel);
      for (const edge of imports) {
        const button = element('button', edge.target_file, panel); button.type = 'button';
        button.onclick = () => { focusPath = edge.target_file; folderPath = ''; rebuild(); selected = model.nodes.find(item => item.type === 'file' && item.file.path === focusPath); details(selected); scheduleDraw(); };
      }
    }
    if (file.imports?.length) element('pre', file.imports.map(item => item.path).join('\n'), panel);
    element('p', `${file.hotspots.length} functions · score ${number(file.total_entropy)}`, panel);
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
}
function rebuild() {
  const options = {folderPath, groupFolders: $('folders').checked, relationships, showImports: $('imports').checked, query: $('search').value, language: $('language').value, focusPath, showFunctions: $('functions').checked, sort: $('sort').value};
  model = buildGraphModel(files, options);
  if (focusPath && !model.fileCount) { focusPath = ''; model = buildGraphModel(files, {...options, focusPath}); }
  selected = selected && model.nodes.find(node => node.id === selected.id);
  $('summary').textContent = `${model.fileCount} of ${files.length} files · ${model.functionCount} functions · ${model.links.filter(edge => edge.kind === 'imports').length} imports${focusPath ? ' · file focus' : folderPath ? ` · ${folderPath}/` : ''}`;
  $('file-count').textContent = `${model.filtered.length} matching files`;
  $('empty').hidden = model.nodes.length > 0;
  $('empty').textContent = files.length ? 'No files match these filters.' : 'Scan a repository in the Scan & Optimize tab to load its files.';
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
  rebuild();
}
window.addEventListener('message', event => {
  if (event.origin !== location.origin || event.source !== parent || event.data?.type !== 'ucs503:report') return;
  if (!Array.isArray(event.data.report?.file_reports)) return;
  loadReport(event.data.report);
});
if (new URLSearchParams(location.search).get('embedded') !== 'true') {
  const root = new URL('../', location.href);
  location.replace(new URL(root.pathname.endsWith('/admin-portal/') ? '#graph' : 'admin-portal/#graph', root));
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
$('search').oninput = () => { clearTimeout(searchTimer); searchTimer = setTimeout(rebuild, 120); };
$('language').onchange = $('sort').onchange = $('functions').onchange = $('imports').onchange = $('folders').onchange = rebuild;
