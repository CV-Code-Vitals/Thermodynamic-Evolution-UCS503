# Analysis Stack Audit Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remediate the actionable analysis-stack audit findings with safer resource ownership, release rollback, and a separately approved cleanup path.

**Architecture:** CloudFormation owns uniquely named resources and exports deployment targets. The deployment workflow resolves those outputs, validates the sandbox host and bundle, and restores the previous release on failure. A read-only inventory plus explicit plan/apply cleanup utility is invoked only through a protected manual workflow.

**Tech Stack:** AWS SAM/CloudFormation YAML, GitHub Actions YAML, Go tests, Node.js 22 tests and cleanup utility, AWS CLI.

**Spec:** `docs/superpowers/specs/2026-10-25-analysis-stack-audit-remediation-design.md`

## Global Constraints

- Local code and documentation changes only; no AWS calls, deployment, cleanup execution, commit, or push.
- Ordinary CloudFormation stack deletion must not silently erase analysis data.
- Cleanup must require protected-environment approval and exact confirmation of `ucs503-analysis-v1`.
- Cleanup must require exact `AnalysisStack` tags for buckets, tables, volumes, and snapshots; logs must match known exact names. Prefix-only ownership is insufficient.
- Cleanup must never target the archive stack, AMIs, or stopped image-builder instances.
- Preserve the current IPv6 subnet allocation, DLM stack-tag condition, DLM target tags, exact SSM parameter paths, and prefix-specific S3 expiration rules.
- Do not add interface VPC endpoints or change IPv6 HTTPS egress in this change.
- Do not invent a general-purpose object expiration period.

## Review Focus

- A cleanup request for the wrong prefix or a still-active stack must refuse before deleting anything; test both cases in the cleanup utility task.
- A bucket, table, volume, or snapshot matching only the prefix but missing the exact ownership tag must be preserved; test in the cleanup utility task.
- An S3 bucket with multiple pages of versions and delete markers must be emptied without skipping markers; test paginated listing and batched deletion in the cleanup utility task.
- A data volume that is attached or snapshots without the expected stack tag must be preserved; test both ownership guards in the cleanup utility task.
- A missing, duplicated, or malformed CloudFormation output must stop deployment before upload or SSM commands; test in the deployment workflow task.
- A failed deployment on either host or failed readiness smoke check must attempt rollback on both hosts and report rollback failure; test in the deployment workflow task.

---

### Task 1: Make stack resources collision-resistant and retention explicit

**Files:**
- Modify: `infra/analysis/template.yaml`
- Modify: `infra/analysis/parameters.example.json`
- Modify: `go-backend/sandbox_lifecycle.go`
- Test: `go-backend/sandbox_lifecycle_test.go`
- Test: `tests/analysis-template.test.mjs`
- Modify: `infra/analysis/README.md`
- Modify: `README_DEPLOY.md`

**Interfaces:**
- Consumes: existing `ResourcePrefix`, current analysis API/table environment references, existing CloudFormation outputs, and existing stack tags.
- Produces: generated physical names for `SourceBucket`, `ArtifactBucket`, and `AnalysisLifecycleTable`; a validated configurable table name with the same local default; stack-unique log-group and Lambda names; `AnalysisDataVolume` retained through ordinary stack deletion; ENIs configured to delete with their owning instances.

- [ ] **Step 1: Add failing naming and lifecycle assertions**

Add assertions to the existing template test suite that verify:
- `SourceBucket` and `ArtifactBucket` have no `BucketName`.
- `AnalysisLifecycleTable` has no `TableName`, and the `SandboxLifecycleTableName` parameter is absent.
- Existing output values still reference the generated bucket/table resources.
- Stack-unique suffixes distinguish explicit log-group and Lambda names.
- `AnalysisDataVolume` has both deletion and replacement policies set to `Retain`.
- Both EC2 network-interface attachments use `DeleteOnTermination: true`.
- Existing IPv6 CIDR and DLM tag-scope assertions remain present.

- [ ] **Step 2: Run the targeted test and verify it fails**

Run: `node --test tests/analysis-template.test.mjs`
Expected: FAIL on the new fixed-name and lifecycle assertions.

- [ ] **Step 3: Update CloudFormation names and retention**

Remove the fixed bucket names and lifecycle table name parameter/property. Add a stack-unique suffix to explicitly named Lambda functions and log groups without exceeding Lambda's 64-character function-name limit. Retain the data volume during normal stack deletion and set both EC2 attachment deletion flags to `true`. Preserve all internal `Ref`/`GetAtt` relationships, runtime table environment variables, tags, IPv6 configuration, DLM target tags, and DLM IAM tag condition.

- [ ] **Step 4: Add failing Go table-name tests**

In `go-backend/sandbox_lifecycle_test.go`, add constructor tests for valid generated-style names and the local `sandbox-lifecycle` name, plus invalid length and invalid-character cases. Verify the lifecycle record ID remains `sandbox-lifecycle`.

Run: `Set-Location go-backend; go test ./...`
Expected: FAIL on valid generated names and/or the new invalid-name assertions.

- [ ] **Step 5: Make the Go table-name contract configurable**

