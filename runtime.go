package docs

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// RuntimeJS is the small docs-owned browser layer. GoFastr still owns its
// core runtime; this file only wires docs-specific controls such as the
// responsive in-page selector, publication search, and active anchor state.
func RuntimeJS() string { return docsRuntimeJS }

const docsRuntimeJS = `(function(){
  function initTocSelect(root){
    var select = root.tagName === 'SELECT' ? root : root.querySelector('select,[data-docs-toc-select]') || root;
    if (!select || select.dataset.docsTocReady === 'true') return;
    if (window.__fastrDocsTocCleanup) window.__fastrDocsTocCleanup();
    select.dataset.docsTocReady = 'true';
    var rail = document.querySelector('.fastr-docs-toc--rail');
    var requestedHref = '';
    function sync(){
      if (!rail) return;
      var active = rail.querySelector('a[aria-current="true"]');
      var activeHref = active && active.getAttribute('href');
      if (!activeHref || (requestedHref && activeHref !== requestedHref)) return;
      select.value = activeHref;
      requestedHref = '';
    }
    var observer = null;
    if (rail && window.MutationObserver) {
      observer = new MutationObserver(sync);
      observer.observe(rail, {subtree:true, attributes:true, attributeFilter:['aria-current']});
    }
    function onChange(){
      var href = select.value;
      if (!href || href.charAt(0) !== '#') return;
      if (!select.value) return;
      var target = document.querySelector(href);
      if (!target) return;
      requestedHref = href;
      window.history.replaceState(null, '', href);
      if (target) target.scrollIntoView({behavior:'smooth', block:'start'});
    }
    select.addEventListener('change', onChange);
    sync();
    window.__fastrDocsTocCleanup = function(){
      if (observer) observer.disconnect();
      select.removeEventListener('change', onChange);
      delete select.dataset.docsTocReady;
      window.__fastrDocsTocCleanup = null;
    };
  }
  function normalizeDocsPath(value){
    var path = String(value || '/').split(/[?#]/, 1)[0] || '/';
    return path.length > 1 ? path.replace(/\/+$/, '') : path;
  }
  function syncDocsDrawerTrigger(){
    var currentPath = normalizeDocsPath(location.pathname);
    document.querySelectorAll('.fastr-docs-mobile-nav-trigger[data-fastr-docs-blog-prefixes]').forEach(function(trigger){
      var prefixes = (trigger.getAttribute('data-fastr-docs-blog-prefixes') || '').split(',').filter(Boolean);
      var matchingPrefix = '';
      prefixes.forEach(function(prefix){
        prefix = normalizeDocsPath(prefix);
        if ((currentPath === prefix || currentPath.indexOf(prefix + '/') === 0) && prefix.length > matchingPrefix.length) matchingPrefix = prefix;
      });
      var drawer = trigger.getAttribute('data-fastr-docs-global-drawer');
      if (matchingPrefix){
        drawer = trigger.getAttribute('data-fastr-docs-blog-drawer') || drawer;
        var mappings = (trigger.getAttribute('data-fastr-docs-blog-drawers') || '').split(';').filter(Boolean);
        mappings.forEach(function(mapping){
          var separator = mapping.indexOf('=');
          if (separator < 0) return;
          var prefix = normalizeDocsPath(mapping.slice(0, separator));
          if (prefix === matchingPrefix) drawer = mapping.slice(separator + 1);
        });
      }
      if (drawer) trigger.setAttribute('data-fui-open', drawer);
    });
  }
  function syncDocsSidebar(sidebar){
    if (!sidebar) return;
    var currentPath = normalizeDocsPath(location.pathname);
    var activeLink = null;
    sidebar.querySelectorAll('a.ui-sidebar__link').forEach(function(link){
      var href = link.getAttribute('href');
      var linkPath = '';
      try { linkPath = normalizeDocsPath(new URL(href || '', document.baseURI).pathname); } catch (_) {}
      var active = !!linkPath && linkPath === currentPath;
      if (active) activeLink = link;
      if (active) {
        link.setAttribute('aria-current', 'page');
        link.classList.add('active');
      } else {
        link.removeAttribute('aria-current');
        link.classList.remove('active');
      }
    });
    sidebar.querySelectorAll('details.ui-sidebar__group').forEach(function(group){
      var summary = group.querySelector('summary.ui-sidebar__link');
      var routeIcon = summary && summary.querySelector('svg[data-fastr-docs-nav-path]');
      var groupPath = '';
      try { groupPath = normalizeDocsPath(new URL(routeIcon && routeIcon.getAttribute('data-fastr-docs-nav-path') || '', document.baseURI).pathname); } catch (_) {}
      var activeParent = !!groupPath && groupPath === currentPath;
      if (routeIcon) routeIcon.toggleAttribute('data-fastr-docs-active', activeParent);
      if (summary) summary.classList.toggle('fastr-docs-nav-group--active', activeParent);
      var active = activeParent || (!!activeLink && group.contains(activeLink));
      if (!active) return;
      group.open = true;
      if (summary) summary.setAttribute('aria-expanded', 'true');
    });
    sidebar.querySelectorAll('[data-fui-sidebar-group-toggle]').forEach(function(toggle){
      var targetID = toggle.getAttribute('aria-controls');
      var target = targetID && document.getElementById(targetID);
      var routeIcon = toggle.querySelector('svg[data-fastr-docs-nav-path]');
      var groupPath = '';
      try { groupPath = normalizeDocsPath(new URL(routeIcon && routeIcon.getAttribute('data-fastr-docs-nav-path') || '', document.baseURI).pathname); } catch (_) {}
      var activeParent = !!groupPath && groupPath === currentPath;
      var active = activeParent || (!!target && !!target.querySelector('a.ui-sidebar__link[aria-current="page"], a.ui-sidebar__link.active'));
      if (!active) return;
      if (routeIcon) routeIcon.toggleAttribute('data-fastr-docs-active', activeParent);
      toggle.setAttribute('aria-expanded', 'true');
      target.hidden = false;
    });
  }
  function syncDocsSidebars(){
    document.querySelectorAll('.ui-sidebar--persistent, [data-fui-widget="fastr-docs-sections"]').forEach(syncDocsSidebar);
  }
  function initSidebarState(){
    syncDocsSidebars();
    if (window.fastrDocsSidebarStateReady || !document.body || !window.MutationObserver) return;
    window.fastrDocsSidebarStateReady = true;
    var observer = new MutationObserver(function(records){
      for (var i = 0; i < records.length; i++) {
        var record = records[i];
        if (record.type === 'childList' || record.attributeName === 'hidden') {
          syncDocsSidebars();
          return;
        }
      }
    });
    observer.observe(document.body, {childList:true, subtree:true, attributes:true, attributeFilter:['hidden']});
    window.addEventListener('gofastr:navigate', function(){ setTimeout(syncDocsSidebars, 0); });
    window.addEventListener('fastr:navigate', function(){ setTimeout(syncDocsSidebars, 0); });
  }
  function escapeHTML(value){
    return String(value == null ? '' : value).replace(/[&<>'"]/g, function(char){
      return {'&':'&amp;','<':'&lt;','>':'&gt;',"'":'&#39;','"':'&quot;'}[char];
    });
  }
  function pagefindTrigger(){
    return document.querySelector('[data-fastr-docs-backend="pagefind"]');
  }
  function jsonSearchTrigger(){
    return document.querySelector('[data-fastr-docs-backend="json"]');
  }
  function pagefindList(input){
    if (!input) return null;
    var id = input.getAttribute('aria-controls');
    return id ? document.getElementById(id) : null;
  }
  function pagefindModule(){
    if (window.fastrDocsPagefind) return window.fastrDocsPagefind;
    var trigger = pagefindTrigger();
    if (!trigger) return null;
    var base = trigger.getAttribute('data-fastr-docs-pagefind-path') || '/pagefind/';
    var href = new URL(base.replace(/\/$/, '') + '/pagefind.js', document.baseURI).href;
    window.fastrDocsPagefind = import(href).then(function(module){
      return module.default || module;
    });
    return window.fastrDocsPagefind;
  }
  function restorePalette(list, query){
    if (!list || !list.dataset.fastrDocsInitialOptions) return;
    list.innerHTML = list.dataset.fastrDocsInitialOptions;
    list.setAttribute('data-fui-static-options', '');
    list.removeAttribute('hidden');
    // Restoring replaces the markup the native combobox filter was tracking,
    // so a pending query has to be reapplied here. Without this the palette
    // answers every keystroke with the complete route list whenever the
    // search index is unreachable.
    filterStaticOptions(list, query);
  }
  function filterStaticOptions(list, query){
    var terms = jsonSearchTerms(query);
    var options = list.querySelectorAll('[role="option"]');
    var matched = 0;
    for (var i = 0; i < options.length; i++){
      var option = options[i];
      if (!terms.length){
        option.hidden = false;
        matched++;
        continue;
      }
      var haystack = (option.getAttribute('data-value') || '') + ' ' + (option.textContent || '');
      haystack = haystack.toLowerCase();
      var hit = terms.every(function(term){ return haystack.indexOf(term) !== -1; });
      option.hidden = !hit;
      if (hit) matched++;
    }
    if (!terms.length || matched) return;
    list.innerHTML = '<li role="option" aria-disabled="true" class="fastr-docs-search-empty"><span class="combobox__opt-label">No matching documentation</span></li>';
    list.removeAttribute('data-fui-static-options');
  }
  function renderPagefindResults(list, results){
    if (!list) return;
    if (!results || !results.length){
      list.innerHTML = '<li role="option" aria-disabled="true" class="fastr-docs-search-empty"><span class="combobox__opt-label">No matching documentation</span></li>';
      list.removeAttribute('data-fui-static-options');
      list.removeAttribute('hidden');
      return;
    }
    list.innerHTML = results.map(function(result, index){
      var data = result.data || {};
      var meta = data.meta || {};
      var title = meta.title || data.url || 'Documentation';
		var excerpt = String(data.excerpt || data.url || '').replace(/<[^>]*>/g, '');
      var url = data.url || '#';
      return '<li role="option" id="fastr-docs-command-palette-list-opt-pagefind-' + index + '" data-value="' + escapeHTML(title) + '" data-fui-push-state="' + escapeHTML(url) + '">' +
        '<span class="combobox__opt-label">' + escapeHTML(title) + '</span>' +
        '<span class="combobox__opt-meta">' + escapeHTML(excerpt) + '</span></li>';
    }).join('');
    list.removeAttribute('data-fui-static-options');
    list.removeAttribute('hidden');
  }
  function jsonSearchModule(){
    if (window.fastrDocsSearchIndex) return window.fastrDocsSearchIndex;
    var trigger = jsonSearchTrigger();
    if (!trigger) return null;
    var href = new URL(trigger.getAttribute('data-fastr-docs-index-path') || '/__fastr-docs/search.json', document.baseURI).href;
    window.fastrDocsSearchIndex = fetch(href, {headers:{'Accept':'application/json'}}).then(function(response){
      if (!response.ok) throw new Error('Search index request failed');
      return response.json();
    });
    return window.fastrDocsSearchIndex;
  }
  function jsonSearchTerms(value){
    return String(value || '').toLowerCase().trim().split(/\s+/).filter(Boolean);
  }
  function jsonSearchExcerpt(entry){
		var value = String(entry.description || entry.text || '').replace(/[#*_>\[\]]/g, '').replace(/\s+/g, ' ').trim();
    return value.length > 140 ? value.slice(0, 137) + '...' : value;
  }
  function renderJSONResults(list, entries, query){
    if (!list) return;
    var terms = jsonSearchTerms(query);
    var ranked = (Array.isArray(entries) ? entries : []).map(function(entry, index){
      var haystack = [entry.title, entry.description, entry.tags, entry.headings, entry.text].join(' ').toLowerCase();
      if (terms.some(function(term){ return haystack.indexOf(term) === -1; })) return null;
      var title = String(entry.title || entry.path || 'Documentation');
      var score = terms.reduce(function(total, term){
        if (String(entry.title || '').toLowerCase().indexOf(term) !== -1) total += 20;
        if (String(entry.headings || '').toLowerCase().indexOf(term) !== -1) total += 8;
        if (String(entry.description || '').toLowerCase().indexOf(term) !== -1) total += 5;
        if (haystack.indexOf(term) !== -1) total += 1;
        return total;
      }, 0);
      return {entry: entry, score: score, index: index, title: title};
    }).filter(Boolean).sort(function(left, right){ return right.score - left.score || left.index - right.index; }).slice(0, 8);
    if (!ranked.length){
      list.innerHTML = '<li role="option" aria-disabled="true" class="fastr-docs-search-empty"><span class="combobox__opt-label">No matching documentation</span></li>';
      list.removeAttribute('data-fui-static-options');
      list.removeAttribute('hidden');
      return;
    }
    list.innerHTML = ranked.map(function(result, index){
      var entry = result.entry || {};
      var url = String(entry.path || '#');
      return '<li role="option" id="fastr-docs-command-palette-list-opt-json-' + index + '" data-value="' + escapeHTML(result.title) + '" data-fui-push-state="' + escapeHTML(url) + '">' +
        '<span class="combobox__opt-label">' + escapeHTML(result.title) + '</span>' +
        '<span class="combobox__opt-meta">' + escapeHTML(jsonSearchExcerpt(entry) || url) + '</span></li>';
    }).join('');
    list.removeAttribute('data-fui-static-options');
    list.removeAttribute('hidden');
  }
  function initJSONSearch(){
    if (!jsonSearchTrigger() || window.fastrDocsJSONSearchReady) return;
    window.fastrDocsJSONSearchReady = true;
    var requestID = 0;
    document.addEventListener('input', function(event){
      var input = event.target && event.target.closest && event.target.closest('#fastr-docs-command-palette-input');
      if (!input) return;
      var list = pagefindList(input);
      if (!list) return;
      if (!list.dataset.fastrDocsInitialOptions) list.dataset.fastrDocsInitialOptions = list.innerHTML;
      var query = (input.value || '').trim();
      if (!query){ restorePalette(list); return; }
      var current = ++requestID;
      var index = jsonSearchModule();
      if (!index) return;
      index.then(function(entries){
        if (current === requestID) renderJSONResults(list, entries, query);
      }).catch(function(){
        if (current === requestID) restorePalette(list, query);
      });
    });
  }
  function initPagefindSearch(){
    if (!pagefindTrigger() || window.fastrDocsPagefindSearchReady) return;
    window.fastrDocsPagefindSearchReady = true;
    document.addEventListener('input', function(event){
      var input = event.target && event.target.closest && event.target.closest('#fastr-docs-command-palette-input');
      if (!input) return;
      var list = pagefindList(input);
      if (!list) return;
      if (!list.dataset.fastrDocsInitialOptions) list.dataset.fastrDocsInitialOptions = list.innerHTML;
      var query = (input.value || '').trim();
      if (!query){ restorePalette(list); return; }
      var pagefind = pagefindModule();
      if (!pagefind) return;
      pagefind.then(function(api){
        if (!api || typeof api.search !== 'function') return;
        return api.search(query);
      }).then(function(response){
        if (!response || !Array.isArray(response.results)) return;
        return Promise.all(response.results.slice(0, 8).map(function(result){
          return result.data().then(function(data){ return {data: data}; });
        }));
      }).then(function(results){
        if (results) renderPagefindResults(list, results);
      }).catch(function(){
        // The portable JSON palette remains usable when Pagefind assets are
        // absent in dev or a deployment omitted the optional static bundle.
      });
    });
  }
  function blogSearchTerms(value){
    return String(value || '').toLowerCase().trim().split(/\s+/).filter(Boolean);
  }
  function initBlogSearch(root){
    if (!root || root.dataset.fastrDocsBlogSearchReady === 'true') return;
    var form = root.querySelector('[data-fastr-docs-blog-search-form]');
    var input = root.querySelector('[data-fastr-docs-blog-search-input]');
    var summary = root.querySelector('[data-fastr-docs-blog-search-summary]');
    var empty = root.querySelector('[data-fastr-docs-blog-search-empty]');
    if (!form || !input) return;
    root.dataset.fastrDocsBlogSearchReady = 'true';
    var items = Array.prototype.slice.call(root.querySelectorAll('[data-fastr-docs-blog-search-item]'));
    function currentQuery(){
      try { return new URL(window.location.href).searchParams.get('q') || ''; } catch (_) { return input.value || ''; }
    }
    function apply(query){
      query = String(query || '').trim();
      var terms = blogSearchTerms(query);
      var visible = 0;
      items.forEach(function(item){
        var text = String(item.getAttribute('data-fastr-docs-blog-search-item') || item.textContent || '').toLowerCase();
        var match = !terms.length || terms.every(function(term){ return text.indexOf(term) !== -1; });
        item.hidden = !match;
        if (match) visible++;
      });
      input.value = query;
      if (summary) {
        var resultsFor = root.getAttribute('data-fastr-docs-blog-results-for') || 'Results for “%s”';
        var idleLabel = root.getAttribute('data-fastr-docs-blog-search-label') || 'Search the publication';
        var matchLabel = root.getAttribute('data-fastr-docs-blog-match-summary') || '%d matches';
        var heading = query ? resultsFor.replace('%s', query) : idleLabel;
        summary.textContent = heading + ' · ' + matchLabel.replace('%d', visible);
      }
      if (empty) empty.hidden = visible > 0;
    }
    function onSubmit(event){
      event.preventDefault();
      var query = input.value.trim();
      var url = new URL(form.getAttribute('action') || window.location.pathname, document.baseURI);
      if (query) url.searchParams.set('q', query); else url.search = '';
      window.history.pushState(null, '', url.pathname + url.search + url.hash);
      apply(query);
      input.focus();
    }
    function onPopState(){ apply(currentQuery()); }
    form.addEventListener('submit', onSubmit);
    window.addEventListener('popstate', onPopState);
    apply(currentQuery() || input.value || '');
    window.__fastrDocsBlogSearchCleanup = function(){
      form.removeEventListener('submit', onSubmit);
      window.removeEventListener('popstate', onPopState);
      delete root.dataset.fastrDocsBlogSearchReady;
      window.__fastrDocsBlogSearchCleanup = null;
    };
  }
  function blogShareURL(path){
    try { return new URL(path || (location.pathname + location.search), document.baseURI).href; }
    catch (_) { return location.href; }
  }
  function announceBlogShare(button, message){
    var surface = button.closest && button.closest('[data-fastr-docs-share-surface]');
    var status = surface && surface.querySelector('[data-fastr-docs-share-status]');
    if (!status) return;
    status.textContent = '';
    window.setTimeout(function(){ status.textContent = message; }, 10);
    window.setTimeout(function(){ status.textContent = ''; }, 4500);
  }
  function copyBlogShareURL(url){
    if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
      return navigator.clipboard.writeText(url);
    }
    return new Promise(function(resolve, reject){
      var area = document.createElement('textarea');
      area.value = url;
      area.setAttribute('readonly', '');
      area.style.position = 'fixed';
      area.style.opacity = '0';
      document.body.appendChild(area);
      area.select();
      var copied = false;
      try { copied = document.execCommand('copy'); } catch (_) {}
      area.remove();
      if (copied) resolve(); else reject(new Error('Clipboard is unavailable'));
    });
  }
  function initBlogShare(){
    var buttons = Array.prototype.slice.call(document.querySelectorAll('[data-fastr-docs-share]'));
    if (!buttons.length) return;
    document.querySelectorAll('[data-fastr-docs-share-target]').forEach(function(target){
      target.textContent = blogShareURL(target.getAttribute('data-fastr-docs-share-path') || target.textContent);
    });
    var listeners = [];
    function fallback(button, url){
      copyBlogShareURL(url).then(function(){ announceBlogShare(button, 'Link copied'); })
        .catch(function(){ announceBlogShare(button, 'Copying the link was not available'); });
    }
    buttons.forEach(function(button){
      if (button.dataset.fastrDocsShareReady === 'true') return;
      button.dataset.fastrDocsShareReady = 'true';
      function onClick(event){
        event.preventDefault();
        var url = blogShareURL(button.getAttribute('data-fastr-docs-share-url'));
        var payload = {
          title: button.getAttribute('data-fastr-docs-share-title') || document.title,
          text: button.getAttribute('data-fastr-docs-share-text') || '',
          url: url
        };
        if (!navigator.share || typeof navigator.share !== 'function') {
          fallback(button, url);
          return;
        }
        try {
          var pending = navigator.share(payload);
          if (pending && typeof pending.then === 'function') {
            pending.then(function(){ announceBlogShare(button, 'Post shared'); })
              .catch(function(error){
                if (error && error.name === 'AbortError') return;
                fallback(button, url);
              });
          } else {
            announceBlogShare(button, 'Post shared');
          }
        } catch (_) {
          fallback(button, url);
        }
      }
      button.addEventListener('click', onClick);
      listeners.push(function(){ button.removeEventListener('click', onClick); delete button.dataset.fastrDocsShareReady; });
    });
    window.__fastrDocsBlogShareCleanup = function(){
      listeners.forEach(function(cleanup){ cleanup(); });
      window.__fastrDocsBlogShareCleanup = null;
    };
  }
  function initVariantSelectors(){
    document.querySelectorAll('[data-docs-variant-select]').forEach(function(select){
      if (select.dataset.docsVariantReady === 'true') return;
      select.dataset.docsVariantReady = 'true';
      var currentPath = select.getAttribute('data-docs-current-path') || location.pathname;
      Array.prototype.forEach.call(select.options, function(option){
        if (option.value === currentPath || option.value.replace(/\/$/, '') === location.pathname.replace(/\/$/, '')) {
          select.value = option.value;
        }
      });
      select.addEventListener('change', function(){
        var destination = select.value;
        if (!destination) return;
        // Variant changes can alter the mounted content slice and the
        // document language, so rebuild the shell from the destination URL
        // instead of keeping a stale selector inside an SPA layout.
        location.href = destination;
      });
    });
  }
  function initVariantActiveLinks(){
    var node = document.querySelector('meta[name="fastr-docs-active-families"]');
    if (!node) return;
    var config;
    try { config = JSON.parse(node.getAttribute('content') || '{}'); } catch (_) { return; }
    if (!config.links || !config.routes) return;

    var state = window.fastrDocsVariantActive = window.fastrDocsVariantActive || {};
    state.config = config;
    state.normalize = function(value){
      var path = String(value || '/').split(/[?#]/, 1)[0] || '/';
      return path.length > 1 ? path.replace(/\/+$/, '') : path;
    };
    state.update = function(path){
      var family = config.routes[state.normalize(path)] || '';
      document.querySelectorAll('header nav a').forEach(function(link){
        var linkFamily = config.links[state.normalize(link.getAttribute('href'))];
        if (!linkFamily) return;
        link.setAttribute('data-fui-activelink-skip', '');
        if (family && linkFamily === family) {
          link.setAttribute('aria-current', 'page');
          link.classList.add('active');
        } else {
          link.removeAttribute('aria-current');
          link.classList.remove('active');
        }
      });
    };
    state.update(location.pathname + location.search);
    if (state.wired) return;
    state.wired = true;
    window.addEventListener('gofastr:navigate', function(event){
      var current = window.fastrDocsVariantActive;
      if (current && current.update) {
        current.update((event.detail && event.detail.path) || (location.pathname + location.search));
      }
      setTimeout(initVariantActiveLinks, 0);
    });
    if (document.body && window.MutationObserver) {
      state.observer = new MutationObserver(function(){
        var current = window.fastrDocsVariantActive;
        if (current && current.update) current.update(location.pathname + location.search);
      });
      state.observer.observe(document.body, {childList:true, subtree:true});
    }
  }
  function init(){
    if (window.__fastrDocsTocCleanup) window.__fastrDocsTocCleanup();
    if (window.__fastrDocsBlogSearchCleanup) window.__fastrDocsBlogSearchCleanup();
    if (window.__fastrDocsBlogShareCleanup) window.__fastrDocsBlogShareCleanup();
    document.querySelectorAll('[data-docs-toc-select]').forEach(function(el){ initTocSelect(el); });
    document.querySelectorAll('[data-fastr-docs-blog-search]').forEach(function(el){ initBlogSearch(el); });
    initBlogShare();
    initSidebarState();
    syncDocsDrawerTrigger();
    initPagefindSearch();
    initJSONSearch();
    initVariantSelectors();
    initVariantActiveLinks();
    if (window.__fastrDocsRuntimeReady) return;
    window.__fastrDocsRuntimeReady = true;
    var scheduleInit = function(){ setTimeout(init, 0); };
    // GoFastr's released SPA runtime emits this event after replacing the
    // page body. Keep the older event as a compatibility hook for hosts that
    // still use the pre-SPA runtime.
    window.addEventListener('gofastr:navigate', scheduleInit);
    document.addEventListener('fastr:navigate', scheduleInit);
  }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();
  window.fastrDocs = window.fastrDocs || {};
  window.fastrDocs.initTocSelect = initTocSelect;
  window.fastrDocs.initPagefindSearch = initPagefindSearch;
  window.fastrDocs.initBlogSearch = initBlogSearch;
  window.fastrDocs.initBlogShare = initBlogShare;
  window.fastrDocs.initVariantSelectors = initVariantSelectors;
})();`

