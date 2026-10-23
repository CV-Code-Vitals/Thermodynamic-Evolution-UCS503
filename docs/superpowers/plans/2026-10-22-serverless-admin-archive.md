# Serverless Admin Archive Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep the public site and Admin Portal on GitHub Pages while moving authenticated archive login, listing, upload, download, and deletion to a cost-conscious AWS serverless backend.

**Architecture:** Refactor the existing Go HTTP API into a shared handler used by local development and a Go Lambda adapter behind API Gateway HTTP API. Store archive metadata in DynamoDB and files in a private S3 bucket; use short-lived HMAC bearer tokens, short-lived presigned S3 POST/GET requests, and a pending-upload/finalize flow. Keep the portal static, store its token in React memory only, and use SAM for repeatable infrastructure.

**Tech Stack:** Go 1.25, AWS Lambda, API Gateway HTTP API, AWS SDK for Go v2, DynamoDB, private S3, AWS SAM, React 19/Vite, Node's built-in test runner.

**Spec:** `docs/superpowers/specs/2026-10-22-serverless-admin-archive-design.md`

## Global Constraints

- AWS region: `ap-south-1` (Mumbai); no custom domain, EC2, RDS, CloudFront, or Cloudflare dependency.
- Keep the public website and Admin Portal UI on GitHub Pages; the production browser origin is `https://zylus08.github.io`.
- Support PDF, PPT, and PPTX only; maximum file size is 50 MiB (52,428,800 bytes).
- Validate extension, expected MIME type, and file signature server-side; validate in the browser only as an early user-feedback measure.
- Keep S3 Block Public Access enabled and use SSE-S3; never put AWS credentials or admin secrets in frontend code or the built site.
- Require a valid short-lived bearer token on every admin API endpoint other than login; use 15-minute tokens and keep them in React memory only.
- Expire upload and download presigned URLs after 5 minutes. Use unique server-generated object keys, not user filenames.
- Expire only abandoned staging objects after one day and pending DynamoDB items after two days; never apply expiration to available archive objects.
- Restrict production CORS to the GitHub Pages origin and required methods/headers; allow local origins only in local development.
- Configure S3 CORS separately for browser upload/download, allowing only `https://zylus08.github.io` and required `POST`/`GET` methods.
- Start with an empty AWS archive; do not migrate existing JSON/PostgreSQL files or metadata.
- Keep Evolution/analysis out of this release and clearly show it as unavailable.
- Assume 10 uploads/month at about 10 MB each, 2 GB initial stored data, and 1 GB monthly downloads for the required AWS Pricing Calculator estimate.
- Treat the existing $20 monthly billing alert as notification only, not a spending cap. Do not create or modify AWS resources, deploy the SAM stack, or enter cloud credentials without the user's explicit approval.
- Use bounded request sizes/timeouts, explicit errors, structured logs, least-privilege IAM, and tests for each storage/auth failure path.

## Review Focus

- **Expired, malformed, or modified bearer tokens:** auth tests must reject them before any protected handler or AWS operation runs.
- **Mismatched extension, MIME type, or file signature (including a renamed ZIP masquerading as PPTX):** validation tests must reject it before metadata is made available.
- **Missing, oversized, wrong-type, or expired staged upload:** finalize tests must reject it and leave no available archive record.
- **S3/DynamoDB partial failures during deletion:** tests must prove that the item is not falsely reported as deleted and that a retry can complete cleanup.
- **Untrusted filenames, traversal strings, same-name uploads, and unauthorized object IDs:** tests must prove server-generated keys cannot collide or escape the controlled `pending/` and `archives/` prefixes.

---

## File Structure

