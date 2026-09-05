// The only KaTeX file every page loads.
//
// The renderer is 266KB and most documentation pages contain no math at all, so
// loading it everywhere would be the largest thing on the page and do nothing.
// This checks for a placeholder first and pulls the renderer in only when there
// is something to render.
//
// It also watches for placeholders arriving later, because the docs shell swaps
// pages without a reload: a reader who lands on a page with no math and
// navigates to one that has it still gets the renderer.

const SELECTOR = '[data-fastr-docs-math]';

// The renderer and stylesheet sit next to this file, wherever the project
// mounted the asset prefix, so derive the location instead of hardcoding it.
function assetDir() {
  const self =
    document.currentScript ||
    document.querySelector('script[src*="katex/katex-loader.js"]');
  const src = self && self.src;
  if (!src) return '/__fastr-docs/katex/';
  return src.slice(0, src.lastIndexOf('/') + 1);
}

// currentScript is only set while this script executes.
const DIR = assetDir();
let started = false;

function load() {
  if (started) return;
  started = true;
  // The stylesheet goes first so it downloads alongside the renderer rather
  // than after it parses, which would show one frame of unstyled math.
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = DIR + 'katex.css';
  document.head.appendChild(link);

  const script = document.createElement('script');
  script.src = DIR + 'katex.js';
  script.defer = true;
  document.head.appendChild(script);
}

function check() {
  if (document.querySelector(SELECTOR)) {
    load();
    return true;
  }
  return false;
}

function watch() {
  if (check()) return;
  if (!window.MutationObserver) return;
  const observer = new MutationObserver(() => {
    if (check()) observer.disconnect();
  });
  observer.observe(document.body, { childList: true, subtree: true });
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', watch, { once: true });
} else {
  watch();
}