In `go-backend/sandbox_lifecycle.go`, validate configured DynamoDB table names against `[A-Za-z0-9_.-]{3,255}` instead of requiring the literal `sandbox-lifecycle`. Keep `sandbox-lifecycle` as the lifecycle record ID and local Compose/environment default.

Run: `Set-Location go-backend; go test ./...`
Expected: all Go tests PASS.

- [ ] **Step 6: Remove the obsolete SAM parameter example**

Remove `SandboxLifecycleTableName` from `infra/analysis/parameters.example.json` and verify the example still parses as JSON.

- [ ] **Step 7: Document generated names and retained-data behavior**

Update deployment instructions to resolve bucket/table names from stack outputs. State that ordinary stack deletion retains data-bearing resources and that only the approved cleanup workflow deletes them.

- [ ] **Step 8: Run the targeted test and verify it passes**

Run: `node --test tests/analysis-template.test.mjs`
Expected: all analysis template tests PASS.

---

### Task 2: Make image deployment verifiable and recoverable

**Files:**
- Modify: `.github/workflows/deploy-analysis.yml`
- Test: `tests/analysis-workflow.test.mjs`
- Modify: `README_DEPLOY.md`
- Modify: `infra/analysis/README.md`

**Interfaces:**
- Consumes: stack outputs named `ArtifactBucketName`, `AnalysisInstanceId`, `SandboxInstanceId`, and `SandboxLifecycleTable`; `ANALYSIS_STACK_NAME`, `ANALYSIS_DEPLOY_ROLE_ARN`, and `ANALYSIS_DEPLOY_REGION` repository variables; the existing `scripts/validate-sandbox-host.mjs`.
- Produces: a SHA-256 checked deployment bundle, sandbox-only preflight validation, and explicit old-release rollback on failed deployment/smoke test.

- [ ] **Step 1: Add failing workflow tests**

Extend `tests/analysis-workflow.test.mjs` to assert that deployment:
- Requires `ANALYSIS_STACK_NAME` but no longer requires fixed artifact-bucket or instance-ID variables.
- Calls `cloudformation describe-stacks` and validates the four required outputs before upload or SSM.
- Passes the generated `SandboxLifecycleTable` value to the analysis Compose command as `SANDBOX_LIFECYCLE_TABLE`.
- Bundles and runs `scripts/validate-sandbox-host.mjs` on the sandbox host only.
- Computes and verifies the bundle SHA-256 before extraction and image loading.
- Captures the previous release on both hosts and attempts to restore both after deployment or smoke-test failure.
- Supports first deployment when no Compose service exists; a failed attempt returns that host to the absent-service state without deleting persistent data.
- Keeps the existing manual opt-in, protected environment, immutable image SHA, SSM-only access, and sandbox power-state restoration.

- [ ] **Step 2: Run the targeted workflow test and verify it fails**

Run: `node --test tests/analysis-workflow.test.mjs`
Expected: FAIL on the missing output resolution, checksum, preflight, and rollback assertions.

- [ ] **Step 3: Resolve deployment values from stack outputs**

Use `ANALYSIS_STACK_NAME` and `ANALYSIS_DEPLOY_REGION` to query CloudFormation outputs. Validate each expected output is present exactly once and non-empty before uploading or sending any SSM command. Remove the three fixed resource-ID/bucket variables. Keep role and region requirements and document `cloudformation:DescribeStacks`.

- [ ] **Step 4: Add host preflight and bundle integrity verification**

Include the sandbox validator in the uploaded bundle and invoke it on the sandbox host before `docker compose`. Pass the generated lifecycle-table output to the analysis Compose command, overriding the local example default only for deployment. Compute the archive SHA-256 on the runner; verify it on each host before extraction or `docker load`. Do not run the sandbox-specific validator on the analysis host.

- [ ] **Step 5: Add checked rollback for partial and smoke-test failures**

Capture the active release independently on each host before deployment. Treat an absent project service as a valid baseline; on failure, remove only the current attempt's project containers without deleting persistent data. If a service exists but its release cannot be identified, stop before mutation. On a failure after either host has been changed, attempt to restore prior known-good releases on both hosts, verify service health and broker readiness, and propagate any rollback failure as a failed workflow. Preserve the initial sandbox power state on all paths.

- [ ] **Step 6: Run the targeted workflow test and verify it passes**

Run: `node --test tests/analysis-workflow.test.mjs`
Expected: all workflow tests PASS.

---

### Task 3: Add a protected inventory-and-cleanup path

**Files:**
- Create: `scripts/analysis-cleanup.mjs`
- Create: `.github/workflows/cleanup-analysis.yml`
- Test: `tests/analysis-cleanup.test.mjs`
- Modify: `tests/analysis-workflow.test.mjs`
- Modify: `README_DEPLOY.md`
- Modify: `infra/analysis/README.md`

**Interfaces:**
- Consumes: exact `ResourcePrefix`; AWS CLI credentials supplied only at workflow runtime; stack tags and generated resource names from Task 1.
- Produces: `plan --stack-name <name> --prefix <prefix> --region <region>` read-only inventory output and `apply --plan <file> --confirm-prefix <prefix> --region <region>` destructive cleanup, gated by a protected GitHub Environment, a false-by-default dispatch boolean, an explicit repository enable variable, and exact prefix confirmation.

