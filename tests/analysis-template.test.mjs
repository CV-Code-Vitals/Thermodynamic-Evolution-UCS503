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

test('configures GitHub OAuth without a Lambda/API Gateway dependency cycle', () => {
  assert.match(archiveTemplate, /GitHubCallbackUrl:\s*\r?\n\s+Type: String\s*\r?\n\s+Default: ''/);
  assert.match(archiveTemplate, /- !Not \[!Equals \[!Ref GitHubCallbackUrl, ''\]\]/);
  assert.match(archiveTemplate, /GITHUB_CALLBACK_URL: !If \[HasGitHubOAuth, !Ref GitHubCallbackUrl, ''\]/);
  assert.match(archiveTemplate, /GitHubOAuthCallbackUrl:[\s\S]*?Value: !Sub https:\/\/\$\{ServerApi\}\.execute-api\./);
  assert.doesNotMatch(archiveTemplate, /GITHUB_CALLBACK_URL: !Sub https:\/\/\$\{ServerApi\}/);
});

test('creates two private instances with split roles', () => {
  assert.equal((template.match(/Type: AWS::EC2::Instance/g) ?? []).length, 2);
  assert.match(resource('AnalysisInstance'), /IamInstanceProfile: !Ref AnalysisInstanceProfile/);
  assert.match(resource('SandboxInstance'), /IamInstanceProfile: !Ref SandboxInstanceProfile/);
  assert.match(resource('AnalysisInstanceProfile'), /- !Ref AnalysisInstanceRole/);
  assert.match(resource('SandboxInstanceProfile'), /- !Ref SandboxInstanceRole/);
  assert.match(resource('AnalysisVpcIpv6'), /AmazonProvidedIpv6CidrBlock: true/);
  assert.doesNotMatch(template, /Type: AWS::EC2::EgressOnlyInternetGateway/);
  assert.match(resource('PrivateS3GatewayEndpoint'), /VpcEndpointType: Gateway/);
  for (const name of ['AnalysisSubnet', 'SandboxSubnet']) {
    assert.match(resource(name), /MapPublicIpOnLaunch: false/);
  }
});

test('preserves pre-created instance network interfaces on termination', () => {
  for (const name of ['SandboxInstance', 'AnalysisInstance']) {
    assert.match(
      resource(name),
      /NetworkInterfaces:\s*\r?\n\s+- DeviceIndex: '0'\s*\r?\n\s+NetworkInterfaceId: !Ref [A-Za-z]+PrivateNetworkInterface\s*\r?\n\s+DeleteOnTermination: false/,
      `${name} must preserve its pre-created network interface when terminated`,
    );
  }
});

test('assigns distinct /64 IPv6 CIDRs after the Amazon-provided VPC block is associated', () => {
  assert.match(resource('AnalysisVpcIpv6'), /AmazonProvidedIpv6CidrBlock: true/);
  for (const name of ['AnalysisSubnet', 'SandboxSubnet']) {
    assert.match(resource(name), /DependsOn:\s*AnalysisVpcIpv6/);
    assert.match(resource(name), /Ipv6CidrBlock:\s*!Select/);
    assert.match(resource(name), /!GetAtt AnalysisVpc\.Ipv6CidrBlocks/);
    assert.match(resource(name), /- 256\s+- 64/);
    assert.match(resource(name), /AssignIpv6AddressOnCreation: true/);
  }
  assert.match(resource('AnalysisSubnet'), /Ipv6CidrBlock:\s*!Select\s+- 0/);
  assert.match(resource('SandboxSubnet'), /Ipv6CidrBlock:\s*!Select\s+- 1/);
  assert.doesNotMatch(template, /Type: AWS::EC2::SubnetCidrBlock/);
});

