// Bundles the in-frame renderer into a single IIFE with Mermaid inlined.
//
// One file, no inline script, no network fetch: the frame document is served
// with script-src 'self', and the sandbox gives it an opaque origin, so a CDN
// import or an inline <script> would both be blocked.
import { build } from 'esbuild';
import { mkdir } from 'node:fs/promises';

await mkdir('../assets', { recursive: true });
await build({
  entryPoints: ['src/frame.js'],
  outfile: '../assets/diagram.js',
  bundle: true,
  format: 'iife',
  target: ['es2020'],
  minify: true,
  legalComments: 'none',
  logLevel: 'info',
});
