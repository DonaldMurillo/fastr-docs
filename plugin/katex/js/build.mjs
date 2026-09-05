// Bundles the page-side KaTeX runtime and copies the stylesheet and fonts.
//
// Unlike the Mermaid plugin there is no frame here. KaTeX's DOM builder applies
// layout through CSSOM (node.style.height = ...), which CSP does not police, so
// math renders directly in the page under default-src 'self'. That is what lets
// inline math sit on the text baseline instead of inside an iframe.
import { build } from 'esbuild';
import { mkdir, readFile, writeFile, readdir, copyFile, rm } from 'node:fs/promises';

const OUT = '../assets';

// The placeholder the Go side emits, styled here so the plugin owns its own
// presentation and a project needs no CSS of its own. Display math scrolls
// rather than overflowing: a wide matrix should not widen the article.
const PLACEHOLDER_CSS = `
.fastr-docs-math--display{display:block;margin:1.25rem 0;text-align:center;overflow-x:auto;overflow-y:hidden}
.fastr-docs-math[data-fastr-docs-math-failed]{color:#b3261e;font-family:ui-monospace,monospace;font-size:.9em;white-space:pre-wrap}
.fastr-docs-math:not([data-fastr-docs-math-rendered]){font-family:ui-monospace,monospace;opacity:.7}
`;
await rm(OUT, { recursive: true, force: true });
await mkdir(OUT + '/fonts', { recursive: true });

// Two entry points, not one. katex-loader.js is what every page loads; it
// pulls katex.js in only on a page that actually has math.
await build({
  entryPoints: ['src/loader.js'],
  outfile: OUT + '/katex-loader.js',
  bundle: true,
  format: 'iife',
  target: ['es2020'],
  minify: true,
  legalComments: 'none',
  logLevel: 'info',
});

await build({
  entryPoints: ['src/page.js'],
  outfile: OUT + '/katex.js',
  bundle: true,
  format: 'iife',
  target: ['es2020'],
  minify: true,
  legalComments: 'none',
  logLevel: 'info',
});

// Only woff2 ships. Every browser that supports MathML-era KaTeX supports it,
// and keeping woff and ttf as well would triple the font payload for nothing.
const css = await readFile('node_modules/katex/dist/katex.min.css', 'utf8');
const trimmed = css.replace(/src:([^;}]*)/g, (whole, sources) => {
  const woff2 = sources.split(',').find((part) => part.includes('.woff2'));
  return woff2 ? 'src:' + woff2.trim() : whole;
});
await writeFile(OUT + '/katex.css', trimmed + PLACEHOLDER_CSS);

const fonts = (await readdir('node_modules/katex/dist/fonts')).filter((name) => name.endsWith('.woff2'));
for (const name of fonts) {
  await copyFile('node_modules/katex/dist/fonts/' + name, OUT + '/fonts/' + name);
}
console.log(`katex.css rewritten to woff2 only, ${fonts.length} font files copied`);
