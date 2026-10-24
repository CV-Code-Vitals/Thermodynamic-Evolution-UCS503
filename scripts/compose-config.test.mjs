import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { resolve } from "node:path";
import test from "node:test";

const repoRoot = resolve(fileURLToPath(new URL("..", import.meta.url)));

function readRepoFile(relativePath) {
  return readFileSync(resolve(repoRoot, relativePath), "utf8");
}

function serviceBlock(compose, serviceName) {
  compose = compose.replace(/\r\n/g, "\n");
  const marker = `  ${serviceName}:\n`;
  const start = compose.indexOf(marker);
  assert.notEqual(start, -1, `missing Compose service ${serviceName}`);
  const afterService = compose.slice(start + marker.length);
  const nextService = afterService.search(/^  [A-Za-z0-9_-]+:\s*$/m);
  return nextService === -1 ? afterService : afterService.slice(0, nextService);
}

function serviceNames(compose) {
  compose = compose.replace(/\r\n/g, "\n");
  const servicesStart = compose.indexOf("services:\n");
  assert.notEqual(servicesStart, -1, "Compose file must define services");
  const afterServices = compose.slice(servicesStart + "services:\n".length);
  const nextTopLevelKey = afterServices.search(/^[A-Za-z][A-Za-z0-9_-]*:\s*$/m);
  const servicesSection = nextTopLevelKey === -1 ? afterServices : afterServices.slice(0, nextTopLevelKey);
  return [...servicesSection.matchAll(/^  ([A-Za-z0-9_-]+):\s*$/gm)].map((match) => match[1]);
}

function assertNoPublishedPorts(service, serviceName) {
  assert.doesNotMatch(service, /^\s*(?:ports|expose):/m, `${serviceName} must not publish ports`);
}

test("analysis compose is isolated from the sandbox broker and Docker socket", () => {
  const compose = readRepoFile("docker-compose.analysis.yml");
  const service = serviceBlock(compose, "ucs503-analysis");

  assert.deepEqual(serviceNames(compose), ["ucs503-analysis"]);
  assert.match(service, /image:\s*ghcr\.io\/cv-code-vitals\/ucs503-analysis:\$\{IMAGE_TAG:/);
  assert.match(service, /network_mode:\s*host/);
  assert.match(service, /ANALYSIS_API_TOKEN:/);
  assert.match(service, /SANDBOX_SERVICE_URL:/);
  assert.match(service, /SANDBOX_SERVICE_TOKEN:/);
  assert.match(service, /SANDBOX_TLS_CA_FILE:/);
  assertNoPublishedPorts(service, "analysis");
  assert.doesNotMatch(service, /docker\.sock|ucs503-sandbox-broker|ucs503-api|\blatest\b/i);
});

test("sandbox compose keeps Docker access private and publishes no broker port", () => {
  const compose = readRepoFile("docker-compose.sandbox.yml");
  const service = serviceBlock(compose, "ucs503-sandbox-broker");

  assert.deepEqual(serviceNames(compose), ["ucs503-sandbox-broker"]);
  assert.match(service, /image:\s*ghcr\.io\/cv-code-vitals\/ucs503-analysis:\$\{IMAGE_TAG:/);
  assert.match(service, /network_mode:\s*host/);
  assert.match(service, /SANDBOX_SERVICE_TOKEN:/);
  assert.match(service, /SANDBOX_TLS_CERT_FILE:/);
  assert.match(service, /SANDBOX_TLS_KEY_FILE:/);
  assert.match(service, /group_add:/);
  assert.match(service, /DOCKER_GID/);
  assert.match(service, /SANDBOX_BROKER_WORKSPACE_DIR:/);
  assert.match(service, /\/var\/run\/docker\.sock:\/var\/run\/docker\.sock/);
  assert.equal((service.match(/\/var\/run\/docker\.sock/g) ?? []).length, 2);
  assertNoPublishedPorts(service, "sandbox broker");
  assert.doesNotMatch(service, /ucs503-api|DATABASE_URL|sandboxes_shared|\blatest\b/i);
});

test("example environment files describe required runtime values without actual secrets", () => {
  const analysisEnv = readRepoFile(".env.analysis.example");
  const sandboxEnv = readRepoFile(".env.sandbox.example");

  assert.match(analysisEnv, /^IMAGE_TAG=replace-with-40-character-commit-sha$/m);
  assert.match(analysisEnv, /^ANALYSIS_API_TOKEN=replace-with-ssm-value$/m);
  assert.match(analysisEnv, /^SANDBOX_SERVICE_TOKEN=replace-with-ssm-value$/m);
  assert.match(analysisEnv, /^SANDBOX_TLS_CA_SOURCE=/m);
  assert.match(sandboxEnv, /^IMAGE_TAG=replace-with-40-character-commit-sha$/m);
  assert.match(sandboxEnv, /^SANDBOX_SERVICE_TOKEN=replace-with-ssm-value$/m);
  assert.match(sandboxEnv, /^SANDBOX_TLS_CERT_SOURCE=/m);
  assert.match(sandboxEnv, /^SANDBOX_TLS_KEY_SOURCE=/m);
  assert.match(sandboxEnv, /^DOCKER_GID=\d+$/m);
});

test("legacy EC2 compose is explicitly documented as manual-only", () => {
  const legacyCompose = readRepoFile("docker-compose.ec2.yml");

  assert.match(legacyCompose, /^# LEGACY: manual-only; do not use for the split EC2 deployment\.$/m);
});
