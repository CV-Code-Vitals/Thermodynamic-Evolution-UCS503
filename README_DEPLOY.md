# Local Development and AWS Deployment

The Admin Portal archive uses GitHub Pages for the static UI and, when you
approve an AWS deployment, API Gateway HTTP API, a Go Lambda, private S3, and
DynamoDB in `ap-south-1` (Mumbai). The archive starts empty. The deployment
does not create EC2, RDS, an Application Load Balancer, or a custom domain.

**Local tests and SAM validation/build do not create AWS resources. Creating
the SSM parameters and deploying the stack both create AWS resources; do not
do either until you have reviewed the cost and explicitly approved.**

## Run locally on Windows

Requirements: Go 1.25, Node.js 22, and npm. From the repository root:

```powershell
$env:ADMIN_PASSKEY = 'choose-a-local-only-passkey'
.\scripts\run-local.ps1
```

The script starts the API, Vite frontend, and local dashboard in separate
PowerShell windows. The Admin Portal is at `http://localhost:5173`; the API
health check is `http://localhost:8080/status`. The development API stores
files and metadata on the local machine and does not require AWS credentials.
Set `ADMIN_PASSKEY` explicitly before starting the API or local scripts. There is no built-in admin fallback; a missing passkey aborts startup.

To run the services manually, open two terminals:

```powershell
# Terminal 1
Set-Location api
$env:APP_ENV = 'development'
$env:ADMIN_PASSKEY = 'choose-a-local-only-passkey'
$env:CORS_ALLOWED_ORIGINS = 'http://localhost:5173,http://127.0.0.1:5173,http://localhost:8000,http://127.0.0.1:8000'
go run .
```

```powershell
# Terminal 2
Set-Location admin-portal
npm ci
npm run dev
```

For local static-site builds, the Windows runner configures the `/api` proxy
automatically. A local `TOKEN_HMAC_KEY` is optional; when omitted the API
generates an ephemeral key at startup, invalidating sessions when it restarts.
Do not put production credentials in `.env`, frontend variables, or source
control.

The separate `go-backend/` analysis service is not the Admin Portal API and is
not part of this serverless archive deployment.

## Separate analysis and sandbox hosts

The deployment Compose files are split by trust boundary:

- `docker-compose.analysis.yml` runs only the Go analysis API and Rust engine.
  It has no Docker socket and stores uploads in an analysis-host-local Docker
  volume.
- `docker-compose.sandbox.yml` runs only the authenticated sandbox broker.
  It runs in the host network namespace without a published Compose port,
  mounts the local Docker socket, and uses a host-local temporary workspace.
  The workspace path is mounted at the same absolute path in the container so
  the host Docker daemon can bind-mount each candidate workspace.
- `docker-compose.ec2.yml` is a legacy, manual-only single-host configuration.
  It retains obsolete archive API/database and shared-volume assumptions; do
  not use it for the split deployment.

Both new Compose files require `IMAGE_TAG` to be a commit SHA and do not fall
back to `latest`. The `.env.analysis.example` and `.env.sandbox.example` files
contain placeholders only. The Compose files do not retrieve secrets from SSM;
an approved deployment process must materialize runtime values from SSM into
protected host-side environment files and certificate files. Never put real
tokens or private keys in the examples, repository, image, or browser bundle.

Before starting the sandbox broker, prepare the host-local workspace directory
with ownership matching the image's unprivileged app user (UID/GID `10002`),
and set `DOCKER_GID` to the numeric group ID that owns `/var/run/docker.sock`.
The broker's container user needs that supplementary group to use Docker; this
is still highly privileged access, which is why the broker and socket must
remain on a dedicated sandbox host. The workspace directory must be writable
by UID `10002`, and its path must match `SANDBOX_BROKER_WORKSPACE_DIR`.
The mounted broker private key must be readable by UID `10002` and not by
unrelated host users; mount the certificate and CA certificate read-only.

