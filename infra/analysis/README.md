# Whole-project analysis hosting

This is a separate SAM/CloudFormation stack for the analysis API and isolated
Evolution sandbox. It does not import or change the working archive stack. It
keeps the website on GitHub Pages, exposes anonymous Graph scans and
authenticated Evolution through API Gateway REST API/Lambda response
streaming with an exact GitHub Pages CORS origin, and uses two private EC2
instances: an always-on analysis host and an on-demand broker host.

## Safety and approval gate

The template has not been deployed. Local builds, tests, and template
validation do not create AWS resources. Do not create SSM parameters, configure
GitHub deployment variables/environments, run `sam deploy`, or send SSM
commands until the itemized estimate, deployment plan, and final explicit
deployment authorization have each been reviewed and approved.

The design target is US$20/month and the hard limit is US$25/month without
additional approval. No live `ap-south-1` pricing or account-specific network
validation has been performed. The five interface endpoints add hourly and
data-processing charges, so the current design may exceed that budget; verify
their current regional price along with the 100-GiB gp3 data volume, seven
retained daily snapshots, EC2, API Gateway/Lambda streaming, CloudWatch,
S3, and transfer charges before approving. The target is not a validated
estimate.

## Stack contracts

- Stack name: `ucs503-analysis-v1`; default region: `ap-south-1`.
- Both instances have private ENIs and no public IPv4 address or SSH ingress.
  The proxy reaches the analysis private IP; broker ingress is allowed only
  from the analysis security group.
- S3 and DynamoDB Gateway Endpoints provide private data access. Interface
  endpoints provide SSM, SSM Messages, EC2 Messages, EC2, and Lambda control
  plane access; endpoint policies and security groups restrict them to the
  stack's workload roles and groups.
- The VPC and subnets remain dual-stack, but public IPv6 egress is disabled.
  AWS SDK clients use the standard regional DNS names so Private DNS routes
  API calls through the interface endpoints. Use a current SSM Agent and
  validate SSM, endpoint, and Gateway Endpoint access from the deployed
  private subnets before treating the stack as ready.
- Staged source objects expire after one day. Image deployment bundles expire
  after 30 days under `bundles/`.
- Analysis outputs bind to `/var/lib/ucs503-analysis/uploads` on a separate,
  encrypted gp3 volume. DLM creates daily snapshots and retains seven. Snapshot
  restore is a manual, unautomated process and must be tested before production.
- The `sandbox-lifecycle` DynamoDB record coordinates leases. The analysis
  service starts the selected sandbox instance, waits for EC2 and authenticated
  broker readiness, and releases leases after Evolution. The scheduled
  watchdog expires stale leases and coordinates idle stop recovery.
- CloudFormation generates the physical names for both retained S3 buckets
  and the retained lifecycle DynamoDB table. Read `SourceBucketName`,
  `ArtifactBucketName`, `AnalysisInstanceId`, `SandboxInstanceId`, and
  `SandboxLifecycleTable` from the stack outputs. The image deployment
  workflow resolves and validates the latter four outputs from the configured
  stack before uploading a bundle or sending SSM commands. It sets the
  analysis Compose invocation's `SANDBOX_LIFECYCLE_TABLE` from the generated
  output; the local Compose default remains `sandbox-lifecycle`, and the
  lifecycle record ID remains `sandbox-lifecycle`.
- Explicit Lambda function and CloudWatch log-group names include a
  stack-specific suffix. The encrypted analysis data volume, buckets, table,
  and retained log groups survive ordinary stack deletion; only the instance
  network interfaces are deleted with their instances. Removing retained
  resources requires the separately approved cleanup workflow.
- The proxy may read only the analysis API token; the authorizer alone may
  read the shared archive signing key. The sandbox role cannot read either.
- `FrontendOrigin` defaults to `https://zylus08.github.io`; update it only if
  the GitHub Pages origin changes. API methods, proxy responses, and API
  Gateway errors use that exact CORS origin rather than reflecting requests.
- The image deployment workflow uses OIDC and SSM. EC2 hosts load a selected
  SHA-addressed bundle from the private artifact bucket and never pull GHCR
  images directly.

## Build and validate locally

From the repository root:

```powershell
node --test tests/analysis-template.test.mjs tests/analysis-workflow.test.mjs tests/graph-visualizer-config.test.mjs
sam validate --lint --template-file infra/analysis/template.yaml
docker compose --env-file .env.analysis.example -f docker-compose.analysis.yml config --quiet
docker compose --env-file .env.sandbox.example -f docker-compose.sandbox.yml config --quiet
```

The example parameter file is documentation only. Replace its AMI placeholders
and managed prefix-list ID placeholders with values verified for the deployment
region only after the separate approval gate; do not commit account-specific
IDs, real secrets, or generated host configuration. Find the two AWS-managed
prefix list IDs with the read-only command below, replacing the region as
needed:

