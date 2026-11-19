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

async function responseError(res, operation) {
  const data = await res.json().catch(() => ({}));
  if (res.status === 404 || data.error === 'remediation_not_configured') {
    return new Error('Autonomous Repair is unavailable on this server.');
  }
  return new Error(data.error || `${operation} failed (HTTP ${res.status}).`);
}

export async function fetchRemediationTargets(apiBase = '/api', token = '') {
  const res = await fetch(`${apiBase}/remediation/targets`, {
    headers: authHeaders(token),
  });
  if (!res.ok) {
    throw await responseError(res, 'Loading targets');
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
    throw await responseError(res, 'Analysis');
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
    throw await responseError(res, 'Proposal');
  }
  return res.json();
}

export async function approveRemediationRepair(apiBase = '/api', token = '', proposalId) {
  const res = await fetch(`${apiBase}/remediation/approve`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ proposal_id: proposalId }),
  });
  if (!res.ok) {
    throw await responseError(res, 'Approval');
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
    throw await responseError(res, 'Execution');
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
    throw await responseError(res, 'Rollback');
  }
  return res.json();
}

export async function fetchRemediationAuditReport(apiBase = '/api', token = '') {
  const res = await fetch(`${apiBase}/remediation/audit`, {
    headers: authHeaders(token),
  });
  if (!res.ok) {
    throw await responseError(res, 'Loading audit report');
  }
  return res.json();
}

export async function runRemediationDemo(apiBase = '/api', token = '') {
  const res = await fetch(`${apiBase}/remediation/demo`, {
    method: 'POST',
    headers: authHeaders(token),
  });
  if (!res.ok) {
    throw await responseError(res, 'Demo');
  }
  return res.json();
}
