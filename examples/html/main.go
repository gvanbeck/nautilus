// Command html demonstrates converting inline HTML to PDF text using the
// pdf/html package.
//
// Page flow is handled by the pdf/layout engine via a DocTemplate story.
// Two custom Flowable types (htmlSample, classSample) render each HTML
// example as a grey label line followed by the rendered spans.
//
// # Usage
//
//	go run ./examples/html \
//	    -font   /Library/Fonts/Lato-Regular.ttf \
//	    -bold   /Library/Fonts/Lato-Bold.ttf \
//	    -italic /Library/Fonts/Lato-Italic.ttf \
//	    -mono   /Library/Fonts/CourierPrime-Regular.ttf \
//	    -out    output.pdf
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gvanbeck/nautilus/pdf"
	htmlpkg "github.com/gvanbeck/nautilus/pdf/html"
	"github.com/gvanbeck/nautilus/pdf/layout"
)

// ── Layout constants ──────────────────────────────────────────────────────────

const (
	marginLeft     = 60.0
	marginTop      = 60.0
	sampleFontSize = 13.0
	labelFontSize  = 9.0
	labelLineH     = labelFontSize * 1.4 // vertical space for the grey label
)

// ── htmlSample ────────────────────────────────────────────────────────────────
// Renders one HTML example: a small grey label followed by the rendered spans.
// Underline and strikethrough decorations are drawn by WriteHTMLSpans.

type htmlSample struct {
	html    string
	classes htmlpkg.ClassStyle
	fontFor func(htmlpkg.Style) string
}

func (s *htmlSample) Wrap(_ *pdf.Document, w, _ float64) (float64, float64) {
	return w, labelLineH + sampleFontSize*1.5 + 4
}

func (s *htmlSample) Draw(doc *pdf.Document, x, y float64) error {
	doc.SetFont("regular", labelFontSize) //nolint
	doc.SetTextColor(130, 130, 130)
	doc.WriteLine(s.html, x, y) //nolint
	spans, _ := htmlpkg.Parse(s.html, s.classes)
	doc.SetTextColor(30, 30, 30)
	_, err := doc.WriteHTMLSpans(spans, s.fontFor, sampleFontSize, x, y+labelLineH)
	return err
}

func (s *htmlSample) Split(_ *pdf.Document, _, _ float64) []layout.Flowable { return nil }
func (s *htmlSample) SpaceBefore() float64                                  { return 0 }
func (s *htmlSample) SpaceAfter() float64                                   { return 8 }
func (s *htmlSample) KeepWithNext() bool                                    { return false }
func (s *htmlSample) MinWidth() float64                                     { return 0 }

// ── classSample ───────────────────────────────────────────────────────────────
// Like htmlSample but draws a tiny ".classname" annotation above each classed
// span to make the class attribute visible.

type classSample struct {
	html    string
	classes htmlpkg.ClassStyle
	fontFor func(htmlpkg.Style) string
}

func (s *classSample) Wrap(_ *pdf.Document, w, _ float64) (float64, float64) {
	return w, 8 + labelLineH + sampleFontSize*1.5 + 4
}

func (s *classSample) Draw(doc *pdf.Document, x, y float64) error {
	doc.SetFont("regular", labelFontSize) //nolint
	doc.SetTextColor(130, 130, 130)
	doc.WriteLine(s.html, x, y) //nolint

	spans, _ := htmlpkg.Parse(s.html, s.classes)
	spansY := y + labelLineH + 8 // extra room above for annotations
	x2 := x
	for _, sp := range spans {
		doc.SetFont(s.fontFor(sp.Style), sampleFontSize) //nolint
		doc.SetTextColor(30, 30, 30)
		endX, _ := doc.WriteLine(sp.Text, x2, spansY)
		w, _ := doc.MeasureText(sp.Text)
		if sp.Class != "" {
			doc.SetFont("regular", 7) //nolint
			doc.SetTextColor(100, 100, 200)
			cw, _ := doc.MeasureText("." + sp.Class)
			doc.WriteLine("."+sp.Class, x2+(w-cw)/2, spansY-8) //nolint
		}
		x2 = endX
	}
	return nil
}

