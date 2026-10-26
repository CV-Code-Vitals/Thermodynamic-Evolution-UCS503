# Analysis Stack Audit Remediation Design

**Status:** Revised draft for user approval  
**Scope:** Local code and documentation changes only. No AWS calls, deployment,
commit, or push.

## Intent

Address the actionable findings in the supplied `ucs503-analysis-v1` audit so
future deployments are less likely to collide with retained resources, target
stale EC2 instances, leave a failed release running, or accumulate untracked
stack-owned data. Keep destructive cleanup separate from normal CloudFormation
stack deletion and require an explicit, protected operator approval.

## Goals

- Remove avoidable globally fixed physical names from S3 buckets and the
  lifecycle table; let the Go service accept a valid configured table name,
  keep `sandbox-lifecycle` as the local default, and use CloudFormation outputs
  for deployment targets and the production table name.
- Make resource retention and cleanup behavior explicit. Ordinary stack
  deletion must not silently erase analysis data. A separate manual cleanup
  workflow must show the exact stack-owned resources, require environment
  approval and a matching confirmation, and exclude the archive stack, AMIs,
  and stopped image-builder instances.
- Improve deployment recovery: validate the sandbox host before changing it,
  verify the uploaded deployment bundle checksum before loading it, and restore
  the prior release on both hosts if deployment or the final smoke test fails.
- Add regression tests for naming, ownership checks, checksum validation,
  deployment-target resolution, and rollback/cleanup gates.
- Update deployment and cleanup documentation, including the permissions that
  must be configured later by an operator.

## Non-goals and constraints

- Do not configure AWS credentials, roles, parameters, endpoints, or resources;
  do not invoke AWS APIs or deploy.
- Do not change the archive stack, its GitHub Pages workflow, AMIs, or stopped
  AMI-builder instances.
- Preserve the already-fixed VPC IPv6 subnet allocation and the existing
  stack-tag scope on the DLM role. Do not add a custom IPv6 waiter without
  evidence that the current CloudFormation dependency is insufficient.
- Do not add interface VPC endpoints or change outbound HTTPS rules in this
  change. Their ongoing cost and required service set need a separate approved
  network design.
- Do not invent a new general-purpose S3 expiration period or delete existing
  AWS resources during implementation.

## Approaches considered

1. **Automatic cleanup on CloudFormation deletion.** Add custom resources to
   empty versioned buckets and delete snapshots during stack deletion. This is
   automatic but makes accidental stack deletion destructive and adds cleanup
   Lambda code and broad delete permissions.
2. **Separate manually approved cleanup workflow — selected.** Keep stack
   deletion non-destructive for retained data. Add a read-only inventory phase
   and a separately gated cleanup phase with protected-environment approval,
   exact stack/prefix confirmation, ownership checks, and a strict resource
   allowlist. This is auditable and avoids accidental data loss; cleanup takes
   an explicit operator action.
3. **Operator-run local script.** This avoids adding GitHub workflow
   permissions, but is less reviewable and repeatable than the selected
   protected workflow.

## Proposed implementation

### Template ownership and naming

- Let CloudFormation generate physical names for the source bucket, artifact
  bucket, and lifecycle table. Remove the fixed lifecycle-table-name parameter.
  Update `NewDynamoSandboxLifecycleStore` to accept DynamoDB table names
  matching `[A-Za-z0-9_.-]{3,255}`; keep the record key `sandbox-lifecycle`
  and the local Compose default unchanged. Pass the generated lifecycle table
  name from the `SandboxLifecycleTable` stack output to the analysis Compose
  command at deployment time.
- Add a stack-unique suffix to explicitly named log groups and Lambda functions
  where old retained names could prevent recreation. Keep useful resource
  prefixes and tags.
- Retain data-bearing resources during ordinary stack deletion. The manual
  cleanup workflow, not a CloudFormation delete policy or custom resource,
  performs approved removal.
- Set ENI delete-on-termination behavior consistently with CloudFormation
  ownership so terminating a managed instance does not leave an attached ENI.
