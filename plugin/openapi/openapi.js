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
      send.addEventListener('click', async () => {
        const option = select.selectedOptions[0];
        const serverURL = root.dataset.openapiServerUrl;
        if (!option || !serverURL) {
          response.textContent = 'No server URL is configured for this reference.';
          return;
        }
        const method = option.dataset.openapiMethod || 'GET';
        const path = option.dataset.openapiPath || '/';
        if (path.includes('{')) {
          response.textContent = 'This operation has path variables. Fill them through a project-specific client before sending.';
          return;
        }
        let url;
        try {
          const server = new URL(serverURL, window.location.href);
          const basePath = server.pathname.replace(/\/+$/, '');
          const operationPath = `/${path.replace(/^\/+/, '')}`;
          const normalizedBase = basePath || '';
          const requestPath = normalizedBase && (operationPath === normalizedBase || operationPath.startsWith(`${normalizedBase}/`))
            ? operationPath
            : `${normalizedBase}${operationPath}`;
          url = new URL(requestPath || '/', server.origin).toString();
        } catch (_) {
          response.textContent = 'The configured OpenAPI server URL is invalid.';
          return;
        }
        response.textContent = `${method} ${url}\n\nLoading…`;
        try {
          const result = await fetch(url, { method, headers: { Accept: 'application/json' } });
          const body = await result.text();
          response.textContent = `${method} ${url}\n\n${result.status} ${result.statusText}\n${body}`;
        } catch (_) {
          response.textContent = `${method} ${url}\n\nRequest failed. Check the server URL, network access, and CORS policy.`;
        }
      });
    });
  };

  // GoFastr's AnchoredRail owns the API index scrollspy. This plugin only
  // supplies the OpenAPI-specific endpoint filter.
  init();
  window.addEventListener('gofastr:navigate', init);
})();