func (s *classSample) Split(_ *pdf.Document, _, _ float64) []layout.Flowable { return nil }
func (s *classSample) SpaceBefore() float64                                  { return 0 }
func (s *classSample) SpaceAfter() float64                                   { return 10 }
func (s *classSample) KeepWithNext() bool                                    { return false }
func (s *classSample) MinWidth() float64                                     { return 0 }

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	fontPath   := flag.String("font",   "/Library/Fonts/Lato-Regular.ttf",        "regular TTF/OTF font")
	boldPath   := flag.String("bold",   "/Library/Fonts/Lato-Bold.ttf",           "bold TTF/OTF font")
	italicPath := flag.String("italic", "/Library/Fonts/Lato-Italic.ttf",         "italic TTF/OTF font")
	monoPath   := flag.String("mono",   "/Library/Fonts/CourierPrime-Regular.ttf", "monospace TTF/OTF font")
	outPath    := flag.String("out",    "output.pdf",                              "output PDF file path")
	flag.Parse()

	doc, err := pdf.New(pdf.Config{
		PageSize:         pdf.PageSizeA4,
		DefaultFontSize:  12,
		LineHeightFactor: 1.5,
	})
	if err != nil {
		log.Fatalf("create document: %v", err)
	}

	if _, err := os.Stat(*fontPath); err != nil {
		log.Fatalf("font not found at %q", *fontPath)
	}
	if err := doc.RegisterFont("regular", *fontPath); err != nil {
		log.Fatalf("register regular font: %v", err)
	}

	hasBold := false
	if _, err := os.Stat(*boldPath); err == nil {
		if err := doc.RegisterFont("bold", *boldPath); err == nil {
			hasBold = true
		}
	}
	hasItalic := false
	if _, err := os.Stat(*italicPath); err == nil {
		if err := doc.RegisterFont("italic", *italicPath); err == nil {
			hasItalic = true
		}
	}
	hasMono := false
	if _, err := os.Stat(*monoPath); err == nil {
		if err := doc.RegisterFont("mono", *monoPath); err == nil {
			hasMono = true
		}
	}

	fontFor := func(s htmlpkg.Style) string {
		switch {
		case s.Monospace && hasMono:
			return "mono"
		case s.Bold && hasBold:
			return "bold"
		case s.Italic && hasItalic:
			return "italic"
		default:
			return "regular"
		}
	}
	boldFont := func() string {
		if hasBold {
			return "bold"
		}
		return "regular"
	}

	contentW := doc.PageWidth() - marginLeft*2

	// ── Paragraph styles ─────────────────────────────────────────────────────
	titleStyle := layout.ParagraphStyle{
		FontName:   boldFont(),
		FontSize:   20,
		TextColor:  &pdf.Color{R: 20, G: 20, B: 100},
		SpaceAfter: 12,
	}
	introStyle := layout.ParagraphStyle{
		FontName:    "regular",
		FontSize:    10,
		TextColor:   &pdf.Color{R: 100, G: 100, B: 100},
		SpaceAfter:  20,
		Leading:     15,
	}
	headingStyle := layout.ParagraphStyle{
		FontName:         boldFont(),
		FontSize:         11,
		TextColor:        &pdf.Color{R: 30, G: 30, B: 120},
		SpaceBefore:      14,
		SpaceAfter:       6,
		KeepWithNextPara: true,
	}
	noteStyle := layout.ParagraphStyle{
		FontName:   "regular",
		FontSize:   10,
		TextColor:  &pdf.Color{R: 80, G: 80, B: 80},
		SpaceAfter: 6,
	}

	// ── ClassStyle map ────────────────────────────────────────────────────────
	classes := htmlpkg.ClassStyle{
		"highlight": {Bold: true},
		"note":      {Italic: true},
		"important": {Bold: true, Underline: true},
	}

	// ── Story helpers ─────────────────────────────────────────────────────────
	para := func(text string, style layout.ParagraphStyle) layout.Flowable {
		return &layout.Paragraph{Text: text, Style: style}
	}
	sample := func(h string) layout.Flowable {
		return &htmlSample{html: h, fontFor: fontFor}
	}
	sampleCls := func(h string) layout.Flowable {
		return &classSample{html: h, classes: classes, fontFor: fontFor}
	}

	// ── Build story ───────────────────────────────────────────────────────────
	story := []layout.Flowable{
		para("HTML → PDF: complete feature tour", titleStyle),
		para(
			"This document exercises every feature of the pdf/html package: "+
				"inline span formatting (bold, italic, underline, strikethrough, "+
				"monospace, class attributes, nesting) and full HTML table parsing "+
				"with thead/tbody/tfoot, colspan/rowspan, cell colours, and rich text.",
			introStyle,
		),

		// ── Section 1 ────────────────────────────────────────────────────────
		para("1. Basic inline tags — <b>, <i>, <u>, <ins>, <strong>, <em>", headingStyle),
		sample(`Normal text, then <b>bold text</b>, then normal again.`),
		sample(`Normal text, then <i>italic text</i>, then normal again.`),
		sample(`Normal text, then <u>underlined text</u>, then normal again.`),
		sample(`Normal text, then <ins>inserted text</ins>, then normal again.`),
		sample(`<strong>strong</strong> and <em>em</em> work the same as b/i.`),

		// ── Section 1b ───────────────────────────────────────────────────────
		para("1b. Semantic italic — <cite>, <var>, <dfn>", headingStyle),
		sample(`<cite>The Go Programming Language</cite> is a great book.`),
		sample(`Set <var>x</var> to the initial value.`),
		sample(`The term <dfn>goroutine</dfn> refers to a lightweight thread.`),

		// ── Section 2 ────────────────────────────────────────────────────────
		para("2. Strikethrough — <s>, <strike>, <del>", headingStyle),
		sample(`Price was <s>€99</s> now €49.`),
		sample(`<strike>This text is struck through.</strike>`),
		sample(`<del>Deleted content</del> replaced by new content.`),

		// ── Section 3 ────────────────────────────────────────────────────────
		para("3. Monospace — <code>, <tt>, <kbd>, <samp>", headingStyle),
		para("Uses the -mono font when provided, otherwise falls back to regular.", noteStyle),
		sample(`Run <code>go build ./...</code> to compile.`),
		sample(`Press <kbd>Ctrl+C</kbd> to cancel.`),
		sample(`The output was <samp>hello, world</samp>.`),
		sample(`Use <tt>fmt.Println</tt> for simple output.`),

		// ── Section 4 ────────────────────────────────────────────────────────
		para("4. Nested tags — styles combine freely", headingStyle),
		sample(`<b><i>bold and italic together</i></b>`),
		sample(`Start <b>bold <i>bold+italic</i> bold again</b> end.`),
		sample(`<u><b>underline+bold</b></u> plain.`),
		sample(`<b><code>bold monospace</code></b> then plain.`),

		// ── Section 5 ────────────────────────────────────────────────────────
		para(`5. Class attribute — <span class="..."> with ClassStyle map`, headingStyle),
		para(`ClassStyle map: "highlight"→Bold, "note"→Italic, "important"→Bold+Underline`, noteStyle),
		sampleCls(`Plain, <span class="highlight">highlighted (bold)</span>, plain.`),
		sampleCls(`Plain, <span class="note">note (italic)</span>, plain.`),
		sampleCls(`Plain, <span class="important">important (bold+underline)</span>, plain.`),
		sampleCls(`<b class="note">bold tag with italic class</b> — styles combine.`),
		sampleCls(`<span class="unknown">unknown class</span> — class preserved, no style change.`),
	}

	// ── DocTemplate: one frame per page, automatic overflow ──────────────────
	frame := &layout.LayoutFrame{
		X:      marginLeft,
		Y:      marginTop,
		Width:  contentW,
		Height: doc.PageHeight() - marginTop*2,
	}
	dt := layout.NewDocTemplate(doc)
	dt.AddPageTemplate(&layout.PageTemplate{
		ID:     "main",
		Frames: []*layout.LayoutFrame{frame},
	})

	if err := dt.Build(story); err != nil {
		log.Fatalf("layout build: %v", err)
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// Tables section — tables manage their own pagination via Table.Draw()
	// ═══════════════════════════════════════════════════════════════════════════
	doc.AddPage()
	y := marginTop

	// heading helper (reused from layout styles above, drawn manually)
	tableHeading := func(text string) {
		doc.SetFont(boldFont(), 11) //nolint
		doc.SetTextColor(30, 30, 120)
		endY, _ := doc.WriteText(text, marginLeft, y, contentW)
		y = endY + 6
	}

	doc.SetFont(boldFont(), 16) //nolint
	doc.SetTextColor(20, 20, 100)
	doc.WriteLine("HTML Table Parsing", marginLeft, y) //nolint
	y += 28

	doc.SetFont("regular", 10) //nolint
	doc.SetTextColor(100, 100, 100)
	endY, _ := doc.WriteText(
		"The following tables are produced by html.ParseTable + doc.TableFromHTML. "+
			"Each table exercises a different set of HTML table features.",
		marginLeft, y, contentW,
	)
	y = endY + 20

	// ── Table A: thead/tbody/tfoot, alignment, text colours ──────────────────
	tableHeading("A. thead / tbody / tfoot — alignment and text colours")

	tableA := `
<table>
  <caption>Q1 Sales Summary</caption>
  <thead>
    <tr><th>Region</th><th align="right">Units</th><th align="right">Revenue</th><th>Trend</th></tr>
  </thead>
  <tbody>
    <tr><td>North</td><td align="right">1,240</td><td align="right" style="color:green">€ 62,000</td><td>▲</td></tr>
    <tr bgcolor="#f0f0f0"><td>South</td><td align="right">980</td><td align="right" style="color:red">€ 44,100</td><td>▼</td></tr>
    <tr><td>East</td><td align="right">2,050</td><td align="right" style="color:green">€ 92,250</td><td>▲</td></tr>
    <tr bgcolor="#f0f0f0"><td>West</td><td align="right">730</td><td align="right">€ 36,500</td><td>—</td></tr>
  </tbody>
  <tfoot>
    <tr><td><b>Total</b></td><td align="right"><b>5,000</b></td><td align="right"><b>€ 234,850</b></td><td></td></tr>
  </tfoot>
</table>`

	htA, err := htmlpkg.ParseTable(tableA, nil)
	if err != nil {
		log.Fatalf("parse table A: %v", err)
	}
	colW := contentW / 4
	tblA := doc.TableFromHTML(htA, pdf.TableConfig{
		X: marginLeft, Y: y,
		ColWidths:  []float64{colW * 1.4, colW * 0.8, colW * 1.0, colW * 0.8},
		PageBottom: doc.PageHeight() - marginTop,
		Border: pdf.Border{
			Top: &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorNavy},
			Bottom: &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorNavy},
			Left: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
			Right: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
		},
		DefaultCellStyle: pdf.CellStyle{
			FontName: "regular", FontSize: 11,
			Padding: pdf.Padding{Top: 4, Bottom: 4, Left: 6, Right: 6},
			Border:  pdf.Border{Bottom: &pdf.BorderSpec{Thickness: 0.3, Color: pdf.ColorLightGray}},
		},
	}, pdf.HtmlTableOptions{
		SpanFontFor: fontFor,
		HeaderStyle: pdf.CellStyle{
			FontName: boldFont(), FontSize: 11,
			Background: &pdf.Color{R: 30, G: 30, B: 120},
			TextColor:  &pdf.ColorWhite,
		},
		FooterStyle: pdf.CellStyle{
			FontName:   boldFont(),
			Background: &pdf.Color{R: 220, G: 220, B: 235},
		},
	})
	if err := tblA.Draw(); err != nil {
		log.Fatalf("draw table A: %v", err)
	}
	y = tblA.CurrentY() + 4
	doc.SetFont("regular", 8) //nolint
	doc.SetTextColor(120, 120, 120)
	doc.WriteText("Table caption: "+htA.Caption, marginLeft, y, contentW) //nolint
	y += 24

	// ── Table B: colspan and rowspan ──────────────────────────────────────────
	tableHeading("B. colspan and rowspan — merged cells")

	tableB := `
<table>
  <thead>
    <tr><th colspan="2">Product</th><th colspan="2">Sales</th></tr>
    <tr><th>Name</th><th>Category</th><th>Units</th><th>Revenue</th></tr>
  </thead>
  <tbody>
    <tr>
      <td rowspan="2">Widget Pro</td><td>Hardware</td>
      <td align="right">500</td><td align="right">€ 25,000</td>
    </tr>
    <tr><td>Software</td><td align="right">300</td><td align="right">€ 15,000</td></tr>
    <tr>
      <td colspan="2" align="center">Other products</td>
      <td align="right">1,200</td><td align="right">€ 48,000</td>
    </tr>
  </tbody>
</table>`

	htB, err := htmlpkg.ParseTable(tableB, nil)
	if err != nil {
		log.Fatalf("parse table B: %v", err)
	}
	cw := contentW / 4
	tblB := doc.TableFromHTML(htB, pdf.TableConfig{
		X: marginLeft, Y: y,
		ColWidths:  []float64{cw * 1.3, cw * 0.9, cw * 0.9, cw * 0.9},
		PageBottom: doc.PageHeight() - marginTop,
		Border: pdf.Border{
			Top: &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorDarkGray},
			Bottom: &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorDarkGray},
			Left: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
			Right: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
		},
		DefaultCellStyle: pdf.CellStyle{
			FontName: "regular", FontSize: 11,
			Padding: pdf.Padding{Top: 4, Bottom: 4, Left: 6, Right: 6},
			Border: pdf.Border{
				Bottom: &pdf.BorderSpec{Thickness: 0.3, Color: pdf.ColorLightGray},
				Right:  &pdf.BorderSpec{Thickness: 0.3, Color: pdf.ColorLightGray},
			},
		},
	}, pdf.HtmlTableOptions{
		SpanFontFor: fontFor,
		HeaderStyle: pdf.CellStyle{
			FontName: boldFont(), FontSize: 11,
			Background: &pdf.Color{R: 60, G: 60, B: 60},
			TextColor:  &pdf.ColorWhite,
			HAlign:     pdf.HAlignCenter,
		},
	})
	if err := tblB.Draw(); err != nil {
		log.Fatalf("draw table B: %v", err)
	}
	y = tblB.CurrentY() + 28

	// ── Table C: bgcolor, rich text, <br> ────────────────────────────────────
	tableHeading("C. Cell background colours, rich text, and <br> line breaks")

	tableC := `
<table>
  <thead>
    <tr><th>Feature</th><th>HTML attribute</th><th>Description</th></tr>
  </thead>
  <tbody>
    <tr>
      <td bgcolor="#dff0d8"><b>Background color</b></td>
      <td><code>bgcolor="#dff0d8"</code></td>
      <td>Cell background via the <code>bgcolor</code> attribute.</td>
    </tr>
    <tr>
      <td bgcolor="#d9edf7"><b>CSS background</b></td>
      <td><code>style="background-color: #d9edf7"</code></td>
      <td>Same result using inline CSS.</td>
    </tr>
    <tr>
      <td style="background-color:#fcf8e3"><b>Text colour</b></td>
      <td><code>style="color: ..."</code></td>
      <td style="color:#8a6d3b">This text is brown via <code>style="color"</code>.</td>
    </tr>
    <tr>
      <td bgcolor="#f2dede"><b>Line breaks</b></td>
      <td><code>&lt;br&gt;</code></td>
      <td>First line.<br>Second line.<br><i>Third line (italic).</i></td>
    </tr>
    <tr>
      <td><b>Rich inline text</b></td>
      <td>mixed tags</td>
      <td>Normal, <b>bold</b>, <i>italic</i>, <code>mono</code>, <b><i>bold+italic</i></b>.</td>
    </tr>
  </tbody>
</table>`

	htC, err := htmlpkg.ParseTable(tableC, nil)
	if err != nil {
		log.Fatalf("parse table C: %v", err)
	}
	tblC := doc.TableFromHTML(htC, pdf.TableConfig{
		X: marginLeft, Y: y,
		ColWidths:  []float64{contentW * 0.22, contentW * 0.3, contentW * 0.48},
		PageBottom: doc.PageHeight() - marginTop,
		Border: pdf.Border{
			Top: &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorDarkGray},
			Bottom: &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorDarkGray},
			Left: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
			Right: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
		},
		DefaultCellStyle: pdf.CellStyle{
			FontName: "regular", FontSize: 10,
			Padding: pdf.Padding{Top: 5, Bottom: 5, Left: 6, Right: 6},
			Border: pdf.Border{
				Bottom: &pdf.BorderSpec{Thickness: 0.3, Color: pdf.ColorLightGray},
				Right:  &pdf.BorderSpec{Thickness: 0.3, Color: pdf.ColorLightGray},
			},
		},
	}, pdf.HtmlTableOptions{
		SpanFontFor: fontFor,
		HeaderStyle: pdf.CellStyle{
			FontName: boldFont(), FontSize: 10,
			Background: &pdf.Color{R: 50, G: 50, B: 100},
			TextColor:  &pdf.ColorWhite,
		},
	})
	if err := tblC.Draw(); err != nil {
		log.Fatalf("draw table C: %v", err)
	}
	y = tblC.CurrentY() + 28

	// ── Table D: valign + MinOrphanRows ───────────────────────────────────────
	tableHeading("D. Vertical alignment + MinOrphanRows (header stays with data)")

	tableD := `
<table>
  <thead>
    <tr><th>valign="top"</th><th>valign="middle"</th><th>valign="bottom"</th></tr>
  </thead>
  <tbody>
    <tr>
      <td valign="top">Line 1<br>Line 2<br>Line 3<br>Line 4<br>Line 5<br>Line 6<br>Line 7<br>Line 8</td>
      <td valign="middle">Only two lines.<br>Vertically centred.</td>
      <td valign="bottom">Only two lines.<br>Pinned to bottom.</td>
    </tr>
  </tbody>
</table>`

	htD, err := htmlpkg.ParseTable(tableD, nil)
	if err != nil {
		log.Fatalf("parse table D: %v", err)
	}
	colW3 := contentW / 3
	tblD := doc.TableFromHTML(htD, pdf.TableConfig{
		X: marginLeft, Y: y,
		ColWidths:     []float64{colW3, colW3, colW3},
		ContinuationY: marginTop,
		RepeatRows:    1,
		MinOrphanRows: 1,
		PageBottom:    doc.PageHeight() - marginTop,
		Border: pdf.Border{
			Top: &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorDarkGray},
			Bottom: &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorDarkGray},
			Left: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
			Right: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
		},
		DefaultCellStyle: pdf.CellStyle{
			FontName: "regular", FontSize: 11,
			Padding: pdf.Padding{Top: 4, Bottom: 4, Left: 6, Right: 6},
			Border:  pdf.Border{Right: &pdf.BorderSpec{Thickness: 0.3, Color: pdf.ColorLightGray}},
		},
	}, pdf.HtmlTableOptions{
		SpanFontFor: fontFor,
		HeaderStyle: pdf.CellStyle{
			FontName: boldFont(), FontSize: 11,
			Background: &pdf.Color{R: 80, G: 80, B: 80},
			TextColor:  &pdf.ColorWhite,
			HAlign:     pdf.HAlignCenter,
		},
	})
	if err := tblD.Draw(); err != nil {
		log.Fatalf("draw table D: %v", err)
	}

	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save PDF: %v", err)
	}
	fmt.Printf("PDF written to %s\n", *outPath)
}
