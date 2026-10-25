import assert from 'node:assert/strict';
import { access, readFile } from 'node:fs/promises';
import { constants } from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const read = (relativePath) => readFile(path.join(root, relativePath), 'utf8');
const template = await read('infra/analysis/template.yaml');
const archiveTemplate = await read('template.yaml');
const workflow = await read('.github/workflows/deploy-analysis.yml');

function resource(name) {
  const marker = new RegExp(`^  ${name}:\\r?\\n`, 'm').exec(template);
  assert.ok(marker, `template is missing resource ${name}`);
  const contentStart = marker.index + marker[0].length;
  const nextResource = template.slice(contentStart).search(/^  [A-Z][A-Za-z0-9]*:\r?$|^Outputs:\r?$/m);
  return template.slice(contentStart, nextResource === -1 ? undefined : contentStart + nextResource);
}

function decryptStatements(source) {
  return source
    .split(/(?=\r?\n\s+- Sid:)/)
    .filter((statement) => /Action:\s*(?:kms:Decrypt|\r?\n\s+- kms:Decrypt)/.test(statement));
}

test('scopes SecureString decryption to the SSM KMS key alias', () => {
  const statements = [
    ...decryptStatements(template),
    ...decryptStatements(archiveTemplate),
  ];
  assert.equal(statements.length, 5, 'expected all four analysis roles and the archive role');
  for (const statement of statements) {
    assert.match(statement, /kms:ViaService:/);
    assert.match(statement, /kms:EncryptionContext:PARAMETER_ARN:/);
    assert.match(statement, /kms:ResourceAliases:\s*\r?\n\s+- alias\/aws\/ssm/);
  }
});

test('creates two private instances with split roles', () => {
  assert.equal((template.match(/Type: AWS::EC2::Instance/g) ?? []).length, 2);
  assert.match(resource('AnalysisInstance'), /IamInstanceProfile: !Ref AnalysisInstanceProfile/);
  assert.match(resource('SandboxInstance'), /IamInstanceProfile: !Ref SandboxInstanceProfile/);
  assert.match(resource('AnalysisInstanceProfile'), /- !Ref AnalysisInstanceRole/);
  assert.match(resource('SandboxInstanceProfile'), /- !Ref SandboxInstanceRole/);
  assert.match(resource('AnalysisVpcIpv6'), /AmazonProvidedIpv6CidrBlock: true/);
  assert.match(resource('EgressOnlyInternetGateway'), /Type: AWS::EC2::EgressOnlyInternetGateway/);
  assert.match(resource('PrivateS3GatewayEndpoint'), /VpcEndpointType: Gateway/);
  for (const name of ['AnalysisSubnet', 'SandboxSubnet']) {
    assert.match(resource(name), /MapPublicIpOnLaunch: false/);
  }
});

test('blocks public and SSH ingress', () => {
  assert.doesNotMatch(template, /AssociatePublicIpAddress:\s*true/);
  assert.doesNotMatch(template, /FromPort:\s*22\b/);
  assert.match(resource('AnalysisFromProxyIngress'), /SourceSecurityGroupId: !Ref ProxySecurityGroup/);
  assert.match(resource('SandboxFromAnalysisIngress'), /SourceSecurityGroupId: !Ref AnalysisSecurityGroup/);
  assert.doesNotMatch(resource('AnalysisSecurityGroup'), /CidrIp: 0\.0\.0\.0\/0/);
  assert.doesNotMatch(resource('SandboxSecurityGroup'), /CidrIp: 0\.0\.0\.0\/0/);
});

test('uses dual-stack control-plane endpoints with the S3 gateway route', async () => {
  const proxy = await read('infra/analysis/stream-proxy/index.mjs');
  const lifecycle = await read('go-backend/sandbox_lifecycle.go');
  assert.match(resource('ProxyFunction'), /Ipv6AllowedForDualStack: true/);
  assert.match(proxy, /new LambdaClient\(\{ useDualstackEndpoint: true \}\)/);
  assert.match(proxy, /new SSMClient\(\{ useDualstackEndpoint: true \}\)/);
  assert.match(lifecycle, /WithUseDualStackEndpoint\(aws\.DualStackEndpointStateEnabled\)/);
  assert.match(resource('PrivateS3GatewayEndpoint'), /VpcEndpointType: Gateway/);
});

test('limits token permissions by role', () => {
  const proxyRole = resource('ProxyExecutionRole');
  const authorizerRole = resource('AuthorizerExecutionRole');
  const analysisRole = resource('AnalysisInstanceRole');
  const sandboxRole = resource('SandboxInstanceRole');
  const watchdogRole = resource('LifecycleWatchdogRole');
  assert.match(proxyRole, /parameter\$\{ApiTokenParameterName\}/);
  assert.doesNotMatch(proxyRole, /SigningKeyParameterName/);
  assert.match(authorizerRole, /SigningKeyParameterName/);
  assert.match(authorizerRole, /ApiTokenParameterName/);
  assert.match(analysisRole, /ApiTokenParameterName/);
  assert.match(analysisRole, /SandboxTokenParameterName/);
  assert.doesNotMatch(analysisRole, /SigningKeyParameterName/);
  assert.match(sandboxRole, /SandboxTokenParameterName/);
  assert.doesNotMatch(sandboxRole, /SigningKeyParameterName|ApiTokenParameterName/);
  assert.match(sandboxRole, /bundles\/analysis-deployments\/\*/);
  assert.match(watchdogRole, /ec2:StopInstances/);
  assert.doesNotMatch(watchdogRole, /ec2:StartInstances/);
});

