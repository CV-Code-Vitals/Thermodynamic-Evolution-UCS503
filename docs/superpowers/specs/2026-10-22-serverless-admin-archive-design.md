# Design: Domain-Free Serverless Admin Archive

## Intent

Keep the public website and Admin Portal on GitHub Pages while moving authenticated archive operations to AWS. The administrator must be able to log in, upload files, list stored deliverables, and download them without buying a domain or relying on browser third-party cookies.

## Why

The current Go API is a long-running HTTP server. It uses PostgreSQL or local JSON for metadata, accepts multipart uploads through the API, and authenticates with an HTTP-only session cookie. The chosen target has no custom domain, uses GitHub Pages for the portal, and prioritizes a $20/month out-of-pocket target. A serverless design avoids idle EC2/RDS charges and uses an AWS-generated HTTPS API endpoint.

## Goals

- Keep the main website and Admin Portal UI on GitHub Pages.
- Provide an AWS-hosted HTTPS API using an AWS-generated hostname; do not require a custom domain.
- Make login work cross-origin without third-party cookies.
- Store uploaded files privately in S3 and archive metadata durably in DynamoDB.
- Support files up to 50 MiB without sending their bytes through API Gateway.
- Fail visibly on storage/API errors; never report browser-only storage as durable success.
- Keep AWS resources and permissions reproducible through infrastructure as code.

## Non-goals for the first release

- Moving the GitHub Pages website or Admin Portal UI to AWS.
- Using Cloudflare DNS, proxying, or WAF; these need a domain the user controls.
- Deploying EC2, RDS, or an Application Load Balancer.
- Migrating the separate repository-analysis/Evolution service. The Evolution tab should be clearly marked unavailable or deferred rather than appearing to work.
- Guaranteeing a hard $20 AWS bill cap. Budgets notify; they do not stop all billable usage.

## Decisions

### Runtime and API

- Use AWS Lambda with the existing Go language and an API Gateway HTTP API.
- Use the API Gateway `execute-api` HTTPS endpoint; configure the GitHub Pages origin explicitly for CORS.
- Use AWS SAM as the deployment template and workflow so IAM, API routes, Lambda configuration, S3, and DynamoDB are created consistently.
- Retain a local development/test path that does not require AWS credentials.

### Authentication

- Replace browser session cookies with a short-lived HMAC-signed bearer token returned by the login endpoint.
- Keep the token in browser memory only; do not put it in localStorage or sessionStorage. A page reload requires signing in again, matching the current portal’s in-memory authenticated state.
- Validate the passkey and token signature/expiry in the API. Store signing and passkey secrets outside source control and inject them securely at deployment.
- Send the token in the `Authorization` header for protected API calls. Configure API Gateway CORS for only the GitHub Pages origin and required headers/methods.

### File storage and metadata

- Keep S3 Block Public Access enabled and use the Lambda execution role for object operations.
- Upload file bytes directly from the browser to S3 using a short-lived, authenticated, size-limited presigned upload request. Finalize metadata only after confirming the object exists and is within the 50 MiB limit.
- For downloads, require the bearer token at the API and return a short-lived presigned S3 download URL. A URL is a temporary bearer credential; it must not be stored as a permanent public link.
- Store archive metadata in a DynamoDB on-demand table. Use a queryable index for newest-first archive listing rather than a full table scan.
- Accept PDF, PPT, and PPTX files only, and validate type and size in both the browser and API before issuing an upload authorization.
- Provide an authenticated delete action that removes both the S3 object and its DynamoDB metadata; require confirmation in the portal. A failed object deletion must remain visible and must not be reported as successful.

### Cost controls and deployment

- Keep the existing $20 monthly cost alert, but treat it as an alert, not a spending limit.
- Before deploying billable resources, estimate API invocations, stored GB, and monthly download GB in AWS Pricing Calculator and show the estimate to the user. Do not deploy if the estimate is not approved.
- Bound Lambda concurrency and API request rates where they meaningfully limit runaway compute/request costs. Explain that these controls do not cap S3 storage or data transfer charges.
- Do not add a CloudFront distribution unless a tested requirement needs it; the API Gateway HTTPS endpoint is sufficient for the API.
- Use AWS SAM deploy from the Windows development environment after tests pass; do not manually create duplicate resources in the Console.

## User choices captured

- No custom domain.
- Public website and Admin Portal remain on GitHub Pages.
- Serverless AWS API and usage-based database/storage chosen over EC2/RDS.
- Login uses a short-lived token held only in memory.
- Initial scope should be the smallest reliable release; the Evolution/analysis tab is deferred.
- AWS account has $80 in credits expiring in 2027; user wants out-of-pocket spending not to exceed $20/month.
- Start the AWS archive empty; keep any existing local/PostgreSQL archive separate (no migration in this release).
- Plan cost estimates around low usage: 10 uploads/month at about 10 MB each, approximately 2 GB initial retained storage, and 1 GB monthly downloads. Files are PDFs, PPTs, or PPTXs only.
- Include authenticated archive deletion; the portal must confirm before deleting.

## Builder decisions

- Select a short token lifetime and presigned URL expiry appropriate for the portal; document the values and test expiry behavior.
- Choose the simplest compatible Go Lambda adapter and define route/error response mapping without changing local API semantics unnecessarily.
- Determine how to represent uploaded-but-not-finalized objects and safely clean up abandoned uploads without an aggressive lifecycle rule that deletes archived files.
- Configure Lambda memory/timeout, API throttling, and SAM packaging based on measured build size and tests.

## Risks and constraints

| Risk | Mitigation |
|---|---|
| A monthly AWS budget is not a hard cap, even with credits. | Estimate first, show exclusions/assumptions, use alerts and bounded request/concurrency settings, and obtain user approval before deployment. |
| API Gateway has request payload limits. | Transfer file bytes directly to S3 with a short-lived upload authorization; keep API calls metadata-only. |
| Presigned URLs can be used by anyone who obtains them until expiry. | Issue them only after authentication, keep expiry short, limit object key/content type/size, and never log full signed URLs. |
| Browser code cannot safely hold long-lived secrets. | Keep short-lived bearer tokens in memory only; never embed the admin passkey or AWS credentials in the static build. |
| The Evolution tab depends on a separate analysis backend. | Clearly defer/disable the feature for this release; do not let it silently fail or use a fake success fallback. |
| AWS service prices and credit eligibility vary by region and usage. | Use the current AWS Pricing Calculator with Mumbai selected and the user's storage/egress assumptions before provisioning. |

## Acceptance

- [ ] The deployed API uses API Gateway HTTPS and the configured GitHub Pages origin.
- [ ] Invalid passkeys are rejected; valid login returns a short-lived signed token.
- [ ] Protected routes reject missing, malformed, and expired tokens.
- [ ] Uploads up to 50 MiB go directly to the private S3 bucket; larger or unsupported files are rejected.
- [ ] Archive metadata is persisted in DynamoDB only after a successful object upload.
- [ ] Archive listing is newest-first and survives Lambda cold starts.
- [ ] Authenticated downloads return a temporary URL to a private S3 object; unauthenticated requests cannot obtain it.
- [ ] Frontend upload/list/download errors are visible, with no fake local success path.
- [ ] Local development and automated tests run without deployed AWS resources.
- [ ] The current AWS Pricing Calculator estimate is shared and explicitly approved before deployment.
