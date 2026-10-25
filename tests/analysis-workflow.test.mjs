import assert from 'node:assert/strict';
import { access, readFile } from 'node:fs/promises';
import { constants } from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const read = (relativePath) => readFile(path.join(root, relativePath), 'utf8');

const publishWorkflow = await read('.github/workflows/publish-api-image.yml');
const deployWorkflow = await read('.github/workflows/deploy-analysis.yml');
const goWorkflow = await read('.github/workflows/ci-go.yml');
const deploymentReadme = await read('README_DEPLOY.md');

test('publishes immutable SHA image tags', () => {
  assert.match(publishWorkflow, /on:\s*\n\s+push:/);
  assert.match(publishWorkflow, /docker\/setup-buildx-action@v3/);
  assert.match(publishWorkflow, /docker\/login-action@v3/);
  assert.match(publishWorkflow, /docker\/build-push-action@v6/);
  assert.match(publishWorkflow, /'go-backend\/\*\*'/);
  assert.match(publishWorkflow, /'thermodynamic-ast-engine\/\*\*'/);
  assert.match(publishWorkflow, /'docker\/\*\*'/);
  assert.match(publishWorkflow, /'\.dockerignore'/);
  assert.match(
    publishWorkflow,
    /ucs503-analysis:\$\{\{\s*github\.sha\s*\}\}/,
  );
  assert.match(
    publishWorkflow,
    /ucs503-sandbox-runner:\$\{\{\s*github\.sha\s*\}\}/,
  );
  assert.doesNotMatch(
    publishWorkflow,
    /ucs503-(?:analysis|sandbox-runner):latest\b/i,
  );
});

test('uses a current setup-go action in the Go CI workflow', () => {
  assert.match(goWorkflow, /actions\/setup-go@v6/);
  assert.doesNotMatch(goWorkflow, /actions\/setup-go@v[1-5]\b/);
  assert.match(goWorkflow, /analysis-authorizer\/\*\*/);
  assert.match(goWorkflow, /Test and vet REST authorizer/);
});

test('gates AWS upload and deploy behind manual opt-in and approval', () => {
  assert.match(deployWorkflow, /workflow_dispatch:/);
  assert.doesNotMatch(deployWorkflow, /^\s+push:/m);
  assert.match(
    deployWorkflow,
    /enable_deployment:[\s\S]*?type:\s*boolean[\s\S]*?default:\s*false/,
  );
  assert.match(
    deployWorkflow,
    /if:\s*\$\{\{\s*inputs\.enable_deployment\s*&&\s*vars\.ANALYSIS_DEPLOY_ENABLED\s*==\s*'true'\s*\}\}/,
  );
  assert.match(deployWorkflow, /environment:\s*\n\s+name:\s*analysis-deployment/);
  assert.match(deployWorkflow, /ref:\s*\$\{\{\s*needs\.validate\.outputs\.image_sha\s*\}\}/);
  assert.match(deployWorkflow, /analysis_image="ghcr\.io\/cv-code-vitals\/ucs503-analysis:\$IMAGE_SHA"/);
  assert.match(deployWorkflow, /runner_image="ghcr\.io\/cv-code-vitals\/ucs503-sandbox-runner:\$IMAGE_SHA"/);
  assert.match(deployWorkflow, /docker pull "\$analysis_image"/);
  assert.match(deployWorkflow, /docker pull "\$runner_image"/);
  assert.match(deployWorkflow, /bundles\/analysis-deployments\/\$IMAGE_SHA\/deployment-bundle\.tar\.gz/);
  assert.match(deployWorkflow, /aws s3 cp/);
  assert.match(deployWorkflow, /--sse AES256/);
  assert.match(deployWorkflow, /--pull never/);
  assert.doesNotMatch(
    deployWorkflow,
    /ucs503-(?:analysis|sandbox-runner):latest\b/i,
  );
});

