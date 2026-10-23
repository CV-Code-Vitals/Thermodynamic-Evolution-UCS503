export const fetchArchiveDownload = async (apiBase, downloadURL, token, fetchImpl = fetch) => {
  const locationOrigin = globalThis.location?.origin || 'http://localhost';
  const apiURL = new URL(apiBase, locationOrigin);
  const targetURL = new URL(downloadURL, apiURL);
  if (targetURL.origin !== apiURL.origin && targetURL.protocol !== 'https:') {
    throw new Error('The download URL must use HTTPS.');
  }
  const headers = targetURL.origin === apiURL.origin && token
    ? { Authorization: `Bearer ${token}` }
    : {};
  const response = await fetchImpl(targetURL.href, {
    headers,
    credentials: 'omit',
  });

  if (!response.ok) {
    const text = await response.text();
    let message = text;
    try {
      const payload = JSON.parse(text);
      message = payload.message || payload.error || text;
    } catch {
      // The API may return plain-text HTTP errors.
    }
    throw new Error(message || `Download failed (${response.status})`);
  }

  return response.blob();
};
