# Custom surfaces

When a product feature needs its own documentation surface, start with the route interface and work outward.

## Surface checklist

1. **Purpose** — what should the reader accomplish here?
2. **Route** — where does it belong in the top-level tree?
3. **Rendering mode** — Markdown page, typed screen, or plugin?
4. **Navigation** — which group owns it and what comes before it?
5. **Search** — what title, description, tags, and terms make it discoverable?
6. **Responsive shape** — what happens at wide, tablet, and phone widths?
7. **Offline behavior** — can it render without a network, and which actions require one?
8. **E2E flow** — what does a user click, type, select, submit, or navigate?

## A useful default

Start with a Markdown explanation and link to a focused interactive screen only when the interaction adds real value. This keeps the durable story readable while giving the product room to demonstrate behavior.

## Ship it as one change

Register the route, add the content, update agent guidance if needed, add E2E coverage, and inspect the rendered page at all supported breakpoints before calling the surface complete.
