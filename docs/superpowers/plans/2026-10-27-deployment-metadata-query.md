# Deployment Metadata Query Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix the deployment workflow's JMESPath lookup of the hyphenated S3 metadata key.

**Architecture:** Keep the existing S3 metadata name and quote it in the AWS CLI query. Update the workflow regression test to require the quoted identifier.

**Tech Stack:** GitHub Actions YAML, Node.js built-in test runner, actionlint.

**Spec:** [2026-10-27-fix-deployment-metadata-query-design.md](../specs/2026-10-27-fix-deployment-metadata-query-design.md)

## Global Constraints

- Preserve the existing `bundle-sha256` S3 metadata key.
- Do not modify AWS resources or trigger GitHub Actions.
- Keep changes local for the user to review and push.

## Review Focus

- A hyphenated JMESPath identifier must be quoted so it is read as a key rather than parsed as an operator; assert the exact quoted query in the regression test.

---

### Task 1: Quote the deployment bundle checksum metadata key

**Files:**
- Modify: `.github/workflows/deploy-analysis.yml` in the `Upload private deployment bundle` step.
- Test: `tests/analysis-workflow.test.mjs` in the idempotent upload test.

**Interfaces:**
- Consumes: Existing S3 user metadata field `bundle-sha256`.
- Produces: AWS CLI query `Metadata."bundle-sha256"` and a regression assertion for that expression.

- [x] **Step 1: Update the regression assertion**

Change the assertion in `makes deployment bundle upload idempotent and refuses a conflicting SHA release` to require `Metadata."bundle-sha256"`.

- [x] **Step 2: Run the targeted test and verify it fails**

Run: `node --test tests/analysis-workflow.test.mjs`
Expected: The upload idempotency test fails because the workflow still contains the unquoted query.

- [x] **Step 3: Quote the key in the workflow**

Change the `--query` expression from `Metadata.bundle-sha256` to `Metadata."bundle-sha256"`; leave the upload metadata and all other behavior unchanged.

- [x] **Step 4: Run targeted tests and workflow syntax validation**

Run: `node --test tests/analysis-workflow.test.mjs`
Expected: All tests pass.

Run: `go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/deploy-analysis.yml`
Expected: No workflow syntax errors.
