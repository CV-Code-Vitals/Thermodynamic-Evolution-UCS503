import React, { useState, useEffect, useCallback } from 'react';
import { useDropzone } from 'react-dropzone';
import {
  Shield, Lock, UploadCloud, File, CheckCircle, Trash2,
  Terminal, Loader2, Database, Download, LayoutDashboard,
  Network, GitBranch, ArrowRight
} from 'lucide-react';
import './AdminPortal.css';
import { fetchArchiveDownload } from '../archive-download';
import { createArchiveApi, getSessionRefreshDelay } from '../archive-api';
import EvolutionWorkspace from './EvolutionWorkspace';
import GraphVisualizer from './GraphVisualizer';

const apiBase = (import.meta.env.VITE_API_BASE || '/api').replace(/\/$/, '');
const archiveApi = createArchiveApi(apiBase);
const analysisApiBase = (import.meta.env.VITE_ANALYSIS_API_BASE || '').replace(/\/+$/, '');
const acceptedContentTypes = {
  '.pdf': 'application/pdf',
  '.ppt': 'application/vnd.ms-powerpoint',
  '.pptx': 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
};

const getContentType = (file) =>
  file.type || acceptedContentTypes[file.name.slice(file.name.lastIndexOf('.')).toLowerCase()] || '';

// ─────────────────────────────────────────────────────────────────────────────
// AdminPortal — top-level component managing all portal views
// ─────────────────────────────────────────────────────────────────────────────

