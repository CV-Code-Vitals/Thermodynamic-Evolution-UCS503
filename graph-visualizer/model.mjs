export function repositoryFiles(report) {
  const root = String(report?.scanned_directory || '').replaceAll('\\', '/').replace(/\/$/, '');
  return (report?.file_reports || []).filter(file => typeof file.file_path === 'string').map(file => {
    let path = file.file_path.replaceAll('\\', '/');
    if (root && path.startsWith(`${root}/`)) path = path.slice(root.length + 1);
    return { ...file, path, language: file.language || 'Text', hotspots: (file.hotspots || []).map((fn, index) => ({
      ...fn, id: `function:${JSON.stringify([path, fn.function_name, fn.byte_range?.[0] ?? fn.line_number, index])}`,
    })) };
  });
}

export function buildGraphModel(files, { query = '', language = '', focusPath = '', showFunctions = false, sort = 'path', relationships = [], showImports = true, groupFolders = false, folderPath = '' } = {}) {
  const term = query.trim().toLowerCase();
  const directory = file => file.path.includes('/') ? file.path.slice(0, file.path.lastIndexOf('/')) : '.';
  const filtered = files.filter(file => (!folderPath || directory(file) === folderPath) && (!language || file.language === language)
    && (!term || `${file.path} ${file.language}`.toLowerCase().includes(term)
      || file.hotspots.some(fn => fn.function_name.toLowerCase().includes(term))))
    .sort((a, b) => sort === 'score' ? b.total_entropy - a.total_entropy || a.path.localeCompare(b.path) : a.path.localeCompare(b.path));
  if (groupFolders && !folderPath && !focusPath && !showFunctions) {
    const folders = new Map();
    for (const file of filtered) {
      const path = directory(file);
      if (!folders.has(path)) folders.set(path, []);
      folders.get(path).push(file);
    }
    const columns = Math.max(1, Math.ceil(Math.sqrt(folders.size * 0.7)));
    const nodes = [...folders].sort(([a], [b]) => a.localeCompare(b)).map(([label, contents], index) => ({
      id:`folder:${label}`, type:'folder', label, files:contents,
      functionCount:contents.reduce((sum, file) => sum + file.hotspots.length, 0),
      score:contents.reduce((sum, file) => sum + (file.total_entropy || 0), 0),
      x:30 + (index % columns) * 300, y:30 + Math.floor(index / columns) * 120, width:264, height:84,
    }));
    const byPath = new Map(nodes.map(node => [node.label, node]));
    const byFile = new Map(filtered.map(file => [file.path, byPath.get(directory(file))]));
    const edges = new Map();
    if (showImports) for (const relation of relationships) {
      if (relation.kind !== 'imports') continue;
      const source = byFile.get(relation.source_file), target = byFile.get(relation.target_file);
      if (!source || !target || source === target) continue;
      const key = JSON.stringify([source.id, target.id]);
      if (edges.has(key)) edges.get(key).count++;
      else edges.set(key, {source, target, kind:'imports', count:1});
    }
    return {nodes, links:[...edges.values()], filtered, fileCount:filtered.length,
      functionCount:filtered.reduce((sum, file) => sum + file.hotspots.length, 0),
      width:Math.max(320, columns * 300), height:Math.max(200, Math.ceil(nodes.length / columns) * 120 + 30)};
  }
  const neighbors = new Set([focusPath]);
  if (focusPath && showImports) {
    for (const edge of relationships) {
      if (edge.kind !== 'imports') continue;
      if (edge.source_file === focusPath) neighbors.add(edge.target_file);
      if (edge.target_file === focusPath) neighbors.add(edge.source_file);
    }
  }
  const visible = focusPath ? filtered.filter(file => neighbors.has(file.path)) : filtered;
  const expanded = showFunctions || Boolean(focusPath);
  const columns = Math.max(1, Math.ceil(Math.sqrt(visible.length * (expanded ? 0.6 : 1.4))));
  const groupWidth = expanded ? 520 : 300;
  const nodes = [], links = [];
  let y = 30;
  for (let offset = 0; offset < visible.length; offset += columns) {
    let rowHeight = 120;
    visible.slice(offset, offset + columns).forEach((file, column) => {
      const x = 30 + column * groupWidth;
      const expandFile = showFunctions || file.path === focusPath;
      const parent = { id: `file:${file.path}`, type: 'file', label: file.path, file,
        score: file.total_entropy || 0, x, y, width: expandFile ? 480 : 264, height: 84 };
      nodes.push(parent);
      if (expandFile) {
        file.hotspots.forEach((fn, index) => {
          const child = { id: fn.id, type: 'function', label: fn.function_name, file, fn,
            score: fn.entropy_score || 0, x: x + (index % 2) * 248, y: y + 126 + Math.floor(index / 2) * 82, width: 232, height: 62 };
          nodes.push(child);
          links.push({ source: parent, target: child, kind: 'contains' });
        });
        rowHeight = Math.max(rowHeight, 152 + Math.ceil(file.hotspots.length / 2) * 82);
      }
    });
    y += rowHeight;
  }
  const fileNodes = new Map(nodes.filter(node => node.type === 'file').map(node => [node.file.path, node]));
  const functionNodes = new Map(nodes.filter(node => node.type === 'function').map(node => [JSON.stringify([node.file.path, node.fn.byte_range?.[0]]), node]));
  for (const edge of relationships) {
    if (edge.kind === 'imports' && showImports) {
      const source = fileNodes.get(edge.source_file), target = fileNodes.get(edge.target_file);
      if (source && target) links.push({source, target, kind:'imports'});
    } else if (edge.kind === 'calls') {
      const source = functionNodes.get(JSON.stringify([edge.source_file, edge.source_start_byte]));
      const target = functionNodes.get(JSON.stringify([edge.target_file, edge.target_start_byte]));
      if (source && target) links.push({source, target, kind:'calls'});
    }
  }
  return { nodes, links, filtered, fileCount: visible.length,
    functionCount: visible.reduce((sum, file) => sum + file.hotspots.length, 0),
    width: nodes.reduce((size, node) => Math.max(size, node.x + node.width + 30), 320),
    height: Math.max(200, y),
  };
}