Both services use host networking to avoid public Compose port mappings. This
does not replace network isolation: deploy them only on private instances with
security groups that allow analysis ingress solely from the selected private
API integration and broker TCP `8082` solely from the analysis instance. The
sandbox instance must have no public IP or inbound SSH. Use SSM for
administration. The broker certificate must cover the private host name used
by `SANDBOX_SERVICE_URL`; the analysis host must trust its issuing CA via the
mounted `SANDBOX_TLS_CA_SOURCE` file.

Validate the configuration without starting containers:

```sh
docker compose --env-file .env.analysis.example -f docker-compose.analysis.yml config
docker compose --env-file .env.sandbox.example -f docker-compose.sandbox.yml config
node --test scripts/compose-config.test.mjs
```

These checks validate Compose syntax and policy only; they do not provision
instances, retrieve secrets, configure security groups, or deploy images.

## Archive behavior and limits

- Accepts PDF, legacy PowerPoint (`.ppt`), and PowerPoint Open XML (`.pptx`)
  files up to 50 MiB. These checks validate type and size; they are not malware
  scanning.
- Uploads go through a short-lived, signed S3 form and are finalized by the API
  only after the object upload succeeds and the backend validates its content.
- Listing is limited to available records. Downloads use short-lived signed
  URLs; the portal keeps its bearer token in memory only.
- Delete is permanent in this initial configuration and idempotent: deleting
  an already-absent archive succeeds, including retries after a lost response.
  S3 versioning and malware scanning are not enabled.
- Pending S3 uploads expire after one day and pending DynamoDB records after
  two days. Available archive objects and metadata have no automatic expiry.

## Pre-deployment cost review

Workload assumption for a planning estimate: 10 uploads/month averaging 10 MB,
2 GB already stored at the start, and 1 GB/month downloaded. This implies about
100 MB of new data each month, with storage growing by that amount if no files
are deleted.

