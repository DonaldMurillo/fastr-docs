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
  var pending = new WeakSet();

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
    // The frame says what it is while it loads and after: assistive tech
    // gets a name, and the reader knows a diagram is coming.
    root.setAttribute('aria-busy', 'true');
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
        // The diagram has drawn: the container stops announcing itself as
        // busy the first time it reports a size.
        roots[i].removeAttribute('aria-busy');
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

  // The bundle behind each frame is 3.4MB, so a frame is created only when its
  // diagram approaches the viewport.
  //
  // iframe loading="lazy" is not enough on its own: Chrome's threshold is
  // generous enough that both diagrams on a normal page load immediately, which
  // is what this measured before the observer existed.
  var visible = window.IntersectionObserver
    ? new IntersectionObserver(function (entries) {
        for (var i = 0; i < entries.length; i++) {
          if (!entries[i].isIntersecting) continue;
          visible.unobserve(entries[i].target);
          mount(entries[i].target);
        }
      }, { rootMargin: '600px 0px' })
    : null;

  function mountAll() {
    var roots = document.querySelectorAll(SELECTOR);
    for (var i = 0; i < roots.length; i++) {
      if (frames.has(roots[i]) || pending.has(roots[i])) continue;
      if (!visible) {
        mount(roots[i]);
        continue;
      }
      pending.add(roots[i]);
      visible.observe(roots[i]);
    }
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

  // Printing has no viewport to scroll, so every diagram is forced in first.
  function mountEverything() {
    var roots = document.querySelectorAll(SELECTOR);
    for (var i = 0; i < roots.length; i++) mount(roots[i]);
  }

  if (window.matchMedia) {
    var print = window.matchMedia('print');
    if (print.addEventListener) print.addEventListener('change', function (e) { if (e.matches) mountEverything(); });
  }
  window.addEventListener('beforeprint', mountEverything);

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init, { once: true });
  } else {
    init();
  }
})();
