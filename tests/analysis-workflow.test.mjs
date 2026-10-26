import assert from 'node:assert/strict';
import { access, readFile } from 'node:fs/promises';
import { constants } from 'node:fs';
import { spawnSync } from 'node:child_process';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const read = (relativePath) => readFile(path.join(root, relativePath), 'utf8');

const publishWorkflow = await read('.github/workflows/publish-api-image.yml');
const deployWorkflow = await read('.github/workflows/deploy-analysis.yml');
const cleanupWorkflow = await read('.github/workflows/cleanup-analysis.yml');
const goWorkflow = await read('.github/workflows/ci-go.yml');
const deploymentReadme = await read('README_DEPLOY.md');
const analysisCompose = await read('docker-compose.analysis.yml');

test('publishes immutable SHA image tags', () => {
  assert.match(publishWorkflow, /on:\s*\n\s+workflow_dispatch:\s*\n\s+push:/);
  assert.match(publishWorkflow, /push:\s*\n\s+branches:\s*\n\s+- main/);
  assert.match(publishWorkflow, /docker\/setup-buildx-action@v3/);
  assert.match(publishWorkflow, /docker\/login-action@v3/);
  assert.match(publishWorkflow, /docker\/build-push-action@v6/);
  assert.match(publishWorkflow, /attestations:\s*write/);
  assert.match(publishWorkflow, /id-token:\s*write/);
  assert.match(publishWorkflow, /actions\/attest@v4/);
  assert.match(publishWorkflow, /subject-digest:\s*\$\{\{\s*steps\.[a-z_]+\.outputs\.digest\s*\}\}/);
  assert.match(publishWorkflow, /push-to-registry:\s*true/);
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
  assert.match(publishWorkflow, /id-token:\s*write/);
  assert.match(publishWorkflow, /id:\s*analysis_image/);
  assert.match(publishWorkflow, /id:\s*sandbox_runner_image/);
  assert.match(publishWorkflow, /sigstore\/cosign-installer@v3/);
  assert.match(publishWorkflow, /cosign sign --yes/);
  assert.match(publishWorkflow, /steps\.analysis_image\.outputs\.digest/);
  assert.match(publishWorkflow, /steps\.sandbox_runner_image\.outputs\.digest/);
  assert.doesNotMatch(
    publishWorkflow,
    /ucs503-(?:analysis|sandbox-runner):latest\b/i,
  );
});

test('verifies signed image provenance and pins deployments to manifest digests', () => {
  assert.match(deployWorkflow, /gh attestation verify/);
  assert.match(deployWorkflow, /oci:\/\/\$analysis_image_digest/);
  assert.match(deployWorkflow, /oci:\/\/\$runner_image_digest/);
  assert.match(deployWorkflow, /--repo CV-Code-Vitals\/Thermodynamic-Evolution-UCS503/);
  assert.match(deployWorkflow, /--signer-workflow CV-Code-Vitals\/Thermodynamic-Evolution-UCS503\/\.github\/workflows\/publish-api-image\.yml/);
  assert.match(deployWorkflow, /--source-digest "\$IMAGE_SHA"/);
  assert.match(deployWorkflow, /docker pull "\$analysis_image_digest"/);
  assert.match(deployWorkflow, /docker pull "\$runner_image_digest"/);
});

test('uses a current setup-go action in the Go CI workflow', () => {
  assert.match(goWorkflow, /actions\/setup-go@v6/);
  assert.doesNotMatch(goWorkflow, /actions\/setup-go@v[1-5]\b/);
  assert.match(goWorkflow, /analysis-authorizer\/\*\*/);
  assert.match(goWorkflow, /Test and vet REST authorizer/);
});