const AdminPortal = () => {

  // ── Auth ──────────────────────────────────────────────────────────────────
  const [token, setToken] = useState('');
  const [expiresAt, setExpiresAt] = useState('');
  const [expiresInMs, setExpiresInMs] = useState(0);
  const [canRefresh, setCanRefresh] = useState(true);
  const isAuthenticated = Boolean(token);
  const [password, setPassword] = useState('');
  const [authError, setAuthError] = useState('');

  // ── Active view: dashboard, upload, archive, or deferred evolution ───────
  const [activeView, setActiveViewState] = useState(() => {
    const requestedView = window.location.hash.slice(1);
    return ['dashboard', 'evolve', 'graph', 'upload', 'archive'].includes(requestedView)
      ? requestedView : 'dashboard';
  });
  const setActiveView = useCallback((view) => {
    setActiveViewState(view);
    const nextURL = new URL(window.location.href);
    nextURL.hash = view;
    window.history.replaceState(window.history.state, '', nextURL);
  }, []);
  const [latestScan, setLatestScan] = useState(null);

  // ── Dropzone / file state ─────────────────────────────────────────────────
  const [selectedFile, setSelectedFile] = useState(null);
  const [fileError, setFileError] = useState('');

  // ── Metadata form ─────────────────────────────────────────────────────────
  const [formData, setFormData] = useState({
    title: '',
    version: '',
    date: new Date().toISOString().split('T')[0],
    summary: '',
  });

  // ── Publish state ─────────────────────────────────────────────────────────
  const [isPublishing, setIsPublishing] = useState(false);
  const [publishSuccess, setPublishSuccess] = useState(false);
  const [publishError, setPublishError] = useState('');
  const [pendingUploadId, setPendingUploadId] = useState('');
  const [pendingUploadUnavailable, setPendingUploadUnavailable] = useState(false);
  // The full Deliverable object returned by the backend after a successful upload
  const [lastDeliverable, setLastDeliverable] = useState(null);

  // ── Archive / deliverables list ───────────────────────────────────────────
  const [deliverables, setDeliverables] = useState([]);
  const [archiveLoading, setArchiveLoading] = useState(false);
  const [archiveError, setArchiveError] = useState('');
  const [downloadingId, setDownloadingId] = useState('');
  const [deletingId, setDeletingId] = useState('');

  // ─────────────────────────────────────────────────────────────────────────
  // Auth
  // ─────────────────────────────────────────────────────────────────────────

  const handleLogin = async (e) => {
    e.preventDefault();
    setAuthError('');
    if (!password.trim()) {
      setAuthError('ENTER YOUR CONFIGURED ADMIN PASSKEY.');
      return;
    }

    try {
      const result = await archiveApi.login(password);
      setToken(result.token);
      setExpiresAt(result.expiresAt);
      setExpiresInMs(result.expiresInMs);
      setCanRefresh(true);
      setPassword('');
    } catch (error) {
      setAuthError(error.message || 'AUTHENTICATION SERVICE UNAVAILABLE. PLEASE TRY AGAIN.');
    }
  };

  const clearAdminSession = useCallback((message = '') => {
    setToken('');
    setExpiresAt('');
    setExpiresInMs(0);
    setCanRefresh(true);
    setAuthError(message);
    setPassword('');
    setDeliverables([]);
    setSelectedFile(null);
    setPublishSuccess(false);
    setPendingUploadId('');
    setPendingUploadUnavailable(false);
    setLastDeliverable(null);
    setLatestScan(null);
    setActiveView('dashboard');
  }, [setActiveView]);

  const handleLogout = () => clearAdminSession('');

  useEffect(() => {
    if (!token || !expiresAt) return undefined;

    const expiration = Date.parse(expiresAt);
    if (!Number.isFinite(expiration)) {
      clearAdminSession('SESSION EXPIRED. PLEASE SIGN IN AGAIN.');
      return undefined;
    }

    let active = true;
    const expireSession = () => {
      if (!active) return;
      clearAdminSession('SESSION EXPIRED. PLEASE SIGN IN AGAIN.');
    };

    const timeout = setTimeout(async () => {
      if (!canRefresh) {
        expireSession();
        return;
      }
      try {
        const refreshed = await archiveApi.refresh(token);
        if (!active) return;
        if (Date.parse(refreshed.expiresAt) <= expiration) setCanRefresh(false);
        setToken(refreshed.token);
        setExpiresAt(refreshed.expiresAt);
        setExpiresInMs(refreshed.expiresInMs);
      } catch {
        expireSession();
      }
    }, getSessionRefreshDelay(expiresInMs, canRefresh));

    return () => {
      active = false;
      clearTimeout(timeout);
    };
  }, [token, expiresAt, expiresInMs, canRefresh, clearAdminSession]);


  // ─────────────────────────────────────────────────────────────────────────
  // Dropzone
  // react-dropzone requires BOTH the correct MIME type key AND the extension
  // in the accept map. Providing both .ppt and .pptx keys fixes the rejection.
  // ─────────────────────────────────────────────────────────────────────────

  const onDrop = useCallback((acceptedFiles, rejectedFiles) => {
    setFileError('');

    // Surface a useful message when the browser rejects a file type
    if (rejectedFiles.length > 0) {
      const reason = rejectedFiles[0].errors?.[0]?.message || 'Invalid file type.';
      setFileError(`FILE REJECTED: ${reason.toUpperCase()}`);
      return;
    }

    if (acceptedFiles.length > 0) {
      const file = acceptedFiles[0];
      if (file.size > 50 * 1024 * 1024) {
        setFileError('FILE EXCEEDS MAXIMUM ALLOWED SIZE (50 MiB).');
        return;
      }
      setSelectedFile(file);
      // Auto-fill the title from the filename (strip extension, clean separators)
      setFormData(prev => ({
        ...prev,
        title: file.name.replace(/\.[^.]+$/, '').replace(/[-_]/g, ' '),
      }));
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    multiple: false,
    disabled: Boolean(pendingUploadId),
    accept: {
      // ── PowerPoint (both legacy .ppt and modern .pptx) ───────────────────
      // react-dropzone validates by MIME type; the browser may report either
      // MIME for a .pptx, so we list all known variants to be safe.
      'application/vnd.ms-powerpoint': ['.ppt'],
      'application/vnd.openxmlformats-officedocument.presentationml.presentation': ['.pptx'],
      // ── Other document formats ────────────────────────────────────────────
      'application/pdf': ['.pdf'],
    },
    maxSize: 50 * 1024 * 1024,
  });

  // ─────────────────────────────────────────────────────────────────────────
  // Form
  // ─────────────────────────────────────────────────────────────────────────

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  // ─────────────────────────────────────────────────────────────────────────
  // Publish — authorize, upload directly to storage, then finalize.
  // ─────────────────────────────────────────────────────────────────────────

  const handlePublish = async () => {
    if (!selectedFile || !formData.title || !formData.version) return;

    setIsPublishing(true);
    setPublishError('');

    try {
      let uploadId = pendingUploadId;
      if (!uploadId) {
        uploadId = await archiveApi.upload(token, {
          originalName: selectedFile.name,
          contentType: getContentType(selectedFile),
          size: selectedFile.size,
          ...formData,
        }, selectedFile);
        setPendingUploadId(uploadId);
      }
      const result = await archiveApi.completeUpload(token, uploadId);
      setPendingUploadId('');
      setPendingUploadUnavailable(false);
      setLastDeliverable(result);
      setPublishSuccess(true);
    } catch (error) {
      if (error.code === 'archive_not_found' || error.code === 'archive_conflict') {
        setPendingUploadUnavailable(true);
        setPublishError('THIS UPLOAD IS NO LONGER AVAILABLE TO FINALIZE. RESET IT TO START A NEW UPLOAD.');
      } else {
        setPublishError(error.message || 'UPLOAD FAILED. PLEASE TRY AGAIN.');
      }
    } finally {
      setIsPublishing(false);
    }
  };

  // ─────────────────────────────────────────────────────────────────────────
  // Archive — fetch deliverables from the Go backend
  // ─────────────────────────────────────────────────────────────────────────

  const fetchDeliverables = useCallback(async () => {
    setArchiveLoading(true);
    setArchiveError('');
    try {
      setDeliverables(await archiveApi.list(token));
    } catch (error) {
      setArchiveError(error.message || 'Unable to load archive. Please try again.');
      setDeliverables([]);
    } finally {
      setArchiveLoading(false);
    }
  }, [token]);

  const handleDownload = async (deliverable) => {
    const id = deliverable.id || deliverable.filename;
    setArchiveError('');
    setDownloadingId(id);
    try {
      const download = await archiveApi.createDownloadUrl(token, id);
      const blob = await fetchArchiveDownload(apiBase, download.url, token);
      const objectUrl = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = objectUrl;
      link.download = deliverable.originalName || deliverable.filename;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.setTimeout(() => URL.revokeObjectURL(objectUrl), 30000);
    } catch (error) {
      setArchiveError(error.message || 'Unable to download archive. Please try again.');
    } finally {
      setDownloadingId('');
    }
  };

  const handleDelete = async (deliverable) => {
    const id = deliverable.id || deliverable.filename;
    if (!window.confirm(`Permanently delete "${deliverable.originalName || deliverable.title || id}"?`)) return;
    setArchiveError('');
    setDeletingId(id);
    try {
      await archiveApi.deleteArchive(token, id);
      await fetchDeliverables();
    } catch (error) {
      setArchiveError(error.message || 'Unable to delete archive. Please try again.');
    } finally {
      setDeletingId('');
    }
  };

  // Reload archive data whenever the user switches to the archive view
  useEffect(() => {
    if (activeView === 'archive' && isAuthenticated) {
      fetchDeliverables();
    }
  }, [activeView, isAuthenticated, fetchDeliverables]);

  // ─────────────────────────────────────────────────────────────────────────
  // Reset upload flow
  // ─────────────────────────────────────────────────────────────────────────

  const resetUpload = () => {
    setSelectedFile(null);
    setFileError('');
    setPublishError('');
    setPublishSuccess(false);
    setLastDeliverable(null);
    setPendingUploadId('');
    setPendingUploadUnavailable(false);
    setFormData({
      title: '',
      version: '',
      date: new Date().toISOString().split('T')[0],
      summary: '',
    });
  };

  // ─────────────────────────────────────────────────────────────────────────
  // RENDER: Login screen
  // ─────────────────────────────────────────────────────────────────────────

  if (!isAuthenticated) {
    return (
      <div className="admin-portal-wrapper">
        <div className="portal-container">
          <div className="portal-header">
            <div className="header-title"><Shield size={20} /> UCS503 // SECURE ACCESS</div>
            <div className="status-indicator">
              <div className="status-dot"></div>
              ENCRYPTED CONNECTION
            </div>
          </div>

          <div className="auth-container">
            <Lock size={48} className="auth-icon" />
            <h1 className="auth-title">Project workspace</h1>
            <p className="auth-subtitle">SIGN IN TO ACCESS REPOSITORY ANALYSIS AND DELIVERABLES</p>

            <form className="auth-form" onSubmit={handleLogin}>
              <div className="input-group">
                <label className="input-label">Passkey</label>
                <input
                  type="password"
                  className="cyber-input"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="• • • • • • • •"
                  autoComplete="current-password"
                  autoFocus
                  required
                />
                <span style={{ fontSize: '0.75rem', color: '#71717a', marginTop: '6px', display: 'block', letterSpacing: '0.05em' }}>
                  Enter the administrator passkey configured on the API server.
                </span>
                {authError && <div className="error-text">{authError}</div>}
              </div>
              <button type="submit" className="cyber-button">INITIALIZE SESSION</button>
            </form>
          </div>
        </div>
      </div>
    );
  }

  // ─────────────────────────────────────────────────────────────────────────
  // RENDER: Success screen
  // ─────────────────────────────────────────────────────────────────────────

  if (publishSuccess && lastDeliverable) {
    return (
      <div className="admin-portal-wrapper">
        <div className="portal-container">
          <div className="portal-header">
            <div className="header-title"><Terminal size={20} /> SYS.ARCHIVE // UPLOAD COMPLETE</div>
            <button type="button" className="cyber-button-sm" onClick={handleLogout}>LOG OUT</button>
          </div>

          <div className="success-container">
            <CheckCircle size={64} className="success-icon" />
            <h1 className="success-title">Archive Generated</h1>
            <p className="success-text">
              {/* Dynamic values from the backend response — no hardcoding */}
              File&nbsp;<strong>[ {lastDeliverable.filename} ]</strong>&nbsp;committed to storage matrix.<br />
              Version&nbsp;<strong>{lastDeliverable.version}</strong>&nbsp;is now archived.<br />
              Existing archive records remain available.
            </p>

            <button
              type="button"
              onClick={() => handleDownload(lastDeliverable)}
              className="cyber-button"
              style={{ marginBottom: '1rem', textDecoration: 'none' }}
              disabled={Boolean(downloadingId)}
            >
              {downloadingId
                ? <><Loader2 className="spinner" size={16} /> PREPARING DOWNLOAD...</>
                : 'ACCESS ARCHIVED FILE ↗'}
            </button>
            {archiveError && (
              <div className="error-text" style={{ marginBottom: '1rem' }}>{archiveError}</div>
            )}

            <div style={{ display: 'flex', gap: '1rem', width: '100%' }}>
              <button className="cyber-button" style={{ flex: 1 }} onClick={resetUpload}>
                NEW UPLOAD
              </button>
              <button
                className="cyber-button"
                style={{ flex: 1 }}
                onClick={() => { resetUpload(); setActiveView('archive'); }}
              >
                VIEW ARCHIVE
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // ─────────────────────────────────────────────────────────────────────────
  // RENDER: Main dashboard (Upload tab + Archive tab)
  // ─────────────────────────────────────────────────────────────────────────

  return (
    <div className="admin-portal-wrapper">
      <div className={`portal-container ${activeView === 'graph' ? 'portal-container-wide' : ''}`}>

        {/* ── Header with tab navigation ──────────────────────────────── */}
        <div className="portal-header">
          <div className="header-title"><Terminal size={20} /> UCS503 // PROJECT WORKSPACE</div>
          <div className="status-indicator">
            <div className="status-dot"></div>
            AUTHORIZED : ADMIN
          </div>
          <button type="button" className="cyber-button-sm" onClick={handleLogout}>LOG OUT</button>
        </div>

        {/* ── Tab switcher ─────────────────────────────────────────────── */}
        <div className="tab-bar">
          <button
            className={`tab-btn ${activeView === 'dashboard' ? 'active' : ''}`}
            onClick={() => setActiveView('dashboard')}
          >
            <LayoutDashboard size={14} /> OVERVIEW
          </button>
          <button
            className={`tab-btn ${activeView === 'evolve' ? 'active' : ''}`}
            onClick={() => setActiveView('evolve')}
          >
            <GitBranch size={14} /> SCAN & OPTIMIZE
          </button>
          <button
            className={`tab-btn ${activeView === 'graph' ? 'active' : ''}`}
            onClick={() => setActiveView('graph')}
          >
            <Network size={14} /> GRAPH
          </button>
          <button
            className={`tab-btn ${activeView === 'upload' ? 'active' : ''}`}
            onClick={() => setActiveView('upload')}
          >
            <UploadCloud size={14} /> UPLOAD
          </button>
          <button
            className={`tab-btn ${activeView === 'archive' ? 'active' : ''}`}
            onClick={() => setActiveView('archive')}
          >
            <Database size={14} /> VIEW ARCHIVE
          </button>
        </div>

        {activeView === 'dashboard' && (
          <div className="dashboard-container">
            <div className="dashboard-heading">
              <div>
                <span className="archive-title">THERMODYNAMIC EVOLUTION · UCS503</span>
                <h1>Project workspace</h1>
                <p>Scan a repository, inspect its structure, optimize Go functions, and manage deliverables after one sign in.</p>
              </div>
              <div className="online-badge"><span className="status-dot" /> AUTHENTICATED</div>
            </div>
            <div className="dashboard-hero">
              <div className="hero-icon"><GitBranch size={28} /></div>
              <div>
                <span className="hero-kicker">REPOSITORY ANALYSIS</span>
                <h2>{latestScan ? 'Latest repository scan is ready.' : 'Start with a repository scan.'}</h2>
                <p>{latestScan
                  ? `${latestScan.report.files_analyzed} files analyzed · ${latestScan.report.total_hotspots} hotspots found`
                  : 'The same scan powers the graph and the Go function optimizer.'}</p>
              </div>
            </div>
            <div className="workspace-metrics">
              <div className="workspace-metric"><span>ANALYZED FILES</span><strong>{latestScan?.report.files_analyzed ?? '—'}</strong></div>
              <div className="workspace-metric"><span>HOTSPOTS</span><strong>{latestScan?.report.total_hotspots ?? '—'}</strong></div>
              <div className="workspace-metric"><span>ANALYSIS ENGINE</span><strong>{latestScan?.report.engine_version ?? 'Ready'}</strong></div>
            </div>
            {latestScan && (
              <p className="archive-note workspace-scan-origin">
                Scanned repository: <span>{latestScan.report.scanned_directory}</span>
              </p>
            )}
            <div className="dashboard-actions">
              <button className="cyber-button" onClick={() => setActiveView('evolve')}>
                <GitBranch size={18} /> SCAN & OPTIMIZE <ArrowRight size={16} />
              </button>
              <button className="cyber-button" onClick={() => setActiveView('graph')}>
                <Network size={18} /> OPEN GRAPH
              </button>
              <button className="cyber-button" onClick={() => setActiveView('upload')}>
                <UploadCloud size={18} /> UPLOAD DELIVERABLE
              </button>
              <button className="cyber-button" onClick={() => setActiveView('archive')}>
                <Database size={18} /> VIEW ARCHIVE
              </button>
            </div>
            <p className="workspace-limit-note">Optimization runs require the server sandbox to be available. The graph uses the authenticated scan shown here.</p>
          </div>
        )}

        {/* ══════════════════════════════════════════════════════════════ */}
        {/* UPLOAD VIEW                                                   */}
        {/* ══════════════════════════════════════════════════════════════ */}
        {activeView === 'upload' && (
          <div className="dashboard-container">

            {/* Dropzone */}
            <div
              {...getRootProps()}
              className={`dropzone-area ${isDragActive ? 'active' : ''} ${selectedFile ? 'has-file' : ''}`}
            >
              <input {...getInputProps()} />
              {selectedFile ? (
                <div className="file-info">
                  <File size={32} />
                  <span>{selectedFile.name}</span>
                </div>
              ) : (
                <>
                  <UploadCloud size={48} className="upload-icon" />
                  <div className="dropzone-text">
                    {isDragActive ? 'DEPLOY FILE HERE' : 'DRAG & DROP ARCHIVE OR CLICK TO BROWSE'}
                  </div>
                  <div className="dropzone-subtext">SUPPORTED: .PPTX · .PPT · .PDF · MAXIMUM 50 MiB</div>
                </>
              )}
            </div>
            <p className="archive-note">
              File type and size are validated; files are not malware-scanned.
            </p>

            {fileError && (
              <div className="error-text" style={{ textAlign: 'center', marginBottom: '1.5rem' }}>
                {fileError}
              </div>
            )}

            {/* Metadata form — only visible once a file is selected */}
            {selectedFile && (
              <>
                {pendingUploadId && (
                  <p className="archive-note">
                    File uploaded. Retrying will finalize the same upload; it will not upload the file again.
                  </p>
                )}
                <div className="metadata-form">
                  <div className="input-group">
                    <label className="input-label">Document Title</label>
                    <input type="text" name="title" className="cyber-input"
                      value={formData.title} onChange={handleInputChange}
                      placeholder="Enter document title" />
                  </div>

                  <div className="input-group">
                    <label className="input-label">Presentation Version</label>
                    <input type="text" name="version" className="cyber-input"
                      value={formData.version} onChange={handleInputChange}
                      placeholder="e.g. v1.2.0" />
                  </div>

                  <div className="input-group">
                    <label className="input-label">Archive Date</label>
                    <input type="date" name="date" className="cyber-input"
                      value={formData.date} onChange={handleInputChange} />
                  </div>

                  <div className="input-group full-width">
                    <label className="input-label">Change Summary</label>
                    <textarea name="summary" className="cyber-input cyber-textarea"
                      value={formData.summary} onChange={handleInputChange}
                      placeholder="Brief description of modifications..." />
                  </div>
                </div>

                {publishError && (
                  <div className="error-text" style={{ textAlign: 'center', marginBottom: '1rem' }}>
                    {publishError}
                  </div>
                )}

                {pendingUploadUnavailable && (
                  <button
                    type="button"
                    className="cyber-button"
                    style={{ width: '100%', marginBottom: '1rem' }}
                    onClick={resetUpload}
                  >
                    RESET UPLOAD AND START OVER
                  </button>
                )}

                <button
                  className="cyber-button"
                  style={{ width: '100%' }}
                  onClick={handlePublish}
                  disabled={!formData.title || !formData.version || isPublishing || pendingUploadUnavailable}
                >
                  {isPublishing
                    ? <><Loader2 className="spinner" size={20} /> TRANSMITTING TO BACKEND...</>
                    : pendingUploadId ? 'RETRY FINALIZATION' : 'PUBLISH TO ARCHIVE'}
                </button>
              </>
            )}
          </div>
        )}

        {/* ══════════════════════════════════════════════════════════════ */}
        {/* ARCHIVE VIEW                                                  */}
        {/* ══════════════════════════════════════════════════════════════ */}
        {activeView === 'archive' && (
          <div className="dashboard-container">

            <div className="archive-header-row">
              <span className="archive-title">DELIVERABLES ARCHIVE</span>
              <button className="cyber-button-sm" onClick={fetchDeliverables} disabled={archiveLoading}>
                {archiveLoading ? <Loader2 className="spinner" size={14} /> : '⟳ REFRESH'}
              </button>
            </div>

            {archiveError && (
              <div className="error-text" style={{ marginBottom: '1rem' }}>{archiveError}</div>
            )}

            {archiveLoading && (
              <div className="archive-loading">
                <Loader2 className="spinner" size={28} style={{ color: 'var(--gold)' }} />
                <span>QUERYING ARCHIVE MATRIX...</span>
              </div>
            )}

            {!archiveLoading && deliverables.length === 0 && (
              <div className="archive-empty">
                <Database size={40} style={{ color: 'var(--text-muted)', marginBottom: '0.75rem' }} />
                <p>NO RECORDS FOUND IN ARCHIVE.</p>
                <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>
                  Upload a deliverable to begin.
                </p>
              </div>
            )}

            {!archiveLoading && deliverables.length > 0 && (
              <div className="archive-table-wrapper">
                <table className="archive-table">
                  <thead>
                    <tr>
                      <th>#</th>
                      <th>TITLE</th>
                      <th>VERSION</th>
                      <th>DATE</th>
                      <th>CHANGE SUMMARY</th>
                      <th>FILE</th>
                      <th>ACTIONS</th>
                    </tr>
                  </thead>
                  <tbody>
                    {deliverables.map((d, index) => {
                      const dateStr = d.date || (d.uploaded_at ? new Date(d.uploaded_at).toLocaleDateString() : (d.uploadedAt ? new Date(d.uploadedAt).toLocaleDateString() : '—'));
                      return (
                        <tr key={d.id || d.filename || index}>
                          <td className="archive-id">{String(index + 1).padStart(3, '0')}</td>
                          <td className="archive-title-cell">{d.title || d.originalName || d.filename}</td>
                          <td className="archive-version">{d.version || '1.0.0'}</td>
                          <td className="archive-date">{dateStr}</td>
                          <td className="archive-summary">{d.summary || '—'}</td>
                          <td>
                            <button
                              type="button"
                              onClick={() => handleDownload(d)}
                              className="archive-download-link"
                              title={d.filename}
                              disabled={Boolean(downloadingId)}
                            >
                              {downloadingId === (d.id || d.filename)
                                ? <Loader2 className="spinner" size={14} />
                                : <Download size={14} />}
                              {downloadingId === (d.id || d.filename) ? 'PREPARING' : 'DOWNLOAD'}
                            </button>
                          </td>
                          <td>
                            <button
                              type="button"
                              onClick={() => handleDelete(d)}
                              className="archive-download-link"
                              disabled={Boolean(deletingId)}
                              aria-label={`Delete ${d.originalName || d.title || d.filename}`}
                            >
                              <Trash2 size={14} />
                              {deletingId === (d.id || d.filename) ? 'DELETING' : 'DELETE'}
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        )}

        {/* ══════════════════════════════════════════════════════════════ */}
        {/* EVOLUTION VIEW                                                */}
        {/* ══════════════════════════════════════════════════════════════ */}
        <section hidden={activeView !== 'evolve'}>
          <EvolutionWorkspace apiBase={apiBase} analysisApiBase={analysisApiBase} token={token} initialScan={latestScan} onScanComplete={setLatestScan} />
        </section>

        {activeView === 'graph' && (
          <div className="graph-workspace">
            <div className="graph-workspace-heading">
              <div>
                <span className="archive-title">REPOSITORY STRUCTURE</span>
                <h1>Graph explorer</h1>
                <p>{latestScan
                  ? `Showing ${latestScan.report.files_analyzed} files from the latest authenticated scan.`
                  : 'Scan a repository first to load its report into the graph.'}</p>
              </div>
              <button className="cyber-button-sm" onClick={() => setActiveView('evolve')}>
                {latestScan ? 'CHANGE SCAN' : 'SCAN REPOSITORY'} <ArrowRight size={14} />
              </button>
            </div>
            {latestScan
              ? <GraphVisualizer report={latestScan.report} />
              : <div className="graph-empty"><Network size={36} /><span>NO SCAN LOADED</span></div>}
          </div>
        )}

      </div>
    </div>
  );
};

export default AdminPortal;
