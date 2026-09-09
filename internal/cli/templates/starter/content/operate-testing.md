# Testing

The product is browser behavior, not just successful compilation. Test the route tree in Go, then exercise the generated project and non-CLI project through a real browser.

## The verification layers

| Layer | Proves |
| --- | --- |
| Go tests | Registration, validation, search records, plugin wiring |
| Generated-project E2E | The CLI output behaves as a complete app |
| Non-CLI E2E | Hand-authored GoFastr integrations work without the generator |
| Static E2E | Exported routes and PWA assets survive deployment |
| Visual inspection | Geometry, hierarchy, spacing, overflow, and responsive composition |

## Write behavior-focused tests

Assert what a reader or operator experiences: a route opens, a parent disclosure is open, a command palette result navigates, an API request shows its response, a drawer closes after navigation, or a mobile document has no horizontal overflow.

Avoid tests that only inspect implementation markers. A marker can be present while the page is clipped, inaccessible, or visually unusable.

## Run the suite

```sh
go test ./...
cd e2e
npm test
```

When a breakpoint changes, add or update a screenshot inspection in addition to the behavioral assertion. The browser should be treated as the final consumer of the product.
