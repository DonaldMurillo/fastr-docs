---
tags: [testing, e2e, validation]
---

# Testing

The product is browser behavior, not only successful compilation. Test the Router model in Go, then exercise a generated or hand-authored project through a real browser.

## Fast checks

```sh
go test ./...
go vet ./...
fastr-docs doctor .
```

Strict validation catches missing titles, unusable content, unsafe or ambiguous paths, duplicate sibling order, broken route registrations, and invalid plugin contributions.

## Browser checks

```sh
cd e2e
npm install
npx playwright install chromium
npm test
```

The suite covers desktop and mobile navigation, active nested routes, command-palette search, theme persistence, in-page navigation, OpenAPI requests, typed GoFastr controls, static export, PWA assets, and offline behavior.

## Test both project paths

The suite creates a project with the CLI and also builds the hand-authored non-CLI fixture. Keep both flows covered when changing the public Router API. For visual changes, review the desktop and mobile snapshots instead of accepting a baseline blindly.
