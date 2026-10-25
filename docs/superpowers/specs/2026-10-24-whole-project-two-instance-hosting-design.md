# Design: Hosting the Whole Project with an Isolated Sandbox

## Intent

Host the complete project experience while keeping the already-working public site and Admin Portal archive deployment intact. The website and portal remain on GitHub Pages; archive operations remain on the existing API Gateway, Lambda, DynamoDB, and private S3 stack. Add hosting only for the separate Go analysis backend, Rust engine, and sandbox execution path.

## Why

The analysis and Evolution services are not part of the archive Lambda. `EvolutionLive` currently uses the archive API base URL and does not send an analysis credential, while the Go analysis service requires `ANALYSIS_API_TOKEN`. That token must not be embedded in a browser build. The current EC2 Compose file also includes a duplicate archive API service requiring `DATABASE_URL`, and runs the sandbox broker as root with the host Docker socket mounted. Deploying that file unchanged would duplicate archive responsibilities and put the API host at risk from sandbox compromise.

## Goals

- Keep the main website and Admin Portal UI on GitHub Pages.
- Leave the live `ucs503-admin-archive-v2` resources and deployment path unchanged.
- Host the Go analysis API and Rust engine behind a stable HTTPS API Gateway endpoint.
- Host the sandbox broker and candidate containers on a separate EC2 instance from the public-facing analysis service.
- Restrict sandbox access to the analysis service over private networking; do not expose the broker publicly.
- Allow only an authenticated Admin Portal session to invoke Evolution and other admin analysis routes; do not expose static service credentials to browser code.
- Keep the public Graph Visualizer separate from the Admin Portal and allow anonymous live repository scans under strict abuse and resource controls.
- Restore the Admin Portal Evolution feature using a public GitHub repository as its source, without coupling it to the Graph Visualizer.
- Use SSM for instance administration; avoid inbound SSH and long-lived SSH private-key secrets.
- Keep the analysis API reachable at all times on the smallest instance that passes a representative resource test.
- Start the isolated sandbox instance only for Evolution work, wait for its health check, and stop it after a 10-minute idle period with no active job or workspace.
- Do not require RDS for the initial analysis/Evolution deployment.
- Produce a region-specific cost estimate and obtain approval before any billable AWS changes.
- Target a low-traffic, single-user demo workload and recovery within approximately one hour, without paying for high availability. Aim for $20/month and do not exceed $25/month without further approval.

## Non-goals

- Moving the website or portal UI off GitHub Pages.
- Changing or replacing the live serverless archive stack.
- Deploying the old `ucs503-api` container or creating RDS merely to satisfy its `DATABASE_URL` requirement.
- Exposing Docker, the sandbox broker, or an instance administration port to the public internet.
- Treating two EC2 instances as a complete defense against container escape or host compromise.
- Provisioning or changing any AWS resources as part of this design/review stage.

## Approaches

### 1. Two EC2 instances with a private sandbox boundary — selected

Run the analysis API and Rust engine on one instance and the broker/runner workload on a dedicated sandbox instance. Keep both instances private; front the analysis API with a separate API Gateway HTTP API and private VPC integration. The broker is reachable only from the analysis service security group. This is the selected pragmatic path because it preserves the current service model while separating public request handling from the host Docker daemon.

Trade-offs: more AWS components and a higher monthly floor than one instance; the sandbox host still requires strong hardening and its compromise remains serious. The analysis instance stays on for availability; the sandbox compute is billed only while running, while its EBS volume remains billable while stopped.

### 2. One EC2 instance for the API and sandbox

This is the lowest-compute-cost path and requires fewer network components, but both public-facing request handling and the Docker host share one machine. Because the broker controls the Docker socket and evaluates candidate code, this option has an unacceptable shared blast radius for the intended public deployment.

### 3. Managed sandbox job service

Move candidate execution to an isolated managed job system such as AWS Batch, with analysis API compute separate. This can improve job isolation and scale-to-zero behavior but requires replacing or substantially adapting the current Docker broker flow and adds infrastructure/operational work. It is not the selected first deployment path.

### Analysis ingress and long-running Evolution transport

The current Go Evolution endpoint emits server-sent events (SSE). This rules out the original HTTP API + Cloud Map ingress for that route: HTTP APIs have a non-increasable 30-second integration timeout and do not support response streaming. REST API response streaming supports SSE for up to 15 minutes, subject to a five-minute idle timeout. REST API private integrations do not support Cloud Map.