test('uses OIDC and SSM without SSH secrets or private-host GHCR pulls', () => {
  assert.match(deployWorkflow, /id-token:\s*write/);
  assert.match(deployWorkflow, /aws-actions\/configure-aws-credentials@v4/);
  assert.match(deployWorkflow, /role-to-assume:\s*\$\{\{\s*vars\.ANALYSIS_DEPLOY_ROLE_ARN\s*\}\}/);
  assert.match(deployWorkflow, /aws ssm send-command/);
  assert.match(deployWorkflow, /AWS-RunShellScript/);
  assert.match(deployWorkflow, /analysis-compose\.yml|docker-compose\.analysis\.yml/);
  assert.doesNotMatch(
    deployWorkflow,
    /EC2_SSH_PRIVATE_KEY|EC2_HOST|appleboy\/|ssh-action|scp-action/,
  );
  const remoteDeployment = deployWorkflow.slice(
    deployWorkflow.indexOf('remote_command="'),
  );
  assert.match(remoteDeployment, /aws s3 cp --region '\$\{AWS_REGION\}'/);
  assert.doesNotMatch(remoteDeployment, /docker pull\b/);
  assert.match(remoteDeployment, /docker load -i/);
  assert.match(deploymentReadme, /hosts do not pull from GHCR/i);
});

test('verifies deployed image identities and analysis health', () => {
  assert.match(deployWorkflow, /analysis_image_id=.*docker image inspect --format/i);
  assert.match(deployWorkflow, /runner_image_id=.*docker image inspect --format/i);
  assert.match(deployWorkflow, /docker inspect --format '\{\{\.Image\}\}'/);
  assert.match(deployWorkflow, /actual_image_id.*expected_image_id|expected_image_id.*actual_image_id/);
  assert.match(deployWorkflow, /SANDBOX_RUNNER_IMAGE/);
  assert.match(deployWorkflow, /Health\.Status|\.State\.Health\.Status/);
  assert.match(deployWorkflow, /127\.0\.0\.1:8080\/health/);
  assert.match(deployWorkflow, /sandbox_available/);
  assert.match(deployWorkflow, /sandbox broker readiness smoke test/i);
  assert.match(deployWorkflow, /sandbox_available.*status/);
  const sandboxDeployment = deployWorkflow.lastIndexOf('"$SANDBOX_RUNNER_IMAGE_ID"');
  const smokeCheck = deployWorkflow.lastIndexOf('verify_analysis_sandbox_smoke');
  assert.ok(sandboxDeployment >= 0 && sandboxDeployment < smokeCheck);
});

test('starts a stopped sandbox only for deployment and restores its state', () => {
  const analysisDeployment = deployWorkflow.search(/deploy_instance \\\r?\n\s+"\$ANALYSIS_INSTANCE_ID"/);
  const sandboxStateCheck = deployWorkflow.indexOf('sandbox_state="$(aws ec2 describe-instances');
  const sandboxDeployment = deployWorkflow.search(/deploy_instance \\\r?\n\s+"\$SANDBOX_INSTANCE_ID"/);
  assert.ok(analysisDeployment >= 0 && analysisDeployment < sandboxStateCheck);
  assert.ok(sandboxStateCheck < sandboxDeployment);
  assert.match(deployWorkflow, /aws ec2 describe-instances/);
  assert.match(deployWorkflow, /State\.Name/);
  assert.match(deployWorkflow, /aws ec2 start-instances/);
  assert.match(deployWorkflow, /aws ec2 wait instance-running/);
  assert.match(deployWorkflow, /PingStatus/);
  assert.match(deployWorkflow, /trap .*EXIT|trap [\s\S]*cleanup_sandbox/);
  assert.match(deployWorkflow, /aws ec2 stop-instances/);
  assert.match(deployWorkflow, /sandbox_started_by_workflow/);
  assert.match(deploymentReadme, /ec2:DescribeInstances/);
  assert.match(deploymentReadme, /ec2:StartInstances/);
  assert.match(deploymentReadme, /ec2:StopInstances/);
  assert.match(deploymentReadme, /ssm:DescribeInstanceInformation/);
});

test('removes the obsolete SSH-based EC2 deployment path', async () => {
  await assert.rejects(
    access(
      path.join(root, '.github/workflows/deploy-to-ec2.yml'),
      constants.F_OK,
    ),
  );
  assert.match(deploymentReadme, /docker-compose\.ec2\.yml.*and the old `ucs503-api` service are legacy/i);
  assert.match(deploymentReadme, /existing GitHub Pages workflow, which remains unchanged/i);
});
