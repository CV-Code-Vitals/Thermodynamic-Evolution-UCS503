const parseResponse = async (response) => {
  const text = await response.text();
  if (!text) return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return { message: text };
  }
};

const responseError = (payload, status) =>
  payload?.message || payload?.error || `Request failed (${status})`;

const archiveURL = (apiBase, path) => `${apiBase.replace(/\/$/, '')}${path}`;

const request = async (fetchImpl, url, options = {}) => {
  const response = await fetchImpl(url, options);
  const payload = await parseResponse(response);
  if (!response.ok) {
    const error = new Error(responseError(payload, response.status));
    error.code = payload?.error;
    error.status = response.status;
    throw error;
  }
  return payload;
};

const requireToken = (token) => {
  if (typeof token !== 'string' || token.length === 0) {
    throw new Error('An active admin session is required.');
  }
};

const authorizedHeaders = (token, additional = {}) => ({
  ...additional,
  Authorization: `Bearer ${token}`,
});

const resolveUploadURL = (apiBase, url) => {
  const locationOrigin = globalThis.location?.origin || 'http://localhost';
  const base = new URL(apiBase, locationOrigin);
  const target = new URL(url, base);
  if (target.origin !== base.origin && target.protocol !== 'https:') {
    throw new Error('The upload URL must use HTTPS.');
  }
  return target;
};

const isValidUpload = (upload) =>
  upload &&
  typeof upload.uploadId === 'string' &&
  upload.uploadId.length > 0 &&
  typeof upload.url === 'string' &&
  upload.url.length > 0 &&
  (upload.fields === undefined ||
    (upload.fields !== null &&
      typeof upload.fields === 'object' &&
      !Array.isArray(upload.fields) &&
      Object.values(upload.fields).every((value) => typeof value === 'string')));

const isArchiveEntry = (entry) =>
  entry &&
  typeof (entry.id || entry.filename) === 'string' &&
  entry.state === 'available';

const isAuthenticatedSession = (payload) =>
  payload?.status === 'authenticated' &&
  typeof payload.token === 'string' &&
  payload.token.length > 0 &&
  typeof payload.expiresAt === 'string' &&
  Number.isFinite(Date.parse(payload.expiresAt)) &&
  Number.isSafeInteger(payload.expiresInMs) &&
  payload.expiresInMs > 0;

export const getSessionRefreshDelay = (expiresInMs, canRefresh) => {
  if (!Number.isSafeInteger(expiresInMs) || expiresInMs < 0) {
    throw new Error('Authentication service returned an invalid session lifetime.');
  }
  return canRefresh ? Math.max(0, expiresInMs - 60_000) : expiresInMs;
};

export const createArchiveApi = (apiBase, fetchImpl = fetch) => {
  const authenticatedJSON = async (token, path, method = 'GET', body) => {
    requireToken(token);
    const headers = authorizedHeaders(token);
    if (body !== undefined) headers['Content-Type'] = 'application/json';
    return request(fetchImpl, archiveURL(apiBase, path), {
      method,
      headers,
      credentials: 'omit',
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
  };

  const login = async (passkey) => {
    const payload = await request(fetchImpl, archiveURL(apiBase, '/login'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'omit',
      body: JSON.stringify({ passkey }),
    });
    if (!isAuthenticatedSession(payload)) {
      throw new Error('Authentication service returned an invalid response.');
    }
    return payload;
  };

  const refresh = async (token) => {
    const payload = await authenticatedJSON(token, '/refresh', 'POST');
    if (!isAuthenticatedSession(payload)) {
      throw new Error('Authentication service returned an invalid response.');
    }
    return payload;
  };

  const beginUpload = async (token, metadata) => {
    const upload = await authenticatedJSON(token, '/upload', 'POST', metadata);
    if (!isValidUpload(upload)) {
      throw new Error('Archive service returned an invalid upload response.');
    }
    return upload;
  };

  const completeUpload = async (token, uploadId) => {
    const entry = await authenticatedJSON(token, '/upload/complete', 'POST', { uploadId });
    if (!isArchiveEntry(entry)) {
      throw new Error('Archive service returned an invalid completed upload.');
    }
    return entry;
  };

  const uploadFile = async (token, metadata, file) => {
    const upload = await beginUpload(token, metadata);
    const target = resolveUploadURL(apiBase, upload.url);
    const apiOrigin = new URL(apiBase, globalThis.location?.origin || 'http://localhost').origin;
    let options;

    if (target.origin === apiOrigin) {
      options = {
        method: 'POST',
        headers: authorizedHeaders(token, { 'Content-Type': metadata.contentType }),
        body: file,
        credentials: 'omit',
      };
    } else {
      const form = new FormData();
      for (const [name, value] of Object.entries(upload.fields || {})) {
        form.append(name, value);
      }
      form.append('file', file, file.name);
      options = { method: 'POST', body: form, credentials: 'omit' };
    }

    const uploadResponse = await fetchImpl(target.href, options);
    const uploadPayload = await parseResponse(uploadResponse);
    if (!uploadResponse.ok) {
      throw new Error(responseError(uploadPayload, uploadResponse.status));
    }
    return upload.uploadId;
  };

  const list = async (token) => {
    const entries = await authenticatedJSON(token, '/deliverables');
    if (!Array.isArray(entries) || !entries.every(isArchiveEntry)) {
      throw new Error('Archive service returned an invalid list response.');
    }
    return entries;
  };

  const createDownloadUrl = async (token, id) => {
    const download = await authenticatedJSON(
      token,
      `/download?filename=${encodeURIComponent(id)}`,
    );
    if (typeof download?.url !== 'string' || download.url.length === 0) {
      throw new Error('Archive service returned an invalid download response.');
    }
    const target = resolveUploadURL(apiBase, download.url);
    return { ...download, url: target.href };
  };

  const deleteArchive = async (token, id) => {
    await authenticatedJSON(token, `/archive?filename=${encodeURIComponent(id)}`, 'DELETE');
  };

  return {
    login,
    refresh,
    beginUpload,
    completeUpload,
    upload: uploadFile,
    list,
    createDownloadUrl,
    deleteArchive,
  };
};
