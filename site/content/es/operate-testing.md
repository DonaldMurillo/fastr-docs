---
tags: [pruebas, e2e, validacion]
locale: es
---

# Pruebas

El producto es el comportamiento en el navegador, no solo que
compile. Prueba el modelo del Router en Go y luego recorre un proyecto generado
o escrito a mano en un navegador de verdad.

## Comprobaciones rápidas

```sh
go test ./...
go vet ./...
fastr-docs doctor .
```

La validación estricta detecta títulos que faltan, contenido inservible, rutas
inseguras o ambiguas, orden duplicado entre hermanas, registros de ruta rotos y
aportaciones de plugin inválidas.

## Comprobaciones en el navegador

```sh
cd e2e
npm install
npx playwright install chromium
npm test
```

La suite cubre la navegación de escritorio y móvil, las rutas anidadas activas,
la búsqueda por la paleta de comandos, la persistencia del tema, la navegación
interna, las peticiones OpenAPI, los controles tipados de GoFastr, la
exportación estática, los recursos PWA y el comportamiento sin conexión.

## Prueba los dos caminos

La suite crea un proyecto con la CLI y además construye el fixture escrito a
mano. Mantén cubiertos los dos flujos cuando cambies la API pública del
Router. Para cambios visuales, revisa las capturas de escritorio y de móvil en
lugar de aceptar la nueva referencia a ciegas.
