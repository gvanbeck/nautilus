# Nautilus

A pure-Go PDF generation library with support for Unicode text, emoji rendering,
tables, frames, borders, headers/footers, HTML inline markup, right-to-left
scripts (Arabic and Hebrew), a high-level layout engine, and 20 chart types.

Built on top of [gopdf](https://github.com/signintech/gopdf).

## Features

- **Paper formats** — A3, A4, A5, Letter, Legal (portrait); custom sizes supported
- **Margins** — configurable per-side page margins with content-area accessors (`ContentX`, `ContentY`, `ContentWidth`, `ContentHeight`, `ContentRightX`, `ContentBottomY`)
- **Font support** — TTF and OTF fonts; register multiple fonts (regular, bold, italic) and switch freely
- **Full Unicode** — Latin extended, CJK, Cyrillic, Greek, Arabic, Hebrew, and more
- **Emoji** — inline PNG substitution via grapheme-cluster resolution (Noto Emoji compatible)
- **Text rendering** — single-line (`WriteLine`) and word-wrapped (`WriteText`) with configurable line height
- **Right-to-left text** — Arabic contextual shaping (presentation forms + lam-alef ligatures) and Unicode BiDi reordering for Arabic and Hebrew
- **HTML inline markup** — convert `<b>`, `<i>`, `<u>`, and classed inline tags to styled text spans
- **Tables** — column spans, row spans, per-cell styling, horizontal/vertical alignment, automatic page overflow
- **Frames** — positioned rectangular content boxes with padding, borders, and background fill; nestable
- **Borders** — per-side control with solid, dashed, dotted, dash-dot, and custom dash patterns
- **Headers & footers** — callbacks invoked on every page with page-number context
- **Two-pass Build** — enables "Page N of M" footers by counting pages before rendering
- **Drawing primitives** — lines, polygons, circles, and rectangles drawn directly onto the page
- **Images** — PNG and JPEG inline via `DrawImage`
- **Layout engine** — `DocTemplate`, `PageTemplate`, `LayoutFrame`, and a `Flowable` story-based composition system inspired by ReportLab/Platypus
- **20 chart types** — line, area, column, bar, pie, polar, scatter, bubble, heatmap, waterfall, funnel, gauge, errorbar, boxplot, columnrange, arearange, bullet, dumbbell, lollipop, treemap
- **Output** — save to file or write to any `io.Writer`

## Documentation

| Guide | Description |
|-------|-------------|
| [Layout engine](docs/layout-guide.en.md) | `DocTemplate`, `PageTemplate`, `LayoutFrame`, the `Flowable` story model, multi-column layouts, and page decorators |
| [Charts](docs/chart-guide.en.md) | All 20 chart types with full configuration reference and examples |
| [HTML](docs/html-guide.en.md) | Inline HTML span parsing and HTML table rendering |
| [RML](docs/rml-guide.en.md) | XML-based document description — define pages, styles, and content without writing Go code |

Nederlandse versies: [Layout](docs/layout-guide.md) · [Charts](docs/chart-guide.md) · [HTML](docs/html-guide.md) · [RML](docs/rml-guide.md)

---

## Installation

```sh
go get github.com/gvanbeck/nautilus
```

## Quick start

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
)

