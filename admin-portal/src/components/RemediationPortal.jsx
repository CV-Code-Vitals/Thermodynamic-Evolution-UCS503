import React, { useState, useEffect } from 'react';
import {
  Play, Lock, RotateCcw, Download, Filter, Search, ChevronLeft, ChevronRight
} from 'lucide-react';
import './RemediationPortal.css';
import {
  fetchRemediationTargets,
  analyzeRemediationTarget,
  proposeRemediationRepair,
  approveRemediationRepair,
  executeRemediationRepair,
  rollbackRemediationSandbox,
  fetchRemediationAuditReport,
  runRemediationDemo
} from '../remediation-api';

function renderVerdictBadge(verdict) {
  switch (verdict) {
    case 'INDEPENDENTLY_VERIFIED':
      return <span className="remediation-badge badge-verified">INDEPENDENTLY_VERIFIED</span>;
    case 'SYNTAX_ONLY_VERIFIED':
      return <span className="remediation-badge badge-review">SYNTAX_ONLY_VERIFIED</span>;
    case 'EXPLICITLY_ABSTAINED':
    case 'ABSTAINED':
      return <span className="remediation-badge badge-abstain">ABSTAINED</span>;
    case 'FALSE_POSITIVE':
      return <span className="remediation-badge badge-false-positive">FALSE_POSITIVE</span>;
    case 'GROUND_TRUTH_UNVERIFIED':
    case 'GROUND_TRUTH_DISQUALIFIED':
      return <span className="remediation-badge badge-unverified">GROUND_TRUTH_UNVERIFIED</span>;
    case 'FAILED':
      return <span className="remediation-badge badge-false-positive">FAILED</span>;
    default:
      return <span className="remediation-badge badge-review">{verdict}</span>;
  }
}

