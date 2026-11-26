#!/usr/bin/env node
// Policy gates for agent-authored changes. Zero dependencies so it runs before install.
//
//   node tools/gates/check.mjs [--base origin/main] [--json]
//
// Exit codes: 0 pass · 1 fail (fix the change) · 3 needs a human (protected paths touched).
import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { globToRegExp } from './glob.mjs';

const args = process.argv.slice(2);
const base = args.includes('--base') ? args[args.indexOf('--base') + 1] : 'origin/main';
const asJson = args.includes('--json');

const git = (...a) => execFileSync('git', a, { encoding: 'utf8' }).trim();
const policy = JSON.parse(readFileSync('.lane/policy.json', 'utf8'));

const matchesAny = (path, globs) => globs.some((g) => globToRegExp(g).test(path));

const mergeBase = git('merge-base', base, 'HEAD');
const changed = git('diff', '--name-only', `${mergeBase}...HEAD`).split('\n').filter(Boolean);

const results = [];
const add = (gate, status, detail) => results.push({ gate, status, detail });

// 1. Protected paths: an agent may propose these changes, but only a human may approve them.
const protectedHits = changed.filter((p) => matchesAny(p, policy.protectedPaths));
add(
  'protected-paths',
  protectedHits.length ? 'needs-human' : 'pass',
  protectedHits.length ? protectedHits : 'no protected paths touched',
);

// 2. Diff size: large agent diffs are hard to review; split them into stacked PRs.
let lines = 0;
for (const row of git('diff', '--numstat', `${mergeBase}...HEAD`).split('\n').filter(Boolean)) {
  const [added, removed, path] = row.split('\t');
  if (added === '-' || matchesAny(path, policy.ignoreFromLineCount ?? [])) continue;
  lines += Number(added) + Number(removed);
}
add(
  'diff-size',
  lines > policy.maxChangedLines ? 'fail' : 'pass',
  `${lines} changed lines (limit ${policy.maxChangedLines})`,
);

// 3. Dependencies: every npm and Go dependency at HEAD must be on the allowlist.
const npmDeps = new Set();
for (const file of git('ls-files', '*package.json', '**/package.json').split('\n').filter(Boolean)) {
  const pkg = JSON.parse(readFileSync(file, 'utf8'));
  for (const field of ['dependencies', 'devDependencies', 'optionalDependencies', 'peerDependencies']) {
    for (const name of Object.keys(pkg[field] ?? {})) {
      if (!name.startsWith('@agent-lane/')) npmDeps.add(name);
    }
  }
}
const goDeps = new Set();
for (const file of git('ls-files', '**/go.mod', 'go.mod').split('\n').filter(Boolean)) {
  const text = readFileSync(file, 'utf8');
  const block = text.match(/require \(([\s\S]*?)\)/g) ?? [];
  const lines = [
    ...block.flatMap((b) => b.split('\n').slice(1, -1)),
    ...(text.match(/^require [^(].*$/gm) ?? []),
  ];
  for (const l of lines) {
    if (l.includes('// indirect')) continue;
    const name = l
      .replace(/^require /, '')
      .trim()
      .split(/\s+/)[0];
    if (name) goDeps.add(name);
  }
}
const unknown = [
  ...[...npmDeps].filter((d) => !policy.allowedDependencies.npm.includes(d)).map((d) => `npm:${d}`),
  ...[...goDeps].filter((d) => !policy.allowedDependencies.go.includes(d)).map((d) => `go:${d}`),
];
add(
  'dependencies',
  unknown.length ? 'fail' : 'pass',
  unknown.length ? unknown : 'all dependencies allowlisted',
);

// 4. Secrets: run gitleaks on the commits in this change when it is installed (CI always installs it).
try {
  execFileSync('gitleaks', ['version'], { stdio: 'ignore' });
  try {
    execFileSync('gitleaks', ['git', '--no-banner', `--log-opts=${mergeBase}..HEAD`, '--redact'], {
      stdio: 'pipe',
    });
    add('secrets', 'pass', 'gitleaks found nothing');
  } catch (e) {
    add('secrets', 'fail', String(e.stdout ?? e.message).slice(-2000));
  }
} catch {
  add('secrets', 'skipped', 'gitleaks not installed');
}

const overall = results.some((r) => r.status === 'fail')
  ? 'fail'
  : results.some((r) => r.status === 'needs-human')
    ? 'needs-human'
    : 'pass';

if (asJson) {
  console.log(JSON.stringify({ base, mergeBase, overall, results }, null, 2));
} else {
  for (const r of results) {
    const d = Array.isArray(r.detail) ? r.detail.join(', ') : r.detail;
    console.log(`${r.status.padEnd(11)} ${r.gate.padEnd(16)} ${d}`);
  }
  console.log(`\noverall: ${overall}`);
}
process.exit(overall === 'fail' ? 1 : overall === 'needs-human' ? 3 : 0);