test('generates collision-resistant bucket and lifecycle-table names', () => {
  for (const name of ['SourceBucket', 'ArtifactBucket']) {
    assert.doesNotMatch(resource(name), /BucketName:/);
  }
  assert.doesNotMatch(resource('AnalysisLifecycleTable'), /TableName:/);
  assert.doesNotMatch(template, /^\s+SandboxLifecycleTableName:/m);
  assert.match(template, /SourceBucketName:\s*\r?\n\s+Value: !Ref SourceBucket/);
  assert.match(template, /ArtifactBucketName:\s*\r?\n\s+Value: !Ref ArtifactBucket/);
  assert.match(template, /SandboxLifecycleTable:\s*\r?\n\s+Value: !Ref AnalysisLifecycleTable/);
});

test('suffixes explicit Lambda and log-group names with the stack ID', () => {
  assert.ok(template.includes("!Select [4, !Split ['-', !Select [2, !Split ['/', !Ref AWS::StackId]]]]"));
  const resourcePrefixMaxLength = Number(/ResourcePrefix:[\s\S]*?MaxLength: (\d+)/.exec(template)?.[1]);
  assert.ok(Number.isInteger(resourcePrefixMaxLength));
  assert.ok(resourcePrefixMaxLength + '-analysis-authorizer-'.length + 12 <= 64);
  for (const name of ['ProxyFunction', 'AuthorizerFunction', 'SourceStagerFunction', 'LifecycleWatchdogFunction']) {
    assert.match(resource(name), /FunctionName: !Sub[\s\S]*?AWS::StackId/);
  }
  for (const name of [
    'AnalysisSsmSessionLogGroup',
    'ProxyLogGroup',
    'AuthorizerLogGroup',
    'SourceStagerLogGroup',
    'LifecycleWatchdogLogGroup',
    'AnalysisApiAccessLogGroup',
  ]) {
    assert.match(resource(name), /LogGroupName: !Sub[\s\S]*?AWS::StackId/);
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

test('uses private-DNS control-plane endpoints with the S3 gateway route', async () => {
  const proxy = await read('infra/analysis/stream-proxy/index.mjs');
  const lifecycle = await read('go-backend/sandbox_lifecycle.go');
  assert.match(resource('ProxyFunction'), /Ipv6AllowedForDualStack: true/);
  assert.match(proxy, /new LambdaClient\(\{\}\)/);
  assert.match(proxy, /new SSMClient\(\{\}\)/);
  assert.doesNotMatch(proxy, /useDualstackEndpoint/);
  assert.doesNotMatch(lifecycle, /WithUseDualStackEndpoint/);
  assert.match(resource('PrivateS3GatewayEndpoint'), /VpcEndpointType: Gateway/);
});

test('keeps private IPv6 subnets but removes public IPv6 egress in favor of AWS endpoints', () => {
  for (const name of [
    'ProxyIpv6HttpsEgress',
    'AnalysisIpv6HttpsEgress',
    'SandboxIpv6HttpsEgress',
  ]) {
    assert.doesNotMatch(template, new RegExp(`^  ${name}:`, 'm'));
  }
  assert.doesNotMatch(template, /Type: AWS::EC2::EgressOnlyInternetGateway/);
  assert.doesNotMatch(template, /DestinationIpv6CidrBlock:\s*::\/0/);

  for (const name of [
    'AnalysisToS3EndpointEgress',
    'SandboxToS3EndpointEgress',
    'AnalysisToDynamoDbEndpointEgress',
    'AnalysisToInterfaceEndpointsEgress',
    'SandboxToInterfaceEndpointsEgress',
    'ProxyToInterfaceEndpointsEgress',
  ]) {
    assert.ok(resource(name), `missing private endpoint egress rule ${name}`);
  }
  assert.match(resource('AnalysisToS3EndpointEgress'), /DestinationPrefixListId: !Ref S3ManagedPrefixListId/);
  assert.match(resource('SandboxToS3EndpointEgress'), /DestinationPrefixListId: !Ref S3ManagedPrefixListId/);
  assert.match(resource('AnalysisToDynamoDbEndpointEgress'), /DestinationPrefixListId: !Ref DynamoDbManagedPrefixListId/);
  assert.match(template, /S3ManagedPrefixListId:\s*\r?\n\s+Type: String\s*\r?\n\s+AllowedPattern: '\^pl-\[0-9a-f\]\+\$'/);
  assert.match(template, /DynamoDbManagedPrefixListId:\s*\r?\n\s+Type: String\s*\r?\n\s+AllowedPattern: '\^pl-\[0-9a-f\]\+\$'/);

  for (const [name, service] of [
    ['SsmEndpoint', 'ssm'],
    ['SsmMessagesEndpoint', 'ssmmessages'],
    ['Ec2MessagesEndpoint', 'ec2messages'],
    ['Ec2ApiEndpoint', 'ec2'],
    ['LambdaApiEndpoint', 'lambda'],
  ]) {
    assert.match(resource(name), /VpcEndpointType: Interface/);
    assert.ok(
      resource(name).includes(`com.amazonaws.\${AWS::Region}.${service}`),
      `${name} should target the ${service} endpoint service`,
    );
    assert.match(resource(name), /PrivateDnsEnabled: true/);
    assert.match(resource(name), /SecurityGroupIds:[\s\S]*?InterfaceEndpointSecurityGroup/);
  }
  assert.match(resource('PrivateDynamoDbGatewayEndpoint'), /ServiceName: !Sub 'com\.amazonaws\.\$\{AWS::Region\}\.dynamodb'/);
});

test('restricts VPC endpoint policies to the workload IAM roles', () => {
  for (const name of [
    'SsmEndpoint',
    'SsmMessagesEndpoint',
    'Ec2MessagesEndpoint',
    'Ec2ApiEndpoint',
    'LambdaApiEndpoint',
  ]) {
    const endpoint = resource(name);
    assert.match(endpoint, /PolicyDocument:/);
    assert.doesNotMatch(endpoint, /Principal:\s*['"]?\*['"]?/);
    assert.match(endpoint, /Principal:[\s\S]*?AWS:[\s\S]*?!GetAtt/);
  }
  for (const name of ['PrivateS3GatewayEndpoint', 'PrivateDynamoDbGatewayEndpoint']) {
    const endpoint = resource(name);
    assert.match(endpoint, /Principal: '\*'/);
    assert.match(endpoint, /aws:PrincipalArn:/);
  }
  assert.match(resource('PrivateS3GatewayEndpoint'), /!GetAtt AnalysisInstanceRole\.Arn/);
  assert.match(resource('PrivateS3GatewayEndpoint'), /!GetAtt SandboxInstanceRole\.Arn/);
  assert.match(resource('PrivateDynamoDbGatewayEndpoint'), /!GetAtt AnalysisInstanceRole\.Arn/);
  assert.match(resource('LambdaApiEndpoint'), /!GetAtt ProxyExecutionRole\.Arn/);
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
  for (const role of [analysisRole, watchdogRole]) {
    assert.match(role, /Sid: DescribeInstancesInStackRegion/);
    assert.match(role, /Action: ec2:DescribeInstances[\s\S]*?aws:RequestedRegion: !Ref AWS::Region/);
    assert.doesNotMatch(role, /ec2:ResourceTag\/AnalysisStack/);
  }
});

test('keeps analysis output snapshots for seven days', async () => {
  assert.match(resource('AnalysisDataVolume'), /VolumeType: gp3/);
  assert.match(resource('AnalysisDataVolume'), /Encrypted: true/);
  assert.match(resource('AnalysisDataVolume'), /DeletionPolicy: Retain/);
  assert.match(resource('AnalysisDataVolume'), /UpdateReplacePolicy: Retain/);
  assert.match(resource('AnalysisDataSnapshotPolicy'), /Description: [0-9A-Za-z _-]+(?=\r?\n)/);
  assert.match(resource('AnalysisDataSnapshotPolicy'), /Count: 7/);
  assert.match(resource('AnalysisDataSnapshotPolicy'), /Interval: 24/);
  assert.match(resource('AnalysisDataSnapshotPolicy'), /TargetTags:[\s\S]*?Key: AnalysisStack\s*\r?\n\s+Value: !Ref ResourcePrefix/);
  assert.match(resource('SnapshotLifecycleRole'), /ec2:ResourceTag\/AnalysisStack: !Ref ResourcePrefix/);
  for (const name of ['SandboxInstance', 'AnalysisInstance']) {
    assert.match(resource(name), /NetworkInterfaces:[\s\S]*?NetworkInterfaceId: !Ref [A-Za-z]+PrivateNetworkInterface/);
    assert.match(resource(name), /DeleteOnTermination: true/);
  }
  assert.match(resource('AnalysisDataVolumeAttachment'), /VolumeId: !Ref AnalysisDataVolume/);
  assert.match(resource('AnalysisInstance'), /install -d -m 0750 \/var\/lib\/ucs503-analysis\/uploads/);
  assert.match(await read('docker-compose.analysis.yml'), /ANALYSIS_UPLOADS_HOST_DIR[\s\S]*?:\/app\/uploads/);
});

test('expires sources after one day and bundles after 30 days', () => {
  assert.match(resource('SourceBucket'), /Prefix: snapshots\/[\s\S]*?ExpirationInDays: 1/);
  assert.match(resource('ArtifactBucket'), /Prefix: bundles\/[\s\S]*?ExpirationInDays: 30/);
});

test('does not grant unnecessary bucket enumeration through the private endpoint', () => {
  const endpoint = resource('PrivateS3GatewayEndpoint');
  assert.doesNotMatch(endpoint, /Action: s3:ListBucket/);
  assert.match(endpoint, /Action: s3:GetObject/);
  assert.match(endpoint, /SourceBucket\.Arn/);
  assert.match(endpoint, /ArtifactBucket\.Arn/);
});

test('runs the sandbox broker through the dedicated rootless Docker socket', async () => {
  const compose = await read('docker-compose.sandbox.yml');
  assert.match(compose, /user:\s*["']0:0["']/);
  assert.match(compose, /cap_drop:\s*\r?\n\s*- ALL/);
  assert.match(compose, /no-new-privileges:true/);
  const exampleEnv = await read('.env.sandbox.example');
  const deploymentWorkflow = await read('.github/workflows/deploy-analysis.yml');

  assert.match(compose, /DOCKER_HOST:\s*unix:\/\/\/run\/docker\.sock/);
  assert.match(compose, /source:\s*\$\{DOCKER_SOCKET_SOURCE:[\s\S]*?target: \/run\/docker\.sock/);
  assert.doesNotMatch(compose, /DOCKER_GID|\/var\/run\/docker\.sock/);
  assert.match(exampleEnv, /^DOCKER_SOCKET_SOURCE=\/run\/user\/10002\/docker\.sock$/m);
  assert.match(deploymentWorkflow, /export DOCKER_HOST=unix:\/\/\/run\/user\/10002\/docker\.sock/);

});

test('configures anonymous scan and protected streaming Evolution routes', () => {
  assert.match(template, /FrontendOrigin:[\s\S]*?Default: https:\/\/cv-code-vitals\.github\.io/);
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
  ]) {
    await access(path.join(root, file), constants.F_OK);
  }
  assert.match(workflow, /bundles\/analysis-deployments\/\$IMAGE_SHA\/\$BUNDLE_SHA256\/deployment-bundle\.tar\.gz/);
  assert.match(resource('AnalysisInstanceRole'), /bundles\/\*/);
  assert.match(resource('SandboxInstanceRole'), /bundles\/analysis-deployments\/\*/);
});
