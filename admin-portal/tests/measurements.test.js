import test from 'node:test';
import assert from 'node:assert/strict';
import { scoreBand, energyBreakdown, summarizeRepository } from '../../graph-visualizer/measurements.mjs';
import { evolutionMeasurements } from '../src/evolution-measurements.js';

test('repository measurements exclude text-only scores and show partial coverage', () => {
  const summary = summarizeRepository([
    {path:'a.go', hotspots:[{function_name:'simple', entropy_score:2}, {function_name:'complex', entropy_score:12}]},
    {path:'b.js', parse_errors:true, hotspots:[{function_name:'partial', entropy_score:4}]},
    {path:'unknown.foo', analysis_mode:'text', hotspots:[], total_entropy:0},
  ]);
  assert.equal(summary.total, 18);
  assert.equal(summary.average, 6);
  assert.equal(summary.peak, 12);
  assert.equal(summary.highCount, 1);
  assert.deepEqual(summary.functions.map(item => item.fn.function_name), ['complex', 'partial', 'simple']);
  assert.deepEqual([summary.astFiles, summary.textFiles, summary.partialFiles, summary.fileCount], [2, 1, 1, 3]);
  assert.equal(scoreBand(summary.average).label, 'Elevated');
});

test('empty and text-only reports have unavailable energy rather than a perfect zero', () => {
  for (const files of [[], [{path:'notes.txt', analysis_mode:'text', hotspots:[]}]]) {
    const summary = summarizeRepository(files);
    assert.equal(summary.total, null);
    assert.equal(summary.average, null);
    assert.equal(summary.peak, null);
    assert.equal(scoreBand(summary.average).label, 'Unmeasured');
  }
});

test('score bands share exact boundaries and contributions match the engine formula', () => {
  assert.deepEqual([null, 4.99, 5, 9.99, 10, 14.99, 15].map(score => scoreBand(score).label),
    ['Unmeasured', 'Cool', 'Elevated', 'Elevated', 'High', 'High', 'Critical']);
  const parts = energyBreakdown({cyclomatic_complexity:10, max_nesting_depth:4, ast_node_count:16, concurrency_primitives:2});
  assert.deepEqual(parts.map(part => part.contribution), [4.5, 1, .6, .3]);
  assert.equal(Number(parts.reduce((sum, part) => sum + part.contribution, 0).toFixed(2)), 6.4);
  assert.equal(energyBreakdown({ast_node_count:0})[2].contribution, 0);
  assert.equal(energyBreakdown()[0].contribution, null);
});

test('evolution preserves the best tested energy during uphill moves and skipped or failed evaluations', () => {
  const telemetry = [
    {step:1, baseline_energy:10, old_energy:10, new_energy:8, temp:100, accepted:true, evaluation_attempted:true, tests_passed:true},
    {step:2, old_energy:8, new_energy:9, temp:85, accepted:true, evaluation_attempted:true, tests_passed:true},
    {step:3, old_energy:9, new_energy:0, temp:72.25, accepted:false, evaluation_attempted:false},
    {step:4, old_energy:9, new_energy:0, temp:61.41, accepted:false, evaluation_attempted:true, tests_passed:false},
    {isError:true, error:'aborted'},
  ];
  const measurements = evolutionMeasurements(99, telemetry, null);
  assert.equal(measurements.baseline, 10, 'Use the measured backend baseline');
  assert.deepEqual(measurements.points.map(point => point.current), [10, 8, 9, 9, 9]);
  assert.deepEqual(measurements.points.map(point => point.best), [10, 8, 8, 8, 8]);
  assert.equal(measurements.improvement, 20);
  assert.deepEqual([measurements.passed, measurements.failed, measurements.skipped], [2, 1, 1]);
  const completed = evolutionMeasurements(10, telemetry, {baseline_energy:10, final_energy:8, correctness_passed:2, correctness_failed:1});
  assert.equal(completed.best, 8);
  assert.equal(completed.current, 9, 'Distinguish the search state from the exported best');
  assert.equal(evolutionMeasurements(10, [], null).improvement, 0);
});
