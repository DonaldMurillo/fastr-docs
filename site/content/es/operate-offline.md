---
tags: [offline, pwa, exportacion-estatica]
locale: es
---

# Sin conexión y PWA

La documentación estática debería seguir leyéndose cuando la red desaparece.
Marca `Offline: true` las rutas que deben durar y exporta el sitio para que
GoFastr precachee la envoltura y fastr-docs incluya los recursos de las rutas.

## Exportar el sitio

```sh
fastr-docs export . --out dist --pagefind
```

La salida incluye el HTML de las rutas, el manifiesto, el service worker, los
iconos de instalación generados, el runtime de la documentación, los registros
de búsqueda, el runtime de OpenAPI, los recursos para agentes y los archivos
públicos del proyecto.

## Instalar y actualizar

El host en vivo se puede instalar en `localhost` y en despliegues HTTPS porque
el layout generado activa el soporte PWA de GoFastr y deriva los iconos
necesarios. Una instalación en vivo mantiene disponibles la envoltura de la
aplicación y la pantalla sin conexión, mientras que las navegaciones a
documentos siguen yendo a la red primero, para que el contenido publicado nunca
esté rancio por diseño.

Para una aplicación de documentación totalmente sin conexión, instala la
exportación estática. Despliega `dist` en un host estático o en una CDN. El
worker exportado precachea todas las rutas exportadas y su huella de contenido
cambia cuando cambia la exportación. En la visita siguiente el navegador instala
el nuevo worker y la nueva caché en segundo plano. GoFastr espera a que se
cierren las pestañas viejas antes de activarlo, así que una sesión de lectura
abierta no se interrumpe.

Mantén alineadas la dependencia del framework y su CLI al actualizar:

```sh
fastr-docs upgrade .
fastr-docs upgrade . --apply
```

El primer comando muestra las migraciones entre la versión de `go.mod` y la
versión objetivo. El segundo aplica los pasos de dependencia, tidy, build y
pruebas. El binario `gofastr` y el módulo de Go se versionan por separado, así
que instala la CLI objetivo antes de aplicar una actualización de varias
versiones.

## Entender el límite

Las páginas de Markdown, la navegación, la búsqueda local y la UI exportada
pueden funcionar sin conexión. Las peticiones interactivas de la API siguen
necesitando red y un servidor alcanzable. La consola de OpenAPI avisa de esa
limitación cuando no hay URL de servidor configurada.

## Mantener la caché a propósito

El host generado precachea solo el runtime y los recursos de ruta que conoce. Si
un plugin es dueño de recursos de navegador, móntalos a través del Router para
que el host y la exportación puedan contarlos.

Prueba el sitio exportado en un navegador real con la red desactivada. La suite
E2E de este proyecto cubre la navegación, la búsqueda con Pagefind, los recursos
PWA y el renderizado de rutas sin conexión.