test('gates AWS upload and deploy behind manual opt-in and approval', () => {
  assert.match(deployWorkflow, /workflow_dispatch:/);
  assert.match(
    deployWorkflow,
    /push:\s*\n\s+branches:\s*\n\s+- main\s*\n\s+paths:\s*\n\s+- '\.github\/workflows\/deploy-analysis\.yml'/,
  );
  assert.match(
    deployWorkflow,
    /registration:\s*\n\s+if:\s*\$\{\{\s*github\.event_name == 'push'\s*\}\}/,
  );
  assert.match(
    deployWorkflow,
    /validate:\s*\n\s+if:\s*\$\{\{\s*github\.event_name == 'workflow_dispatch'\s*\}\}/,
  );
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
  assert.match(deployWorkflow, /sigstore\/cosign-installer@v3/);
  assert.match(deployWorkflow, /docker image inspect --format '\{\{range \.RepoDigests\}\}\{\{println \.\}\}\{\{end\}\}'/);
  assert.match(deployWorkflow, /cosign verify --certificate-identity/);
  assert.match(deployWorkflow, /certificate-oidc-issuer/);
  assert.match(deployWorkflow, /publish-api-image\.yml@refs\/heads\/main/);
  assert.match(deployWorkflow, /analysis_repository=ghcr\.io\/cv-code-vitals\/ucs503-analysis/);
  assert.match(deployWorkflow, /runner_repository=ghcr\.io\/cv-code-vitals\/ucs503-sandbox-runner/);
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
  assert.match(deploymentReadme, /hosts do not pull from\s+GHCR/i);
});

test('resolves deployment resources from the analysis stack outputs', () => {
  assert.match(deployWorkflow, /ANALYSIS_STACK_NAME:\s*\$\{\{\s*vars\.ANALYSIS_STACK_NAME\s*\}\}/);
  assert.match(deployWorkflow, /ANALYSIS_DEPLOY_ROLE_ARN:\s*\$\{\{\s*vars\.ANALYSIS_DEPLOY_ROLE_ARN\s*\}\}/);
  assert.match(deployWorkflow, /ANALYSIS_DEPLOY_REGION:\s*\$\{\{\s*vars\.ANALYSIS_DEPLOY_REGION\s*\}\}/);
  assert.doesNotMatch(
    deployWorkflow,
    /vars\.ANALYSIS_ARTIFACT_BUCKET|vars\.ANALYSIS_INSTANCE_ID|vars\.SANDBOX_INSTANCE_ID/,
  );
  assert.match(deployWorkflow, /ANALYSIS_STACK_NAME[\s\S]*ANALYSIS_DEPLOY_REGION/);
  assert.match(deployWorkflow, /aws cloudformation describe-stacks/);

  const describeStacks = deployWorkflow.indexOf('aws cloudformation describe-stacks');
  const uploadBundle = deployWorkflow.indexOf('aws s3 cp analysis-deployment-bundle.tar.gz');
  const sendCommand = deployWorkflow.indexOf('aws ssm send-command');
  assert.ok(describeStacks >= 0 && describeStacks < uploadBundle);
  assert.ok(describeStacks < sendCommand);
  for (const outputName of [
    'ArtifactBucketName',
    'AnalysisInstanceId',
    'SandboxInstanceId',
    'SandboxLifecycleTable',
  ]) {
    assert.match(deployWorkflow, new RegExp(outputName));
    assert.match(deployWorkflow, /matches.*!= 1/);
    assert.match(deployWorkflow, /OutputValue/);
  }
  assert.match(deployWorkflow, /SANDBOX_LIFECYCLE_TABLE=.*SANDBOX_LIFECYCLE_TABLE/);
  assert.match(deploymentReadme, /ANALYSIS_STACK_NAME/);
  assert.match(deploymentReadme, /cloudformation:DescribeStacks/i);
});