- `api/main.go`: keep the local server entry point and initialization; retain local development support.
- `api/router.go`: create shared, dependency-injected HTTP routes for both local Go and Lambda execution.
- `api/auth.go`, `api/auth_test.go`: passkey verification, short-lived HMAC bearer-token issue/verification, and secret loading.
- `api/archive.go`, `api/archive_test.go`: archive metadata types, accepted-file validation, upload state transitions, and testable service interfaces.
- `api/aws_store.go`, `api/aws_store_test.go`: DynamoDB repository and private S3/presign operations behind small interfaces.
- `api/local_store.go`, `api/local_store_test.go`: local-only filesystem/JSON implementation and authenticated upload endpoint so end-to-end local development does not require AWS credentials.
- `api/lambda.go`, `api/lambda_test.go`: API Gateway v2 to `net/http` adapter and Lambda startup.
- `api/Makefile`: reproducible Linux `bootstrap` binary build for the SAM `provided.al2023` runtime.
- `api/main_test.go`: preserve and adapt existing local/private-download, origin, and storage failure tests.
- `api/go.mod`, `api/go.sum`: add only AWS Lambda adapter/runtime and DynamoDB SDK dependencies needed by the implementation.
- `template.yaml`: SAM HTTP API, Go Lambda, private S3 bucket, DynamoDB table/index, bounded log retention, CORS, and least-privilege role policies.
- `admin-portal/src/components/AdminPortal.jsx`: memory-only auth token, archive-only file types, presigned upload/download, delete confirmation, and Evolution deferred state.
- `admin-portal/src/archive-api.js`, `admin-portal/tests/archive-api.test.js`: authenticated API request helpers with injectable `fetch` for Node tests.
- `admin-portal/src/archive-download.js`, `admin-portal/tests/archive-download.test.js`: retain the public helper surface where practical while adapting downloads to a short-lived URL and bearer authorization.
- `.github/workflows/ci-go.yml`: run relevant Go API tests and builds; preserve the existing analysis backend checks.
- `.github/workflows/deploy-frontend.yml`: continue requiring the API Gateway HTTPS URL via `VITE_API_BASE` and run frontend tests before publishing.
- `.env.example`: document local-only configuration and public frontend API URL without real secrets or AWS keys.
- `README_DEPLOY.md`, `DEPLOYMENT_CHECKLIST.md`: replace stale EC2/RDS deployment instructions with SAM setup, local configuration, cost/billing review, explicit approval gate, deployment, rollback, and cleanup steps.
- `.github/workflows/deploy-to-ec2.yml`: keep the legacy EC2 path clearly disabled from the serverless workflow; do not leave an automatic EC2 deploy path enabled by the new instructions.

## API Contract

Keep the existing API prefix and archive-facing routes so the frontend change is contained:

- `POST /api/login`, JSON `{ "passkey": "..." }` -> `200 { "status": "authenticated", "token": "...", "expiresAt": "..." }`; invalid passkey -> `401`.
- `POST /api/upload`, bearer auth and trusted origin, JSON `{ "originalName": "...", "contentType": "...", "size": 123, "title": "...", "version": "...", "date": "YYYY-MM-DD", "summary": "..." }` -> `201 { "uploadId": "...", "url": "...", "fields": { ... }, "expiresAt": "..." }`.
- Browser submits the returned form fields and file to the presigned S3 POST URL; AWS credentials never reach the browser.
- `POST /api/upload/complete`, bearer auth and trusted origin, JSON `{ "uploadId": "..." }` -> `201` with the available `ArchiveEntry`, only after S3 object verification and metadata persistence.
- `GET /api/deliverables`, bearer auth -> newest-first array of available `ArchiveEntry` records.
- `GET /api/download?filename=<opaque archive id>`, bearer auth -> `200 { "url": "...", "expiresAt": "..." }`; unknown IDs -> `404`.
- `DELETE /api/archive?filename=<opaque archive id>`, bearer auth and trusted origin -> `204` only after object and metadata deletion finish; missing IDs -> `404`.
- `GET /api/status` remains a public health check. All other admin operations require valid authorization.

Use a 5-minute upload authorization, a 5-minute download URL, and a 15-minute HMAC token. Validate a filename extension and declared MIME against the allowlist, then verify file signatures on the server before exposing archive metadata. Validate PPTX as an OOXML ZIP package, not merely any ZIP file. Reject names containing path separators or traversal, normalize the stored display filename to a basename, and cap filename/title/version/summary lengths at 255/200/64/2,000 UTF-8 bytes respectively; validate `date` as an ISO calendar date.

For uploads, create a DynamoDB `pending` record with a two-day TTL and generate an opaque UUID key under `pending/`. Enforce a 50 MiB `content-length-range` and the allowed MIME value in the S3 POST policy. A presigned client can write only under `pending/`, never to `archives/`; the permanent archive key is a separate server-generated UUID and is never overwritten by the browser. On completion, verify the object size, stored content type, and signature; copy the verified object to the permanent key under `archives/`; update the DynamoDB record to `available` and remove its TTL; then remove the staged object. Configure an S3 lifecycle rule to expire only objects under `pending/` after one day. If a finalize step fails, return an explicit error, leave the record retryable, compensate for partial copy failures, and log enough structured context to diagnose/retry safely. Local development returns a local-only upload endpoint instead of an S3 URL; that endpoint enforces the same authentication, size/type/signature validation, and staged/finalize states using the filesystem and JSON metadata.

