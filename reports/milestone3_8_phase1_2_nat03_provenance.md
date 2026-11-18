# Milestone 3.8 Phase 1.2 — Forensic Investigation of NAT-03 Provenance

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Branch:** `release/milestone-3.8-mvp-rc`  
**Date:** 2026-11-08  
**Investigator:** Antigravity Autonomous Systems Audit Agent  

---

## 1. Executive Summary

A forensic investigation was conducted into the provenance, authenticity, and technical validity of evaluation target **NAT-03** (`gorilla_websocket/client.go:375`), previously classified in Milestone 3.7 Phase 1 and Phase 2 as a "historical natural defect" repaired autonomously.

### Core Forensic Findings:
1. **Pristine Upstream Origin:** The benchmark file `benchmarks/external_v2/gorilla_websocket/client.go` is an exact, byte-for-byte, unmutated copy of `gorilla/websocket` commit `9111bb834a68b893cebbbaed5060bdbc1d9ab7d2` (official release `v1.5.0`, tagged April 2022).
2. **Fictitious Fix Reference:** Upstream GitHub item `gorilla/websocket#789` is a user discussion issue titled *"Investigating Error code 1006"*, created in 2021 regarding WebSocket close codes. It is **not** a resource-leak repair pull request, contains no code fixes, and has no associated leak patch.
3. **No Unclosed Socket Defect:** The underlying network connection `netConn` is acquired at line 320 and immediately protected by a deferred cleanup handler at lines 330–334 (`defer func() { if netConn != nil { netConn.Close() } }()`). On all error paths, `netConn` is reliably closed. On successful handshake, ownership transfers cleanly to `*Conn` at line 413 (`netConn = nil`).
4. **Explicit Contract Exemption:** The upstream package documentation for `DialContext` (lines 157–158) explicitly specifies:  
   > *"The response body may not contain the entire response and does not need to be closed by the application."*
5. **Prior Disqualification in Benchmark Metadata:** The benchmark repository's own ground-truth amendment file (`benchmarks/external_v2/ground_truth_amendment_v2_1.json`) explicitly disqualified this target prior to Milestone 3.8:  
   > *"Claimed line 82 is a struct field. Socket at line 320 is protected by deferred Close handler at line 330. Not a defect."*

**Final Provenance Verdict:** **`GROUND_TRUTH_DISQUALIFIED / BENIGN_IDIOM`**. NAT-03 is not a defect. Zero natural defects exist in the verified repair denominator.

---

## 2. Upstream Pinned Commit Verification

The evaluation benchmark pins `gorilla/websocket` at upstream commit:
- **Commit SHA:** `9111bb834a68b893cebbbaed5060bdbc1d9ab7d2`
- **Tag:** `v1.5.0`
- **File:** `client.go`
- **File Size:** 42,977 bytes
- **SHA-256 Digest:** `7a885bebb08e1a129ef993a4c498308bf33291583d7c588a44ecad7e7a1772bf`

A cryptographic comparison against the upstream tree confirmed:
- Zero lines were modified, inserted, or deleted in `benchmarks/external_v2/gorilla_websocket/client.go`.
- The benchmark file was imported directly from the pristine tagged repository.

---

## 3. Analysis of GitHub Reference `gorilla/websocket#789`

Previous milestone documentation attributed NAT-03 to `gorilla/websocket#789` as a "historical bug fix". Forensic audit of GitHub records reveals:
- **URL:** `https://github.com/gorilla/websocket/issues/789`
- **Title:** "Investigating Error code 1006"
- **Author:** External user reporting connection resets during high-throughput messaging.
- **Type:** Community discussion issue.
- **Pull Requests / Commits:** None linked to issue #789 that modify `DialContext` or close response bodies.
- **Conclusion:** The citation of #789 as a bug-fix PR was a misattribution introduced during initial corpus indexing in Milestone 3.7 Phase 1.

---

## 4. Resource Lifecycle & Control-Flow Analysis

The alleged defect claimed that `client.go:375` (inside `DialContext`) leaks an `*http.Response` body allocated by `http.ReadResponse`.

```go
// client.go lines 369-378
resp, err := http.ReadResponse(conn.br, req)
if err != nil {
    if netConn != nil {
        netConn.Close()
    }
    return nil, nil, err
}
```

### Complete Socket & Body Lifecycle:
1. **Socket Allocation (Line 320):**
   ```go
   netConn, err := netDialer.DialContext(ctx, "tcp", net.JoinHostPort(u.Host, port))
   ```
