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
  mounts only the dedicated UID's rootless Docker socket, and uses a host-local
  temporary workspace. The workspace path is mounted at the same absolute
  path in the container so the rootless daemon can bind-mount each candidate
  workspace.
- `docker-compose.ec2.yml` is a legacy, manual-only single-host configuration.
  It retains obsolete archive API/database and shared-volume assumptions; do
  not use it for the split deployment.

Both new Compose files require `IMAGE_TAG` to be a commit SHA and do not fall
back to `latest`. The `.env.analysis.example` and `.env.sandbox.example` files
contain placeholders only. The Compose files do not retrieve secrets from SSM;
an approved deployment process must materialize runtime values from SSM into
protected host-side environment files and certificate files. Never put real
tokens or private keys in the examples, repository, image, or browser bundle.

Before starting the sandbox broker, prepare a dedicated host user named
`ucs503-sandbox` with UID/GID `10002`, install rootless Docker for that user,
enable systemd lingering (`loginctl enable-linger ucs503-sandbox`), and enable
the user's Docker service. Verify its socket is
`/run/user/10002/docker.sock`; set `DOCKER_SOCKET_SOURCE` to that host path.
The broker runs as UID `10002` and uses the mounted socket at
`/run/docker.sock`. Never mount the rootful `/var/run/docker.sock` into the
broker. The rootless daemon and workspace must share the same host path
namespace. The workspace directory must be writable by UID `10002`, and its
path must match `SANDBOX_BROKER_WORKSPACE_DIR`.
The mounted broker private key must be readable by UID `10002` and not by
unrelated host users; mount the certificate and CA certificate read-only.
Before starting the broker, run `sudo node scripts/validate-sandbox-host.mjs`
from a repository checkout on the sandbox host (Node.js 22 required). This
read-only check verifies the account, lingering, socket ownership, workspace
access, and rootless daemon.

Both services use host networking to avoid public Compose port mappings. This
does not replace network isolation: deploy them only on private instances with
security groups that allow analysis ingress solely from the selected private
API integration and broker TCP `8082` solely from the analysis instance. The
sandbox instance must have no public IP or inbound SSH. Use SSM for
administration. The broker certificate must cover the private host name used
by `SANDBOX_SERVICE_URL`; the analysis host must trust its issuing CA via the
mounted `SANDBOX_TLS_CA_SOURCE` file. The analysis client uses only that CA
trust bundle for the broker; it does not fall back to public/system roots.
The broker must present a certificate whose chain and hostname validate
against the configured CA.

### Broker authentication and secret rotation

The analysis service sends `X-Sandbox-Token` only over the private broker
connection after TLS certificate validation. The broker compares it in
constant time. Generate at least 32 random bytes for this shared secret, store
it as the `/ucs503/analysis/sandbox-service-token` SSM SecureString, and
materialize it only into protected host-side runtime configuration for the
analysis and sandbox services. Never put it in URLs, browser code, logs,
Compose files, or container images.

Rotate the broker token during a maintenance window because the broker accepts
one configured token at a time: block broker ingress, update the SecureString
and both hosts' protected runtime configuration, restart both services, verify
the authenticated health check and a sandbox evaluation, then restore broker
ingress. For a suspected compromise, block ingress immediately and perform
this procedure before reopening it. After deployment, schedule this rotation
at least every 90 days and immediately after any suspected exposure. Systems
Manager Parameter Store retains parameter versions; record the version and
rotation date, never the value. Automated rotation is not provisioned and
requires a separately approved deployment change.

For broker CA rotation, first deploy a trust bundle containing both the
current and replacement CA to the analysis host, then replace the broker
certificate and verify TLS and authenticated evaluation. Remove the old CA
from the bundle after verification. Never make the client temporarily trust
system/public roots to bridge a CA change.

