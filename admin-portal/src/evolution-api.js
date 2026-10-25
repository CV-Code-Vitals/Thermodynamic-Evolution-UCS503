function parseErrorMessage(body, fallback) {
  if (!body) return fallback;
  try {
    const parsed = JSON.parse(body);
    return parsed.error || parsed.message || fallback;
  } catch {
    return body;
  }
}

function parseSSEFrame(frame) {
  let type = 'message';
  const data = [];

  for (const line of frame.split(/\r?\n/)) {
    if (line.startsWith('event:')) type = line.slice(6).trim();
    if (line.startsWith('data:')) data.push(line.slice(5).trimStart());
  }

  if (data.length === 0) return null;

  let parsed;
  try {
    parsed = JSON.parse(data.join('\n'));
  } catch {
    throw new Error('Evolution sent an invalid event payload.');
  }

  return { type, data: parsed };
}

export function createEvolutionApi(apiBase, fetchImpl = globalThis.fetch) {
  const base = typeof apiBase === 'string' ? apiBase.trim().replace(/\/+$/, '') : '';

  return {
    async start({ token, payload, signal, onEvent = () => {} }) {
      if (!base) throw new Error('Evolution analysis API is not configured.');
      if (!token) throw new Error('An authenticated Admin Portal session is required.');

      const response = await fetchImpl(`${base}/api/evolve`, {
        method: 'POST',
        headers: {
          Accept: 'text/event-stream',
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
        signal,
      });

      if (!response.ok) {
        const body = await response.text();
        throw new Error(`Evolution request failed (${response.status}): ${parseErrorMessage(body, response.statusText || 'Request rejected')}`);
      }
      if (!response.body) throw new Error('Evolution response did not include a stream.');

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let streamCompleted = false;

      const deliverFrame = (frame) => {
        const event = parseSSEFrame(frame);
        if (!event) return;
        onEvent(event);
        if (event.type === 'error') {
          throw new Error(event.data.error || event.data.message || 'Evolution failed.');
        }
      };

      try {
        while (true) {
          const { done, value } = await reader.read();
          if (done) {
            streamCompleted = true;
            buffer += decoder.decode();
            break;
          }

          buffer += decoder.decode(value, { stream: true });
          const frames = buffer.split(/\r?\n\r?\n/);
          buffer = frames.pop() || '';
          for (const frame of frames) deliverFrame(frame);
        }

        if (buffer.trim()) deliverFrame(buffer);
      } finally {
        if (!streamCompleted) {
          await reader.cancel().catch(() => {});
        }
        reader.releaseLock();
      }
    },
  };
}