- Keep existing prefix-specific S3 expiry rules and the exact SSM parameter
  path contract. Do not set broader data expiry without an approved retention
  period.

### Deployment workflow

- Require the analysis stack name as a non-secret workflow setting, read the
  artifact bucket and both instance IDs from that stack's outputs, and fail
  closed if outputs are missing, duplicated, or malformed. Document the
  additional read-only CloudFormation permission.
- Package the sandbox host validator and run it on the sandbox host only;
  it is not an analysis-host validator.
- Calculate the deployment archive SHA-256 on the runner and verify it on each
  host before extraction or `docker load`.
- Identify the currently running release on each host before changing it. If
  deployment or the final broker-readiness smoke check fails, attempt to
  restore the prior known-good release on both hosts and verify health. Surface
  rollback failures explicitly; do not report a successful rollback when either
  host cannot be restored. A host with no running service is a valid initial
  state: on failure, remove only the current attempt's project containers and
  leave persistent data untouched. If a running service's release cannot be
  identified, stop before mutation.
- Preserve the existing manual deployment opt-in, GitHub Environment approval,
  immutable image SHA selection, SSM deployment path, and sandbox
  start/stop-state restoration.

### Manual cleanup workflow

- Require the CloudFormation stack to be absent or fully deleted before
  cleanup. A read-only inventory job uses the exact `ucs503-analysis-v1`
  prefix and discovers buckets, tables, volumes, and snapshots only when their
  `AnalysisStack` tag exactly matches that prefix. Log groups must match the
  template's known exact names or stack-suffixed forms. A prefix match alone is
  never sufficient ownership proof. Publish an exact human-readable inventory
  before any deletion job can run.
- Require the protected cleanup environment and an exact confirmation value
  matching the requested resource prefix. Keep cleanup disabled unless both a
  `workflow_dispatch` boolean and the repository variable
  `ANALYSIS_CLEANUP_ENABLED` are explicitly enabled. Require a separately
  configured protected GitHub Environment with reviewers; document that it
  must be configured before the enable variable is set. Re-query and
  re-validate ownership after approval before performing any destructive
  action.
- Limit deletion to this stack's tagged buckets and object versions, tagged
  lifecycle table, tagged analysis data volume and DLM snapshots, and only the
  known exact log-group names. Do not delete resources based on a broad name
  prefix alone or enumerate-and-delete arbitrary account resources. Never
  target AMIs, standalone builder instances, or the archive stack.
- Keep the required OIDC role and environment setup as documented manual
  prerequisites. Code changes must not create or configure them.

### Audit findings intentionally unchanged or deferred

- The IPv6 subnet dependency and distinct `/64` allocation are already present.
- The DLM role already scopes snapshot actions using the `AnalysisStack` tag;
  preserve that condition and the existing policy target tags.
- Outbound IPv6 HTTPS is currently restricted to TCP 443 but has a `::/0`
  destination. Keep it for now; narrowing it safely requires a separate
  endpoint/cost review.
- Keep the current prefix-specific object expiration rules. A broader retention
  period requires an explicit data-retention decision.

## Validation

- Add targeted template and Go tests for generated physical names,
  configurable lifecycle-table names, retained-data policies, ENI lifecycle,
  existing IPv6/DLM guardrails, and stack outputs.
- Extend workflow tests for stack output resolution (including lifecycle-table
  injection), bundle checksum checks, sandbox-only host validation, rollback
  behavior, cleanup inventory and confirmation gates, and the resource
  allowlist.
- Run the repository's relevant Node tests, `sam validate --lint`,
  `git diff --check`, and `sam build`.
- Do not run any AWS command, deployment, cleanup, commit, or push.

## Approval needed

The earlier design approval is superseded by the newly discovered fixed
`sandbox-lifecycle` contract. Confirm this revised design before implementation.
The cleanup workflow is intentionally destructive only after its separate
protected approval; it will delete the exact retained resources shown in its
inventory and will not preserve their S3 objects, DynamoDB items, EBS data, or
DLM snapshots.