test('keeps analysis output snapshots for seven days', async () => {
  assert.match(resource('AnalysisDataVolume'), /VolumeType: gp3/);
  assert.match(resource('AnalysisDataVolume'), /Encrypted: true/);
  assert.match(resource('AnalysisDataSnapshotPolicy'), /Count: 7/);
  assert.match(resource('AnalysisDataSnapshotPolicy'), /Interval: 24/);
  assert.match(resource('AnalysisDataVolumeAttachment'), /VolumeId: !Ref AnalysisDataVolume/);
  assert.match(resource('AnalysisInstance'), /install -d -m 0750 \/var\/lib\/ucs503-analysis\/uploads/);
  assert.match(await read('docker-compose.analysis.yml'), /ANALYSIS_UPLOADS_HOST_DIR[\s\S]*?:\/app\/uploads/);
});

test('expires sources after one day and bundles after 30 days', () => {
  assert.match(resource('SourceBucket'), /Prefix: snapshots\/[\s\S]*?ExpirationInDays: 1/);
  assert.match(resource('ArtifactBucket'), /Prefix: bundles\/[\s\S]*?ExpirationInDays: 30/);
});

test('runs the sandbox broker through the dedicated rootless Docker socket', async () => {
  const compose = await read('docker-compose.sandbox.yml');
  const exampleEnv = await read('.env.sandbox.example');
  const deploymentWorkflow = await read('.github/workflows/deploy-analysis.yml');
  const deploymentReadme = await read('README_DEPLOY.md');
  const analysisReadme = await read('infra/analysis/README.md');

  assert.match(compose, /DOCKER_HOST:\s*unix:\/\/\/run\/docker\.sock/);
  assert.match(compose, /source:\s*\$\{DOCKER_SOCKET_SOURCE:[\s\S]*?target: \/run\/docker\.sock/);
  assert.doesNotMatch(compose, /DOCKER_GID|\/var\/run\/docker\.sock/);
  assert.match(exampleEnv, /^DOCKER_SOCKET_SOURCE=\/run\/user\/10002\/docker\.sock$/m);
  assert.match(deploymentWorkflow, /export DOCKER_HOST=unix:\/\/\/run\/user\/10002\/docker\.sock/);
  for (const readme of [deploymentReadme, analysisReadme]) {
    assert.match(readme, /rootless Docker/i);
    assert.match(readme, /10002/);
    assert.match(readme, /linger/i);
  }
});

test('configures anonymous scan and protected streaming Evolution routes', () => {
  assert.match(resource('ScanApiMethod'), /AuthorizationType: NONE/);
  assert.match(resource('ScanApiMethod'), /ResponseTransferMode: STREAM/);
  assert.match(resource('EvolveApiMethod'), /AuthorizationType: CUSTOM/);
  assert.match(resource('EvolveApiMethod'), /AuthorizerId: !Ref AnalysisAuthorizer/);
  assert.match(resource('EvolveApiMethod'), /ResponseTransferMode: STREAM/);
  assert.match(resource('AnalysisAuthorizer'), /AuthorizerResultTtlInSeconds: 0/);
  for (const name of ['ScanOptionsMethod', 'EvolveOptionsMethod']) {
    assert.match(resource(name), /HttpMethod: OPTIONS/);
    assert.match(resource(name), /AuthorizationType: NONE/);
    assert.match(resource(name), /Access-Control-Allow-Origin/);
  }
  assert.match(resource('AnalysisApiDefault4xxCors'), /ResponseType: DEFAULT_4XX/);
  assert.match(resource('AnalysisApiDefault5xxCors'), /ResponseType: DEFAULT_5XX/);
  assert.match(resource('ProxyFunction'), /FRONTEND_ORIGIN: !Ref FrontendOrigin/);
});

test('omits credentials from logs and archive resources', () => {
  const accessLog = resource('AnalysisApiStage');
  assert.match(accessLog, /AccessLogSetting:/);
  assert.doesNotMatch(accessLog, /Authorization|authorizer|requestBody|responseBody/i);
  assert.doesNotMatch(accessLog, /DataTraceEnabled: true/);
  assert.doesNotMatch(template, /AWS::RDS::DBInstance|ucs503-admin-archive-v2|ucs503-api/);
});

test('provides the deployment support files and artifact prefix contract', async () => {
  for (const file of [
    'infra/analysis/samconfig.toml',
    'infra/analysis/parameters.example.json',
    'infra/analysis/README.md',
  ]) {
    await access(path.join(root, file), constants.F_OK);
  }
  assert.match(workflow, /bundles\/analysis-deployments\/\$IMAGE_SHA\/deployment-bundle\.tar\.gz/);
  assert.match(resource('AnalysisInstanceRole'), /bundles\/\*/);
  assert.match(resource('SandboxInstanceRole'), /bundles\/analysis-deployments\/\*/);
});
