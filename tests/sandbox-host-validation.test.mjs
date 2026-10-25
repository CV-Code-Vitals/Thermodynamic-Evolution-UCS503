import assert from 'node:assert/strict';
import test from 'node:test';
import { validateSandboxHost } from '../scripts/validate-sandbox-host.mjs';

const socketPath = '/run/user/10002/docker.sock';
const workspacePath = '/var/lib/ucs503/sandbox-workspaces';

function makeOptions(overrides = {}) {
  const options = {
    platform: 'linux',
    socketPath,
    workspacePath,
    user: 'ucs503-sandbox',
    stat: (path) => {
      if (path === socketPath) {
        return { isSocket: () => true, uid: 10002, gid: 10002 };
      }
      if (path === workspacePath) {
        return { isDirectory: () => true };
      }
      throw new Error(`unexpected stat path: ${path}`);
    },
    run: (command, args) => {
      if (command === 'id' && args[0] === '-u') return '10002\n';
      if (command === 'id' && args[0] === '-g') return '10002\n';
      if (command === 'loginctl') return 'yes\n';
      if (command === 'sudo' && args.includes('test')) return '';
      if (command === 'sudo' && args.includes('docker')) {
        assert.ok(args.includes(`DOCKER_HOST=unix://${socketPath}`));
        return '["name=rootless"]\n';
      }
      throw new Error(`unexpected command: ${command} ${args.join(' ')}`);
    },
    ...overrides,
  };
  return options;
}

test('accepts a correctly provisioned rootless sandbox host', () => {
  assert.equal(validateSandboxHost(makeOptions()).rootless, true);
});

test('rejects a sandbox account with the wrong UID or GID', () => {
  const options = makeOptions({
    run: (command, args) => {
      if (command === 'id' && args[0] === '-u') return '10003\n';
      return makeOptions().run(command, args);
    },
  });

  assert.throws(() => validateSandboxHost(options), /UID\/GID 10002/);
});

test('rejects an incorrectly owned rootless socket', () => {
  const options = makeOptions({
    stat: () => ({ isSocket: () => true, uid: 0, gid: 0 }),
  });

  assert.throws(() => validateSandboxHost(options), /owned by UID\/GID 10002/);
});

test('rejects a socket path that is not a Unix socket', () => {
  const options = makeOptions({
    stat: () => ({ isSocket: () => false, uid: 10002, gid: 10002 }),
  });

  assert.throws(() => validateSandboxHost(options), /must be a socket/i);
});

test('rejects a sandbox account without systemd lingering', () => {
  const options = makeOptions({
    run: (command, args) => {
      if (command === 'loginctl') return 'no\n';
      return makeOptions().run(command, args);
    },
  });

  assert.throws(() => validateSandboxHost(options), /lingering must be enabled/i);
});

test('rejects a workspace the sandbox account cannot write', () => {
  const options = makeOptions({
    run: (command, args) => {
      if (command === 'sudo' && args.includes('test')) throw new Error('workspace is not writable');
      return makeOptions().run(command, args);
    },
  });

  assert.throws(() => validateSandboxHost(options), /workspace.*writable/i);
});

test('rejects a Docker daemon that is not rootless', () => {
  const options = makeOptions({
    run: (command, args) => {
      if (command === 'sudo' && args.includes('docker')) return '["name=seccomp"]\n';
      return makeOptions().run(command, args);
    },
  });

  assert.throws(() => validateSandboxHost(options), /rootless Docker daemon/i);
});
