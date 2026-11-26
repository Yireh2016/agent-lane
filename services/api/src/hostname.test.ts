import { describe, expect, it } from 'vitest';
import { normalizeHostname } from './hostname.js';

describe('normalizeHostname', () => {
  it('should lowercase and strip a trailing dot', () => {
    expect(normalizeHostname(' Example.COM. ')).toEqual({ ok: true, hostname: 'example.com' });
  });

  it.each([
    ['https://example.com', 'scheme'],
    ['example.com:443', 'port'],
    ['example.com/login', 'path'],
    ['-bad.example.com', 'leading hyphen'],
    ['', 'empty'],
    [42, 'not a string'],
  ] as [unknown, string][])('should reject %s (%s)', (input) => {
    expect(normalizeHostname(input).ok).toBe(false);
  });
});
