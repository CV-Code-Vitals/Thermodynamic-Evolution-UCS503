import assert from 'node:assert/strict';
import { mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const utility = path.join(root, 'scripts', 'analysis-cleanup.mjs');
const prefix = 'ucs503-analysis-v1';
const stackName = 'ucs503-analysis-v1';
const region = 'ap-south-1';
const stackSuffix = '0123456789ab';
const otherStackSuffix = 'fedcba987654';
const stackId = `arn:aws:cloudformation:${region}:123456789012:stack/${stackName}/11111111-2222-3333-4444-${stackSuffix}`;
const knownLogGroups = [
  `/aws/ssm/sessions/${prefix}-${stackSuffix}`,
  `/aws/lambda/${prefix}-analysis-proxy-${stackSuffix}`,
  `/aws/lambda/${prefix}-analysis-authorizer-${stackSuffix}`,
  `/aws/lambda/${prefix}-source-stager-${stackSuffix}`,
  `/aws/lambda/${prefix}-sandbox-watchdog-${stackSuffix}`,
  `/aws/apigateway/${prefix}-analysis-${stackSuffix}`,
];

const owned = [{ Key: 'AnalysisStack', Value: prefix }];

function baseScenario() {
  return {
    stack: { StackName: stackName, StackStatus: 'DELETE_COMPLETE', StackId: stackId },
    buckets: ['ucs503-analysis-source-123', 'ucs503-analysis-unowned-123'],
    tableIds: {
      'ucs503-analysis-lifecycle': 'table-id-lifecycle',
      'ucs503-analysis-unowned': 'table-id-unowned',
    },
    bucketTags: {
      'ucs503-analysis-source-123': owned,
      'ucs503-analysis-unowned-123': [],
    },
    tables: ['ucs503-analysis-lifecycle', 'ucs503-analysis-unowned'],
    tableTags: {
      'ucs503-analysis-lifecycle': owned,
      'ucs503-analysis-unowned': [],
    },
    volumes: [
      { VolumeId: 'vol-0123456789abcdef0', State: 'available', Tags: owned, Attachments: [] },
      { VolumeId: 'vol-0123456789abcdef1', State: 'available', Tags: [], Attachments: [] },
    ],
    snapshots: [
      { SnapshotId: 'snap-0123456789abcdef0', Tags: owned, State: 'completed' },
      { SnapshotId: 'snap-0123456789abcdef1', Tags: [], State: 'completed' },
    ],
    logGroups: [
      ...knownLogGroups.map((logGroupName) => ({ logGroupName, creationTime: 1760000000000 })),
      {
        logGroupName: `/aws/lambda/${prefix}-analysis-proxy-${otherStackSuffix}`,
        creationTime: 1760000000000,
      },
      {
        logGroupName: `/aws/lambda/${prefix}-analysis-proxy`,
        creationTime: 1760000000000,
      },
      { logGroupName: `/aws/lambda/${prefix}-unapproved-function-${stackSuffix}`, creationTime: 1760000000000 },
    ],
    versionPages: [
      {
        Versions: Array.from({ length: 1001 }, (_, index) => ({
          Key: `object-${index}`,
          VersionId: `version-${index}`,
        })),
        DeleteMarkers: [{ Key: 'marker-first-page', VersionId: 'delete-marker-first' }],
        IsTruncated: true,
        NextKeyMarker: 'object-1000',
        NextVersionIdMarker: 'version-1000',
      },
      {
        Versions: [{ Key: 'last-object', VersionId: 'version-last' }],
        DeleteMarkers: [{ Key: 'deleted-object', VersionId: 'delete-marker-last' }],
        IsTruncated: false,
      },
    ],
  };
}

const mockAwsSource = String.raw`
import { appendFileSync, readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

const [service, operation, ...args] = process.argv.slice(2);
const scenario = JSON.parse(readFileSync(process.env.ANALYSIS_CLEANUP_MOCK_SCENARIO, 'utf8'));

function option(name) {
  const index = args.indexOf(name);
  return index >= 0 ? args[index + 1] : undefined;
}
function respond(value) {
  process.stdout.write(JSON.stringify(value));
}
function fail(message, code = 255) {
  process.stderr.write(message);
  process.exit(code);
}
function tags(value) {
  return { TagSet: value ?? [] };
}
function deletePayload() {
  const value = option('--delete');
  return value.startsWith('file://')
    ? JSON.parse(readFileSync(fileURLToPath(value), 'utf8'))
    : JSON.parse(value);
}

const call = { service, operation, args };
call.childEnvironment = Object.fromEntries([
  'AWS_ENDPOINT_URL',
  'AWS_ENDPOINT_URL_SSM',
  'AWS_PROFILE',
  'AWS_CONFIG_FILE',
  'AWS_SHARED_CREDENTIALS_FILE',
  'NODE_OPTIONS',
  'HTTP_PROXY',
  'HTTPS_PROXY',
].map((key) => [key, process.env[key]]));
if (service === 's3api' && operation === 'delete-objects') {
  call.deleteCount = deletePayload().Objects.length;
}
appendFileSync(process.env.ANALYSIS_CLEANUP_MOCK_LOG, JSON.stringify(call) + '\n');
if (scenario.failOperation === service + ' ' + operation) {
  fail('mock AWS operation failed: ' + service + ' ' + operation, 1);
}

if (service === 'cloudformation' && operation === 'describe-stacks') {
  if (scenario.stack) respond({ Stacks: [scenario.stack] });
  else fail('ValidationError: Stack with id ' + option('--stack-name') + ' does not exist');
} else if (service === 's3api' && operation === 'list-buckets') {
  respond({
    Buckets: scenario.buckets.map((Name) => ({
      Name,
      CreationDate: scenario.bucketCreationDates?.[Name] ?? '2025-01-01T00:00:00Z',
    })),
  });
} else if (service === 's3api' && operation === 'get-bucket-tagging') {
  const bucket = option('--bucket');
  const priorCalls = readFileSync(process.env.ANALYSIS_CLEANUP_MOCK_LOG, 'utf8')
    .trim()
    .split('\n')
    .filter(Boolean)
    .map((line) => JSON.parse(line))
    .filter((entry) => entry.service === 's3api' && entry.operation === 'get-bucket-tagging' && entry.args.includes(bucket));
  const removedBeforeDelete = bucket === scenario.dropBucketTagFor &&
    priorCalls.length === scenario.dropBucketTagAtCall;
  respond(tags(removedBeforeDelete ? [] : scenario.bucketTags[bucket]));
} else if (service === 's3api' && operation === 'list-object-versions') {
  const marker = option('--key-marker');
  const page = marker ? scenario.versionPages[1] : scenario.versionPages[0];
  respond(page);
} else if (service === 's3api' && operation === 'delete-objects') {
  const payload = deletePayload();
  if (payload.Objects.length > 1000) fail('Delete batch exceeded 1000 objects', 1);
  respond({ Deleted: payload.Objects });
} else if (service === 's3api' && operation === 'delete-bucket') {
  respond({});
} else if (service === 'dynamodb' && operation === 'list-tables') {
  respond({ TableNames: scenario.tables });
} else if (service === 'dynamodb' && operation === 'describe-table') {
  const tableName = option('--table-name');
  respond({
    Table: {
      TableName: tableName,
      TableId: scenario.tableIds[tableName],
      TableArn: 'arn:aws:dynamodb:' + option('--region') + ':123456789012:table/' + tableName,
    },
  });
} else if (service === 'dynamodb' && operation === 'list-tags-of-resource') {
  const tableName = option('--resource-arn').split('/').pop();
  respond({ Tags: scenario.tableTags[tableName] ?? [] });
} else if (service === 'dynamodb' && operation === 'delete-table') {
  respond({});
} else if (service === 'ec2' && operation === 'describe-volumes') {
  respond({ Volumes: scenario.volumes });
} else if (service === 'ec2' && operation === 'describe-snapshots') {
  respond({ Snapshots: scenario.snapshots });
} else if (service === 'ec2' && operation === 'delete-volume') {
  respond({});
} else if (service === 'ec2' && operation === 'delete-snapshot') {
  respond({});
} else if (service === 'logs' && operation === 'describe-log-groups') {
  respond({ logGroups: scenario.logGroups });
} else if (service === 'logs' && operation === 'delete-log-group') {
  respond({});
} else {
  fail('Unexpected AWS operation: ' + service + ' ' + operation, 1);
}
`;

async function withMockAws(scenario, callback) {
  const directory = await mkdtemp(path.join(root, '.analysis-cleanup-test-'));
  const mockPath = path.join(directory, 'mock-aws.mjs');
  const scenarioPath = path.join(directory, 'scenario.json');
  const logPath = path.join(directory, 'calls.jsonl');
  await writeFile(mockPath, mockAwsSource);
  await writeFile(scenarioPath, JSON.stringify(scenario));
  try {
    return await callback({
      directory,
      logPath,
      scenarioPath,
      async run(args, extraEnv = {}) {
        return spawnSync(process.execPath, [utility, ...args], {
          cwd: root,
          encoding: 'utf8',
          env: {
            ...process.env,
            ...extraEnv,
            ANALYSIS_CLEANUP_TEST_MODE: '1',
            ANALYSIS_CLEANUP_AWS_CLI: mockPath,
            ANALYSIS_CLEANUP_MOCK_SCENARIO: scenarioPath,
            ANALYSIS_CLEANUP_MOCK_LOG: logPath,
          },
        });
      },
      async calls() {
        try {
          return (await readFile(logPath, 'utf8'))
            .trim()
            .split('\n')
            .filter(Boolean)
            .map((line) => JSON.parse(line));
        } catch (error) {
          if (error.code === 'ENOENT') return [];
          throw error;
        }
      },
      async writeScenario(nextScenario) {
        await writeFile(scenarioPath, JSON.stringify(nextScenario));
      },
    });
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
}

function planArgs() {
  return ['plan', '--stack-name', stackName, '--prefix', prefix, '--region', region];
}

async function createPlan(context) {
  const result = await context.run(planArgs());
  assert.equal(result.status, 0, result.stderr);
  return JSON.parse(result.stdout);
}

test('plans only exactly tagged resources and log groups suffixed for the deleted target stack', async () => {
  await withMockAws(baseScenario(), async (context) => {
    const result = await context.run(planArgs());
    assert.equal(result.status, 0, result.stderr);
    const plan = JSON.parse(result.stdout);
    assert.equal(plan.prefix, prefix);
    assert.equal(plan.stackName, stackName);
    assert.match(result.stderr, /inventory/i);
    assert.deepEqual(
      plan.resources.map(({ type, id }) => `${type}:${id}`).sort(),
      [
        'bucket:ucs503-analysis-source-123',
        'table:ucs503-analysis-lifecycle',
        'volume:vol-0123456789abcdef0',
        'snapshot:snap-0123456789abcdef0',
        ...knownLogGroups.map((name) => `log-group:${name}`),
        `log-group:/aws/lambda/${prefix}-analysis-proxy`,
      ].sort(),
    );
    assert.equal(
      plan.resources.some(({ id }) => id === `/aws/lambda/${prefix}-analysis-proxy-${otherStackSuffix}`),
      false,
    );
    assert.equal(plan.resources.some((resource) => resource.id.includes('unowned')), false);
    const calls = await context.calls();
    assert.ok(calls.some(({ service, operation }) => service === 's3api' && operation === 'get-bucket-tagging'));
    assert.ok(calls.some(({ service, operation }) => service === 'dynamodb' && operation === 'list-tags-of-resource'));
    assert.ok(calls.every(({ operation }) => !/^delete-|^terminate-|^deregister-/.test(operation)));
    assert.match(calls.find(({ service, operation }) => service === 'cloudformation' && operation === 'describe-stacks').args.join(' '), /ucs503-analysis-v1/);
  });
});

test('does not pass endpoint, profile, proxy, or Node override variables to AWS subprocesses', async () => {
  await withMockAws(baseScenario(), async (context) => {
    const result = await context.run(planArgs(), {
      AWS_ENDPOINT_URL: 'http://attacker.invalid',
      AWS_ENDPOINT_URL_SSM: 'http://attacker.invalid',
      AWS_PROFILE: 'untrusted',
      AWS_CONFIG_FILE: 'C:\\untrusted\\config',
      AWS_SHARED_CREDENTIALS_FILE: 'C:\\untrusted\\credentials',
      NODE_OPTIONS: '--no-warnings',
      HTTP_PROXY: 'http://attacker.invalid',
      HTTPS_PROXY: 'http://attacker.invalid',
    });
    assert.equal(result.status, 0, result.stderr);
    const calls = await context.calls();
    assert.ok(calls.length > 0);
    for (const call of calls) {
      assert.deepEqual(call.childEnvironment, {});
    }
  });
});

test('an absent stack cannot authorize suffixed log groups but exact legacy names remain eligible', async () => {
  const absent = baseScenario();
  absent.stack = null;
  await withMockAws(absent, async (context) => {
    const plan = await createPlan(context);
    const logGroupNames = plan.resources
      .filter(({ type }) => type === 'log-group')
      .map(({ id }) => id);
    assert.deepEqual(logGroupNames, [`/aws/lambda/${prefix}-analysis-proxy`]);
  });
});

test('apply refuses when the deleted target stack identity changes after planning', async () => {
  await withMockAws(baseScenario(), async (context) => {
    const plan = await createPlan(context);
    const planPath = path.join(context.directory, 'plan.json');
    await writeFile(planPath, JSON.stringify(plan));
    const changedStack = baseScenario();
    changedStack.stack.StackId = `arn:aws:cloudformation:${region}:123456789012:stack/${stackName}/aaaaaaaa-bbbb-cccc-dddd-${otherStackSuffix}`;
    await context.writeScenario(changedStack);
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      prefix,
      '--region',
      region,
    ]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /stack identity changed|inventory changed/i);
    assert.equal((await context.calls()).some(({ operation }) => /^delete-/.test(operation)), false);
  });
});