func main() {
    doc, err := pdf.New(pdf.Config{
        PageSize:        pdf.PageSizeA4,
        DefaultFontSize: 12,
    })
    if err != nil {
        log.Fatal(err)
    }

    doc.AddPage()

    if err := doc.RegisterFont("regular", "/path/to/font.ttf"); err != nil {
        log.Fatal(err)
    }
    if err := doc.SetFont("regular", 14); err != nil {
        log.Fatal(err)
    }

    if _, err := doc.WriteLine("Hello, World!", 50, 100); err != nil {
        log.Fatal(err)
    }

    if err := doc.Save("hello.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

## API reference

### Document creation

```go
// Create a document with default settings (A4, 12 pt, 1.2× line height).
doc, err := pdf.New(pdf.Config{})

// Create with explicit settings.
doc, err := pdf.New(pdf.Config{
    PageSize:         pdf.PageSizeA4,      // or PageSizeA3, A5, Letter, Legal
    EmojiResolver:    resolver,            // optional emoji.Resolver
    DefaultFontSize:  12,
    LineHeightFactor: 1.4,
})
```

**Available page sizes:**

| Constant          | Width (pt) | Height (pt) |
|-------------------|-----------|-------------|
| `PageSizeA3`      | 841.89    | 1190.55     |
| `PageSizeA4`      | 595.28    | 841.89      |
| `PageSizeA5`      | 419.53    | 595.28      |
| `PageSizeLetter`  | 612       | 792         |
| `PageSizeLegal`   | 612       | 1008        |

### Pages

```go
doc.AddPage()                   // append a new page and make it active
doc.PageWidth()                 // page width in points
doc.PageHeight()                // page height in points
doc.PageCount()                 // number of pages added so far
```

### Margins

Set margins once in `Config`.  All write methods still accept explicit
coordinates; the margin accessors give you named references to the content
area so you never need to hard-code numeric offsets.

```go
doc, err := pdf.New(pdf.Config{
    PageSize: pdf.PageSizeA4,
    Margins:  pdf.UniformMargins(50),           // 50 pt on all sides
    // or:
    Margins:  pdf.Margins{Top: 60, Right: 50, Bottom: 60, Left: 50},
})

// Content area accessors
doc.ContentX()        // left edge  = margins.Left
doc.ContentY()        // top edge   = margins.Top
doc.ContentWidth()    // usable width  = page width  − left − right margin
doc.ContentHeight()   // usable height = page height − top  − bottom margin
doc.ContentRightX()   // right edge = page width − margins.Right  (RTL anchor)
doc.ContentBottomY()  // bottom edge = page height − margins.Bottom

// Use them when writing content
doc.WriteText(text, doc.ContentX(), doc.ContentY(), doc.ContentWidth())

// Right-to-left text
shaped := rtl.Shape("مرحبا بالعالم")
doc.WriteLineRTL(shaped, doc.ContentRightX(), y)

// Table overflow threshold
tbl := doc.NewTable(pdf.TableConfig{
    X:             doc.ContentX(),
    Y:             startY,
    ColWidths:     []float64{...},
    PageBottom:    doc.ContentBottomY(),
    ContinuationY: doc.ContentY(),
})
```

### Fonts

Both `.ttf` and `.otf` files are supported.

```go
// Register fonts under named aliases.
doc.RegisterFont("regular", "/path/to/NotoSans-Regular.ttf")
doc.RegisterFont("bold",    "/path/to/NotoSans-Bold.ttf")
doc.RegisterFont("italic",  "/path/to/NotoSans-Italic.otf")

// Activate a font at a given size (points).
doc.SetFont("regular", 12)
doc.SetFont("bold", 14)

// Measure text width in the current font.
width, err := doc.MeasureText("Hello")
```

### Text rendering

```go
// Write a single line at (x, y). Returns the X after the last character.
endX, err := doc.WriteLine("Hello, World! 👋", 50, 100)

// Write word-wrapped text within maxWidth. Returns the Y below the last line.
endY, err := doc.WriteText(longText, 50, 100, 495)

// Set text colour (RGB, 0–255).
doc.SetTextColor(60, 60, 60)

// Adjust the line-height multiplier (default 1.2).
doc.SetLineHeightFactor(1.5)
```

`WriteText` honours explicit `\n` newlines and breaks long lines at word
boundaries.

### Cursor position

```go
x := doc.GetX()  // current horizontal position
y := doc.GetY()  // current vertical position
```

### Right-to-left text (Arabic and Hebrew)

The `pdf/rtl` package prepares RTL text for rendering by applying Arabic
contextual letter shaping and Unicode Bidirectional Algorithm (BiDi) reordering.
The result is a string in visual (left-to-right glyph) order that can be passed
to the RTL rendering methods.

```go
import "github.com/gvanbeck/nautilus/pdf/rtl"
```

**Single-line RTL rendering:**

```go
// 1. Shape and reorder the text.
shaped := rtl.Shape("مرحبا بالعالم")

// 2. Render with the right edge at rightX.
leftX, err := doc.WriteLineRTL(shaped, rightEdge, y)

// Hebrew (no Arabic shaping needed — Shape still applies BiDi reordering).
shaped := rtl.Shape("שלום עולם")
leftX, err := doc.WriteLineRTL(shaped, rightEdge, y)
```

**Multi-line (word-wrapped) RTL rendering:**

`WriteTextRTL` handles shaping, wrapping, and per-line BiDi reordering
internally so that word order is preserved correctly across line breaks.
Pass the original (logical-order) text directly.

```go
// Explicit newlines (\n) are treated as paragraph breaks.
endY, err := doc.WriteTextRTL("مرحبا بالعالم\nكيف حالك", rightEdge, y, maxWidth)
```

**RTL inside a Frame:**

```go
f := doc.NewFrame(pdf.FrameConfig{
    X: 50, Y: 200, Width: 495,
    Padding: pdf.UniformPadding(8),
})
f.SetFont("arabic", 12)

// Single line — text must be pre-shaped.
shaped := rtl.Shape("مرحبا")
f.WriteLineRTL(shaped)

// Multi-line — pass original text, shaping is applied internally.
f.WriteTextRTL("مرحبا بالعالم\nكيف حالك")

f.Close()
```

**`rtl` package functions:**

| Function | Description |
|----------|-------------|
| `rtl.Shape(text)` | Arabic shaping + BiDi reorder → use for single lines. |
| `rtl.ShapeOnly(text)` | Arabic shaping only, logical order preserved → rarely needed directly. |
| `rtl.Reorder(text)` | BiDi reorder only, no Arabic shaping → suitable for Hebrew. |

**Font requirements for Arabic:**

The font must include the **Unicode Arabic Presentation Forms-B** block
(U+FE70–U+FEFF). Recommended fonts: [Noto Naskh Arabic](https://fonts.google.com/noto/specimen/Noto+Naskh+Arabic),
[Amiri](https://fonts.google.com/specimen/Amiri).
For Hebrew, any font covering the Hebrew block (U+0590–U+05FF) is sufficient.

**Covered Arabic letters:**

All letters of the basic Arabic alphabet (U+0621–U+064A) including alef
variants (آ أ إ ا), ta marbuta (ة), waw (و), ya (ي), and the mandatory
lam-alef ligatures (لا لأ لإ لآ).  Arabic diacritics (harakat) are treated
as transparent during joining and pass through unchanged.

### HTML inline markup

The `pdf/html` package converts a string of inline HTML into a slice of `Span`
values, each carrying text, formatting flags, and an optional CSS class name.

```go
import "github.com/gvanbeck/nautilus/pdf/html"
```

**Supported tags:** `<b>`, `<strong>`, `<i>`, `<em>`, `<cite>`, `<var>`,
`<dfn>`, `<u>`, `<ins>`, `<s>`, `<strike>`, `<del>`, `<code>`, `<tt>`,
`<kbd>`, `<samp>`, and any tag with a `class` attribute. Tags may be freely
nested.

```go
spans, err := html.Parse("<b>bold</b> and <i>italic</i>", nil)
```

Each `Span` contains:

```go
type Span struct {
    Text  string      // plain text content
    Style html.Style  // Bold, Italic, Underline, Strikethrough, Monospace flags
    Class string      // CSS class name (innermost), if present
}
```

**Class-based styles:**

Pass a `ClassStyle` map to merge additional style flags onto spans whose
tag carries a matching `class` attribute.  The class name is always preserved
in `Span.Class` regardless.

```go
cs := html.ClassStyle{
    "highlight": {Bold: true},
    "note":      {Italic: true},
    "important": {Bold: true, Underline: true},
}
spans, err := html.Parse(`<span class="highlight">text</span>`, cs)
// spans[0].Style.Bold == true
// spans[0].Class == "highlight"
```

**Rendering spans with `WriteHTMLSpans`:**

```go
fontFor := func(s html.Style) string {
    switch {
    case s.Bold:      return "bold"
    case s.Italic:    return "italic"
    case s.Monospace: return "mono"
    default:          return "regular"
    }
}
endX, err := doc.WriteHTMLSpans(spans, fontFor, fontSize, x, y)
```

Underline and strikethrough decorations are drawn automatically.

**HTML table parsing:**

Parse a `<table>` element and render it as a `pdf.Table`:

```go
htmlTable, err := html.ParseTable(htmlString)

pdfTable, err := doc.TableFromHTML(htmlTable, pdf.TableConfig{
    ColWidths: []float64{200, 100, 80},
    DefaultCellStyle: pdf.CellStyle{FontName: "regular", FontSize: 10},
}, pdf.HtmlTableOptions{
    SpanFontFor: fontFor,
    HeaderStyle: pdf.CellStyle{FontName: "bold", Background: &headerBg, TextColor: &white},
})
endY, err := pdfTable.Draw(doc, x, y)
```

Supports `colspan`, `rowspan`, `align`, `valign`, `bgcolor`, inline `style`,
`<thead>` / `<tbody>` / `<tfoot>`, and inline HTML tags within cells.

→ **Full reference:** [docs/html-guide.en.md](docs/html-guide.en.md)

### Emoji support

Emoji are rendered as inline PNG images sized to match the current font.
Supply a `Resolver` that maps each emoji grapheme cluster to a PNG file path.

```go
import "github.com/gvanbeck/nautilus/pdf/emoji"

// Use the built-in Noto Emoji resolver.
resolver := &emoji.NotoResolver{Dir: "/path/to/noto-emoji/png/128"}

doc, _ := pdf.New(pdf.Config{
    EmojiResolver: resolver,
})
```

Download Noto Emoji PNGs (Apache 2.0) from
[googlefonts/noto-emoji](https://github.com/googlefonts/noto-emoji/tree/main/png/128).

**Emoji segmentation** — the `emoji` package also exposes text segmentation:

```go
segments := emoji.Split("Hi 👋 there 🌍")
// → [{KindText "Hi "}, {KindEmoji "👋"}, {KindText " there "}, {KindEmoji "🌍"}]

// Convert a cluster to a Noto-style filename.
emoji.ClusterToFilename("👨‍👩‍👧")
// → "emoji_u1f468_200d_1f469_200d_1f467.png"
```

**Custom resolver** — implement the `emoji.Resolver` interface:

```go
type Resolver interface {
    Resolve(cluster string) (path string, found bool)
}
```

### Borders

Borders can be drawn around any rectangle. Each side is independently
configurable with its own thickness, colour, and line pattern.

```go
// Uniform border — all four sides identical.
border := pdf.NewUniformBorder(pdf.BorderSpec{
    Thickness: 1.5,
    Color:     pdf.ColorNavy,
    Pattern:   pdf.PatternSolid,
})
doc.DrawBorder(50, 100, 495, 40, border)

// Per-side border — only top and bottom.
doc.DrawBorder(50, 100, 495, 40, pdf.Border{
    Top:    &pdf.BorderSpec{Thickness: 2, Color: pdf.ColorNavy, Pattern: pdf.PatternSolid},
    Bottom: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorGray, Pattern: pdf.PatternDashed},
})
```

**Border patterns:**

| Constant          | Description                              |
|-------------------|------------------------------------------|
| `PatternSolid`    | Continuous unbroken line                 |
| `PatternDashed`   | Long-dash / gap pattern                  |
| `PatternDotted`   | Short-dot / gap pattern                  |
| `PatternDashDot`  | Alternating long dash and short dot      |
| `PatternCustom`   | Custom dash array via `DashArray` field  |

```go
// Custom dash pattern.
spec := pdf.BorderSpec{
    Thickness: 2,
    Color:     pdf.ColorRed,
    Pattern:   pdf.PatternCustom,
    DashArray: []float64{12, 4, 4, 4},
    DashPhase: 0,
}
```

**Predefined colours:**

`ColorBlack`, `ColorWhite`, `ColorLightGray`, `ColorGray`, `ColorDarkGray`,
`ColorRed`, `ColorGreen`, `ColorBlue`, `ColorNavy`, `ColorOrange`

```go
custom := pdf.Color{R: 235, G: 245, B: 255}
```

### Frames

Frames are positioned rectangular content boxes — similar to LaTeX minipages.
Content flows downward automatically within the frame.

```go
// Fixed-height frame with background fill and accent border.
f := doc.NewFrame(pdf.FrameConfig{
    X: 50, Y: 200, Width: 495, Height: 80,
    Background: &pdf.Color{R: 235, G: 245, B: 255},
    Border: pdf.Border{
        Left: &pdf.BorderSpec{Thickness: 4, Color: pdf.ColorNavy},
    },
    Padding: pdf.Padding{Top: 8, Right: 12, Bottom: 8, Left: 16},
})
f.SetFont("regular", 10)
f.SetTextColor(20, 20, 80)
f.WriteText("This text flows inside the frame.")
f.Close() // draws the border

// Auto-height frame (Height: 0) — border adapts to content.
f := doc.NewFrame(pdf.FrameConfig{
    X: 50, Y: 300, Width: 230,
    Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
    Padding: pdf.UniformPadding(8),
})
f.SetFont("regular", 10)
f.WriteText("The frame height adjusts to fit this content.")
f.Close()
```

**Frame methods:**

```go
f.ContentX()       // left edge of content area (X + padding left)
f.ContentWidth()   // usable width (frame width − horizontal padding)
f.CurrentY()       // Y position of the next content line
f.FrameHeight()    // current outer height (fixed or computed)

f.WriteLine(text)                // render on current line (no Y advance)
f.WriteLineAt(text, xOffset)    // render at offset from content left edge
f.WriteText(text)                // word-wrapped, advances Y
f.WriteLineRTL(shaped)          // RTL single line, right-aligned (pre-shaped)
f.WriteTextRTL(text)            // RTL word-wrapped, shaping applied internally
f.Advance(n)                     // move Y down by n points
f.NewLine()                      // move Y down by one line height

f.SetFont(name, size)            // delegates to Document.SetFont
f.SetTextColor(r, g, b)         // delegates to Document.SetTextColor
f.MeasureText(text)              // delegates to Document.MeasureText

// Draw a border inside the frame at a relative offset.
f.DrawInnerBorder(xOffset, yOffset, width, height, border)

f.Close()  // finalise: draw outer border (idempotent)
```

**Padding helpers:**

```go
pdf.UniformPadding(8)           // 8 pt on all sides
pdf.HorizontalPadding(12, 6)   // 12 pt left/right, 6 pt top/bottom
pdf.Padding{Top: 5, Right: 8, Bottom: 5, Left: 8}  // explicit
```

### Tables

Tables provide grid-based layout with column spans, row spans, per-cell
styling, and automatic page overflow.

```go
tbl := doc.NewTable(pdf.TableConfig{
    X: 50, Y: 100,
    ColWidths: []float64{120, 260, 115},     // explicit column widths
    PageBottom:    doc.PageHeight() - 60,     // overflow threshold
    ContinuationY: 60,                       // Y on continuation pages
    Border: pdf.NewUniformBorder(pdf.BorderSpec{
        Thickness: 1.5, Color: pdf.ColorNavy,
    }),
    DefaultCellStyle: pdf.CellStyle{
        Padding:  pdf.Padding{Top: 5, Right: 8, Bottom: 5, Left: 8},
        Border:   pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
        FontName: "regular",
        FontSize: 10,
    },
})
```

**Adding rows:**

```go
// Single row.
tbl.AddRow(pdf.Row{
    Height: 24,  // fixed height; 0 = auto-height from content
    Cells: []pdf.Cell{
        {Text: "Name"},
        {Text: "Description"},
        {Text: "Value"},
    },
})

// Multiple rows at once.
tbl.AddRows(row1, row2, row3)
```

**Column span:**

```go
tbl.AddRow(pdf.Row{
    Cells: []pdf.Cell{
        {Text: "Spanning all columns", ColSpan: 3},
    },
})
```

**Row span:**

```go
tbl.AddRow(pdf.Row{
    Cells: []pdf.Cell{
        {Text: "Spans 2 rows", RowSpan: 2},
        {Text: "Row 1, Col 2"},
        {Text: "Row 1, Col 3"},
    },
})
tbl.AddRow(pdf.Row{
    Cells: []pdf.Cell{
        // Column 1 occupied by rowspan — omit it.
        {Text: "Row 2, Col 2"},
        {Text: "Row 2, Col 3"},
    },
})
```

**Per-cell styling:**

```go
navy := pdf.ColorNavy
white := pdf.ColorWhite

tbl.AddRow(pdf.Row{
    Cells: []pdf.Cell{
        {Text: "Header", Style: pdf.CellStyle{
            Background: &navy,
            TextColor:  &white,
            FontName:   "bold",
            FontSize:   11,
            HAlign:     pdf.HAlignCenter,
            VAlign:     pdf.VAlignMiddle,
            Padding:    pdf.UniformPadding(6),
            Border:     pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1, Color: pdf.ColorNavy}),
        }},
    },
})
```

**Row background:**

```go
bg := pdf.Color{R: 230, G: 240, B: 255}
tbl.AddRow(pdf.Row{
    Background: &bg,
    Cells:      []pdf.Cell{{Text: "Shaded row"}},
})
```

**Cell alignment:**

| Horizontal          | Vertical            |
|---------------------|---------------------|
| `HAlignDefault`     | `VAlignDefault`     |
| `HAlignLeft`        | `VAlignTop`         |
| `HAlignCenter`      | `VAlignMiddle`      |
| `HAlignRight`       | `VAlignBottom`      |

**Drawing the table:**

```go
if err := tbl.Draw(); err != nil {
    log.Fatal(err)
}
```

Tables automatically call `doc.AddPage()` when a row group exceeds the
remaining space on the current page. Rows joined by a `RowSpan` are kept
together and never split across a page break.

### Headers and footers

Register callbacks that are invoked automatically on every page.

```go
doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
    d.SetFont("regular", 8)
    d.SetTextColor(100, 100, 100)
    d.WriteLine("My Document", 50, 15)
})

doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
    d.SetFont("regular", 8)
    d.SetTextColor(120, 120, 120)
    label := fmt.Sprintf("Page %d of %d", info.Number, info.Total)
    w, _ := d.MeasureText(label)
    d.WriteLine(label, (d.PageWidth()-w)/2, d.PageHeight()-20)
})
```

`PageInfo` provides:

- `Number` — 1-based index of the current page
- `Total` — total number of pages (0 when unknown)

**Known total pages** — if you know the count upfront:

```go
doc.SetTotalPages(10)
```

### Two-pass Build

Use `Build` when footers need to display the total page count but the count
is not known in advance. Build executes the callback twice: first to count
pages, then to render with the total available.

```go
doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
    d.SetFont("regular", 8)
    label := fmt.Sprintf("Page %d of %d", info.Number, info.Total)
    d.WriteLine(label, 50, d.PageHeight()-20)
})

doc.Build(func() {
    doc.AddPage()
    doc.SetFont("regular", 12)
    doc.WriteText("First page content…", 50, 60, 495)

    doc.AddPage()
    doc.SetFont("regular", 12)
    doc.WriteText("Second page content…", 50, 60, 495)
})

doc.Save("report.pdf")
```

During the counting pass:
- `AddPage` increments the counter but produces no PDF content
- `SetFont` tracks font state but does not call gopdf
- `WriteLine`, `WriteText`, `WriteLineRTL`, `WriteTextRTL`, `SetTextColor`, `DrawBorder` are no-ops
- `RegisterFont` always executes (fonts must be available for both passes)

### Output

```go
// Save to a file.
doc.Save("output.pdf")

// Write to any io.Writer.
var buf bytes.Buffer
doc.Output(&buf)
```

## Drawing primitives

The coordinate system for all drawing methods places the origin at the
top-left corner of the page, with X increasing rightward and Y increasing
downward. All measurements are in points (1 pt = 1/72 inch).

All drawing primitives are no-ops during the counting pass of `Build`.

### Lines

```go
// Draw a straight line from (x1,y1) to (x2,y2).
doc.DrawLine(x1, y1, x2, y2, lineWidth float64, color pdf.Color)
```

```go
// Draw a diagonal separator line.
doc.DrawLine(50, 100, 545, 100, 0.5, pdf.ColorGray)
```

### Polygons

```go
// Draw a filled polygon. Requires at least 3 points.
// The polygon is automatically closed (last point connects to first).
doc.FillPolygon(points []pdf.Point, color pdf.Color)

// Draw a filled polygon with a stroked outline.
doc.FillAndStrokePolygon(points []pdf.Point, fillColor pdf.Color, lineWidth float64, strokeColor pdf.Color)
```

```go
// Draw a filled triangle.
doc.FillPolygon([]pdf.Point{
    {X: 100, Y: 200},
    {X: 150, Y: 120},
    {X: 200, Y: 200},
}, pdf.Color{R: 70, G: 130, B: 180})

// Draw a filled triangle with a dark outline.
doc.FillAndStrokePolygon(
    []pdf.Point{{X: 100, Y: 200}, {X: 150, Y: 120}, {X: 200, Y: 200}},
    pdf.Color{R: 70, G: 130, B: 180},
    1.5,
    pdf.ColorNavy,
)
```

### Circles

```go
// Draw a filled circle centered at (cx, cy) with radius r.
doc.FillCircle(cx, cy, r float64, color pdf.Color)

// Draw the outline of a circle.
doc.StrokeCircle(cx, cy, r, lineWidth float64, color pdf.Color)
```

```go
// Filled circle.
doc.FillCircle(150, 200, 30, pdf.Color{R: 255, G: 100, B: 100})

// Outlined circle.
doc.StrokeCircle(150, 200, 30, 1.5, pdf.ColorNavy)
```

Circles are approximated using 32 polygon vertices, which is indistinguishable
from a true circle at normal document resolutions.

### Rectangles

```go
// Draw a solid filled rectangle. Origin is top-left; Y increases downward.
doc.FillRect(x, y, w, h float64, color pdf.Color)
```

```go
// Draw a navy header band across the full content width.
doc.FillRect(doc.ContentX(), 40, doc.ContentWidth(), 24, pdf.ColorNavy)
```

### Images

```go
// Render a PNG or JPEG image at (x, y) scaled to the given width and height.
err := doc.DrawImage(path string, x, y, width, height float64) error
```

```go
// Place a logo in the top-left corner of the content area.
if err := doc.DrawImage("logo.png", doc.ContentX(), doc.ContentY(), 80, 40); err != nil {
    log.Fatal(err)
}
```

### Graphics state

The graphics state stack allows you to save and restore drawing attributes
(line width, stroke color, fill color) so that chart and drawing code cannot
interfere with each other.

```go
doc.SaveGraphicsState()    // push current state onto the stack
// ... draw with temporary settings ...
doc.RestoreGraphicsState() // pop and restore previous state
```

Calls must be balanced. Both methods are no-ops during the counting pass of
`Build`.

## Layout engine (`pdf/layout`)

The layout engine provides high-level, automatic page flow. Instead of
computing positions manually, you build a flat list of `Flowable` elements
(the _story_) and let the engine distribute them across frames and pages.

```go
import "github.com/gvanbeck/nautilus/pdf/layout"
```

### Architecture

```
Story  ([]Flowable)
    ↓ consumed by
DocTemplate          — page/frame scheduler
    ↓ manages
PageTemplate         — page geometry (ordered LayoutFrames + decorators)
    ↓ contains
LayoutFrame          — rectangular region with downward Y cursor
    ↓ draws into
pdf.Document         — underlying PDF canvas
```

### The Flowable interface

Every content element implements `Flowable`:

```go
type Flowable interface {
    // Wrap measures the flowable within the available space.
    // Returns the actual (width, height) the flowable will occupy.
    // A returned height greater than availHeight signals the flowable does
    // not fit and must be split or moved to the next frame.
    Wrap(doc *pdf.Document, availWidth, availHeight float64) (float64, float64)

    // Draw renders the flowable with its top-left corner at (x, y).
    // Always called after a successful Wrap.
    Draw(doc *pdf.Document, x, y float64) error

    // Split divides the flowable so the first part fits within availHeight.
    // Returns nil when splitting is not possible; the engine moves the
    // flowable to the next frame. Returned parts must reproduce all content.
    Split(doc *pdf.Document, availWidth, availHeight float64) []Flowable

    SpaceBefore() float64  // extra whitespace above this flowable
    SpaceAfter() float64   // extra whitespace below this flowable
    KeepWithNext() bool    // prevent a break between this and the next flowable
    MinWidth() float64     // minimum width required
}
```

Key invariants:
- `Wrap` is always called before `Draw` or `Split`.
- Leading space (`SpaceBefore`) is suppressed at the top of a fresh frame.
- `DocTemplate.Build` has loop detection: it errors after 10 consecutive
  failed placements.

### Built-in flowables

#### Paragraph

Renders word-wrapped text with per-paragraph font, colour, alignment, and
spacing control.

```go
style := layout.ParagraphStyle{
    FontName:         "regular",    // registered font name
    FontSize:         12,           // points; 0 uses document default
    Leading:          16,           // line height; 0 defaults to FontSize × 1.2
    Alignment:        layout.AlignLeft, // AlignLeft, AlignCenter, AlignRight
    SpaceBefore:      8,            // extra space above the paragraph
    SpaceAfter:       6,            // extra space below the paragraph
    KeepWithNextPara: true,         // prevent break before next flowable
    LeftIndent:       20,           // reduce usable width from the left
    RightIndent:      20,           // reduce usable width from the right
    TextColor:        &pdf.Color{R: 40, G: 40, B: 40},
}

p := &layout.Paragraph{Text: "Hello, layout engine!", Style: style}
```

Long paragraphs are automatically split across frames at line boundaries.

#### Spacer

Reserves a fixed amount of vertical space without rendering anything.

```go
&layout.Spacer{Height: 12}             // 12 pt gap
&layout.Spacer{Width: 80, Height: 12}  // 80 pt wide, 12 pt tall
```

#### HRFlowable

Draws a horizontal rule as a solid filled bar.

```go
&layout.HRFlowable{
    Width:     0.8,              // fraction of available width (0..1) or absolute pts (>1)
    Thickness: 1,                // bar height in points; defaults to 1
    Color:     pdf.ColorGray,
    Align:     layout.AlignCenter, // AlignLeft, AlignCenter, AlignRight
    Before:    6,                // space above the rule
    After:     6,                // space below the rule
}
```

#### KeepTogether

Prevents a group of flowables from being split across frames. If the group
does not fit in the remaining frame space, the engine inserts a `FrameBreak`
and retries on the next frame. If the group is larger than an entire frame,
individual flowables are returned for independent splitting.

```go
// Keep a heading together with its first body paragraph.
&layout.KeepTogether{
    Flowables: []layout.Flowable{
        &layout.Paragraph{Text: "Section Heading", Style: h1Style},
        &layout.Paragraph{Text: "First paragraph…", Style: bodyStyle},
    },
}
```

### Action flowables

Action flowables are zero-height elements that control the engine rather than
rendering visible content.

#### PageBreak

Forces an immediate page break.

```go
// Simple page break.
&layout.PageBreak{}

// Page break with immediate template switch.
&layout.PageBreak{NextTemplate: "TwoColumn"}
```

#### FrameBreak

Advances the engine to the next frame (or the next page when no more frames
remain on the current page).

```go
&layout.FrameBreak{}
```

#### CondPageBreak

Inserts a page break only when fewer than `MinHeight` points remain in the
current frame.

```go
// Break if less than 72 pt (one inch) remains.
&layout.CondPageBreak{MinHeight: 72}
```

#### NextPageTemplate

Schedules a template switch that takes effect on the next page break. The
current page continues to use the existing template.

```go
story = append(story,
    titleContent...,
    &layout.NextPageTemplate{TemplateID: "TwoColumn"},
    &layout.PageBreak{},
    bodyContent...,
)
```

### LayoutFrame

A `LayoutFrame` is a rectangular region that receives flowables. The frame
maintains an internal Y cursor that advances downward as content is added.

```go
frame := &layout.LayoutFrame{
    X:            50,    // top-left X in page coordinates (points)
    Y:            80,    // top-left Y in page coordinates (points)
    Width:        495,   // outer width in points
    Height:       700,   // outer height in points
    Padding:      pdf.Padding{Top: 8, Right: 8, Bottom: 8, Left: 8},
    ID:           "main",    // optional name for debugging
    ShowBoundary: false,     // draw a thin outline when true (useful during development)
}
```

### PageTemplate

A `PageTemplate` groups one or more `LayoutFrame`s with optional decorators.

```go
tmpl := &layout.PageTemplate{
    ID:               "single",          // referenced by NextPageTemplate / PageBreak
    Frames:           []*layout.LayoutFrame{frame},
    OnPage:           headerFooterFunc,  // called after AddPage (draw headers, watermarks)
    OnPageEnd:        func(doc *pdf.Document, pageNum int) { /* ... */ },
    AutoNextTemplate: "single",          // switch to this template after each page
}
```

`PageDecorator` signature:

```go
type PageDecorator func(doc *pdf.Document, pageNum int)
```

### DocTemplate

`DocTemplate` is the engine that processes the story.

```go
dt := layout.NewDocTemplate(doc)
dt.AddPageTemplate(singleTemplate)
dt.AddPageTemplate(twoColTemplate)

if err := dt.Build(story); err != nil {
    log.Fatal(err)
}
```

- `NewDocTemplate(doc)` — create an engine for the given `pdf.Document`.
- `AddPageTemplate(pt)` — register a page template. The first registered
  template is used for the first page.
- `Build(story)` — flow all flowables across frames and pages. Returns an
  error if no templates are registered or if a flowable cannot fit in any
  frame after 10 attempts.

### Minimal example

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/layout"
)

func main() {
    doc, _ := pdf.New(pdf.Config{
        PageSize: pdf.PageSizeA4,
        Margins:  pdf.UniformMargins(50),
    })
    doc.RegisterFont("regular", "/path/to/font.ttf")
    doc.SetFont("regular", 12)

    style := layout.ParagraphStyle{FontName: "regular", FontSize: 12}
    story := []layout.Flowable{
        &layout.Paragraph{Text: "Hello, Nautilus!", Style: style},
        &layout.Spacer{Height: 12},
        &layout.Paragraph{Text: "Second paragraph.", Style: style},
    }

    frame := &layout.LayoutFrame{
        X: doc.ContentX(), Y: doc.ContentY(),
        Width: doc.ContentWidth(), Height: doc.ContentHeight(),
    }
    tmpl := &layout.PageTemplate{ID: "main", Frames: []*layout.LayoutFrame{frame}}

    dt := layout.NewDocTemplate(doc)
    dt.AddPageTemplate(tmpl)
    if err := dt.Build(story); err != nil {
        log.Fatal(err)
    }
    doc.Save("output.pdf")
}
```

### Multi-column layout

Provide two `LayoutFrame`s per `PageTemplate`. The engine fills the left
frame first, then the right frame, then starts a new page.

```go
const (
    margin  = 50.0
    gutter  = 12.0
    headerH = 40.0
    footerH = 36.0
)

pageW := doc.PageWidth()
pageH := doc.PageHeight()
contentX := margin
contentY := margin + headerH
contentW := pageW - 2*margin
contentH := pageH - margin - headerH - footerH

colW := (contentW - gutter) / 2

leftFrame := &layout.LayoutFrame{
    X: contentX,           Y: contentY,
    Width: colW,           Height: contentH,
    ShowBoundary: true,    // show outline during development
}
rightFrame := &layout.LayoutFrame{
    X: contentX + colW + gutter, Y: contentY,
    Width: colW,                 Height: contentH,
    ShowBoundary: true,
}

pageDecorator := func(d *pdf.Document, pageNum int) {
    d.SetFont("regular", 8)
    d.WriteLine("My Document", margin, margin+10)
}

twoColTemplate := &layout.PageTemplate{
    ID:               "two-column",
    Frames:           []*layout.LayoutFrame{leftFrame, rightFrame},
    OnPage:           pageDecorator,
    AutoNextTemplate: "two-column",
}
```

### Template switching

Switch templates at page boundaries to implement first-page vs. body-page
layouts:

```go
singleTemplate := &layout.PageTemplate{ID: "single", Frames: []*layout.LayoutFrame{singleFrame}}
twoColTemplate  := &layout.PageTemplate{ID: "two-column", Frames: []*layout.LayoutFrame{leftFrame, rightFrame}}

dt := layout.NewDocTemplate(doc)
dt.AddPageTemplate(singleTemplate)
dt.AddPageTemplate(twoColTemplate)

story := []layout.Flowable{
    // ... title page content ...
    &layout.NextPageTemplate{TemplateID: "two-column"},
    &layout.PageBreak{},
    // ... body content flows in two columns ...
    &layout.NextPageTemplate{TemplateID: "single"},
    &layout.PageBreak{},
    // ... back to single column ...
}

dt.Build(story)
```

## RML (`pdf/rml`)

The `pdf/rml` package lets you describe complete PDF documents in XML without
writing any Go rendering code.  An RML file declares fonts, page templates,
paragraph styles, table styles, and story content in a single file.

```go
import "github.com/gvanbeck/nautilus/pdf/rml"

doc, err := rml.ParseFile("invoice.rml", rml.Options{FontDir: "/path/to/fonts"})
doc.Save("invoice.pdf")
```

Or from the command line:

```sh
go run ./examples/rml -rml examples/rml/invoice.rml -fontdir /Library/Fonts -out invoice.pdf
```

**Supported elements:** page templates, frames, page graphics (headers/footers),
paragraph and table styles, `<para>`, `<blockTable>`, `<image>`, `<ul>` / `<ol>`,
`<spacer>`, `<hr>`, `<keepTogether>`, `<pageBreak>`, `<condPageBreak>`,
`<nextPageTemplate>`, and font registration.

→ **Full reference:** [docs/rml-guide.en.md](docs/rml-guide.en.md)

---

## Charts (`pdf/chart`)

Nautilus includes 20 chart types with a declarative API that mirrors the
Highcharts JSON configuration model. Charts draw directly onto `pdf.Document`
and integrate seamlessly with the layout engine via `chart.NewFlowable`.

```go
import (
    "github.com/gvanbeck/nautilus/pdf/chart"
    "github.com/gvanbeck/nautilus/pdf/chart/line"
)
```

### Chart sub-packages

Each chart type lives in its own importable sub-package so that binaries only
pay for the renderers they use.

| Package | Chart type |
|---------|------------|
| `pdf/chart/line` | Line chart — X/Y lines with optional markers |
| `pdf/chart/area` | Area chart — filled line chart |
| `pdf/chart/column` | Column chart — vertical bars, grouped or stacked |
| `pdf/chart/bar` | Bar chart — horizontal bars |
| `pdf/chart/pie` | Pie and donut chart |
| `pdf/chart/polar` | Polar / spider / radar chart |
| `pdf/chart/scatter` | Scatter chart — X/Y point cloud |
| `pdf/chart/bubble` | Bubble chart — scatter with Z-sized circles |
| `pdf/chart/heatmap` | Heatmap — color-coded grid |
| `pdf/chart/waterfall` | Waterfall — running-total bar chart |
| `pdf/chart/funnel` | Funnel and pyramid chart |
| `pdf/chart/gauge` | Gauge and solid-gauge chart |
| `pdf/chart/errorbar` | Error bar chart |
| `pdf/chart/boxplot` | Box-and-whisker chart |
| `pdf/chart/columnrange` | Column range — low/high vertical bars |
| `pdf/chart/arearange` | Area range — low/high filled band |
| `pdf/chart/bullet` | Bullet chart — bar with target marker and qualitative bands |
| `pdf/chart/dumbbell` | Dumbbell — low/high range dots connected by a line |
| `pdf/chart/lollipop` | Lollipop — stick with terminal dot |
| `pdf/chart/treemap` | Treemap — hierarchical rectangle packing |

### The Drawable interface

All chart types implement `Drawable`:

```go
type Drawable interface {
    Draw(doc *pdf.Document, x, y, width, height float64) error
}
```

`x, y` is the top-left corner of the bounding box in points.

### Embedding charts in a layout story

Use `chart.NewFlowable` to wrap any `Drawable` as a `layout.Flowable`:

```go
// width: 0 fills the available frame width; height is fixed.
story = append(story, chart.NewFlowable(myChart, 0, 220))
```

### chart.Options

`chart.Options` is the top-level configuration object.

```go
opts := chart.Options{
    FontName:   "regular",             // registered font name; must be on the Document
    FontSize:   9,                     // base font size in points; defaults to 9
    Title:      &chart.Title{Text: "Sales by Quarter"},
    Subtitle:   &chart.Title{Text: "2023 vs 2024"},
    XAxis:      &chart.Axis{Categories: []string{"Q1", "Q2", "Q3", "Q4"}},
    YAxis:      &chart.Axis{},
    Series:     []chart.Series{...},
    Legend:     &chart.Legend{},
    PlotOptions: &chart.PlotOptions{...},
    Colors:     nil,                   // nil uses DefaultColors
    Background: &pdf.Color{R: 250, G: 250, B: 250},
}
```

### chart.Title

```go
&chart.Title{
    Text:     "Chart Title",
    FontName: "bold",       // overrides Options.FontName
    FontSize: 11,           // overrides Options.FontSize when > 0
    Color:    &pdf.Color{R: 30, G: 30, B: 30},
}
```

### chart.Axis

```go
&chart.Axis{
    Title:         &chart.Title{Text: "Revenue (USD)"},
    Categories:    []string{"Q1", "Q2", "Q3", "Q4"}, // discrete tick labels
    Min:           chart.Float(0),    // clamp minimum visible value
    Max:           chart.Float(500),  // clamp maximum visible value
    TickInterval:  chart.Float(100),  // fixed gridline spacing
    GridLineWidth: 0.5,               // 0 = default; negative = hide gridlines
    GridLineColor: &pdf.Color{R: 220, G: 220, B: 220},
    Labels: &chart.AxisLabels{
        Enabled:  chart.Bool(true),
        Format:   "{value}%",         // "{value}" is replaced with the tick label
        FontName: "regular",
        FontSize: 8,
    },
    Visible: chart.Bool(true),
}
```

### chart.Series

```go
chart.Series{
    Name:  "Product A",                       // shown in the legend
    Data:  []float64{43, 55, 57, 60},         // y-values for line/area/column/bar/pie
    Color: &pdf.Color{R: 124, G: 181, B: 236}, // overrides palette assignment
}

// Rich data for scatter, bubble, heatmap, range charts, box plot, etc.
chart.Series{
    Name: "Measurements",
    Points: []chart.Point{
        {X: 1.5, Y: 23.4},
        {X: 2.3, Y: 17.8},
    },
}
```

### chart.Point

`Point` is a rich data point for chart types that require more than a single
Y value. Set only the fields meaningful for your chart type.

```go
chart.Point{
    X:    1.5,     // horizontal value (scatter, bubble, heatmap column index)
    Y:    23.4,    // primary value
    Z:    50,      // bubble radius source; heatmap cell value

    Low:    10.0,  // lower bound (range charts, box plot, error bar, dumbbell)
    Q1:     20.0,  // first quartile (box plot only)
    Median: 30.0,  // median (box plot only)
    Q3:     40.0,  // third quartile (box plot only)
    High:   55.0,  // upper bound (range charts, box plot, error bar, dumbbell)

    Target: 220,   // reference/target value (bullet chart)

    Name:  "Category label", // waterfall steps, funnel stages, treemap nodes
    Color: &pdf.Color{...},  // per-point color override

    IsSum:             true, // waterfall: show cumulative total (Y ignored)
    IsIntermediateSum: true, // waterfall: show running subtotal
}
```

### chart.Legend

```go
&chart.Legend{
    Enabled:       chart.Bool(true),
    Layout:        "horizontal",   // "horizontal" (default) or "vertical"
    Align:         "center",       // "left", "center" (default), "right"
    VerticalAlign: "bottom",       // "top", "middle", "bottom" (default)
    FontName:      "regular",
    FontSize:      8,
}
```

### chart.PlotOptions

`PlotOptions` contains per-chart-type rendering knobs. Only the field
corresponding to the chart type being rendered has any effect.

```go
opts.PlotOptions = &chart.PlotOptions{
    Line:        &chart.LineOptions{...},
    Area:        &chart.AreaOptions{...},
    Column:      &chart.ColumnOptions{...},
    Bar:         &chart.BarOptions{...},  // alias for ColumnOptions
    Pie:         &chart.PieOptions{...},
    Polar:       &chart.PolarOptions{...},
    Scatter:     &chart.ScatterOptions{...},
    Bubble:      &chart.BubbleOptions{...},
    Heatmap:     &chart.HeatmapOptions{...},
    Waterfall:   &chart.WaterfallOptions{...},
    Funnel:      &chart.FunnelOptions{...},
    Gauge:       &chart.GaugeOptions{...},
    Errorbar:    &chart.ErrorbarOptions{...},
    Boxplot:     &chart.BoxplotOptions{...},
    ColumnRange: &chart.ColumnRangeOptions{...},
    AreaRange:   &chart.AreaRangeOptions{...},
    Bullet:      &chart.BulletOptions{...},
    Dumbbell:    &chart.DumbbellOptions{...},
    Lollipop:    &chart.LollipopOptions{...},
    Treemap:     &chart.TreemapOptions{...},
}
```

Key fields for common chart types:

| Type | Key fields |
|------|-----------|
| `LineOptions` | `LineWidth` (default 2), `Marker`, `DataLabels` |
| `AreaOptions` | `LineWidth`, `FillAlpha` (0–1, default 0.3), `Marker`, `DataLabels` |
| `ColumnOptions` | `Stacking` (`""` grouped, `"normal"`, `"percent"`), `GroupPadding`, `PointPadding`, `BorderWidth`, `DataLabels` |
| `PieOptions` | `InnerSize` (`"50%"` for donut), `StartAngle` (degrees, default −90 = top), `DataLabels` |
| `PolarOptions` | `GridLineInterpolation` (`"polygon"` or `"circle"`), `FillAlpha`, `LineWidth`, `Marker`, `DataLabels` |
| `BubbleOptions` | `MinSize`, `MaxSize`, `ZMin`, `ZMax`, `DataLabels` |
| `HeatmapOptions` | `MinColor`, `MaxColor`, `BorderWidth`, `DataLabels` |
| `WaterfallOptions` | `UpColor`, `NegativeColor`, `LineWidth`, `DataLabels` |
| `FunnelOptions` | `NeckWidth`, `NeckHeight`, `Width`, `Reversed` (pyramid), `DataLabels` |
| `GaugeOptions` | `PaneStartAngle`, `PaneEndAngle`, `PlotBands`, `Solid` (solid-gauge), `DataLabels` |
| `BulletOptions` | `PlotBands`, `TargetWidth`, `TargetColor`, `DataLabels` |
| `TreemapOptions` | `ColorByPoint` (default true), `BorderWidth`, `BorderColor`, `DataLabels` |

### Helper functions

```go
// Float returns a *float64 — use it for optional float fields.
chart.Float(0.5)

// Bool returns a *bool — use it for optional bool fields.
chart.Bool(true)

// SeriesColor returns the color for series index i, cycling through the
// configured palette (opts.Colors) or DefaultColors when opts.Colors is nil.
chart.SeriesColor(opts, i)

// DefaultColors is the built-in 10-color Highcharts palette.
var chart.DefaultColors []pdf.Color
```

### DataLabels and Marker

```go
// DataLabels configures value labels rendered next to data points or bars.
&chart.DataLabels{
    Enabled:  chart.Bool(true),
    Format:   "{y}",      // "{y}" is replaced with the value; default "{y}"
    FontName: "regular",
    FontSize: 8,
    Color:    &pdf.Color{R: 50, G: 50, B: 50},
}

// Marker controls the symbol drawn at each data point on line/area/scatter charts.
&chart.Marker{
    Enabled: chart.Bool(true),
    Symbol:  "circle",   // "circle" (default), "square", "diamond"
    Radius:  3,          // radius in points
}
```

### GaugePlotBand

Colored arc bands used by both `GaugeOptions` and `BulletOptions`:

```go
chart.GaugePlotBand{
    From:      0,
    To:        80,
    Color:     pdf.Color{R: 85, G: 191, B: 59},  // green zone
    Thickness: 12,  // arc width in points; defaults to 10
}
```

### Complete example — line chart via layout engine

```go
package main

import (
    "log"

    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/chart"
    "github.com/gvanbeck/nautilus/pdf/chart/line"
    "github.com/gvanbeck/nautilus/pdf/layout"
)

func main() {
    doc, _ := pdf.New(pdf.Config{
        PageSize: pdf.PageSizeA4,
        Margins:  pdf.UniformMargins(40),
    })
    doc.RegisterFont("regular", "/path/to/font.ttf")
    doc.SetFont("regular", 11)

    opts := chart.Options{
        FontName: "regular",
        FontSize: 8,
        Title:    &chart.Title{Text: "Monthly Revenue"},
        XAxis:    &chart.Axis{Categories: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}},
        YAxis:    &chart.Axis{},
        Legend:   &chart.Legend{},
        Series: []chart.Series{
            {Name: "2023", Data: []float64{120, 150, 130, 180, 160, 200}},
            {Name: "2024", Data: []float64{140, 165, 175, 195, 210, 240}},
        },
    }

    lc := &line.LineChart{Options: opts}

    story := []layout.Flowable{
        chart.NewFlowable(lc, 0, 220),
    }

    frame := &layout.LayoutFrame{
        X: doc.ContentX(), Y: doc.ContentY(),
        Width: doc.ContentWidth(), Height: doc.ContentHeight(),
    }
    tmpl := &layout.PageTemplate{ID: "main", Frames: []*layout.LayoutFrame{frame}}

    dt := layout.NewDocTemplate(doc)
    dt.AddPageTemplate(tmpl)
    if err := dt.Build(story); err != nil {
        log.Fatal(err)
    }
    doc.Save("chart.pdf")
}
```

### Common chart examples

**Stacked column chart:**

```go
opts := chart.Options{
    FontName: "regular",
    FontSize: 8,
    XAxis:    &chart.Axis{Categories: []string{"Q1", "Q2", "Q3", "Q4"}},
    YAxis:    &chart.Axis{},
    Series: []chart.Series{
        {Name: "North", Data: []float64{43, 55, 57, 60}},
        {Name: "South", Data: []float64{23, 35, 41, 47}},
        {Name: "West",  Data: []float64{31, 28, 38, 44}},
    },
    PlotOptions: &chart.PlotOptions{
        Column: &chart.ColumnOptions{Stacking: "normal"},
    },
}
cc := &column.ColumnChart{Options: opts}
cc.Draw(doc, x, y, width, height)
```

**Donut chart:**

```go
opts := chart.Options{
    FontName: "regular",
    FontSize: 8,
    Series: []chart.Series{
        {Name: "Chrome",  Data: []float64{65}},
        {Name: "Firefox", Data: []float64{15}},
        {Name: "Safari",  Data: []float64{12}},
        {Name: "Other",   Data: []float64{8}},
    },
    PlotOptions: &chart.PlotOptions{
        Pie: &chart.PieOptions{
            InnerSize:  "50%",
            DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
        },
    },
}
pc := &pie.PieChart{Options: opts}
```

**Gauge with plot bands:**

```go
opts := chart.Options{
    FontName: "regular",
    FontSize: 8,
    YAxis:    &chart.Axis{Min: chart.Float(0), Max: chart.Float(200)},
    Series:   []chart.Series{{Name: "Speed km/h", Data: []float64{120}}},
    PlotOptions: &chart.PlotOptions{
        Gauge: &chart.GaugeOptions{
            PaneStartAngle: -150,
            PaneEndAngle:   150,
            PlotBands: []chart.GaugePlotBand{
                {From: 0,   To: 80,  Color: pdf.Color{R: 85,  G: 191, B: 59},  Thickness: 12},
                {From: 80,  To: 140, Color: pdf.Color{R: 221, G: 223, B: 13},  Thickness: 12},
                {From: 140, To: 200, Color: pdf.Color{R: 223, G: 83,  B: 83},  Thickness: 12},
            },
            DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
        },
    },
}
gc := &gauge.GaugeChart{Options: opts}
```

**Waterfall chart with sum flags:**

```go
opts := chart.Options{
    FontName: "regular",
    FontSize: 8,
    YAxis:    &chart.Axis{},
    Series: []chart.Series{{
        Points: []chart.Point{
            {Name: "Start",      Y: 120000},
            {Name: "Revenue",    Y: 569000},
            {Name: "Costs",      Y: -342000},
            {Name: "Subtotal",   IsIntermediateSum: true},
            {Name: "More costs", Y: -233000},
            {Name: "Balance",    IsSum: true},
        },
    }},
}
wc := &waterfall.WaterfallChart{Options: opts}
```

→ **Full reference:** [docs/chart-guide.en.md](docs/chart-guide.en.md)

## Examples

| Example | Description |
|---------|-------------|
| [`examples/basic`](examples/basic/main.go) | Multi-page demo covering fonts, Unicode, emoji, borders, frames, tables, headers/footers, and the two-pass Build mechanism. |
| [`examples/html`](examples/html/main.go) | Demonstrates `pdf/html`: inline HTML parsing, HTML table parsing and rendering with `WriteHTMLSpans` and `TableFromHTML`. |
| [`examples/layout`](examples/layout/main.go) | Multi-column, frame switching, `KeepTogether`, `CondPageBreak`, `HRFlowable`, and the `OnPage` decorator. |
| [`examples/rtl`](examples/rtl/main.go) | Arabic and Hebrew right-to-left text: contextual shaping, lam-alef ligatures, BiDi reordering, mixed RTL/LTR, and RTL inside a Frame. |
| [`examples/rml`](examples/rml/main.go) | XML-based document generation using the RML package; includes a full invoice template. |
| [`examples/celltag`](examples/celltag/main.go) | Generate table rows from Go structs using `cell` struct tags with `CellsFromStruct`. |
| [`examples/chart`](examples/chart/main.go) | All 20 chart types rendered via the layout engine across multiple pages. |

### Running the basic example

```sh
go run ./examples/basic \
    -font  /Library/Fonts/Lato-Medium.ttf \
    -bold  /Library/Fonts/Lato-Black.ttf \
    -emoji path/to/noto-emoji/png/128 \
    -out   output.pdf
```

### Running the HTML markup example

```sh
go run ./examples/html \
    -font   /Library/Fonts/Lato-Regular.ttf \
    -bold   /Library/Fonts/Lato-Bold.ttf \
    -italic /Library/Fonts/Lato-Italic.ttf \
    -out    output.pdf
```

### Running the layout example

```sh
go run ./examples/layout \
    -font /Library/Fonts/Lato-Medium.ttf \
    -bold /Library/Fonts/Lato-Black.ttf \
    -out  output.pdf
```

### Running the RTL example

```sh
go run ./examples/rtl \
    -arabic /System/Library/Fonts/Supplemental/DecoTypeNaskh.ttc \
    -hebrew /System/Library/Fonts/SFHebrew.ttf \
    -latin  /Library/Fonts/Lato-Regular.ttf \
    -out    output.pdf
```

### Running the chart example

```sh
go run ./examples/chart \
    -font /Library/Fonts/Lato-Medium.ttf \
    -bold /Library/Fonts/Lato-Black.ttf \
    -out  chart_output.pdf
```

## License

See [LICENSE](LICENSE) for details.
