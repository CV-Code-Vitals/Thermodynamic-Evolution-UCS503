# Security Findings Remediation Design

## Intent

Address the actionable findings from the supplied security screenshots without
changing or deploying AWS resources. Keep the existing archive service
functional, maintain the separate analysis/sandbox architecture, and make
deployment failures visible instead of reporting success for a partial or
unhealthy rollout.

## Goals

- Fail analysis-backend startup if its engine binary is missing, not a regular
  file, or not executable on Unix.
- Reduce the sandbox broker's host privilege boundary by using rootless Docker
  on the sandbox host and connecting the broker to that user's socket.
- Make deployment confirm the selected image IDs and backend health. Handle a
  stopped, on-demand sandbox host by starting it only for a gated deployment
  and returning it to stopped state if the workflow started it.
- Narrow SecureString decryption policies to the AWS-managed SSM KMS key alias,
  while retaining the existing SSM service and parameter encryption-context
  conditions.
- Add bearer-token renewal for active admin sessions, preserving in-memory
  token storage and imposing a finite absolute session lifetime.

## Non-goals

- No AWS calls, stack deployment, secret creation, GitHub configuration,
  workflow dispatch, commit, or push.
- No change to the archive API's allowed-origin model; current code uses an
  exact configured-origin allowlist.
- No HttpOnly-cookie migration. The static GitHub Pages frontend and API are
  cross-origin; a cookie/CSRF redesign is broader than necessary for renewal.
- No attempt to fix findings the review confirmed are already addressed:
  mandatory Docker CI skip detection, token persistence claims, authorizer
  context exposure, ordinary cleanup, and bounded broker readiness retries.
- No change to candidate container resource limits or its no-network policy.

## Approaches

### Sandbox isolation

- **Rootless Docker (selected):** Run the sandbox daemon as the dedicated
  unprivileged host user matching the broker UID, and mount only that user's
  socket into the broker. This materially reduces daemon compromise impact
  while preserving the broker's existing Docker API calls. It requires
  rootless-daemon installation and systemd linger configuration in the
  sandbox AMI.
- **Restricted socket proxy:** Add a proxy between broker and host Docker.
  This avoids a rootful socket mount but remains dependent on correctly
  constraining container creation, bind mounts, and daemon API access; it is
  not the selected host boundary.

### Admin session continuity

- **In-memory bearer renewal (selected):** Add a refresh endpoint that accepts
  only a currently valid signed token, rotates it, and refuses renewal beyond
  an absolute session lifetime. The portal renews shortly before expiry and
  clears the session if renewal fails. This preserves current API and browser
  storage behavior.
- **HttpOnly cookie session:** Better resistance to token theft by injected
  page scripts, but requires a cross-origin cookie and CSRF redesign across
  GitHub Pages and the archive API. Not selected for this remediation.

### KMS key scope

- **Restrict by the AWS-managed SSM key alias (selected):** Add
  `kms:ResourceAliases = alias/aws/ssm` to the existing decrypt conditions,
  preserving the current default SecureString encryption and avoiding a
  customer-managed-key migration.
- **Require customer-managed keys:** Scope to explicit key ARNs, but requires
  new key resources, parameter migration, key-policy operations, and cost.
  Not selected.

## Implementation decisions

- Add a testable engine-binary validation helper and make startup terminate
  with the configured path and actionable `ENGINE_PATH` guidance on failure.
- Set `DOCKER_HOST` to the broker container's mounted rootless socket. Replace
  the rootful Docker group mount contract with a required host socket path.
  Document host UID, daemon service, linger, socket permissions, and path
  visibility for candidate workspace bind mounts.
- The deployment workflow records whether the sandbox was stopped before
  deployment. If stopped, it starts it, waits for SSM availability, loads the
  selected image bundle, verifies the broker image ID and runner image ID,
  then stops it again on both success and failure. If it was already running,
  it leaves it running. The analysis host additionally must pass image-ID and
  HTTP health checks. Document the additional narrowly scoped EC2 start/stop
  permissions required by the future protected OIDC role.
- Add an authenticated refresh route and a versioned token payload that can
  preserve issue time while continuing to verify unexpired legacy tokens.
  Renewal is allowed only for tokens created in the new format and stops at
  an eight-hour absolute session age; older tokens require a fresh login.
- Apply KMS alias scoping to the archive and analysis templates, but do not
  deploy either template or change current AWS settings.

## Constraints and calls left to the builder

- Keep each fix surgical and add focused regression tests before production
  changes.
- Do not edit unrelated user changes or regenerate the modified API executable.
- Rootless Docker must be represented consistently in the sandbox Compose
  file, deployment command, example environment, and operator documentation.
  This repository change documents the required sandbox AMI setup but does
  not build, update, or provision an AMI; the host must be prepared separately
  before a future deployment.
- If the installed AWS policy validator rejects `kms:ResourceAliases` for the
  AWS-managed SSM key, stop and report the compatibility blocker rather than
  silently broadening access or introducing a customer-managed key.
