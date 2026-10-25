import React, { useEffect, useRef, useState } from 'react';
import { Play, Square, CheckCircle, XCircle, AlertTriangle, ArrowRight, Thermometer, Activity, Code } from 'lucide-react';
import './EvolutionLive.css';
import { createEvolutionApi } from '../evolution-api.js';

const apiBase = (import.meta.env.VITE_ANALYSIS_API_BASE || '').replace(/\/+$/, '');
const evolutionApi = createEvolutionApi(apiBase);

const EvolutionLive = ({ token, initialData, onBack }) => {
  const [isEvolving, setIsEvolving] = useState(false);
  const [telemetry, setTelemetry] = useState([]);
  const [stats, setStats] = useState({
    temp: initialData?.config?.initial_temp || 100.0,
    currentEnergy: initialData?.base_energy || 0.0,
    deltaE: 0.0,
  });
  const [finalResult, setFinalResult] = useState(null);
  const [repoURL, setRepoURL] = useState(initialData?.repo_url || '');
  const [filePath, setFilePath] = useState(initialData?.file_path || '');
  const [functionName, setFunctionName] = useState(initialData?.function_name || '');
  const [error, setError] = useState('');
  const [status, setStatus] = useState('');

  const abortControllerRef = useRef(null);
  const logEndRef = useRef(null);

  const codeLeft = initialData?.base_code || "// Original code snippet";
  const [codeRight, setCodeRight] = useState(initialData?.base_code || "// Baseline code");

  useEffect(() => {
    if (logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [telemetry]);

  useEffect(() => () => abortControllerRef.current?.abort(), []);

  const startEvolution = async () => {
    if (!apiBase) {
      setError('Evolution analysis API is not configured for this deployment.');
      return;
    }
    if (!repoURL.trim() || !filePath.trim()) {
      setError('Enter a GitHub repository URL and target file before starting.');
      return;
    }

    setIsEvolving(true);
    setTelemetry([]);
    setFinalResult(null);
    setError('');
    setStatus('Connecting to the analysis service…');
    abortControllerRef.current = new AbortController();

    const payload = {
      repo_url: repoURL.trim(),
      file_path: filePath.trim(),
      ...(functionName.trim() ? { function_name: functionName.trim() } : {}),
      base_energy: initialData?.base_energy ?? 50.0,
      config: initialData?.config || {
        initial_temp: 100.0,
        cooling_rate: 0.85,
        max_steps: 20
      }
    };

    try {
      await evolutionApi.start({
        token,
        payload,
        signal: abortControllerRef.current.signal,
        onEvent: ({ type, data }) => {
          if (type === 'step') {
            setTelemetry((previous) => [...previous, data]);
            setStatus(`Evolution in progress — step ${data.step ?? '…'}.`);
            setStats((previous) => ({
              temp: Number.isFinite(data.temp) ? data.temp : previous.temp,
              currentEnergy: Number.isFinite(data.new_energy)
                ? data.new_energy
                : Number.isFinite(data.old_energy) ? data.old_energy : previous.currentEnergy,
              deltaE: Number.isFinite(data.delta_e) ? data.delta_e : 0,
            }));
            if (data.accepted && data.code_patch) setCodeRight(data.code_patch);
          } else if (type === 'complete') {
            setFinalResult(data);
            setStatus('Evolution completed.');
            if (data.optimized_code) setCodeRight(data.optimized_code);
          } else if (type === 'error') {
            setTelemetry((previous) => [...previous, { error: data.error || 'Evolution failed.', isError: true }]);
          }
        },
      });
    } catch (err) {
      if (err.name === 'AbortError') {
        setStatus('Evolution aborted.');
      } else {
        setError(err.message || 'Evolution failed. Please try again.');
        setStatus('');
      }
    } finally {
      abortControllerRef.current = null;
      setIsEvolving(false);
    }
  };

  const abortEvolution = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  };

  return (
    <div className="evolution-container">
      <div className="evolution-header">
        <div>
          <h2>Thermodynamic Evolution Live</h2>
          <p className="subtitle">Authenticated repository optimization</p>
        </div>
        <div className="evolution-controls">
          {!isEvolving ? (
            <button className="cyber-button" onClick={startEvolution}>
              <Play size={16} /> START EVOLUTION
            </button>
          ) : (
            <button className="cyber-button abort-button" onClick={abortEvolution}>
              <Square size={16} /> ABORT RUN
            </button>
          )}
          {onBack && (
            <button className="cyber-button-alt" onClick={onBack} disabled={isEvolving}>BACK</button>
          )}
        </div>
      </div>

      {!apiBase && (
        <div className="evolution-error" role="alert">
          Evolution analysis API is not configured for this deployment.
        </div>
      )}
      {error && <div className="evolution-error" role="alert">{error}</div>}
      {status && <div className="evolution-status" role="status" aria-live="polite">{status}</div>}

      <div className="evolution-target panel">
        <label>
          GitHub repository URL
          <input
            type="url"
            value={repoURL}
            onChange={(event) => setRepoURL(event.target.value)}
            placeholder="https://github.com/owner/repository"
            disabled={isEvolving}
            required
          />
        </label>
        <label>
          Target file
          <input
            type="text"
            value={filePath}
            onChange={(event) => setFilePath(event.target.value)}
            placeholder="src/main.go"
            disabled={isEvolving}
            required
          />
        </label>
        <label>
          Function name <span>(optional)</span>
          <input
            type="text"
            value={functionName}
            onChange={(event) => setFunctionName(event.target.value)}
            placeholder="ProcessData"
            disabled={isEvolving}
          />
        </label>
      </div>

      <div className="evolution-dashboard">
        {/* Thermodynamics Monitor */}
        <div className="thermo-monitor panel">
          <h3><Thermometer size={18} /> System Thermodynamics</h3>
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
            <div className="final-result-banner">
              Evolution Complete. Total Steps: {finalResult.total_steps ?? '—'}. Total ΔE: {Number(finalResult.energy_delta || 0).toFixed(2)}.
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
                      {t.tests_passed ? <CheckCircle size={14} className="icon-green"/> : <XCircle size={14} className="icon-red"/>}
                    </span>
                    <span className="accept-status">
                      {t.accepted ? 'ACCEPTED' : 'REJECTED'}
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
