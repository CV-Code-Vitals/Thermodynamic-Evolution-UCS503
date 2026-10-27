> **Superseded draft:** Do not execute this plan independently. The authoritative plan is [2026-10-24-whole-project-hosting-plan.md](2026-10-24-whole-project-hosting-plan.md), which selects a REST API response-streaming Lambda proxy, S3-staged source/image delivery, and the explicit cost/authorization gates.

# AWS Hosting and Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deploy the analysis API and isolated sandbox on separate EC2 instances with a cost-reviewed HTTPS entry point, while leaving the existing serverless archive untouched.

**Architecture:** Use a separate analysis CloudFormation/SAM stack, a public HTTPS API Gateway entry point, an isolated private broker path, least-privilege instance profiles, and SSM-based administration/deployment. Select the least-cost secure ingress and egress design only after an `ap-south-1` comparison; do not assume NAT, a load balancer, public IPv4, or paid interface endpoints.

**Tech Stack:** AWS SAM/CloudFormation, API Gateway HTTP API, Lambda authorizer, EC2, SSM, CloudWatch, GitHub Actions OIDC, GHCR or an approved private artifact path.

**Spec:** `docs/superpowers/specs/2026-10-24-whole-project-two-instance-hosting-design.md`

## Global Constraints

- Preserve the live `ucs503-admin-archive-v2` stack and GitHub Pages archive configuration unchanged.
- Keep the website and Admin Portal UI on GitHub Pages.
- Use two EC2 instances; the analysis instance must not access the Docker socket.
- Keep the broker private with ingress limited to the analysis service.
- Use SSM instead of inbound SSH or long-lived SSH key secrets.
- Do not create RDS or deploy the duplicate `ucs503-api` container.
- Use `ap-south-1`, a few runs per day, one concurrent run, and approximately one-hour recovery as estimate assumptions.
- Treat $20/month out-of-pocket as a target; explain any higher estimate and obtain approval before provisioning.
- No AWS resource mutation, deployment, secret change, or GitHub setting change until separately authorized.
- Keep anonymous Graph scans separate from the authorizer-protected Admin Portal Evolution routes.

## Review Focus

- Network topology must provide end-to-end HTTPS for users and a private, authenticated, TLS-verified broker path without silently adding fixed-cost services.
- The authorizer role may read the token HMAC key; neither EC2 instance role may read it.
- The analysis host must receive only the internal analysis and broker secrets it needs; no browser bundle receives them.
- API Gateway access logs must exclude authorization headers, request/response bodies, and every authorizer context value; payload/data tracing is disabled.
- The broker token is a random 32-byte minimum SecureString at `/ucs503/analysis/sandbox-service-token`; access is limited to the analysis and sandbox runtime roles and rotation requires a documented maintenance/revocation procedure.
- The analysis broker client requires the dedicated CA bundle and trusts no public/system roots for broker connections.
- A failed deployment or instance replacement must not affect the archive stack and must preserve analysis output according to the approved restore policy.
- GitHub Actions must deploy only immutable image references using OIDC/SSM, never SSH private keys or `latest`.

---

### Task 1: Compare and approve costed network and data paths (read-only)

**Files:**
- Read: `docs/superpowers/specs/2026-10-24-whole-project-two-instance-hosting-design.md`
- Read: `docker-compose.analysis.yml` and `docker-compose.sandbox.yml` from the sandbox plan
- Output: Costed ingress/egress decision to be recorded in this plan's implementation notes after user approval; do not edit or deploy AWS resources in this task

**Interfaces:**
- Consumes: The separate analysis/API and sandbox services from the first two sub-plans.
- Produces: A user-approved option with HTTPS ingress, private broker routing, image delivery, SSM access, and itemized recurring/usage costs for `ap-south-1`.

- [ ] **Step 1: Build a like-for-like topology comparison**

Compare (a) API Gateway private integration with VPC Link/internal load balancer, (b) direct HTTPS ingress if a domain/certificate path is available, and (c) any lower-fixed-cost secure alternative validated against API Gateway, streaming, and EC2 networking requirements. For each, include both EC2 instances, EBS, API Gateway, authorizer invocations, network egress/image delivery, CloudWatch, backups, and required endpoints.

- [ ] **Step 2: Estimate low-traffic costs**

Use the AWS Pricing Calculator for `ap-south-1`, few runs per day, one concurrent run, and the one-hour recovery target. Separate monthly fixed charges from usage charges and show the credit-independent cost.

- [ ] **Step 3: Obtain explicit user approval of topology and cost**

If the safe two-instance option is above the $20 target, present the cost drivers and alternatives. Do not silently switch to one EC2 or reduce the sandbox isolation. Do not run `sam deploy`, create a stack, alter SSM parameters, or modify AWS settings in this task.

### Task 2: Create a separate analysis infrastructure template

**Files:**
- Create: `infra/analysis/template.yaml`
- Create: `infra/analysis/samconfig.toml`
- Create: `infra/analysis/README.md`
- Create: `infra/analysis/parameters.example.json`
- Create: `tests/analysis-template.test.mjs`

**Interfaces:**
- Consumes: The approved network path from Task 1; authorizer and runtime contracts from the first two sub-plans.
- Produces: A separate stack named `ucs503-analysis-v1`, with isolated resource names and no references that update or import the archive stack.