For deletion, require the token and trusted origin, confirm in the UI, transition the DynamoDB record to `deleting`, delete the private S3 object, then delete the DynamoDB record. On S3 failure, restore `available` if possible and return an error. If metadata removal fails after S3 deletion, leave a retryable `deleting` record; a repeated delete must safely treat an already-missing S3 object as deleted and finish removing metadata. Emit a structured deletion outcome log without logging tokens, passkeys, signed URLs, or file contents.

## Tasks

### Task 1: Extract shared API routing and define authenticated archive contracts

**Files:**
- Create: `api/router.go`
- Create: `api/auth.go`
- Create: `api/auth_test.go`
- Create: `api/archive.go`
- Create: `api/archive_test.go`
- Modify: `api/main.go`
- Modify: `api/main_test.go`

**Interfaces:**
- `newRouter(deps Dependencies) http.Handler` constructs the same routes for local development and Lambda without connecting to AWS at package initialization.
- `type Dependencies struct { Auth Authenticator; Archives ArchiveService; Logger *slog.Logger }`.
- `type Authenticator interface { Issue(passkey string) (token string, expiresAt time.Time, err error); Verify(token string, now time.Time) bool }`.
- `type ArchiveService interface { BeginUpload(context.Context, UploadMetadata) (PresignedUpload, error); CompleteUpload(context.Context, string) (ArchiveEntry, error); List(context.Context) ([]ArchiveEntry, error); Download(context.Context, string) (PresignedDownload, error); Delete(context.Context, string) error }`.
- `type ArchiveEntry struct` carries an opaque ID, sanitized original filename, content type, size, title/version/date/summary, uploaded timestamp, and availability state; user-supplied names never become storage keys.
- `validateUpload(name, contentType string, size int64, signature []byte) error` rejects unsupported extension/MIME/signature combinations and sizes outside `1..52,428,800`.
- `validateArchiveMetadata(title, version, date, summary string) error` enforces the exact field limits and ISO date format in the API Contract section.

- [ ] **Step 1: Add failing unit tests** for valid/invalid passkeys, token expiry and tampering, accepted PDF/PPT/PPTX signatures, invalid MIME/extension combinations, a renamed arbitrary ZIP, path traversal names, empty files, 50 MiB plus one byte, metadata byte limits, and invalid dates.
- [ ] **Step 2: Run the targeted Go tests** with `cd api; go test ./...`; confirm the new tests fail for the missing pure auth/validation behavior.
- [ ] **Step 3: Implement token and file-validation functions** in `api/auth.go` and `api/archive.go`. Use HMAC-SHA-256, constant-time MAC/passkey comparisons, cryptographically random token material, 15-minute expiry, and the exact accepted extensions/MIME values from Global Constraints.
- [ ] **Step 4: Extract the route registration into `newRouter`** and keep the local `main` path using it; preserve the existing local disk/JSON mode and health route. Remove any permissive/fake auth fallback from production.
- [ ] **Step 5: Run `cd api; go test ./...; go vet ./...; go build ./...`** and confirm existing local and private-download tests still pass.

### Task 2: Implement private S3 and DynamoDB archive operations

**Files:**
- Create: `api/aws_store.go`
- Create: `api/aws_store_test.go`
- Modify: `api/archive.go`
- Modify: `api/go.mod`
- Modify: `api/go.sum`

