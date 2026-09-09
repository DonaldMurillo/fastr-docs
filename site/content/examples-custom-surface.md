---
tags: [examples, screens, product-surfaces]
---

# Custom surface checklist

A custom docs surface is still a route. Start with the user task, choose the rendering strategy, and keep the route connected to the project’s shared navigation and search model.

## Decide the route interface

1. Name the task in the route title and description.
2. Choose a Markdown Page for durable explanation or a typed Screen for interaction.
3. Assign an explicit sibling order.
4. Decide whether the surface works offline.
5. Add tags and search text when the visible copy is not enough.

## Compose with GoFastr

Use the framework’s controls for forms, tabs, disclosures, data views, notifications, and responsive behavior. Keep state and server actions in the screen component. Let the docs shell continue to own the header, drawer, command palette, in-page navigation, and theme.

## Verify the user flow

Test the route at desktop and mobile widths. Cover the first render, the main interaction, error or empty states, navigation away and back, keyboard access, and static export. The [playground](/examples/playground) shows this pattern in a small screen.
