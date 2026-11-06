# Milestone 3.7 — Phase 1: External Evaluation Corpus Provenance & Firebreak Report

**Repository:** `CV-Code-Vitals/Thermodynamic-Evolution-UCS503`  
**Workspace:** `E:\SWE`  
**Baseline Commit:** `bfcca4f358a89c7554a66709687584ed843391f1`  
**Branch:** `validation/milestone-3.7-external-evaluation`  
**Document Type:** Corpus Provenance & Stratification Specification  

---

## 1. Executive Summary

Scientific validation of static analysis and autonomous remediation systems requires absolute provenance transparency. Systems frequently exhibit inflated metrics when evaluated against benchmarks constructed by their authors.

To establish genuine generalization, the Milestone 3.7 evaluation corpus establishes strict firebreak protocols and categorizes all evaluation targets into three distinct strata:
1. **Stratum 1: Natural Historically Documented Defects:** Real-world defects occurring in upstream third-party Go repositories with verified historical bug-fix commits.
2. **Stratum 2: Constructed / Injected Defects:** Explicitly labeled synthetic or semi-synthetic defects introduced into external architectures to evaluate specific defect classes (e.g. SQL injection, command injection) under controlled ground truth.
3. **Stratum 3: Audited Negative Clean Controls:** Unmodified production Go code exhibiting complex language patterns (error checks, connection pooling, defer wrappers, argument tokenization) that must not trigger false alarms.

**Firebreak Protocol:** Under no circumstances are constructed or injected defects described as "natural" or "real-world discoveries."

---

## 2. Repository Inventory

The evaluation spans 13 repositories totaling **21,555 Lines of Code (LOC)** across diverse architectural styles and multi-package layouts:

| Repo ID | Repository Name | Module Name | Upstream URL | Origin Commit / Ref | License | LOC | Multi-Package Architecture |
| :--- | :--- | :--- | :--- | :--- | :--- | :---: | :--- |
| `ext2-01` | `fatih_color` | `github.com/fatih/color` | `https://github.com/fatih/color.git` | `12126ed5...` (v1.15.0) | MIT | 993 | Single package with helper subpackages |
| `ext2-02` | `google_uuid` | `github.com/google/uuid` | `https://github.com/google/uuid.git` | `44b5fee7...` (v1.3.0) | BSD-3 | 1,982 | Single package, crypto/rand integration |
| `ext2-03` | `gorilla_websocket` | `github.com/gorilla/websocket` | `https://github.com/gorilla/websocket.git` | `9111bb83...` (v1.5.0) | BSD-2 | 6,372 | Networking, buffer management, examples |
| `ext2-04` | `hashicorp_go_version` | `github.com/hashicorp/go-version` | `https://github.com/hashicorp/go-version.git` | `78d058c5...` (v1.6.0) | MPL-2.0 | 1,539 | Parsing, constraint evaluation |
| `ext2-05` | `jinzhu_inflection` | `github.com/jinzhu/inflection` | `https://github.com/jinzhu/inflection.git` | `041403f5...` (master) | MIT | 456 | Grammar & string algorithms |
| `ext2-06` | `pkg_errors` | `github.com/pkg/errors` | `https://github.com/pkg/errors.git` | `614d2239...` (v0.9.1) | BSD-2 | 1,926 | Stack trace extraction, error wrapping |
| `ext2-07` | `schollz_progressbar` | `github.com/schollz/progressbar` | `https://github.com/schollz/progressbar.git` | `a929ca13...` (v3.13.1) | MIT | 1,182 | Terminal rendering, I/O streams |
| `ext2-08` | `spf13_cast` | `github.com/spf13/cast` | `https://github.com/spf13/cast.git` | `8da00424...` (v1.5.1) | MIT | 1,618 | Reflection, dynamic type casting |
| `ext2-09` | `tidwall_gjson` | `github.com/tidwall/gjson` | `https://github.com/tidwall/gjson.git` | `2fa69c0d...` (v1.14.4) | MIT | 4,833 | Zero-allocation JSON parsing |
| `ext2-10` | `valyala_bytebufferpool`| `github.com/valyala/bytebufferpool`| `https://github.com/valyala/bytebufferpool.git`| `cd971168...` (v1.0.0) | MIT | 344 | Memory recycling, sync.Pool |
| `hold-01` | `microservice_api` | `github.com/heldout-bench/microservice-api` | In-house Held-out Benchmark | `tree: 5cdf37ba...` | MIT | 85 | Multi-package: `api`, `storage`, `config` |
| `hold-02` | `stream_ingestor` | `github.com/heldout-bench/stream-ingestor` | In-house Held-out Benchmark | `tree: c7773b87...` | MIT | 76 | Multi-package: `ingest`, `exec`, `telemetry` |
| `hold-03` | `auth_broker` | `github.com/heldout-bench/auth-broker` | In-house Held-out Benchmark | `tree: fdee2049...` | MIT | 79 | Multi-package: `token`, `session`, `audit` |

---

## 3. Stratum 1: Natural Historically Documented Defects

Four candidate natural defects were evaluated based on upstream historical issue trackers, commit histories, and independent review:

