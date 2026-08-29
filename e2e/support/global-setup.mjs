import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { spawn, execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const e2eRoot = path.resolve(here, '..');
const repoRoot = path.resolve(e2eRoot, '..');
const runtimePath = path.join(e2eRoot, '.runtime.json');
const appPort = 4175;
const apiPort = 4176;
const staticPort = 4177;
const manualPort = 4178;
const manualStaticPort = 4179;
const docsLocale = 'pt-BR';

const run = (command, args, cwd, env = {}) => execFileSync(command, args, {
  cwd,
  env: { ...process.env, ...env },
  stdio: 'pipe',
  encoding: 'utf8',
  windowsHide: true,
});

const start = (command, args, cwd, env = {}) => spawn(command, args, {
  cwd,
  env: { ...process.env, ...env },
  stdio: 'ignore',
  windowsHide: true,
});

const stop = (child) => {
  if (!child?.pid) return;
  if (process.platform === 'win32') {
    try { execFileSync('taskkill', ['/PID', String(child.pid), '/T', '/F'], { stdio: 'ignore', windowsHide: true }); } catch {}
    return;
  }
  try { process.kill(child.pid, 'SIGTERM'); } catch {}
};

const waitFor = async (url, timeout = 60_000) => {
  const started = Date.now();
  while (Date.now() - started < timeout) {
    try {
      const response = await fetch(url);
      if (response.status < 500) return;
    } catch {}
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Timed out waiting for ${url}`);
};

export default async function globalSetup() {
  const target = await fs.mkdtemp(path.join(os.tmpdir(), 'fastr-docs-e2e-'));
  let api;
  let staticSite;
  let app;
  let manual;
  let manualStatic;
  try {
    run('go', ['run', './cmd/fastr-docs', 'init', target, '--name', 'E2E Docs', '--module', 'example.com/e2e-docs'], repoRoot);
    run('go', ['mod', 'tidy'], target);

    api = start(process.execPath, [path.join(here, 'mock-api.mjs')], e2eRoot, { PORT: String(apiPort) });
    await waitFor(`http://127.0.0.1:${apiPort}/v1/projects`);

    const dist = path.join(target, 'dist-e2e');
    run('go', ['run', '.', '--export', dist], target, {
      API_SERVER_URL: `http://127.0.0.1:${apiPort}/v1`,
      DOCS_LOCALE: docsLocale,
      DOCS_SEARCH_BACKEND: 'pagefind',
    });
    const pagefindBin = process.platform === 'win32'
      ? path.join(e2eRoot, 'node_modules', '.bin', 'pagefind.cmd')
      : path.join(e2eRoot, 'node_modules', '.bin', 'pagefind');
    if (process.platform === 'win32') {
      run(process.env.ComSpec || 'cmd.exe', ['/d', '/s', '/c', `npx --no-install pagefind --site "${dist}"`], e2eRoot);
    } else {
      run(pagefindBin, ['--site', dist], e2eRoot);
    }
    staticSite = start(process.execPath, [path.join(here, 'static-server.mjs'), dist], e2eRoot, { PORT: String(staticPort) });
    await waitFor(`http://127.0.0.1:${staticPort}/`);

    const appBinary = path.join(target, 'e2e-docs.exe');
    run('go', ['build', '-o', appBinary, '.'], target);
    app = start(appBinary, [], target, {
      PORT: String(appPort),
      API_SERVER_URL: `http://127.0.0.1:${apiPort}/v1`,
      DOCS_LOCALE: docsLocale,
    });
    await waitFor(`http://127.0.0.1:${appPort}/`);

    const manualRoot = path.join(e2eRoot, 'fixtures', 'non-cli');
    const manualBinary = path.join(target, 'non-cli-fixture.exe');
    run('go', ['build', '-o', manualBinary, '.'], manualRoot);
    const manualDist = path.join(target, 'manual-dist');
    run(manualBinary, ['--export', manualDist], manualRoot, {
      API_SERVER_URL: `http://127.0.0.1:${apiPort}/v1`,
      DOCS_LOCALE: docsLocale,
    });
    manualStatic = start(process.execPath, [path.join(here, 'static-server.mjs'), manualDist], e2eRoot, { PORT: String(manualStaticPort) });
    await waitFor(`http://127.0.0.1:${manualStaticPort}/`);

    manual = start(manualBinary, [], manualRoot, {
      PORT: String(manualPort),
      API_SERVER_URL: `http://127.0.0.1:${apiPort}/v1`,
      DOCS_LOCALE: docsLocale,
    });
    await waitFor(`http://127.0.0.1:${manualPort}/`);

    await fs.writeFile(runtimePath, JSON.stringify({
      target,
      dist,
      appPid: app.pid,
      apiPid: api.pid,
      staticPid: staticSite.pid,
      appURL: `http://127.0.0.1:${appPort}`,
      apiURL: `http://127.0.0.1:${apiPort}`,
      staticURL: `http://127.0.0.1:${staticPort}`,
      manualURL: `http://127.0.0.1:${manualPort}`,
      manualStaticURL: `http://127.0.0.1:${manualStaticPort}`,
      manualPid: manual.pid,
      manualStaticPid: manualStatic.pid,
    }, null, 2));
  } catch (error) {
    stop(app);
    stop(staticSite);
    stop(api);
    stop(manual);
    stop(manualStatic);
    await fs.rm(target, { recursive: true, force: true });
    throw error;
  }
}