test('planning is deterministic and refuses a stack that has not been deleted', async () => {
  await withMockAws(baseScenario(), async (context) => {
    const first = await createPlan(context);
    const second = await createPlan(context);
    assert.deepEqual(first, second);
  });
  const active = baseScenario();
  active.stack = { StackName: stackName, StackStatus: 'CREATE_COMPLETE', StackId: stackId };
  await withMockAws(active, async (context) => {
    const result = await context.run(planArgs());
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /still exists|not deleted/i);
  });
});

test('apply requires exact prefix confirmation before making any AWS calls', async () => {
  await withMockAws(baseScenario(), async (context) => {
    const plan = await createPlan(context);
    const planPath = path.join(context.directory, 'plan.json');
    await writeFile(planPath, JSON.stringify(plan));
    const before = (await context.calls()).length;
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      `${prefix}-wrong`,
      '--region',
      region,
    ]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /confirmation|prefix/i);
    assert.equal((await context.calls()).length, before);
  });
});

test('apply rejects malformed plans before making any AWS calls', async () => {
  await withMockAws(baseScenario(), async (context) => {
    const planPath = path.join(context.directory, 'malformed-plan.json');
    await writeFile(planPath, '{"formatVersion":1,"resources":[]}');
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      prefix,
      '--region',
      region,
    ]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /schema|malformed|prefix/i);
    assert.equal((await context.calls()).length, 0);
  });
});

