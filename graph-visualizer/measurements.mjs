export function scoreBand(score) {
  if (!Number.isFinite(score)) return {label:'Unmeasured', color:'#888888'};
  if (score >= 15) return {label:'Critical', color:'#ff6565'};
  if (score >= 10) return {label:'High', color:'#ef4444'};
  if (score >= 5) return {label:'Elevated', color:'#f59e0b'};
  return {label:'Cool', color:'#e0c65a'};
}

export function energyBreakdown(metrics = {}) {
  return [
    {label:'Branches + base path', raw:metrics.cyclomatic_complexity, weight:0.45},
    {label:'Maximum nesting', raw:metrics.max_nesting_depth, weight:0.25},
    {label:'AST nodes', raw:metrics.ast_node_count, weight:0.15, logarithmic:true},
    {label:'Concurrency primitives', raw:metrics.concurrency_primitives, weight:0.15},
  ].map(part => ({...part, contribution:Number.isFinite(part.raw)
    ? part.weight * (part.logarithmic ? Math.log2(Math.max(1, part.raw)) : part.raw) : null}));
}

export function summarizeRepository(files) {
  const functions = files.filter(file => file.analysis_mode !== 'text').flatMap(file =>
    file.hotspots.filter(fn => Number.isFinite(fn.entropy_score)).map(fn => ({file, fn, score:fn.entropy_score})));
  functions.sort((a, b) => b.score - a.score || a.file.path.localeCompare(b.file.path) || a.fn.line_number - b.fn.line_number);
  const total = functions.reduce((sum, fn) => sum + fn.score, 0);
  const average = functions.length ? total / functions.length : null;
  return {
    functions, total:functions.length ? total : null, average,
    peak:functions[0]?.score ?? null,
    highCount:functions.filter(fn => fn.score >= 10).length,
    astFiles:files.filter(file => file.analysis_mode !== 'text').length,
    textFiles:files.filter(file => file.analysis_mode === 'text').length,
    partialFiles:files.filter(file => file.parse_errors).length,
    fileCount:files.length,
  };
}
