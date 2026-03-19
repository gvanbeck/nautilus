# Nautilus Layout System — Developer Guide

The layout package (`pdf/layout`) is a high-level layout engine built on top of the Nautilus
`pdf.Document` drawing layer, inspired by the Platypus system from the Python
ReportLab library.

Instead of calculating coordinates manually, you describe your document as a
**Story**: a flat list of `Flowable` objects. The engine — `DocTemplate` —
automatically distributes that content across frames and pages.

---

## Table of Contents

1. [Core concepts](#1-core-concepts)
2. [Units and coordinates](#2-units-and-coordinates)
3. [Minimal example](#3-minimal-example)
4. [Document and fonts](#4-document-and-fonts)
5. [Paragraph](#5-paragraph)
6. [Spacer](#6-spacer)
7. [HRFlowable](#7-hrflowable)
8. [KeepTogether](#8-keeptogether)
9. [Action flowables](#9-action-flowables)
10. [LayoutFrame](#10-layoutframe)
11. [PageTemplate](#11-pagetemplate)
12. [DocTemplate and Build](#12-doctemplate-and-build)
13. [Page decorators (headers and footers)](#13-page-decorators-headers-and-footers)
14. [Template switching](#14-template-switching)
15. [Multi-column layout](#15-multi-column-layout)
16. [Debugging with ShowBoundary](#16-debugging-with-showboundary)
17. [Implementing a custom Flowable](#17-implementing-a-custom-flowable)
18. [Frequently asked questions](#18-frequently-asked-questions)

---

## 1. Core concepts

The system consists of four layers stacked on top of each other:

```
Story  ([]Flowable)
   ↓  consumed by
DocTemplate    —  manages frames and pages
   ↓  manages
PageTemplate   —  page geometry: ordered LayoutFrames + decorators
   ↓  contains
LayoutFrame    —  rectangular region with a downward Y cursor
   ↓  draws into
pdf.Document   —  the underlying PDF drawing layer
```

### Story

A `Story` is simply a `[]layout.Flowable`. You build the complete document
content as a flat list and pass it to `DocTemplate.Build`.

```go
story := []layout.Flowable{
    &layout.Paragraph{Text: "Title", Style: titleStyle},
    &layout.Spacer{Height: 12},
    &layout.Paragraph{Text: "First paragraph...", Style: bodyStyle},
    &layout.PageBreak{},
    &layout.Paragraph{Text: "Starts on a new page.", Style: bodyStyle},
}
```

### Flowable

A `Flowable` is anything that can be measured and drawn. The interface:

```go
type Flowable interface {
    Wrap(doc *pdf.Document, availWidth, availHeight float64) (width, height float64)
    Draw(doc *pdf.Document, x, y float64) error
    Split(doc *pdf.Document, availWidth, availHeight float64) []Flowable
    SpaceBefore() float64
    SpaceAfter()  float64
    KeepWithNext() bool
    MinWidth() float64
}
```

The engine always calls `Wrap` first (for measurement), then `Draw` (for rendering).
`Split` is only called when a flowable does not fit in the remaining space.

### LayoutFrame

A `LayoutFrame` is a rectangular region on the page with an internal
Y cursor that moves downward as content is added. When the cursor reaches
the bottom, the frame is full and the engine switches to the next frame
or a new page.

### PageTemplate

A `PageTemplate` associates a name with an ordered list of `LayoutFrame` objects
plus optional callbacks for headers and footers. Multiple templates enable
alternating page layouts (title page, single column, two columns, etc.).

### DocTemplate

`DocTemplate` is the engine. It processes the story step by step, places
flowables in frames, switches frames, starts new pages, and executes
template switches.

---

## 2. Units and coordinates

All dimensions are in **points** (pt). 1 point = 1/72 inch ≈ 0.353 mm.

Common conversions:

| Unit    | Points   |
|---------|----------|
| 1 cm    | 28.35 pt |
| 1 inch  | 72 pt    |
| 10 mm   | 28.35 pt |

The coordinate system has its origin at the **top-left** of the page.
The Y axis runs **downward** (consistent with the rest of the Nautilus library).

Standard page sizes:

```go
pdf.PageSizeA3     // 841.89 × 1190.55 pt
pdf.PageSizeA4     // 595.28 × 841.89 pt
pdf.PageSizeA5     // 419.53 × 595.28 pt
pdf.PageSizeLetter // 612 × 792 pt
pdf.PageSizeLegal  // 612 × 1008 pt
```

---

## 3. Minimal example

```go
package main

import (
    "log"
    "github.com/gvanbeck/nautilus/pdf"
    "github.com/gvanbeck/nautilus/pdf/layout"
)

func main() {
    // 1. Create a document.
    doc, err := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
    if err != nil {
        log.Fatal(err)
    }

    // 2. Register and activate a font.
    if err := doc.RegisterFont("regular", "/path/to/NotoSans-Regular.ttf"); err != nil {
        log.Fatal(err)
    }
    doc.SetFont("regular", 12) // set default font before Build

    // 3. Define a style and build the story.
    body := layout.ParagraphStyle{FontName: "regular", FontSize: 12, SpaceAfter: 8}
    story := []layout.Flowable{
        &layout.Paragraph{Text: "Hello, Nautilus!", Style: body},
        &layout.Spacer{Height: 12},
        &layout.Paragraph{Text: "Second paragraph.", Style: body},
    }

    // 4. Define the frame: the full content area of the page.
    const margin = 50.0
    frame := &layout.LayoutFrame{
        X:      margin,
        Y:      margin,
        Width:  doc.PageWidth() - 2*margin,
        Height: doc.PageHeight() - 2*margin,
    }

    // 5. Attach the frame to a PageTemplate.
    tmpl := &layout.PageTemplate{
        ID:     "main",
        Frames: []*layout.LayoutFrame{frame},
    }

    // 6. Create the DocTemplate, register the template, and build.
    dt := layout.NewDocTemplate(doc)
    dt.AddPageTemplate(tmpl)
    if err := dt.Build(story); err != nil {
        log.Fatal(err)
    }

    // 7. Save the PDF.
    if err := doc.Save("output.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

> **Note:** `doc.SetFont` must be called before `Build` so that text measurement
> works from the very first paragraph.

---

## 4. Document and fonts

```go
doc, err := pdf.New(pdf.Config{
    PageSize:         pdf.PageSizeA4,
    DefaultFontSize:  12,
    LineHeightFactor: 1.4,
})
```

Registering and switching fonts:

```go
doc.RegisterFont("regular", "NotoSans-Regular.ttf")
doc.RegisterFont("bold",    "NotoSans-Bold.ttf")
doc.RegisterFont("italic",  "NotoSans-Italic.ttf")

doc.SetFont("regular", 12) // activate before Build
```

In a `ParagraphStyle`, the `FontName` field automatically switches to the correct
font during rendering. If `FontName` is empty, the paragraph uses whichever font
is currently active on the document.

---

## 5. Paragraph

`Paragraph` is the primary text element. It supports word wrapping,
font switching, color, indentation, and horizontal alignment.

### Struct definition

```go
type Paragraph struct {
    Text  string         // text to render; use \n for explicit line breaks
    Style ParagraphStyle // visual formatting
}
```

### ParagraphStyle

```go
type ParagraphStyle struct {
    FontName         string     // registered font name; empty = current font
    FontSize         float64    // size in points; 0 = 12 pt
    Leading          float64    // line spacing in points; 0 = FontSize × 1.2
    Alignment        HAlign     // AlignLeft (default), AlignCenter, AlignRight
    SpaceBefore      float64    // extra whitespace above the paragraph
    SpaceAfter       float64    // extra whitespace below the paragraph
    KeepWithNextPara bool       // no frame/page break after this paragraph
    LeftIndent       float64    // left indentation in points
    RightIndent      float64    // right indentation in points
    TextColor        *pdf.Color // text color; nil = current document color
}
```

### Alignment

```go
const (
    AlignLeft   HAlign = iota // default
    AlignCenter
    AlignRight
)
```

### Examples

**Basic paragraph:**
```go
body := layout.ParagraphStyle{
    FontName:   "regular",
    FontSize:   11,
    SpaceAfter: 6,
}
p := &layout.Paragraph{Text: "The quick brown fox jumps over the lazy dog.", Style: body}
```

**Centered subtitle:**
```go
subtitle := layout.ParagraphStyle{
    FontName:   "regular",
    FontSize:   14,
    Alignment:  layout.AlignCenter,
    SpaceAfter: 16,
}
```

**Colored and indented:**
```go
navy := pdf.ColorNavy
callout := layout.ParagraphStyle{
    FontName:    "bold",
    FontSize:    10,
    LeftIndent:  20,
    RightIndent: 20,
    SpaceBefore: 8,
    SpaceAfter:  8,
    TextColor:   &navy,
}
```

**Custom line spacing:**
```go
spacious := layout.ParagraphStyle{
    FontName: "regular",
    FontSize: 11,
    Leading:  18, // fixed 18 pt line spacing instead of 11 × 1.2 = 13.2 pt
}
```

**Explicit line breaks with \n:**
```go
&layout.Paragraph{
    Text: "Line one\nLine two\nLine three",
    Style: body,
}
```

### Splitting across frames

Long paragraphs are automatically split when they do not fit in the remaining
frame space. The first part is drawn in the current frame; the remainder is
carried over to the next frame. No action is required on your part.

---

## 6. Spacer

`Spacer` reserves a fixed amount of vertical space without drawing anything.

```go
type Spacer struct {
    Width  float64 // 0 or negative = full available width
    Height float64 // height to reserve in points
}
```

**Examples:**

```go
// 12 points of whitespace
&layout.Spacer{Height: 12}

// Half inch (36 pt) of whitespace
&layout.Spacer{Height: 36}
```

> **Tip:** Use `SpaceBefore` and `SpaceAfter` in `ParagraphStyle` for
> automatic spacing around paragraphs. Use `Spacer` for one-off,
> explicit whitespace.

---

## 7. HRFlowable

`HRFlowable` draws a horizontal line as a filled bar.

```go
type HRFlowable struct {
    Width     float64   // width: > 1.0 = absolute points;
                        //        0..1.0 = fraction of available width (e.g. 0.8 = 80%)
                        //        0 = full available width
    Thickness float64   // bar height in points; default 1
    Color     pdf.Color // fill color
    Align     HAlign    // alignment when Width < available width
    Before    float64   // whitespace above the line
    After     float64   // whitespace below the line
}
```

**Examples:**

```go
// Thin gray line spanning the full width
&layout.HRFlowable{
    Thickness: 0.75,
    Color:     pdf.ColorLightGray,
    Before:    8,
    After:     8,
}

// Thick navy line, 60% width, centered
&layout.HRFlowable{
    Width:     0.6,
    Thickness: 2,
    Color:     pdf.ColorNavy,
    Align:     layout.AlignCenter,
    Before:    12,
    After:     12,
}

// Red line with absolute width of 100 points, right-aligned
&layout.HRFlowable{
    Width:     100,
    Thickness: 1,
    Color:     pdf.ColorRed,
    Align:     layout.AlignRight,
}
```

---

## 8. KeepTogether

`KeepTogether` prevents a group of flowables from being split across frames or pages.

```go
type KeepTogether struct {
    Flowables []Flowable // elements to keep together
}
```

**Behaviour:**

1. Group fits in remaining frame space → draw immediately.
2. Does not fit → engine inserts a `FrameBreak` and retries in the next frame.
3. Does not even fit in an empty frame → engine splits the individual flowables
   separately (to avoid infinite loops).

**Typical use — always keep a heading with its first paragraph:**

```go
h1 := layout.ParagraphStyle{FontName: "bold", FontSize: 14, SpaceAfter: 4}
body := layout.ParagraphStyle{FontName: "regular", FontSize: 11, SpaceAfter: 6}

story = append(story, &layout.KeepTogether{
    Flowables: []layout.Flowable{
        &layout.Paragraph{Text: "Chapter 3 — Results", Style: h1},
        &layout.Paragraph{Text: "In this chapter we discuss...", Style: body},
    },
})
```

**Alternative via KeepWithNextPara:**

When a paragraph style has `KeepWithNextPara: true`, the engine automatically
groups that paragraph together with the next one. This is convenient for
heading styles:

```go
h1Style := layout.ParagraphStyle{
    FontName:         "bold",
    FontSize:         14,
    KeepWithNextPara: true, // never separated from the first paragraph below it
}
```

> **Note:** `KeepWithNextPara` only binds the paragraph to the *immediately
> following* flowable. For longer groups use `KeepTogether` explicitly.

---

## 9. Action flowables

Action flowables are invisible, zero-height elements that direct the engine
rather than drawing visible content. They are inserted into the story list
as control signals.

### PageBreak

Forces an immediate page break.

```go
type PageBreak struct {
    NextTemplate string // optional template ID for the new page
}
```

```go
// Simple page break
&layout.PageBreak{}

// Page break and immediately switch to the "two-column" template
&layout.PageBreak{NextTemplate: "two-column"}
```

### FrameBreak

Advances to the next frame in the current template (or to a new page if
the current frame is the last one).

```go
&layout.FrameBreak{}
```

Use this to push content explicitly to the second column:

```go
story = append(story,
    leftColumnContent...,
    &layout.FrameBreak{}, // jump to right column
    rightColumnContent...,
)
```

### CondPageBreak

Inserts a page break *only* if fewer than `MinHeight` points remain in the
current frame. Useful to prevent a section from starting on a very small
remaining space.

```go
type CondPageBreak struct {
    MinHeight float64
}
```

```go
// Page break if less than 72 pt (1 inch) remains
&layout.CondPageBreak{MinHeight: 72}

// Page break if fewer than 3 lines of 14 pt remain
&layout.CondPageBreak{MinHeight: 3 * 14 * 1.2}
```

**Good placement:** just before a section heading or a group that should
start on a "clean" new page.

### NextPageTemplate

Schedules a template switch that takes effect at the *next* page break.
The current page continues to use the current template.

```go
type NextPageTemplate struct {
    TemplateID string
}
```

```go
// Switch to two columns on the next page
&layout.NextPageTemplate{TemplateID: "two-column"},
&layout.PageBreak{},
```

The difference from `PageBreak{NextTemplate: "..."}`:

| | `NextPageTemplate` + `PageBreak{}` | `PageBreak{NextTemplate: "..."}` |
|---|---|---|
| When does it switch? | At the next `PageBreak` | Immediately at the `PageBreak` |
| Current page | remains unchanged | already ends |
| Use | Always safe | When you want to switch immediately |

In practice both are equivalent when placed immediately after each other.

---

## 10. LayoutFrame

`LayoutFrame` defines a rectangular content area on the page.

```go
type LayoutFrame struct {
    X, Y          float64     // top-left coordinate on the page (points)
    Width, Height float64     // outer dimensions (points)
    Padding       pdf.Padding // inner whitespace
    ID            string      // optional name (for debugging)
    ShowBoundary  bool        // draw an outline around the frame (debugging)
}
```

### Padding

```go
frame := &layout.LayoutFrame{
    X: 50, Y: 80,
    Width: 495, Height: 700,
    Padding: pdf.Padding{Top: 8, Right: 12, Bottom: 8, Left: 12},
}

// Helpers:
frame.Padding = pdf.UniformPadding(10)       // 10 pt on all sides
frame.Padding = pdf.HorizontalPadding(12, 8) // 12 pt left/right, 8 pt top/bottom
```

### Available width

The *inner* width available for flowables to render into:

```
inner width = Width − Padding.Left − Padding.Right
```

Flowables receive this value as `availWidth` in their `Wrap` call.

### Useful sizing calculation

Use the document dimensions as a starting point:

```go
const (
    margin  = 50.0
    headerH = 40.0
    footerH = 36.0
)
frame := &layout.LayoutFrame{
    X:      margin,
    Y:      margin + headerH,
    Width:  doc.PageWidth() - 2*margin,
    Height: doc.PageHeight() - margin - headerH - footerH,
}
```

### ShowBoundary

Set `ShowBoundary: true` during development to make frame outlines visible as
thin gray rectangles. Remove this in production.

```go
frame := &layout.LayoutFrame{
    X: 50, Y: 80, Width: 495, Height: 700,
    ShowBoundary: true, // debugging only
}
```

---

## 11. PageTemplate

`PageTemplate` associates a name with an ordered list of frames plus
optional page decorators.

```go
type PageTemplate struct {
    ID               string          // unique name, used by NextPageTemplate
    Frames           []*LayoutFrame  // frames in fill order
    OnPage           PageDecorator   // callback after AddPage (headers, watermarks)
    OnPageEnd        PageDecorator   // callback before page finalisation (footers)
    AutoNextTemplate string          // template ID after this page; empty = same template
}
```

### AutoNextTemplate

With `AutoNextTemplate` you do not need to place `NextPageTemplate` actions
in your story when you want all pages of the same type:

```go
singleTemplate := &layout.PageTemplate{
    ID:               "single",
    Frames:           []*layout.LayoutFrame{singleFrame},
    OnPage:           headerFooter,
    AutoNextTemplate: "single", // every subsequent page also uses "single"
}
```

Without `AutoNextTemplate` the engine always uses the current template until
you explicitly switch it.

### Registering multiple templates

The first registered template is used for page 1.

```go
dt := layout.NewDocTemplate(doc)
dt.AddPageTemplate(titleTemplate)   // page 1
dt.AddPageTemplate(bodyTemplate)    // activated after a switch
dt.AddPageTemplate(twoColTemplate)
```

---

## 12. DocTemplate and Build

```go
dt := layout.NewDocTemplate(doc)
dt.AddPageTemplate(tmpl)            // register at least one template
err := dt.Build(story)              // process the story
```

### What Build does

```
startPage()
  → pageNum++
  → doc.AddPage()
  → reset all frames in the current template
  → call OnPage decorator

Main loop (for each element in the story):
  1. ActionFlowable? → execute immediately, continue
  2. KeepWithNext=true? → collect chain, wrap in KeepTogether
  3. frame.add(flowable) → does it fit?
     YES → draw, move cursor down, continue
     NO  → try frame.split(flowable)
           Split succeeded? → re-queue parts
           Split failed?    → advanceFrame() (to next frame or new page)

endPage()
  → call OnPageEnd decorator
  → process pending template switch
```

### Error handling

`Build` returns an error if:
- no templates are registered.
- a flowable fails to fit in a frame more than 10 consecutive times (infinite
  loop or content that is too large).

```go
if err := dt.Build(story); err != nil {
    log.Fatalf("build failed: %v", err)
}
```

### After Build

The `pdf.Document` contains all pages. Save as usual:

```go
doc.Save("output.pdf")
// or to an io.Writer:
doc.Output(os.Stdout)
```

---

## 13. Page decorators (headers and footers)

`PageDecorator` is a function that the engine calls at the beginning or end
of each page:

```go
type PageDecorator func(doc *pdf.Document, pageNum int)
```

`pageNum` is 1-based. You can draw anything that `pdf.Document` supports:
text, borders, lines, images.

### Example

```go
const (
    margin  = 50.0
    headerH = 40.0
    footerH = 36.0
)
pageW := doc.PageWidth()
pageH := doc.PageHeight()
contentW := pageW - 2*margin

pageDecorator := func(d *pdf.Document, pageNum int) {
    // ── Header ────────────────────────────────────────────────────────────
    spec := &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}
    d.DrawBorder(margin, margin, contentW, headerH-4, pdf.Border{Bottom: spec})

    d.SetFont("regular", 8)
    d.SetTextColor(100, 100, 100)
    d.WriteLine("My Document", margin, margin+10)

    num := fmt.Sprintf("Page %d", pageNum)
    w, _ := d.MeasureText(num)
    d.WriteLine(num, pageW-margin-w, margin+10)

    // ── Footer ────────────────────────────────────────────────────────────
    footerY := pageH - footerH
    d.DrawBorder(margin, footerY, contentW, 0, pdf.Border{
        Top: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
    })
    d.SetFont("regular", 8)
    d.SetTextColor(150, 150, 150)
    d.WriteLine("Confidential", margin, footerY+8)
}

tmpl := &layout.PageTemplate{
    ID:     "main",
    Frames: []*layout.LayoutFrame{frame},
    OnPage: pageDecorator, // called at the start of each page
}
```

### Skipping the first page

```go
pageDecorator := func(d *pdf.Document, pageNum int) {
    if pageNum == 1 {
        return // no header on the title page
    }
    // ... rest of header
}
```

### Reserving space for header and footer

The decorator draws *outside* the `LayoutFrame` areas. Make sure your frames
leave sufficient margin for the header and footer:

```go
frame := &layout.LayoutFrame{
    X:      margin,
    Y:      margin + headerH, // below the header
    Width:  pageW - 2*margin,
    Height: pageH - margin - headerH - footerH, // above the footer
}
```

---

## 14. Template switching

Use template switching when different pages need a different layout:
a title page, a table of contents, single-column content, and two columns
for dense tables.

### Method 1 — `NextPageTemplate` + `PageBreak`

The most common approach. `NextPageTemplate` schedules the switch; `PageBreak`
executes the jump.

```go
story = append(story,
    // ... single-column content ...
    &layout.NextPageTemplate{TemplateID: "two-column"},
    &layout.PageBreak{},
    // ... two-column content ...
    &layout.NextPageTemplate{TemplateID: "single"},
    &layout.PageBreak{},
    // ... single-column content ...
)
```

### Method 2 — `PageBreak` with `NextTemplate`

Combines page break and template switch in one step:

```go
&layout.PageBreak{NextTemplate: "two-column"},
```

### Method 3 — `AutoNextTemplate`

Automatic switch after each page, without story actions. Useful for
"title page → rest":

```go
titleTemplate := &layout.PageTemplate{
    ID:               "title",
    Frames:           []*layout.LayoutFrame{titleFrame},
    OnPage:           titleDecorator,
    AutoNextTemplate: "body", // automatically switch to "body" after page 1
}
bodyTemplate := &layout.PageTemplate{
    ID:               "body",
    Frames:           []*layout.LayoutFrame{bodyFrame},
    OnPage:           bodyDecorator,
    AutoNextTemplate: "body", // stay on "body"
}
```

### Registration priority

The **first** registered template is used for page 1:

```go
dt.AddPageTemplate(titleTemplate) // page 1
dt.AddPageTemplate(bodyTemplate)  // active after switch
```

---

## 15. Multi-column layout

Multiple columns are achieved by adding multiple `LayoutFrame` objects to a
`PageTemplate`. The engine fills them left to right (in the order of the
`Frames` list).

```go
const (
    margin    = 50.0
    colGutter = 12.0  // space between columns
)
contentW := doc.PageWidth() - 2*margin
colW     := (contentW - colGutter) / 2
contentY := 90.0
contentH := doc.PageHeight() - contentY - 50.0

leftFrame := &layout.LayoutFrame{
    X: margin,          Y: contentY,
    Width: colW,        Height: contentH,
}
rightFrame := &layout.LayoutFrame{
    X: margin + colW + colGutter, Y: contentY,
    Width: colW,                  Height: contentH,
}

twoColTemplate := &layout.PageTemplate{
    ID:               "two-column",
    Frames:           []*layout.LayoutFrame{leftFrame, rightFrame},
    OnPage:           pageDecorator,
    AutoNextTemplate: "two-column",
}
```

> **Important:** Content flows *sequentially* from column to column.
> The left column is fully filled before the right column receives any content.
> This is the same behaviour as Platypus. There is no automatic balancing
> (as in newspaper column layout).

### Jumping explicitly to the right column

```go
story = append(story,
    leftColumnContent...,
    &layout.FrameBreak{}, // force transition to right column
    rightColumnContent...,
)
```

### Three-column layout

```go
colW := (contentW - 2*colGutter) / 3

col1 := &layout.LayoutFrame{X: margin,                    Y: contentY, Width: colW, Height: contentH}
col2 := &layout.LayoutFrame{X: margin+colW+colGutter,     Y: contentY, Width: colW, Height: contentH}
col3 := &layout.LayoutFrame{X: margin+2*(colW+colGutter), Y: contentY, Width: colW, Height: contentH}

threeColTemplate := &layout.PageTemplate{
    ID:     "three-column",
    Frames: []*layout.LayoutFrame{col1, col2, col3},
    OnPage: pageDecorator,
    AutoNextTemplate: "three-column",
}
```

---

## 16. Debugging with ShowBoundary

Enable `ShowBoundary: true` on frames to see their outlines as thin gray
rectangles. This lets you confirm that:

- Frames are positioned correctly.
- The correct template is active.
- Both columns are activated (even if one is empty).

```go
leftFrame := &layout.LayoutFrame{
    X: margin, Y: contentY,
    Width: colW, Height: contentH,
    ShowBoundary: true, // visible in the PDF
}
```

Remove `ShowBoundary: true` when the document is ready for production.

---

## 17. Implementing a custom Flowable

Implement the `Flowable` interface to create custom elements. All
seven methods are required.

### Minimal implementation

```go
type MyFlowable struct {
    Width, Height float64
    // ... custom fields ...
}

func (m *MyFlowable) Wrap(_ *pdf.Document, availWidth, _ float64) (float64, float64) {
    w := m.Width
    if w <= 0 || w > availWidth {
        w = availWidth
    }
    return w, m.Height
}

func (m *MyFlowable) Draw(doc *pdf.Document, x, y float64) error {
    // Draw here using doc.WriteLine, doc.DrawBorder, doc.FillRect, etc.
    return nil
}

func (m *MyFlowable) Split(_ *pdf.Document, _, _ float64) []Flowable { return nil }
func (m *MyFlowable) SpaceBefore() float64                           { return 0 }
func (m *MyFlowable) SpaceAfter() float64                            { return 0 }
func (m *MyFlowable) KeepWithNext() bool                             { return false }
func (m *MyFlowable) MinWidth() float64                              { return 0 }
```

### Guidelines

| Method | Guideline |
|--------|-----------|
| `Wrap` | Store the available width for use in `Draw`. Ignore `availHeight` if the height is fixed. |
| `Draw` | Draw only within the area `(x, y)` to `(x+width, y+height)`. |
| `Split` | Return `nil` if splitting is not possible (engine moves the flowable to the next frame). Return two flowables when splitting: the first part fits within `availHeight`, the second contains the remainder. |
| `SpaceBefore/After` | Use for extra whitespace; the engine applies this and suppresses `SpaceBefore` at the top of a frame. |
| `KeepWithNext` | Return `true` to prevent a frame break after this element. |

### Example — colored box flowable

```go
type ColorBox struct {
    BoxWidth, BoxHeight float64
    FillColor           pdf.Color
    BorderColor         pdf.Color
    SpaceBeforeVal      float64
    SpaceAfterVal       float64
}

func (cb *ColorBox) Wrap(_ *pdf.Document, availWidth, _ float64) (float64, float64) {
    w := cb.BoxWidth
    if w <= 0 || w > availWidth {
        w = availWidth
    }
    return w, cb.BoxHeight
}

func (cb *ColorBox) Draw(doc *pdf.Document, x, y float64) error {
    doc.FillRect(x, y, cb.BoxWidth, cb.BoxHeight, cb.FillColor)
    spec := pdf.BorderSpec{Thickness: 1, Color: cb.BorderColor}
    return doc.DrawBorder(x, y, cb.BoxWidth, cb.BoxHeight, pdf.NewUniformBorder(spec))
}

func (cb *ColorBox) Split(_ *pdf.Document, _, _ float64) []layout.Flowable { return nil }
func (cb *ColorBox) SpaceBefore() float64                                   { return cb.SpaceBeforeVal }
func (cb *ColorBox) SpaceAfter() float64                                    { return cb.SpaceAfterVal }
func (cb *ColorBox) KeepWithNext() bool                                     { return false }
func (cb *ColorBox) MinWidth() float64                                      { return cb.BoxWidth }
```

Usage:

```go
story = append(story,
    &ColorBox{
        BoxWidth: 0, BoxHeight: 40, // 0 width = full available width
        FillColor:      pdf.Color{R: 235, G: 245, B: 255},
        BorderColor:    pdf.ColorNavy,
        SpaceBeforeVal: 8,
        SpaceAfterVal:  8,
    },
)
```

### Splittable flowable

If a flowable can be split across frames, implement `Split`:

```go
func (mf *MyFlowable) Split(doc *pdf.Document, availWidth, availHeight float64) []layout.Flowable {
    if availHeight < mf.MinimumHeight() {
        return nil // does not even fit a small piece; move to next frame
    }
    part1 := &MyFlowable{/* content that fits in availHeight */}
    part2 := &MyFlowable{/* remaining content */}
    return []layout.Flowable{part1, part2}
}
```

> **Contract rule:** the sum of the heights of all returned parts must equal
> the original height. No content may be lost.

---

## 18. Frequently asked questions

### Why does all content appear in one column with a two-column layout?

The engine fills columns sequentially: the left column is fully filled
before the right column receives any content. If your content fits in the
left column, the right column remains empty.

**Solution:** Add more content, or use `FrameBreak` to jump explicitly to
the right column.

Enable `ShowBoundary: true` to inspect frame outlines.

---

### Why does text overflow the bottom edge of the frame?

Flowables are only split when they *do not fit* in the remaining space. If
a flowable has no `Split` implementation (returns `nil`), the engine moves
it to the next frame. Check that your frames are large enough and that
`Split` is implemented correctly for custom flowables.

---

### How do I add "Page X of Y" to the footer?

The standard `pdf.Document` supports this via the two-pass `Build` method.
Use the `doc.Build(func(){...})` approach combined with `doc.SetFooter`
instead of `DocTemplate` to use this feature.

If you are using `DocTemplate`, you can set the total page count upfront
(if known) or calculate it after the fact via a second build pass.

---

### Can I use `pdf.Frame` and `DocTemplate` together?

Yes. `pdf.Frame` (the lower-level frame API) and `DocTemplate` are completely
independent. Use `pdf.Frame` for freely positioned boxes (e.g. text blocks
next to images in a header); use `DocTemplate` for the main content flow.

---

### My flowable does not fit in any frame and Build returns an error.

`Build` returns an error if a flowable cannot be placed more than 10 consecutive
times. Check that:

- The `Height` of your frame is large enough.
- `Wrap` returns a realistic height (not larger than the frame height).
- There is no infinite loop in a custom `Split` implementation.

---

### How do I reset the font size after a page break?

The `PageDecorator` (`OnPage`) is called after every `AddPage` and typically
sets the font for the header. Your first paragraph style on the new page calls
`applyFont` during `Wrap`, so the font is restored immediately. Make sure every
`ParagraphStyle` has a `FontName` and `FontSize`.

---

## Quick reference

### Types

| Type | Description |
|------|-------------|
| `Flowable` | Interface for all placeable elements |
| `Paragraph` | Text element with word wrapping |
| `ParagraphStyle` | Visual formatting for `Paragraph` |
| `HAlign` | Alignment constant: `AlignLeft`, `AlignCenter`, `AlignRight` |
| `Spacer` | Reserves vertical space |
| `HRFlowable` | Horizontal line as a filled bar |
| `KeepTogether` | Keeps a group of elements together |
| `PageBreak` | Forces a page break |
| `FrameBreak` | Advances to the next frame |
| `CondPageBreak` | Conditional page break |
| `NextPageTemplate` | Schedules a template switch |
| `LayoutFrame` | Rectangular content area on a page |
| `PageTemplate` | Page layout: frames + decorators |
| `PageDecorator` | `func(doc *pdf.Document, pageNum int)` |
| `DocTemplate` | The layout engine |

### Functions

| Function | Description |
|----------|-------------|
| `NewDocTemplate(doc)` | Create a new engine |
| `(dt) AddPageTemplate(pt)` | Register a template |
| `(dt) Build(story)` | Process the story and generate the PDF |

### ParagraphStyle — quick reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `FontName` | `string` | `""` | Registered font name |
| `FontSize` | `float64` | `0` (→ 12) | Size in points |
| `Leading` | `float64` | `0` (→ FontSize×1.2) | Line spacing |
| `Alignment` | `HAlign` | `AlignLeft` | Alignment |
| `SpaceBefore` | `float64` | `0` | Space above |
| `SpaceAfter` | `float64` | `0` | Space below |
| `KeepWithNextPara` | `bool` | `false` | Bind to next flowable |
| `LeftIndent` | `float64` | `0` | Left indentation |
| `RightIndent` | `float64` | `0` | Right indentation |
| `TextColor` | `*pdf.Color` | `nil` | Text color |

### LayoutFrame — quick reference

| Field | Type | Description |
|-------|------|-------------|
| `X, Y` | `float64` | Top-left position on the page |
| `Width, Height` | `float64` | Outer dimensions |
| `Padding` | `pdf.Padding` | Inner whitespace |
| `ID` | `string` | Optional name |
| `ShowBoundary` | `bool` | Draw debug outline |