**Interfaces:**
- `type ArchiveRepository interface { CreatePending(context.Context, ArchiveEntry) error; Get(context.Context, string) (ArchiveEntry, error); ListAvailable(context.Context) ([]ArchiveEntry, error); SetStatus(context.Context, string, string) error; Delete(context.Context, string) error }`.
- `type ObjectStore interface { CreateUpload(ctx context.Context, pendingKey, contentType string, size int64, expiresIn time.Duration) (PresignedUpload, error); VerifyUpload(ctx context.Context, pendingKey, originalName, contentType string, expectedSize int64) error; PromoteUpload(ctx context.Context, pendingKey, archiveKey string) error; PresignDownload(ctx context.Context, archiveKey string, expiresIn time.Duration) (string, error); DeleteObject(ctx context.Context, key string) error }`.
- `type PresignedUpload struct { URL string; Fields map[string]string; ExpiresAt time.Time }`.
- `type UploadMetadata struct { OriginalName, ContentType, Title, Version, Date, Summary string; Size int64 }`.
- `type PresignedDownload struct { URL string; ExpiresAt time.Time }`.
- Implement `archiveService` with those method signatures to compose the repository/object-store interfaces; all storage calls accept request contexts and use bounded timeouts.

- [ ] **Step 1: Add failing fake-backed tests** for pending record creation, wrong/missing/oversized object rejection, PDF/PPT/PPTX signature verification (including OOXML PPTX structure), promotion to `archives/`, newest-first listing, 300-second upload and download URL expiries, expired-URL upload failure, repeated presigned upload not overwriting the permanent archive object, compensation after copy/metadata failures, and every S3/DynamoDB operation failure.
- [ ] **Step 2: Run `cd api; go test ./...`** and confirm the new service tests fail before implementation.
- [ ] **Step 3: Add the AWS SDK v2 DynamoDB dependency and update the S3 SDK to a release with `s3.PresignClient.PresignPostObject`**; implement the repository/object-store adapters. The S3 POST policy must bind the unique pending key, accepted content type, 50 MiB `content-length-range`, and 5-minute expiry; never log the full URL or POST fields.
- [ ] **Step 4: Implement upload finalization and deletion state transitions** exactly as specified above. Make finalize and delete retry-safe; list only records whose state is `available`.
- [ ] **Step 5: Run `cd api; go test ./...; go vet ./...; go build ./...`** and confirm fake AWS tests prove failures are surfaced without success-shaped fallbacks.

### Task 3: Preserve local end-to-end uploads without AWS

**Files:**
- Create: `api/local_store.go`
- Create: `api/local_store_test.go`
- Modify: `api/archive.go`
- Modify: `api/router.go`

**Interfaces:**
- The local object-store adapter returns an upload URL under the local API and an opaque one-use upload ID; production continues returning an S3 presigned POST.
- The local upload route accepts only the issued pending ID, requires the admin bearer token, limits the request to 50 MiB, and writes to a unique staging path before the shared finalize validation.
- Local JSON metadata and files remain under the existing configured archive/upload paths; no browser-only or success-shaped fallback is permitted.

- [ ] **Step 1: Add failing local end-to-end tests** for the authorized upload/finalize/list/download/delete flow, missing/expired local upload ID, oversized body, invalid signatures, and local storage write failure.
- [ ] **Step 2: Run `cd api; go test ./...`** and confirm the local flow tests fail before implementation.
- [ ] **Step 3: Implement the filesystem/JSON adapters** using the same `ArchiveService` rules as AWS and the existing local paths/configuration.
- [ ] **Step 4: Use the same 5-minute expiry for local upload authorizations** and reject expired or already-finalized upload IDs.
- [ ] **Step 5: Run `cd api; go test ./...; go vet ./...; go build ./...`** and verify local operation needs no AWS credentials.

### Task 4: Expose the archive API through Go Lambda and preserve local execution

**Files:**
- Create: `api/lambda.go`
- Create: `api/lambda_test.go`
- Modify: `api/router.go`
- Modify: `api/main.go`
- Modify: `api/Dockerfile`
- Modify: `scripts/run-local.ps1`
- Modify: `scripts/run-local.sh`
- Modify: `api/go.mod`
- Modify: `api/go.sum`
- Modify: `api/main_test.go`

**Interfaces:**
- `api/lambda.go` adapts API Gateway v2 requests to the shared `http.Handler`, preserving method, path, body, headers, query, and request context.
- Routes use the API Contract. Login is unauthenticated; status is public; upload initiation/finalization, archive listing, download URL creation, and deletion all require bearer authentication.
- The production token verifier reads the passkey and signing key from SSM standard-tier SecureString parameter names encrypted with the AWS-managed SSM key; local mode reads `ADMIN_PASSKEY` and may read `TOKEN_HMAC_KEY` from the developer's environment, generating an ephemeral random signing key if none is set.