1. **HTTP API + Cloud Map + asynchronous jobs:** Avoids a load balancer, but replaces the current SSE interaction with a durable job/status/polling flow and requires backend and Admin Portal changes. Not selected because it is a larger user-visible change.
2. **REST API + VPC Link + ALB:** Preserves SSE with less application refactoring, but adds an always-billed load balancer and VPC Link and is unlikely to fit the monthly cap. Not selected.
3. **REST API + response-streaming Lambda proxy — selected:** Keep the existing SSE experience. API Gateway invokes a response-streaming Lambda proxy in the VPC; the proxy forwards requests to the analysis instance over a restricted private connection. This avoids a VPC Link, Cloud Map integration, and load balancer. The API Gateway endpoint provides public HTTPS without a custom domain. The proxy must stream valid API Gateway response metadata and payload, never log credentials or bodies, and send heartbeats often enough to stay below the five-minute stream idle timeout.

### Private source and image delivery

The EC2 instances must not need public IPv4 or NAT. Current DNS lookups returned no AAAA records for `github.com` or `ghcr.io`, so an IPv6-only egress route cannot be assumed to support repository cloning or GHCR pulls. The selected design stages those artifacts instead:

- A narrowly scoped Lambda outside the VPC downloads only allowlisted public GitHub repository archives and writes bounded, temporary snapshots to a dedicated private S3 bucket.
- The VPC proxy invokes the stager using the AWS Lambda dual-stack endpoint; the analysis instance reads snapshots through an S3 Gateway Endpoint and safely extracts them into temporary workspaces. The backend must not run `git clone` against user-supplied URLs.
- GitHub Actions publishes immutable SHA-tagged images and uploads deployment image bundles to a separate private S3 artifact bucket using short-lived OIDC credentials. Private EC2 hosts load the bundles from S3; GHCR remains a publication target only and is not a runtime dependency.
- Use an S3 Gateway Endpoint for private S3 access and an egress-only Internet Gateway for required IPv6-capable AWS control-plane endpoints such as SSM, EC2, and Lambda. Confirm endpoint reachability and IAM permissions before deployment. Do not add NAT, public IPv4, or paid interface endpoints by default.

This staging path adds Lambda, S3 storage, and request costs but avoids fixed NAT, load-balancer, and VPC Link costs. Apply strict GitHub host/path validation, archive byte/file limits, S3 prefixes and lifecycle expiration, and least-privilege permissions. If the staging or IPv6 control-plane path cannot be validated, stop and present the costed alternatives rather than silently adding NAT or paid endpoints.

## Proposed architecture

