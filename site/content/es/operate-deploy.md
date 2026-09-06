---
tags: [despliegue, marca, hosting]
locale: es
---

# Desplegar y personalizar

El sitio generado es de marca blanca. Conserva el Router y el modelo de host, y
sustituye la marca, los recursos públicos, el origen de ejecución y la ruta de
despliegue por los del proyecto.

## Poner la marca

```go
router := docs.NewRouter(
    docs.WithSiteName("Acme Docs"),
    docs.WithBrand(docs.BrandConfig{
        Name: "Acme Docs", FaviconURL: "/assets/favicon.svg",
    }),
)
```

Usa `WithCustomCSS` para los cambios visuales propios del producto y
`MountAssets` para archivos públicos seguros. El tema por defecto expone tokens
claros y oscuros para la envoltura y para los componentes de GoFastr.

## Configurar los orígenes de despliegue

Define `PUBLIC_SITE_URL` para que el sitemap y el robots lleven URLs absolutas.
Define `API_SERVER_URL` cuando el servidor OpenAPI cambie entre local, staging y
producción. Los orígenes de API declarados entran en la CSP del host; los
destinos de red arbitrarios no.

## Desplegar bajo una subruta

```sh
fastr-docs export . --out dist --base /docs
```

La exportación reescribe las URLs del runtime, de la búsqueda y de OpenAPI para
la base elegida, y graba la base en cada página para que el selector de idioma,
los resultados de búsqueda y el estado activo de la barra lateral funcionen
también debajo de ella. Sirve el resultado como sitio estático o ponlo detrás
del host de GoFastr.

## GitHub Pages

Un repositorio llamado `<usuario>.github.io` se sirve en la raíz de ese
dominio. Cualquier otro es una página de proyecto bajo
`https://<usuario>.github.io/<repo>/`, que es el caso de la subruta de arriba.
El flujo de `.github/workflows/pages.yml` cubre los dos: deduce la base del
nombre del repositorio, exporta, pasa Pagefind sobre el resultado y publica con
`actions/deploy-pages`. Activa Pages en el repositorio con GitHub Actions como
origen y haz push. Un dominio propio deja la base vacía otra vez; no cambia
nada más.

{{< tip title="El idioma no cambia el despliegue" >}}
Un sitio traducido se exporta igual que uno monolingüe: las páginas en español
son rutas del mismo árbol, así que `export` las escribe junto a las demás.
{{< /tip >}}