The Lambda authorizer reads the HMAC key and internal analysis API token from
SSM on each authorization request. The API Gateway integration may use the
`analysisApiToken` authorizer context value only to set the backend's internal
authentication header. Configure access logs with operational metadata only;
never log authorization headers, request/response bodies, or any
`$context.authorizer.*` value, and do not enable payload/data tracing. Add a
template test for these exclusions before creating the analysis API Gateway.

Replacing `/ucs503/admin/token-hmac-key` immediately invalidates outstanding
admin bearer tokens; users must sign in again. Coordinate that change with the
existing archive API because it uses the same signing key. Replacing
`/ucs503/analysis/api-token` must be coordinated with the analysis service's
runtime configuration; block the admin routes, update SSM and the protected
host-side value, restart the analysis service, verify a protected request,
then restore the routes. Record rotation date and parameter version, but never
record secret values. The infrastructure deployment must grant each role read
access only to the parameters it needs and include a documented emergency
revocation procedure.

Validate the configuration without starting containers:

```sh
docker compose --env-file .env.analysis.example -f docker-compose.analysis.yml config
docker compose --env-file .env.sandbox.example -f docker-compose.sandbox.yml config
node --test scripts/compose-config.test.mjs
```

These checks validate Compose syntax and policy only; they do not provision
instances, retrieve secrets, configure security groups, or deploy images.
CI runs these checks alongside Go tests and vetting. It builds the mandatory
live Docker runner under the triggering commit SHA and fails if the integration
test skips; the image publishing workflow publishes only commit-SHA tags for
the analysis and sandbox runner images.

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

## Split analysis deployment operations

The split analysis deployment is separate from the Admin Portal archive and
its existing GitHub Pages workflow, which remains unchanged. The removed
SSH-based `deploy-to-ec2.yml` workflow cannot deploy the legacy single-host
stack. `docker-compose.ec2.yml` and the old `ucs503-api` service are legacy;
the new workflow does not deploy them or use the archive stack.

CloudFormation generates the physical names for the source bucket, artifact
bucket, and lifecycle table. After a separately approved deployment, resolve
their values from the stack outputs rather than deriving names from
`ResourcePrefix`:

```powershell
aws cloudformation describe-stacks `
  --stack-name <analysis-stack-name> `
  --query "Stacks[0].Outputs[?OutputKey=='SourceBucketName' || OutputKey=='ArtifactBucketName' || OutputKey=='AnalysisInstanceId' || OutputKey=='SandboxInstanceId' || OutputKey=='SandboxLifecycleTable'].[OutputKey,OutputValue]" `
  --output table
