(() => {
  const init = () => {
    document.querySelectorAll('[data-openapi-reference]').forEach(root => {
      const filter = root.querySelector('[data-openapi-filter]');
      if (!filter || filter.dataset.openapiFilterBound) return;
      filter.dataset.openapiFilterBound = 'true';
      const operations = [...root.querySelectorAll('[data-openapi-operation]')];
      filter.addEventListener('input', () => {
        const query = filter.value.trim().toLowerCase();
        operations.forEach(operation => {
          operation.hidden = Boolean(query) && !operation.dataset.openapiSearch.includes(query);
        });
      });

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
      syncInputs();

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
        const headers = { Accept: 'application/json' };
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
        try {
          const request = { method, headers };
          if (body !== undefined) request.body = body;
          const result = await fetch(url, request);
          const responseBody = await result.text();
          response.textContent = `${method} ${url}\n\n${result.status} ${result.statusText}\n${responseBody}`;
        } catch (error) {
          const detail = error && error.message ? ` (${error.message})` : '';
          response.textContent = `${method} ${url}\n\nRequest failed. Check the server URL, network access, and CORS policy.${detail}`;
        }
      });
    });
  };

  // GoFastr's AnchoredRail owns the API index scrollspy. This plugin only
  // supplies the OpenAPI-specific endpoint filter.
  init();
  window.addEventListener('gofastr:navigate', init);
})();
