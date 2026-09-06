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
const selfSitePort = 4180;
const devPort = 4181;
const selfPrefixPort = 4182;
const docsLocale = 'pt-BR';
let goCommand = 'go';

const resolveGoCommand = async () => {
  const candidates = [
    process.env.GO_BINARY,
    process.platform === 'win32' && process.env.GOROOT
      ? path.join(process.env.GOROOT, 'bin', 'go.exe')
      : null,
    process.platform === 'win32' && process.env.ProgramFiles
      ? path.join(process.env.ProgramFiles, 'Go', 'bin', 'go.exe')
      : null,
    process.platform === 'win32' && process.env.LOCALAPPDATA
      ? path.join(process.env.LOCALAPPDATA, 'Programs', 'Go', 'bin', 'go.exe')
      : null,
    process.platform === 'win32' && process.env.USERPROFILE
      ? path.join(process.env.USERPROFILE, 'sdk', 'go', 'bin', 'go.exe')
      : null,
    'go',
  ].filter(Boolean);
  for (const candidate of candidates) {
    if (candidate !== 'go') {
      try {
        const stat = await fs.stat(candidate);
        if (stat.isFile()) return candidate;
      } catch {}
      continue;
    }
    try {
      execFileSync(candidate, ['version'], { stdio: 'ignore', windowsHide: true });
      return candidate;
    } catch {}
  }
  return 'go';
};

const run = (command, args, cwd, env = {}) => execFileSync(command, args, {
  cwd,
  env: { ...process.env, ...env },
  stdio: 'pipe',
  encoding: 'utf8',
  windowsHide: true,
});

const start = (command, args, cwd, env = {}, stdio = 'ignore') => spawn(command, args, {
  cwd,
  env: { ...process.env, ...env },
  stdio,
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

// The dev loop compiles the generated project before it serves anything, and
// that project grows as the starter template does. 60s was already marginal.
const waitFor = async (url, timeout = 180_000) => {
  const started = Date.now();
  while (Date.now() - started < timeout) {
    try {
      const response = await fetch(url);
      if (response.status >= 200 && response.status < 400) return;
    } catch {}
    await new Promise((resolve) => setTimeout(resolve, 250));
  }
  throw new Error(`Timed out waiting for ${url}`);
};

export default async function globalSetup() {
  goCommand = await resolveGoCommand();
  const target = await fs.mkdtemp(path.join(os.tmpdir(), 'fastr-docs-e2e-'));
  let api;
  let staticSite;
  let app;
  let manual;
  let manualStatic;
  let selfSite;
  let selfPrefixStatic;
  let dev;
  try {
    run(goCommand, ['run', './cmd/fastr-docs', 'init', target, '--name', 'E2E Docs', '--module', 'example.com/e2e-docs'], repoRoot);
    run(goCommand, ['mod', 'tidy'], target);

    api = start(process.execPath, [path.join(here, 'mock-api.mjs')], e2eRoot, { PORT: String(apiPort) });
    await waitFor(`http://127.0.0.1:${apiPort}/v1/projects`);

    const dist = path.join(target, 'dist-e2e');
    run(goCommand, ['run', '.', '--export', dist], target, {
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
    run(goCommand, ['build', '-o', appBinary, '.'], target);
    app = start(appBinary, [], target, {
      PORT: String(appPort),
      API_SERVER_URL: `http://127.0.0.1:${apiPort}/v1`,
      DOCS_LOCALE: docsLocale,
    });
    await waitFor(`http://127.0.0.1:${appPort}/`);

    dev = start(goCommand, ['run', './cmd/fastr-docs', 'dev', target, '--addr', `127.0.0.1:${devPort}`, '--no-a11y'], repoRoot, {
      PORT: String(devPort),
    }, 'inherit');
    await waitFor(`http://127.0.0.1:${devPort}/`);

    const manualRoot = path.join(e2eRoot, 'fixtures', 'non-cli');
    const manualBinary = path.join(target, 'non-cli-fixture.exe');
    run(goCommand, ['build', '-o', manualBinary, '.'], manualRoot);
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

    const selfRoot = path.join(repoRoot, 'site');
    const selfBinary = path.join(target, 'fastr-docs-self.exe');
    run(goCommand, ['build', '-o', selfBinary, '.'], selfRoot);
    selfSite = start(selfBinary, [], selfRoot, {
      PORT: String(selfSitePort),
      API_SERVER_URL: `http://127.0.0.1:${apiPort}/v1`,
      PUBLIC_SITE_URL: `http://127.0.0.1:${selfSitePort}`,
    });
    await waitFor(`http://127.0.0.1:${selfSitePort}/`);

    // The same site exported below a prefix, the shape of a GitHub project
    // page, served from a folder whose name is the prefix so the URLs carry
    // it. Pagefind indexes the prefixed folder, so its result URLs are
    // root-relative and the runtime has to add the base itself.
    const selfPrefix = '/prefix';
    const selfPrefixRoot = path.join(target, 'self-prefix');
    const selfPrefixDist = path.join(selfPrefixRoot, 'prefix');
    run(selfBinary, ['--export', selfPrefixDist, '--export-base', selfPrefix], selfRoot, {
      API_SERVER_URL: `http://127.0.0.1:${apiPort}/v1`,
      PUBLIC_SITE_URL: `http://127.0.0.1:${selfPrefixPort}${selfPrefix}`,
      DOCS_SEARCH_BACKEND: 'pagefind',
    });
    if (process.platform === 'win32') {
      run(process.env.ComSpec || 'cmd.exe', ['/d', '/s', '/c', `npx --no-install pagefind --site "${selfPrefixDist}"`], e2eRoot);
    } else {
      run(pagefindBin, ['--site', selfPrefixDist], e2eRoot);
    }
    selfPrefixStatic = start(process.execPath, [path.join(here, 'static-server.mjs'), selfPrefixRoot], e2eRoot, { PORT: String(selfPrefixPort) });
    await waitFor(`http://127.0.0.1:${selfPrefixPort}${selfPrefix}/`);

    await fs.writeFile(runtimePath, JSON.stringify({
      target,
      dist,
      appPid: app.pid,
      devPid: dev.pid,
      apiPid: api.pid,
      staticPid: staticSite.pid,
      appURL: `http://127.0.0.1:${appPort}`,
      devURL: `http://127.0.0.1:${devPort}`,
      apiURL: `http://127.0.0.1:${apiPort}`,
      staticURL: `http://127.0.0.1:${staticPort}`,
      manualURL: `http://127.0.0.1:${manualPort}`,
      manualStaticURL: `http://127.0.0.1:${manualStaticPort}`,
      selfURL: `http://127.0.0.1:${selfSitePort}`,
      selfPrefixURL: `http://127.0.0.1:${selfPrefixPort}`,
      selfPrefix,
      selfPrefixPid: selfPrefixStatic.pid,
      manualPid: manual.pid,
      manualStaticPid: manualStatic.pid,
      selfPid: selfSite.pid,
    }, null, 2));
  } catch (error) {
    stop(app);
    stop(dev);
    stop(staticSite);
    stop(api);
    stop(manual);
    stop(manualStatic);
    stop(selfSite);
    stop(selfPrefixStatic);
    await fs.rm(target, { recursive: true, force: true });
    throw error;
  }
}
