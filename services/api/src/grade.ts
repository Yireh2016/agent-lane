import type { Severity } from './repository.js';

export type Grade = 'A' | 'B' | 'C' | 'D' | 'F';

const GRADES_WORST_FIRST: [Severity, Grade][] = [
  ['critical', 'F'],
  ['high', 'D'],
  ['medium', 'C'],
  ['low', 'B'],
];

export function gradeFor(findings: { severity: Severity }[]): Grade {
  for (const [severity, grade] of GRADES_WORST_FIRST) {
    if (findings.some((f) => f.severity === severity)) return grade;
  }
  return 'A';
}
