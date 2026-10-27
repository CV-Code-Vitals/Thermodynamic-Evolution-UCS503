export function evolutionMeasurements(initialEnergy, telemetry, result) {
  const steps = telemetry.filter(step => !step.isError);
  const baseline = result?.baseline_energy ?? steps.find(step => step.baseline_energy > 0)?.baseline_energy ?? initialEnergy;
  let current = baseline, best = baseline;
  const points = [{step:0, current, best, temperature:100}];
  for (const step of steps) {
    if (step.accepted && Number.isFinite(step.new_energy) && step.new_energy > 0) current = step.new_energy;
    else if (Number.isFinite(step.old_energy) && step.old_energy > 0) current = step.old_energy;
    best = Math.min(best, current);
    points.push({step:step.step, current, best, temperature:step.temp});
  }
  if (result) best = result.final_energy;
  const passed = steps.filter(step => step.evaluation_attempted && step.tests_passed).length;
  const failed = steps.filter(step => step.evaluation_attempted && !step.tests_passed).length;
  return {
    points, baseline, current, best, temperature:points.at(-1).temperature,
    improvement:baseline > 0 ? (baseline - best) / baseline * 100 : 0,
    passed:result?.correctness_passed ?? passed,
    failed:result?.correctness_failed ?? failed,
    skipped:steps.filter(step => !step.evaluation_attempted).length,
  };
}
