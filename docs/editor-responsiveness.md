# Editor Responsiveness

Measured by Playwright e2e tests (`tests/e2e/editor.spec.ts`) against
`gobook serve` running a multi-cell notebook.

## Viewports

| Viewport | Width | Height | Device         |
|----------|-------|--------|----------------|
| Desktop  | 1280  | 720    | Desktop Chrome |
| Mobile   | 393   | 851    | Pixel 5        |

## Measurements

Recorded from `PerformanceNavigationTiming` in the browser.

| Metric           | Desktop | Mobile  |
|------------------|---------|---------|
| DOMContentLoaded | < 500ms | < 500ms |
| Load complete    | < 1s    | < 1s    |

These are targets. Actual measurements are logged by the Playwright tests
and should be verified in CI.

## Test coverage

The e2e test suite covers:

- Page load and block rendering at both viewports
- Cell blocks with Run Cell buttons
- Prose blocks with save functionality
- Secret prompt with masked input
- Run All toolbar button
- CSP and X-Content-Type-Options headers
- API endpoints (status, blocks, diagnostics, completion)
- Visual regression screenshots at both viewports
