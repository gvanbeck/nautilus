# Nautilus RML — User Guide

The `pdf/rml` package lets you describe PDF documents using an XML dialect
based on **RML (Report Markup Language)** by ReportLab.
Instead of writing Go code you describe the entire layout in an XML file,
which the library parses and converts into a PDF.

---

## Table of Contents

1. [Quick Start](#1-quick-start)
2. [Document Structure](#2-document-structure)
3. [Docinit — Registering Fonts](#3-docinit--registering-fonts)
4. [Template — Page Layout](#4-template--page-layout)
5. [PageGraphics — Headers and Footers](#5-pagegraphics--headers-and-footers)
6. [Stylesheet](#6-stylesheet)
   - [paraStyle](#61-parastyle)
   - [blockTableStyle](#62-blocktablestyle)
7. [Story — Content](#7-story--content)
   - [para](#71-para)
   - [spacer](#72-spacer)
   - [blockTable](#73-blocktable)
   - [image](#74-image)
   - [ul / ol](#75-ul--ol)
   - [indent](#76-indent)
   - [keepTogether](#77-keeptogether)
   - [condPageBreak](#78-condpagebreak)
   - [pageBreak / frameBreak](#79-pagebreak--framebreak)
   - [nextPageTemplate](#710-nextpagetemplate)
   - [hr / hRule](#711-hr--hrule)
8. [Units and Colors](#8-units-and-colors)
9. [Coordinate System](#9-coordinate-system)
10. [Go API](#10-go-api)
11. [Full Example](#11-full-example)

---

## 1. Quick Start

```bash
go run ./examples/rml \
  -rml     examples/rml/invoice.rml \
  -fontdir /Library/Fonts \
  -out     invoice.pdf
```

Or from Go code:

```go
import "github.com/gvanbeck/nautilus/pdf/rml"

doc, err := rml.ParseFile("invoice.rml", rml.Options{
    FontDir: "/Library/Fonts",
})
if err != nil {
    log.Fatal(err)
}
if err := doc.Save("invoice.pdf"); err != nil {
    log.Fatal(err)
}
```

---

## 2. Document Structure

An RML document always contains these four sections in order:

```xml
<?xml version="1.0" encoding="utf-8"?>
<document>
  <docinit>    <!-- fonts -->         </docinit>
  <template>   <!-- page layout -->   </template>
  <stylesheet> <!-- styles -->        </stylesheet>
  <story>      <!-- content -->       </story>
</document>
```

---

## 3. Docinit — Registering Fonts

### `<registerTTFont>`

Registers a TrueType font under an internal name.

```xml
<docinit>
  <registerTTFont fontName="regular"     fontFile="Lato-Regular.ttf"/>
  <registerTTFont fontName="regularBold" fontFile="Lato-Bold.ttf"/>
</docinit>
```

| Attribute  | Description                                              |
|------------|----------------------------------------------------------|
| `fontName` | Internal name used in styles                             |
| `fontFile` | File name (relative to `FontDir` option, or absolute)    |

### `<registerFontFamily>`

Groups four variants of a font family so that inline markup (`<b>`, `<i>`)
automatically selects the correct variant.

```xml
<docinit>
  <registerTTFont fontName="sans"          fontFile="Roboto-Regular.ttf"/>
  <registerTTFont fontName="sans-bold"     fontFile="Roboto-Bold.ttf"/>
  <registerTTFont fontName="sans-italic"   fontFile="Roboto-Italic.ttf"/>
  <registerTTFont fontName="sans-boldital" fontFile="Roboto-BoldItalic.ttf"/>

  <registerFontFamily name="sans"
                      fontName="sans"
                      bold="sans-bold"
                      italic="sans-italic"
                      boldItalic="sans-boldital"/>
</docinit>
```

| Attribute    | Description                            |
|--------------|----------------------------------------|
| `name`       | Family name (use in `fontName=`)       |
| `fontName`   | Regular variant                        |
| `bold`       | Bold variant                           |
| `italic`     | Italic variant                         |
| `boldItalic` | Bold-italic variant                    |

---

## 4. Template — Page Layout

```xml
<template pageSize="A4"
          leftMargin="55"  rightMargin="55"
          topMargin="65"   bottomMargin="62"
          title="Invoice 2026-0042"
          author="Nautilus Systems BV"
          subject="Invoice"
          creator="MyApp 1.0">

  <pageTemplate id="main">
    <pageGraphics>…</pageGraphics>   <!-- optional -->
    <frame id="body" x1="55" y1="62" width="485" height="720"/>
  </pageTemplate>

</template>
```

### `<template>` Attributes

| Attribute           | Default | Description                                                     |
|---------------------|---------|-----------------------------------------------------------------|
| `pageSize`          | `A4`    | `A3`, `A4`, `A5`, `letter`, `legal`, or `(width,height)` in pt |
| `leftMargin`        | `55`    | Left margin in pt                                               |
| `rightMargin`       | `55`    | Right margin in pt                                              |
| `topMargin`         | `55`    | Top margin in pt                                                |
| `bottomMargin`      | `55`    | Bottom margin in pt                                             |
| `title`             | —       | PDF metadata title                                              |
| `author`            | —       | PDF metadata author                                             |
| `subject`           | —       | PDF metadata subject                                            |
| `creator`           | —       | PDF metadata creator                                            |
| `firstPageTemplate` | first   | ID of the template to use for page 1                            |

### `<pageTemplate>`

| Attribute | Description                    |
|-----------|--------------------------------|
| `id`      | Unique name for this template  |

### `<frame>`

Defines a text area into which story content flows. Coordinates use
**PDF coordinates** (origin at bottom-left).

| Attribute | Description                      |
|-----------|----------------------------------|
| `id`      | Frame name                       |
| `x1`      | Left edge (from left of page)    |
| `y1`      | Bottom edge (from bottom of page)|
| `width`   | Frame width                      |
| `height`  | Frame height                     |

---

## 5. PageGraphics — Headers and Footers

`<pageGraphics>` contains drawing commands that run on **every page**
before the story content. Perfect for fixed headers and footers.

```xml
<pageGraphics>
  <saveState/>
  <setFont name="regular" size="8"/>
  <fill color="gray"/>

  <!-- Header line -->
  <lines>55 790 540 790</lines>
  <drawString x="55" y="793">Company Name BV</drawString>
  <drawRightString x="540" y="793">Invoice #2026-0042</drawRightString>

  <!-- Footer with page number -->
  <lines>55 52 540 52</lines>
  <drawCentredString x="297" y="41">Page %p</drawCentredString>

  <restoreState/>
</pageGraphics>
```

### Supported Drawing Commands

| Element               | Attributes                              | Description                              |
|-----------------------|-----------------------------------------|------------------------------------------|
| `<saveState/>`        | —                                       | Save graphics state                      |
| `<restoreState/>`     | —                                       | Restore graphics state                   |
| `<setFont/>`          | `name`, `size`                          | Set font and size                        |
| `<fill/>`             | `color`                                 | Set fill/text color                      |
| `<stroke/>`           | `color`, `width`                        | Set stroke color                         |
| `<drawString/>`       | `x`, `y`                                | Draw text left-aligned                   |
| `<drawRightString/>`  | `x`, `y`                                | Draw text right-aligned (x = right edge) |
| `<drawCentredString/>`| `x`, `y`                                | Draw text centered on x                  |
| `<lines/>`            | content: `x1 y1 x2 y2 …`               | Draw one or more line segments           |
| `<line/>`             | `x1`, `y1`, `x2`, `y2`, `width`, `color`| Single line segment                     |
| `<rect/>`             | `x`, `y`, `width`, `height`, `fill`, `stroke`, `round` | Rectangle        |
| `<circle/>`           | `x`, `y`, `radius`, `fill`, `stroke`   | Circle                                   |

### Page Number Variables

Inside text elements of `<pageGraphics>`:

| Variable | Content                       |
|----------|-------------------------------|
| `%p`     | Current page number           |
| `%P`     | Total page count *(not yet supported — see §11)* |

---

## 6. Stylesheet

### 6.1 `<paraStyle>`

Defines a reusable paragraph style.

```xml
<stylesheet>
  <paraStyle name="title"
             fontName="regularBold" fontSize="22"
             spaceAfter="6"/>

  <paraStyle name="body"
             fontName="regular" fontSize="11"
             leading="16" spaceAfter="5"/>

  <paraStyle name="small"
             fontName="regular" fontSize="9"
             textColor="gray" spaceAfter="3"/>
</stylesheet>
```

| Attribute        | Description                                            |
|------------------|--------------------------------------------------------|
| `name`           | Unique style name                                      |
| `parent`         | Inherit settings from another style                    |
| `fontName`       | Font name                                              |
| `fontSize`       | Font size in pt                                        |
| `leading`        | Line height in pt                                      |
| `alignment`      | `left`, `center`, `right`                              |
| `spaceBefore`    | Space before paragraph in pt                           |
| `spaceAfter`     | Space after paragraph in pt                            |
| `textColor`      | Text color (name or `#rrggbb`)                         |
| `backColor`      | Paragraph background color                             |
| `leftIndent`     | Left indent in pt                                      |
| `rightIndent`    | Right indent in pt                                     |
| `firstLineIndent`| Additional first-line indent in pt                     |
| `underline`      | `1` or `true` for underline                            |
| `strike`         | `1` or `true` for strikethrough                        |
| `keepWithNext`   | `1` or `true` to keep with following paragraph         |

### 6.2 `<blockTableStyle>`

Defines the formatting of a table.

```xml
<blockTableStyle id="invoice">
  <!-- Full grid in light gray, thick outer border in navy -->
  <lineStyle kind="GRID"    colorName="lightgray" thickness="0.4"
             start="0,0" stop="-1,-1"/>
  <lineStyle kind="OUTLINE" colorName="navy"      thickness="1.2"
             start="0,0" stop="-1,-1"/>

  <!-- Header row: navy background, white bold text -->
  <blockBackground colorName="navy"  start="0,0" stop="-1,0"/>
  <blockTextColor  colorName="white" start="0,0" stop="-1,0"/>
  <blockFont       name="regularBold" start="0,0" stop="-1,0"/>

  <!-- Last row: light-blue background, bold -->
  <blockBackground colorName="#e8f0fe" start="0,-1" stop="-1,-1"/>
  <blockFont       name="regularBold"  start="0,-1" stop="-1,-1"/>

  <!-- Right-align numeric columns (cols 3–4) -->
  <blockAlignment  value="right" start="3,0" stop="4,-1"/>

  <!-- Vertical centering in all cells -->
  <blockValign     value="middle" start="0,0" stop="-1,-1"/>

  <!-- Padding -->
  <blockPadding      value="4" start="0,0" stop="-1,-1"/>
  <blockLeftPadding  value="6" start="0,0" stop="-1,-1"/>
</blockTableStyle>
```

#### `<lineStyle>`

| Attribute   | Values                            | Description                               |
|-------------|-----------------------------------|-------------------------------------------|
| `kind`      | `GRID`, `OUTLINE`, `INNERGRID`, `BOX` | Line type                             |
| `colorName` | color value                       | Line color                                |
| `thickness` | number in pt                      | Line thickness (default `0.5`)            |
| `start`     | `col,row`                         | Start cell (inclusive)                    |
| `stop`      | `col,row`                         | End cell (inclusive, negative = from end) |

#### Cell Styling Commands

| Element              | Relevant attributes           | Description                    |
|----------------------|-------------------------------|--------------------------------|
| `<blockBackground>`  | `colorName`, `start`, `stop`  | Cell background color          |
| `<blockTextColor>`   | `colorName`, `start`, `stop`  | Cell text color                |
| `<blockFont>`        | `name`, `size`, `start`, `stop` | Cell font                    |
| `<blockAlignment>`   | `value`, `start`, `stop`      | Horizontal alignment           |
| `<blockValign>`      | `value`, `start`, `stop`      | Vertical alignment             |
| `<blockPadding>`     | `value`, `start`, `stop`      | Uniform padding                |
| `<blockTopPadding>`  | `value`, `start`, `stop`      | Top padding                    |
| `<blockRightPadding>`| `value`, `start`, `stop`      | Right padding                  |
| `<blockBottomPadding>`| `value`, `start`, `stop`     | Bottom padding                 |
| `<blockLeftPadding>` | `value`, `start`, `stop`      | Left padding                   |

#### Cell Range with Negative Indexing

`start="0,0" stop="-1,-1"` applies to **all** cells.
`start="0,0" stop="-1,0"` applies to the **header row** (row 0).
`start="0,-1" stop="-1,-1"` applies to the **last row**.

---

## 7. Story — Content

### 7.1 `<para>`

A paragraph with optional inline markup.

```xml
<para style="body">Plain text without markup.</para>
<para style="body">Text with <b>bold</b> and <i>italic</i>.</para>
<para style="body">Text with <u>underline</u>.</para>
```

Supported inline tags: `<b>`, `<i>`, `<u>`.
When a paragraph contains inline markup and a `<registerFontFamily>` is declared,
bold/italic variants are selected automatically.

The shorthand `h1`–`h6` is available as an alternative to `<para style="…">`:

```xml
<h1>Chapter 1</h1>
<h2>Section</h2>
```

### 7.2 `<spacer>`

Adds vertical (or horizontal) whitespace.

```xml
<spacer length="18"/>        <!-- 18 pt vertical -->
<spacer length="10" width="0"/>
```

### 7.3 `<blockTable>`

A table that can automatically flow across pages.

```xml
<blockTable colWidths="55,220,40,80,90"
            rowHeights="24,22,22,22"
            style="invoice"
            repeatRows="1"
            align="left"
            spaceAfter="10">

  <!-- Header row -->
  <tr height="24">
    <td>Ref.</td>
    <td>Description</td>
    <td>Qty</td>
    <td>Unit Price</td>
    <td>Subtotal</td>
  </tr>

  <!-- Data rows with row background -->
  <tr>
    <td>A-001</td>
    <td>Mechanical keyboard</td>
    <td>2</td>
    <td>€ 89.95</td>
    <td>€ 179.90</td>
  </tr>
  <tr bg="#eef4ff">
    <td>A-002</td>
    <td>27" 4K IPS monitor</td>
    <td>1</td>
    <td>€ 549.00</td>
    <td>€ 549.00</td>
  </tr>

</blockTable>
```

#### `<blockTable>` Attributes

| Attribute    | Description                                                       |
|--------------|-------------------------------------------------------------------|
| `colWidths`  | Comma-separated column widths in pt (required)                    |
| `rowHeights` | Comma-separated row heights in pt (optional)                      |
| `style`      | Reference to a `blockTableStyle` id                               |
| `repeatRows` | Number of leading rows to repeat after page overflow              |
| `align`      | Horizontal table alignment: `left`, `center`, `right`             |
| `spaceBefore`| Space before table in pt                                          |
| `spaceAfter` | Space after table in pt                                           |

#### `<tr>` Attributes

| Attribute | Description                                   |
|-----------|-----------------------------------------------|
| `height`  | Fixed row height (overrides `rowHeights`)      |
| `bg`      | Background color for the entire row            |

#### `<td>` Attributes

| Attribute      | Description                                         |
|----------------|-----------------------------------------------------|
| `colspan`      | Number of columns to span (default 1)               |
| `rowspan`      | Number of rows to span (default 1)                  |
| `style`        | Paragraph style for cell content                    |
| `fontName`     | Override font                                       |
| `fontSize`     | Override font size                                  |
| `bold`         | `1` or `true` for bold                              |
| `bg`           | Cell background color                               |
| `textColor`    | Cell text color                                     |
| `halign`       | `left`, `center`, `right`                           |
| `valign`       | `top`, `middle`, `bottom`                           |
| `topPadding`   | Top padding in pt                                   |
| `rightPadding` | Right padding in pt                                 |
| `bottomPadding`| Bottom padding in pt                                |
| `leftPadding`  | Left padding in pt                                  |

### 7.4 `<image>`

Inserts an image.

```xml
<image file="logo.png" width="120" height="60"
       align="right" spaceAfter="10"/>
```

| Attribute    | Description                                       |
|--------------|---------------------------------------------------|
| `file`       | Path to the image (JPEG or PNG)                   |
| `width`      | Width in pt (optional, scales proportionally)     |
| `height`     | Height in pt (optional)                           |
| `align`      | `left`, `center`, `right`                         |
| `spaceBefore`| Space before in pt                                |
| `spaceAfter` | Space after in pt                                 |

### 7.5 `<ul>` / `<ol>`

Unordered or ordered list.

```xml
<ul bulletIndent="12">
  <li>First item</li>
  <li style="body">Second item</li>
</ul>

<ol start="3">
  <li>Third item</li>
  <li>Fourth item</li>
</ol>
```

#### `<ul>` / `<ol>` Attributes

| Attribute      | Description                                        |
|----------------|----------------------------------------------------|
| `bulletIndent` | Indent of the bullet/number in pt                  |
| `style`        | Default paragraph style for all list items         |
| `start`        | Starting number for ordered lists (default `1`)    |

#### `<li>` Attributes

| Attribute | Description                              |
|-----------|------------------------------------------|
| `style`   | Paragraph style (overrides list style)   |

### 7.6 `<indent>`

Wraps child content in an indent.

```xml
<indent left="30" right="15">
  <para style="body">Indented text.</para>
  <blockTable …>…</blockTable>
</indent>
```

| Attribute | Description             |
|-----------|-------------------------|
| `left`    | Left indent in pt       |
| `right`   | Right indent in pt      |

### 7.7 `<keepTogether>`

Keeps all child flowables on the same page.

```xml
<keepTogether maxHeight="200">
  <para style="h2">Section Title</para>
  <para style="body">First paragraph of the section.</para>
</keepTogether>
```

| Attribute   | Description                                                              |
|-------------|--------------------------------------------------------------------------|
| `maxHeight` | Maximum height in pt (default unlimited). If the group exceeds this, it will be split anyway. |

### 7.8 `<condPageBreak>`

Inserts a page break **if** the remaining space is less than the given height.

```xml
<condPageBreak height="150"/>
```

| Attribute | Description                     |
|-----------|---------------------------------|
| `height`  | Minimum required height in pt   |

### 7.9 `<pageBreak>` / `<frameBreak>`

```xml
<pageBreak/>    <!-- Hard page break -->
<frameBreak/>   <!-- Advance to next frame -->
```

### 7.10 `<nextPageTemplate>`

Switches to a different page template starting from the next page.

```xml
<nextPageTemplate id="twocolumn"/>
```

| Attribute | Description                     |
|-----------|---------------------------------|
| `id`      | ID of the target page template  |

### 7.11 `<hr>` / `<hRule>`

Horizontal rule.

```xml
<hr width="100%" thickness="0.5" colorName="lightgray"/>
```

| Attribute   | Description                                    |
|-------------|------------------------------------------------|
| `width`     | Width in pt or `%` of available width          |
| `thickness` | Line thickness in pt (default `0.5`)           |
| `colorName` | Color (default black)                          |

---

## 8. Units and Colors

### Units

All numeric values are interpreted as **points (pt)** by default.
The following units are also recognized:

| Input      | Example    | Conversion        |
|------------|------------|-------------------|
| `pt`       | `12pt`     | 1 pt = 1/72 inch  |
| `cm`       | `2.1cm`    | 1 cm = 28.35 pt   |
| `mm`       | `21mm`     | 1 mm = 2.835 pt   |
| `in`       | `0.83in`   | 1 in = 72 pt      |
| *(number)* | `595`      | Interpreted as pt |

### Colors

Colors can be specified as:

| Format        | Example             | Description                              |
|---------------|---------------------|------------------------------------------|
| Name          | `navy`, `lightgray` | Well-known CSS color names               |
| Hex           | `#e8f0fe`           | RGB hexadecimal                          |
| Comma         | `220,220,220`       | Three decimal values (0–255)             |

Supported color names (selection): `black`, `white`, `red`, `green`, `blue`,
`navy`, `gray` / `grey`, `lightgray`, `darkgray`, `yellow`, `orange`, `purple`,
`cyan`, `magenta`, `pink`, `brown`, `lime`, `teal`, `silver`, `gold`,
`transparent` / `none`.

---

## 9. Coordinate System

RML uses **PDF coordinates**: the origin is at the **bottom-left** of the page.
The Y-axis runs upward.

```
(0, 842)  ──────────────── (595, 842)   ← top of A4
          │                │
          │   page content │
          │                │
(0,   0)  ──────────────── (595,   0)   ← bottom of A4
```

This applies to `<frame>`, `<drawString>`, `<lines>`, `<rect>`, etc.

The library converts internally to the Nautilus coordinate system
(origin at top-left).

**Guidelines for A4 (595 × 842 pt):**

- Header text at y ≈ 790–800 (near the top)
- Footer text at y ≈ 30–55 (near the bottom)
- Text frame: `y1=62, height=720` leaves room for header and footer

---

## 10. Go API

```go
package rml

// Options holds configuration parameters for parsing.
type Options struct {
    FontDir string // directory containing font files
}

// ParseFile reads an RML file and returns a pdf.Document.
func ParseFile(path string, opts Options) (*pdf.Document, error)

// Parse reads an RML stream and returns a pdf.Document.
func Parse(r io.Reader, opts Options) (*pdf.Document, error)
```

### Example CLI Tool

```go
package main

import (
    "flag"
    "log"

    "github.com/gvanbeck/nautilus/pdf/rml"
)

func main() {
    rmlFile := flag.String("rml",     "document.rml", "RML input file")
    fontDir := flag.String("fontdir", ".",             "Font directory")
    outFile := flag.String("out",     "document.pdf",  "PDF output file")
    flag.Parse()

    doc, err := rml.ParseFile(*rmlFile, rml.Options{FontDir: *fontDir})
    if err != nil {
        log.Fatalf("parse: %v", err)
    }
    if err := doc.Save(*outFile); err != nil {
        log.Fatalf("save: %v", err)
    }
}
```

---

## 11. Full Example

See `examples/rml/invoice.rml` for a complete invoice example featuring:

- `<docinit>` with two font registrations
- `<template>` with metadata, margins, and a frame definition
- `<pageGraphics>` with a header (company name + invoice number), separator
  lines and a footer with page number (`%p`)
- `<stylesheet>` with multiple paragraph styles and two table styles
  (`invoice` with navy header row, and `meta` for the metadata table)
- `<story>` with letterhead, a metadata table, an invoice-line table
  (with alternating row background colors), and a summary table

```bash
go run ./examples/rml \
  -rml     examples/rml/invoice.rml \
  -fontdir /Library/Fonts \
  -out     invoice.pdf
```

---

## Known Limitations

The following RML features are currently **not supported**:

| Feature               | Description                                          |
|-----------------------|------------------------------------------------------|
| `%P` in pageGraphics  | Total page count (requires two-pass rendering)       |
| `<storyPlace>`        | Placing content at fixed coordinates                 |
| `<balancedColumns>`   | Automatic multi-column text balancing                |
| Barcodes              | `<barCode>`, QR codes, etc.                          |
| Form fields           | Interactive PDF form elements                        |
| Cross-references      | `<seq>`, `<bookmark>`, `<getName>`, etc.             |
| `<illustration>`      | Inline SVG/vector drawings                           |
| `<plugInFlowable>`    | External flowable plug-ins                           |
