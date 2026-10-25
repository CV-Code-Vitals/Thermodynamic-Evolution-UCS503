import { LambdaClient, InvokeCommand } from '@aws-sdk/client-lambda';
import { GetParameterCommand, SSMClient } from '@aws-sdk/client-ssm';
import { isIP } from 'node:net';
import { Readable } from 'node:stream';
import { pipeline } from 'node:stream/promises';

const snapshotIdPattern = /^[a-f0-9]{32}$/;
const scanTokenParameter = '/ucs503/analysis/api-token';
const defaultMaxRequestBytes = 64 * 1024;
const defaultMaxScanResponseBytes = 5 * 1024 * 1024;
const allowedRoutes = new Map([
  ['/api/scan', 'scan'],
  ['/api/evolve', 'evolve'],
]);

function isPrivateIPv4(address) {
  const octets = address.split('.').map(Number);
  if (octets.length !== 4 || octets.some((part) => !Number.isInteger(part) || part < 0 || part > 255)) {
    return false;
  }
  return octets[0] === 10
    || octets[0] === 127
    || (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31)
    || (octets[0] === 192 && octets[1] === 168);
}

function isPrivateIPv6(address) {
  const normalized = address.toLowerCase();
  return normalized === '::1' || normalized.startsWith('fc') || normalized.startsWith('fd');
}

function validateAnalysisBaseURL(value) {
  let base;
  try {
    base = new URL(value);
  } catch {
    throw new Error('analysis API base URL is invalid');
  }
  if (!['http:', 'https:'].includes(base.protocol)
    || base.username
    || base.password
    || base.search
    || base.hash
    || isIP(base.hostname) === 0) {
    throw new Error('analysis API base URL must use a private IP address');
  }
  const privateAddress = isIP(base.hostname) === 4
    ? isPrivateIPv4(base.hostname)
    : isPrivateIPv6(base.hostname);
  if (!privateAddress) throw new Error('analysis API base URL must use a private IP address');
  return base;
}

function validateRepositoryURL(value) {
  if (typeof value !== 'string' || value.length > 2048) {
    throw new Error('repository URL is invalid');
  }
  let parsed;
  try {
    parsed = new URL(value);
  } catch {
    throw new Error('repository URL is invalid');
  }
  const segments = parsed.pathname.split('/').filter(Boolean);
  if (parsed.protocol !== 'https:'
    || parsed.hostname.toLowerCase() !== 'github.com'
    || parsed.username
    || parsed.password
    || parsed.port
    || parsed.search
    || parsed.hash
    || segments.length !== 2
    || parsed.pathname.includes('%')
    || segments.some((segment) => segment === '.' || segment === '..')) {
    throw new Error('repository URL must identify a public GitHub repository');
  }
}

function parseRequestBody(event, maxBytes) {
  if (typeof event.body !== 'string') throw Object.assign(new Error('request body is required'), { statusCode: 400 });
  const maximumEncodedLength = event.isBase64Encoded ? Math.ceil(maxBytes * 4 / 3) : maxBytes;
  if (event.body.length > maximumEncodedLength) {
    throw Object.assign(new Error('request body is too large'), { statusCode: 413 });
  }
  let bodyBytes;
  try {
    bodyBytes = event.isBase64Encoded
      ? Buffer.from(event.body, 'base64')
      : Buffer.from(event.body, 'utf8');
  } catch {
    throw Object.assign(new Error('request body encoding is invalid'), { statusCode: 400 });
  }
  if (bodyBytes.length > maxBytes) {
    throw Object.assign(new Error('request body is too large'), { statusCode: 413 });
  }
  let payload;
  try {
    payload = JSON.parse(bodyBytes.toString('utf8'));
  } catch {
    throw Object.assign(new Error('request body must be valid JSON'), { statusCode: 400 });
  }
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
    throw Object.assign(new Error('request body must be a JSON object'), { statusCode: 400 });
  }
  return payload;
}

function sourceIpFromGateway(event) {
  const sourceIp = event.requestContext?.identity?.sourceIp;
  if (typeof sourceIp !== 'string' || isIP(sourceIp) === 0) {
    throw Object.assign(new Error('API Gateway client IP is unavailable'), { statusCode: 400 });
  }
  return sourceIp;
}

