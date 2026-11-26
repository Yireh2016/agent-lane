import { describe, expect, it } from 'vitest';
import { gradeFor } from './grade.js';
import type { Severity } from './repository.js';

const f = (...severities: Severity[]) => severities.map((severity) => ({ severity }));

describe('gradeFor', () => {
  it('should grade A when there are no findings', () => {
    expect(gradeFor([])).toBe('A');
  });

  it('should grade by the worst severity present', () => {
    expect(gradeFor(f('low', 'low'))).toBe('B');
    expect(gradeFor(f('low', 'medium'))).toBe('C');
    expect(gradeFor(f('medium', 'high', 'low'))).toBe('D');
    expect(gradeFor(f('low', 'critical', 'high'))).toBe('F');
  });
});