- [ ] **Step 1: Add failing handler tests** for login success/failure, every route without a token, malformed/expired token, trusted-origin enforcement for mutations, unsupported methods, unknown archive IDs, and API Gateway v2 request/response mapping.
- [ ] **Step 2: Run `cd api; go test ./...`** and confirm handler/adapter tests fail before wiring.
- [ ] **Step 3: Wire routes to `ArchiveService`** with stable JSON errors, structured `slog` fields, request-scoped contexts, explicit timeouts, and no logs containing secrets or presigned URLs.
- [ ] **Step 4: Add the Lambda startup path** using the existing shared router and AWS Lambda Go HTTP adapter; build the production HMAC authenticator and DynamoDB/S3-backed `ArchiveService` from SSM parameters and Lambda role credentials, then pass both through `newRouter` (fail closed if any dependency/secret is unavailable). Update local launch scripts and the Docker build from `go run/build main.go` to package mode (`go run .` / `go build .`) so they include the new Go files and still require no AWS credentials locally.
- [ ] **Step 5: Run `cd api; go test ./...; go vet ./...; go build ./...`** and verify both local and Lambda paths use the same auth and handler rules; run the Windows local script's API startup command as a smoke check.

### Task 5: Add SAM infrastructure with private defaults and least-privilege access

**Files:**
- Create: `template.yaml`
- Create: `api/Makefile`
- Modify: `.github/workflows/ci-go.yml`

**Interfaces:**
- SAM provisions one API Gateway HTTP API, one Go Lambda, one DynamoDB on-demand table with a status/time index and TTL on pending records, and one private S3 bucket.
- Lambda receives only configuration values and SSM parameter names; secrets are not template literals, GitHub Pages build variables, or frontend bundle values.
- IAM allows only the required S3 actions on `pending/*` and `archives/*`, required DynamoDB actions on the one table/index, and read/decrypt access to the two configured SSM SecureString parameters using the AWS-managed SSM key.
- API Gateway CORS allows only `https://zylus08.github.io`, `GET/POST/DELETE/OPTIONS`, and `Authorization`/`Content-Type`; local origins are not production template defaults.
- S3 CORS separately allows only the same frontend origin and the browser's required `GET`/`POST` operations and headers.

- [ ] **Step 1: Add template validation checks** for public-access block, SSE-S3, exact API and S3 CORS origins, 1-day `pending/` lifecycle only, 2-day pending-item TTL, 50 MiB API-independent S3 upload path, DynamoDB on-demand mode, bounded Lambda timeout/concurrency, 14-day log retention, and resource-retention protections.
- [ ] **Step 2: Run `sam validate --lint`** and confirm the checks fail before the template exists or has the required properties.
- [ ] **Step 3: Implement `template.yaml` and `api/Makefile`** using Go `provided.al2023`, a 512 MiB Lambda with a 25-second timeout, API Gateway's generated HTTPS endpoint, and narrow explicit IAM statements. Do not create an AWS Budget resource; the existing alert is reviewed separately.
- [ ] **Step 4: Install the AWS SAM CLI in CI and add validation for `sam validate --lint` and `sam build`** without invoking `sam deploy` or requiring production secrets; include `template.yaml` in the workflow path filters.
- [ ] **Step 5: Run `sam validate --lint`, `sam build`, and the Go test/build/vet commands**; verify CloudFormation build/validation does not create AWS resources.

### Task 6: Integrate memory-only authentication and archive workflows in the portal

**Files:**
- Create: `admin-portal/src/archive-api.js`
- Create: `admin-portal/tests/archive-api.test.js`
- Modify: `admin-portal/src/archive-download.js`
- Modify: `admin-portal/tests/archive-download.test.js`
- Modify: `admin-portal/src/components/AdminPortal.jsx`
- Modify: `admin-portal/src/components/AdminPortal.css`

**Interfaces:**
- `createArchiveApi(apiBase, fetchImpl = fetch)` provides `login(passkey)`, `beginUpload(token, metadata)`, `completeUpload(token, uploadId)`, `list(token)`, `createDownloadUrl(token, id)`, and `deleteArchive(token, id)`.
- Upload uses a two-step backend authorization + browser-to-S3 presigned POST; the completion call occurs only after S3 returns success.
- Download obtains a 5-minute URL from the authenticated API and fetches the file; do not persist tokens or URLs.
- `AdminPortal.jsx` stores the token only in React state, clears it on logout/reload, and shows an explicit “Evolution is not available in this release” state rather than invoking `EvolutionLive`.

