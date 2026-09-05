// Host side of the Mermaid bridge.
//
// The page renders a placeholder carrying the diagram source. This turns each
// one into a sandboxed iframe, hands it the source over postMessage, and sizes
// the frame from the height the frame reports back.
//
// The source never reaches the page's own DOM as markup, and the frame never
// reaches the page: sandbox="allow-scripts" without allow-same-origin gives it
// an opaque origin.
(function () {
  var PROTOCOL_VERSION = 1;
  var SELECTOR = '[data-fastr-docs-mermaid]';
  var frames = new WeakMap();

  function frameSrc(root) {
    return root.getAttribute('data-fastr-docs-mermaid-frame') || '/__fastr-docs/mermaid/diagram.html';
  }

  function currentTheme() {
    var attr = document.documentElement.getAttribute('data-color-scheme');
    if (attr === 'dark' || attr === 'light') return attr;
    return window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  function send(iframe, method, params) {
    if (!iframe || !iframe.contentWindow) return;
    iframe.contentWindow.postMessage({ v: PROTOCOL_VERSION, src: 'mermaid-host', method: method, params: params }, '*');
  }

  function mount(root) {
    if (frames.has(root)) return;
    var source = root.getAttribute('data-fastr-docs-mermaid') || root.textContent || '';
    var iframe = document.createElement('iframe');
    iframe.className = 'fastr-docs-mermaid__frame';
    // No allow-same-origin: that is what makes the frame an opaque origin and
    // keeps its relaxed style policy from reaching this document.
    iframe.setAttribute('sandbox', 'allow-scripts');
    iframe.setAttribute('loading', 'lazy');
    iframe.setAttribute('title', root.getAttribute('data-fastr-docs-mermaid-title') || 'Diagram');
    iframe.setAttribute('scrolling', 'no');
    iframe.src = frameSrc(root);
    root.textContent = '';
    root.appendChild(iframe);
    frames.set(root, { iframe: iframe, source: source, ready: false });
  }

  function entryFor(iframe) {
    var roots = document.querySelectorAll(SELECTOR);
    for (var i = 0; i < roots.length; i++) {
      var entry = frames.get(roots[i]);
      if (entry && entry.iframe === iframe) return { root: roots[i], entry: entry };
    }
    return null;
  }

  function onMessage(event) {
    var data = event.data;
    if (!data || data.v !== PROTOCOL_VERSION || data.src !== 'mermaid-frame') return;
    var roots = document.querySelectorAll(SELECTOR);
    for (var i = 0; i < roots.length; i++) {
      var entry = frames.get(roots[i]);
      if (!entry || !entry.iframe.contentWindow || entry.iframe.contentWindow !== event.source) continue;
      if (data.method === 'ready') {
        entry.ready = true;
        send(entry.iframe, 'render', { source: entry.source, theme: currentTheme() });
      } else if (data.method === 'resize' && data.params && data.params.height > 0) {
        entry.iframe.style.height = data.params.height + 'px';
      } else if (data.method === 'error') {
        roots[i].setAttribute('data-fastr-docs-mermaid-failed', 'true');
      }
      return;
    }
  }

  function rerenderAll() {
    var roots = document.querySelectorAll(SELECTOR);
    for (var i = 0; i < roots.length; i++) {
      var entry = frames.get(roots[i]);
      if (entry && entry.ready) send(entry.iframe, 'render', { source: entry.source, theme: currentTheme() });
    }
  }

  function mountAll() {
    var roots = document.querySelectorAll(SELECTOR);
    for (var i = 0; i < roots.length; i++) mount(roots[i]);
  }

  function init() {
    window.addEventListener('message', onMessage);
    mountAll();
    // The docs shell swaps pages without a reload, so new diagrams appear
    // after the first mount.
    if (window.MutationObserver) {
      new MutationObserver(function () { mountAll(); }).observe(document.body, { childList: true, subtree: true });
    }
    // Theme is mirrored by re-rendering: the frame cannot read the host's CSS.
    if (window.MutationObserver) {
      new MutationObserver(rerenderAll).observe(document.documentElement, {
        attributes: true, attributeFilter: ['data-color-scheme'],
      });
    }
    if (window.matchMedia) {
      var media = window.matchMedia('(prefers-color-scheme: dark)');
      if (media.addEventListener) media.addEventListener('change', rerenderAll);
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init, { once: true });
  } else {
    init();
  }
})();
