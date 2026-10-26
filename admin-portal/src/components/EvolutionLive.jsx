import React, { useState, useEffect, useRef } from 'react';
import { Play, Square, CheckCircle, XCircle, AlertTriangle, ArrowRight, Thermometer, Activity, Code, Download } from 'lucide-react';
import './EvolutionLive.css';
import { readEvolutionStream } from '../evolution-api';

const EvolutionLive = ({ initialData, onBack, apiBase, analysisApiBase = '', token }) => {
  const [isEvolving, setIsEvolving] = useState(false);
  const [telemetry, setTelemetry] = useState([]);
  const [stats, setStats] = useState({
    temp: 100,
    currentEnergy: initialData.base_energy,
    deltaE: 0.0,
  });
  const [finalResult, setFinalResult] = useState(null);
  
  const abortControllerRef = useRef(null);
  const logEndRef = useRef(null);

  const codeLeft = initialData.base_code;
  const [codeRight, setCodeRight] = useState(initialData.base_code);

  useEffect(() => () => abortControllerRef.current?.abort(), []);

  // Auto-scroll logs
  useEffect(() => {
    if (telemetry.length && logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [telemetry]);

  const startEvolution = async () => {
    setIsEvolving(true);
    setTelemetry([]);
    setFinalResult(null);
    setCodeRight(codeLeft);
    abortControllerRef.current = new AbortController();

    const payload = {
      workspace_id: initialData.workspace_id,
      repo_url: initialData.repo_url,
      file_path: initialData.file_path,
      function_name: initialData.function_name,
      start_byte: initialData.start_byte,
      source_hash: initialData.source_hash,
      config: { initial_temp: 100, cooling_rate: 0.85, max_steps: 20, enable_duplicate_detection: true }
    };

    try {
      const response = await fetch(analysisApiBase ? `${analysisApiBase}/api/evolve` : `${apiBase}/analysis/evolve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify(payload),
        signal: abortControllerRef.current.signal
      });
      await readEvolutionStream(response, (event, data) => {
        if (event === 'step') {
          setTelemetry(previous => [...previous, data]);
          setStats({ temp: data.temp, currentEnergy: data.accepted ? data.new_energy : data.old_energy, deltaE: data.delta_e });
          if (data.is_best_so_far && data.code_patch) setCodeRight(data.code_patch);
        } else if (event === 'complete') {
          setFinalResult(data);
          setCodeRight(data.optimized_code);
          setStats(previous => ({ ...previous, currentEnergy: data.final_energy, deltaE: -data.energy_delta }));
        }
      });
    } catch (err) {
      if (err.name === 'AbortError') {
        setTelemetry(prev => [...prev, { error: 'Evolution aborted by user.', isError: true }]);
      } else {
        setTelemetry(prev => [...prev, { error: err.message, isError: true }]);
      }
    } finally {
      setIsEvolving(false);
    }
  };

  const abortEvolution = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  };

  const downloadResult = (kind) => {
    const contents = kind === 'patch' ? finalResult.patch : JSON.stringify({
      repository: initialData.repo_url,
      file_path: initialData.file_path,
      function_name: initialData.function_name,
      baseline_code: codeLeft,
      ...finalResult,
    }, null, 2);
    const url = URL.createObjectURL(new Blob([contents], {
      type: kind === 'patch' ? 'text/plain;charset=utf-8' : 'application/json',
    }));
    const link = document.createElement('a');
    link.href = url;
    link.download = `evolution-${initialData.function_name}.${kind === 'patch' ? 'patch' : 'json'}`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 30000);
  };

  return (
    <div className="evolution-container">
      <div className="evolution-header">
        <div>
          <h2>Thermodynamic Evolution Live</h2>
          <p className="subtitle">Target: <code>{initialData.function_name}</code></p>
        </div>
        <div className="evolution-controls">
          {!isEvolving ? (
            <button className="cyber-button" onClick={startEvolution} disabled={Boolean(finalResult)}>
              <Play size={16} /> START EVOLUTION
            </button>
          ) : (
            <button className="cyber-button abort-button" onClick={abortEvolution}>
              <Square size={16} /> ABORT RUN
            </button>
          )}
          {onBack && (
            <button className="cyber-button-alt" onClick={onBack} disabled={isEvolving}>SCAN ANOTHER REPOSITORY</button>
          )}
        </div>
      </div>

      <div className="evolution-dashboard">
        {/* Thermodynamics Monitor */}
        <div className="thermo-monitor panel">
          <h3><Thermometer size={18} /> Structural complexity</h3>
          <div className="stats-grid">
            <div className="stat-card">
              <span className="stat-label">Temperature (T)</span>
              <span className="stat-value">{stats.temp.toFixed(2)}</span>
            </div>
            <div className="stat-card">
              <span className="stat-label">Current Energy (E)</span>
              <span className="stat-value">{stats.currentEnergy.toFixed(2)}</span>
            </div>
            <div className="stat-card">
              <span className="stat-label">Energy Delta (ΔE)</span>
              <span className={`stat-value ${stats.deltaE < 0 ? 'good' : stats.deltaE > 0 ? 'bad' : ''}`}>
                {stats.deltaE > 0 ? '+' : ''}{stats.deltaE.toFixed(2)}
              </span>
            </div>
          </div>
        </div>

        {/* Unified Diff Viewer */}
        <div className="diff-viewer panel">
          <h3><Code size={18} /> Code Evolution (Best Candidate)</h3>
          <div className="diff-split">
            <div className="code-pane">
              <div className="pane-header">Baseline</div>
              <pre><code>{codeLeft}</code></pre>
            </div>
            <div className="diff-arrow"><ArrowRight size={24} /></div>
            <div className="code-pane">
              <div className="pane-header">Optimized</div>
              <pre><code>{codeRight}</code></pre>
            </div>
          </div>
          {finalResult && (
            <div className="final-result-banner" role="status">
              <p>{finalResult.best_improved
                ? `Evolution complete. Structural score reduced by ${finalResult.energy_delta.toFixed(2)}.`
                : 'Run complete. No lower-scoring candidate passed evaluation; the original code was retained.'}</p>
              <p>{finalResult.total_steps} steps completed. {finalResult.correctness_passed} candidates passed tests.</p>
              <div className="evolution-downloads">
                {finalResult.patch && <button className="cyber-button" onClick={() => downloadResult('patch')}>
                  <Download size={16} /> DOWNLOAD PATCH
                </button>}
                <button className="cyber-button-alt" onClick={() => downloadResult('report')}>
                  <Download size={16} /> DOWNLOAD REPORT
                </button>
              </div>
              {finalResult.patch && <p>Review the patch, then apply it to the scanned revision with <code>git apply</code>. Downloaded results remain available after the scanned copy expires.</p>}
            </div>
          )}
        </div>

        {/* Telemetry Log */}
        <div className="telemetry-log panel">
          <h3><Activity size={18} /> Telemetry Event Log</h3>
          <div className="log-container">
            {telemetry.length === 0 && <div className="log-empty">Waiting for telemetry...</div>}
            {telemetry.map((t, idx) => (
              <div key={idx} className={`log-entry ${t.isError ? 'log-error' : t.accepted ? 'log-accepted' : 'log-rejected'}`}>
                {t.isError ? (
                  <><AlertTriangle size={14}/> <span>{t.error}</span></>
                ) : (
                  <>
                    <span className="step-num">Step {t.step}</span>
                    <span className="test-status">
                      {t.evaluation_attempted
                        ? t.tests_passed ? <CheckCircle size={14} className="icon-green"/> : <XCircle size={14} className="icon-red"/>
                        : <span aria-label="Tests not run">—</span>}
                    </span>
                    <span className="accept-status">
                      {t.accepted ? 'ACCEPTED' : t.evaluation_attempted ? 'REJECTED' : 'SKIPPED'}
                    </span>
                    <span className="reason-text">({t.reason}) ΔE: {t.delta_e?.toFixed(2)}</span>
                  </>
                )}
              </div>
            ))}
            <div ref={logEndRef} />
          </div>
        </div>
      </div>
    </div>
  );
};

export default EvolutionLive;