```

The image deployment workflow reads `ArtifactBucketName`,
`AnalysisInstanceId`, `SandboxInstanceId`, and `SandboxLifecycleTable` from the
stack and supplies the lifecycle table output to analysis Compose. For manual
host-side Compose use, set `SANDBOX_LIFECYCLE_TABLE` to that output rather than
guessing the generated name. The local Compose default remains
`sandbox-lifecycle`, which is also the fixed lifecycle record ID. The
explicitly named Lambda functions and CloudWatch log groups include
stack-specific suffixes.

Ordinary stack deletion retains both S3 buckets, the lifecycle table, the
encrypted analysis data volume, and retained log groups; it does not erase
their data. The EC2 network interfaces are configured to delete with their
instances. Do not treat `sam delete` as data cleanup: retained resources may
be removed only through the separately approved cleanup workflow after its
required review and backup checks.

### Deployment approval and cost gate

The image publisher builds only commit-SHA tags for the analysis and sandbox
runner images. To deploy or roll back, dispatch `.github/workflows/deploy-analysis.yml`
with the full 40-character SHA and leave `enable_deployment` false unless the
deployment is separately approved. The deployment job runs only when both that
input is true and the repository variable `ANALYSIS_DEPLOY_ENABLED` is exactly
`true`. It then waits for approval from the protected `analysis-deployment`
GitHub Environment before downloading images, assuming an AWS role, uploading,
or invoking SSM. No repository variable, environment rule, secret, AWS role,
or AWS resource is configured by this change.

Repository owners must review the itemized estimate and approve the deployment
before configuring these Actions variables: `ANALYSIS_DEPLOY_ENABLED`,
`ANALYSIS_STACK_NAME`, `ANALYSIS_DEPLOY_ROLE_ARN`, and
`ANALYSIS_DEPLOY_REGION`. Keep account-specific values in repository
variables, not source code. The workflow reads `ArtifactBucketName`,
`AnalysisInstanceId`, `SandboxInstanceId`, and `SandboxLifecycleTable` from the
named CloudFormation stack, requiring each output exactly once and non-empty
before it uploads or sends SSM commands. Grant the AWS role
`cloudformation:DescribeStacks` for that stack, object upload to the designated
private bucket prefix, `ssm:SendCommand` for the two approved managed instances
and the `AWS-RunShellScript` and `AWS-RunRemoteScript` documents,
`ssm:GetCommandInvocation`, and `ssm:DescribeInstanceInformation` for
the SSM readiness check. It also needs `ssm:GetParameters` for the six named
external SecureStrings; it checks their names and types without requesting or
printing their values. The sandbox power-state safeguard additionally
requires `ec2:DescribeInstances` and
`ec2:StartInstances`/`ec2:StopInstances`; scope start/stop to only the approved
sandbox instance ARN. `DescribeInstances` has no resource-level IAM scope; its
allow statement is deliberately limited to the deployment region and is used
only with the resolved sandbox instance ID. The workflow never starts or stops the analysis
instance. Ensure the bucket blocks public access, uses encryption, and expires
deployment bundles under an owner-approved retention policy. The workflow
uploads to `bundles/analysis-deployments/<sha>/<bundle-sha256>/`; an existing
object is reused only when its recorded SHA-256 equals the newly built bundle.
Using the bundle checksum in the key keeps retries safe when Docker archive
metadata changes between runs. The two EC2 roles need
read-only access to that bundle prefix.
Do not create SSH keys, GHCR credentials for EC2, or long-lived AWS keys.

The GitHub runner resolves each selected SHA tag to its immutable manifest
digest, verifies GitHub provenance and keyless Sigstore signatures against
this repository's image-publishing workflow, records the Docker image IDs,
and saves the images with the selected revision's Compose files and
`scripts/validate-sandbox-host.mjs`. It then calculates a SHA-256 for the
bundle. SSM asks each private host to download the bundle from S3 and
verify that checksum before extraction or `docker load`; hosts do not pull from
GHCR or require inbound SSH. Only the sandbox host runs the bundled validator,
before its Compose command. The analysis Compose invocation receives the
stack's `SandboxLifecycleTable` output as
`SANDBOX_LIFECYCLE_TABLE`; the local Compose example default remains
`sandbox-lifecycle`.

Before changing either service, the workflow independently records each
host's active SHA release. If either deployment or the final
`sandbox_available: true` smoke test fails, it attempts to restore both
captured releases and verifies their service health plus analysis-to-broker
readiness. Any failed rollback check keeps the workflow failed for operator
follow-up. If the sandbox was stopped before deployment, the workflow starts
it, waits for SSM readiness, and restores its stopped state on success or
failure; an initially running sandbox remains running. Use the same dispatch
flow with a previously published SHA to roll forward or deliberately deploy an
older release. The workflow does not change API Gateway, archive services, or
other AWS resources.

Before any deployment, confirm the expected monthly total against the
approved budget (target about US$20/month; no more than US$25/month without
additional explicit approval). Budget notifications are alerts, not spending
caps. Review EC2 hours, EBS, S3 requests/storage/transfer, CloudWatch logs,
SSM, and snapshot/backup charges after each release and monthly. Do not
enable the deployment variable until the cost and target instances are
approved.

### SSM operations, health, and monitoring

Both hosts must be registered as Systems Manager managed nodes, have the
required SSM agent and Docker Compose runtime, and reach SSM and the private
artifact bucket without a public IP. Use SSM Run Command or Session Manager
for diagnostics; do not open SSH or add SSH secrets. The deployment workflow
stores each selected release under `/opt/ucs503/releases/<commit-sha>`. The
SSM agent downloads the bundle directly, so the hosts do not need the AWS CLI.
It keeps
the analysis Compose volume under the fixed project name
`ucs503-analysis`.

After deployment, check the SSM command result for both instances, run
`docker compose ps` on each host, and check the analysis health endpoint at
`http://127.0.0.1:8080/health`. Exercise an authenticated analysis request and
a sandbox evaluation over the private TLS connection before closing the
change. Configure monitoring before production use: alarm on EC2 status
checks, failed SSM commands, API Gateway/Lambda errors, throttles and latency,
disk exhaustion, failed analysis health checks, and sandbox start/readiness
failures or stale work leases. Set an explicit CloudWatch log retention period
and alert recipients; never include bearer tokens, service tokens, private
keys, request bodies, or authorizer context in logs. The Compose files rotate
local container logs, but do not create CloudWatch alarms, centralized log
retention, or host metrics collection. Confirm the exact alarm/log cost.

