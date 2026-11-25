# CI Stabilization Dependency Security Audit Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Audit Target:** Frontend Web Application (`admin-portal`)  
**Audit Date:** 2026-11-23  
**Auditor:** Antigravity Autonomous Security Engineer  

---

## 1. Executive Summary

As part of the Milestone 3.8 MVP Release Candidate CI Stabilization Hotfix, an exhaustive security audit of the frontend dependency graph was performed. The audit focused on npm dependencies specified in [admin-portal/package.json](file:///E:/SWE/admin-portal/package.json), locked in [admin-portal/package-lock.json](file:///E:/SWE/admin-portal/package-lock.json), and tested under GitHub Actions workflow [.github/workflows/check-frontend.yml](file:///E:/SWE/.github/workflows/check-frontend.yml).

The primary security finding identified in CI runs is advisory **`GHSA-fc5q-39xc-vjpf` / `CVE-2026-112366`** affecting the glob-matching library `picomatch`.

### Summary of Audit Results

| Metric / Check | Value / Status | Verification Command |
| :--- | :--- | :--- |
| **Production Dependencies (`--omit=dev`)** | **0 Vulnerabilities** | `npm audit --omit=dev --audit-level=high` |
| **Vulnerable Package 1** | `picomatch@4.0.7` | `npm ls picomatch` (via Vite / tinyglobby / fdir) |
| **Vulnerable Package 2** | `picomatch@2.3.2` | `npm ls picomatch` (via gh-pages / globby / micromatch) |
| **Upstream Fix Availability** | **None** (latest release is 4.0.7) | `npm view picomatch versions` |
| **Breaking `npm audit fix --force` Proposal** | Proposes downgrade to `vite@6.2.7` | Rejected per security guidelines |
| **Runtime Exposure** | **Zero (0%)** | Bundled static client output does not include globs |
| **Untrusted Input Exposure** | **Zero (0%)** | CI/Build globs are static repo paths |
| **Frontend Test Suite** | **49/49 Passed** (0 failed) | `npm test` |
| **Frontend Linter (`oxlint`)** | **0 errors, 0 warnings** | `npm run lint` |
| **Production Build** | **Success** (705 ms) | `VITE_API_BASE=... npm run build` |

---

## 2. Dependency Graph Tracing

Execution of `npm ls picomatch micromatch fdir tinyglobby vite gh-pages` produces the following exact dependency hierarchy:

```text
admin-portal@0.0.0 E:\SWE\admin-portal
+-- @vitejs/plugin-react@6.0.5
| `-- vite@8.3.5 deduped
+-- gh-pages@6.3.0
| `-- globby@11.1.0
|   `-- fast-glob@3.3.3
|     `-- micromatch@4.0.8
|       `-- picomatch@2.3.2
`-- vite@8.3.5
  +-- picomatch@4.0.7
  `-- tinyglobby@0.2.17
    +-- fdir@6.5.0
    | `-- picomatch@4.0.7 deduped
    `-- picomatch@4.0.7 deduped
```

### Dependency Chain Analysis

1. **Chain A (Vite Bundler & Dev Server):**
   - Direct devDependency: `vite@^8.2.0` (resolved to `8.3.5`).
   - `vite@8.3.5` directly depends on `picomatch@4.0.7`.
   - `vite@8.3.5` depends on `tinyglobby@0.2.17`.
   - `tinyglobby@0.2.17` depends on `fdir@6.5.0` and `picomatch@4.0.7`.
   - `fdir@6.5.0` depends on `picomatch@4.0.7`.

2. **Chain B (GitHub Pages Deployment Tooling):**
   - Direct devDependency: `gh-pages@^6.3.0` (resolved to `6.3.0`).
   - `gh-pages@6.3.0` depends on `globby@11.1.0`.
   - `globby@11.1.0` depends on `fast-glob@3.3.3`.
   - `fast-glob@3.3.3` depends on `micromatch@4.0.8`.
   - `micromatch@4.0.8` depends on `picomatch@2.3.2`.

---

## 3. Advisory Vulnerability Assessment

### GHSA-fc5q-39xc-vjpf / CVE-2026-112366
- **Vulnerability Type:** Regular Expression Denial of Service (ReDoS) in `picomatch.scan()`.
- **Mechanism:** Repeated single-branch extglobs (e.g. `@(a|a)*`) induce catastrophic backtracking in pattern compilation and string matching routines within `picomatch`.
- **CVSS Base Score:** High (7.5)
- **Attack Vector:** Network / Local, requiring untrusted user-supplied glob pattern input.
- **Affected Releases:** `picomatch` all versions `<= 4.0.7` and `<= 2.3.2`.

### Upstream Registry Inspection
Querying the npm registry (`npm view picomatch versions --json`) confirms that `4.0.7` is the latest published version across the entire `picomatch` package lifetime. No patched release (e.g. `4.0.8` or `4.1.0` or `2.3.3`) has been published to npm as of the audit date.

### Rejection of Destructive Automatic Fixes
Running `npm audit fix --force` was evaluated and strictly **rejected**:
1. It requests downgrading Vite from `8.3.5` to `6.2.7`, introducing major breaking changes to ESM compilation and React 19 compatibility (`@vitejs/plugin-react@6.x`).
2. Downgrading to Vite 6.x reintroduces historical security issues and breaks React 19 JSX runtime bindings.
3. Installing unreviewed third-party forks violates repository integrity constraints.

---

## 4. Threat Modeling & Attack Surface Verification

An end-to-end evaluation was conducted to determine whether untrusted input can reach the vulnerable routines:

1. **Production Runtime Isolation:**
   `picomatch` is exclusively a build-time dependency. Vite compiles React JSX and static CSS/JS chunks into `dist/assets/*.js`. The compiled bundle contains zero glob engines and zero references to `picomatch`.
   Verification: `npm audit --omit=dev --audit-level=high` reports `found 0 vulnerabilities`.

2. **Build and CI Ingress:**
   In CI (`.github/workflows/check-frontend.yml`) and local builds (`npm run build`), glob patterns passed to Vite and tinyglobby are static string literals hardcoded in `vite.config.js` and `package.json` (e.g. `admin-portal/**`, `src/**/*.jsx`).

3. **Untrusted Input Ingress:**
   No untrusted or external user input is parsed as a glob pattern at any stage during CI, build, deployment, or browser runtime.

---

## 5. Security Decision & Recommended Policy

Because no verified upstream fix exists in the npm registry, and because the component is isolated strictly to development and build tooling with zero runtime reachability, a time-limited **Security Exception Proposal** has been formulated in [reports/ci_stabilization_security_exception_proposal.md](file:///E:/SWE/reports/ci_stabilization_security_exception_proposal.md).
