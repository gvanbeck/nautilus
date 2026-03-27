# Nautilus HTML — Developer Guide

The `pdf/html` package converts inline HTML markup and HTML tables into
structures that can be rendered by the Nautilus PDF library.  There are two
independent features:

1. **Inline HTML span parsing** (`html.Parse`) — converts a string containing
   HTML tags into a slice of styled `Span` values for rendering in-line text.
2. **HTML table parsing** (`html.ParseTable`) — converts a full `<table>`
   element into an `HtmlTable` that the document can render via
   `doc.TableFromHTML`.

---

## Table of Contents

1. [Inline HTML — supported tags](#1-inline-html--supported-tags)
2. [Parsing inline HTML](#2-parsing-inline-html)
3. [CSS class mapping](#3-css-class-mapping)
4. [Rendering inline spans with WriteHTMLSpans](#4-rendering-inline-spans-with-writehtmlspans)
5. [HTML table parsing](#5-html-table-parsing)
6. [Rendering HTML tables with TableFromHTML](#6-rendering-html-tables-with-tablefromhtml)
7. [Full example — inline HTML](#7-full-example--inline-html)
8. [Full example — HTML table](#8-full-example--html-table)
9. [Known limitations](#9-known-limitations)

---

## 1. Inline HTML — supported tags

The inline parser recognises the following HTML tags.  Tags may be freely
nested.  All other tags are ignored (their content is still preserved as plain
text).

| Tag(s)                              | Effect          |
|-------------------------------------|-----------------|
| `<b>`, `<strong>`                   | Bold            |
| `<i>`, `<em>`, `<cite>`, `<var>`, `<dfn>` | Italic   |
| `<u>`, `<ins>`                      | Underline       |
| `<s>`, `<strike>`, `<del>`          | Strikethrough   |
| `<code>`, `<tt>`, `<kbd>`, `<samp>` | Monospace       |
| any tag with `class="…"`            | CSS class       |

Self-closing tags (e.g. `<br/>`) are accepted but produce no output; use
`\n` or `<br>` converted to `\n` at the call site for line breaks.

---

## 2. Parsing inline HTML

```go
import "github.com/gvanbeck/nautilus/pdf/html"

spans, err := html.Parse(
    `Hello <b>world</b> and <i>italic <b>bold-italic</b></i> text.`,
    nil, // no CSS class mapping
)
```

`Parse` returns a `[]html.Span` where each span holds:

```go
type Span struct {
    Text  string     // plain text content of this span
    Style Style      // accumulated formatting
    Class string     // innermost CSS class name, if any
}

type Style struct {
    Bold          bool
    Italic        bool
    Underline     bool
    Strikethrough bool
    Monospace     bool
}
```

Consecutive characters with the same style are grouped into a single span.
The parser handles unclosed or misordered tags gracefully.

---

## 3. CSS class mapping

Pass a `ClassStyle` map to `Parse` to translate CSS class names into styling
flags. The class name is always preserved in `Span.Class` regardless.

```go
classes := html.ClassStyle{
    "highlight": html.Style{Bold: true},
    "warning":   html.Style{Italic: true, Underline: true},
    "code":      html.Style{Monospace: true},
}

spans, err := html.Parse(
    `Normal <span class="highlight">important</span> text.`,
    classes,
)
// spans[1].Style.Bold == true
// spans[1].Class     == "highlight"
```

If no class mapping is provided (or a class name is not in the map), the class
name is still recorded in `Span.Class` so the caller can apply custom rendering
logic.

---

## 4. Rendering inline spans with WriteHTMLSpans

`doc.WriteHTMLSpans` renders a `[]html.Span` inline on the page.  You supply
a callback (`fontFor`) that maps a `Style` to a registered font name, allowing
you to switch between your regular, bold, italic, and monospace fonts.

```go
import (
    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/html"
)

// Register four font variants.
doc.RegisterFont("regular",    "/path/to/Roboto-Regular.ttf")
doc.RegisterFont("bold",       "/path/to/Roboto-Bold.ttf")
doc.RegisterFont("italic",     "/path/to/Roboto-Italic.ttf")
doc.RegisterFont("mono",       "/path/to/RobotoMono-Regular.ttf")

fontFor := func(s html.Style) string {
    switch {
    case s.Bold && s.Italic: return "bold"   // no bold-italic font here
    case s.Bold:             return "bold"
    case s.Italic:           return "italic"
    case s.Monospace:        return "mono"
    default:                 return "regular"
    }
}

spans, _ := html.Parse(`Price: <b>€ 99</b> <i>(excl. VAT)</i>`, nil)

// endX is the X position after the last character — useful for continuing
// content on the same line.
endX, err := doc.WriteHTMLSpans(spans, fontFor, 11, x, y)
```

Underline and strikethrough decorations are drawn automatically as thin
horizontal lines immediately after the text.

---

## 5. HTML table parsing

`html.ParseTable` parses a string containing a `<table>` element and returns
an `HtmlTable`.

```go
htmlSrc := `
<table>
  <thead>
    <tr><th>Product</th><th>Price</th><th>Stock</th></tr>
  </thead>
  <tbody>
    <tr><td>Widget A</td><td align="right">€ 12.50</td><td>145</td></tr>
    <tr bgcolor="#f0f0f0"><td>Widget B</td><td align="right">€ 8.00</td><td>32</td></tr>
  </tbody>
</table>`

table, err := html.ParseTable(htmlSrc)
```

### Supported table structure elements

| Element             | Description                                              |
|---------------------|----------------------------------------------------------|
| `<table>`           | Root element                                             |
| `<caption>`         | Optional caption (available as `HtmlTable.Caption`)      |
| `<thead>`           | Header section — rows are marked `IsHeader: true`        |
| `<tbody>`           | Body section                                             |
| `<tfoot>`           | Footer section — rows are marked `IsFooter: true`        |
| `<tr>`              | Table row                                                |
| `<th>`              | Header cell — always `Bold: true`, marks row as header   |
| `<td>`              | Data cell                                                |

### Supported cell and row attributes

| Attribute / property        | Where      | Effect                                              |
|-----------------------------|------------|-----------------------------------------------------|
| `colspan="N"`               | `<td>/<th>`| Cell spans N columns (`HtmlCell.ColSpan`)           |
| `rowspan="N"`               | `<td>/<th>`| Cell spans N rows (`HtmlCell.RowSpan`)              |
| `align="left|center|right"` | `<tr>/<td>`| Horizontal text alignment                           |
| `valign="top|middle|bottom"`| `<tr>/<td>`| Vertical text alignment                             |
| `bgcolor="#RRGGBB"`         | `<tr>/<td>`| Background color (also via `style="background-color:…"`) |
| `style="color:…"`           | `<td>/<th>`| Text color                                          |
| `style="font-weight:bold"`  | `<td>/<th>`| Bold text                                           |
| `style="text-align:…"`      | `<td>/<th>`| Horizontal alignment (overrides `align`)            |
| `style="vertical-align:…"`  | `<td>/<th>`| Vertical alignment (overrides `valign`)             |
| `nowrap`                    | `<td>/<th>`| Sets `HtmlCell.NoWrap: true`                        |
| `width="…"`                 | `<td>/<th>`| Width hint — stored in `HtmlCell.Width` as a string |

Cell content may contain inline HTML tags (`<b>`, `<i>`, `<u>`, etc.) and
`<br>` (converted to `\n`).

---

## 6. Rendering HTML tables with TableFromHTML

After parsing, convert the `HtmlTable` to a `pdf.Table` with
`doc.TableFromHTML` and draw it like any other table.

```go
import (
    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/html"
)

htmlTable, err := html.ParseTable(htmlSrc)
if err != nil {
    log.Fatal(err)
}

cfg := pdf.TableConfig{
    ColWidths: []float64{200, 100, 80},  // required — no auto-sizing
    DefaultCellStyle: pdf.CellStyle{
        FontName: "regular",
        FontSize: 10,
        Padding:  pdf.Padding{Top: 4, Right: 6, Bottom: 4, Left: 6},
    },
}

htmlOpts := pdf.HtmlTableOptions{
    SpanFontFor: func(s html.Style) string {
        switch {
        case s.Bold:      return "bold"
        case s.Italic:    return "italic"
        case s.Monospace: return "mono"
        default:          return "regular"
        }
    },
    HeaderStyle: pdf.CellStyle{
        FontName:   "bold",
        FontSize:   10,
        Background: &pdf.Color{R: 0.2, G: 0.4, B: 0.7},
        TextColor:  &pdf.Color{R: 1, G: 1, B: 1},
    },
    FooterStyle: pdf.CellStyle{
        FontName: "italic",
        FontSize: 9,
    },
}

pdfTable, err := doc.TableFromHTML(htmlTable, cfg, htmlOpts)
if err != nil {
    log.Fatal(err)
}

// Draw the table at (x=50, y=100). Returns the Y position below the table.
endY, err := pdfTable.Draw(doc, 50, 100)
```

### HtmlTableOptions fields

| Field         | Type                             | Description                                          |
|---------------|----------------------------------|------------------------------------------------------|
| `SpanFontFor` | `func(html.Style) string`        | Maps a Style to a registered font name; nil = default|
| `HeaderStyle` | `pdf.CellStyle`                  | Applied to `<thead>` rows and all-`<th>` rows        |
| `FooterStyle` | `pdf.CellStyle`                  | Applied to `<tfoot>` rows                            |

Color values from `bgcolor` and inline `style="color:…"` are parsed
automatically and support:
- Named colors: `red`, `blue`, `green`, `gray`, `white`, `black`, …
- Hex: `#RRGGBB` and `#RGB`
- RGB function: `rgb(255, 128, 0)`

---

## 7. Full example — inline HTML

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/html"
)

func main() {
    doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
    doc.RegisterFont("regular", "/path/to/Roboto-Regular.ttf")
    doc.RegisterFont("bold",    "/path/to/Roboto-Bold.ttf")
    doc.RegisterFont("italic",  "/path/to/Roboto-Italic.ttf")
    doc.AddPage()

    fontFor := func(s html.Style) string {
        switch {
        case s.Bold:   return "bold"
        case s.Italic: return "italic"
        default:       return "regular"
        }
    }

    src := `Status: <b>Active</b> — owner: <i>Alice</i> — ref: <u>INV-2024-0042</u>`
    spans, err := html.Parse(src, nil)
    if err != nil {
        log.Fatal(err)
    }

    doc.SetFont("regular", 11)
    if _, err := doc.WriteHTMLSpans(spans, fontFor, 11, 50, 80); err != nil {
        log.Fatal(err)
    }

    if err := doc.Save("inline.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 8. Full example — HTML table

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/html"
)

func main() {
    doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
    doc.RegisterFont("regular", "/path/to/Roboto-Regular.ttf")
    doc.RegisterFont("bold",    "/path/to/Roboto-Bold.ttf")
    doc.AddPage()

    src := `
    <table>
      <thead>
        <tr>
          <th>Item</th>
          <th>Qty</th>
          <th align="right">Price</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>Widget <b>Pro</b></td>
          <td>3</td>
          <td align="right">€ 45.00</td>
        </tr>
        <tr bgcolor="#f5f5f5">
          <td>Gadget</td>
          <td>1</td>
          <td align="right">€ 12.50</td>
        </tr>
      </tbody>
    </table>`

    htmlTable, err := html.ParseTable(src)
    if err != nil {
        log.Fatal(err)
    }

    cfg := pdf.TableConfig{
        ColWidths: []float64{250, 60, 100},
        DefaultCellStyle: pdf.CellStyle{
            FontName: "regular",
            FontSize: 10,
            Padding:  pdf.Padding{Top: 4, Right: 6, Bottom: 4, Left: 6},
        },
    }

    opts := pdf.HtmlTableOptions{
        SpanFontFor: func(s html.Style) string {
            if s.Bold {
                return "bold"
            }
            return "regular"
        },
        HeaderStyle: pdf.CellStyle{
            FontName:   "bold",
            Background: &pdf.Color{R: 0.15, G: 0.35, B: 0.65},
            TextColor:  &pdf.Color{R: 1, G: 1, B: 1},
        },
    }

    pdfTable, err := doc.TableFromHTML(htmlTable, cfg, opts)
    if err != nil {
        log.Fatal(err)
    }

    if _, err := pdfTable.Draw(doc, 50, 80); err != nil {
        log.Fatal(err)
    }

    if err := doc.Save("table.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

---

## 9. Known limitations

- **Block-level elements** — only `<table>` is recognised as a block element.
  Divs, paragraphs, headings, and other block tags inside inline HTML are
  ignored; their text content is still included.
- **Nested tables** — nested `<table>` elements inside a cell are skipped.
- **CSS stylesheets** — only a small subset of inline `style="…"` properties
  are parsed (see the attribute table in section 5). External or embedded CSS
  is not supported.
- **`<br>` in cells** — `<br>` tags are converted to `\n` within cell content.
- **`width` hints** — the `width` attribute on `<td>` / `<th>` is stored in
  `HtmlCell.Width` as a raw string but is **not** applied automatically.
  Translate it to explicit `ColWidths` in `TableConfig` yourself if needed.
- **HTML entities** — only the most common entities (`&amp;`, `&lt;`, `&gt;`,
  `&nbsp;`, `&quot;`) are decoded.  Numeric character references (`&#160;`)
  are not supported.
