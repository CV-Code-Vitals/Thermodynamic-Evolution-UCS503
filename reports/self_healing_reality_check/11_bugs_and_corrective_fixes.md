# Corrective engineering
- Removed implicit demo append and exposed discovery failures; separate explicit demo routes preserve fixtures.
- Removed live use of untrusted hotspot reports and blanket allocation fallback.
- Reject parse failures, missing snapshots, expired snapshots, and stale/deleted source during analysis.
- Preserve source URL/SHA, rule, symbol, line, evidence, confidence and snapshot hashes; deterministic finding ordering.
- Propagate analyzer failures instead of silently swallowing them.
- Correct same-line statement ordering and demonstrated Gin Query, hex.Encode, hash.Write, and incompatible Write-arity signature errors.
- Respect scanner abstention in proposal generation; bind synthesis to the selected function.
- Missing reproducer is NOT_RUN; missing compilation environment is UNAVAILABLE rather than PASSED.
- Approval consumption fails if durable ledger persistence fails; truncated/unreadable ledgers fail closed.
- Clear UI selection and pipeline state on workspace changes; remove hardcoded baseline display and default demo selection.
- Replace fixed live test IDs with discovered candidates; update tests that previously required demo contamination or manufactured missing-source analysis.

Remaining: general receiver typing is incomplete; timer/control-flow findings require review; public independent reproducer/regression configuration is absent; private/not-found error distinctions and explicit unsupported-language status are incomplete. No arbitrary-defect repair capability is established.