test('apply revalidates identity and ownership, then empties S3 versions before deleting resources', async () => {
  await withMockAws(baseScenario(), async (context) => {
    const plan = await createPlan(context);
    const planPath = path.join(context.directory, 'plan.json');
    await writeFile(planPath, JSON.stringify(plan));
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      prefix,
      '--region',
      region,
    ]);
    assert.equal(result.status, 0, result.stderr);
    const calls = await context.calls();
    const deletes = calls.filter(({ operation }) => /^delete-/.test(operation));
    assert.ok(deletes.some(({ service, operation }) => service === 'dynamodb' && operation === 'delete-table'));
    assert.ok(deletes.some(({ service, operation }) => service === 'ec2' && operation === 'delete-volume'));
    assert.ok(deletes.some(({ service, operation }) => service === 'ec2' && operation === 'delete-snapshot'));
    assert.equal(deletes.filter(({ service, operation }) => service === 'logs' && operation === 'delete-log-group').length, 7);
    const versionLists = calls.filter(({ service, operation }) => service === 's3api' && operation === 'list-object-versions');
    assert.equal(versionLists.length, 2);
    assert.ok(versionLists[1].args.join(' ').includes('object-1000'));
    const objectDeletes = calls.filter(({ service, operation }) => service === 's3api' && operation === 'delete-objects');
    assert.deepEqual(objectDeletes.map(({ deleteCount }) => deleteCount), [1000, 2, 2]);
    const bucketDeleteIndex = calls.findIndex(({ service, operation }) => service === 's3api' && operation === 'delete-bucket');
    const lastObjectDeleteIndex = calls.findLastIndex(({ service, operation }) => service === 's3api' && operation === 'delete-objects');
    assert.ok(bucketDeleteIndex > lastObjectDeleteIndex);
    assert.ok(calls.every(({ service, operation, args }) => {
      const invocation = `${service} ${operation} ${args.join(' ')}`;
      return !/terminate-instances|deregister-image|delete-stack|archive/i.test(invocation);
    }));
  });
});

