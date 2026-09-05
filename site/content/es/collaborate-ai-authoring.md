---
tags: [agentes, escritura-ia, flujo]
locale: es
---

# Escribir con agentes

Un agente no debería tener que deducir por ingeniería inversa cómo está montado
un proyecto de documentación. El generador incluye referencias a nivel de
proyecto y una habilidad de escritura reutilizable que apuntan al Router, al
árbol de contenido, a los comandos de validación y a las comprobaciones de
navegador.

## Deja clara la fuente de la verdad

- `docs/router.go` gobierna el registro y el orden de las rutas.
- `content/` gobierna el Markdown y el front matter.
- `openapi.json` gobierna el contrato de la API que consume el plugin.
- `agents/claude.md` explica las convenciones del proyecto.
- `.agents/skills/docs-authoring/SKILL.md` describe las tareas seguras.

## Dale al agente un endpoint vivo

El host generado también sirve el endpoint `/mcp` de GoFastr y lo anuncia en la
tarjeta de agente y en `/.well-known/mcp.json`. `framework.WithMCP()` monta el
transporte. `framework.WithMCPIntrospection()` añade herramientas de solo
lectura para inspeccionar rutas, plugins, configuración y estado.

El endpoint pertenece a un servidor en marcha. La exportación estática incluye
las páginas y los archivos de descubrimiento, pero no puede incluir un
transporte MCP vivo. Quita las dos opciones del framework y
`AgentCard.MCPEndpoint` a la vez si un despliegue no debe exponer MCP.

## Cambia las cosas juntas

Al añadir una página, actualiza el registro de la ruta y su Markdown en el mismo
cambio. Dale título, descripción, orden explícito y un distintivo solo cuando el
matiz vaya a seguir siendo útil. Deja el contenido en borrador fuera del build
público y pasa las comprobaciones estrictas antes de entregar.

{{< note title="Traducir es lo mismo" >}}
Una traducción también son dos cambios en uno: el archivo en `content/es/` y su
ruta en `docs/router.go`. Un agente que añade solo el archivo deja una página
que no existe.
{{< /note >}}

## Verifica el resultado

Los agentes deberían ejecutar `fastr-docs doctor .`, las pruebas del proyecto y
la suite de navegador cuando cambie el comportamiento. El árbol de rutas es un
contrato: una página no está terminada cuando existe su archivo, sino cuando
funcionan su navegación, su búsqueda, su exportación y su recorrido.