### Output backup, restore, and recovery

Analysis uploads bind to `/var/lib/ucs503-analysis/uploads` on a separate,
encrypted gp3 data volume attached to the analysis instance. The infrastructure
template configures daily EBS snapshots and seven-snapshot retention through
Data Lifecycle Manager. Snapshot charges must be included in the itemized
estimate. The template does not automate restore or verify snapshots; test a
restore before production and retain the previous host/disk until expected
outputs are confirmed.

For an approved manual restore, provision a replacement analysis host and an
encrypted data volume in the same Availability Zone, restore the selected EBS
snapshot to that volume, attach and mount it at `/var/lib/ucs503-analysis`,
and verify ownership for the container UID/GID `10002`. Start the service only
after confirming `/var/lib/ucs503-analysis/uploads` contains the expected
outputs. Record snapshot IDs and test dates, never credentials. Do not claim
one-hour recovery until a replacement-host restore has been timed.

For sandbox lease recovery, do not stop the host while an evaluation or
workspace lease is active. Inspect analysis and broker logs through SSM after
a crash; allow the lifecycle controller/watchdog to reconcile stale leases,
confirm no runner container remains active, then restart the broker or allow
the idle-stop controller to stop the sandbox. Escalate rather than manually
clearing a lease whose owner or job state is uncertain. Verify the broker's
authenticated health and a test evaluation after recovery. Lease recovery
alarms and automated reconciliation must be enabled and tested as part of
host operations; this workflow does not change their AWS settings.

### Token and certificate rotation

Use the existing **Broker authentication and secret rotation** procedures
above for `SANDBOX_SERVICE_TOKEN`, the broker CA bundle/certificate,
`ANALYSIS_API_TOKEN`, and the archive signing key. Fetch or update values only
through the approved SSM/host process. Record rotation date and SSM parameter
version, never values. Roll out CA trust overlap before replacing the broker
certificate; coordinate shared-token changes across both hosts and verify
authenticated health/evaluation before restoring ingress.

### Retained analysis resource cleanup

The cleanup workflow is a separate, manual path for resources retained after
the analysis CloudFormation stack has been deleted. Delete the analysis stack
first; do not use cleanup while the stack exists: the read-only inventory
refuses every state other than a missing stack or `DELETE_COMPLETE`. Before
dispatching, configure the repository
variables `ANALYSIS_STACK_NAME`, `ANALYSIS_RESOURCE_PREFIX`, and
`ANALYSIS_CLEANUP_REGION`, plus separate
`ANALYSIS_CLEANUP_READ_ROLE_ARN` and `ANALYSIS_CLEANUP_WRITE_ROLE_ARN` OIDC
role variables. Set `ANALYSIS_CLEANUP_ENABLED` to the exact string `true` only
for an approved cleanup window. Configure required reviewers for the
protected `analysis-cleanup` GitHub Environment in repository settings; that
protection is an operator prerequisite and is not created by this workflow.

