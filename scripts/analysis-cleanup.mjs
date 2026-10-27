import { spawnSync } from 'node:child_process';
import { mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

const OWNERSHIP_TAG = 'AnalysisStack';
const FORMAT_VERSION = 1;
const MAX_DELETE_BATCH = 1000;

function fail(message) {
  throw new Error(message);
}

function parseArgs(args) {
  if (!['plan', 'apply'].includes(args[0])) {
    fail('Usage: analysis-cleanup.mjs plan|apply [options]');
  }
  const mode = args[0];
  const options = {};
  for (let index = 1; index < args.length; index += 2) {
    const key = args[index];
    const value = args[index + 1];
    if (!key?.startsWith('--') || value === undefined || value.startsWith('--')) {
      fail(`Invalid option near ${key ?? 'end of arguments'}`);
    }
    const name = key.slice(2);
    if (Object.hasOwn(options, name)) fail(`Duplicate option: ${key}`);
    options[name] = value;
  }
  const required = mode === 'plan'
    ? ['stack-name', 'prefix', 'region']
    : ['plan', 'confirm-prefix', 'region'];
  for (const name of required) {
    if (!options[name]) fail(`Missing required option: --${name}`);
  }
  const allowed = mode === 'plan'
    ? new Set(['stack-name', 'prefix', 'region'])
    : new Set(['plan', 'confirm-prefix', 'region']);
  for (const name of Object.keys(options)) {
    if (!allowed.has(name)) fail(`Unknown option: --${name}`);
  }
  return { mode, options };
}

function validatePrefix(prefix) {
  if (!/^[a-z0-9][a-z0-9-]{2,62}$/.test(prefix)) {
    fail('Resource prefix must be 3-63 lowercase letters, digits, or hyphens and start with a letter or digit');
  }
}

function validateStackName(stackName) {
  if (!/^[A-Za-z][A-Za-z0-9-]{0,127}$/.test(stackName)) {
    fail('Invalid CloudFormation stack name');
  }
}

function validateRegion(region) {
  if (!/^[a-z0-9-]{3,64}$/.test(region)) fail('Invalid AWS region');
}

function parseJson(output, description) {
  try {
    return JSON.parse(output);
  } catch {
    fail(`AWS CLI returned invalid JSON for ${description}`);
  }
}

function awsProcessEnvironment(testMode) {
  const environment = {};
  const allowedNames = [
    'PATH',
    'SystemRoot',
    'WINDIR',
    'TEMP',
    'TMP',
    'AWS_ACCESS_KEY_ID',
    'AWS_SECRET_ACCESS_KEY',
    'AWS_SESSION_TOKEN',
    'AWS_SECURITY_TOKEN',
  ];
  for (const name of allowedNames) {
    if (process.env[name] !== undefined) environment[name] = process.env[name];
  }
  if (testMode) {
    for (const name of ['ANALYSIS_CLEANUP_MOCK_SCENARIO', 'ANALYSIS_CLEANUP_MOCK_LOG']) {
      if (process.env[name] !== undefined) environment[name] = process.env[name];
    }
  }
  return environment;
}

function aws(region, service, operation, args = []) {
  const testMode = process.env.NODE_TEST_CONTEXT !== undefined &&
    process.env.ANALYSIS_CLEANUP_TEST_MODE === '1';
  const configuredCli = testMode ? process.env.ANALYSIS_CLEANUP_AWS_CLI : undefined;
  if (configuredCli && !/\.mjs$/i.test(configuredCli)) {
    fail('The test AWS CLI override must be an .mjs module');
  }
  const cli = configuredCli || 'aws';
  const isNodeScript = /\.mjs$/i.test(cli);
  const command = isNodeScript ? process.execPath : cli;
  const commandArgs = [
    ...(isNodeScript ? [cli] : []),
    service,
    operation,
    ...args,
    '--region',
    region,
    '--output',
    'json',
  ];
  const result = spawnSync(command, commandArgs, {
    encoding: 'utf8',
    maxBuffer: 32 * 1024 * 1024,
    env: awsProcessEnvironment(testMode),
  });
  if (result.error) throw new Error(`AWS CLI could not run ${service} ${operation}: ${result.error.message}`);
  if (result.status !== 0) {
    const detail = (result.stderr || result.stdout || `exit status ${result.status}`).trim();
    const error = new Error(`AWS ${service} ${operation} failed: ${detail}`);
    error.awsError = detail;
    throw error;
  }
  return parseJson(result.stdout || '{}', `${service} ${operation}`);
}

function resourceHasOwnershipTag(tags, prefix) {
  return Array.isArray(tags) && tags.some(
    (tag) => tag.Key === OWNERSHIP_TAG && tag.Value === prefix,
  );
}

function logGroupAllowed(logGroupName, prefix, stackSuffix) {
  const stems = [
    `/aws/ssm/sessions/${prefix}`,
    `/aws/lambda/${prefix}-analysis-proxy`,
    `/aws/lambda/${prefix}-analysis-authorizer`,
    `/aws/lambda/${prefix}-source-stager`,
    `/aws/lambda/${prefix}-sandbox-watchdog`,
    `/aws/apigateway/${prefix}-analysis`,
  ];
  return stems.some((stem) =>
    logGroupName === stem ||
    (stackSuffix !== null && logGroupName === `${stem}-${stackSuffix}`));
}

function stackSuffixFromId(stackId, stackName, region) {
  if (typeof stackId !== 'string') return null;
  const escaped = (value) => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const pattern = new RegExp(
    `^arn:[^:]+:cloudformation:${escaped(region)}:\\d{12}:stack/${escaped(stackName)}/[a-f0-9]{8}(?:-[a-f0-9]{4}){3}-([a-f0-9]{12})$`,
  );
  return pattern.exec(stackId)?.[1] ?? null;
}

function normalizeResource(resource) {
  if (!resource || typeof resource !== 'object' || Array.isArray(resource)) {
    fail('Malformed plan resource');
  }
  const assertKeys = (expected) => {
    if (JSON.stringify(Object.keys(resource).sort()) !== JSON.stringify([...expected].sort())) {
      fail('Plan resource has an unexpected schema');
    }
  };
  switch (resource?.type) {
    case 'snapshot':
      assertKeys(['type', 'id']);
      if (typeof resource.id !== 'string' || resource.id.length === 0) fail('Malformed plan resource identifier');
      return { type: resource.type, id: resource.id };
    case 'bucket':
      assertKeys(['type', 'id', 'creationDate']);
      if (typeof resource.id !== 'string' || resource.id.length === 0 ||
        typeof resource.creationDate !== 'string' || resource.creationDate.length === 0) {
        fail('Malformed bucket identity in plan');
      }
      return { type: resource.type, id: resource.id, creationDate: resource.creationDate };
    case 'table':
      assertKeys(['type', 'id', 'tableId']);
      if (typeof resource.id !== 'string' || resource.id.length === 0 ||
        typeof resource.tableId !== 'string' || resource.tableId.length === 0) {
        fail('Malformed DynamoDB table identity in plan');
      }
      return { type: resource.type, id: resource.id, tableId: resource.tableId };
    case 'log-group':
      assertKeys(['type', 'id', 'creationTime']);
      if (typeof resource.id !== 'string' || resource.id.length === 0 ||
        !Number.isSafeInteger(resource.creationTime) || resource.creationTime < 0) {
        fail('Malformed log-group identity in plan');
      }
      return { type: resource.type, id: resource.id, creationTime: resource.creationTime };
    case 'volume':
      assertKeys(['type', 'id', 'attachments']);
      if (typeof resource.id !== 'string' || resource.id.length === 0 || !Array.isArray(resource.attachments)) {
        fail('Malformed volume resource in plan');
      }
      if (!resource.attachments.every((attachment) =>
        attachment &&
        typeof attachment === 'object' &&
        !Array.isArray(attachment) &&
        JSON.stringify(Object.keys(attachment).sort()) === JSON.stringify(['device', 'instanceId', 'state']) &&
        typeof attachment.instanceId === 'string' &&
        typeof attachment.state === 'string' &&
        (attachment.device === null || typeof attachment.device === 'string'))) {
        fail('Malformed volume attachments in plan');
      }
      return {
        type: 'volume',
        id: resource.id,
        attachments: resource.attachments.map((attachment) => ({
          instanceId: attachment.instanceId,
          state: attachment.state,
          device: attachment.device,
        })),
      };
    default:
      fail('Plan contains an unsupported resource type');
  }
}

function normalizePlan(plan) {
  if (!plan || typeof plan !== 'object' || Array.isArray(plan)) fail('Cleanup plan must be a JSON object');
  const keys = Object.keys(plan).sort();
  if (JSON.stringify(keys) !== JSON.stringify(['formatVersion', 'prefix', 'region', 'resources', 'stackId', 'stackName'].sort())) {
    fail('Cleanup plan has an unexpected schema');
  }
  if (plan.formatVersion !== FORMAT_VERSION || !Array.isArray(plan.resources)) {
    fail('Unsupported or malformed cleanup plan');
  }
  validatePrefix(plan.prefix);
  validateStackName(plan.stackName);
  validateRegion(plan.region);
  if (plan.stackId !== null && (typeof plan.stackId !== 'string' || plan.stackId.length === 0)) {
    fail('Malformed CloudFormation stack identity in plan');
  }
  const resources = plan.resources.map(normalizeResource);
  resources.sort(compareResources);
  return {
    formatVersion: FORMAT_VERSION,
    stackName: plan.stackName,
    stackId: plan.stackId,
    prefix: plan.prefix,
    region: plan.region,
    resources,
  };
}

function compareResources(left, right) {
  return left.type.localeCompare(right.type) || left.id.localeCompare(right.id);
}

function buildPlan(stackName, stackId, prefix, region, resources) {
  const ordered = resources.map(normalizeResource).sort(compareResources);
  return { formatVersion: FORMAT_VERSION, stackName, stackId, prefix, region, resources: ordered };
}

function canonical(plan) {
  return JSON.stringify(normalizePlan(plan));
}

function assertStackDeleted(stackName, region, expectedStackId) {
  let stackId = null;
  try {
    const response = aws(region, 'cloudformation', 'describe-stacks', ['--stack-name', stackName]);
    const stack = response.Stacks?.find((item) => item.StackName === stackName);
    if (!stack) fail(`CloudFormation stack ${stackName} was not found in the describe response`);
    if (stack.StackStatus !== 'DELETE_COMPLETE') {
      fail(`CloudFormation stack ${stackName} still exists in state ${stack.StackStatus}`);
    }
    stackId = typeof stack.StackId === 'string' && stack.StackId.length > 0 ? stack.StackId : null;
  } catch (error) {
    if (!error.awsError || !/does not exist|not exist|doesn't exist/i.test(error.awsError)) throw error;
  }
  if (expectedStackId !== undefined && stackId !== expectedStackId) {
    fail(`CloudFormation stack identity changed for ${stackName}`);
  }
  return stackId;
}

function getBucketTags(bucket, region) {
  try {
    const response = aws(region, 's3api', 'get-bucket-tagging', ['--bucket', bucket]);
    return response.TagSet ?? [];
  } catch (error) {
    if (/NoSuchTagSet/i.test(error.awsError || error.message)) return [];
    throw error;
  }
}

function getBucketInfo(bucketName, region) {
  const response = aws(region, 's3api', 'list-buckets');
  const matches = (response.Buckets ?? []).filter((bucket) => bucket.Name === bucketName);
  if (matches.length !== 1 || typeof matches[0].CreationDate !== 'string') {
    fail(`S3 bucket identity changed for ${bucketName}`);
  }
  return { bucket: matches[0], tags: getBucketTags(bucketName, region) };
}

function getTableInfo(tableName, region) {
  const tableResponse = aws(region, 'dynamodb', 'describe-table', ['--table-name', tableName]);
  const table = tableResponse.Table;
  if (!table?.TableArn || !table.TableId || table.TableName !== tableName) {
    fail(`DynamoDB table identity changed for ${tableName}`);
  }
  const tagsResponse = aws(region, 'dynamodb', 'list-tags-of-resource', ['--resource-arn', table.TableArn]);
  return { table, tags: tagsResponse.Tags ?? [] };
}

function getVolume(volumeId, region) {
  const response = aws(region, 'ec2', 'describe-volumes', ['--volume-ids', volumeId]);
  const matches = (response.Volumes ?? []).filter((volume) => volume.VolumeId === volumeId);
  if (matches.length !== 1) fail(`EBS volume identity changed for ${volumeId}`);
  return matches[0];
}

function getSnapshot(snapshotId, region) {
  const response = aws(region, 'ec2', 'describe-snapshots', ['--snapshot-ids', snapshotId, '--owner-ids', 'self']);
  const matches = (response.Snapshots ?? []).filter((snapshot) => snapshot.SnapshotId === snapshotId);
  if (matches.length !== 1) fail(`EBS snapshot identity changed for ${snapshotId}`);
  return matches[0];
}

function getLogGroup(logGroupName, prefix, region, stackSuffix) {
  if (!logGroupAllowed(logGroupName, prefix, stackSuffix)) fail(`Log group is outside the cleanup allowlist: ${logGroupName}`);
  const response = aws(region, 'logs', 'describe-log-groups', [
    '--log-group-name-prefix',
    logGroupName,
  ]);
  const matches = (response.logGroups ?? []).filter((group) => group.logGroupName === logGroupName);
  if (matches.length !== 1) fail(`CloudWatch log group identity changed for ${logGroupName}`);
  if (!Number.isSafeInteger(matches[0].creationTime) || matches[0].creationTime < 0) {
    fail(`CloudWatch log group has no valid creation time: ${logGroupName}`);
  }
  return matches[0];
}

function assertOwnedResource(resource, prefix, region, stackId, stackName) {
  switch (resource.type) {
    case 'bucket':
      const { bucket, tags } = getBucketInfo(resource.id, region);
      if (bucket.CreationDate !== resource.creationDate) {
        fail(`S3 bucket identity changed for ${resource.id}`);
      }
      if (!resourceHasOwnershipTag(tags, prefix)) {
        fail(`Bucket ownership tag is missing or changed: ${resource.id}`);
      }
      break;
    case 'table': {
      const { table, tags } = getTableInfo(resource.id, region);
      if (table.TableId !== resource.tableId) fail(`DynamoDB table identity changed for ${resource.id}`);
      if (!resourceHasOwnershipTag(tags, prefix)) fail(`DynamoDB table ownership tag is missing or changed: ${resource.id}`);
      break;
    }
    case 'volume': {
      const volume = getVolume(resource.id, region);
      if (!resourceHasOwnershipTag(volume.Tags, prefix)) fail(`EBS volume ownership tag is missing or changed: ${resource.id}`);
      if ((volume.Attachments ?? []).length !== 0) fail(`Refusing to delete attached EBS volume ${resource.id}`);
      break;
    }
    case 'snapshot': {
      const snapshot = getSnapshot(resource.id, region);
      if (!resourceHasOwnershipTag(snapshot.Tags, prefix)) fail(`EBS snapshot ownership tag is missing or changed: ${resource.id}`);
      break;
    }
    case 'log-group':
      if (getLogGroup(
        resource.id,
        prefix,
        region,
        stackSuffixFromId(stackId, stackName, region),
      ).creationTime !== resource.creationTime) {
        fail(`CloudWatch log group identity changed for ${resource.id}`);
      }
      break;
    default:
      fail('Unsupported resource in cleanup plan');
  }
}

function inventory(stackName, prefix, region, expectedStackId) {
  const stackId = assertStackDeleted(stackName, region, expectedStackId);
  const stackSuffix = stackSuffixFromId(stackId, stackName, region);
  const resources = [];

  const bucketResponse = aws(region, 's3api', 'list-buckets');
  for (const bucket of bucketResponse.Buckets ?? []) {
    if (!bucket.Name) continue;
    if (typeof bucket.CreationDate !== 'string') fail(`S3 bucket has no creation date: ${bucket.Name}`);
    if (resourceHasOwnershipTag(getBucketTags(bucket.Name, region), prefix)) {
      resources.push({ type: 'bucket', id: bucket.Name, creationDate: bucket.CreationDate });
    }
  }

  const tableNames = aws(region, 'dynamodb', 'list-tables').TableNames ?? [];
  for (const tableName of tableNames) {
    const { table, tags } = getTableInfo(tableName, region);
    if (resourceHasOwnershipTag(tags, prefix)) {
      resources.push({ type: 'table', id: tableName, tableId: table.TableId });
    }
  }

  const volumeResponse = aws(region, 'ec2', 'describe-volumes', []);
  for (const volume of volumeResponse.Volumes ?? []) {
    if (volume.VolumeId && resourceHasOwnershipTag(volume.Tags, prefix)) {
      resources.push({
        type: 'volume',
        id: volume.VolumeId,
        attachments: (volume.Attachments ?? []).map((attachment) => ({
          instanceId: attachment.InstanceId,
          state: attachment.State,
          device: attachment.Device ?? null,
        })).sort((left, right) => left.instanceId.localeCompare(right.instanceId)),
      });
    }
  }

  const snapshotResponse = aws(region, 'ec2', 'describe-snapshots', ['--owner-ids', 'self']);
  for (const snapshot of snapshotResponse.Snapshots ?? []) {
    if (snapshot.SnapshotId && resourceHasOwnershipTag(snapshot.Tags, prefix)) {
      resources.push({ type: 'snapshot', id: snapshot.SnapshotId });
    }
  }

  const logResponse = aws(region, 'logs', 'describe-log-groups', ['--log-group-name-prefix', '/aws/']);
  for (const group of logResponse.logGroups ?? []) {
    if (group.logGroupName && logGroupAllowed(group.logGroupName, prefix, stackSuffix)) {
      if (!Number.isSafeInteger(group.creationTime) || group.creationTime < 0) {
        fail(`CloudWatch log group has no valid creation time: ${group.logGroupName}`);
      }
      resources.push({
        type: 'log-group',
        id: group.logGroupName,
        creationTime: group.creationTime,
      });
    }
  }

  return buildPlan(stackName, stackId, prefix, region, resources);
}

function formatInventory(plan) {
  const lines = [
    `Analysis cleanup inventory for ${plan.prefix} (${plan.region})`,
    `CloudFormation stack: ${plan.stackName} (not present or DELETE_COMPLETE)`,
    `Resources: ${plan.resources.length}`,
  ];
  if (plan.resources.length === 0) lines.push('  (none)');
  for (const resource of plan.resources) {
    const attachmentNote = resource.type === 'volume' && resource.attachments.length
      ? ` [ATTACHED: ${resource.attachments.map((item) => item.instanceId).join(', ')}]`
      : '';
    lines.push(`  ${resource.type}: ${resource.id}${attachmentNote}`);
  }
  return `${lines.join('\n')}\n`;
}

async function planMode(options) {
  validatePrefix(options.prefix);
  validateStackName(options['stack-name']);
  validateRegion(options.region);
  const plan = inventory(options['stack-name'], options.prefix, options.region);
  process.stderr.write(formatInventory(plan));
  process.stdout.write(`${JSON.stringify(plan, null, 2)}\n`);
}

async function withDeleteFile(payload, stackName, stackId, prefix, region, resource) {
  const directory = await mkdtemp(path.join(process.cwd(), '.analysis-cleanup-'));
  const filename = path.join(directory, 'delete.json');
  try {
    await writeFile(filename, JSON.stringify(payload), { mode: 0o600 });
    assertStackDeleted(stackName, region, stackId);
    assertOwnedResource(resource, prefix, region, stackId, stackName);
    const fileUri = pathToFileURL(filename).href;
    const response = aws(region, 's3api', 'delete-objects', [
      '--bucket',
      resource.id,
      '--delete',
      fileUri,
    ]);
    if ((response.Errors ?? []).length) {
      fail(`S3 reported errors deleting objects from ${resource.id}: ${JSON.stringify(response.Errors)}`);
    }
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
}

async function emptyBucket(resource, stackName, stackId, prefix, region) {
  const bucket = resource.id;
  let keyMarker;
  let versionIdMarker;
  const seenMarkers = new Set();
  while (true) {
    const args = ['--bucket', bucket, '--no-paginate'];
    if (keyMarker !== undefined) args.push('--key-marker', keyMarker);
    if (versionIdMarker !== undefined) args.push('--version-id-marker', versionIdMarker);
    const page = aws(region, 's3api', 'list-object-versions', args);
    const objects = [
      ...(page.Versions ?? []),
      ...(page.DeleteMarkers ?? []),
    ].map((item) => ({ Key: item.Key, VersionId: item.VersionId }));
    for (let index = 0; index < objects.length; index += MAX_DELETE_BATCH) {
      const batch = objects.slice(index, index + MAX_DELETE_BATCH);
      await withDeleteFile({ Objects: batch, Quiet: true }, stackName, stackId, prefix, region, resource);
    }
    if (!page.IsTruncated) break;
    if (typeof page.NextKeyMarker !== 'string' || page.NextKeyMarker.length === 0) {
      fail(`S3 returned a truncated version listing without a next marker for ${bucket}`);
    }
    const nextKeyMarker = page.NextKeyMarker;
    const nextVersionIdMarker = page.NextVersionIdMarker;
    const markerKey = `${nextKeyMarker}\u0000${nextVersionIdMarker ?? ''}`;
    if (seenMarkers.has(markerKey)) fail(`S3 version pagination did not advance for ${bucket}`);
    seenMarkers.add(markerKey);
    keyMarker = nextKeyMarker;
    versionIdMarker = nextVersionIdMarker;
  }
  assertStackDeleted(stackName, region, stackId);
  assertOwnedResource(resource, prefix, region, stackId, stackName);
  aws(region, 's3api', 'delete-bucket', ['--bucket', bucket]);
}

function deleteResource(resource, stackName, stackId, prefix, region) {
  assertStackDeleted(stackName, region, stackId);
  assertOwnedResource(resource, prefix, region, stackId, stackName);
  switch (resource.type) {
    case 'bucket':
      return emptyBucket(resource, stackName, stackId, prefix, region);
    case 'table':
      aws(region, 'dynamodb', 'delete-table', ['--table-name', resource.id]);
      break;
    case 'volume':
      aws(region, 'ec2', 'delete-volume', ['--volume-id', resource.id]);
      break;
    case 'snapshot':
      aws(region, 'ec2', 'delete-snapshot', ['--snapshot-id', resource.id]);
      break;
    case 'log-group':
      aws(region, 'logs', 'delete-log-group', ['--log-group-name', resource.id]);
      break;
    default:
      fail('Unsupported resource in cleanup plan');
  }
}

async function applyMode(options) {
  validateRegion(options.region);
  let parsedPlan;
  try {
    parsedPlan = JSON.parse(await readFile(options.plan, 'utf8'));
  } catch (error) {
    fail(`Unable to read cleanup plan: ${error.message}`);
  }
  const plan = normalizePlan(parsedPlan);
  if (options['confirm-prefix'] !== plan.prefix) fail('Confirmation prefix must exactly match the plan prefix');
  if (options.region !== plan.region) fail('Requested region does not match the cleanup plan region');

  const freshPlan = inventory(plan.stackName, plan.prefix, plan.region, plan.stackId);
  if (canonical(freshPlan) !== canonical(plan)) {
    fail('Cleanup plan inventory changed; produce a new read-only plan before applying');
  }
  const attachedVolume = plan.resources.find(
    (resource) => resource.type === 'volume' && resource.attachments.length > 0,
  );
  if (attachedVolume) fail(`Refusing to apply a plan containing attached EBS volume ${attachedVolume.id}`);
  for (const resource of plan.resources) {
    await deleteResource(resource, plan.stackName, plan.stackId, plan.prefix, plan.region);
  }
  process.stdout.write(`Cleanup complete for ${plan.prefix}; removed ${plan.resources.length} resources.\n`);
}

try {
  const { mode, options } = parseArgs(process.argv.slice(2));
  if (mode === 'plan') await planMode(options);
  else await applyMode(options);
} catch (error) {
  process.stderr.write(`Analysis cleanup refused: ${error.message}\n`);
  process.exitCode = 1;
}
