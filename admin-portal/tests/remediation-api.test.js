import { describe, it } from 'node:test';
import assert from 'node:assert';
import {
  fetchRemediationTargets,
  analyzeRemediationTarget,
  proposeRemediationRepair,
  approveRemediationRepair,
  executeRemediationRepair,
  rollbackRemediationSandbox,
  fetchRemediationAuditReport,
  runRemediationDemo
} from '../src/remediation-api.js';

describe('Autonomous Remediation API Client', () => {
  it('fetchRemediationTargets calls /remediation/targets and returns targets array', async () => {
    const mockFetch = async (url) => {
      assert.strictEqual(url, 'http://localhost:8080/api/remediation/targets');
      return {
        ok: true,
        json: async () => ({ targets: [{ target_id: 'NAT-03', defect_category: 'RESOURCE_LEAK' }] }),
      };
    };
    globalThis.fetch = mockFetch;
    const res = await fetchRemediationTargets('http://localhost:8080/api');
    assert.strictEqual(res.length, 1);
    assert.strictEqual(res[0].target_id, 'NAT-03');
  });

  it('analyzeRemediationTarget sends target_id in POST body', async () => {
    const mockFetch = async (url, opts) => {
      assert.strictEqual(url, 'http://localhost:8080/api/remediation/analyze');
      assert.strictEqual(opts.method, 'POST');
      const body = JSON.parse(opts.body);
      assert.strictEqual(body.target_id, 'NAT-03');
      return {
        ok: true,
        json: async () => ({
          analysis_id: 'ANALYSIS-NAT-03-123',
          automatic_eligibility: true,
          uncertainty_score: 0.05,
        }),
      };
    };
    globalThis.fetch = mockFetch;
    const res = await analyzeRemediationTarget('http://localhost:8080/api', '', 'NAT-03');
    assert.strictEqual(res.analysis_id, 'ANALYSIS-NAT-03-123');
    assert.strictEqual(res.automatic_eligibility, true);
  });

  it('proposeRemediationRepair calls /remediation/propose with analysis_id', async () => {
    const mockFetch = async (url, opts) => {
      assert.strictEqual(url, 'http://localhost:8080/api/remediation/propose');
      const body = JSON.parse(opts.body);
      assert.strictEqual(body.analysis_id, 'ANALYSIS-NAT-03-123');
      return {
        ok: true,
        json: async () => ({
          proposal_id: 'PROP-NAT-03-123',
          status: 'PROPOSED',
          unified_diff: '--- a/client.go\n+++ b/client.go\n',
        }),
      };
    };
    globalThis.fetch = mockFetch;
    const res = await proposeRemediationRepair('http://localhost:8080/api', '', 'ANALYSIS-NAT-03-123');
    assert.strictEqual(res.status, 'PROPOSED');
  });

  it('approveRemediationRepair sends proposal_id, principal, and role', async () => {
    const mockFetch = async (url, opts) => {
      assert.strictEqual(url, 'http://localhost:8080/api/remediation/approve');
      const body = JSON.parse(opts.body);
      assert.strictEqual(body.proposal_id, 'PROP-NAT-03-123');
      assert.strictEqual(body.principal, 'operator@internal');
      assert.strictEqual(body.role, 'Lead SRE');
      return {
        ok: true,
        json: async () => ({
          token_id: 'tok-abc-123',
          valid: true,
        }),
      };
    };
    globalThis.fetch = mockFetch;
    const res = await approveRemediationRepair('http://localhost:8080/api', '', 'PROP-NAT-03-123', 'operator@internal', 'Lead SRE');
    assert.strictEqual(res.token_id, 'tok-abc-123');
    assert.strictEqual(res.valid, true);
  });

  it('executeRemediationRepair sends token_id and returns oracle results', async () => {
    const mockFetch = async (url, opts) => {
      assert.strictEqual(url, 'http://localhost:8080/api/remediation/execute');
      const body = JSON.parse(opts.body);
      assert.strictEqual(body.token_id, 'tok-abc-123');
      return {
        ok: true,
        json: async () => ({
          execution_id: 'EXEC-NAT-03-123',
          all_oracles_passed: true,
          verification_verdict: 'INDEPENDENTLY_VERIFIED',
          original_repo_immutable: true,
        }),
      };
    };
    globalThis.fetch = mockFetch;
    const res = await executeRemediationRepair('http://localhost:8080/api', '', 'tok-abc-123');
    assert.strictEqual(res.all_oracles_passed, true);
    assert.strictEqual(res.original_repo_immutable, true);
  });

  it('rollbackRemediationSandbox calls /remediation/rollback', async () => {
    const mockFetch = async (url, opts) => {
      assert.strictEqual(url, 'http://localhost:8080/api/remediation/rollback');
      const body = JSON.parse(opts.body);
      assert.strictEqual(body.execution_id, 'EXEC-NAT-03-123');
      return {
        ok: true,
        json: async () => ({ success: true }),
      };
    };
    globalThis.fetch = mockFetch;
    const res = await rollbackRemediationSandbox('http://localhost:8080/api', '', 'EXEC-NAT-03-123');
    assert.strictEqual(res.success, true);
  });

  it('fetchRemediationAuditReport returns machine-readable audit report', async () => {
    const mockFetch = async (url) => {
      assert.strictEqual(url, 'http://localhost:8080/api/remediation/audit');
      return {
        ok: true,
        json: async () => ({
          baseline_commit: 'de06b6e',
          total_sessions_recorded: 2,
          source_immutability_100_percent: true,
        }),
      };
    };
    globalThis.fetch = mockFetch;
    const res = await fetchRemediationAuditReport('http://localhost:8080/api');
    assert.strictEqual(res.baseline_commit, 'de06b6e');
    assert.strictEqual(res.source_immutability_100_percent, true);
  });
});