1. **Static frontend:** Keep GitHub Pages publication unchanged. Keep `VITE_API_BASE` for the archive API. Add a separate analysis API base setting for Evolution so the browser never sends analysis calls to the archive endpoint.
2. **Archive:** Keep the existing `ucs503-admin-archive-v2` API Gateway, Lambda, DynamoDB, private S3 bucket, secrets, and CI/CD unchanged.
3. **Analysis ingress:** Use a separate regional API Gateway REST API with an AWS-generated HTTPS URL and a response-streaming Lambda proxy attached to the analysis VPC. The Lambda proxy connects only to the analysis instance's stable private address and port; use TLS 1.2+ with a self-managed, pinned trust certificate rather than AWS Private CA. Configure `AWS_PROXY` response streaming and a maximum Evolution run shorter than the 15-minute API Gateway stream limit. Send an SSE heartbeat more frequently than the five-minute idle limit. Do not create a VPC Link, Cloud Map integration, ALB, NLB, or public EC2 address. Verify response-streaming availability in `ap-south-1` before implementation.
4. **Analysis instance:** Keep the smallest instance that passes representative scan/Evolution memory and CPU tests continuously running for API availability. Run only the Go analysis service and Rust engine; do not mount `/var/run/docker.sock`. The analysis instance must not receive the archive token-signing key. Use a stable private address, permit its API port only from the proxy Lambda security group, and add the GitHub Pages origin to the analysis API CORS allowlist.
5. **Authentication:** Keep the Graph Visualizer separate from the Admin Portal. Its public live-scan route is anonymous but receives strict API Gateway throttling, server-side per-client rate and concurrency limits, bounded clone/engine timeouts and input sizes, and GitHub-only repository URL validation. The client IP used for limits must come from trusted API Gateway context, not an arbitrary browser header. Admin-only routes such as Evolution require the short-lived bearer token obtained through the unchanged Admin Portal login. A separate, low-volume API Gateway Lambda authorizer verifies that token using the existing HMAC format and signing key read from the existing SSM SecureString. The authorizer is verification-only, has least-privilege access, and is not added to or deployed with the archive API. The streaming proxy forwards the authorizer's internal `analysisApiToken` context only in a server-to-server header, overwriting any client-supplied value. Access logs must contain operational metadata only and must exclude authorization headers, request/response bodies, and all `$context.authorizer.*` values; do not enable payload/data tracing. Add a template test for this log policy. No passkey, HMAC key, sandbox token, or `ANALYSIS_API_TOKEN` may be included in frontend bundles, URLs, logs, or repository files.
6. **Repository input:** Keep Evolution inside the authenticated Admin Portal and independent of the public Graph Visualizer. The browser sends a public GitHub repository URL and target file/function. The proxy obtains a bounded archive from the GitHub stager, then passes an unforgeable internal snapshot reference to the analysis service. The service reads only the corresponding staging prefix from S3, safely extracts the archive, and resolves the target; users never submit server-local paths or S3 keys. Public Graph scans use the same stager and remain anonymous only at the API route, subject to stricter rate and size limits.
7. **Sandbox lifecycle:** Keep the broker and candidate runner on a separate private EC2 instance with no public IP or inbound SSH. Only Evolution requests can start it. The analysis service uses a narrowly scoped instance role to start/describe/stop only this sandbox instance, waits for the authenticated broker health check before continuing, and returns a retryable error if startup exceeds the bounded readiness timeout. Stop the instance only when there are no active jobs or workspaces and the 10-minute idle timer expires; include a recovery/watchdog path for a crashed analysis process so the sandbox does not remain billed indefinitely. Do not mount the sandbox Docker socket on the analysis instance.
8. **Sandbox security and workspace transfer:** Allow broker traffic only from the analysis service security group. Require a cryptographically random shared service token of at least 32 bytes in `X-Sandbox-Token`, constant-time comparison, TLS 1.2 or newer, and a dedicated configured CA trust bundle on the analysis client; do not fall back to system/public trust roots. Store the token as `/ucs503/analysis/sandbox-service-token` SecureString and grant read access only to the two runtime roles that need it. Rotate the token in a maintenance window and document emergency revocation. Preserve existing CPU, memory, process, timeout, and admission limits, and enforce non-root candidate containers, dropped capabilities, no privileged mode, restricted networking, and ephemeral writable storage. Send a bounded compressed source snapshot over the private broker connection; extract it into an ephemeral workspace with traversal and size protections. Do not add EFS.
9. **Private egress and artifact delivery:** Keep EC2 instances private with no public IPv4. Use an S3 Gateway Endpoint for source/image artifacts and an egress-only Internet Gateway only for validated dual-stack AWS control-plane endpoints needed by SSM, EC2 lifecycle control, and Lambda invocation. Keep the source-stager Lambda outside the VPC so it can reach public GitHub. Mirror immutable CI image bundles into a separate private artifact bucket with short-lived GitHub OIDC credentials; instances load images from S3, not GHCR. Validate endpoint reachability, permissions, S3 lifecycle, and artifact limits before deployment. Do not use NAT, paid interface endpoints, or AWS Private CA without separate approval.
10. **Operations and data:** Use least-privilege instance and Lambda roles, encrypted storage, CloudWatch logs/metrics with retention, health checks, SSM administration, and cost alarms. Document credential rotation and emergency revocation for the HMAC key, internal analysis token, broker token, and certificates. Keep analysis deliverables on explicitly persistent encrypted storage with a backup/restore procedure; do not introduce RDS. GitHub source snapshots and deployment artifacts use separate S3 buckets/prefixes and short lifecycle retention; do not reuse or alter the archive bucket.
11. **Existing deployment assets:** Reuse SHA-tagged images where practical, but do not deploy `ucs503-api` from `docker-compose.ec2.yml`. The existing Compose file is a reference, not a production-ready deployment manifest.

## Cost and approval constraints

