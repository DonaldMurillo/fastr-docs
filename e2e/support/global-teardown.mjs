import fs from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const e2eRoot = path.resolve(here, '..');
const runtimePath = path.join(e2eRoot, '.runtime.json');

const stop = (pid) => {
  if (!pid) return;
  if (process.platform === 'win32') {
    try { execFileSync('taskkill', ['/PID', String(pid), '/T', '/F'], { stdio: 'ignore', windowsHide: true }); } catch {}
    return;
  }
  try { process.kill(pid, 'SIGTERM'); } catch {}
};

export default async function globalTeardown() {
  let runtime;
  try { runtime = JSON.parse(await fs.readFile(runtimePath, 'utf8')); } catch { return; }
  stop(runtime.appPid);
  stop(runtime.devPid);
  stop(runtime.staticPid);
  stop(runtime.apiPid);
  stop(runtime.manualPid);
  stop(runtime.manualStaticPid);
  stop(runtime.selfPid);
  await fs.rm(runtime.target, { recursive: true, force: true });
  await fs.rm(runtimePath, { force: true });
}