// DefaultIconPNG provides one source image for GoFastr's derived PWA icons.
// Projects can replace it with their own image or use BrandConfig for the
// browser-facing logo.
func DefaultIconPNG() []byte {
	const size = 512
	canvas := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.RGBA{R: 247, G: 245, B: 239, A: 255}}, image.Point{}, draw.Src)
	ink := color.RGBA{R: 28, G: 31, B: 29, A: 255}
	orange := color.RGBA{R: 236, G: 113, B: 49, A: 255}
	// A simple rounded-ish square mark with three rising bars. The generated
	// icon is intentionally brand-neutral and remains legible at 16px.
	for y := 108; y < 404; y++ {
		for x := 108; x < 404; x++ {
			dx, dy := x-256, y-256
			if dx*dx+dy*dy < 148*148 && dx*dx+dy*dy > 128*128 {
				canvas.Set(x, y, ink)
			}
		}
	}
	for _, bar := range []struct{ x, top int }{{205, 275}, {250, 220}, {295, 165}} {
		for y := bar.top; y < 360; y++ {
			for x := bar.x; x < bar.x+25; x++ {
				canvas.Set(x, y, orange)
			}
		}
	}
	var body bytes.Buffer
	_ = png.Encode(&body, canvas)
	return body.Bytes()
}