function buildUpstreamBody(route, payload, snapshotID) {
  if (route === 'scan') return { snapshot_id: snapshotID };
  if (typeof payload.file_path !== 'string' || payload.file_path.trim() === '') {
    throw Object.assign(new Error('file_path is required'), { statusCode: 400 });
  }
  return {
    snapshot_id: snapshotID,
    file_path: payload.file_path,
    ...(typeof payload.function_name === 'string' ? { function_name: payload.function_name } : {}),
    ...(Number.isInteger(payload.start_byte) ? { start_byte: payload.start_byte } : {}),
    ...(Number.isInteger(payload.end_byte) ? { end_byte: payload.end_byte } : {}),
    ...(typeof payload.base_energy === 'number' ? { base_energy: payload.base_energy } : {}),
    ...(payload.config && typeof payload.config === 'object' && !Array.isArray(payload.config)
      ? { config: payload.config }
      : {}),
  };
}

async function readBoundedBody(body, maximumBytes) {
  if (!body) return Buffer.alloc(0);
  const chunks = [];
  let size = 0;
  for await (const chunk of body) {
    const buffer = Buffer.from(chunk);
    size += buffer.length;
    if (size > maximumBytes) {
      throw new Error('analysis response exceeded the limit');
    }
    chunks.push(buffer);
  }
  return Buffer.concat(chunks);
}

function responseMetadata(statusCode, headers = {}, frontendOrigin) {
  return {
    statusCode,
    headers: {
      'Content-Type': headers['Content-Type'] ?? 'application/json; charset=utf-8',
      'Cache-Control': headers['Cache-Control'] ?? 'no-store',
      ...(headers['X-Accel-Buffering'] ? { 'X-Accel-Buffering': headers['X-Accel-Buffering'] } : {}),
      ...(frontendOrigin ? {
        'Access-Control-Allow-Origin': frontendOrigin,
        'Access-Control-Allow-Headers': 'Authorization,Content-Type',
        'Access-Control-Allow-Methods': 'OPTIONS,POST',
        Vary: 'Origin',
      } : {}),
    },
  };
}

async function writeBuffer(rawResponseStream, metadata, body, factory) {
  const output = factory(rawResponseStream, metadata);
  await pipeline(Readable.from([body]), output);
}

async function invokeStager(lambdaClient, repositoryURL) {
  const functionName = process.env.SOURCE_STAGER_FUNCTION;
  if (!functionName) throw new Error('source stager function is not configured');
  const result = await lambdaClient.send(new InvokeCommand({
    FunctionName: functionName,
    InvocationType: 'RequestResponse',
    Payload: Buffer.from(JSON.stringify({ repository_url: repositoryURL })),
  }));
  if (result.FunctionError) throw new Error('source stager invocation failed');
  let payload;
  try {
    payload = JSON.parse(Buffer.from(result.Payload ?? []).toString('utf8'));
  } catch {
    throw new Error('source stager returned an invalid response');
  }
  return payload;
}

async function getScanToken(ssmClient) {
  const result = await ssmClient.send(new GetParameterCommand({
    Name: scanTokenParameter,
    WithDecryption: true,
  }));
  const token = result.Parameter?.Value;
  if (typeof token !== 'string' || token.length === 0) throw new Error('scan API token is not configured');
  return token;
}

