(() => {
  const initTocSelect = () => {
    if (window.__fastrDocsTocCleanup) window.__fastrDocsTocCleanup();
    const select = document.querySelector('[data-docs-toc-select]');
    const rail = document.querySelector('.fastr-docs-toc--rail');
    if (!select || !rail) return;

    let requestedHref = '';
    const sync = () => {
      const active = rail.querySelector('a[aria-current="true"]');
      const activeHref = active?.getAttribute('href');
      if (!activeHref || (requestedHref && activeHref !== requestedHref)) return;
      select.value = activeHref;
      requestedHref = '';
    };
    const observer = new MutationObserver(sync);
    observer.observe(rail, { subtree: true, attributes: true, attributeFilter: ['aria-current'] });
    const onChange = event => {
      const href = event.target.value;
      if (!href || !href.startsWith('#')) return;
      const target = document.querySelector(href);
      if (!target) return;
      requestedHref = href;
      select.value = href;
      history.pushState(null, '', href);
      target.scrollIntoView({ behavior: 'smooth', block: 'start' });
    };
    select.addEventListener('change', onChange);
    window.__fastrDocsTocCleanup = () => {
      observer.disconnect();
      select.removeEventListener('change', onChange);
      window.__fastrDocsTocCleanup = null;
    };
    sync();
  };

  // GoFastr owns Sidebar, its drawer, active navigation, and AnchoredRail
  // scrollspy. This runtime is deliberately limited to the docs-specific
  // local search behavior so those framework interactions never diverge.
  const init = () => { initTocSelect(); };
  init();
  window.addEventListener('gofastr:navigate', init);
})();