- [ ] **Step 1: Add cleanup utility tests with a mocked AWS CLI**

Create tests that inject a fake AWS CLI executable and assert:
- Planning performs read-only listing/tagging calls and emits only buckets, tables, volumes, and snapshots with the exact `AnalysisStack=ucs503-analysis-v1` tag; log groups must match the known exact resource names or stack-suffixed forms.
- A resource matching only a name prefix but missing the required ownership tag is not included in the deletion plan.
- Planning refuses while the named CloudFormation stack still exists in a non-deleted state.
- Applying requires the plan prefix and exact confirmation to match, then revalidates resources and ownership before deletion.
- Applying empties all current and versioned S3 objects across pagination, then deletes the bucket; deletes only tagged tables and known exact log groups; deletes only tagged snapshots; and deletes only unattached tagged data volumes.
- No code path invokes commands to terminate EC2 instances, delete AMIs, or touch the archive stack.
- Any AWS CLI error or ownership mismatch stops cleanup with a non-zero result.

- [ ] **Step 2: Run the cleanup utility tests and verify they fail**

Run: `node --test tests/analysis-cleanup.test.mjs`
Expected: FAIL because the cleanup utility does not exist.

- [ ] **Step 3: Implement read-only planning and guarded apply**

Implement the two CLI modes in `scripts/analysis-cleanup.mjs` using the AWS CLI without adding dependencies. Planning must emit deterministic JSON plus a readable inventory. Applying must reject malformed plans, wrong confirmation, a live stack, changed resource identity, missing exact ownership tags, unknown log-group names, attached volumes, and failed AWS operations. Delete S3 versions/delete markers in batches of at most 1,000 and handle pagination markers until exhausted.

- [ ] **Step 4: Add cleanup utility tests and verify they pass**

Run: `node --test tests/analysis-cleanup.test.mjs`
Expected: all cleanup tests PASS, including the safety and pagination cases.

- [ ] **Step 5: Add the manual workflow gates**

Create a `workflow_dispatch` workflow with a read-only inventory job and a destructive job in a protected `analysis-cleanup` environment. The destructive job must require both a false-by-default `enable_cleanup` boolean and `ANALYSIS_CLEANUP_ENABLED == 'true'`, plus exact resource-prefix confirmation. Show the inventory in the job summary before the destructive environment approval. The destructive job must download the exact plan artifact, revalidate it, and invoke `apply`; do not configure AWS roles or environment protections in code.

- [ ] **Step 6: Test the workflow's gate and allowlist**

Extend `tests/analysis-workflow.test.mjs` to assert the cleanup workflow is manual-only, has separate inventory and protected deletion jobs, requires the default-false boolean, the repository enable variable, the protected environment, and exact prefix confirmation, and grants no EC2 instance termination, AMI deletion, or archive-stack operations.

Run: `node --test tests/analysis-workflow.test.mjs`
Expected: all workflow tests PASS.

- [ ] **Step 7: Document review and execution requirements**

Document that the stack must already be deleted before cleanup, how to inspect the inventory, the protected environment requirement, the typed confirmation, the exact resource allowlist, and the least-privilege OIDC permissions required for future operator configuration. State explicitly that apply irreversibly removes S3 object versions, DynamoDB items, the retained EBS data volume, log groups, and tagged snapshots.

---

### Task 4: Run the complete local validation

**Files:**
- Verify: `infra/analysis/template.yaml`
- Verify: `.github/workflows/deploy-analysis.yml`
- Verify: `.github/workflows/cleanup-analysis.yml`
- Verify: `tests/analysis-template.test.mjs`
- Verify: `tests/analysis-workflow.test.mjs`
- Verify: `tests/analysis-cleanup.test.mjs`

- [ ] **Step 1: Run all focused tests**

Run: `node --test tests/analysis-template.test.mjs tests/analysis-workflow.test.mjs tests/analysis-cleanup.test.mjs`
Expected: all tests PASS.

- [ ] **Step 2: Run the Go test suite**

Run: `Set-Location go-backend; go test ./...`
Expected: all Go tests PASS.

- [ ] **Step 3: Validate SAM and the workflows**

Run: `sam validate --lint --template-file infra/analysis/template.yaml`
Expected: valid SAM template with no lint errors.

Run: `actionlint .github/workflows/deploy-analysis.yml .github/workflows/cleanup-analysis.yml`
Expected: no workflow errors. If `actionlint` is unavailable, record that explicitly and rely on workflow tests without installing dependencies.

- [ ] **Step 4: Check whitespace and build**

Run: `git diff --check`
Expected: no whitespace errors.

Run: `sam build --template-file infra/analysis/template.yaml`
Expected: `Build Succeeded`.

- [ ] **Step 5: Confirm scope**

Inspect `git status --short` and the final diff. Confirm there are no AWS API calls from the local session, deployments, cleanup executions, commits, or pushes.