export default function RemediationPortal({ apiBase = '/api', token = '', latestScan = null, onNavigateToScan = null }) {
  const [targets, setTargets] = useState([]);
  const [selectedTarget, setSelectedTarget] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Filter & Pagination state for large target corpora
  const [categoryFilter, setCategoryFilter] = useState('ALL');
  const [actionFilter, setActionFilter] = useState('ALL');
  const [searchQuery, setSearchQuery] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 25;

  // Pipeline step state
  const [analysis, setAnalysis] = useState(null);
  const [proposal, setProposal] = useState(null);
  const [approval, setApproval] = useState(null);
  const [execution, setExecution] = useState(null);
  const [auditReport, setAuditReport] = useState(null);
  const [rollbackStatus, setRollbackStatus] = useState('');

  useEffect(() => {
    let isMounted = true;
    setTargets([]);
    setSelectedTarget('');
    setAnalysis(null);
    setProposal(null);
    setApproval(null);
    setExecution(null);
    setAuditReport(null);
    setRollbackStatus('');
    async function loadTargets() {
      if (!latestScan?.workspace_id) return;
      try {
        const list = await fetchRemediationTargets(apiBase, token, latestScan?.workspace_id);
        if (isMounted) {
          setTargets(list);
          setSelectedTarget(prev => list.some(target => target.target_id === prev) ? prev : (list[0]?.target_id || ''));
          setError('');
        }
      } catch (loadError) {
        if (isMounted) {
          setTargets([]);
          setSelectedTarget('');
          setError(loadError.message);
        }
      }
    }
    void loadTargets();
    return () => { isMounted = false; };
  }, [apiBase, token, latestScan?.workspace_id]);

  const resetPipeline = () => {
    setAnalysis(null);
    setProposal(null);
    setApproval(null);
    setExecution(null);
    setError('');
    setRollbackStatus('');
  };

  const handleRunAnalysis = async () => {
    setLoading(true);
    setError('');
    resetPipeline();
    try {
      const res = await analyzeRemediationTarget(apiBase, token, selectedTarget);
      setAnalysis(res);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleProposeRepair = async () => {
    if (!analysis) return;
    setLoading(true);
    setError('');
    try {
      const res = await proposeRemediationRepair(apiBase, token, analysis.analysis_id);
      setProposal(res);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleApproveRepair = async () => {
    if (!proposal) return;
    setLoading(true);
    setError('');
    try {
      const res = await approveRemediationRepair(
        apiBase,
        token,
        proposal.proposal_id
      );
      setApproval(res);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleExecuteRepair = async () => {
    if (!approval) return;
    setLoading(true);
    setError('');
    try {
      const res = await executeRemediationRepair(apiBase, token, approval.token_id);
      setExecution(res);
      // Auto-refresh audit report
      const audit = await fetchRemediationAuditReport(apiBase, token);
      setAuditReport(audit);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleRollback = async () => {
    if (!execution) return;
    setLoading(true);
    try {
      await rollbackRemediationSandbox(apiBase, token, execution.execution_id);
      setRollbackStatus('Sandbox deletion verified. No persisted repository revision was restored.');
    } catch (err) {
      setError(`Rollback failed: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleQuickDemo = async () => {
    setLoading(true);
    setError('');
    try {
      await runRemediationDemo(apiBase, token);
      const audit = await fetchRemediationAuditReport(apiBase, token);
      setAuditReport(audit);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const downloadAuditJSON = () => {
    if (!auditReport) return;
    const blob = new Blob([JSON.stringify(auditReport, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `remediation_audit_${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const scannedTargets = targets.filter(t => !t.is_demo_target && (t.stratum === 'SCANNED_REPOSITORY' || t.target_id.startsWith('SCAN-')));
  const demoTargets = targets.filter(t => t.is_demo_target || (t.stratum !== 'SCANNED_REPOSITORY' && !t.target_id.startsWith('SCAN-')));

  const filteredScanned = scannedTargets.filter(t => {
    if (categoryFilter !== 'ALL' && t.defect_category !== categoryFilter) return false;
    if (actionFilter !== 'ALL' && t.expected_action !== actionFilter) return false;
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      const matchId = (t.target_id || '').toLowerCase().includes(q);
      const matchFile = (t.target_file || '').toLowerCase().includes(q);
      const matchDesc = (t.description || '').toLowerCase().includes(q);
      if (!matchId && !matchFile && !matchDesc) return false;
    }
    return true;
  });

  const totalPages = Math.max(1, Math.ceil(filteredScanned.length / pageSize));
  const paginatedScanned = filteredScanned.slice((currentPage - 1) * pageSize, currentPage * pageSize);
  const currentTargetObj = targets.find(t => t.target_id === selectedTarget);

  return (
    <div className="remediation-container">
      <div className="remediation-header">
        <h2>Autonomous Code Repair MVP</h2>
        <p>Controlled, Sandboxed AST-Level Remediation with Decoupled Independent Oracles & Cryptographic Authorization</p>
        <div style={{ display: 'flex', gap: '0.75rem', marginTop: '0.5rem' }}>
          <span className="remediation-badge badge-review">SOURCE: {targets.find(t => t.target_id === selectedTarget)?.upstream_commit?.slice(0, 12) || 'No selected revision'}</span>
          <span className="remediation-badge badge-review">BRANCH: release/milestone-3.8-mvp-rc</span>
          <span className="remediation-badge badge-abstain">FAIL-CLOSED SAFETY ACTIVE</span>
        </div>
      </div>

      {/* CONNECTED WORKSPACE BANNER */}
      {latestScan ? (
        <div className="connected-repo-banner" style={{ background: 'rgba(0, 255, 136, 0.08)', border: '1px solid rgba(0, 255, 136, 0.3)', borderRadius: '6px', padding: '0.75rem 1rem', marginBottom: '1.25rem' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
            <div>
              <div style={{ fontSize: '0.75rem', color: '#00ff88', fontWeight: 600, letterSpacing: '0.05em' }}>CONNECTED SCANNED REPOSITORY</div>
              <div style={{ fontSize: '1rem', fontWeight: 600, color: '#e0e6ed', marginTop: '2px' }}>
                {latestScan.repo_url || 'Active Workspace'}
                <span style={{ fontSize: '0.8rem', color: '#8892b0', marginLeft: '0.5rem', fontWeight: 400 }}>
                  (Workspace: {latestScan.workspace_id ? `${latestScan.workspace_id.slice(0, 8)}...` : 'retained'})
                </span>
              </div>
              <div style={{ fontSize: '0.8rem', color: '#8892b0', marginTop: '2px' }}>
                {latestScan.report?.files_analyzed ?? '—'} files analyzed · {latestScan.report?.total_hotspots ?? '—'} hotspots detected · {scannedTargets.length} findings available for remediation
              </div>
            </div>
            {onNavigateToScan && (
              <button type="button" className="action-btn secondary" style={{ fontSize: '0.75rem', padding: '4px 10px' }} onClick={onNavigateToScan}>
                CHANGE / RE-SCAN
              </button>
            )}
          </div>
        </div>
      ) : (
        <div className="connected-repo-banner" style={{ background: 'rgba(56, 189, 248, 0.05)', border: '1px solid rgba(56, 189, 248, 0.2)', borderRadius: '6px', padding: '0.75rem 1rem', marginBottom: '1.25rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '0.5rem' }}>
          <div>
            <div style={{ fontSize: '0.75rem', color: '#38bdf8', fontWeight: 600 }}>NO LIVE REPOSITORY SCANNED</div>
            <div style={{ fontSize: '0.8rem', color: '#8892b0' }}>
              Scan a GitHub repository in <strong>Scan & Optimize</strong> to discover and repair live findings, or explicitly run the labeled demo.
            </div>
          </div>
          {onNavigateToScan && (
            <button type="button" className="action-btn" style={{ fontSize: '0.75rem', padding: '4px 10px' }} onClick={onNavigateToScan}>
              GO TO SCAN & OPTIMIZE
            </button>
          )}
        </div>
      )}

      {error && (
        <div className="abstain-banner" style={{ borderColor: '#ff4444', color: '#ffaaaa' }}>
          <h4>Error Encountered</h4>
          <p>{error}</p>
        </div>
      )}

      {/* STEP 1: TARGET SELECTION */}
      <div className={`remediation-step-card ${!analysis ? 'active' : ''}`}>
        <h3 className="step-title">
          <span className="step-number">1</span>
          Repository Target Selection
        </h3>
        <p>
          {scannedTargets.length > 0
            ? `Select a diagnostic finding from the scanned repository (${scannedTargets.length} discovered) :`
            : 'No live findings available. Scan a repository or explicitly run the demo.'}
        </p>

        {/* Filter Toolbar for Scanned Targets */}
        {scannedTargets.length > 0 && (
          <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap', alignItems: 'center', marginBottom: '1rem', background: 'rgba(255,255,255,0.02)', padding: '0.5rem 0.75rem', borderRadius: '6px', border: '1px solid #1f2937' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
              <Filter size={14} style={{ color: '#00ff88' }} />
              <span style={{ fontSize: '0.75rem', color: '#9ca3af', fontWeight: 600 }}>FILTERS:</span>
            </div>
            <select
              className="cyber-input"
              style={{ fontSize: '0.75rem', padding: '3px 8px', height: 'auto' }}
              value={categoryFilter}
              onChange={(e) => { setCategoryFilter(e.target.value); setCurrentPage(1); }}
            >
              <option value="ALL">All Categories</option>
              <option value="RESOURCE_LEAK">Resource Leaks</option>
              <option value="SQL_INJECTION">SQL Injection</option>
              <option value="COMMAND_INJECTION">Command Injection</option>
              <option value="HIGH_COMPLEXITY">High Complexity Hotspots</option>
              <option value="UNHANDLED_ERROR">Unhandled Errors</option>
            </select>
            <select
              className="cyber-input"
              style={{ fontSize: '0.75rem', padding: '3px 8px', height: 'auto' }}
              value={actionFilter}
              onChange={(e) => { setActionFilter(e.target.value); setCurrentPage(1); }}
            >
              <option value="ALL">All Actions</option>
              <option value="REPAIR">Eligible for Repair (REPAIR)</option>
              <option value="ABSTAIN">Requires Abstention / Review (ABSTAIN)</option>
            </select>
            <div style={{ position: 'relative', flexGrow: 1, minWidth: '180px' }}>
              <input
                className="cyber-input"
                style={{ fontSize: '0.75rem', padding: '3px 8px 3px 24px', width: '100%', height: 'auto' }}
                type="text"
                placeholder="Search by file, rule ID, or description..."
                value={searchQuery}
                onChange={(e) => { setSearchQuery(e.target.value); setCurrentPage(1); }}
              />
              <Search size={12} style={{ position: 'absolute', left: '7px', top: '7px', color: '#6b7280' }} />
            </div>
            {(categoryFilter !== 'ALL' || actionFilter !== 'ALL' || searchQuery) && (
              <button
                type="button"
                className="action-btn secondary"
                style={{ fontSize: '0.7rem', padding: '2px 8px', height: 'auto' }}
                onClick={() => { setCategoryFilter('ALL'); setActionFilter('ALL'); setSearchQuery(''); setCurrentPage(1); }}
              >
                CLEAR
              </button>
            )}
          </div>
        )}

        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap' }}>
          <select
            className="cyber-input"
            value={selectedTarget}
            onChange={(e) => {
              setSelectedTarget(e.target.value);
              resetPipeline();
            }}
            disabled={loading || targets.length === 0}
            style={{ minWidth: '320px' }}
          >
            {scannedTargets.length > 0 && (
              <optgroup label={`Scanned Findings (${filteredScanned.length} matching / ${scannedTargets.length} total, Page ${currentPage} of ${totalPages})`}>
                {paginatedScanned.map((t) => (
                  <option key={t.target_id} value={t.target_id}>
                    {t.target_id}: {t.description || `${t.target_file} (${t.defect_category})`}
                  </option>
                ))}
              </optgroup>
            )}
            <optgroup label="Preconfigured Deterministic Demo Targets (5)">
              {demoTargets.map((t) => (
                <option key={t.target_id} value={t.target_id}>
                  {t.target_id}: {t.repository} — {t.defect_category} [{t.stratum || (t.target_id === 'NAT-03' ? 'GROUND_TRUTH_DISQUALIFIED' : 'DEMO_FIXTURE')}]
                </option>
              ))}
            </optgroup>
          </select>
          <button className="action-btn" onClick={handleRunAnalysis} disabled={loading || targets.length === 0}>
            {loading ? 'ANALYZING...' : 'RUN SEMANTIC ANALYSIS'}
          </button>
          <button className="action-btn secondary" onClick={handleQuickDemo} disabled={loading}>
            RUN ONE-CLICK DETERMINISTIC DEMO
          </button>
        </div>

        {/* Pagination controls */}
        {scannedTargets.length > 0 && totalPages > 1 && (
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginTop: '0.75rem', fontSize: '0.8rem', color: '#9ca3af' }}>
            <button
              type="button"
              className="action-btn secondary"
              style={{ padding: '2px 8px', fontSize: '0.75rem', display: 'flex', alignItems: 'center', gap: '2px' }}
              disabled={currentPage <= 1 || loading}
              onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
            >
              <ChevronLeft size={12} /> PREV
            </button>
            <span>
              Page {currentPage} of {totalPages} (Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, filteredScanned.length)} of {filteredScanned.length} targets)
            </span>
            <button
              type="button"
              className="action-btn secondary"
              style={{ padding: '2px 8px', fontSize: '0.75rem', display: 'flex', alignItems: 'center', gap: '2px' }}
              disabled={currentPage >= totalPages || loading}
              onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
            >
              NEXT <ChevronRight size={12} />
            </button>
          </div>
        )}

        {/* Selected target preview snippet */}
        {currentTargetObj && (
          <div style={{ marginTop: '0.75rem', padding: '0.5rem 0.75rem', background: 'rgba(0,0,0,0.25)', borderRadius: '4px', border: '1px solid #1f2937', fontSize: '0.75rem', color: '#9ca3af' }}>
            <span style={{ color: '#00ff88', fontWeight: 600 }}>SELECTED TARGET:</span> {currentTargetObj.target_id} |{' '}
            <span style={{ color: '#38bdf8' }}>CATEGORY:</span> {currentTargetObj.defect_category} |{' '}
            <span style={{ color: '#f59e0b' }}>EXPECTED:</span> {currentTargetObj.expected_action} |{' '}
            <span style={{ color: '#e5e7eb' }}>FILE:</span> {currentTargetObj.target_file}:{currentTargetObj.line_number || 1}
            {currentTargetObj.snapshot_id && (
              <span style={{ marginLeft: '0.5rem', color: '#8892b0' }}>
                [SNAPSHOT: {currentTargetObj.snapshot_id}]
              </span>
            )}
          </div>
        )}
      </div>

      {/* STEP 2: FINDINGS & SEMANTIC EVIDENCE */}
      {analysis && (
        <div className={`remediation-step-card ${analysis && !proposal ? 'active' : ''}`}>
          <h3 className="step-title">
            <span className="step-number">2</span>
            Diagnostic Findings & Semantic Evidence
          </h3>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '0.5rem' }}>
            <div><strong>Finding ID:</strong> {analysis.finding_id}</div>
            <div><strong>Defect Category:</strong> {analysis.defect_category}</div>
            <div>
              <strong>Provenance Stratum:</strong>{' '}
              <span className={`remediation-badge ${analysis.stratum === 'SCANNED_REPOSITORY' || analysis.target_id.startsWith('SCAN-') ? 'badge-verified' : (analysis.target_id === 'NAT-03' ? 'badge-abstain' : (analysis.target_id.startsWith('DEF-HOLD') ? 'badge-review' : 'badge-verified'))}`}>
                {analysis.stratum || (analysis.target_id === 'NAT-03' ? 'GROUND_TRUTH_DISQUALIFIED' : (analysis.target_id.startsWith('DEF-HOLD') ? 'CONSTRUCTED_MICROSERVICE' : (analysis.target_id.startsWith('SCAN-') ? 'SCANNED_REPOSITORY' : 'AUTHENTIC_REPO_INJECTED')))}
              </span>
            </div>
            <div><strong>File & Line:</strong> {analysis.target_file}:{analysis.line_number}</div>
            <div>
              <strong>Automatic Eligibility:</strong>{' '}
              <span className={`remediation-badge ${analysis.automatic_eligibility ? 'badge-verified' : 'badge-abstain'}`}>
                {analysis.automatic_eligibility ? 'ELIGIBLE' : 'INELIGIBLE (REQUIRES HUMAN REVIEW / ABSTAIN)'}
              </span>
            </div>
            <div><strong>Uncertainty Score:</strong> {(analysis.uncertainty_score * 100).toFixed(1)}%</div>
          </div>

          <div className="evidence-box">
            <strong>Reaching Definitions & Taint Evidence:</strong>
            <ul style={{ margin: '0.25rem 0', paddingLeft: '1.25rem' }}>
              {analysis.semantic_evidence.map((ev, i) => (
                <li key={i}>{ev}</li>
              ))}
            </ul>
            <strong>Eligibility Rationale:</strong> {analysis.eligibility_rationale}
          </div>

          {!proposal && (
            <div className="btn-group">
              <button className="action-btn" onClick={handleProposeRepair} disabled={loading}>
                {loading ? 'SYNTHESIZING...' : 'PROPOSE AUTONOMOUS REPAIR'}
              </button>
            </div>
          )}
        </div>
      )}

      {/* STEP 3: REPAIR PROPOSAL & DIFF */}
      {proposal && (
        <div className={`remediation-step-card ${proposal && !approval ? 'active' : ''}`}>
          <h3 className="step-title">
            <span className="step-number">3</span>
            Autonomous Patch Synthesis & Review
          </h3>

          {proposal.status === 'ABSTAINED' ? (
            <div className="abstain-banner">
              <h4>CONSERVATIVE FAIL-CLOSED ABSTENTION ENFORCED</h4>
              <p><strong>Category:</strong> {proposal.abstain_category}</p>
              <p><strong>Architectural Rationale:</strong> {proposal.abstain_reason}</p>
              <p>The system refuses to apply unverified mutations that require architectural redesign or modify public API signatures.</p>
            </div>
          ) : (
            <>
              <p>
                <strong>Intent:</strong> {proposal.intent_type} |{' '}
                <strong>Diff SHA-256:</strong> <code>{proposal.diff_hash?.slice(0, 16)}...</code> |{' '}
                <strong>Synthesis Latency:</strong> {proposal.duration_ms} ms
              </p>
              <div className="diff-box">
                {proposal.unified_diff.split('\n').map((line, idx) => {
                  let cls = '';
                  if (line.startsWith('+') && !line.startsWith('+++')) cls = 'diff-line-add';
                  else if (line.startsWith('-') && !line.startsWith('---')) cls = 'diff-line-del';
                  return <div key={idx} className={cls}>{line}</div>;
                })}
              </div>

              {!approval && (
                <div style={{ marginTop: '1rem' }}>
                  <h4>Authorize repair</h4>
                  <button className="action-btn" onClick={handleApproveRepair} disabled={loading}>
                    <Lock size={14} style={{ marginRight: '6px', verticalAlign: 'middle' }} />
                    AUTHORIZE REPAIR
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      )}

      {/* STEP 4: AUTHORIZATION TOKEN & SANDBOX EXECUTION */}
      {approval && (
        <div className={`remediation-step-card ${approval && !execution ? 'active' : ''}`}>
          <h3 className="step-title">
            <span className="step-number">4</span>
            Cryptographic Authorization Token
          </h3>
          <div className="token-box">
            <div><strong>Token ID:</strong> {approval.token_id}</div>
            <div><strong>Approving Principal:</strong> {approval.principal} ({approval.role})</div>
            <div><strong>HMAC-SHA256 Signature:</strong> {approval.signature}</div>
            <div><strong>Expires At:</strong> {new Date(approval.expires_at).toLocaleString()}</div>
            <div><strong>Single-Use Replay Protection:</strong> ACTIVE (Atomic Persistent Ledger)</div>
          </div>

          {!execution && (
            <div className="btn-group">
              <button className="action-btn" onClick={handleExecuteRepair} disabled={loading}>
                <Play size={14} style={{ marginRight: '6px', verticalAlign: 'middle' }} />
                EXECUTE IN ISOLATED SANDBOX & VERIFY ORACLES
              </button>
            </div>
          )}
        </div>
      )}

      {/* STEP 5: ORACLES & IMMUTABILITY */}
      {execution && (
        <div className="remediation-step-card active">
          <h3 className="step-title">
            <span className="step-number">5</span>
            Multi-Stage Independent Verification Oracles
          </h3>

          <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', marginBottom: '0.75rem' }}>
            <strong>Overall Verdict:</strong>
            {renderVerdictBadge(execution.verification_verdict)}
            <span>(Verification Latency: {execution.duration_ms} ms)</span>
          </div>

          <div className="oracle-grid">
            {Object.entries(execution.oracles || {}).map(([key, oracle]) => {
              const status = oracle.status || (oracle.passed ? 'PASSED' : 'FAILED');
              let statusClass = 'fail';
              let statusLabel = '✗ FAILED';
              if (status === 'PASSED') {
                statusClass = 'pass';
                statusLabel = '✓ PASSED';
              } else if (status === 'NOT_RUN') {
                statusClass = 'not-run';
                statusLabel = '○ NOT RUN';
              } else if (status === 'UNAVAILABLE') {
                statusClass = 'unavailable';
                statusLabel = '⚠ UNAVAILABLE';
              }

              return (
                <div key={key} className={`oracle-card ${statusClass}`}>
                  <span className="oracle-name">{oracle.oracle_type}</span>
                  <span className={`oracle-status ${statusClass}`}>{statusLabel}</span>
                  <small style={{ color: '#aaa' }}>{oracle.description}</small>
                  {oracle.details && <small style={{ color: statusClass === 'pass' ? '#88ff88' : '#ff8888' }}>{oracle.details}</small>}
                  {oracle.container_info && (
                    <small style={{ color: '#00f3ff', fontSize: '0.7rem' }}>
                      Container: {oracle.container_info.container_id} ({oracle.container_info.network_policy})
                    </small>
                  )}
                </div>
              );
            })}
          </div>

          <div style={{ marginTop: '1rem', borderTop: '1px solid #222938', paddingTop: '0.75rem' }}>
            <p>
              <strong>Original Repository Immutability:</strong>{' '}
              <span style={{ color: execution.original_repo_immutable ? '#00ff88' : '#ff4444' }}>
                {execution.original_repo_immutable ? '100% BITWISE IDENTICAL (Zero In-Place Mutations)' : 'VIOLATION'}
              </span>
            </p>
            <small style={{ color: '#888', wordBreak: 'break-all' }}>
              SHA-256 Tree Pre: {execution.original_repo_pre_digest}
              <br />
              SHA-256 Tree Post: {execution.original_repo_post_digest}
            </small>
          </div>

          <div className="btn-group">
            <button className="action-btn danger" onClick={handleRollback} disabled={loading || !execution.rollback_available}>
              <RotateCcw size={14} style={{ marginRight: '6px', verticalAlign: 'middle' }} />
              ROLLBACK / DISCARD SANDBOX WORKSPACE
            </button>
            <button className="action-btn secondary" onClick={downloadAuditJSON} disabled={!auditReport}>
              <Download size={14} style={{ marginRight: '6px', verticalAlign: 'middle' }} />
              DOWNLOAD AUDIT REPORT (JSON)
            </button>
          </div>

          {rollbackStatus && (
            <p style={{ color: '#00ff88', marginTop: '0.5rem' }}>✓ {rollbackStatus}</p>
          )}
        </div>
      )}

      {/* STEP 6: MACHINE-READABLE AUDIT REPORT */}
      {auditReport && (
        <div className="remediation-step-card">
          <h3 className="step-title">
            <span className="step-number">6</span>
            Machine-Readable Audit Trail
          </h3>
          <p>
            Total Sessions: {auditReport.total_sessions_recorded} | Verified Repairs: {auditReport.verified_repairs_count} |{' '}
            Abstentions: {auditReport.abstained_repairs_count} | 100% Source Immutability: {auditReport.source_immutability_100_percent ? 'YES' : 'NO'}
          </p>
          <button className="action-btn secondary" onClick={downloadAuditJSON}>
            <Download size={14} style={{ marginRight: '6px', verticalAlign: 'middle' }} />
            DOWNLOAD FULL AUDIT LOG
          </button>
        </div>
      )}
    </div>
  );
}
