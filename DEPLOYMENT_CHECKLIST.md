# Serverless Admin Archive Deployment Checklist

## Before AWS setup

- [ ] Confirm this is the intended AWS account and `ap-south-1` (Mumbai) region.
- [ ] Review the low-volume cost range and verify it in the AWS Pricing
  Calculator with current regional rates.
- [ ] Confirm the existing $20 monthly budget alert and notification recipients.
  It is an alert, not a spending cap.
- [ ] Obtain explicit approval before creating SSM parameters or deploying the
  SAM stack. No AWS resources should be created before approval.
- [ ] Confirm the GitHub Pages origin is exactly `https://zylus08.github.io`.
  Do not include the project path in CORS `Origin`.

## Local preflight

- [ ] Run `sam validate --lint --template-file template.yaml`.
- [ ] Run `sam build --template-file template.yaml`.
- [ ] In `api/`, run `go test ./...`, `go vet ./...`, and `go build ./...`.
- [ ] In `admin-portal/`, run `npm test`, `npm run lint`, and
  `npm run build` with a test HTTPS `VITE_API_BASE`.
- [ ] Verify the built static frontend contains no passkey, token-signing
  secret, AWS credential, bearer token, or presigned URL.
- [ ] Confirm `.env.example` contains only local development guidance and no
  real credential values.

## After approval: parameters and stack

- [ ] In Parameter Store in `ap-south-1`, create standard-tier SecureString
  parameters `/ucs503/admin/passkey` and `/ucs503/admin/token-hmac-key` using
  the AWS-managed SSM key.
- [ ] Use independent, random production values; never use the local passkey.
- [ ] Deploy `template.yaml` with stack name `ucs503-admin-archive`,
  `FrontendOrigin=https://zylus08.github.io`, and the exact SSM parameter names.
- [ ] Review the SAM change set before accepting resource creation.
- [ ] Save the `ApiBaseUrl`, archive bucket name, and DynamoDB table name
  outputs in an approved secure operations record; never save secret values.
- [ ] Set the repository Actions variable `VITE_API_BASE` to the public API
  HTTPS URL ending in `/api`. It is not a secret.
- [ ] Leave `EC2_DEPLOY_ENABLED` unset/false. The legacy EC2 workflow should
  remain unused.

## Smoke tests

- [ ] Check `GET https://<api-id>.execute-api.ap-south-1.amazonaws.com/status`.
- [ ] Verify the GitHub Pages portal can log in with the configured passkey and
  that an incorrect passkey is rejected.
- [ ] Upload one disposable PDF, PPT, or PPTX smaller than 50 MiB.
- [ ] Confirm it appears in the archive list, can be downloaded, and can be
  deleted after the UI confirmation.
- [ ] Confirm no files other than supported PDF/PPT/PPTX types are accepted.
  The service does not scan for malware.
- [ ] Confirm the S3 bucket blocks public access, and unauthenticated requests
  cannot read archive objects; confirm the bucket policy denies insecure
  (non-HTTPS) transport.
- [ ] Confirm failed upload finalization is not reported as a successful
  archive entry.

## Rollback and cleanup

- [ ] For a code rollback, deploy the previous known-good commit/template to
  the existing stack. Do not delete the stack to roll back.
- [ ] Keep S3 and DynamoDB data during rollback. The SAM template retains the
  bucket, table, and log group to reduce accidental data loss.
- [ ] For full teardown, separately approve the destructive cleanup, export
  required archive data, empty/delete the S3 archive bucket, delete the
  retained DynamoDB table and CloudWatch log group, remove both SSM parameters,
  and remove the SAM artifact bucket if it is no longer needed.
- [ ] Recheck AWS Budgets after deployment. Notifications do not automatically
  stop charges.
