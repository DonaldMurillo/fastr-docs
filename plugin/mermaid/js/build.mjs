// Bundles the in-frame renderer.
//
// Code splitting is the point of this build. Mermaid loads each diagram type
// through a dynamic import, so with format 'esm' and splitting on, esbuild
// emits them as separate chunks and the frame downloads only the ones a
// diagram actually uses instead of all 3.4MB.
//
// That matters more here than in a normal page. Every frame is a distinct
// opaque origin, so it gets its own HTTP cache partition: two diagrams on a
// page cannot share a download, and neither can a reload. Shrinking what a
// single frame needs is the only lever there is.
//
// One directory, no inline script, no network fetch: the frame document is
// served with script-src 'self', and the sandbox gives it an opaque origin, so
// a CDN import or an inline <script> would both be blocked.
import { build } from 'esbuild';
import { mkdir, rm } from 'node:fs/promises';

const OUT = '../assets/frame';
await rm(OUT, { recursive: true, force: true });
await mkdir(OUT, { recursive: true });

const result = await build({
  entryPoints: ['src/frame.js'],
  outdir: OUT,
  bundle: true,
  format: 'esm',
  splitting: true,
  target: ['es2020'],
  minify: true,
  legalComments: 'none',
  metafile: true,
  logLevel: 'error',
});

const outputs = Object.entries(result.metafile.outputs);
const total = outputs.reduce((n, [, o]) => n + o.bytes, 0);
const entry = outputs.find(([name]) => name.endsWith('frame.js'));
console.log(`${outputs.length} files, ${(total / 1e6).toFixed(2)}MB total, entry ${(entry[1].bytes / 1e3).toFixed(1)}kb`);