- Use `ap-south-1`, 720 hours/month, and a planning assumption of 30 sandbox-running hours/month (one hour/day) for a first low-traffic comparison. Include both EC2 instances, EBS, REST API request/data charges, authorizer and proxy/stager Lambda duration, S3 snapshots and image artifacts, logs, backups, and likely data transfer.
- Current Mumbai public price-list candidates: `t3a.small` is $0.0123/hour and `t3a.medium` is $0.0246/hour; gp3 is $0.0912/GB-month. Two 20-GB gp3 volumes cost about $3.65/month. With an always-on `t3a.small` analysis instance, a `t3a.medium` sandbox running 30 hours/month, and those volumes, the compute/storage baseline is about $13.25/month before API, Lambda, S3, logs, backup, and transfer usage. This is not a final quote, and the analysis `t3a.small` is not approved until a representative resource test passes.
- If the analysis instance requires `t3a.medium`, the same baseline is about $22.10/month before usage-based costs. Do not deploy at an estimated total above $25/month without explicit approval; show any higher estimate and alternatives first. AWS credits do not count toward the cap.
- The selected topology avoids fixed NAT, VPC Link, load-balancer, public IPv4, Cloud Map, and AWS Private CA charges. Include any required endpoint, S3 storage, internet transfer, log, or backup charges rather than treating them as free.
- Estimate for a few analysis/Evolution runs per day with one concurrent run. This is a planning baseline, not a guaranteed capacity commitment.
- Target roughly one-hour recovery for this low-traffic demo. Do not add multi-AZ hot standby or other high-availability capacity unless later requirements change.
- If the safe two-instance design cannot meet the target under realistic low-usage assumptions, present the cost drivers and alternatives before provisioning; do not silently downgrade to the one-instance design.
- No AWS resource, secret, or GitHub setting may be created, updated, or deleted until the user separately approves the design, final itemized cost estimate, deployment plan, and then explicitly authorizes deployment.

## Builder decisions

- Validate the REST API/Lambda response-streaming path in `ap-south-1`, including SSE framing, CORS, authorization-context forwarding, and stream/idle limits.
- Validate the source-stager and S3 artifact flow end to end without enabling public IPv4 or NAT on the EC2 instances.
- Confirm IPv6 reachability for the selected SSM, EC2, and Lambda dual-stack endpoints from a private subnet; use only free S3 Gateway Endpoint routing for private S3 access.
- Treat NAT Gateways, load balancers, public IPv4 addresses, AWS Private CA, and paid interface endpoints as exceptions requiring a documented justification, itemized cost, and approval.
- Select the minimal durable-storage approach for analysis outputs and define retention, backup, and restore behavior.
- Define health checks, log retention, measured instance sizing, analysis availability, sandbox start/readiness/idle-stop behavior, crash recovery, and operational alarms.
- Define and test the authorizer-to-analysis trust boundary; never grant the analysis EC2 role access to the archive token-signing key.
- Confirm the existing runner resource limits and network restrictions against the actual Dockerfiles and Compose configuration.
- Set conservative per-client scan rate, concurrency, clone time, repository size, and allowed GitHub URL limits; reject excess work explicitly.
- Set the compressed/expanded workspace-transfer limits, reject malformed archives, and prove the broker needs no shared host path or persistent volume.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| A compromised broker or container can affect the sandbox host through its Docker daemon. | Put the broker on a dedicated instance, keep it private, restrict who can call it, harden runner containers, and document residual host risk. |
| A browser-embedded shared API token would be extractable by every visitor. | Use the Admin Portal's short-lived bearer session for user authentication; keep service credentials server-side. |
| A compromised analysis host could forge archive tokens if it received the archive signing key. | Keep that key accessible only to the separate authorizer role; pass only an authorization result across the private API integration. |
| Anonymous repository scans could be abused to consume compute or probe internal resources. | Apply route-level API Gateway throttling, trusted-client per-IP limits, one-run-at-a-time admission, strict repository URL validation, and bounded clone/engine work. |
| Analysis and sandbox EC2 instances do not share the existing Compose named volumes or repository paths. | Transfer a bounded ephemeral repository snapshot through the authenticated private broker protocol; never assume a shared path across hosts. |
| A large or malicious repository snapshot could exhaust memory, disk, or broker request limits. | Cap compressed and expanded bytes/files, reject unsafe archive paths, use temporary storage, and clean up on all exits. |
| Existing Compose deployment starts a redundant archive API and expects a database. | Do not deploy `ucs503-api`; preserve archive operations on the existing serverless stack. |
| Private instances need image and operating-system updates. | Compare secure egress and endpoint options in the cost/design review; do not add a NAT Gateway or paid endpoint without showing its recurring cost and alternatives. |
| Private EC2 instances cannot resolve public GitHub/GHCR hosts to IPv6 in current DNS checks. | Stage bounded public GitHub archives through a non-VPC Lambda and deploy SHA-tagged image bundles through a separate private S3 bucket and Gateway Endpoint. Revalidate before deployment. |
| Long-running SSE requests exceed HTTP API limits or API Gateway stream idles. | Use REST API response streaming through the Lambda proxy; cap job duration below 15 minutes and send heartbeats more frequently than every five minutes. |
| A sandbox start/stop race disrupts an Evolution job or leaves compute running. | Serialize Evolution, track active workspaces, cancel idle-stop when new work arrives, stop only after the idle grace period, and add a watchdog for stale leases. |
| EC2-local files can be lost or become unavailable during replacement. | Use encrypted persistent storage, tested backups, and documented restore steps; confirm output retention before implementation. |
| Two instances plus network services may exceed the desired monthly spend. | Estimate all fixed and variable costs in `ap-south-1` before approval; show a lower-cost option without representing it as equally safe. |
| Infrastructure changes could disrupt the working archive. | Use a separate analysis stack/API, separate frontend configuration, and no edits to the archive stack during initial backend hosting. |

