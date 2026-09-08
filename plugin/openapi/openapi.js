(() => {
  const init = () => {
    document.querySelectorAll('[data-openapi-reference]').forEach(root => {
      const filter = root.querySelector('[data-openapi-filter]');
      if (!filter || filter.dataset.openapiFilterBound) return;
      filter.dataset.openapiFilterBound = 'true';
      const operations = [...root.querySelectorAll('[data-openapi-operation]')];
      const groups = [...root.querySelectorAll('[data-openapi-group]')];
      // Accents fold on both sides: a reader typing "operacion" finds
      // "operación".
      const foldText = value => String(value || '').normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
      operations.forEach(operation => {
        operation.dataset.openapiSearchFolded = foldText(operation.dataset.openapiSearch);
      });
      const applyFilter = () => {
        const query = foldText(filter.value.trim());
        operations.forEach(operation => {
          operation.hidden = Boolean(query) && !operation.dataset.openapiSearchFolded.includes(query);
        });
        // A group heading with nothing left under it is noise.
        groups.forEach(group => {
          const visible = group.querySelectorAll('[data-openapi-operation]:not([hidden])');
          group.hidden = Boolean(query) && visible.length === 0;
        });
      };
      filter.addEventListener('input', applyFilter);
      // The narrowed list never survives a navigation: Escape clears it,
      // and so does arriving somewhere else.
      const resetFilter = () => {
        if (!filter.value) return;
        filter.value = '';
        operations.forEach(operation => { operation.hidden = false; });
        groups.forEach(group => { group.hidden = false; });
      };
      filter.addEventListener('keydown', event => {
        if (event.key === 'Escape' && filter.value) {
          resetFilter();
          event.stopPropagation();
        }
      });
      window.addEventListener('gofastr:navigate', resetFilter);
      window.addEventListener('fastr:navigate', resetFilter);

      const select = root.querySelector('[data-openapi-operation-select]');
      const send = root.querySelector('[data-openapi-try]');
      const response = root.querySelector('[data-openapi-response]');
      if (!select || !send || !response || send.dataset.openapiTryBound) return;
      send.dataset.openapiTryBound = 'true';

      const inputGroups = [...root.querySelectorAll('[data-openapi-inputs-for]')];
      const syncInputs = () => {
        inputGroups.forEach(group => {
          const visible = group.dataset.openapiInputsFor === select.value;
          group.hidden = !visible;
        });
      };
      select.addEventListener('change', syncInputs);
      // A deep link names an operation by its card id; landing on it
      // preselects that operation in the console instead of the first.
      if (location.hash) {
        const target = root.querySelector(location.hash);
        if (target && target.dataset.openapiOperation !== undefined) {
          const wanted = [...select.options].find(option => option.value === target.id);
          if (wanted) select.value = wanted.value;
          // The deep-linked card sits under the sticky header otherwise.
          requestAnimationFrame(() => target.scrollIntoView({ block: 'start' }));
        }
      }
      syncInputs();
      // Picking an operation is a choice worth sharing: the URL says which
      // one is loaded, without adding a history entry.
      select.addEventListener('change', () => {
        try { history.replaceState(null, '', '#' + select.value); } catch (_) {}
      });
      // A multi-server contract can be switched without a reload.
      const serverSelect = root.querySelector('[data-openapi-servers]');
      if (serverSelect) {
        serverSelect.addEventListener('change', () => {
          root.dataset.openapiServerUrl = serverSelect.value;
          const note = root.querySelector('[data-openapi-server-note]');
          if (note) note.textContent = serverSelect.value;
        });
      }

      send.addEventListener('click', async () => {
        const option = select.selectedOptions[0];
        const serverURL = root.dataset.openapiServerUrl;
        if (!option || !serverURL) {
          response.textContent = 'No server URL is configured for this reference.';
          return;
        }
        const method = option.dataset.openapiMethod || 'GET';
        const path = option.dataset.openapiPath || '/';
        const group = inputGroups.find(item => item.dataset.openapiInputsFor === option.value);
        const fields = group ? [...group.querySelectorAll('[data-openapi-param-name]')] : [];
        const valueFor = field => (field.value || '').trim();
        const missing = fields.find(field => field.dataset.openapiParamRequired === 'true' && !valueFor(field));
        if (missing) {
          response.textContent = `Enter the required ${missing.dataset.openapiParamIn || 'parameter'} “${missing.dataset.openapiParamName}”.`;
          missing.focus();
          return;
        }
        let resolvedPath = path.replace(/\{([^}]+)\}/g, (match, name) => {
          const field = fields.find(item => item.dataset.openapiParamName === name && item.dataset.openapiParamIn === 'path');
          return field ? encodeURIComponent(valueFor(field)) : match;
        });
        if (resolvedPath.includes('{')) {
          response.textContent = 'Enter values for every path parameter before sending the request.';
          return;
        }
        let url;
        try {
          const server = new URL(serverURL, window.location.href);
          const basePath = server.pathname.replace(/\/+$/, '');
          const operationPath = `/${resolvedPath.replace(/^\/+/, '')}`;
          const normalizedBase = basePath || '';
          const requestPath = normalizedBase && (operationPath === normalizedBase || operationPath.startsWith(`${normalizedBase}/`))
            ? operationPath
            : `${normalizedBase}${operationPath}`;
          url = new URL(requestPath || '/', server.origin).toString();
          const parsedURL = new URL(url);
          fields.forEach(field => {
            const value = valueFor(field);
            const location = field.dataset.openapiParamIn;
            if (!value || location === 'path') return;
            if (location === 'query') parsedURL.searchParams.set(field.dataset.openapiParamName, value);
          });
          url = parsedURL.toString();
        } catch (_) {
          response.textContent = 'The configured OpenAPI server URL is invalid.';
          return;
        }
        response.textContent = `${method} ${url}\n\nLoading…`;
        response.dataset.state = 'loading';
        response.setAttribute('aria-busy', 'true');
        root.setAttribute('aria-busy', 'true');
        send.disabled = true;
        const headers = { Accept: 'application/json' };
        const apiKeyInput = root.querySelector('[data-openapi-api-key-value]');
        if (apiKeyInput && apiKeyInput.value.trim()) {
          const headerName = root.getAttribute('data-openapi-api-key') || 'X-Api-Key';
          headers[headerName] = apiKeyInput.value.trim();
        }
        const tokenInput = root.querySelector('[data-openapi-token]');
        if (tokenInput && tokenInput.value.trim()) {
          headers['Authorization'] = `Bearer ${tokenInput.value.trim()}`;
        }
        fields.forEach(field => {
          if (field.dataset.openapiParamIn !== 'header' || !valueFor(field)) return;
          const name = field.dataset.openapiParamName;
          const forbidden = ['accept', 'content-length', 'cookie', 'host', 'origin', 'referer', 'user-agent'];
          if (!forbidden.includes(name.toLowerCase())) headers[name] = valueFor(field);
        });
        const bodyInput = group && group.querySelector('[data-openapi-body]');
        let body;
        if (bodyInput && bodyInput.value.trim()) {
          body = bodyInput.value.trim();
          const contentType = bodyInput.dataset.openapiContentType || 'application/json';
          if (contentType.toLowerCase().includes('json')) {
            try {
              body = JSON.stringify(JSON.parse(body));
            } catch (_) {
              response.textContent = 'The request body is not valid JSON.';
              bodyInput.focus();
              return;
            }
          }
          headers['Content-Type'] = contentType;
        } else if (bodyInput && bodyInput.dataset.openapiBodyRequired === 'true') {
          response.textContent = 'Enter the required request body before sending the request.';
          bodyInput.focus();
          return;
        }
        if (fields.some(field => field.dataset.openapiParamIn === 'cookie' && valueFor(field))) {
          response.textContent = `${method} ${url}\n\nCookie parameters cannot be set by a cross-origin browser request.`;
          return;
        }
        // A newer click cancels the older request rather than racing it.
        if (window.__fastrOpenAPIAbort) window.__fastrOpenAPIAbort.abort();
        const controller = window.AbortController ? new AbortController() : null;
        window.__fastrOpenAPIAbort = controller;
        try {
          try {
            const request = { method, headers, signal: controller ? controller.signal : undefined };
            if (body !== undefined) request.body = body;
            const result = await fetch(url, request);
            const responseBody = await result.text();
            // JSON renders indented, the way a person reads it.
            let displayBody = responseBody;
            try {
              displayBody = JSON.stringify(JSON.parse(responseBody), null, 2);
            } catch (_) { /* not JSON: show it verbatim */ }
            // The headers the server actually sent, because rate limits
            // and pagination live there.
            let headerLines = '';
            if (result.headers && typeof result.headers.forEach === 'function') {
              const shown = [];
              result.headers.forEach((value, name) => { shown.push(name + ': ' + value); });
              if (shown.length) headerLines = shown.join('\n') + '\n\n';
            }
            const contentType = result.headers ? (result.headers.get('content-type') || '') : '';
            response.textContent = `${method} ${url}\n\n${result.status} ${result.statusText}${contentType ? ' · ' + contentType : ''}\n${headerLines}${displayBody}`;
          } catch (error) {
            const detail = error && error.message ? ` (${error.message})` : '';
            response.textContent = `${method} ${url}\n\nRequest failed. Check the server URL, network access, and CORS policy.${detail}`;
          }
        } finally {
          // Errors land inline in the response pane rather than only the
          // console, the button comes back, and the pane stops announcing
          // itself as loading.
          delete response.dataset.state;
          response.removeAttribute('aria-busy');
          root.removeAttribute('aria-busy');
          send.disabled = false;
          if (window.__fastrOpenAPIAbort === controller) window.__fastrOpenAPIAbort = null;
        }
      });
      // A prepared request is worth taking out of the page: the console
      // can emit the same call as curl.
      const curlButton = root.querySelector('[data-openapi-curl]');
      if (curlButton && !curlButton.dataset.openapiCurlBound) {
        curlButton.dataset.openapiCurlBound = 'true';
        curlButton.addEventListener('click', () => {
          const option = select.selectedOptions[0];
          if (!option) return;
          const method = option.dataset.openapiMethod || 'GET';
          const path = option.dataset.openapiPath || '/';
          const group = inputGroups.find(item => item.dataset.openapiInputsFor === option.value);
          const fields = group ? [...group.querySelectorAll('[data-openapi-param-name]')] : [];
          const query = new URLSearchParams();
          fields.forEach(field => {
            const value = (field.value || '').trim();
            if (value && field.dataset.openapiParamIn === 'query') query.set(field.dataset.openapiParamName, value);
          });
          let resolvedPath = path;
          fields.forEach(field => {
            const value = (field.value || '').trim();
            if (field.dataset.openapiParamIn === 'path' && value) {
              resolvedPath = resolvedPath.split(`{${field.dataset.openapiParamName}}`).join(encodeURIComponent(value));
            }
          });
          const target = `${root.dataset.openapiServerUrl || ''}${resolvedPath}${query.toString() ? '?' + query : ''}`;
          const lines = [`curl -X ${method} '${target}'`];
          if (headers && headers.Accept) lines.push(`  -H 'Accept: ${headers.Accept}'`);
          const tokenInput = root.querySelector('[data-openapi-token]');
          if (tokenInput && tokenInput.value.trim()) lines.push(`  -H 'Authorization: Bearer ${tokenInput.value.trim()}'`);
          const apiKeyInput = root.querySelector('[data-openapi-api-key-value]');
          if (apiKeyInput && apiKeyInput.value.trim()) {
            const apiKeyHeader = root.getAttribute('data-openapi-api-key') || 'X-Api-Key';
            lines.push(`  -H '${apiKeyHeader}: ${apiKeyInput.value.trim()}'`);
          }
          const bodyInput = group && group.querySelector('[data-openapi-body]');
          if (bodyInput && bodyInput.value.trim()) lines.push(`  -d '${bodyInput.value.trim().replace(/'/g, `'\\''`)}'`);
          const command = lines.join(' \\\n');
          const copied = () => { curlButton.textContent = 'Copied'; setTimeout(() => { curlButton.textContent = 'Copy as cURL'; }, 1600); };
          if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
            navigator.clipboard.writeText(command).then(copied).catch(copied);
          } else { copied(); }
          response.textContent = command;
        });
      }
    });
  };

  // GoFastr's AnchoredRail owns the API index scrollspy. This plugin only
  // supplies the OpenAPI-specific endpoint filter.
  init();
  window.addEventListener('gofastr:navigate', init);
})();
