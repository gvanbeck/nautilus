// Command html demonstrates converting inline HTML to PDF text using the
// pdf/html package.
//
// It generates output.pdf with examples of:
//   - <b>, <strong>              → bold
//   - <i>, <em>, <cite>, <var>, <dfn> → italic
//   - <u>, <ins>                 → underline (tracked in Span.Style)
//   - <s>, <strike>, <del>       → strikethrough (tracked in Span.Style)
//   - <code>, <tt>, <kbd>, <samp> → monospace
//   - Nesting: bold+italic combined
//   - <span class="..."> with a ClassStyle map
//
// # Usage
//
//	go run ./examples/html \
//	    -font   /Library/Fonts/Lato-Regular.ttf \
//	    -bold   /Library/Fonts/Lato-Bold.ttf \
//	    -italic /Library/Fonts/Lato-Italic.ttf \
//	    -mono   /Library/Fonts/CourierPrime-Regular.ttf \
//	    -out    output.pdf
//
// Only -font is required; -bold, -italic and -mono fall back to the regular
// font when omitted.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gvanbeck/nautilus/pdf"
	htmlpkg "github.com/gvanbeck/nautilus/pdf/html"
)

func main() {
	fontPath   := flag.String("font",   "/Library/Fonts/Lato-Regular.ttf",        "regular TTF/OTF font")
	boldPath   := flag.String("bold",   "/Library/Fonts/Lato-Bold.ttf",           "bold TTF/OTF font (falls back to regular)")
	italicPath := flag.String("italic", "/Library/Fonts/Lato-Italic.ttf",         "italic TTF/OTF font (falls back to regular)")
	monoPath   := flag.String("mono",   "/Library/Fonts/CourierPrime-Regular.ttf", "monospace TTF/OTF font (falls back to regular)")
	outPath    := flag.String("out",    "output.pdf",                              "output PDF file path")
	flag.Parse()

	// ── Create document ───────────────────────────────────────────────────
	doc, err := pdf.New(pdf.Config{
		PageSize:         pdf.PageSizeA4,
		DefaultFontSize:  12,
		LineHeightFactor: 1.5,
	})
	if err != nil {
		log.Fatalf("create document: %v", err)
	}

	// ── Register fonts ────────────────────────────────────────────────────
	mustExist := func(path, flag string) {
		if _, err := os.Stat(path); err != nil {
			log.Fatalf("font not found at %q — use -%s to set a valid path", path, flag)
		}
	}
	mustExist(*fontPath, "font")

	if err := doc.RegisterFont("regular", *fontPath); err != nil {
		log.Fatalf("register regular font: %v", err)
	}

	hasBold := false
	if _, err := os.Stat(*boldPath); err == nil {
		if err := doc.RegisterFont("bold", *boldPath); err != nil {
			log.Printf("warning: bold font: %v — falling back to regular", err)
		} else {
			hasBold = true
		}
	}

	hasItalic := false
	if _, err := os.Stat(*italicPath); err == nil {
		if err := doc.RegisterFont("italic", *italicPath); err != nil {
			log.Printf("warning: italic font: %v — falling back to regular", err)
		} else {
			hasItalic = true
		}
	}

	hasMono := false
	if _, err := os.Stat(*monoPath); err == nil {
		if err := doc.RegisterFont("mono", *monoPath); err != nil {
			log.Printf("warning: mono font: %v — falling back to regular", err)
		} else {
			hasMono = true
		}
	}

	// ── Font mapper ───────────────────────────────────────────────────────
	// Given a Span.Style, return the registered font name to use.
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

	// ── Layout constants ──────────────────────────────────────────────────
	const (
		marginLeft = 60.0
		marginTop  = 60.0
		fontSize   = 13.0
	)
	pageW    := doc.PageWidth()
	contentW := pageW - marginLeft*2

	// ── writeSpans renders a slice of Spans inline at (x, y). ─────────────
	// Each span may switch the font. Returns the final X position.
	writeSpans := func(spans []htmlpkg.Span, x, y float64) float64 {
		for _, span := range spans {
			if err := doc.SetFont(fontFor(span.Style), fontSize); err != nil {
				log.Printf("set font: %v", err)
				continue
			}
			doc.SetTextColor(30, 30, 30)
			endX, err := doc.WriteLine(span.Text, x, y)
			if err != nil {
				log.Printf("write span: %v", err)
				continue
			}
			x = endX
		}
		return x
	}

	// ── Helper: section heading ───────────────────────────────────────────
	heading := func(text string, y float64) float64 {
		name := "bold"
		if !hasBold {
			name = "regular"
		}
		doc.SetFont(name, 11) //nolint
		doc.SetTextColor(30, 30, 120)
		endY, _ := doc.WriteText(text, marginLeft, y, contentW)
		return endY + 6
	}

	// ── Helper: label in gray ─────────────────────────────────────────────
	label := func(text string, x, y float64) {
		doc.SetFont("regular", 9) //nolint
		doc.SetTextColor(130, 130, 130)
		doc.WriteLine(text, x, y) //nolint
	}

	// ── ClassStyle map ────────────────────────────────────────────────────
	// Maps CSS class names to Style overrides. Class names are always preserved
	// in Span.Class regardless of this map.
	classes := htmlpkg.ClassStyle{
		"highlight": {Bold: true},
		"note":      {Italic: true},
		"important": {Bold: true, Underline: true},
	}

	// ── Build document ────────────────────────────────────────────────────
	doc.AddPage()
	y := marginTop

	// Title
	doc.SetFont(func() string {
		if hasBold {
			return "bold"
		}
		return "regular"
	}(), 20) //nolint
	doc.SetTextColor(20, 20, 100)
	doc.WriteLine("HTML → PDF text spans", marginLeft, y) //nolint
	y += 32

	doc.SetFont("regular", 10) //nolint
	doc.SetTextColor(100, 100, 100)
	endY, _ := doc.WriteText(
		"Demonstrates the pdf/html package: inline HTML is parsed into Spans "+
			"carrying Style (Bold, Italic, Underline) and Class. "+
			"The caller switches fonts and applies class styles at render time.",
		marginLeft, y, contentW,
	)
	y = endY + 20

	// ── Section 1: basic b / i / u / ins ─────────────────────────────────
	y = heading("1. Tags <b>, <i>, <u>, <ins>, <strong>, <em>", y)

	samples := []struct {
		html  string
		descY float64
	}{
		{`Normal text, then <b>bold text</b>, then normal again.`, 0},
		{`Normal text, then <i>italic text</i>, then normal again.`, 0},
		{`Normal text, then <u>underlined text</u>, then normal again.`, 0},
		{`Normal text, then <ins>inserted text</ins>, then normal again.`, 0},
		{`<strong>strong</strong> and <em>em</em> work too.`, 0},
	}

	for _, s := range samples {
		spans, err := htmlpkg.Parse(s.html, nil)
		if err != nil {
			log.Fatalf("parse html: %v", err)
		}

		// Print the source HTML in gray above the rendered line.
		label(s.html, marginLeft, y)
		y += 14

		// Render the spans.
		writeSpans(spans, marginLeft, y)

		// Annotate underlined spans (underline not yet rendered by the library).
		x2 := marginLeft
		for _, sp := range spans {
			w, _ := doc.MeasureText(sp.Text)
			if sp.Style.Underline {
				doc.SetFont("regular", 8) //nolint
				doc.SetTextColor(200, 80, 0)
				doc.WriteLine("↑ underline (tracked)", x2, y+fontSize+2) //nolint
			}
			x2 += w
		}

		y += fontSize*1.5 + 18
	}

	y += 8

	// ── Section 1b: semantic italic ───────────────────────────────────────
	y = heading("1b. Semantic italic: <cite>, <var>, <dfn>", y)

	semanticSamples := []string{
		`<cite>The Go Programming Language</cite> is a great book.`,
		`Set <var>x</var> to the initial value.`,
		`The term <dfn>goroutine</dfn> refers to a lightweight thread.`,
	}
	for _, h := range semanticSamples {
		spans, _ := htmlpkg.Parse(h, nil)
		label(h, marginLeft, y)
		y += 14
		writeSpans(spans, marginLeft, y)
		y += fontSize*1.5 + 18
	}

	y += 8

	// ── Section 2: nesting ────────────────────────────────────────────────
	y = heading("2. Nested tags", y)

	nestSamples := []string{
		`<b><i>bold and italic</i></b>`,
		`Start <b>bold <i>bold+italic</i> bold again</b> end.`,
		`<u><b>underline+bold</b></u> plain.`,
	}
	for _, h := range nestSamples {
		spans, _ := htmlpkg.Parse(h, nil)
		label(h, marginLeft, y)
		y += 14
		writeSpans(spans, marginLeft, y)
		y += fontSize*1.5 + 18
	}

	y += 8

	// ── Section 3: class attribute ────────────────────────────────────────
	y = heading("3. Class attribute", y)

	doc.SetFont("regular", 10) //nolint
	doc.SetTextColor(80, 80, 80)
	endY, _ = doc.WriteText(
		`Class map: "highlight" → Bold, "note" → Italic, "important" → Bold+Underline`,
		marginLeft, y, contentW,
	)
	y = endY + 12

	classSamples := []string{
		`Plain, <span class="highlight">highlighted (bold)</span>, plain.`,
		`Plain, <span class="note">note (italic)</span>, plain.`,
		`Plain, <span class="important">important (bold+underline)</span>, plain.`,
		`<b class="note">bold tag with italic class</b> — styles combine.`,
		`<span class="unknown">unknown class</span> — class preserved, no style change.`,
	}
	for _, h := range classSamples {
		spans, _ := htmlpkg.Parse(h, classes)
		label(h, marginLeft, y)
		y += 14
		writeSpans(spans, marginLeft, y)

		// Show Class values for each span.
		x2 := marginLeft
		for _, sp := range spans {
			doc.SetFont("regular", fontSize) //nolint
			w, _ := doc.MeasureText(sp.Text)
			if sp.Class != "" {
				doc.SetFont("regular", 7) //nolint
				doc.SetTextColor(100, 100, 200)
				cw, _ := doc.MeasureText("." + sp.Class)
				doc.WriteLine("."+sp.Class, x2+(w-cw)/2, y-8) //nolint
			}
			x2 += w
		}

		y += fontSize*1.5 + 22
	}

	y += 8

	// ── Section 4: strikethrough ──────────────────────────────────────────
	y = heading("4. Strikethrough: <s>, <strike>, <del>", y)

	strikeSamples := []string{
		`Price was <s>€99</s> now €49.`,
		`<strike>This text is struck through.</strike>`,
		`<del>Deleted content</del> replaced by new content.`,
	}
	for _, h := range strikeSamples {
		spans, _ := htmlpkg.Parse(h, nil)
		label(h, marginLeft, y)
		y += 14
		writeSpans(spans, marginLeft, y)

		// Annotate strikethrough spans (not yet drawn by the library).
		x2 := marginLeft
		for _, sp := range spans {
			w, _ := doc.MeasureText(sp.Text)
			if sp.Style.Strikethrough {
				doc.SetFont("regular", 8) //nolint
				doc.SetTextColor(200, 80, 0)
				doc.WriteLine("↑ strikethrough (tracked)", x2, y+fontSize+2) //nolint
			}
			x2 += w
		}

		y += fontSize*1.5 + 22
	}

	y += 8

	// ── Section 5: monospace ──────────────────────────────────────────────
	y = heading("5. Monospace: <code>, <tt>, <kbd>, <samp>", y)

	doc.SetFont("regular", 10) //nolint
	doc.SetTextColor(80, 80, 80)
	monoNote := "Uses the -mono font when provided, otherwise falls back to the regular font."
	endY, _ = doc.WriteText(monoNote, marginLeft, y, contentW)
	y = endY + 12

	monoSamples := []string{
		`Run <code>go build ./...</code> to compile.`,
		`Press <kbd>Ctrl+C</kbd> to cancel.`,
		`The output was <samp>hello, world</samp>.`,
		`Use <tt>fmt.Println</tt> for simple output.`,
	}
	for _, h := range monoSamples {
		spans, _ := htmlpkg.Parse(h, nil)
		label(h, marginLeft, y)
		y += 14
		writeSpans(spans, marginLeft, y)
		y += fontSize*1.5 + 18
	}

	// ── Write output ──────────────────────────────────────────────────────
	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save PDF: %v", err)
	}
	fmt.Printf("PDF written to %s\n", *outPath)
}