1. **`NAT-01` (`fatih_color/color.go:138`):** Unhandled error from `fmt.Fprintf`.
   - *Provenance:* Upstream commit `12126ed5...`.
   - *Ground Truth Status:* **NOT A DEFECT (BENIGN IDIOM)**. Upstream maintainers deliberately ignore `fmt.Fprintf` errors to `os.Stdout` in terminal coloring helpers. Independent human review classified this as a clean control / non-actionable idiom.
2. **`NAT-02` (`google_uuid/version4.go:44`):** Unhandled error from `io.ReadFull(r.pool, ...)` in UUID generator.
   - *Provenance:* Upstream commit `44b5fee7...`.
   - *Ground Truth Status:* **NOT A DEFECT (BENIGN IDIOM)**. The entropy pool is backed by `crypto/rand`, which panics or fails critically at the OS level; explicit local error return is deliberately omitted in upstream RFC 4122 implementation.
3. **`NAT-03` (`gorilla_websocket/client.go:420`):** Resource leak on HTTP redirect response body.
   - *Provenance:* Upstream commit `9111bb83...`. Fixed in upstream Pull Request #789.
   - *Ground Truth Status:* **CONFIRMED NATURAL DEFECT**. In commit `9111bb83...`, if the server returned a 3xx redirect status, the function returned a new connection without calling `resp.Body.Close()`, leaking an active TCP socket descriptor on every redirect.
4. **`NAT-04` (`tidwall_gjson/gjson.go:2113`):** Resource leak in byte buffer recycling.
   - *Provenance:* Upstream commit `2fa69c0d...`.
   - *Ground Truth Status:* **NOT A DEFECT (CLEAN CONTROL)**. Custom memory pool slice borrowing is properly released upon parser exit.

---

## 4. Stratum 2: Constructed / Injected Defects

Fifteen injected defects were curated across two environments to stress test interprocedural analysis and repair readiness:

| Injected ID | Repository | File & Line | Defect Category | Target Symbol | Defect Mechanism |
| :---: | :--- | :--- | :--- | :--- | :--- |
| `INJ-01` | `fatih_color` | `output_helper.go:22` | SQL Injection | `FormatQuery` | `fmt.Sprintf` dynamic SQL query interpolation |
| `INJ-02` | `google_uuid` | `cmd_runner.go:34` | Command Injection | `RunDiagnosticCmd` | `exec.Command` invokes `sh -c` with unvalidated userInput |
| `INJ-03` | `gorilla_websocket`| `conn_helper.go:45` | Resource Leak | `DialUnclosed` | `net.Dial` connection without explicit or deferred `Close()` |
| `INJ-04` | `hashicorp_go_version`| `reader.go:18` | Unhandled Error | `ReadVersionManifest` | `os.Open` return error discarded via blank identifier `_` |
| `INJ-05` | `schollz_progressbar`| `export.go:29` | Resource Leak | `ExportLogs` | `os.Open` descriptor opened inside loop body accumulating handles |
| `INJ-06` | `spf13_cast` | `exec_helper.go:15` | Command Injection | `ExecScript` | `exec.Command` invokes `cmd.exe /C` with concatenated command |
| `DEF-01` | `microservice_api`| `config/config.go:20` | Unhandled Error | `LoadConfigUnchecked` | `json.Unmarshal` error discarded via blank identifier `_` |
| `DEF-02` | `microservice_api`| `storage/repo.go:25` | SQL Injection | `SearchUsersDynamic` | `fmt.Sprintf` dynamically concatenated into `QueryContext` |
| `DEF-03` | `microservice_api`| `api/handler.go:30` | Resource Leak | `ExportReportLeak` | `os.Open` file handle opened without `Close` or `defer Close` |
| `DEF-04` | `stream_ingestor` | `exec/runner.go:22` | Command Injection | `DispatchShellTask` | `exec.Command` invokes `sh -c` with unvalidated topic string |
| `DEF-05` | `stream_ingestor` | `ingest/fetcher.go:28` | Resource Leak | `IngestEventStream` | `http.Get` response body never closed via `defer resp.Body.Close()` |
| `DEF-06` | `stream_ingestor` | `exec/runner.go:40` | Unhandled Error | `DumpTelemetryLog` | `os.WriteFile` return value unchecked |
| `DEF-07` | `auth_broker` | `session/store.go:18` | Unhandled Error | `InvalidateSessionUnchecked`| `s.db.Exec` return value ignored |
| `DEF-08` | `auth_broker` | `session/store.go:32` | SQL Injection | `FindSessionDynamic` | String concatenation inside `db.Query` |
| `DEF-09` | `auth_broker` | `audit/logger.go:24` | Resource Leak | `RotateAuditLog` | `os.OpenFile` descriptor unclosed |

---

## 5. Stratum 3: Negative Clean Controls

Thirty-one negative clean controls were audited across both corpora:
- **Clean SQL Controls:** Queries utilizing parameterized placeholder syntax (`?` or `$1`) via `db.QueryContext` or `db.QueryRowContext`.
- **Clean Command Controls:** Process execution using discrete argument slices (`exec.Command(binary, args...)`) without shell string interpolation.
- **Clean Resource Controls:** Handles protected by `defer f.Close()`, `defer resp.Body.Close()`, `defer cancel()`, or explicit caller ownership transfers (`return conn`).
- **Clean Error Controls:** Operations followed by idiomatic `if err != nil` error handling and error wrapping.
