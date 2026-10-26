import assert from 'node:assert/strict';
import test from 'node:test';
import viteConfig from '../vite.config.js';

function configureProduction(apiBase, allowRelative = false) {
  const original = process.env.VITE_API_BASE;
  const originalAllowRelative = process.env.VITE_ALLOW_RELATIVE_API_BASE;
  if (apiBase === undefined) {
    delete process.env.VITE_API_BASE;
  } else {
    process.env.VITE_API_BASE = apiBase;
  }
  if (allowRelative) {
    process.env.VITE_ALLOW_RELATIVE_API_BASE = 'true';
  } else {
    delete process.env.VITE_ALLOW_RELATIVE_API_BASE;
  }

  try {
    return viteConfig({ mode: 'production' });
  } finally {
    if (original === undefined) {
      delete process.env.VITE_API_BASE;
    } else {
      process.env.VITE_API_BASE = original;
    }
    if (originalAllowRelative === undefined) {
      delete process.env.VITE_ALLOW_RELATIVE_API_BASE;
    } else {
      process.env.VITE_ALLOW_RELATIVE_API_BASE = originalAllowRelative;
    }
  }
}

test('production build rejects a missing API URL', () => {
  assert.throws(() => configureProduction(undefined), /VITE_API_BASE/);
});

test('production build rejects a relative API path', () => {
  assert.throws(() => configureProduction('/api'), /absolute HTTPS URL/);
});

test('production build accepts a public HTTPS API URL', () => {
  const config = configureProduction('https://api.example.com');
  assert.equal(config.base, './');
});

test('local static build accepts the same-origin API proxy explicitly', () => {
  const config = configureProduction('/api', true);
  assert.equal(config.base, './');
});

test('production validates the optional separate analysis gateway', () => {
  const original = process.env.VITE_ANALYSIS_API_BASE;
  try {
    process.env.VITE_ANALYSIS_API_BASE = 'http://analysis.example';
    assert.throws(() => configureProduction('https://archive.example/api'), /VITE_ANALYSIS_API_BASE/);
    process.env.VITE_ANALYSIS_API_BASE = 'https://analysis.example/prod';
    assert.equal(configureProduction('https://archive.example/api').base, './');
  } finally {
    if (original === undefined) delete process.env.VITE_ANALYSIS_API_BASE;
    else process.env.VITE_ANALYSIS_API_BASE = original;
  }
});
