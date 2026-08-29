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
// responsive in-page selector and the active anchor state.
func RuntimeJS() string { return docsRuntimeJS }

const docsRuntimeJS = `(function(){
  function initTocSelect(root){
    var select = root.tagName === 'SELECT' ? root : root.querySelector('select,[data-docs-toc-select]') || root;
    if (!select || select.dataset.docsTocReady === 'true') return;
    select.dataset.docsTocReady = 'true';
    select.addEventListener('change', function(){
      if (!select.value) return;
      var target = document.querySelector(select.value);
      if (target) window.history.replaceState(null, '', select.value);
      if (target) target.scrollIntoView({behavior:'smooth', block:'start'});
    });
  }
  function normalizeDocsPath(value){
    var path = String(value || '/').split(/[?#]/, 1)[0] || '/';
    return path.length > 1 ? path.replace(/\/+$/, '') : path;
  }
  function syncMobileSidebar(){
    var drawer = document.querySelector('[data-fui-widget="fastr-docs-sections"]');
    if (!drawer) return;
    var currentPath = normalizeDocsPath(location.pathname);
    var activeLink = null;
    drawer.querySelectorAll('a.ui-sidebar__link').forEach(function(link){
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
    drawer.querySelectorAll('details.ui-sidebar__group').forEach(function(group){
      var active = !!activeLink && group.contains(activeLink);
      if (!active) return;
      group.open = true;
      var summary = group.querySelector('summary');
      if (summary) summary.setAttribute('aria-expanded', 'true');
    });
  }
  function initMobileSidebar(){
    syncMobileSidebar();
    if (window.fastrDocsMobileSidebarReady || !document.body || !window.MutationObserver) return;
    window.fastrDocsMobileSidebarReady = true;
    var observer = new MutationObserver(function(records){
      for (var i = 0; i < records.length; i++) {
        var record = records[i];
        if (record.type === 'childList' || record.attributeName === 'hidden') {
          syncMobileSidebar();
          return;
        }
      }
    });
    observer.observe(document.body, {childList:true, subtree:true, attributes:true, attributeFilter:['hidden']});
    window.addEventListener('gofastr:navigate', function(){ setTimeout(syncMobileSidebar, 0); });
    window.addEventListener('fastr:navigate', function(){ setTimeout(syncMobileSidebar, 0); });
  }
  function escapeHTML(value){
    return String(value == null ? '' : value).replace(/[&<>'"]/g, function(char){
      return {'&':'&amp;','<':'&lt;','>':'&gt;',"'":'&#39;','"':'&quot;'}[char];
    });
  }
  function pagefindTrigger(){
    return document.querySelector('[data-fastr-docs-backend="pagefind"]');
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
  function restorePalette(list){
    if (!list || !list.dataset.fastrDocsInitialOptions) return;
    list.innerHTML = list.dataset.fastrDocsInitialOptions;
    list.setAttribute('data-fui-static-options', '');
    list.removeAttribute('hidden');
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
      var excerpt = data.excerpt || data.url || '';
      var url = data.url || '#';
      return '<li role="option" id="fastr-docs-command-palette-list-opt-pagefind-' + index + '" data-value="' + escapeHTML(title) + '" data-fui-push-state="' + escapeHTML(url) + '">' +
        '<span class="combobox__opt-label">' + escapeHTML(title) + '</span>' +
        '<span class="combobox__opt-meta">' + escapeHTML(excerpt) + '</span></li>';
    }).join('');
    list.removeAttribute('data-fui-static-options');
    list.removeAttribute('hidden');
  }
  function initPagefindSearch(){
    var trigger = pagefindTrigger();
    if (!trigger || trigger.dataset.fastrDocsPagefindReady === 'true') return;
    trigger.dataset.fastrDocsPagefindReady = 'true';
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
    document.querySelectorAll('[data-docs-toc-select]').forEach(function(el){ initTocSelect(el); });
    initMobileSidebar();
    initPagefindSearch();
    initVariantSelectors();
    initVariantActiveLinks();
    document.addEventListener('fastr:navigate', function(){ setTimeout(init, 0); });
  }
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
  else init();
  window.fastrDocs = window.fastrDocs || {};
  window.fastrDocs.initTocSelect = initTocSelect;
  window.fastrDocs.initPagefindSearch = initPagefindSearch;
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