## Acceptance criteria

- [ ] Main website and Admin Portal continue to load from GitHub Pages.
- [ ] Archive login, upload, list, download, and delete continue to use the existing serverless archive API without changes.
- [ ] Evolution calls a separate HTTPS analysis API and succeeds only for a valid, unexpired Admin Portal session.
- [ ] Evolution SSE is streamed through the REST API/Lambda proxy, stays within the 15-minute stream and five-minute idle limits, and retains user-visible progress and cancellation behavior.
- [ ] The analysis API remains reachable while idle; only an authenticated Evolution job can start the sandbox, and the sandbox stops after the approved idle period once all work is complete.
- [ ] The public Graph Visualizer remains independent of the Admin Portal and can start anonymous scans only within the approved route, rate, concurrency, URL, size, and timeout limits.
- [ ] Anonymous scans cannot target loopback, private, link-local, or metadata-service addresses, and rate-limit identity cannot be forged by a client-supplied header.
- [ ] Evolution remains inside the Admin Portal, requires its existing login, accepts a public GitHub repository and target file/function, and does not depend on Graph Visualizer page state.
- [ ] A sandbox evaluation succeeds with analysis and broker on separate hosts using a bounded ephemeral source snapshot; it does not rely on shared host paths or volumes.
- [ ] Missing, malformed, expired, and tampered bearer tokens are rejected by the separate authorizer; the analysis EC2 role cannot read the archive HMAC signing key.
- [ ] No service token, passkey, signing key, or AWS credential appears in frontend assets or requests other than the short-lived bearer token.
- [ ] The analysis instance has no Docker socket; only the dedicated private sandbox instance runs the broker.
- [ ] The broker has no public ingress and accepts requests only from the analysis service; candidate containers meet documented resource and isolation limits.
- [ ] The old `ucs503-api` service and RDS are not required for the initial deployment.
- [ ] Analysis outputs survive instance replacement according to the approved backup/restore policy.
- [ ] CI validates images/configuration and deployment can be rolled back without affecting the archive stack.
- [ ] A complete `ap-south-1` estimate compares secure ingress/egress alternatives and itemizes recurring networking and storage costs; NAT Gateways, load balancers, public IPv4, and paid endpoints are included only when justified and approved.
- [ ] Public GitHub repositories and SHA-tagged images reach private instances through the approved S3 staging path; private EC2 hosts do not require direct IPv4 internet access.
- [ ] The complete monthly estimate, including usage and a resource-tested instance size, is at or below $25 or has explicit additional user approval.
- [ ] No AWS resources are changed before separate user authorization.

## User choices captured

- The goal is to host the whole project: main website, Admin Portal archive, analysis/Evolution backend, and sandbox execution.
- Keep the currently working Admin Portal archive deployment unchanged.
- Use two EC2 instances, separating analysis API from the sandbox worker.
- Keep the analysis API always reachable; start the sandbox automatically for Evolution and stop it after a 10-minute idle period with no active work.
- Plan around low traffic, single-user demo use, and approximately one-hour recovery.
- Aim for $20/month and require explicit approval before exceeding $25/month.
- Avoid NAT, load balancers, public IPv4, AWS Private CA, and paid interface endpoints; the selected ingress is REST API plus a VPC Lambda response-streaming proxy.
- Use an S3-staged GitHub source and image delivery path so private EC2 does not depend on IPv4-only GitHub/GHCR endpoints.
- Use a separate API Gateway Lambda authorizer for Admin Portal-only routes; keep the archive API Lambda and infrastructure unchanged.
- Keep Graph Visualizer separate and allow anonymous scans only with strict server-side abuse and resource controls.
- Keep Evolution independent of Graph Visualizer state; accept an admin-supplied public GitHub repository and target file/function.
- Use bounded ephemeral repository transfer between analysis and sandbox hosts instead of shared filesystem infrastructure.
- Do not provision or change AWS resources, secrets, or GitHub settings during discovery/design; provide the final cost before provisioning.
- Do not provision or change AWS resources during discovery/design.
