// Command layout demonstrates the Platypus-inspired high-level layout engine.
//
// It generates output.pdf that shows how a story (a flat list of Flowables)
// is automatically flowed across frames and pages by the DocTemplate engine.
//
// The example covers:
//   - Single-column layout with Paragraphs and Spacers
//   - Horizontal rules (HRFlowable)
//   - KeepTogether (heading + body never separated)
//   - Conditional page breaks
//   - Two-column layout via a second PageTemplate
//   - Explicit PageBreak and NextPageTemplate
//
// # Usage
//
//	go run ./examples/layout \
//	    -font /Library/Fonts/Lato-Medium.ttf \
//	    -bold /Library/Fonts/Lato-Black.ttf \
//	    -out output.pdf
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/layout"
)

func main() {
	fontPath := flag.String("font", "/Library/Fonts/Lato-Medium.ttf", "path to regular TTF font")
	boldPath := flag.String("bold", "/Library/Fonts/Lato-Black.ttf", "path to bold TTF font")
	outPath := flag.String("out", "output.pdf", "output PDF path")
	flag.Parse()

	// ── Create document ───────────────────────────────────────────────────
	doc, err := pdf.New(pdf.Config{
		PageSize: pdf.PageSizeA4,
	})
	if err != nil {
		log.Fatalf("new document: %v", err)
	}

	// ── Register fonts ────────────────────────────────────────────────────
	if _, err := os.Stat(*fontPath); err != nil {
		log.Fatalf("regular font not found at %q", *fontPath)
	}
	if err := doc.RegisterFont("regular", *fontPath); err != nil {
		log.Fatalf("register regular: %v", err)
	}

	hasBold := false
	if _, err := os.Stat(*boldPath); err == nil {
		if err := doc.RegisterFont("bold", *boldPath); err != nil {
			log.Printf("warning: bold font: %v (using regular)", err)
		} else {
			hasBold = true
		}
	}
	boldFont := "regular"
	if hasBold {
		boldFont = "bold"
	}

	// Set initial font so MeasureText works from the start.
	if err := doc.SetFont("regular", 11); err != nil {
		log.Fatalf("set font: %v", err)
	}

	// ── Layout geometry ───────────────────────────────────────────────────
	const (
		margin  = 50.0
		headerH = 40.0 // reserved at the top for the header band
		footerH = 36.0 // reserved at the bottom
	)
	pageW := doc.PageWidth()
	pageH := doc.PageHeight()

	contentX := margin
	contentY := margin + headerH
	contentW := pageW - 2*margin
	contentH := pageH - margin - headerH - footerH

	// ── Shared decorator: header + footer on every page ───────────────────
	pageDecorator := func(d *pdf.Document, pageNum int) {
		// Header line.
		spec := &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorNavy}
		d.DrawBorder(margin, margin, contentW, headerH-4, pdf.Border{Bottom: spec}) //nolint

		d.SetFont("regular", 8)   //nolint
		d.SetTextColor(80, 80, 80)
		d.WriteLine("Nautilus Layout Engine — Demo", margin, margin+10) //nolint

		num := fmt.Sprintf("Page %d", pageNum)
		w, _ := d.MeasureText(num)
		d.WriteLine(num, pageW-margin-w, margin+10) //nolint

		// Footer line.
		footerY := pageH - footerH
		d.DrawBorder(margin, footerY, contentW, 0, pdf.Border{ //nolint
			Top: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
		})
		d.SetFont("regular", 8)  //nolint
		d.SetTextColor(150, 150, 150)
		d.WriteLine("github.com/vanbeckevoort", margin, footerY+8) //nolint
	}

	// ── Page templates ────────────────────────────────────────────────────
	// Template 1: single-column.
	singleFrame := &layout.LayoutFrame{
		X: contentX, Y: contentY,
		Width: contentW, Height: contentH,
	}
	singleTemplate := &layout.PageTemplate{
		ID:               "single",
		Frames:           []*layout.LayoutFrame{singleFrame},
		OnPage:           pageDecorator,
		AutoNextTemplate: "single",
	}

	// Template 2: two-column.
	colGutter := 12.0
	colW := (contentW - colGutter) / 2
	leftFrame := &layout.LayoutFrame{
		X: contentX, Y: contentY,
		Width: colW, Height: contentH,
		ShowBoundary: true,
	}
	rightFrame := &layout.LayoutFrame{
		X: contentX + colW + colGutter, Y: contentY,
		Width: colW, Height: contentH,
		ShowBoundary: true,
	}
	twoColTemplate := &layout.PageTemplate{
		ID:               "two-column",
		Frames:           []*layout.LayoutFrame{leftFrame, rightFrame},
		OnPage:           pageDecorator,
		AutoNextTemplate: "two-column",
	}

	// ── Style definitions ─────────────────────────────────────────────────
	title := layout.ParagraphStyle{
		FontName:    boldFont,
		FontSize:    26,
		SpaceBefore: 0,
		SpaceAfter:  20,
	}
	h1 := layout.ParagraphStyle{
		FontName:        boldFont,
		FontSize:        14,
		SpaceBefore:     16,
		SpaceAfter:      4,
		KeepWithNextPara: true,
	}
	h2 := layout.ParagraphStyle{
		FontName:        boldFont,
		FontSize:        11,
		SpaceBefore:     10,
		SpaceAfter:      2,
		KeepWithNextPara: true,
	}
	body := layout.ParagraphStyle{
		FontName:   "regular",
		FontSize:   11,
		SpaceAfter: 6,
	}
	bodyCenter := layout.ParagraphStyle{
		FontName:   "regular",
		FontSize:   11,
		SpaceAfter: 6,
		Alignment:  layout.AlignCenter,
	}
	caption := layout.ParagraphStyle{
		FontName:   "regular",
		FontSize:   9,
		SpaceAfter: 4,
		TextColor:  &pdf.ColorGray,
	}

	p := func(text string, style layout.ParagraphStyle) layout.Flowable {
		return &layout.Paragraph{Text: text, Style: style}
	}
	hr := func() layout.Flowable {
		return &layout.HRFlowable{
			Width: 1.0, Thickness: 0.75,
			Color:  pdf.ColorLightGray,
			Before: 8, After: 8,
		}
	}
	sp := func(h float64) layout.Flowable {
		return &layout.Spacer{Height: h}
	}

	// ── Story ─────────────────────────────────────────────────────────────
	// A "story" is a flat slice of Flowables. The DocTemplate engine
	// consumes it in order, distributing content across frames and pages.
	story := []layout.Flowable{
		// ── Title section ─────────────────────────────────────────────────
		p("Nautilus Layout Engine", title),
		p("A Platypus-inspired flow layout for Go PDF generation.", bodyCenter),
		hr(),

		// ── Introduction ──────────────────────────────────────────────────
		// KeepTogether ensures the heading always appears with its first paragraph.
		&layout.KeepTogether{Flowables: []layout.Flowable{
			p("1. Introduction", h1),
			p("The layout engine builds on top of the Nautilus pdf.Document "+
				"drawing primitives and adds a higher-level abstraction: the "+
				"Flowable interface.  Any piece of content — a paragraph, an "+
				"image, a spacer, a horizontal rule — implements Flowable and "+
				"can be placed in a Story.", body),
		}},

		p("The DocTemplate engine processes the story sequentially.  When a "+
			"frame fills up, it automatically advances to the next frame or "+
			"starts a new page, calling user-supplied decorators to render "+
			"headers and footers along the way.", body),

		p("This mirrors the design of Platypus from the Python ReportLab "+
			"library, adapted to Go idioms: explicit struct initialization "+
			"instead of keyword arguments, interfaces instead of class "+
			"inheritance, and value semantics where appropriate.", body),

		// ── Flowable types ─────────────────────────────────────────────────
		&layout.KeepTogether{Flowables: []layout.Flowable{
			p("2. Flowable Types", h1),
			p("The following built-in Flowable types are provided:", body),
		}},

		&layout.KeepTogether{Flowables: []layout.Flowable{
			p("Paragraph", h2),
			p("The primary text element.  Supports word-wrapping, font and "+
				"colour overrides, left/right indentation, and horizontal "+
				"alignment (left, centre, right).  Long paragraphs can be "+
				"split across frames.", body),
		}},

		&layout.KeepTogether{Flowables: []layout.Flowable{
			p("Spacer", h2),
			p("Reserves a fixed amount of vertical space without rendering "+
				"anything visible.  Useful for adding breathing room between "+
				"sections.", body),
		}},

		&layout.KeepTogether{Flowables: []layout.Flowable{
			p("HRFlowable", h2),
			p("Draws a horizontal rule as a solid filled bar.  Width can be "+
				"an absolute point value or a fraction of the available width "+
				"(e.g. 0.8 = 80 %).  Horizontal alignment is configurable.", body),
		}},

		&layout.KeepTogether{Flowables: []layout.Flowable{
			p("KeepTogether", h2),
			p("Wraps a group of flowables that must appear on the same frame. "+
				"If the group does not fit in the remaining space, the engine "+
				"inserts a FrameBreak and retries on the next frame.  If the "+
				"group is larger than an entire frame, it falls back to "+
				"individual splitting.", body),
		}},

		// ── Action flowables ───────────────────────────────────────────────
		&layout.KeepTogether{Flowables: []layout.Flowable{
			p("3. Action Flowables", h1),
			p("Action flowables are zero-height elements that control the "+
				"engine rather than rendering visible content.", body),
		}},

		p("PageBreak forces an immediate page break.  An optional NextTemplate "+
			"ID switches the page template on the new page.  FrameBreak "+
			"advances to the next frame within the current page.  "+
			"CondPageBreak inserts a page break only when fewer than "+
			"MinHeight points remain in the current frame.  "+
			"NextPageTemplate schedules a template switch that takes "+
			"effect on the next page break.", body),

		// Conditional page break: only break if less than 120 pt remains.
		&layout.CondPageBreak{MinHeight: 120},

		// ── Page templates ─────────────────────────────────────────────────
		&layout.KeepTogether{Flowables: []layout.Flowable{
			p("4. Page Templates", h1),
			p("A PageTemplate groups one or more LayoutFrames with optional "+
				"OnPage / OnPageEnd decorators.  The engine fills frames in "+
				"the order they are listed; when the last frame on a page is "+
				"full, a new page is started.", body),
		}},

		p("Switching between templates is done with NextPageTemplate (deferred) "+
			"or by passing a NextTemplate ID to PageBreak (immediate).  "+
			"AutoNextTemplate lets a template automatically switch to another "+
			"template after each page — enabling first-page vs. later-page "+
			"layouts without any story-side boilerplate.", body),

		hr(),
		p("The next section demonstrates a two-column layout produced by "+
			"switching to the 'two-column' PageTemplate.", caption),
		sp(4),

		// Switch to two-column layout on the next page.
		&layout.NextPageTemplate{TemplateID: "two-column"},
		&layout.PageBreak{},

		// ── Two-column content ─────────────────────────────────────────────
		// The grey outlines show the two LayoutFrame boundaries.
		// Content flows left→right: when the left frame fills up the engine
		// automatically continues in the right frame.
		p("5. Two-Column Layout", h1),
		p("This section is rendered in a two-column layout.  The story is "+
			"exactly the same flat list of Flowables — the only difference is "+
			"the active PageTemplate, which now provides two LayoutFrames per "+
			"page instead of one.", body),

		p("When the left frame fills up, the engine automatically continues "+
			"in the right frame.  When both frames are full, a new page is "+
			"started with the same two-column template (set via "+
			"AutoNextTemplate).", body),

		hr(),

		p("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do "+
			"eiusmod tempor incididunt ut labore et dolore magna aliqua.  "+
			"Ut enim ad minim veniam, quis nostrud exercitation ullamco "+
			"laboris nisi ut aliquip ex ea commodo consequat.", body),

		p("Duis aute irure dolor in reprehenderit in voluptate velit esse "+
			"cillum dolore eu fugiat nulla pariatur.  Excepteur sint occaecat "+
			"cupidatat non proident, sunt in culpa qui officia deserunt mollit "+
			"anim id est laborum.", body),

		p("Sed ut perspiciatis unde omnis iste natus error sit voluptatem "+
			"accusantium doloremque laudantium, totam rem aperiam eaque ipsa "+
			"quae ab illo inventore veritatis et quasi architecto beatae vitae "+
			"dicta sunt explicabo.  Nemo enim ipsam voluptatem quia voluptas "+
			"sit aspernatur aut odit aut fugit.", body),

		p("At vero eos et accusamus et iusto odio dignissimos ducimus qui "+
			"blanditiis praesentium voluptatum deleniti atque corrupti quos "+
			"dolores et quas molestias excepturi sint occaecati cupiditate "+
			"non provident.", body),

		p("Nam libero tempore, cum soluta nobis est eligendi optio cumque "+
			"nihil impedit quo minus id quod maxime placeat facere possimus, "+
			"omnis voluptas assumenda est, omnis dolor repellendus.", body),

		p("Temporibus autem quibusdam et aut officiis debitis aut rerum "+
			"necessitatibus saepe eveniet ut et voluptates repudiandae sint "+
			"et molestiae non recusandae.  Itaque earum rerum hic tenetur a "+
			"sapiente delectus.", body),

		p("Quis autem vel eum iure reprehenderit qui in ea voluptate velit "+
			"esse quam nihil molestiae consequatur, vel illum qui dolorem eum "+
			"fugiat quo voluptas nulla pariatur.", body),

		p("Neque porro quisquam est qui dolorem ipsum quia dolor sit amet, "+
			"consectetur, adipisci velit, sed quia non numquam eius modi "+
			"tempora incidunt ut labore et dolore magnam aliquam quaerat "+
			"voluptatem.", body),

		p("Ut enim ad minima veniam, quis nostrum exercitationem ullam "+
			"corporis suscipit laboriosam, nisi ut aliquid ex ea commodi "+
			"consequatur?  Quis autem vel eum iure reprehenderit.", body),

		p("Et harum quidem rerum facilis est et expedita distinctio.  Nam "+
			"libero tempore, cum soluta nobis est eligendi optio cumque nihil "+
			"impedit quo minus id quod maxime placeat facere possimus.", body),

		p("Similique sunt in culpa qui officia deserunt mollitia animi, id "+
			"est laborum et dolorum fuga.  Et harum quidem rerum facilis est "+
			"et expedita distinctio.", body),

		// Switch back to single column.
		&layout.NextPageTemplate{TemplateID: "single"},
		&layout.PageBreak{},

		// ── Final section ──────────────────────────────────────────────────
		p("6. Summary", h1),
		p("The Nautilus layout engine provides a clean, composable approach "+
			"to high-level PDF layout in Go:", body),
		p("  • A flat story list — no document tree to manage", body),
		p("  • Flowable interface — extend with custom elements", body),
		p("  • Automatic pagination — frames and pages managed for you", body),
		p("  • Template switching — single, two-column, or any geometry", body),
		p("  • KeepTogether — headings always accompanied by their body", body),
		sp(10),
		hr(),
		p("End of layout demo.", caption),
	}

	// ── Build ─────────────────────────────────────────────────────────────
	dt := layout.NewDocTemplate(doc)
	dt.AddPageTemplate(singleTemplate)
	dt.AddPageTemplate(twoColTemplate)

	if err := dt.Build(story); err != nil {
		log.Fatalf("build layout: %v", err)
	}

	// ── Save ──────────────────────────────────────────────────────────────
	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save PDF: %v", err)
	}
	fmt.Printf("PDF written to %s  (%d pages)\n", *outPath, doc.PageCount())
}