Dispatch **Cleanup retained analysis resources** with `enable_cleanup` left
false for inventory-only review. The inventory job uses its read-only role and
adds the resource list to the workflow run summary before the destructive job
can request environment approval. Review the whole list and the uploaded
`analysis-cleanup-plan` artifact. To proceed, dispatch again with
`enable_cleanup` set to true and type the exact `ANALYSIS_RESOURCE_PREFIX` in
`confirm_prefix`. Cleanup remains disabled unless that boolean, the repository
enable variable, and exact prefix confirmation all match. The protected
environment approval is an additional independent gate.

The allowlist is restricted to S3 buckets, DynamoDB tables, EBS data volumes,
and EBS snapshots carrying the exact `AnalysisStack=<resource prefix>` tag,
plus only these six exact legacy log-group stems from the analysis template.
The exact unsuffixed names remain eligible even when the deleted stack ID is
unavailable. A suffixed name is eligible only when its 12-character lowercase
hexadecimal suffix matches the exact deleted stack ID:

- `/aws/ssm/sessions/<prefix>` or `/aws/ssm/sessions/<prefix>-<suffix>`
- `/aws/lambda/<prefix>-analysis-proxy` or `/aws/lambda/<prefix>-analysis-proxy-<suffix>`
- `/aws/lambda/<prefix>-analysis-authorizer` or `/aws/lambda/<prefix>-analysis-authorizer-<suffix>`
- `/aws/lambda/<prefix>-source-stager` or `/aws/lambda/<prefix>-source-stager-<suffix>`
- `/aws/lambda/<prefix>-sandbox-watchdog` or `/aws/lambda/<prefix>-sandbox-watchdog-<suffix>`
- `/aws/apigateway/<prefix>-analysis` or `/aws/apigateway/<prefix>-analysis-<suffix>`

The script rejects missing or changed ownership tags, changed inventory,
unknown log groups, attached volumes, incomplete S3 version pagination, and
failed AWS operations. It rechecks ownership immediately before each delete.
Applying is irreversible: it deletes all S3 object versions and delete
markers (including current objects) from each eligible bucket before deleting
the bucket; drops the eligible DynamoDB tables and all their items; deletes
unattached tagged EBS data volumes and tagged snapshots; and deletes the six
eligible log groups. Check backups and retention obligations first. This path never
terminates EC2 instances, deletes AMIs, or changes/deletes the archive stack.

For future least-privilege OIDC configuration, scope the inventory role to
`cloudformation:DescribeStacks` for the named analysis stack;
`s3:ListAllMyBuckets` and `s3:GetBucketTagging`; DynamoDB
`ListTables`, `DescribeTable`, and `ListTagsOfResource`; EC2
`DescribeVolumes`/`DescribeSnapshots`; and CloudWatch Logs
`DescribeLogGroups`. Scope the separate cleanup role's matching read
permissions plus `s3:ListBucketVersions`, `s3:DeleteObjectVersion`, and
`s3:DeleteBucket` only to tagged analysis buckets; `dynamodb:DeleteTable`
only to tagged lifecycle tables; `ec2:DeleteVolume`/`ec2:DeleteSnapshot`
only to the exact `AnalysisStack` tag; and `logs:DeleteLogGroup` only to the
six allowlisted log-group name patterns. EC2 describe/list APIs that do not
support resource scoping may require `Resource: "*"`. Restrict both OIDC
trust policies to this repository and the intended workflow/job identities.
Do not grant EC2 instance termination, AMI deletion, CloudFormation stack
deletion, or archive-resource permissions.
