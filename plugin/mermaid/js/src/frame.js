// The in-frame Mermaid renderer.
//
// This runs inside <iframe sandbox="allow-scripts"> with no allow-same-origin,
// so the document has an opaque origin: cookies, storage, and the host DOM are
// all unreachable. window.parent.postMessage is the only channel, and it is
// deliberately one narrow one.
//
// It is read-only by design. The editor plugin in gofastr-plugins has a save
// protocol and a capability handshake; a documentation page needs neither, so
// the whole protocol here is "host sends a diagram, frame reports its height".

import mermaid from 'mermaid';

const PROTOCOL_VERSION = 1;
const ROOT_ID = 'fastr-docs-mermaid-root';

let rendered = 0;

// targetOrigin is "*" because the parent's origin is opaque to a sandboxed
// frame and must not be guessed. Nothing sensitive crosses this boundary: the
// diagram source came from the host in the first place.
function post(method, params) {
  try {
    window.parent.postMessage({ v: PROTOCOL_VERSION, src: 'mermaid-frame', method, params }, '*');
  } catch {
    // A closed or detached parent is not recoverable and not worth reporting.
  }
}

function reportHeight() {
  const root = document.getElementById(ROOT_ID);
  if (!root) return;
  const height = Math.ceil(root.getBoundingClientRect().height);
  if (height > 0) post('resize', { height });
}

function showError(message) {
  const root = document.getElementById(ROOT_ID);
  if (root) {
    root.textContent = '';
    const box = document.createElement('p');
    box.className = 'fastr-docs-mermaid-error';
    // textContent, never innerHTML: the message can carry the diagram source.
    box.textContent = message;
    root.appendChild(box);
  }
  post('error', { message });
  reportHeight();
}

async function render(source, theme) {
  const root = document.getElementById(ROOT_ID);
  if (!root) return;
  const text = String(source || '').trim();
  if (text === '') {
    root.textContent = '';
    reportHeight();
    return;
  }
  mermaid.initialize({
    startOnLoad: false,
    securityLevel: 'strict',
    theme: theme === 'dark' ? 'dark' : 'default',
    fontFamily: 'inherit',
  });
  rendered += 1;
  try {
    const { svg } = await mermaid.render('fastr-docs-mermaid-' + rendered, text);
    root.innerHTML = svg;
  } catch (error) {
    showError(String((error && error.message) || error));
    return;
  }
  // Mermaid sizes the SVG after insertion, so measure on the next frame.
  requestAnimationFrame(reportHeight);
}

function onMessage(event) {
  // Anything that is not the embedding page is ignored outright.
  if (event.source !== window.parent) return;
  const data = event.data;
  if (!data || data.v !== PROTOCOL_VERSION || data.src !== 'mermaid-host') return;
  if (data.method === 'render') {
    void render(data.params && data.params.source, data.params && data.params.theme);
  }
}

function boot() {
  window.addEventListener('message', onMessage);
  window.addEventListener('resize', reportHeight);
  // The host cannot know when this bundle finished parsing, so the frame
  // speaks first.
  post('ready', {});
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', boot, { once: true });
} else {
  boot();
}
