# AI authoring

The generated project includes `agents/claude.md` and `.agents/skills/docs-authoring/SKILL.md` so an agent can extend the site without guessing where the source of truth lives.

## Give agents the same model

The authoring guidance says to:

1. Register routes in `docs/router.go`.
2. Keep Markdown in `content/`.
3. Assign explicit sibling order values.
4. Use screens only when the interaction needs typed state or actions.
5. Run the project checks and browser suite before handoff.

This is not a second product workflow. It is a compact, machine-readable explanation of the human workflow already enforced by the project.

## Use the live agent endpoint

The generated host serves GoFastr's `/mcp` endpoint and advertises it through
the agent card and `/.well-known/mcp.json`. `framework.WithMCP()` mounts the
transport. `framework.WithMCPIntrospection()` adds read-only tools for
inspecting routes, plugins, configuration, and readiness.

Static exports include the docs and discovery files, but they cannot serve a
live MCP transport. Remove the two framework options and
`AgentCard.MCPEndpoint` together if a deployment should not expose MCP.

## Safe agent changes

Ask an agent to add a page, update a route description, or extend an existing plugin with a concrete path and acceptance behavior. Require it to report changed routes, tests run, and any visual surfaces inspected.

Keep strict validation enabled. If an advanced integration needs an opt-out, document the reason next to the configuration and add a test for the new invariant.
