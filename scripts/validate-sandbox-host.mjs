import { execFileSync } from 'node:child_process';
import { statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const requiredID = '10002';

const runCommand = (command, args) =>
  execFileSync(command, args, {
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'pipe'],
  });

export const validateSandboxHost = ({
  platform = process.platform,
  socketPath = process.env.DOCKER_SOCKET_SOURCE || '/run/user/10002/docker.sock',
  workspacePath = process.env.SANDBOX_WORKSPACE_HOST_DIR || '/var/lib/ucs503/sandbox-workspaces',
  user = 'ucs503-sandbox',
  run = runCommand,
  stat = statSync,
} = {}) => {
  if (platform !== 'linux') {
    throw new Error('Sandbox host validation must run on the target Linux host.');
  }

  if (run('id', ['-u', user]).trim() !== requiredID ||
      run('id', ['-g', user]).trim() !== requiredID) {
    throw new Error(`${user} must have UID/GID 10002.`);
  }

  if (run('loginctl', ['show-user', user, '-p', 'Linger', '--value']).trim() !== 'yes') {
    throw new Error(`systemd lingering must be enabled for ${user}.`);
  }

  const socket = stat(socketPath);
  if (!socket.isSocket() || socket.uid !== Number(requiredID)) {
    throw new Error(`Rootless Docker socket must be a socket owned by UID 10002: ${socketPath}`);
  }

  if (!stat(workspacePath).isDirectory()) {
    throw new Error(`Sandbox workspace must be an existing directory: ${workspacePath}`);
  }

  try {
    run('sudo', ['-u', user, '--', 'test', '-w', workspacePath]);
  } catch (error) {
    throw new Error(`Sandbox workspace is not writable by ${user}: ${workspacePath}`, { cause: error });
  }

  const securityOptions = run('sudo', [
    '-u', user, '--', 'env', `DOCKER_HOST=unix://${socketPath}`,
    'docker', 'info', '--format', '{{json .SecurityOptions}}',
  ]);
  if (!securityOptions.includes('rootless')) {
    throw new Error('A rootless Docker daemon is not running for the dedicated sandbox user.');
  }

  return { user, uid: Number(requiredID), gid: Number(requiredID), socketPath, workspacePath, rootless: true };
};

const invokedPath = process.argv[1] && path.resolve(process.argv[1]);
if (invokedPath === fileURLToPath(import.meta.url)) {
  try {
    const result = validateSandboxHost();
    console.log(`Rootless sandbox host ready: ${result.socketPath} (workspace ${result.workspacePath})`);
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
