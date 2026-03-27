# Changelog

All notable changes to this project will be documented in this file.

---

## v0.2.0 — 2026-03-27

### New features

#### RML — XML-based document description (`pdf/rml`)

A new package that lets you define complete PDF documents in XML without writing
Go rendering code.  Declare fonts, page templates, paragraph styles, table
styles, and story content in a single `.rml` file.

- `rml.ParseFile(path, opts)` and `rml.Parse(r, opts)` return a ready-to-save
  `*pdf.Document`.
- Supports `<registerTTFont>` and `<registerFontFamily>` in `<docinit>`.
- Full page template support: `<pageTemplate>`, `<frame>`, `<pageGraphics>`
  with drawing commands (`<drawString>`, `<lines>`, `<rect>`, `<setFont>`,
  `<fill>`, `<stroke>`, `<saveState>`, `<restoreState>`, etc.).
- Stylesheet: `<paraStyle>` and `<blockTableStyle>` with line styling,
  background colors, alignment, and padding.
- Story elements: `<para>`, `<blockTable>`, `<image>`, `<ul>`, `<ol>`,
  `<spacer>`, `<hr>`, `<indent>`, `<keepTogether>`, `<condPageBreak>`,
  `<pageBreak>`, `<frameBreak>`, `<nextPageTemplate>`.
- Inline markup in paragraphs: `<b>`, `<i>`, `<u>`.
- Page number variable `%p` in page graphics.
- New example: [`examples/rml`](examples/rml/main.go) with a full invoice
  template ([`examples/rml/invoice.rml`](examples/rml/invoice.rml)).
- New documentation: [docs/rml-guide.en.md](docs/rml-guide.en.md) /
  [docs/rml-guide.md](docs/rml-guide.md).

#### HTML table parsing and rendering (`pdf/html`, `pdf`)

- New `html.ParseTable(string)` function parses a `<table>` element into an
  `html.HtmlTable` structure.
- Supported structural elements: `<table>`, `<caption>`, `<thead>`, `<tbody>`,
  `<tfoot>`, `<tr>`, `<th>`, `<td>`.
- Supported attributes: `colspan`, `rowspan`, `align`, `valign`, `bgcolor`,
  `nowrap`, `width`, and inline `style` properties (`background-color`,
  `color`, `text-align`, `vertical-align`, `font-weight`).
- Cell content supports inline HTML tags and `<br>`.
- New `doc.TableFromHTML(htmlTable, TableConfig, HtmlTableOptions)` converts a
  parsed `HtmlTable` to a `pdf.Table` with full colspan/rowspan and rich-span
  support.
- New `pdf.HtmlTableOptions` with `SpanFontFor`, `HeaderStyle`, `FooterStyle`.
- New `doc.WriteHTMLSpans(spans, fontFor, fontSize, x, y)` renders a
  `[]html.Span` inline, including automatic underline and strikethrough
  decoration.
- Color parsing in table attributes supports named colors, `#RRGGBB`, `#RGB`,
  and `rgb(r,g,b)`.

#### Extended inline HTML tag support (`pdf/html`)

The `html.Style` struct and `html.Parse` gained two new formatting flags:

- `Strikethrough` — set by `<s>`, `<strike>`, `<del>`
- `Monospace` — set by `<code>`, `<tt>`, `<kbd>`, `<samp>`

Additional italic aliases recognized: `<cite>`, `<var>`, `<dfn>`.
Additional underline alias: `<ins>`.

#### Struct-driven table cells (`pdf`)

New `pdf.CellsFromStruct` and `pdf.HeaderCellsFromStruct` generate table cells
directly from exported struct fields, driven by the `cell` struct tag.

Tag syntax supports: `text`, `format`, `header`, `halign`, `valign`, `bold`,
`font`, `size`, `bg`, `color`, `border`, `colspan`, `rowspan`, and `-` to skip
a field.

```go
type Item struct {
    Name  string  `cell:"halign=left"`
    Price float64 `cell:"halign=right;format=%.2f"`
    Stock int     `cell:"halign=center"`
}
cells := pdf.CellsFromStruct(item)
```

New example: [`examples/celltag`](examples/celltag/main.go).

### Improvements

#### Table (`pdf/table.go`)

Significant internal improvements and additions to the `pdf.Table` rendering
engine (396 lines added), including improved row-span handling, continuation
page logic, and support for the rich-span cells used by `TableFromHTML`.

#### Layout guide updated (`docs/layout-guide.md`)

The Dutch layout guide received 124 lines of additions keeping it in sync with
the English version.

### New documentation

- [docs/chart-guide.en.md](docs/chart-guide.en.md) — complete chart developer
  guide covering all 20 chart types with configuration reference and examples.
- [docs/chart-guide.md](docs/chart-guide.md) — Dutch version.
- [docs/html-guide.en.md](docs/html-guide.en.md) — inline HTML and HTML table
  rendering guide.
- [docs/html-guide.md](docs/html-guide.md) — Dutch version.
- [docs/rml-guide.en.md](docs/rml-guide.en.md) — RML user guide (English).
- [docs/rml-guide.md](docs/rml-guide.md) — RML user guide (Dutch).

### Examples

| Example | What's new |
|---------|-----------|
| [`examples/rml`](examples/rml/main.go) | New — full invoice rendered from an `.rml` file |
| [`examples/celltag`](examples/celltag/main.go) | New — struct-driven table generation |
| [`examples/html`](examples/html/main.go) | Extended to cover HTML table parsing and `WriteHTMLSpans` |
| All examples | Output PDFs committed alongside source |

---

## v0.1.0 — 2026-03-19

Initial public release.

- Core `pdf.Document` with text, borders, frames, tables, images, drawing
  primitives, headers/footers, two-pass `Build`, emoji, and RTL support.
- Layout engine: `DocTemplate`, `PageTemplate`, `LayoutFrame`, `Paragraph`,
  `Spacer`, `HRFlowable`, `KeepTogether`, action flowables.
- 20 chart types with Highcharts-style declarative API and `chart.NewFlowable`
  integration.
- Inline HTML span parsing (`<b>`, `<i>`, `<u>`, class attributes).
- Examples: `basic`, `html`, `layout`, `rtl`, `chart`.
- Documentation: [docs/layout-guide.en.md](docs/layout-guide.en.md) /
  [docs/layout-guide.md](docs/layout-guide.md).