- [ ] **Step 1: Add template validation checks**

Add checks in `tests/analysis-template.test.mjs` for two separate EC2 roles/instances, no Docker socket permission on the analysis role, no public broker ingress, no SSH ingress, least-privilege SSM access, encrypted storage, CloudWatch retention, API Gateway route split, and authorizer SSM access limited to exact parameter names.

- [ ] **Step 2: Validate the initial template**

Run from the repository root: `sam validate --lint --template-file infra/analysis/template.yaml`

Expected: FAIL until the separate template and resources are defined.

- [ ] **Step 3: Implement only the approved topology**

Use the approved option from Task 1. Add the API Gateway HTTP API routes: public `POST /api/scan` with route throttling initially set to 1 request/second and burst 2; authorizer-protected admin analysis/Evolution routes; and minimally exposed health checks. Map `$context.identity.sourceIp` to an integration header that overwrites any client-provided scan identity header. Include the Lambda authorizer, SSM SecureString parameter ARNs with separate exact-path read permissions for the HMAC key, internal analysis token, broker service token, broker certificate/private key, and analysis trust CA; private broker security group; instance profiles; persistent encrypted output storage; log retention; and rollback-safe resource names. Configure API access logs with operational metadata only, exclude all authorizer context values and authorization headers/bodies, and do not enable payload/data tracing. Add a test that rejects log formats containing these sensitive fields. Include a secret rotation and emergency revocation runbook. Do not use a paid certificate authority by default without including its cost in the comparison. Do not modify root `template.yaml`.

- [ ] **Step 4: Build and validate without deployment**

Run from the repository root: `sam validate --lint --template-file infra/analysis/template.yaml` and `sam build --template-file infra/analysis/template.yaml`.

Expected: PASS; no AWS resources are created by these commands.

- [ ] **Step 5: Commit**

```text
feat: define isolated analysis hosting stack
```

### Task 3: Add OIDC/SSM deployment automation

**Files:**
- Create: `.github/workflows/deploy-analysis.yml`
- Create: `tests/deploy-analysis-workflow.test.mjs`
- Modify: `.github/workflows/publish-api-image.yml`
- Modify: `.github/workflows/ci-go.yml`
- Modify: `infra/analysis/README.md`

**Interfaces:**
- Consumes: SHA-tagged analysis and sandbox images, the approved OIDC role ARN, the `ucs503-analysis-v1` stack, and SSM Run Command.
- Produces: A manual/environment-gated deployment that assumes a short-lived GitHub OIDC role and uses SSM to roll out images/configuration; no SSH actions or private-key secrets.

- [ ] **Step 1: Add workflow-policy tests**

Add Node `node:test` checks in `tests/deploy-analysis-workflow.test.mjs` that `deploy-analysis.yml` is manual/environment gated, requests only required GitHub permissions, uses OIDC role assumption, pins images to `github.sha`, deploys through SSM, and never references `EC2_SSH_PRIVATE_KEY`, `EC2_HOST`, or `latest`.

- [ ] **Step 2: Verify the current legacy workflow does not satisfy the checks**

Run from the repository root: `node --test tests/deploy-analysis-workflow.test.mjs`

Expected: The new workflow test fails because the secure analysis deployment workflow does not exist.

- [ ] **Step 3: Implement gated deployment and rollback**

Build and publish SHA-tagged images, transfer them using the approved low-cost artifact route, and deploy via SSM. Include the template and workflow policy tests in CI. Require a GitHub environment approval before the deployment job. Keep `.github/workflows/deploy-to-ec2.yml` manual-only and disabled; do not repurpose it for this deployment.

- [ ] **Step 4: Validate CI and deployment artifacts**

Run `node --test tests/deploy-analysis-workflow.test.mjs`, `sam validate --lint --template-file infra/analysis/template.yaml`, `sam build --template-file infra/analysis/template.yaml`, and the Docker/Go tests from the other plans.

Expected: PASS; deployment remains unexecuted until a separate explicit user authorization.

- [ ] **Step 5: Commit**

```text
feat: deploy analysis services through OIDC and SSM
```

### Task 4: Document operations, recovery, and final deployment gate

**Files:**
- Modify: `README_DEPLOY.md`
- Modify: `DEPLOYMENT_CHECKLIST.md`
- Create: `infra/analysis/OPERATIONS.md`

- [ ] **Step 1: Add recovery and non-interference checks**

Document health checks, logs, cost alarms, token rotation, image rollback, one-hour recovery procedure, output backup/restore, and confirmation that the archive stack is not a deployment target.

- [ ] **Step 2: Verify readiness locally**

Run from the repository root: `sam validate --lint --template-file infra/analysis/template.yaml`, `sam build --template-file infra/analysis/template.yaml`, plus all Go, frontend, and sandbox test commands in the other plans.

Expected: PASS; cost estimate and deployment approval are recorded separately.

- [ ] **Step 3: Stop before provisioning**

Do not execute `sam deploy`, `aws cloudformation deploy`, SSM `send-command`, or any AWS mutation until the user separately approves the final costed plan and explicitly authorizes deployment.

- [ ] **Step 4: Commit**

```text
docs: add analysis deployment and recovery runbook
```
