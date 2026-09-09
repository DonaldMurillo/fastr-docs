---
tags: [agents, ai-authoring, workflow]
---

# AI authoring

AI agents should not have to reverse-engineer a documentation project. The generator includes project-level references and a reusable authoring skill that point agents to the Router, content tree, validation commands, and browser checks.

## Keep the source of truth clear

- `docs/router.go` owns route registration and ordering.
- `content/` owns Markdown and front matter.
- `openapi.json` owns the API spec consumed by the plugin.
- `agents/claude.md` explains project conventions.
- `.agents/skills/docs-authoring/SKILL.md` describes safe authoring tasks.

## Give agents a live endpoint

The generated host also serves GoFastr's `/mcp` endpoint and advertises it in
the agent card and `/.well-known/mcp.json`. `framework.WithMCP()` mounts the
transport. `framework.WithMCPIntrospection()` adds read-only tools for
inspecting routes, plugins, configuration, and readiness.

The endpoint belongs to a running server. Static export includes the pages and
agent discovery files, but it cannot include a live MCP transport. Remove the
two framework options and `AgentCard.MCPEndpoint` together if a deployment
should not expose MCP.

## Make changes together

When adding a page, update the route registration and its Markdown source in the same change. Give it a title, description, explicit order, and a badge only when the qualifier will remain useful. Keep draft content out of the public build and run strict checks before handoff.

## Verify the result

Agents should run `fastr-docs doctor .`, the project tests, and the browser suite for behavior changes. The route tree is a promise: a page is not complete when its file exists; it is complete when navigation, search, export, and its user flow all work.
