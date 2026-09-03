import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..');

const jobs = [
  { name: '@forma/api-client', dir: 'frontend/packages/forma-api-client', args: ['--noEmit'] },
  { name: '@forma/data', dir: 'frontend/packages/forma-data', args: ['--noEmit'] },
  {
    name: '@forma/app',
    dir: 'frontend/apps/forma',
    args: ['--noEmit', '-p', 'tsconfig.build.json'],
  },
];

function resolveTsc(pkgDir) {
  const req = createRequire(join(pkgDir, 'package.json'));
  return req.resolve('typescript/bin/tsc');
}

let failed = false;
for (const job of jobs) {
  const pkgDir = join(root, job.dir);
  const tsc = resolveTsc(pkgDir);
  if (!job.args.includes('--noEmit')) {
    throw new Error(`${job.name} typecheck must pass --noEmit`);
  }
  console.log(`\n[forma-typecheck] ${job.name}: tsc ${job.args.join(' ')}`);
  const result = spawnSync(process.execPath, [tsc, ...job.args], {
    cwd: pkgDir,
    stdio: 'inherit',
    env: process.env,
  });
  if (result.status !== 0) {
    failed = true;
    console.error(`[forma-typecheck] ${job.name} FAILED`);
  } else {
    console.log(`[forma-typecheck] ${job.name} PASS`);
  }
}

if (failed) {
  process.exit(1);
}