test('apply stops before deletion if ownership changes or the plan has an attached volume', async () => {
  const ownershipScenario = baseScenario();
  await withMockAws(ownershipScenario, async (context) => {
    const plan = await createPlan(context);
    const planPath = path.join(context.directory, 'plan.json');
    await writeFile(planPath, JSON.stringify(plan));
    const changed = baseScenario();
    changed.bucketTags['ucs503-analysis-source-123'] = [];
    await context.writeScenario(changed);
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      prefix,
      '--region',
      region,
    ]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /changed|ownership|tag/i);
    assert.equal((await context.calls()).some(({ operation }) => /^delete-/.test(operation)), false);
  });

  await withMockAws(baseScenario(), async (context) => {
    const plan = await createPlan(context);
    const planPath = path.join(context.directory, 'plan.json');
    await writeFile(planPath, JSON.stringify(plan));
    const replacementTable = baseScenario();
    replacementTable.tableIds['ucs503-analysis-lifecycle'] = 'replacement-table-id';
    await context.writeScenario(replacementTable);
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      prefix,
      '--region',
      region,
    ]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /inventory changed/i);
    assert.equal((await context.calls()).some(({ operation }) => /^delete-/.test(operation)), false);
  });

  const changingTagScenario = baseScenario();
  changingTagScenario.dropBucketTagFor = 'ucs503-analysis-source-123';
  changingTagScenario.dropBucketTagAtCall = 3;
  await withMockAws(changingTagScenario, async (context) => {
    const plan = await createPlan(context);
    const planPath = path.join(context.directory, 'plan.json');
    await writeFile(planPath, JSON.stringify(plan));
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      prefix,
      '--region',
      region,
    ]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /ownership tag/i);
    assert.equal((await context.calls()).some(({ operation }) => /^delete-/.test(operation)), false);
  });

  const attachedScenario = baseScenario();
  attachedScenario.volumes[0].Attachments = [{ InstanceId: 'i-0123456789abcdef0', State: 'attached' }];
  await withMockAws(attachedScenario, async (context) => {
    const plan = await createPlan(context);
    const planPath = path.join(context.directory, 'plan.json');
    await writeFile(planPath, JSON.stringify(plan));
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      prefix,
      '--region',
      region,
    ]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /attached EBS volume/i);
    assert.equal((await context.calls()).some(({ operation }) => operation === 'delete-volume'), false);
    assert.equal((await context.calls()).some(({ operation }) => /^delete-/.test(operation)), false);
  });
});

test('AWS operation failures stop cleanup with non-zero status', async () => {
  const failingScenario = baseScenario();
  failingScenario.failOperation = 's3api list-object-versions';
  await withMockAws(failingScenario, async (context) => {
    const plan = await createPlan(context);
    const planPath = path.join(context.directory, 'plan.json');
    await writeFile(planPath, JSON.stringify(plan));
    const result = await context.run([
      'apply',
      '--plan',
      planPath,
      '--confirm-prefix',
      prefix,
      '--region',
      region,
    ]);
    assert.notEqual(result.status, 0);
    assert.match(result.stderr, /mock AWS operation failed/i);
    assert.equal((await context.calls()).some(({ service, operation }) => service === 's3api' && operation === 'delete-bucket'), false);
  });
});
