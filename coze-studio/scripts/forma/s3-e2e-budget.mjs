/**
 * FORMA S3 Live E2E — real model call budget + gate resume state.
 * Local harness only; state file is gitignored via cursor-results policy.
 */
import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

import { countModelCalls } from './s3-e2e-fixtures.mjs';

const root = join(dirname(fileURLToPath(import.meta.url)), '..', '..', '..');
export const statePath = join(root, 'forma', 'cursor-results', 's3-e2e-state.json');

export const GATES = [
  'auth',
  'real-model',
  'no-silent-mutation',
  'gap',
  'confirmation',
  'edit-confirm',
  'conflict',
  'proposal-diff',
  'apply-provenance',
  'stale-proposal',
  'tenant-isolation',
];

export const MAX_REAL_MODEL_CALLS = parseInt(process.env.MAX_REAL_MODEL_CALLS || '8', 10);
export const MAX_REAL_MODEL_USER_TURNS = parseInt(process.env.MAX_REAL_MODEL_USER_TURNS || '3', 10);

export function loadState() {
  if (!existsSync(statePath)) return null;
  try {
    return JSON.parse(readFileSync(statePath, 'utf8'));
  } catch {
    return null;
  }
}

export function saveState(state) {
  writeFileSync(statePath, `${JSON.stringify(state, null, 2)}\n`, 'utf8');
}

export function gateIndex(name) {
  const idx = GATES.indexOf(name);
  if (idx < 0) throw new Error(`unknown gate: ${name}`);
  return idx;
}

export function shouldRunGate(gate, fromGate, completedGates = []) {
  if (completedGates.includes(gate)) return false;
  if (!fromGate) return true;
  return gateIndex(gate) >= gateIndex(fromGate);
}

export function markGateComplete(state, gate) {
  if (!state.completedGates.includes(gate)) {
    state.completedGates.push(gate);
  }
  state.lastGate = gate;
  state.updatedAt = new Date().toISOString();
  saveState(state);
}

export function initFreshLogs(writeFileSyncFn, paths) {
  if (process.env.FORMA_S3_E2E_RESUME === '1') return;
  for (const p of paths) writeFileSyncFn(p, '', 'utf8');
}

export function reportBudgetStop({ completedGates, pendingGates, modelCalls, failurePoint, reason }) {
  const report = {
    status: 'BUDGET_STOP',
    maxRealModelCalls: MAX_REAL_MODEL_CALLS,
    modelCalls,
    completedGates,
    pendingGates,
    failurePoint,
    reason,
    stoppedAt: new Date().toISOString(),
  };
  // eslint-disable-next-line no-console
  console.error('\n=== FORMA S3 REAL MODEL BUDGET STOP ===');
  // eslint-disable-next-line no-console
  console.error(JSON.stringify(report, null, 2));
  return report;
}

export function assertModelBudget(sessionId, { completedGates = [], failurePoint = 'real-model' } = {}) {
  if (!sessionId) return 0;
  const modelCalls = countModelCalls(sessionId);
  if (modelCalls >= MAX_REAL_MODEL_CALLS) {
    const pending = GATES.filter(g => !completedGates.includes(g));
    reportBudgetStop({
      completedGates,
      pendingGates: pending,
      modelCalls,
      failurePoint,
      reason: `model call budget reached (${modelCalls}/${MAX_REAL_MODEL_CALLS})`,
    });
    process.exit(2);
  }
  return modelCalls;
}

export function realModelProbePass(sessionId) {
  if (process.env.FORMA_S3_REAL_MODEL_PROBE_PASS === '1') return true;
  const saved = loadState();
  if (saved?.realModelProbePass && saved.sessionId === sessionId) return true;
  const rows = countModelCalls(sessionId);
  return rows >= 2;
}
