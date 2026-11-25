# Ground truth
Confirmed natural defects: 0. No independent failing natural-defect reproducer executed.

Initial mlsc-timetable findings: seven false positives, caused by treating gin.Context.Query string results as sql.Rows/error results. Corrected.
Initial uuid findings: five hex.Encode false positives (returns int), plus two hash.Hash.Write false positives (standard-library contract never returns an error). Corrected.
chi middleware/logger.go:52 uses LogEntry.Write with no return value; generic Write signature inference was incorrect. Arity-based correction added. Timer finding in middleware/throttle.go:105 ignores timer firing and explicit Stop paths; downgraded to unsupported/manual review, not eligible Close repair. Remaining chi diagnostics are NEEDS_MANUAL_REVIEW, not confirmed defects. Source-pattern presence alone does not establish defective behavior.

Four controlled mutations are separate from natural findings. They add a labeled isolated package to copies of the pinned user repository; they are not natural production defects. Reproducer source is supplied, but failing/passing behavior is NOT_RUN because Docker is unavailable.