export function createProxyHandler(dependencies = {}) {
  const lambdaClient = dependencies.lambdaClient ?? new LambdaClient({ useDualstackEndpoint: true });
  const ssmClient = dependencies.ssmClient ?? new SSMClient({ useDualstackEndpoint: true });
  const stager = dependencies.stager ?? ((repositoryURL) => invokeStager(lambdaClient, repositoryURL));
  const readScanToken = dependencies.getScanToken ?? (() => getScanToken(ssmClient));
  const fetchImpl = dependencies.fetchImpl ?? globalThis.fetch;
  const maxRequestBytes = dependencies.maxRequestBytes ?? defaultMaxRequestBytes;
  const maxScanResponseBytes = dependencies.maxScanResponseBytes ?? defaultMaxScanResponseBytes;
  const logger = dependencies.logger ?? (() => {});
  const frontendOrigin = dependencies.frontendOrigin ?? process.env.FRONTEND_ORIGIN;
  const responseStreamFactory = dependencies.responseStreamFactory ?? ((stream, metadata) => {
    const runtime = globalThis.awslambda;
    return runtime?.HttpResponseStream?.from
      ? runtime.HttpResponseStream.from(stream, metadata)
      : stream;
  });
  const baseValue = dependencies.analysisBaseURL ?? process.env.ANALYSIS_API_BASE;
  const buildResponseMetadata = (statusCode, headers) => responseMetadata(statusCode, headers, frontendOrigin);

  return async (event, rawResponseStream) => {
    const abortController = new AbortController();
    let responseFinished = false;
    let responseStarted = false;
    const abortOnClose = () => {
      if (!responseFinished) abortController.abort(new Error('client disconnected'));
    };
    rawResponseStream.once?.('close', abortOnClose);

    const sendError = async (statusCode, error) => {
      const body = Buffer.from(JSON.stringify({ error }));
      responseStarted = true;
      await writeBuffer(
        rawResponseStream,
        buildResponseMetadata(statusCode),
        body,
        responseStreamFactory,
      );
      responseFinished = true;
    };

    try {
      if (event.httpMethod !== 'POST') {
        await sendError(405, 'method_not_allowed');
        return;
      }
      const path = event.path ?? event.rawPath;
      const route = allowedRoutes.get(path);
      if (!route) {
        await sendError(404, 'route_not_found');
        return;
      }

      const payload = parseRequestBody(event, maxRequestBytes);
      validateRepositoryURL(payload.repo_url);
      const base = validateAnalysisBaseURL(baseValue);
      const trustedClientIP = route === 'scan' ? sourceIpFromGateway(event) : undefined;
      let analysisToken;
      if (route === 'evolve') {
        if (typeof payload.file_path !== 'string' || payload.file_path.trim() === '') {
          throw Object.assign(new Error('file_path is required'), { statusCode: 400 });
        }
        analysisToken = event.requestContext?.authorizer?.analysisApiToken;
        if (typeof analysisToken !== 'string' || analysisToken.length === 0) {
          throw Object.assign(new Error('Evolution requires an authenticated admin session'), { statusCode: 401 });
        }
      } else {
        analysisToken = await readScanToken();
        if (typeof analysisToken !== 'string' || analysisToken.length === 0) {
          throw new Error('scan API token is not configured');
        }
      }
      const snapshot = await stager(payload.repo_url);
      const snapshotID = typeof snapshot === 'string' ? snapshot : snapshot?.snapshot_id;
      if (typeof snapshotID !== 'string' || !snapshotIdPattern.test(snapshotID)) {
        throw new Error('source stager returned an invalid snapshot ID');
      }

      const upstreamPath = path.startsWith('/') ? path : `/${path}`;
      const basePath = base.pathname.replace(/\/+$/, '');
      const upstreamURL = new URL(`${basePath}${upstreamPath}`, base.origin);
      const headers = {
        Accept: route === 'evolve' ? 'text/event-stream' : 'application/json',
        'Content-Type': 'application/json',
        'X-API-Token': analysisToken,
      };
      if (trustedClientIP) headers['X-Trusted-Client-IP'] = trustedClientIP;

      const upstream = await fetchImpl(upstreamURL, {
        method: 'POST',
        headers,
        body: JSON.stringify(buildUpstreamBody(route, payload, snapshotID)),
        signal: abortController.signal,
      });
      const scanBody = route === 'scan'
        ? await readBoundedBody(upstream.body, maxScanResponseBytes)
        : undefined;
      const contentType = upstream.headers.get('content-type') ?? 'application/octet-stream';
      const metadata = buildResponseMetadata(upstream.status, {
        'Content-Type': contentType,
        'Cache-Control': upstream.headers.get('cache-control') ?? 'no-store',
        'X-Accel-Buffering': contentType.includes('text/event-stream') ? 'no' : undefined,
      });
      responseStarted = true;
      const output = responseStreamFactory(rawResponseStream, metadata);

      if (route === 'scan') {
        await pipeline(Readable.from([scanBody]), output);
      } else if (!upstream.body) {
        await pipeline(Readable.from([]), output);
      } else {
        await pipeline(Readable.fromWeb(upstream.body), output);
      }
      responseFinished = true;
    } catch (error) {
      if (abortController.signal.aborted) return;
      const statusCode = Number.isInteger(error?.statusCode) ? error.statusCode : 502;
      const errorCode = statusCode === 401
        ? 'unauthorized'
        : statusCode === 400
          ? 'invalid_request'
          : statusCode === 413
            ? 'request_too_large'
            : 'analysis_proxy_failed';
      logger('analysis proxy request failed', { statusCode, route: event.path ?? event.rawPath ?? '' });
      if (responseStarted) {
        rawResponseStream.destroy?.();
        return;
      }
      if (!rawResponseStream.destroyed) await sendError(statusCode, errorCode);
    } finally {
      rawResponseStream.removeListener?.('close', abortOnClose);
    }
  };
}

const lambdaHandler = createProxyHandler();
export const handler = globalThis.awslambda?.streamifyResponse
  ? globalThis.awslambda.streamifyResponse(lambdaHandler)
  : lambdaHandler;