- [ ] **Step 1: Add failing Node tests** for bearer headers on every protected request, no token storage, valid/invalid login responses, upload-policy form submission, failed S3 upload, finalize failure, expired/download errors, and delete response/errors.
- [ ] **Step 2: Run `cd admin-portal; npm test`** and confirm the new API-helper tests fail before implementation.
- [ ] **Step 3: Implement `archive-api.js` and update `archive-download.js`** to follow the API Contract, surface HTTP/API errors, and reject malformed success responses.
- [ ] **Step 4: Update the UI** to accept only PDF/PPT/PPTX under 50 MiB, explain these are file-type checks rather than malware scanning, show upload/list/download/delete progress/errors, confirm permanent delete, add a logout control that clears in-memory token state, and defer Evolution clearly.
- [ ] **Step 5: Run `cd admin-portal; npm test; npm run lint; npm run build`** with a test HTTPS `VITE_API_BASE`; confirm no token, passkey, AWS credential, or signed URL is persisted or embedded in the built static site.

### Task 7: Replace deployment guidance with cost review, approval gates, rollback, and cleanup

**Files:**
- Modify: `README_DEPLOY.md`
- Modify: `DEPLOYMENT_CHECKLIST.md`
- Modify: `.env.example`
- Modify: `.github/workflows/deploy-frontend.yml`
- Modify: `.github/workflows/deploy-to-ec2.yml`

**Interfaces:**
- Local development continues through the existing PowerShell runner and Go HTTP server; document `ADMIN_PASSKEY`/`TOKEN_HMAC_KEY` as local-only values.
- The Pages workflow receives only `VITE_API_BASE`; that value is the public API Gateway HTTPS URL with the `/api` prefix, not a secret.
- Deployment guide records a current AWS Pricing Calculator estimate for Mumbai and the stated workload, with API Gateway, Lambda, DynamoDB, S3 storage/requests, data transfer, CloudWatch logs, and any secret-management charges identified.
- The deployment process stops after SAM build/tests and cost estimate until the user explicitly approves creating/updating AWS resources.

- [ ] **Step 1: Add documentation checks** for an empty-archive/no-migration note, exact file size/types, SSM secret setup, least-privilege policy, origin/CORS setup, deployment commands, post-deploy smoke tests, rollback, cleanup, optional malware scanning, and the no-hard-cap warning.
- [ ] **Step 2: Replace stale EC2/RDS instructions** with Windows PowerShell SAM setup/validation/deployment steps and clear instructions for registering an SSM SecureString passkey/signing secret only after explicit approval.
- [ ] **Step 3: Document billing and security operations:** verify the existing $20 AWS Budget alert and notification recipients, explain that alerts do not stop charges, capture a dated calculator estimate and assumptions, identify optional paid malware scanning (for example GuardDuty Malware Protection for S3) without enabling it by default, and provide manual test-stack/bucket/table cleanup instructions.
- [ ] **Step 4: Document safe deployment and rollback:** obtain explicit user approval before `sam deploy`; deploy the reviewed stack; set the GitHub Actions `VITE_API_BASE`; publish Pages; test login/upload/list/download/delete; on failure redeploy the prior template/code revision while retaining archive data. Change the legacy EC2 workflow to manual-only and keep it disabled unless the user separately opts back in.
- [ ] **Step 5: Run `git diff --check`, all Go/frontend/SAM checks from prior tasks, and review the deployment docs** to ensure no stale instruction can accidentally launch EC2/RDS or imply a $20 hard spending cap.

## Deployment Gate

The implementation is complete only after local Go tests/vet/build, frontend tests/lint/build, SAM validation/build, and the review of all changed files pass. Before any AWS resource creation or modification:

1. Use the AWS Pricing Calculator for `ap-south-1` and the stated volume assumptions; list all potentially chargeable services and exclusions.
2. Confirm the existing monthly billing alert and its notification recipients; do not describe it as a hard stop.
3. Show the estimate and planned SAM resources to the user.
4. Ask for explicit approval to create/update the SSM parameters and SAM stack. Without that approval, stop after local validation and documentation.

No malware-scanning service is enabled by default. The documentation must explain the trade-off and cost of optional scanning; adding one later requires a separate user decision.