2. **Immediate Deferred Protection (Lines 330–334):**
   ```go
   defer func() {
       if netConn != nil {
           netConn.Close()
       }
   }()
   ```
3. **HTTP Handshake Failure Handling (Lines 387–396):**
   ```go
   if resp.StatusCode != 101 {
       n, _ := io.ReadFull(resp.Body, buf[:cap(buf)])
       resp.Body = ioutil.NopCloser(bytes.NewReader(buf[:n]))
       return nil, resp, ErrBadHandshake
   }
   ```
   *Insight:* The response body is read into an in-memory buffer, and `resp.Body` is replaced with `ioutil.NopCloser(bytes.NewReader(...))`. Calling `Close()` on a `NopCloser` is a no-op that releases no OS resources. The actual OS socket `netConn` is closed by the deferred handler when `DialContext` returns `ErrBadHandshake`.
4. **Successful Handshake (Lines 407–414):**
   ```go
   resp.Body = ioutil.NopCloser(bytes.NewReader(nil))
   conn.subprotocol = resp.Header.Get("Sec-Websocket-Protocol")
   netConn = nil // to avoid close in defer
   return conn, resp, nil
   ```
   *Insight:* On success, `netConn` is set to `nil` so the defer does not close the active socket, and ownership of `netConn` transfers to the newly constructed `*Conn` object. `resp.Body` is an empty in-memory reader.

### Resulting Failure Mode of Previous Synthetic Patch:
In Milestone 3.8 Phase 1.1, the autonomous synthesizer inserted:
```go
resp, err := http.ReadResponse(conn.br, req)
defer resp.Body.Close()
if err != nil { ... }
```
When `http.ReadResponse` encountered an invalid handshake, it returned `(nil, err)`. The deferred call evaluated `nil.Body.Close()`, triggering an immediate fatal panic:
`runtime error: invalid memory address or nil pointer dereference`.

---

## 5. Provenance Timeline & Evolution

| Milestone / Date | Milestone Classification | Finding / Metric Reported | Audit Verdict |
| :--- | :--- | :--- | :--- |
| **Milestone 3.7 Phase 1** (Oct 2026) | Corpus Construction | Ingested `gorilla_websocket` as real-world Go repo; tagged `NAT-03` with reference `#789`. | Fictitious PR citation; candidate was pristine upstream code. |
| **Milestone 3.7 Phase 2** (Oct 2026) | Pilot External Repair | Included in 10-case repair set; reported as verified natural repair. | **False Positive Verification:** Oracles lacked negative-path nil-pointer tests. |
| **Milestone 3.8 Phase 0** (Nov 2026) | Scientific Freeze | Questioned validity of NAT-03; mandated forensic audit of #789. | Investigation initiated. |
| **Milestone 3.8 Phase 1.1** (Nov 2026) | Defect Reproduction | Built deterministic reproducer (`nat03_reproducer_test.go`); confirmed patch panics on nil response. | AST rule corrected to require post-error check placement; synthesizer abstains on escaping return. |
| **Milestone 3.8 Phase 1.2** (Nov 2026) | Provenance & Claims Audit | Byte-for-byte upstream comparison; #789 confirmed unrelated; contract verified. | **Formally Disqualified:** Stratum updated to `GROUND_TRUTH_DISQUALIFIED`. Natural repair count corrected to 0. |

---

## 6. Corrective Actions Enforced

1. **Ground Truth Reclassification:** `NAT-03` is permanently recorded as `GROUND_TRUTH_DISQUALIFIED` in `reports/milestone3_8_phase1_2_corrected_ground_truth.json`.
2. **Metrics Recalculation:** All claims of natural defect repair have been retracted. The numerator of verified natural repairs is **0**.
3. **Demo & CLI Alignment:** The demo application (`scripts/run-demo.ps1`, `remediation_mvp.exe`) now uses **`INJ-EXT2-RES-02`** (authentic injected defect in `fatih/color`) for the positive repair demo, and presents `NAT-03` as an audited disqualified candidate where the engine safely **abstains**.
4. **UI Dashboard Strata Badges:** The React dashboard (`RemediationPortal.jsx`) displays explicit provenance stratum pills: `AUTHENTIC_INJECTED`, `CONSTRUCTED_MICROSERVICE`, `CLEAN_CONTROL`, and `GROUND_TRUTH_DISQUALIFIED`.

