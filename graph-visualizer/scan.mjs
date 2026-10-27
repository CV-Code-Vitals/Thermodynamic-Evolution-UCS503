export async function scanPublicRepository({ apiBase, analysisApiBase }, repoURL, signal) {
  const endpoint = analysisApiBase
    ? `${analysisApiBase.replace(/\/+$/, '')}/api/scan`
    : `${apiBase.replace(/\/+$/, '')}/analysis/scan`;
  const response = await fetch(endpoint, {
    method: 'POST',
    credentials: 'omit',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ repo_url: repoURL.trim() }),
    signal,
  });
  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    const messages = {
      analysis_not_configured: 'Repository scanning is not configured on this server.',
      analysis_unavailable: 'The analysis service is unavailable. Please try again later.',
      scan_rate_limited: 'Please wait a minute before scanning another repository.',
      scan_capacity_reached: 'The scanner is busy. Please try again shortly.',
      invalid_repository: 'Enter a public GitHub repository URL, such as https://github.com/owner/repository.',
      invalid_request: 'Enter a valid public GitHub repository URL.',
    };
    throw new Error(messages[payload?.error] || payload?.detail || payload?.error || `Repository scan failed (${response.status}).`);
  }
  const report = payload?.report || payload;
  if (!Array.isArray(report?.file_reports)) throw new Error('The scanner did not return a repository report.');
  return report;
}
