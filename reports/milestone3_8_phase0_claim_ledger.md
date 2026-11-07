# Milestone 3.8 — Phase 0: Scientific Release Claim Ledger

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `de06b6e568e0cce07d20b319d6618c15a9d3c899`  
**Document Type:** Formal Scientific Claim Ledger & Anti-Hype Guardrails  
**Status:** FROZEN & RATIFIED FOR RELEASE CANDIDATE  

---

## 1. Epistemological Framework

To maintain scientific integrity and prevent unjustified extrapolation, this Release Claim Ledger classifies all system capabilities into three explicit tiers:
1. **Demonstrated Empirical Claims:** Capabilities mathematically and experimentally proven across reproducible test suites.
2. **Bounded Operational Claims:** Capabilities operating under documented technical limitations and constraints.
3. **Unsupported / Prohibited Claims:** Out-of-scope behaviors that the system strictly disclaims and refuses to attempt.

---

## 2. Demonstrated Empirical Claims

| Claim ID | Capability Claim | Empirical Evidence | Verified Ground Truth |
| :---: | :--- | :--- | :--- |
| **DEM-01** | **Deterministic AST Resource Leak Repair** | Correctly synthesizes `defer <handle>.Close()` for `os.File`, `net.Conn`, `http.Response.Body`, and stream reassignments. | Verified across `NAT-03`, `INJ-EXT2-RES-02`, `INJ-EXT2-RES-03`, `DEF-HOLD-03`, `DEF-HOLD-05`, `DEF-HOLD-09`. |
| **DEM-02** | **Deterministic AST SQL Injection Parameterization** | Replaces formatted SQL strings (`fmt.Sprintf`, binary string concatenation) with parameterized `?` placeholders, binds arguments to call signatures, and cleans unused imports. | Verified across `INJ-EXT2-SQL-01`, `INJ-EXT2-SQL-02`, `DEF-HOLD-02`, `DEF-HOLD-08`. |
| **DEM-03** | **Decoupled 5-Stage Verification Oracles** | Evaluates patches against 5 distinct oracles independent of original detector (Compilation, Regression, Reproducer, Invariance, Negative Properties). | Zero detector circularity; 10/10 attempted candidate repairs pass all 5 oracles. |
| **DEM-04** | **Cryptographic Authorization & Replay Resistance** | HMAC-SHA256 token verification binding principal, task ID, source commit, file hashes, patch diff hash, workspace, and timestamp with persistent atomic disk ledger. | 10/10 adversarial penetration vectors pass (`approval_test.go`). |
| **DEM-05** | **Fail-Closed Conservative Abstention** | Safely detects unsupported defect categories (shell injections `sh -c`, void return signatures) and yields explicit `ABSTAIN` with zero crashes or code mutations. | Verified across `INJ-EXT2-CMD-01`, `DEF-HOLD-04`, `INJ-EXT2-ERR-04`, `DEF-HOLD-01`, `DEF-HOLD-06`, `DEF-HOLD-07`. |
| **DEM-06** | **Bitwise Source Repository Immutability** | All patch synthesis and trial execution occur in ephemeral disposable sandboxes; original benchmark repositories remain 100% byte-for-byte identical. | Pre- and post-repair SHA-256 tree digests verified bitwise identical. |

---

## 3. Bounded Operational Claims

| Bound ID | Operational Limitation | Technical Root Cause | System Guardrail |
| :---: | :--- | :--- | :--- |
| **BND-01** | **Intra-Procedural AST Mutation Scope** | AST synthesizer mutates single-file syntax trees; does not refactor multi-file public interfaces. | Tasks requiring cross-file signature modifications trigger explicit `AMBIGUOUS_RETURN_SIGNATURE` abstentions. |
| **BND-02** | **Single Natural Defect Sample Size ($N=1$)** | While `NAT-03` is an authentic upstream defect in `gorilla/websocket`, $N=1$ yields a wide confidence interval ([20.7%, 100%]). | System claims demonstrated capability on this class of natural bug, but disclaims generalization across all arbitrary Go bugs. |
| **BND-03** | **Bounded SQL Placeholder Dialects** | Parameterization engine targets standard `?` syntax. | Dialects utilizing numbered positional parameters (`$1`, `$2` in PostgreSQL) require future extension. |
| **BND-04** | **Local Ephemeral Sandbox Dependency** | Verification requires local execution of `go test` and syntax analysis tools in isolated directories. | Requires standard Go toolchain installed on host; unsupported platforms must be explicitly documented. |

---

## 4. Unsupported / Prohibited Claims (Anti-Hype Guardrails)

The following claims are **STRICTLY DISCLAIMED**:

1. **"Fully Autonomous Production Sweeps":** The system is NOT authorized or designed to autonomously push commits, merge pull requests, or alter production code without authenticated human approval.
2. **"General Artificial Program Repair":** The system does NOT synthesize arbitrary algorithmic code, invent new business logic, or solve complex logic bugs outside its formal deterministic AST mutators.
3. **"Automated Command Injection Resolution":** Arbitrary shell strings passed to interpreters (`sh -c`) cannot be safely converted to structured subprocess calls without architectural redesign. The system intentionally and safely fails closed.
4. **"Zero-Human Oversight Safety":** The cryptographic token and approval gate explicitly enforce human accountability. Automated eligibility does not equal execution permission.
