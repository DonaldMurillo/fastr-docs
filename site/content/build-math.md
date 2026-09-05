---
tags: [math, katex, plugins]
locale: en
---

# Math

TeX renders in the page itself. No iframe, no relaxed content policy, and inline
formulas sit on the text baseline where they belong.

```go title="docs/router.go"
router.Use(katex.Plugin{})
```

That is the whole setup. The plugin contributes its runtime, its stylesheet, and
its fonts, so nothing in `main.go` changes.

## Writing it

Wrap inline math in single dollars and display math in double dollars, the way
every TeX-aware editor already expects.

```md
The Lorentz factor is $\gamma = 1/\sqrt{1 - v^2/c^2}$.

$$
\int_{-\infty}^{\infty} e^{-x^2}\,dx = \sqrt{\pi}
$$
```

Which renders as: the Lorentz factor is $\gamma = 1/\sqrt{1 - v^2/c^2}$.

$$
\int_{-\infty}^{\infty} e^{-x^2}\,dx = \sqrt{\pi}
$$

Matrices, sums, and alignment all work, because this is the real KaTeX, not a
subset:

$$
\begin{pmatrix} a & b \\ c & d \end{pmatrix}
\begin{pmatrix} x \\ y \end{pmatrix}
=
\begin{pmatrix} ax + by \\ cx + dy \end{pmatrix}
$$

$$
\sum_{i=1}^{n} i^2 = \frac{n(n+1)(2n+1)}{6}
$$

If dollars are awkward, the `math` shortcode does display math with no
delimiters at all:

```md
{{< math >}}
e^{i\pi} + 1 = 0
{{< /math >}}
```

{{< math >}}
e^{i\pi} + 1 = 0
{{< /math >}}

## Dollars that are not math

Prose is full of dollar signs that have nothing to do with TeX, and a scanner
that claims one deletes the reader's text. This one follows the rules Pandoc
settled on, so all of these stay exactly as written:

| You write | Why it stays prose |
| --- | --- |
| `Plans run from $5 to $10.` | The closing dollar is followed by a digit |
| `It costs $ 5.` | The opening dollar is followed by a space |
| `Pay $5 or $ 6.` | The closing dollar is preceded by a space |
| `Set $HOME and $PATH.` | Same rule, on the closing dollar |
| `A literal \$5 stays.` | A backslash escapes it outright |

Code is never scanned at all, neither fenced blocks nor inline spans, so
`echo $x$` in a shell example is safe.

For a site whose prose is mostly prices, turn the scanner off and keep the
shortcode:

```go
router.Use(katex.Plugin{DisableDollarSyntax: true})
```

## Why this one is not sandboxed

The [Mermaid plugin](/docs/build/diagrams) renders inside an opaque-origin
iframe, because Mermaid injects a `<style>` element and emits SVG carrying
inline styles, both of which `default-src 'self'` blocks.

KaTeX does not need that. It builds its layout as DOM nodes and assigns sizes
through CSSOM, as `node.style.height = "0.68em"`, and CSP polices style
attributes in parsed markup and `<style>` elements, not CSSOM. So the strict
policy holds as written, with no exception carved out for math.

The one place KaTeX would trip the policy is its own error path, which writes a
`style` attribute to color the message red. The runtime turns `throwOnError` on
and renders its own error text instead, so a mistyped formula shows you the
parse error rather than a console full of policy violations.

Rendering in the page rather than in a frame is also what makes inline math
work. A frame is a rectangle; it cannot share a line box with the sentence
around it, so $a^2 + b^2 = c^2$ inside a paragraph has to be a real element.

## What a page actually downloads

The renderer is 266 KB, and most documentation pages have no math on them. So
the only file every page carries is a 783-byte loader. It looks for a formula,
and fetches the renderer and stylesheet only when it finds one.

| Page | KaTeX bytes |
| --- | --- |
| No math on it | 783 |
| This one | ~347 KB, including 4 font faces |

The loader keeps watching after the first pass, so arriving here from a page
with no math still works: the docs shell swaps pages without a reload.

Fonts are lazy in their own right. The stylesheet declares twenty faces and the
browser fetches only the ones a formula uses, each cached for a year. Nothing
comes from a CDN, so an offline export renders math the same as the live site.
