# Security Exception Proposal: Transient Build-Tooling Dependency Vulnerability

**Proposal ID:** `SEC-EXP-2026-001`  
**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Target Package:** `admin-portal` (`picomatch@4.0.7`, `picomatch@2.3.2`)  
**Advisory ID:** `GHSA-fc5q-39xc-vjpf` / `CVE-2026-112366`  
**Severity:** High (ReDoS)  
**Status:** **PENDING REVIEW & APPROVAL**  
**Submission Date:** 2026-11-18  
**Expiry / Re-evaluation Date:** 2026-12-18 (30-day window)  
**Proposed By:** Antigravity Autonomous Systems Engineering Team  

---

## 1. Exception Request Overview

This security exception proposal requests temporary, time-bounded authorization to accept the transient presence of `picomatch@4.0.7` and `picomatch@2.3.2` strictly within `devDependencies` of the `admin-portal` workspace.

A breaking dependency downgrade (`npm audit fix --force` to Vite 6.2.7) or an unvetted third-party fork is prohibited by release policy, and no official upstream patch exists in the npm registry.

---

## 2. Justification & Component Necessity

### Why the Component is Required
1. **`vite@8.3.5` (Direct DevDependency):**
   Vite is the core ESM build tool and development server for `admin-portal`. It provides JSX compilation, modern TypeScript/JavaScript bundling, CSS extraction, and static asset generation. Downgrading to Vite 6.2.7 breaks React 19 compatibility (`@vitejs/plugin-react@6.0.5`), causes test failures in ES module mocking, and reintroduces older known vulnerabilities.

2. **`gh-pages@6.3.0` (Direct DevDependency):**
   `gh-pages` is used exclusively for publishing release builds of the static portal to GitHub Pages via Git worktree commits. It relies on `globby` / `fast-glob` / `micromatch` for selecting build artifacts under `admin-portal/dist`.

---

## 3. Vulnerability Reachability & Threat Surface Analysis

### Why Untrusted Inputs Cannot Reach the Vulnerable Code
- **No Untrusted Input Ingestion:**
  The advisory `GHSA-fc5q-39xc-vjpf` describes a Regular Expression Denial of Service that occurs when `picomatch` parses or matches malicious, untrusted glob strings containing repeated nested extglobs (e.g. `@(a|a)*`).
- **Static Configuration Globs Only:**
  All globs processed by Vite, `tinyglobby`, `fdir`, and `gh-pages` are static, hardcoded constants within repository configuration files (`vite.config.js`, `package.json`).
- **Complete Runtime Isolation:**
  `picomatch` is not bundled into the production JavaScript output (`dist/assets/index-*.js`). It is absent from client runtime, web requests, and server responses.
- **Verification Proof:**
  Execution of `npm audit --omit=dev --audit-level=high` confirms **0 vulnerabilities** exist in production runtime dependencies.

---

## 4. Compensating Controls in Place

The following multi-layered compensating controls are active and verified:

1. **Build Environment Sandboxing:**
   Frontend builds run in isolated GitHub Actions container runners (`ubuntu-latest`) or developer workstations with strictly bounded execution lifetimes.
2. **Deterministic Step Timeouts:**
   All CI workflow steps in `.github/workflows/check-frontend.yml` and Go/Rust CI are governed by execution timeouts (15 minutes). Any hypothetical catastrophic backtracking loop would be terminated without impacting release integrity.
3. **Artifact Integrity Verification:**
   All generated static assets (`dist/`) are validated by automated node tests (`tests/*.test.js`) before deployment.
4. **Clean Production Tree:**
   `package.json` production dependencies (`dependencies`) are restricted to `react`, `react-dom`, `lucide-react`, and `react-dropzone`. None depend on `picomatch`.

---

## 5. Upgrade Path & Remediation Roadmap

1. **Upstream Monitoring:**
   The engineering team will monitor releases of `picomatch`, `micromatch`, and `vite` for patch releases resolving `GHSA-fc5q-39xc-vjpf`.
2. **Immediate Lockfile Regeneration:**
   As soon as a verified upstream patch is published to `https://registry.npmjs.org`, `npm update picomatch` or corresponding dependency upgrades will be executed and locked via standard `npm install`.
3. **Sunset / Expiry Date:**
   This exception expires exactly 30 days from issuance: **2026-12-18**. If an upstream patch has not been released by this date, a formal review will be conducted to evaluate alternative bundler configurations.

---

## 6. Formal Approval Status

- **Current Status:** **PENDING FORMAL APPROVAL**  
  *(This proposal is documented for repository maintainers and release leads. No unauthorized silent suppression of npm audit flags has been introduced).*
