const analysisErrorMessage = (payload, fallback) => payload.error === 'analysis_unavailable'
  ? 'The analysis service is unavailable. Ask the administrator to start it, then retry.'
  : payload.detail || payload.error || fallback;

export const readEvolutionStream = async (response, onEvent) => {
  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(analysisErrorMessage(error, `Evolution failed (${response.status}).`));
  }
  if (!response.headers.get('content-type')?.includes('text/event-stream') || !response.body) {
    throw new Error('The analysis service did not return an event stream.');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let eventName = '';
  let data = [];
  let terminal = false;
  const processLine = (line) => {
    if (line === '') {
      if (data.length) {
        const payload = JSON.parse(data.join('\n'));
        if (eventName === 'error') throw new Error(analysisErrorMessage(payload, 'Evolution failed.'));
        onEvent(eventName, payload);
        if (eventName === 'complete') terminal = true;
      }
      eventName = '';
      data = [];
    } else if (line.startsWith('event:')) {
      eventName = line.slice(6).trim();
    } else if (line.startsWith('data:')) {
      data.push(line.slice(5).replace(/^ /, ''));
    }
  };
  try {
    while (true) {
      const { done, value } = await reader.read();
      buffer += decoder.decode(value, { stream: !done });
      let newline;
      while ((newline = buffer.indexOf('\n')) !== -1) {
        processLine(buffer.slice(0, newline).replace(/\r$/, ''));
        buffer = buffer.slice(newline + 1);
      }
      if (done) break;
    }
    if (!terminal) throw new Error('The evolution stream ended before completion. Please scan again before retrying.');
  } finally {
    await reader.cancel().catch(() => {});
    reader.releaseLock();
  }
};

export const scanEvolutionRepository = async (apiBase, token, repoURL, signal) => {
  const response = await fetch(`${apiBase}/analysis/scan`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ repo_url: repoURL, retain_workspace: true }),
    signal,
  });
  const payload = await response.json();
  if (!response.ok) {
    throw new Error(analysisErrorMessage(payload, 'Repository scan failed.'));
  }
  if (!payload.workspace_id || !Array.isArray(payload.targets)) {
    throw new Error('The analysis service did not return a usable workspace.');
  }
  return { ...payload, repo_url: repoURL };
};
