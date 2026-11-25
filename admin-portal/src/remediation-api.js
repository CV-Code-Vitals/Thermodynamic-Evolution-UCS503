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

async function request(url, options, operation) {
  let res;
  try {
    res = await fetch(url, options);
  } catch {
    throw new Error(`${operation} failed because the repair service could not be reached. Try again.`);
  }
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    if (res.status === 404 || data?.error === 'remediation_not_configured') {
      throw new Error('Autonomous Repair is unavailable on this server.');
    }
    if (res.status === 401 || res.status === 403) {
      throw new Error(`${operation} requires an active administrator session. Sign in again.`);
    }
    if (res.status >= 500) {
      throw new Error(`${operation} failed because the repair service returned HTTP ${res.status}. Try again.`);
    }
    const failure = new Error((typeof data?.error === 'string' && data.error.trim()) || `${operation} failed (HTTP ${res.status}).`);
    failure.execution = data?.execution;
    failure.verdict = data?.verification_verdict;
    throw failure;
  }
  if (data === null) {
    throw new Error(`${operation} returned an invalid response from the repair service.`);
  }
  return data;
}

export async function fetchRemediationTargets(apiBase = '/api', token = '', workspaceId = '') {
  const query = workspaceId ? `?workspace_id=${encodeURIComponent(workspaceId)}` : '';
  const data = await request(`${apiBase}/remediation/targets${query}`, {
    headers: authHeaders(token),
  }, 'Loading targets');
  if (!Array.isArray(data.targets)) {
    throw new Error('The repair service returned an invalid repository target list.');
  }
  return data.targets;
}

export async function analyzeRemediationTarget(apiBase = '/api', token = '', targetId) {
  return request(`${apiBase}/remediation/analyze`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ target_id: targetId }),
  }, 'Analysis');
}

export async function proposeRemediationRepair(apiBase = '/api', token = '', analysisId) {
  return request(`${apiBase}/remediation/propose`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ analysis_id: analysisId }),
  }, 'Proposal');
}

export async function approveRemediationRepair(apiBase = '/api', token = '', proposalId) {
  return request(`${apiBase}/remediation/approve`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ proposal_id: proposalId }),
  }, 'Approval');
}

export async function executeRemediationRepair(apiBase = '/api', token = '', tokenId) {
  return request(`${apiBase}/remediation/execute`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ token_id: tokenId }),
  }, 'Execution');
}

export async function rollbackRemediationSandbox(apiBase = '/api', token = '', executionId) {
  return request(`${apiBase}/remediation/rollback`, {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ execution_id: executionId }),
  }, 'Discard sandbox');
}

export async function fetchRemediationAuditReport(apiBase = '/api', token = '') {
  return request(`${apiBase}/remediation/audit`, {
    headers: authHeaders(token),
  }, 'Loading audit report');
}

export async function runRemediationDemo(apiBase = '/api', token = '') {
  return request(`${apiBase}/remediation/demo`, {
    method: 'POST',
    headers: authHeaders(token),
  }, 'Demo');
}
