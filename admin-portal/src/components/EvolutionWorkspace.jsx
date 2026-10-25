import React, { useEffect, useRef, useState } from 'react';
import EvolutionLive from './EvolutionLive';
import { scanEvolutionRepository } from '../evolution-api';

export default function EvolutionWorkspace({ apiBase, token, initialScan = null, onScanComplete = () => {} }) {
  const [repoURL, setRepoURL] = useState(initialScan?.repo_url || '');
  const [scan, setScan] = useState(initialScan);
  const [targetIndex, setTargetIndex] = useState('0');
  const [target, setTarget] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const abortRef = useRef(null);
  useEffect(() => () => abortRef.current?.abort(), []);

  const scanRepository = async (event) => {
    event.preventDefault();
    abortRef.current?.abort();
    abortRef.current = new AbortController();
    setLoading(true);
    setError('');
    setScan(null);
    onScanComplete(null);
    setTargetIndex('0');
    try {
      const result = await scanEvolutionRepository(apiBase, token, repoURL.trim(), abortRef.current.signal);
      setScan(result);
      onScanComplete(result);
    } catch (scanError) {
      if (scanError.name !== 'AbortError') {
        setError(scanError.message === 'analysis_not_configured'
          ? 'Repository analysis has not been enabled on this server.' : scanError.message);
      }
    } finally {
      setLoading(false);
    }
  };

  if (target) {
    return <EvolutionLive apiBase={apiBase} token={token} initialData={target}
      onBack={() => setTarget(null)} />;
  }
  return (
    <div className="dashboard-container">
      <h2>Analyze a repository</h2>
      <p>Scan a public GitHub repository and explore its files and functions in the graph.</p>
      <p>Function analysis supports Go, Python, JavaScript / JSX, TypeScript / TSX, Rust, Java, C, C++, C#, Ruby, PHP, Swift and shell. Other text files appear with their language and line count. Optimization supports Go modules with tests and dependencies available in the offline sandbox. The score measures code structure; runtime performance needs a separate benchmark.</p>
      <form className="auth-form" onSubmit={scanRepository}>
        <label className="input-label" htmlFor="evolution-repository">GitHub repository URL</label>
        <input id="evolution-repository" className="cyber-input" type="url" required
          placeholder="https://github.com/owner/repository" value={repoURL}
          onChange={(event) => setRepoURL(event.target.value)} disabled={loading} />
        <button className="cyber-button" disabled={loading} type="submit">
          {loading ? 'SCANNING REPOSITORY…' : 'SCAN REPOSITORY'}
        </button>
      </form>
      {error && <p className="error-text" role="alert">{error}</p>}
      {scan && (
        <div className="input-group">
          <p>Analyzed {scan.report.files_analyzed} files. The scanned copy expires at {new Date(scan.expires_at).toLocaleTimeString()}.</p>
          {scan.targets.length ? (
            <>
              <label className="input-label" htmlFor="evolution-function">Go function</label>
              <select id="evolution-function" className="cyber-input" value={targetIndex}
                onChange={(event) => setTargetIndex(event.target.value)}>
                {scan.targets.map((item, index) => (
                  <option key={`${item.file_path}:${item.start_byte}`} value={index}>
                    {item.file_path} · {item.function_name} · score {item.base_energy.toFixed(2)}
                  </option>
                ))}
              </select>
              <p>Changes apply to the scanned copy. Review the best candidate before applying it to your own repository.</p>
              <button className="cyber-button" onClick={() => setTarget({
                ...scan.targets[Number(targetIndex)], workspace_id: scan.workspace_id,
                repo_url: scan.repo_url,
              })}>REVIEW FUNCTION</button>
            </>
          ) : <p>Open the Graph tab to explore the scan. No eligible Go functions were found for automated optimization.</p>}
        </div>
      )}
    </div>
  );
}
