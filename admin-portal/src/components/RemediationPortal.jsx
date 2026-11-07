import React, { useState, useEffect } from 'react';
import {
  Shield, CheckCircle, AlertTriangle, FileCode, Play,
  Lock, ArrowRight, RotateCcw, Download, Terminal, Loader2
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
    case 'VERIFIED':
      return <span className="remediation-badge badge-verified">VERIFIED</span>;
    case 'EXPLICITLY_ABSTAINED':
    case 'ABSTAINED':
      return <span className="remediation-badge badge-abstain">ABSTAINED</span>;
    case 'FALSE_POSITIVE':
      return <span className="remediation-badge badge-false-positive">FALSE_POSITIVE</span>;
    case 'GROUND_TRUTH_UNVERIFIED':
    case 'GROUND_TRUTH_DISQUALIFIED':
      return <span className="remediation-badge badge-unverified">GROUND_TRUTH_UNVERIFIED</span>;
    default:
      return <span className="remediation-badge badge-review">{verdict}</span>;
  }
}

export default function RemediationPortal({ apiBase = '/api', token = '' }) {
  const [targets, setTargets] = useState([]);
  const [selectedTarget, setSelectedTarget] = useState('INJ-EXT2-RES-02');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Pipeline step state
  const [analysis, setAnalysis] = useState(null);
  const [proposal, setProposal] = useState(null);
  const [approval, setApproval] = useState(null);
  const [execution, setExecution] = useState(null);
  const [auditReport, setAuditReport] = useState(null);
  const [rollbackStatus, setRollbackStatus] = useState('');

  // Human approval form state
  const [operatorPrincipal, setOperatorPrincipal] = useState('lead-sre@code-vitals.internal');
  const [operatorRole, setOperatorRole] = useState('Principal Reliability Engineer');

  useEffect(() => {
    loadTargets();
  }, []);

  const loadTargets = async () => {
    try {
      const list = await fetchRemediationTargets(apiBase, token);
      setTargets(list);
    } catch (err) {
      // Fallback to pre-configured defaults if offline
      setTargets([
        {
          target_id: 'INJ-EXT2-RES-02',
          repository: 'fatih_color',
          defect_category: 'RESOURCE_LEAK',
          stratum: 'AUTHENTIC_REPO_INJECTED',
          description: 'Authentic Repo Injected Defect: os.Open file descriptor leak in DumpColorProfile',
        },
        {
          target_id: 'INJ-EXT2-SQL-01',
          repository: 'stretchr_objx',
          defect_category: 'SQL_INJECTION',
          stratum: 'AUTHENTIC_REPO_INJECTED',
          description: 'Authentic Repo Injected Defect: Dynamic SQL string formatting via fmt.Sprintf',
        },
        {
          target_id: 'INJ-EXT2-CMD-01',
          repository: 'schollz_progressbar',
          defect_category: 'COMMAND_INJECTION',
          stratum: 'AUTHENTIC_REPO_INJECTED',
          description: 'Authentic Repo Injected Defect: Unsanitized command executed via sh -c (Must safely abstain)',
        },
        {
          target_id: 'DEF-HOLD-01',
          repository: 'microservice_api',
          defect_category: 'UNHANDLED_ERROR',
          stratum: 'CONSTRUCTED_MICROSERVICE',
          description: 'Constructed Microservice Defect: json.Unmarshal error discarded in non-error return signature',
        },
        {
          target_id: 'NAT-03',
          repository: 'gorilla_websocket',
          defect_category: 'RESOURCE_LEAK',
          stratum: 'GROUND_TRUTH_DISQUALIFIED',
          description: 'Disqualified / False Alarm: DialContext manages socket in defer at line 330; response body does not require application close',
        },
      ]);
    }
  };

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
        proposal.proposal_id,
        operatorPrincipal,
        operatorRole
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
      setRollbackStatus('Sandbox successfully discarded. Original repository is 100% immutable.');
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
      const demoRes = await runRemediationDemo(apiBase, token);
      const audit = await fetchRemediationAuditReport(apiBase, token);
      setAuditReport(audit);
      setSelectedTarget('INJ-EXT2-RES-02');
      // Trigger normal flow for visual display
      await handleRunAnalysis();
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

  return (
    <div className="remediation-container">
      <div className="remediation-header">
        <h2>Autonomous Code Repair MVP</h2>
        <p>Controlled, Sandboxed AST-Level Remediation with Decoupled Independent Oracles & Cryptographic Authorization</p>
        <div style={{ display: 'flex', gap: '0.75rem', marginTop: '0.5rem' }}>
          <span className="remediation-badge badge-verified">BASELINE: de06b6e</span>
          <span className="remediation-badge badge-review">BRANCH: release/milestone-3.8-mvp-rc</span>
          <span className="remediation-badge badge-abstain">FAIL-CLOSED SAFETY ACTIVE</span>
        </div>
      </div>

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
        <p>Select a candidate defect target from the verified external pilot corpus:</p>
        <div style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap' }}>
          <select
            className="cyber-input"
            value={selectedTarget}
            onChange={(e) => {
              setSelectedTarget(e.target.value);
              resetPipeline();
            }}
            disabled={loading}
            style={{ minWidth: '320px' }}
          >
            {targets.map((t) => (
              <option key={t.target_id} value={t.target_id}>
                {t.target_id}: {t.repository} — {t.defect_category} [{t.stratum || (t.target_id === 'NAT-03' ? 'GROUND_TRUTH_DISQUALIFIED' : (t.target_id.startsWith('DEF-HOLD') ? 'CONSTRUCTED_MICROSERVICE' : 'AUTHENTIC_REPO_INJECTED'))}]
              </option>
            ))}
          </select>
          <button className="action-btn" onClick={handleRunAnalysis} disabled={loading}>
            {loading ? 'ANALYZING...' : 'RUN SEMANTIC ANALYSIS'}
          </button>
          <button className="action-btn secondary" onClick={handleQuickDemo} disabled={loading}>
            RUN ONE-CLICK DETERMINISTIC DEMO
          </button>
        </div>
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
              <span className={`remediation-badge ${analysis.target_id === 'NAT-03' ? 'badge-abstain' : (analysis.target_id.startsWith('DEF-HOLD') ? 'badge-review' : 'badge-verified')}`}>
                {analysis.target_id === 'NAT-03' ? 'GROUND_TRUTH_DISQUALIFIED' : (analysis.target_id.startsWith('DEF-HOLD') ? 'CONSTRUCTED_MICROSERVICE' : 'AUTHENTIC_REPO_INJECTED')}
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
                  <h4>Human Authorization Sign-Off Gate</h4>
                  <div style={{ display: 'flex', gap: '1rem', flexWrap: 'wrap', marginBottom: '0.75rem' }}>
                    <input
                      className="cyber-input"
                      type="text"
                      placeholder="Approver Principal ID"
                      value={operatorPrincipal}
                      onChange={(e) => setOperatorPrincipal(e.target.value)}
                    />
                    <input
                      className="cyber-input"
                      type="text"
                      placeholder="Operator Role"
                      value={operatorRole}
                      onChange={(e) => setOperatorRole(e.target.value)}
                    />
                  </div>
                  <button className="action-btn" onClick={handleApproveRepair} disabled={loading}>
                    <Lock size={14} style={{ marginRight: '6px', verticalAlign: 'middle' }} />
                    SIGN & AUTHORIZE CRYPTOGRAPHIC TOKEN
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
            {Object.entries(execution.oracles || {}).map(([key, oracle]) => (
              <div key={key} className={`oracle-card ${oracle.passed ? 'pass' : 'fail'}`}>
                <span className="oracle-name">{oracle.oracle_type}</span>
                <span className={`oracle-status ${oracle.passed ? 'pass' : 'fail'}`}>
                  {oracle.passed ? '✓ PASSED' : '✗ FAILED'}
                </span>
                <small style={{ color: '#888' }}>{oracle.description}</small>
                {oracle.details && <small style={{ color: '#ff8888' }}>{oracle.details}</small>}
              </div>
            ))}
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
