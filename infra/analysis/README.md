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
validation has been performed. In particular, verify the 100-GiB gp3 data
volume, seven retained daily snapshots, two EC2 instance profiles, API
Gateway/Lambda streaming, CloudWatch retention, S3, and transfer charges
before approving. The target is not a validated estimate.

## Stack contracts

- Stack name: `ucs503-analysis-v1`; default region: `ap-south-1`.
- Both instances have private ENIs and no public IPv4 address or SSH ingress.
  The proxy reaches the analysis private IP; broker ingress is allowed only
  from the analysis security group.
- A S3 Gateway Endpoint provides private access to source and deployment
  buckets. The source stager runs outside the VPC to access GitHub safely.
- The VPC proxy's Lambda/SSM SDK clients and the analysis host's EC2/DynamoDB
  clients request dual-stack endpoints. The S3 client retains the regional
  endpoint to use the Gateway Endpoint. The AMIs must also have an SSM Agent
  configured for working dual-stack control-plane access; verify this from
  the deployed private subnets before treating the stack as ready.
- Staged source objects expire after one day. Image deployment bundles expire
  after 30 days under `bundles/`.
- Analysis outputs bind to `/var/lib/ucs503-analysis/uploads` on a separate,
  encrypted gp3 volume. DLM creates daily snapshots and retains seven. Snapshot
  restore is a manual, unautomated process and must be tested before production.
- The `sandbox-lifecycle` DynamoDB record coordinates leases. The analysis
  service starts the selected sandbox instance, waits for EC2 and authenticated
  broker readiness, and releases leases after Evolution. The scheduled
  watchdog expires stale leases and coordinates idle stop recovery.
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
with reviewed, supported images only after the separate approval gate; do not
commit account-specific IDs, real secrets, or generated host configuration.

## Deployment prerequisites and checks

Before a separately authorized deployment, prepare and review:

1. Supported analysis and sandbox AMIs with the SSM Agent, Docker Engine,
   Docker Compose plugin, AWS CLI, Node.js 22 for the read-only host validator,
   and host bootstrap configuration. The
   analysis host must not have access to a Docker socket. Prepare the sandbox
   host with rootless Docker for UID/GID `10002`, enable systemd linger for
   that dedicated user, and verify the daemon socket is
   `/run/user/10002/docker.sock`; the broker mounts that socket inside its
   container at `/run/docker.sock`. The daemon and broker must see the same
   absolute sandbox workspace path. Run
   `sudo node scripts/validate-sandbox-host.mjs` on the sandbox host to check
   the UID/GID, lingering, socket, workspace access, and rootless daemon.
2. A protected `analysis-deployment` GitHub Environment with required
   reviewers and repository variables for the deploy role ARN, region, private
   artifact bucket, and both instance IDs. Restrict the OIDC role to this
   repository and environment. Grant it only `s3:PutObject` to the deployment
   bundle prefix, `ssm:SendCommand`/`ssm:GetCommandInvocation` for the two
   managed nodes, and `ssm:DescribeInstanceInformation` for the sandbox
   readiness check. The workflow's power-state safeguard also requires
   `ec2:DescribeInstances` and `ec2:StartInstances`/`ec2:StopInstances`; scope
   start/stop to only the sandbox instance ARN. It never starts or stops the
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
   `SANDBOX_LIFECYCLE_TABLE`, `SANDBOX_TLS_CA_SOURCE`, and
   `ANALYSIS_UPLOADS_HOST_DIR`. The sandbox env file must configure the
   broker token, certificate/key sources, workspace directory, and
   `DOCKER_SOCKET_SOURCE=/run/user/10002/docker.sock`.
5. Validated IPv6 access from the private subnets to the required SSM and EC2
   control-plane endpoints, S3 Gateway Endpoint access, successful broker
   certificate hostname verification, and REST API response streaming in
   `ap-south-1`.

The deployment action is deliberately manual and remains disabled unless both
the dispatch input and `ANALYSIS_DEPLOY_ENABLED` repository variable are true.
The protected environment approval is an additional gate. Do not set those
controls until after the cost review and deployment authorization.

After a deployment, verify the template stack outputs, API health, anonymous
scan limits, authenticated Evolution stream, sandbox start/readiness/idle
stop, stale-lease recovery, backup snapshot, and a tested data-volume restore.
Review CloudWatch logs and alarms without logging authorization headers,
request bodies, internal tokens, signing keys, or authorizer context.
