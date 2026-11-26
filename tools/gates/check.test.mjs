// node --test tools/gates/check.test.mjs
import assert from 'node:assert/strict';
import { test } from 'node:test';

const { globToRegExp } = await import('./glob.mjs');

test('should match ** across directories and * within one', () => {
  assert.ok(globToRegExp('db/migrations/**').test('db/migrations/002_add.sql'));
  assert.ok(globToRegExp('**/Dockerfile').test('services/api/Dockerfile'));
  assert.ok(globToRegExp('**/Dockerfile').test('Dockerfile'));
  assert.ok(globToRegExp('docker-compose*.yml').test('docker-compose.proxy.yml'));
  assert.ok(!globToRegExp('docker-compose*.yml').test('ops/docker-compose.yml'));
  assert.ok(!globToRegExp('services/api/src/auth.ts').test('services/api/src/auth.test.ts'));
});