For that low-volume workload, budget **approximately US$0.15–$1/month** before
taxes, credits, optional scanning, and any usage shared with other workloads.
This is a planning range, not a saved AWS Pricing Calculator quote. It assumes
no EC2/RDS/NAT/custom-domain/paid malware scanning, low API and DynamoDB
request counts, short-retention low-volume logs, and no unusual retry or abuse
traffic. Main cost drivers are S3 Standard storage, requests and data transfer,
API Gateway requests, Lambda execution, DynamoDB on-demand reads/writes, and
CloudWatch logs. Actual rates vary by region and account eligibility; confirm
the final configuration in the [AWS Pricing Calculator](https://calculator.aws/)
for Asia Pacific (Mumbai) immediately before deployment.

AWS Standard Parameter Store parameters use SecureString with the
AWS-managed SSM key. Confirm the current Parameter Store and KMS pricing in
the calculator as part of the final estimate. The estimate does not assume
free-tier eligibility or apply the previously mentioned account credits.
Earlier project context mentioned about US$80 in account credits; verify the
remaining balance, expiry, and service eligibility directly in Billing before
including any credits in a decision.
Verify the existing **$20 monthly AWS Budget alert** and its recipients before
creating resources. A budget alert sends notifications; it is **not** a hard
spending cap and does not stop resources or charges.

Pricing references:

- [API Gateway pricing](https://aws.amazon.com/api-gateway/pricing/)
- [Lambda pricing](https://aws.amazon.com/lambda/pricing/)
- [S3 pricing](https://aws.amazon.com/s3/pricing/)
- [DynamoDB pricing](https://aws.amazon.com/dynamodb/pricing/)
- [CloudWatch pricing](https://aws.amazon.com/cloudwatch/pricing/)
- [Systems Manager pricing](https://aws.amazon.com/systems-manager/pricing/)
- [AWS Pricing Calculator](https://calculator.aws/)

## AWS setup (only after you approve)

1. Install AWS CLI and AWS SAM CLI. Authenticate with an authorized IAM Identity
   Center profile; do not create or store long-lived access keys in this
   repository. Check the selected account and region:

   ```powershell
   aws configure sso --profile your-profile
   aws sso login --profile your-profile
   aws sts get-caller-identity --profile your-profile
   aws configure get region --profile your-profile
   ```

2. From the repository root, validate and build locally:

   ```powershell
   sam validate --lint --template-file template.yaml
   sam build --template-file template.yaml
   ```

   On Windows, `sam build` needs GNU Make available on `PATH`; alternatively,
   start Docker Desktop and use `sam build --use-container`. These commands do
   not create AWS resources. CI performs `sam validate --lint` and `sam build`
   on every relevant Go/template change.

3. Review the current calculator estimate, the exact template resources, and
   the account's budget alert. **Stop and ask for explicit approval before
   creating the parameters or stack.**

4. Only after approval, in the AWS Console select **Asia Pacific (Mumbai) /
   `ap-south-1`**, then open **Systems Manager → Parameter Store**. Create two
   **Standard** **SecureString** parameters encrypted with the AWS-managed
   `alias/aws/ssm` key:

   - `/ucs503/admin/passkey` — a unique, strong admin passkey.
   - `/ucs503/admin/token-hmac-key` — an independent, randomly generated
     signing secret with at least 32 random bytes.

   Never reuse the local development passkey. Do not paste either value into
   chat, GitHub Actions, `.env`, or the Vite frontend. The template gives the
   Lambda role `GetParameter` access only to these configured parameter names;
   its separate `DescribeParameters` permission can inspect metadata but not
   read other parameter values.

5. Only after approval and successful build, deploy the reviewed template:

   ```powershell
   sam deploy --guided `
     --stack-name ucs503-admin-archive `
     --region ap-south-1 `
     --profile your-profile `
     --capabilities CAPABILITY_IAM `
     --resolve-s3 `
     --parameter-overrides `
       FrontendOrigin=https://zylus08.github.io `
       AdminPasskeyParameterName=/ucs503/admin/passkey `
       TokenHMACKeyParameterName=/ucs503/admin/token-hmac-key
   ```

   Review the change set and costs before confirming. SAM may create a separate
   S3 artifact bucket for deployment packages. The template creates the API,
   Lambda, execution role, private archive bucket (public access blocked and
   insecure transport denied), DynamoDB table/index, and retained CloudWatch
   log group.

6. Copy the `ApiBaseUrl` output from the completed stack. In the GitHub
   repository, add the Actions **variable** `VITE_API_BASE` with that public
   HTTPS URL, including `/api` (for example,
   `https://<api-id>.execute-api.ap-south-1.amazonaws.com/api`). It is a public
   endpoint, not a secret. The allowed browser origin is the exact origin
   `https://zylus08.github.io`; the project path is not part of an Origin
   header. Change the SAM `FrontendOrigin` parameter if the Pages hostname
   changes.

7. Publish the frontend using the GitHub Pages workflow. After deployment,
   verify `GET https://<api-id>.execute-api.ap-south-1.amazonaws.com/status`
   (the API health route is `/status`),
   then test login, upload, list, download, and delete with a disposable file.
   Verify the archive bucket is private and that an unauthenticated S3 request
   cannot read an object. Do not paste presigned URLs or secrets into logs or
   tickets.

## Rollback and cleanup

- To roll back application code, check out the last known-good commit and run
  `sam build` then `sam deploy` against the same stack and parameters. Do not
  delete the stack as a rollback; the archive data is retained.
- Keep the previous frontend revision and restore it by rerunning the Pages
  workflow from that revision. Confirm `VITE_API_BASE` still matches the stack.
- The template intentionally retains the archive S3 bucket, DynamoDB table,
  and log group when the stack is deleted or replaced. `sam delete` therefore
  does **not** erase all data or stop every retained resource. Before cleanup,
  download any needed files and metadata. Then, after a separate explicit
  approval, run `sam delete --stack-name ucs503-admin-archive --region
  ap-south-1 --profile your-profile` to remove the stack-managed API, Lambda,
  and role. Empty and delete the retained archive bucket, delete the retained
  DynamoDB table and log group, remove both SSM parameters, and remove the SAM
  artifact bucket if it is no longer needed. This is destructive and cannot
  be undone without backups.

The legacy EC2 deployment workflow is manual-only and requires both a
deliberate workflow input and the `EC2_DEPLOY_ENABLED` repository variable.
Leave that variable unset/false; it is unrelated to this serverless plan.