```powershell
aws ec2 describe-managed-prefix-lists `
  --region ap-south-1 `
  --filters Name=owner-id,Values=AWS `
            Name=prefix-list-name,Values=com.amazonaws.ap-south-1.s3,com.amazonaws.ap-south-1.dynamodb `
  --query "PrefixLists[].{Name:PrefixListName,Id:PrefixListId}" `
  --output table
```

## Deployment prerequisites and checks

Before a separately authorized deployment, prepare and review:

1. Supported analysis and sandbox AMIs with the SSM Agent, Docker Engine,
   Docker Compose plugin, Node.js 22 for the read-only host validator, and host
   bootstrap configuration. The SSM agent downloads deployment bundles directly.
   The
   analysis host must not have access to a Docker socket. Prepare the sandbox
   host with rootless Docker for UID/GID `10002`, enable systemd linger for
   that dedicated user, and verify the daemon socket is
   `/run/user/10002/docker.sock`; the broker mounts that socket inside its
   container at `/run/docker.sock`. The daemon and broker must see the same
   absolute sandbox workspace path. Run
   `sudo node scripts/validate-sandbox-host.mjs` on the sandbox host to check
   the UID/GID, lingering, socket, workspace access, and rootless daemon.
2. A protected `analysis-deployment` GitHub Environment with required
   reviewers and repository variables for `ANALYSIS_STACK_NAME`,
   `ANALYSIS_DEPLOY_ROLE_ARN`, and `ANALYSIS_DEPLOY_REGION`. Restrict the OIDC
   role to this repository and environment. Grant it
   `cloudformation:DescribeStacks` for the analysis stack,
   `s3:PutObject` to the deployment bundle prefix,
   `ssm:SendCommand` for the two managed nodes and the `AWS-RunShellScript` and
   `AWS-RunRemoteScript` documents, `ssm:GetCommandInvocation`, and
   `ssm:DescribeInstanceInformation` for the sandbox readiness check, plus
   `ssm:GetParameters` for the six configured external SecureStrings. The
   preflight reads names and parameter types only, never secret values. The
   workflow resolves `ArtifactBucketName`, `AnalysisInstanceId`,
   `SandboxInstanceId`, and `SandboxLifecycleTable` from CloudFormation
   outputs and fails closed if any expected output is missing, duplicated, or
   empty. The workflow's power-state safeguard also requires
   `ec2:DescribeInstances` and `ec2:StartInstances`/`ec2:StopInstances`; scope
   start/stop to only the sandbox instance ARN. AWS does not support
   resource-level scoping for `DescribeInstances`, so the template limits that
   read permission to its Region and the workflow sends a single resolved
   sandbox instance ID. It never starts or stops the
   analysis instance. Do not configure these permissions as part of local
   validation.
3. The SecureString parameters named in the template, including the shared
   signing key and separate API and broker tokens, plus the broker TLS
   certificate, key, and CA. Generate and install them through an approved
   secret-handling process; never store their values in the repository.
4. Protected `/etc/ucs503/analysis.env` and `/etc/ucs503/sandbox.env` files,
   the broker certificate/CA files, and correctly permissioned host paths.
   The analysis env file must configure `ANALYSIS_API_TOKEN`, `SOURCE_BUCKET`,
   `SANDBOX_SERVICE_URL`, `SANDBOX_SERVICE_TOKEN`, `SANDBOX_INSTANCE_ID`,
   `SANDBOX_TLS_CA_SOURCE`, and `ANALYSIS_UPLOADS_HOST_DIR`. The deployment
   workflow supplies `SANDBOX_LIFECYCLE_TABLE` from the stack output to
   analysis Compose; manual host-side Compose use must set it to that same
   generated table name. The sandbox env file must configure the broker token,
   certificate/key sources, workspace directory, and
   `DOCKER_SOCKET_SOURCE=/run/user/10002/docker.sock`.
5. Validated private endpoint access from the dual-stack subnets to SSM and
   the AWS control plane, S3/DynamoDB Gateway Endpoint access, successful
   broker certificate hostname verification, and REST API response streaming
   in `ap-south-1`.

The deployment action is deliberately manual and remains disabled unless both
the dispatch input and `ANALYSIS_DEPLOY_ENABLED` repository variable are true.
The protected environment approval is an additional gate. The runner includes
the sandbox host validator in the SHA-addressed bundle, computes its SHA-256,
and each target host verifies that checksum before extraction and image load.
The validator is run only on the sandbox host before its Compose deployment.
Both hosts' current SHA releases are captured before deployment; if deployment
or the broker-readiness smoke test fails, the workflow attempts to restore
both old releases and checks analysis health and broker readiness. A failed
rollback verification fails the workflow. If the sandbox began stopped, its
initial power state is restored on both success and failure. Do not set those
controls until after the cost review and deployment authorization.

After a deployment, verify the template stack outputs, API health, anonymous
scan limits, authenticated Evolution stream, sandbox start/readiness/idle
stop, stale-lease recovery, backup snapshot, and a tested data-volume restore.
Review CloudWatch logs and alarms without logging authorization headers,
request bodies, internal tokens, signing keys, or authorizer context.

## Retained-resource cleanup

The retained S3 buckets, lifecycle table, analysis data volume, tagged
snapshots, and retained CloudWatch log groups have a separate manual cleanup
workflow. Delete the analysis stack first; the inventory and apply commands
both refuse to run while the named stack remains in any state other than
missing or `DELETE_COMPLETE`.

Before use, an operator must configure repository variables for
`ANALYSIS_STACK_NAME`, `ANALYSIS_RESOURCE_PREFIX`, `ANALYSIS_CLEANUP_REGION`,
`ANALYSIS_CLEANUP_READ_ROLE_ARN`, and
`ANALYSIS_CLEANUP_WRITE_ROLE_ARN`. The prefix must be the exact value used by
the stack's `AnalysisStack` tags. Set `ANALYSIS_CLEANUP_ENABLED` to the exact
string `true` only after cleanup authorization, and configure required
reviewers for the `analysis-cleanup` GitHub Environment outside the workflow
source. The read role must be read-only; the write role should be a distinct
least-privilege OIDC role.

The `workflow_dispatch` input `enable_cleanup` defaults to false. Dispatch
with it false to inspect the inventory. The inventory is printed in the
workflow run summary and its exact JSON plan is saved as the
`analysis-cleanup-plan` artifact before the destructive job waits for
`analysis-cleanup` environment approval. Review this full inventory and the
plan first. Cleanup proceeds only when the input is true,
`ANALYSIS_CLEANUP_ENABLED` is exactly `true`, and the required
`confirm_prefix` exactly equals `ANALYSIS_RESOURCE_PREFIX`. The apply command
downloads the plan from that inventory job and revalidates stack state,
resource identity, exact tags, allowlisted log-group names, and volume
attachment state immediately before each deletion.

Eligible resources are limited to S3 buckets, DynamoDB tables, EBS volumes,
and EBS snapshots with `AnalysisStack=<ANALYSIS_RESOURCE_PREFIX>`, and the
following six exact legacy log-group stems. Their exact unsuffixed names remain
eligible even if the CloudFormation stack ID is unavailable. A suffixed name
is eligible only when its 12-character lowercase hexadecimal suffix matches
the exact deleted stack ID:

| Template resource | Allowed log-group name |
| --- | --- |
| SSM sessions | `/aws/ssm/sessions/<prefix>` or `/aws/ssm/sessions/<prefix>-<suffix>` |
| Analysis proxy | `/aws/lambda/<prefix>-analysis-proxy` or `/aws/lambda/<prefix>-analysis-proxy-<suffix>` |
| Analysis authorizer | `/aws/lambda/<prefix>-analysis-authorizer` or `/aws/lambda/<prefix>-analysis-authorizer-<suffix>` |
| Source stager | `/aws/lambda/<prefix>-source-stager` or `/aws/lambda/<prefix>-source-stager-<suffix>` |
| Sandbox watchdog | `/aws/lambda/<prefix>-sandbox-watchdog` or `/aws/lambda/<prefix>-sandbox-watchdog-<suffix>` |
| API access | `/aws/apigateway/<prefix>-analysis` or `/aws/apigateway/<prefix>-analysis-<suffix>` |

Apply irreversibly deletes all current and versioned S3 objects and delete
markers before deleting eligible buckets, deletes DynamoDB tables and every
item in them, removes only unattached tagged EBS data volumes and tagged
snapshots, and deletes the allowlisted log groups. Confirm independent
backups, recovery needs, and retention requirements before approval. The
workflow and utility never terminate EC2 instances, delete AMIs, or operate on
the serverless archive stack.

Future OIDC roles should be restricted to this repository and workflow
identities. The inventory role needs only CloudFormation `DescribeStacks`;
S3 `ListAllMyBuckets`/`GetBucketTagging`; DynamoDB
`ListTables`/`DescribeTable`/`ListTagsOfResource`; EC2
`DescribeVolumes`/`DescribeSnapshots`; and Logs `DescribeLogGroups`. The
separate cleanup role additionally needs S3 `ListBucketVersions`,
`DeleteObjectVersion`, and `DeleteBucket` on eligible bucket ARNs; DynamoDB
`DeleteTable` on eligible table ARNs; EC2 `DeleteVolume`/`DeleteSnapshot`
conditioned on the exact `AnalysisStack` tag; and Logs `DeleteLogGroup` only
for the six allowlisted log-group patterns. EC2 describe actions that do not
support resource-level permissions may require `Resource: "*"`. Do not grant
instance termination, AMI deletion, CloudFormation stack deletion, or archive
permissions. Configure these IAM policies, OIDC trust, and GitHub environment
reviewers through the approved operator process, not in this repository.