test('rejects a deployment region that could alter an interpolated SSM command', () => {
  const checkVariables = deployWorkflow.slice(
    deployWorkflow.indexOf('name: Check deployment variables'),
    deployWorkflow.indexOf('name: Log in to GitHub Container Registry'),
  );
  assert.match(checkVariables, /ANALYSIS_DEPLOY_REGION must be a safe AWS region identifier/);
  assert.match(checkVariables, /\^\[a-z\]\{2\}\(-\[a-z0-9\]\+\)\+-\[0-9\]\+\$/);
  assert.doesNotMatch('us-east-1; touch injected', /^[a-z]{2}(-[a-z0-9]+)+-[0-9]+$/);
  assert.doesNotMatch("us-east-1' injected", /^[a-z]{2}(-[a-z0-9]+)+-[0-9]+$/);
  assert.match('us-gov-west-1', /^[a-z]{2}(-[a-z0-9]+)+-[0-9]+$/);
});

test('validates stack output formats and lengths before writing workflow outputs', () => {
  const outputResolutionStart = deployWorkflow.indexOf(
    'artifact_bucket_name="$(resolve_output ArtifactBucketName)"',
  );
  const outputWriteStart = deployWorkflow.indexOf(
    'echo "artifact_bucket_name=$artifact_bucket_name"',
    outputResolutionStart,
  );
  const outputValidation = deployWorkflow.slice(
    outputResolutionStart,
    outputWriteStart,
  );

  assert.ok(outputResolutionStart >= 0 && outputWriteStart > outputResolutionStart);
  assert.match(outputValidation, /\$\{#artifact_bucket_name\} < 3/);
  assert.match(outputValidation, /\$\{#artifact_bucket_name\} > 63/);
  assert.match(outputValidation, /artifact_bucket_name.*\^\[a-z0-9\]\[a-z0-9.-\]\*\[a-z0-9\]\$/);
  assert.match(outputValidation, /artifact_bucket_name.*\*\.\.\*/);
  assert.match(outputValidation, /artifact_bucket_name.*== xn--\*/);
  assert.match(outputValidation, /artifact_bucket_name.*== \*--table-s3/);
  assert.match(outputValidation, /"\$analysis_instance_id" =~ \^i-\[a-f0-9\]\{8\}\(\[a-f0-9\]\{9\}\)\?\$/);
  assert.match(outputValidation, /"\$sandbox_instance_id" =~ \^i-\[a-f0-9\]\{8\}\(\[a-f0-9\]\{9\}\)\?\$/);
  assert.match(outputValidation, /\$\{#sandbox_lifecycle_table\} < 3/);
  assert.match(outputValidation, /\$\{#sandbox_lifecycle_table\} > 255/);
  assert.match(outputValidation, /sandbox_lifecycle_table.*\^\[A-Za-z0-9_.-\]\+\$/);
  assert.ok(outputWriteStart > outputValidation.indexOf('sandbox_lifecycle_table'));
  assert.match(
    deployWorkflow,
    /select\(type == "string" and length > 0\)/,
  );

  const isValidBucket = (value) =>
    value.length >= 3 &&
    value.length <= 63 &&
    /^[a-z0-9][a-z0-9.-]*[a-z0-9]$/.test(value) &&
    !value.includes('..') &&
    !value.includes('.-') &&
    !value.includes('-.') &&
    !value.startsWith('xn--') &&
    !value.startsWith('sthree-') &&
    !value.endsWith('-s3alias') &&
    !value.endsWith('--ol-s3') &&
    !value.endsWith('.mrap') &&
    !value.endsWith('--x-s3') &&
    !value.endsWith('--table-s3') &&
    !/^([0-9]{1,3}\.){3}[0-9]{1,3}$/.test(value);
  const isValidLifecycleTable = (value) =>
    value.length >= 3 &&
    value.length <= 255 &&
    /^[A-Za-z0-9_.-]+$/.test(value);

  assert.equal(isValidBucket('ucs503-artifacts-123'), true);
  assert.equal(isValidBucket('AnalysisBucket'), false);
  assert.equal(isValidBucket('ab'), false);
  assert.equal(isValidBucket('x'.repeat(64)), false);
  assert.equal(isValidBucket('bad..bucket'), false);
  assert.equal(isValidBucket('192.168.0.1'), false);
  assert.equal(isValidBucket('xn--bucket'), false);
  assert.equal(isValidBucket('bucket--table-s3'), false);
  assert.equal(isValidLifecycleTable('sandbox-lifecycle'), true);
  assert.equal(isValidLifecycleTable('ab'), false);
  assert.equal(isValidLifecycleTable('x'.repeat(256)), false);
  assert.equal(isValidLifecycleTable('table/name'), false);
  assert.doesNotMatch('i-a', /^i-[a-f0-9]{8}([a-f0-9]{9})?$/);
  assert.doesNotMatch('i-1234567', /^i-[a-f0-9]{8}([a-f0-9]{9})?$/);
  assert.doesNotMatch(`i-${'a'.repeat(18)}`, /^i-[a-f0-9]{8}([a-f0-9]{9})?$/);
});

test('preflights external secret dependencies without exposing secret values', () => {
  const secretPreflight = deployWorkflow.slice(
    deployWorkflow.indexOf('name: Verify required external SSM secrets'),
    deployWorkflow.indexOf('name: Upload private deployment bundle'),
  );
  for (const key of [
    'ApiTokenParameterName',
    'SigningKeyParameterName',
    'SandboxTokenParameterName',
    'SandboxTlsCertificateParameterName',
    'SandboxTlsPrivateKeyParameterName',
    'SandboxCaParameterName',
  ]) assert.match(secretPreflight, new RegExp(key));
  assert.match(secretPreflight, /aws ssm get-parameters/);
  assert.match(secretPreflight, /Type == "SecureString"/);
  assert.doesNotMatch(secretPreflight, /--with-decryption/);
});

test('makes deployment bundle upload idempotent and refuses a conflicting SHA release', () => {
  const upload = deployWorkflow.slice(
    deployWorkflow.indexOf('name: Upload private deployment bundle'),
    deployWorkflow.indexOf('name: Deploy both private hosts through SSM'),
  );
  assert.match(upload, /aws s3api head-object/);
  assert.match(upload, /Metadata\.bundle-sha256/);
  assert.match(upload, /Refusing to overwrite a different deployment bundle/);
  assert.match(upload, /--metadata "bundle-sha256=\$BUNDLE_SHA256"/);
  assert.match(upload, /NotFound\|404\|NoSuchKey/);
});

test('preflights only the sandbox and verifies the bundle before extraction', () => {
  assert.match(deployWorkflow, /scripts\/validate-sandbox-host\.mjs/);
  assert.match(deployWorkflow, /node .*validate-sandbox-host\.mjs/);
  assert.match(deployWorkflow, /tar -czf analysis-deployment-bundle\.tar\.gz[\s\S]*scripts\/validate-sandbox-host\.mjs/);
  assert.match(deployWorkflow, /sha256sum analysis-deployment-bundle\.tar\.gz/);
  assert.match(deployWorkflow, /sha256sum -c/);
  assert.match(deployWorkflow, /SANDBOX_LIFECYCLE_TABLE/);

  const extraction = deployWorkflow.indexOf('tar -xzf');
  const imageLoad = deployWorkflow.indexOf('docker load -i');
  const remoteChecksum = deployWorkflow.indexOf('sha256sum -c');
  const bundleDownload = deployWorkflow.indexOf('bundle=/var/tmp/ucs503-deploy-');
  const preflight = deployWorkflow.indexOf('node ', bundleDownload);
  assert.ok(remoteChecksum >= 0 && remoteChecksum < extraction);
  assert.ok(remoteChecksum < imageLoad);
  assert.ok(preflight >= 0 && preflight < deployWorkflow.indexOf('docker compose --project-name ${project_name}', preflight));
  assert.match(
    deployWorkflow,
    /if \[ '\$\{project_name\}' = 'ucs503-sandbox' \]; then[\s\S]*?node .*validate-sandbox-host\.mjs[\s\S]*?fi/,
  );
  assert.match(
    deployWorkflow,
    /deploy_instance \\\r?\n\s+"\$ANALYSIS_INSTANCE_ID"[\s\S]*?ucs503-analysis/,
  );
});

test('captures both current releases and rolls both hosts back on any deployment failure', () => {
  assert.match(deployWorkflow, /capture_previous_release/);
  assert.match(deployWorkflow, /previous_analysis_image_sha/);
  assert.match(deployWorkflow, /previous_sandbox_image_sha/);
  assert.match(deployWorkflow, /rollback_instance/);
  assert.match(deployWorkflow, /rollback_both_hosts/);
  assert.match(deployWorkflow, /rollback_instance[\s\S]*?ANALYSIS_INSTANCE_ID/);
  assert.match(deployWorkflow, /rollback_instance[\s\S]*?SANDBOX_INSTANCE_ID/);
  assert.match(deployWorkflow, /verify_analysis_sandbox_smoke/);
  assert.match(deployWorkflow, /trap .*rollback|trap [\s\S]*rollback_both_hosts/);
  assert.match(deployWorkflow, /rollback.*health|health.*rollback/i);
  assert.match(deployWorkflow, /rollback.*broker readiness|broker readiness.*rollback/i);
  assert.match(deployWorkflow, /sandbox_started_by_workflow/);
  assert.match(deployWorkflow, /cleanup_sandbox/);
  assert.match(deploymentReadme, /rollback/i);
  assert.match(deployWorkflow, /SANDBOX_LIFECYCLE_TABLE/);

  const captureAnalysis = deployWorkflow.indexOf('previous_analysis_capture="$(capture_previous_release');
  const captureSandbox = deployWorkflow.indexOf('previous_sandbox_capture="$(capture_previous_release');
  const firstDeployment = deployWorkflow.search(/deploy_instance \\\r?\n\s+"\$ANALYSIS_INSTANCE_ID"/);
  assert.ok(captureAnalysis >= 0 && captureAnalysis < firstDeployment);
  assert.ok(captureSandbox >= 0 && captureSandbox < firstDeployment);
});

test('restores the captured lifecycle table on rollback without overriding pre-change defaults', () => {
  const captureStart = deployWorkflow.indexOf('capture_previous_release() {');
  const deploymentStart = deployWorkflow.indexOf('deploy_instance() {');
  const rollbackStart = deployWorkflow.indexOf('rollback_instance() {');
  const smokeCheckStart = deployWorkflow.indexOf('verify_analysis_sandbox_smoke() {');
  const capture = deployWorkflow.slice(captureStart, deploymentStart);
  const candidateDeployment = deployWorkflow.slice(deploymentStart, rollbackStart);
  const rollback = deployWorkflow.slice(rollbackStart, smokeCheckStart).replace(/\r/g, '');

  assert.match(
    capture,
    /lifecycle_environment=.*docker inspect --format '\{\{range \.Config\.Env\}\}\{\{println \.\}\}\{\{end\}\}'/,
  );
  assert.match(capture, /lifecycle_table_entry=.*grep '\^SANDBOX_LIFECYCLE_TABLE='/);
  assert.match(capture, /previous_lifecycle_table_state=unset/);
  assert.match(capture, /previous_lifecycle_table_state=set/);
  assert.match(capture, /previous_lifecycle_table_b64=/);
  assert.match(capture, /printf '%s\|%s\|%s\|%s'/);
  assert.match(
    deployWorkflow,
    /IFS='\|' read -r \\\r?\n\s+previous_analysis_image_sha \\\r?\n\s+previous_analysis_container_ids \\\r?\n\s+previous_analysis_lifecycle_table_state \\\r?\n\s+previous_analysis_lifecycle_table_b64/,
  );

  assert.match(
    candidateDeployment,
    /export SANDBOX_LIFECYCLE_TABLE='\$\{SANDBOX_LIFECYCLE_TABLE\}'/,
  );
  assert.match(rollback, /previous_lifecycle_table_state="\$9"/);
  const declarations = rollback.match(/rollback_instance\(\) \{\n([\s\S]*?)\n\s+local remote_command/)?.[1];
  assert.ok(declarations, 'rollback positional arguments should be captured');
  const lifecycleTableDeclaration = declarations
    .split('\n')
    .find((line) => line.includes('local previous_lifecycle_table_b64='));
  assert.match(lifecycleTableDeclaration, /local previous_lifecycle_table_b64="\$\{10\}"/);
  const shell = [
    'rollback_instance() {',
    lifecycleTableDeclaration,
    '  printf "%s" "$previous_lifecycle_table_b64"',
    '}',
    'rollback_instance id sha compose env project service image baseline set Y2FwdHVyZWQtdGFibGU=',
  ].join('\n');
  const parsedArgument = spawnSync('bash', ['-s'], { encoding: 'utf8', input: shell });
  assert.equal(parsedArgument.status, 0, parsedArgument.stderr);
  assert.equal(parsedArgument.stdout, 'Y2FwdHVyZWQtdGFibGU=');
  assert.match(
    rollback,
    /'\$\{previous_lifecycle_table_state\}' = set/,
  );
  assert.match(rollback, /base64 -d/);
  assert.match(rollback, /'\$\{previous_lifecycle_table_b64\}'/);
  assert.doesNotMatch(
    rollback,
    /export SANDBOX_LIFECYCLE_TABLE='\$\{SANDBOX_LIFECYCLE_TABLE\}'/,
  );
  assert.match(rollback, /docker compose[\s\S]*? up --pull never -d/);
  assert.match(
    analysisCompose,
    /SANDBOX_LIFECYCLE_TABLE: \$\{SANDBOX_LIFECYCLE_TABLE:-sandbox-lifecycle\}/,
  );

});

test('attempts both host rollbacks when either host rollback fails', () => {
  const rollbackStart = deployWorkflow.indexOf('rollback_both_hosts() {');
  const cleanupStart = deployWorkflow.indexOf('cleanup_sandbox() {', rollbackStart);
  const rollbackFunction = deployWorkflow.slice(rollbackStart, cleanupStart).replace(/\r/g, '').trim();
  assert.ok(rollbackStart >= 0 && cleanupStart > rollbackStart);

  for (const failedHost of ['analysis-instance', 'sandbox-instance']) {
    const rollbackWithFixedInputs = rollbackFunction
      .replace(/\$ANALYSIS_INSTANCE_ID/g, 'analysis-instance')
      .replace(/\$SANDBOX_INSTANCE_ID/g, 'sandbox-instance')
      .replace(/\$previous_analysis_image_sha/g, 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa')
      .replace(/\$previous_sandbox_image_sha/g, 'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb')
      .replace(/\$previous_analysis_container_ids/g, '')
      .replace(/\$previous_sandbox_container_ids/g, '')
      .replace(/\$previous_analysis_lifecycle_table_state/g, 'unset')
      .replace(/\$previous_analysis_lifecycle_table_b64/g, '');
    const shell = [
      'rollback_instance() {',
      '  ROLLBACK_CALLS="${ROLLBACK_CALLS},$1"',
      `  [[ "$1" != "${failedHost}" ]]`,
      '}',
      'verify_analysis_sandbox_smoke() { return 0; }',
      'ROLLBACK_CALLS=',
      rollbackWithFixedInputs,
      'rollback_both_hosts >/dev/null 2>&1',
      'printf "%s|%s" "$?" "$ROLLBACK_CALLS"',
    ].join('\n');
    const result = spawnSync('bash', ['-s'], { encoding: 'utf8', input: shell });
    assert.equal(result.status, 0, result.stderr);
    assert.equal(result.stdout, '1|,analysis-instance,sandbox-instance');
  }
});

test('encodes complete SSM scripts before constructing the SendCommand request', async () => {
  const encoderPath = path.join(root, 'scripts', 'encode-ssm-command.mjs');
  const hostileScript = "echo safe'; touch /tmp/untrusted; echo ' $(id) `whoami`";
  const encoded = spawnSync(process.execPath, [encoderPath], {
    encoding: 'utf8',
    input: hostileScript,
  });
  assert.equal(encoded.status, 0, encoded.stderr);
  const match = /^printf '%s' '([A-Za-z0-9+/=]+)' \| base64 -d \| bash$/.exec(encoded.stdout);
  assert.ok(match, `encoder must emit a shell-safe base64 wrapper; got: ${encoded.stdout}`);
  assert.equal(Buffer.from(match[1], 'base64').toString('utf8'), hostileScript);

  const sendFunction = deployWorkflow.slice(
    deployWorkflow.indexOf('send_ssm_command() {'),
    deployWorkflow.indexOf('capture_previous_release() {'),
  );
  assert.match(sendFunction, /encode-ssm-command\.mjs/);
  assert.match(sendFunction, /jq -n --arg command/);
});

test('rolls back older sandbox releases without requiring the new validator', () => {
  const deploymentStart = deployWorkflow.indexOf('deploy_instance() {');
  const rollbackStart = deployWorkflow.indexOf('rollback_instance() {');
  const smokeCheckStart = deployWorkflow.indexOf('verify_analysis_sandbox_smoke() {');
  const deployment = deployWorkflow.slice(deploymentStart, rollbackStart);
  const rollback = deployWorkflow.slice(rollbackStart, smokeCheckStart);

  assert.match(deployment, /node .*validate-sandbox-host\.mjs/);
  assert.doesNotMatch(rollback, /validate-sandbox-host\.mjs/);
  assert.match(rollback, /Captured rollback Compose release is unavailable/);
  assert.match(rollback, /docker compose[\s\S]*? up --pull never -d/);
  assert.match(rollback, /export DOCKER_HOST=unix:\/\/\/run\/user\/10002\/docker\.sock/);
});

test('preserves non-empty absent-service baselines exactly during rollback', () => {
  assert.match(deployWorkflow, /printf 'absent\|%s\|unset\|'/);
  assert.match(
    deployWorkflow,
    /docker ps -aq --filter 'label=com\.docker\.compose\.project=\$\{project_name\}' --filter 'label=com\.docker\.compose\.service=\$\{service_name\}'/,
  );
  assert.match(deployWorkflow, /previous_analysis_container_ids/);
  assert.match(deployWorkflow, /previous_sandbox_container_ids/);
  assert.match(deployWorkflow, /if \[ '\$\{previous_sha\}' = absent \]; then[\s\S]*?docker rm -f/);
  assert.match(deployWorkflow, /baseline_container_ids/);
  assert.match(deployWorkflow, /docker ps -aq --filter 'label=com\.docker\.compose\.project=\$\{project_name\}'/);
  assert.match(deployWorkflow, /expected_baseline_container_ids/);
  assert.match(deployWorkflow, /actual_baseline_container_ids/);
  assert.match(deployWorkflow, /actual_baseline_container_ids.*expected_baseline_container_ids/s);
  assert.doesNotMatch(deployWorkflow, /--remove-orphans/);
  assert.doesNotMatch(deployWorkflow, /docker compose[^\n]*down[^\n]*-v/);
  assert.doesNotMatch(deployWorkflow, /docker volume rm/);
  assert.match(deployWorkflow, /previous_analysis_image_sha" != absent/);
  assert.match(deployWorkflow, /previous_sandbox_image_sha" != absent/);
  assert.match(deployWorkflow, /Active service image is not a SHA-addressed release/);
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
  const sandboxStateCheck = deployWorkflow.indexOf('sandbox_state="$(aws ec2 describe-instances');
  const analysisDeployment = deployWorkflow.search(/deploy_instance \\\r?\n\s+"\$ANALYSIS_INSTANCE_ID"/);
  const sandboxDeployment = deployWorkflow.search(/deploy_instance \\\r?\n\s+"\$SANDBOX_INSTANCE_ID"/);
  assert.ok(sandboxStateCheck >= 0 && sandboxStateCheck < analysisDeployment);
  assert.ok(analysisDeployment < sandboxDeployment);
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

test('keeps retained-resource cleanup manual, separately gated, and protected', () => {
  assert.match(cleanupWorkflow, /^name:.*$/m);
  assert.match(cleanupWorkflow, /on:\s*\n\s+workflow_dispatch:/);
  assert.doesNotMatch(cleanupWorkflow, /^\s{2}(?:push|pull_request|schedule):/m);
  assert.match(
    cleanupWorkflow,
    /enable_cleanup:[\s\S]*?type:\s*boolean[\s\S]*?default:\s*false/,
  );
  assert.match(cleanupWorkflow, /confirm_prefix:[\s\S]*?type:\s*string/);
  assert.match(cleanupWorkflow, /jobs:\s*\n\s+inventory:/);
  assert.match(cleanupWorkflow, /\n\s+cleanup:\s*\n\s+needs:\s*inventory/);
  assert.match(cleanupWorkflow, /vars\.ANALYSIS_CLEANUP_READ_ROLE_ARN/);
  assert.match(cleanupWorkflow, /vars\.ANALYSIS_CLEANUP_WRITE_ROLE_ARN/);
  assert.match(cleanupWorkflow, /vars\.ANALYSIS_CLEANUP_ENABLED\s*==\s*'true'/);
  assert.match(cleanupWorkflow, /inputs\.enable_cleanup/);
  assert.match(cleanupWorkflow, /inputs\.confirm_prefix\s*==\s*vars\.ANALYSIS_RESOURCE_PREFIX/);
  assert.match(cleanupWorkflow, /environment:\s*\n\s+name:\s*analysis-cleanup/);
  assert.match(cleanupWorkflow, /scripts\/analysis-cleanup\.mjs plan/);
  assert.match(cleanupWorkflow, /scripts\/analysis-cleanup\.mjs apply/);
  assert.match(cleanupWorkflow, /analysis-cleanup-plan/);
  assert.match(cleanupWorkflow, /--confirm-prefix/);
  assert.match(cleanupWorkflow, /GITHUB_STEP_SUMMARY/);
  assert.ok(
    cleanupWorkflow.indexOf('GITHUB_STEP_SUMMARY') <
    cleanupWorkflow.indexOf('name: analysis-cleanup'),
    'inventory must be summarized before the protected cleanup job is considered',
  );
  assert.doesNotMatch(
    cleanupWorkflow,
    /TerminateInstances|terminate-instances|DeregisterImage|deregister-image|DeleteStack|delete-stack|archive/i,
  );
});

test('documents cleanup prerequisites, destructive effects, allowlist, and future OIDC scope', () => {
  for (const document of [deploymentReadme, cleanupWorkflow]) {
    assert.match(document, /analysis-cleanup/);
  }
  assert.match(deploymentReadme, /Delete the analysis stack\s+first|stack must already be deleted/i);
  assert.match(deploymentReadme, /confirm_prefix/);
  assert.match(deploymentReadme, /exact\s+`ANALYSIS_RESOURCE_PREFIX`/i);
  assert.match(deploymentReadme, /exact `AnalysisStack=/i);
  assert.match(deploymentReadme, /S3 object versions/i);
  assert.match(deploymentReadme, /DynamoDB.*items/i);
  assert.match(deploymentReadme, /EBS data volume/i);
  assert.match(deploymentReadme, /tagged snapshots/i);
  assert.match(deploymentReadme, /least-privilege.*OIDC|OIDC.*least-privilege/i);
  assert.match(deploymentReadme, /analysis-cleanup.*protected GitHub Environment|protected `analysis-cleanup`/i);
});
