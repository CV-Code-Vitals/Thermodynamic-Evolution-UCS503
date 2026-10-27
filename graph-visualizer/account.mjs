export function createAccountClient(apiBase, fetcher = fetch) {
  const base = apiBase.replace(/\/+$/, '');
  async function request(path, {token = '', method = 'GET', body} = {}) {
    const response = await fetcher(`${base}/account/${path}`, {
      method, credentials:'omit', headers:{...(token ? {Authorization:`Bearer ${token}`} : {}), ...(body ? {'Content-Type':'application/json'} : {})},
      ...(body ? {body:JSON.stringify(body)} : {}),
    });
    if (response.status === 204) return null;
    const data = await response.json().catch(() => null);
    if (!response.ok) {
      const messages = {user_signin_required:'Your session expired. Sign in again.', invalid_signin_ticket:'This sign-in link expired or was already used. Sign in again.', saved_repository_limit:'You can save up to 50 repositories. Remove a saved link first.', invalid_repository:'Enter a GitHub repository URL, such as https://github.com/owner/repository.', github_signin_not_configured:'GitHub sign-in is currently unavailable.', account_unavailable:'Your account service is temporarily unavailable. Try again shortly.'};
      const error = new Error(messages[data?.error] || 'The account request failed. Please try again.');
      error.status = response.status;
      throw error;
    }
    if (data === null) throw new Error('The account server returned an invalid response.');
    return data;
  }
  return {
    signInURL:`${base}/auth/github/start`,
    config:() => request('config'),
    exchange:(ticket, verifier) => request('session', {method:'POST', body:{ticket, verifier}}),
    me:token => request('me', {token}),
    repositories:token => request('repositories', {token}),
    save:(token, url) => request('repositories', {token, method:'POST', body:{url}}),
    remove:(token, id) => request(`repositories/${encodeURIComponent(id)}`, {token, method:'DELETE'}),
    logout:token => request('logout', {token, method:'POST'}),
  };
}

export async function mountAccountPanel(apiBase) {
  const $ = id => document.getElementById(id);
  const api = createAccountClient(apiBase);
  const key = `ucs503:personal:${new URL(apiBase, location.href).href}`;
  const read = name => { try { return sessionStorage.getItem(name) || ''; } catch { return ''; } };
  const store = (name, value) => { try { if (value) sessionStorage.setItem(name, value); else sessionStorage.removeItem(name); } catch { /* Session remains usable in memory when browser storage is blocked. */ } };
  let token = read(key), expiryTimer;
  const status = message => { $('account-status').textContent = message; };
  const signedOut = () => {
    clearTimeout(expiryTimer); token = ''; store(key, '');
    $('github-signin').hidden = false; $('github-signout').hidden = true;
    $('save-repository').hidden = true; $('saved-repositories').hidden = true;
    $('saved-list').replaceChildren(); $('account-identity').textContent = '';
  };
  async function savedList() {
    const repositories = await api.repositories(token);
    if (!Array.isArray(repositories)) throw new Error('Could not load saved repositories.');
    const list = $('saved-list'); list.replaceChildren();
    if (!repositories.length) { const empty = document.createElement('p'); empty.textContent = 'No saved repositories yet. Enter a repository URL above and choose Save repository.'; list.append(empty); }
    for (const repository of repositories) {
      const row = document.createElement('div'); row.className = 'saved-repository-row';
      const scan = document.createElement('button'); scan.type = 'button'; scan.textContent = `Scan ${repository.name}`;
      scan.onclick = () => { $('repository-url').value = repository.url; $('scan-form').requestSubmit(); };
      const remove = document.createElement('button'); remove.type = 'button'; remove.textContent = 'Remove'; remove.setAttribute('aria-label', `Remove ${repository.name}`);
      remove.onclick = () => act(async () => { await api.remove(token, repository.id); await savedList(); status('Saved link removed.'); });
      row.append(scan, remove); list.append(row);
    }
  }
  async function signedIn(session) {
    if (!session.user?.id || !session.user?.login || !Number.isFinite(session.expires_in_ms)) throw new Error('The account server returned an invalid session.');
    $('account-identity').textContent = `Signed in as ${session.user.login}`;
    $('github-signin').hidden = true; $('github-signout').hidden = false;
    $('save-repository').hidden = false; $('saved-repositories').hidden = false;
    clearTimeout(expiryTimer);
    expiryTimer = setTimeout(() => { signedOut(); status('Your session expired. Sign in again to access saved repositories.'); }, Math.max(0, session.expires_in_ms));
    await savedList();
    status('Save repository links to your account and return to scan them. Scanning currently requires public repositories.');
  }
  async function act(action) {
    try { await action(); } catch (error) {
      if (error.status === 401) signedOut();
      status(error.message || 'The account service is unavailable.');
    }
  }
  $('github-signin').onclick = () => act(async () => {
    const verifier = [...crypto.getRandomValues(new Uint8Array(32))].map(value => value.toString(16).padStart(2, '0')).join('');
    store(`${key}:verifier`, verifier);
    if (read(`${key}:verifier`) !== verifier) throw new Error('Allow session storage in your browser to complete GitHub sign-in.');
    const hash = new Uint8Array(await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier)));
    const challenge = btoa(String.fromCharCode(...hash)).replaceAll('+', '-').replaceAll('/', '_').replace(/=+$/, '');
    const destination = new URL(api.signInURL, location.href); destination.searchParams.set('handoff_challenge', challenge);
    store(`${key}:repository`, $('repository-url').value); location.assign(destination.href);
  });
  $('save-repository').onclick = () => act(async () => {
    $('save-repository').disabled = true;
    try { await api.save(token, $('repository-url').value); await savedList(); $('saved-repositories').open = true; status('Repository link saved to your account.'); }
    finally { $('save-repository').disabled = false; }
  });
  $('github-signout').onclick = () => act(async () => { await api.logout(token); signedOut(); status('Signed out. Your saved repository links remain in your account.'); });
  $('account-panel').hidden = false;
  const fragment = new URLSearchParams(location.hash.slice(1));
  const ticket = fragment.get('auth_ticket'), authError = fragment.get('auth_error');
  if (ticket || authError) history.replaceState(history.state, '', location.pathname+location.search);
  await act(async () => {
    const config = await api.config();
    $('github-signin').disabled = !config.github_enabled;
    if (!config.github_enabled) { signedOut(); status('GitHub sign-in is not configured yet. Public scanning remains available.'); return; }
    const pendingRepository = read(`${key}:repository`);
    if (pendingRepository) { $('repository-url').value = pendingRepository; store(`${key}:repository`, ''); }
    if (ticket) {
      const verifier = read(`${key}:verifier`); store(`${key}:verifier`, '');
      const session = await api.exchange(ticket, verifier);
      if (!/^[a-f0-9]{64}$/.test(session.token)) throw new Error('The account server returned an invalid session.');
      token = session.token; store(key, token); await signedIn(session);
    } else if (token) { await signedIn(await api.me(token)); }
    else status(authError ? 'GitHub sign-in did not complete. Please try again.' : 'Sign in with GitHub to save repository links. Public scanning works without signing in.');
  });
}
