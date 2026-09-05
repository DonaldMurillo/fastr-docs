// Page-side KaTeX runtime.
//
// This renders in the page itself, with no iframe and no CSP relaxation. KaTeX
// builds its layout as DOM nodes and assigns sizes through CSSOM
// (node.style.height = "0.68em"), and CSP does not police CSSOM, only style
// attributes in parsed markup and <style> elements. So the strict
// default-src 'self' policy the docs pages ship with holds as written.
//
// The one exception is KaTeX's own error path, which calls
// setAttribute("style", "color:...") and would be blocked. Errors are caught
// here and rendered as our own markup instead, so a bad formula never trips
// the policy.

import katex from 'katex';

const SELECTOR = '[data-fastr-docs-math]';
const RENDERED = 'data-fastr-docs-math-rendered';

// The stylesheet and fonts sit next to this bundle, wherever the project
// mounted the asset prefix, so derive the location instead of hardcoding it.
function assetDir() {
  const self =
    document.currentScript ||
    document.querySelector('script[src*="katex/katex.js"]');
  const src = self && self.src;
  if (!src) return '/__fastr-docs/katex/';
  return src.slice(0, src.lastIndexOf('/') + 1);
}

// A <link> element is an external same-origin stylesheet, which style-src
// 'self' allows. Injecting the CSS as a <style> block would not be.
function loadStylesheet(dir) {
  const href = dir + 'katex.css';
  if (document.querySelector('link[rel="stylesheet"][href="' + href + '"]')) return;
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = href;
  document.head.appendChild(link);
}

function showError(root, message) {
  root.textContent = message;
  root.setAttribute('data-fastr-docs-math-failed', 'true');
}

function renderOne(root) {
  if (root.hasAttribute(RENDERED)) return;
  const tex = root.getAttribute('data-fastr-docs-math') || '';
  const displayMode = root.getAttribute('data-fastr-docs-math-display') === 'true';
  root.setAttribute(RENDERED, 'true');
  try {
    katex.render(tex, root, {
      displayMode,
      // Left on so a formula error is an exception here rather than KaTeX
      // writing a style attribute the page policy would block.
      throwOnError: true,
      output: 'htmlAndMathml',
      strict: false,
      trust: false,
    });
  } catch (error) {
    showError(root, String((error && error.message) || error));
  }
}

function renderAll() {
  const roots = document.querySelectorAll(SELECTOR);
  for (let i = 0; i < roots.length; i++) renderOne(roots[i]);
}

// currentScript is only set while this script executes, so resolve the asset
// directory now and pull the stylesheet in straight away rather than waiting
// for DOMContentLoaded, which would show one frame of unstyled math.
const DIR = assetDir();
loadStylesheet(DIR);

function init() {
  renderAll();
  // The docs shell swaps pages without a reload, so formulas appear after the
  // first pass.
  if (window.MutationObserver) {
    new MutationObserver(renderAll).observe(document.body, { childList: true, subtree: true });
  }
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init, { once: true });
} else {
  init();
}
