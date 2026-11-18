// =============================================================================
// admin-portal · src/remediation-api.js
// Client API library for Autonomous Code Repair MVP Release Candidate
// =============================================================================

function authHeaders(token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

export async function fetchRemediationTargets(apiBase = '/api', token = '') {
  const res = await fetch(`${apiBase}/remediation/targets`, {
    headers: authHeaders(token),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch targets: ${res.statusText}`);
  }
  const data = await res.json();
  return data.targets || [];
}

export async function analyzeRemediationTarget(apiBase = '/api', token = '', targetId) {
  const res = await fetch(`${apiBase}/remediation/analyze`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ target_id: targetId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `Analysis failed: ${res.statusText}`);
  }
  return res.json();
}

export async function proposeRemediationRepair(apiBase = '/api', token = '', analysisId) {
  const res = await fetch(`${apiBase}/remediation/propose`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ analysis_id: analysisId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `Proposal failed: ${res.statusText}`);
  }
  return res.json();
}

export async function approveRemediationRepair(apiBase = '/api', token = '', proposalId, principal = '', role = '') {
  const res = await fetch(`${apiBase}/remediation/approve`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ proposal_id: proposalId, principal, role }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `Approval failed: ${res.statusText}`);
  }
  return res.json();
}

export async function executeRemediationRepair(apiBase = '/api', token = '', tokenId) {
  const res = await fetch(`${apiBase}/remediation/execute`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ token_id: tokenId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `Execution failed: ${res.statusText}`);
  }
  return res.json();
}

export async function rollbackRemediationSandbox(apiBase = '/api', token = '', executionId) {
  const res = await fetch(`${apiBase}/remediation/rollback`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ execution_id: executionId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `Rollback failed: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchRemediationAuditReport(apiBase = '/api', token = '') {
  const res = await fetch(`${apiBase}/remediation/audit`, {
    headers: authHeaders(token),
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch audit report: ${res.statusText}`);
  }
  return res.json();
}

export async function runRemediationDemo(apiBase = '/api', token = '') {
  const res = await fetch(`${apiBase}/remediation/demo`, {
    method: 'POST',
    headers: authHeaders(token),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `Demo run failed: ${res.statusText}`);
  }
  return res.json();
}
