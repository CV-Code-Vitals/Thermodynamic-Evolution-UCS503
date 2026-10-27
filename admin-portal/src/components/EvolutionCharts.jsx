import React from 'react';

function TrendChart({title, points, series}) {
  const values = points.flatMap(point => series.map(line => point[line.key])).filter(Number.isFinite);
  const minimum = Math.min(...values), maximum = Math.max(...values);
  const padding = Math.max((maximum - minimum) * .1, .1);
  const low = Math.max(0, minimum - padding), high = maximum + padding;
  const range = high - low;
  const x = step => 100 + step / Math.max(1, points.at(-1).step) * 450;
  const y = value => 155 - (value - low) / range * 120;
  return <div className="trend-chart">
    <h4>{title}</h4>
    <svg viewBox="0 0 580 195" role="img" aria-label={`${title}. ${series.map(line => `${line.label}: ${points.at(-1)[line.key]?.toFixed(2) ?? 'unavailable'}`).join(', ')}`}>
      <path d="M100 25V155H550" fill="none" stroke="#555" />
      <text x="92" y="38" textAnchor="end">{high.toFixed(2)}</text>
      <text x="92" y="155" textAnchor="end">{low.toFixed(2)}</text>
      <text x="100" y="178">Step 0</text><text x="550" y="178" textAnchor="end">Step {points.at(-1).step}</text>
      {series.map(line => <g key={line.key}>
        <polyline points={points.filter(point => Number.isFinite(point[line.key])).map(point => `${x(point.step)},${y(point[line.key])}`).join(' ')} fill="none" stroke={line.color} strokeWidth="2.5" strokeDasharray={line.dashed ? '7 5' : undefined} />
        {points.filter(point => Number.isFinite(point[line.key])).map(point => <circle key={point.step} cx={x(point.step)} cy={y(point[line.key])} r="3" fill={line.color}><title>Step {point.step}: {line.label} {point[line.key].toFixed(2)}</title></circle>)}
      </g>)}
    </svg>
    <div className="chart-legend">{series.map(line => <span key={line.key} style={{color:line.color}}>{line.dashed ? '┄' : '━'} {line.label}</span>)}</div>
  </div>;
}

export default function EvolutionCharts({measurements}) {
  const {points, baseline} = measurements;
  return <section className="evolution-trends panel" aria-label="Optimization measurements">
    <h3>Energy &amp; annealing temperature</h3>
    <p className="measurement-explanation">Lower energy means less structural complexity. Temperature controls exploration and has no physical units. The best tested candidate is retained even if the search temporarily accepts higher energy.</p>
    <div className="trend-grid">
      <TrendChart title="Structural energy" points={points.map(point => ({...point, baseline}))} series={[{key:'baseline', label:'Baseline', color:'#999', dashed:true}, {key:'current', label:'Current search', color:'#f59e0b'}, {key:'best', label:'Best tested', color:'#ffd700'}]} />
      <TrendChart title="Annealing temperature" points={points} series={[{key:'temperature', label:'Temperature at step start', color:'#e0e0e0'}]} />
    </div>
    <details><summary>View measured values</summary><div className="measurement-table"><table><thead><tr><th>Step</th><th>Current energy</th><th>Best energy</th><th>Temperature</th></tr></thead><tbody>{points.map(point => <tr key={point.step}><td>{point.step}</td><td>{point.current.toFixed(2)}</td><td>{point.best.toFixed(2)}</td><td>{point.temperature?.toFixed(2) ?? 'Unavailable'}</td></tr>)}</tbody></table></div></details>
  </section>;
}
