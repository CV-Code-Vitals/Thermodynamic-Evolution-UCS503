# FAIL — PIPELINE NOT TRUSTWORTHY
The mandatory release gate is not met. Real source scanning is demonstrated and several false-success paths were corrected, but no natural or injected defect has been dynamically verified as repaired. Docker is unavailable; the public executor does not wire independent reproducer/invariance commands. Browser coverage is partial. Remaining semantic limitations prevent treating diagnostics as proven defects.

Starting HEAD: 4b992caf951aef397ec9ef043bb9f1e3755b0cae. Corrections are local working-tree changes; no local commit yet. No push, merge, deployment, cloud-resource creation, or original repository modification occurred.

Confirmed natural defects: 0. Controlled mutations: 4. Generated controlled patches: 2 (resource cleanup, SQL binding). Controlled abstentions: 2 (dropped error, command injection). Independently verified successful repairs: 0. Dynamic repair failures: UNVERIFIED, not zero failures inferred from absent execution.

Supported demonstrated subset: source-only Go semantic scanning, snapshot-bound evidence, conservative abstention, and deterministic proposal generation for simple single-function cleanup and SQL parameterization. Correctness/generalization remains unproven.

See 02 for pinned repository inventory, 04 for measured scans/digests, 06 for exact proposals, and evidence logs for test commands/outcomes. Frontend tests: 56 passed. Frontend production build passed. Semantic suite and targeted security/HTTP tests passed in recorded runs; latest checks are separately logged. Historical benchmark-writing/host-execution tests were not run, preserving the user ledger and sandbox constraint. Live Docker test skipped as unavailable.

## Measured final source-only matrix
Grounded means the referenced source exists, not that the diagnostic is a confirmed defect. Files scanned here counts non-test Go files.

|Repository|Commit SHA|Files Scanned|Grounded Findings|Confirmed Defects|Patches Attempted|Dynamically Verified|Abstentions / manual-review targets|False Positives remaining|Original Immutable|
|---|---|---|---|---|---|---|---|---|---|
|Zylus08/mlsc-timetable|cbd082eafd24b290391145cb0399e81c6b41303c|4|0|0|0|0|0|UNVERIFIED|True|
|go-chi/chi|c64ed7d6635f33a8b2feaff7dcfa55677023ac7d|55|52|0|0|0|52|UNVERIFIED|True|
|google/uuid|2d3c2a9cc518326daf99a383f07c4d3c44317e4d|16|0|0|0|0|0|UNVERIFIED|True|
|Zylus08/Finger-Counter|fefd57c524fae161701f69c64b7175c8dbc45157|0|0|0|0|0|0|UNVERIFIED|True|
