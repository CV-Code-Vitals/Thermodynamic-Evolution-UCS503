import React, { useState, useEffect, useRef } from 'react';
import { Play, Square, CheckCircle, XCircle, AlertTriangle, ArrowRight, Thermometer, Activity, Code } from 'lucide-react';
import './EvolutionLive.css';

const apiBase = (import.meta.env.VITE_API_BASE || '/api').replace(/\/$/, '');

const EvolutionLive = ({ initialData, onBack }) => {
  const [isEvolving, setIsEvolving] = useState(false);
  const [telemetry, setTelemetry] = useState([]);
  const [stats, setStats] = useState({
    temp: initialData?.config?.initial_temp || 100.0,
    currentEnergy: initialData?.base_energy || 0.0,
    deltaE: 0.0,
  });
  const [finalResult, setFinalResult] = useState(null);
  
  const abortControllerRef = useRef(null);
  const logEndRef = useRef(null);

  const [codeLeft, setCodeLeft] = useState(initialData?.base_code || "// Original code snippet");
  const [codeRight, setCodeRight] = useState(initialData?.base_code || "// Baseline code");

  // Auto-scroll logs
  useEffect(() => {
    if (logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [telemetry]);

  const startEvolution = async () => {
    setIsEvolving(true);
    setTelemetry([]);
    setFinalResult(null);
    abortControllerRef.current = new AbortController();

    const payload = {
      repo_path: initialData?.repo_path || "/tmp/repo",
      file_path: initialData?.file_path || "main.go",
      function_name: initialData?.function_name || "ProcessData",
      start_byte: initialData?.start_byte || 0,
      end_byte: initialData?.end_byte || 100,
      base_energy: initialData?.base_energy || 50.0,
      config: initialData?.config || {
        initial_temp: 100.0,
        cooling_rate: 0.85,
        max_steps: 20
      }
    };

    try {
      const response = await fetch(`${apiBase}/evolve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
        signal: abortControllerRef.current.signal
      });

      if (!response.body) throw new Error("No response body");

      const reader = response.body.getReader();
      const decoder = new TextDecoder('utf-8');
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        
        const lines = buffer.split('\n');
        buffer = lines.pop() || ''; // Keep incomplete line in buffer

        let currentEvent = '';
        
        for (const line of lines) {
          if (line.startsWith('event: ')) {
            currentEvent = line.replace('event: ', '').trim();
          } else if (line.startsWith('data: ')) {
            const dataStr = line.replace('data: ', '').trim();
            if (!dataStr) continue;
            
            try {
              const data = JSON.parse(dataStr);
              
              if (currentEvent === 'step') {
                setTelemetry(prev => [...prev, data]);
                setStats({
                  temp: data.temp,
                  currentEnergy: data.new_energy || data.old_energy,
                  deltaE: data.delta_e || 0
                });
                if (data.accepted && data.code_patch) {
                  setCodeRight(data.code_patch);
                }
              } else if (currentEvent === 'complete') {
                setFinalResult(data);
                setIsEvolving(false);
                if (data.optimized_code) setCodeRight(data.optimized_code);
              } else if (currentEvent === 'error') {
                setTelemetry(prev => [...prev, { error: data.error, isError: true }]);
                setIsEvolving(false);
              }
            } catch (e) {
              console.error("Failed to parse SSE data", e, dataStr);
            }
          }
        }
      }
    } catch (err) {
      if (err.name === 'AbortError') {
        setTelemetry(prev => [...prev, { error: 'Evolution aborted by user.', isError: true }]);
      } else {
        setTelemetry(prev => [...prev, { error: err.message, isError: true }]);
      }
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
          <p className="subtitle">Target: <code>{initialData?.function_name || 'ProcessData'}</code></p>
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
            <button className="cyber-button-alt" onClick={onBack}>BACK</button>
          )}
        </div>
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
              Evolution Complete. Total Steps: {finalResult.total_steps}. Total ΔE: {finalResult.energy_delta.toFixed(2)}.
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
