---
tags: [matematicas, katex, plugins]
locale: es
---

# Matemáticas

El TeX se renderiza en la propia página. Sin marco, sin política de contenido
relajada, y las fórmulas en línea se apoyan en la línea base del texto, que es
donde tienen que estar.

```go title="docs/router.go"
router.Use(katex.Plugin{})
```

Esa es toda la configuración. El plugin aporta su runtime, su hoja de estilos y
sus tipografías, así que nada de `main.go` cambia.

## Cómo se escribe

Encierra las matemáticas en línea entre dólares simples y las de bloque entre
dólares dobles, tal como espera cualquier editor que conozca TeX.

```md
El factor de Lorentz es $\gamma = 1/\sqrt{1 - v^2/c^2}$.

$$
\int_{-\infty}^{\infty} e^{-x^2}\,dx = \sqrt{\pi}
$$
```

Que se ve así: el factor de Lorentz es $\gamma = 1/\sqrt{1 - v^2/c^2}$.

$$
\int_{-\infty}^{\infty} e^{-x^2}\,dx = \sqrt{\pi}
$$

Matrices, sumatorios y alineaciones funcionan, porque esto es KaTeX de verdad y
no un subconjunto:

$$
\sum_{i=1}^{n} i^2 = \frac{n(n+1)(2n+1)}{6}
$$

Si los dólares estorban, el shortcode `math` hace matemáticas de bloque sin
delimitadores:

{{< math >}}
e^{i\pi} + 1 = 0
{{< /math >}}

## Dólares que no son matemáticas

La prosa está llena de signos de dólar que no tienen nada que ver con TeX, y un
escáner que se apropie de uno borra el texto de quien lee. Este sigue las reglas
que fijó Pandoc, así que todo esto se queda tal cual:

| Escribes | Por qué sigue siendo prosa |
| --- | --- |
| `Los planes van de $5 a $10.` | Tras el dólar de cierre hay un dígito |
| `Cuesta $ 5.` | Tras el dólar de apertura hay un espacio |
| `Paga $5 o $ 6.` | Antes del dólar de cierre hay un espacio |
| `Define $HOME y $PATH.` | La misma regla, en el dólar de cierre |
| `Un literal \$5 se queda.` | La barra invertida lo escapa |

El código no se escanea nunca, ni las vallas ni los tramos en línea, así que
`echo $x$` en un ejemplo de shell está a salvo.

Para un sitio cuya prosa sean sobre todo precios, apaga el escáner y quédate con
el shortcode:

```go
router.Use(katex.Plugin{DisableDollarSyntax: true})
```

## Por qué este no está aislado

El [plugin de Mermaid](/es/docs/build/diagrams) renderiza dentro de un iframe de
origen opaco, porque Mermaid inyecta un elemento `<style>` y emite SVG con
estilos en línea, y `default-src 'self'` bloquea ambas cosas.

KaTeX no lo necesita. Construye su composición como nodos del DOM y asigna los
tamaños por CSSOM, como `node.style.height = "0.68em"`, y la CSP vigila los
atributos de estilo del marcado y los elementos `<style>`, no el CSSOM. Así que
la política estricta se mantiene tal cual, sin ninguna excepción para las
matemáticas.

El único punto donde KaTeX rompería la política es su propia vía de error, que
escribe un atributo `style` para pintar el mensaje en rojo. El runtime activa
`throwOnError` y dibuja su propio texto de error, así que una fórmula mal
escrita te enseña el error de análisis en vez de llenar la consola de
violaciones.

Renderizar en la página es además lo que hace posible las matemáticas en línea.
Un marco es un rectángulo; no puede compartir línea con la frase que lo rodea,
así que $a^2 + b^2 = c^2$ dentro de un párrafo tiene que ser un elemento de
verdad.

## Qué descarga una página

El renderizador ocupa 266 KB, y la mayoría de las páginas no tienen
matemáticas. Por eso lo único que lleva cada página es un cargador de 783 bytes,
que busca una fórmula y solo entonces pide el renderizador y la hoja de estilos.

| Página | Bytes de KaTeX |
| --- | --- |
| Sin matemáticas | 783 |
| Esta | unos 347 KB, con 4 tipografías |

El cargador sigue vigilando después de la primera pasada, así que llegar aquí
desde una página sin matemáticas también funciona.

Las tipografías son perezosas por su cuenta: la hoja de estilos declara veinte y
el navegador solo pide las que usa una fórmula. Nada viene de una CDN, así que
una exportación sin conexión dibuja las matemáticas igual que el sitio en vivo.
