// Command basic demonstrates the nautilus PDF library.
//
// It generates output.pdf containing three pages:
//   - A cover page with title, subtitle, and emoji characters
//   - A content page with Unicode and font-size demonstrations
//   - A third page with word-wrap demonstration
//
// Every page carries a header with the document title and a footer showing
// "Page N of M" – rendered via the two-pass Build mechanism so the total page
// count is available from the first page onward.
//
// # Usage
//
//	go run ./examples/basic \
//	    -font /Library/Fonts/Lato-Medium.ttf \
//	    -bold /Library/Fonts/Lato-Black.ttf \
//	    -emoji assets/emoji/png/128 \
//	    -out output.pdf
//
// The -emoji flag is optional; emoji characters are silently skipped when
// omitted.
//
// # Obtaining Noto Emoji PNGs (Apache 2.0)
//
//	git clone --depth 1 https://github.com/googlefonts/noto-emoji.git
//	# use noto-emoji/png/128 as -emoji value
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/emoji"
)

func main() {
	fontPath := flag.String("font", "/Library/Fonts/Lato-Medium.ttf", "path to regular TTF or OTF font")
	boldPath := flag.String("bold", "/Library/Fonts/Lato-Black.ttf", "path to bold TTF or OTF font (optional)")
	emojiDir := flag.String("emoji", "", "directory with Noto Emoji PNG files (optional)")
	outPath := flag.String("out", "output.pdf", "output PDF file path")
	flag.Parse()

	// ── Emoji resolver ────────────────────────────────────────────────────
	var resolver emoji.Resolver
	if *emojiDir != "" {
		resolver = &emoji.NotoResolver{Dir: *emojiDir}
	}

	// ── Create document ───────────────────────────────────────────────────
	doc, err := pdf.New(pdf.Config{
		PageSize:         pdf.PageSizeA4,
		EmojiResolver:    resolver,
		DefaultFontSize:  12,
		LineHeightFactor: 1.4,
	})
	if err != nil {
		log.Fatalf("create document: %v", err)
	}

	// ── Register fonts ────────────────────────────────────────────────────
	if _, err := os.Stat(*fontPath); err != nil {
		log.Fatalf("regular font not found at %q\nUse -font to specify a TTF or OTF file.", *fontPath)
	}
	if err := doc.RegisterFont("regular", *fontPath); err != nil {
		log.Fatalf("register regular font: %v", err)
	}

	hasBold := false
	if _, err := os.Stat(*boldPath); err == nil {
		if err := doc.RegisterFont("bold", *boldPath); err != nil {
			log.Printf("warning: register bold font: %v (continuing without bold)", err)
		} else {
			hasBold = true
		}
	}

	// ── Layout constants ──────────────────────────────────────────────────
	const (
		marginLeft  = 60.0
		marginRight = 60.0
		headerY     = 18.0 // top of header text
		footerY     = 820.0 // top of footer text (near bottom of A4)
	)
	pageW := doc.PageWidth()
	contentW := pageW - marginLeft - marginRight

	// headerBoxH is the full height of the header band (text + padding).
	const headerBoxH = 28.0
	// footerBoxH is the full height of the footer band.
	const footerBoxH = 24.0

	// ── Header: title + chapter indicator + border ─────────────────────────
	// The header callback is registered before Build so that it is active
	// during the rendering pass.
	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
		if info.Number == 1 {
			return // no header on the cover page
		}

		// ── Border around the header band ──────────────────────────────
		// Solid navy top/bottom, no left/right sides.
		topLine := &pdf.BorderSpec{
			Thickness: 1.5,
			Color:     pdf.ColorNavy,
			Pattern:   pdf.PatternSolid,
		}
		bottomLine := &pdf.BorderSpec{
			Thickness: 0.5,
			Color:     pdf.ColorLightGray,
			Pattern:   pdf.PatternDashed,
		}
		d.DrawBorder(marginLeft, headerY-4, contentW, headerBoxH, pdf.Border{ //nolint
			Top:    topLine,
			Bottom: bottomLine,
		})

		// ── Header text ────────────────────────────────────────────────
		if err := d.SetFont("regular", 8); err != nil {
			return
		}
		d.SetTextColor(100, 100, 100)
		d.WriteLine("Nautilus PDF Library — Demo Document", marginLeft, headerY+6) //nolint

		chapter := fmt.Sprintf("Chapter %d", info.Number-1)
		w, _ := d.MeasureText(chapter)
		d.WriteLine(chapter, pageW-marginRight-w, headerY+6) //nolint
	})

	// ── Footer: page numbers + border ─────────────────────────────────────
	doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
		// ── Border above the footer band ───────────────────────────────
		// The pattern varies per page to demonstrate all styles.
		patterns := []pdf.BorderPattern{
			pdf.PatternSolid,
			pdf.PatternDashed,
			pdf.PatternDotted,
			pdf.PatternDashDot,
		}
		pat := patterns[(info.Number-1)%len(patterns)]

		topSpec := &pdf.BorderSpec{
			Thickness: 0.75,
			Color:     pdf.ColorGray,
			Pattern:   pat,
		}
		// Thicker solid left accent line in navy.
		leftSpec := &pdf.BorderSpec{
			Thickness: 2,
			Color:     pdf.ColorNavy,
			Pattern:   pdf.PatternSolid,
		}
		d.DrawBorder(marginLeft, footerY-6, contentW, footerBoxH, pdf.Border{ //nolint
			Top:  topSpec,
			Left: leftSpec,
		})

		// ── Footer text ────────────────────────────────────────────────
		if err := d.SetFont("regular", 8); err != nil {
			return
		}
		d.SetTextColor(120, 120, 120)

		label := fmt.Sprintf("Page %d of %d", info.Number, info.Total)
		w, _ := d.MeasureText(label)
		d.WriteLine(label, (pageW-w)/2, footerY+4) //nolint

		d.WriteLine("nautilus example", marginLeft+8, footerY+4) //nolint

		right := "2026"
		rw, _ := d.MeasureText(right)
		d.WriteLine(right, pageW-marginRight-rw, footerY+4) //nolint
	})

	// ── Build: two-pass so footer knows Total pages ───────────────────────
	doc.Build(func() {

		// ── Page 1: Cover ─────────────────────────────────────────────────
		doc.AddPage()

		setFont := func(name string, size float64) {
			if name == "bold" && !hasBold {
				name = "regular"
			}
			doc.SetFont(name, size) //nolint
		}

		// Decorative border box around the cover title area.
		// All four sides visible with different patterns to showcase the API.
		doc.DrawBorder(marginLeft, 110, contentW, 80, pdf.Border{ //nolint
			Top:    &pdf.BorderSpec{Thickness: 3, Color: pdf.ColorNavy, Pattern: pdf.PatternSolid},
			Bottom: &pdf.BorderSpec{Thickness: 3, Color: pdf.ColorNavy, Pattern: pdf.PatternSolid},
			Left:   &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorLightGray, Pattern: pdf.PatternDotted},
			Right:  &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorLightGray, Pattern: pdf.PatternDotted},
		})

		// Large title
		setFont("bold", 32)
		doc.SetTextColor(20, 20, 100)
		doc.WriteLine("Nautilus PDF", marginLeft, 130) //nolint

		setFont("regular", 16)
		doc.SetTextColor(60, 60, 60)
		doc.WriteLine("Pure Go document generation library", marginLeft, 178) //nolint

		// Feature bullets with emoji
		setFont("regular", 12)
		doc.SetTextColor(40, 40, 40)
		bullets := []string{
			"📄  A3 · A4 · A5 · Letter · Legal paper formats",
			"🔤  TTF and OTF font support",
			"🌍  Full Unicode — CJK, Arabic, accented Latin, Cyrillic",
			"😀  Emoji via inline PNG substitution (Noto Emoji)",
			"📐  Automatic word wrapping",
			"🔢  Headers and footers with page numbering",
		}
		y := 230.0
		for _, b := range bullets {
			y, _ = doc.WriteText(b, marginLeft, y, contentW)
		}

		// ── Page 2: Unicode samples ───────────────────────────────────────
		doc.AddPage()

		contentY := 60.0 // below header

		section := func(title string, startY float64) float64 {
			setFont("bold", 13)
			doc.SetTextColor(20, 20, 100)
			y, _ := doc.WriteText(title, marginLeft, startY, contentW)
			return y + 4
		}

		body := func(text string, startY float64) float64 {
			setFont("regular", 11)
			doc.SetTextColor(40, 40, 40)
			y, _ := doc.WriteText(text, marginLeft, startY, contentW)
			return y + 6
		}

		contentY = section("Unicode text samples", contentY)
		contentY = body("Latin extended: café résumé naïve Ångström", contentY)
		contentY = body("CJK: こんにちは世界  你好世界  안녕하세요", contentY)
		contentY = body("Cyrillic: Привет мир", contentY)
		contentY = body("Greek: Γεια σου κόσμε", contentY)
		contentY = body("Arabic: مرحبا بالعالم", contentY)

		contentY += 8
		contentY = section("Font size scale", contentY)
		for _, size := range []float64{8, 10, 12, 14, 18, 24} {
			setFont("regular", size)
			doc.SetTextColor(40, 40, 40)
			doc.WriteText(fmt.Sprintf("%.0f pt — The quick brown fox jumps over the lazy dog", size), //nolint
				marginLeft, contentY, contentW)
			contentY += size*1.4 + 2
		}

		// ── Page 3: Word wrap demo ────────────────────────────────────────
		doc.AddPage()

		contentY = 60.0
		contentY = section("Automatic word wrapping", contentY)
		contentY = body(fmt.Sprintf(
			"The paragraph below is wrapped at %.0f pt (content width = page width − margins). "+
				"The library splits at word boundaries and honours explicit \\n newlines.\n\n"+
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor "+
				"incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud "+
				"exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure "+
				"dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
			contentW,
		), contentY)

		contentY += 12
		contentY = section("Emoji inline with text", contentY)
		contentY = body("Emoji are replaced with PNG images at the current font size:\n"+
			"Smiling: 😀  Waving: 👋  Earth: 🌍  Party: 🎉  Heart: ❤️  Thumbs: 👍",
			contentY)

		// ── Page 4: Border showcase ───────────────────────────────────────
		doc.AddPage()

		contentY = 60.0
		contentY = section("Border styles", contentY)
		contentY = body("DrawBorder supports five patterns, per-side specs,\nand any RGB colour.", contentY)
		contentY += 4

		type borderDemo struct {
			label  string
			border pdf.Border
		}
		demos := []borderDemo{
			{
				"Solid — uniform 1.5 pt navy",
				pdf.NewUniformBorder(pdf.BorderSpec{
					Thickness: 1.5, Color: pdf.ColorNavy, Pattern: pdf.PatternSolid,
				}),
			},
			{
				"Dashed — 0.75 pt gray",
				pdf.NewUniformBorder(pdf.BorderSpec{
					Thickness: 0.75, Color: pdf.ColorGray, Pattern: pdf.PatternDashed,
				}),
			},
			{
				"Dotted — 1 pt orange",
				pdf.NewUniformBorder(pdf.BorderSpec{
					Thickness: 1, Color: pdf.ColorOrange, Pattern: pdf.PatternDotted,
				}),
			},
			{
				"DashDot — 1 pt green",
				pdf.NewUniformBorder(pdf.BorderSpec{
					Thickness: 1, Color: pdf.ColorGreen, Pattern: pdf.PatternDashDot,
				}),
			},
			{
				"Custom — dash array [12 4 4 4] 2 pt red",
				pdf.NewUniformBorder(pdf.BorderSpec{
					Thickness: 2, Color: pdf.ColorRed, Pattern: pdf.PatternCustom,
					DashArray: []float64{12, 4, 4, 4},
				}),
			},
			{
				"Mixed sides — thick top (navy solid), thin bottom (gray dashed)",
				pdf.Border{
					Top:    &pdf.BorderSpec{Thickness: 2.5, Color: pdf.ColorNavy, Pattern: pdf.PatternSolid},
					Bottom: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorGray, Pattern: pdf.PatternDashed},
				},
			},
		}

		boxH := 28.0
		gap := 10.0
		for _, demo := range demos {
			doc.DrawBorder(marginLeft, contentY, contentW, boxH, demo.border) //nolint
			setFont("regular", 9)
			doc.SetTextColor(60, 60, 60)
			doc.WriteLine(demo.label, marginLeft+8, contentY+9) //nolint
			contentY += boxH + gap
		}

		// ── Page 5: Frame / minipage layout ───────────────────────────────
		doc.AddPage()

		contentY = 60.0
		contentY = section("Frames — positioned content boxes", contentY)
		contentY = body("Each frame is a positioned rectangular region with its own\n"+
			"content area, padding, border and optional background fill.", contentY)
		contentY += 8

		// ── Callout box (fixed height, background fill) ──────────────
		calloutBg := pdf.Color{R: 235, G: 245, B: 255}
		callout := doc.NewFrame(pdf.FrameConfig{
			X: marginLeft, Y: contentY, Width: contentW, Height: 56,
			Background: &calloutBg,
			Border: pdf.Border{
				Left: &pdf.BorderSpec{Thickness: 4, Color: pdf.ColorNavy, Pattern: pdf.PatternSolid},
			},
			Padding: pdf.Padding{Top: 8, Right: 12, Bottom: 8, Left: 16},
		})
		callout.SetFont("regular", 10) //nolint
		callout.SetTextColor(20, 20, 80)
		callout.WriteText("ℹ️  Frame with background fill and left accent border.\n" + //nolint
			"Background is drawn before text so the fill stays behind the content.")
		callout.Close() //nolint
		contentY += 56 + 14

		// ── Two-column layout ─────────────────────────────────────────
		colGutter := 10.0
		colW := (contentW - colGutter) / 2

		leftBody := "The quick brown fox jumps over the lazy dog.\n\n" +
			"Frames allow independent layout regions on the same page, " +
			"useful for multi-column reports, sidebars, and callout boxes."
		rightBody := "Jeder Frame hat seine eigene Breite, sein Padding und seinen Border.\n\n" +
			"Le renard brun rapide saute par-dessus le chien paresseux.\n\n" +
			"De snelle bruine vos springt over de luie hond."

		for i, body := range []string{leftBody, rightBody} {
			xPos := marginLeft + float64(i)*(colW+colGutter)
			f := doc.NewFrame(pdf.FrameConfig{
				X: xPos, Y: contentY, Width: colW,
				Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray, Pattern: pdf.PatternSolid}),
				Padding: pdf.HorizontalPadding(8, 8),
			})
			f.SetFont("regular", 10) //nolint
			f.SetTextColor(40, 40, 40)
			// Column heading
			f.SetFont("regular", 10) //nolint
			f.SetTextColor(20, 20, 100)
			if i == 0 {
				f.WriteText("Column A") //nolint
			} else {
				f.WriteText("Column B") //nolint
			}
			// Separator below heading
			f.DrawInnerBorder(0, f.CurrentY()-contentY-4, colW, f.CurrentY()-contentY+4, //nolint
				pdf.Border{Bottom: &pdf.BorderSpec{Thickness: 0.4, Color: pdf.ColorLightGray}})
			f.Advance(6)
			f.SetFont("regular", 9) //nolint
			f.SetTextColor(50, 50, 50)
			f.WriteText(body) //nolint
			f.Close() //nolint
		}
		contentY += 130

		// ── Nested frames ────────────────────────────────────────────
		outer := doc.NewFrame(pdf.FrameConfig{
			X: marginLeft, Y: contentY, Width: contentW, Height: 110,
			Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1, Color: pdf.ColorDarkGray, Pattern: pdf.PatternDashed}),
			Padding: pdf.UniformPadding(10),
		})
		outer.SetFont("regular", 9) //nolint
		outer.SetTextColor(80, 80, 80)
		outer.WriteText("Outer frame (dashed border)") //nolint
		outer.Advance(4)

		// Inner frame nested inside outer.
		innerBg := pdf.Color{R: 245, G: 245, B: 220}
		inner := doc.NewFrame(pdf.FrameConfig{
			X:          marginLeft + 10 + 20,
			Y:          outer.CurrentY(),
			Width:      contentW - 20 - 60,
			Height:     48,
			Background: &innerBg,
			Border:     pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1, Color: pdf.ColorOrange, Pattern: pdf.PatternSolid}),
			Padding:    pdf.UniformPadding(7),
		})
		inner.SetFont("regular", 9) //nolint
		inner.SetTextColor(80, 60, 0)
		inner.WriteText("Inner frame (solid orange border, cream background)\n" + //nolint
			"Frames can be positioned inside other frames for complex layouts.")
		inner.Close() //nolint
		outer.Close() //nolint

		// ── Page 6: Table demo ─────────────────────────────────────────
		doc.AddPage()

		contentY = 60.0
		contentY = section("Tables — grid-based layout", contentY)
		contentY = body("Tables support colspan, rowspan, per-cell styling,\n"+
			"background colours, and automatic page overflow.", contentY)
		contentY += 8

		navy := pdf.ColorNavy
		white := pdf.ColorWhite
		lightBlue := pdf.Color{R: 230, G: 240, B: 255}
		lightGray := pdf.ColorLightGray

		cellBorder := pdf.NewUniformBorder(pdf.BorderSpec{
			Thickness: 0.5,
			Color:     pdf.ColorLightGray,
			Pattern:   pdf.PatternSolid,
		})

		tbl := doc.NewTable(pdf.TableConfig{
			X: marginLeft, Y: contentY,
			ColWidths: []float64{120, 200, 80, 75},
			Border: pdf.NewUniformBorder(pdf.BorderSpec{
				Thickness: 1.5,
				Color:     pdf.ColorNavy,
				Pattern:   pdf.PatternSolid,
			}),
			DefaultCellStyle: pdf.CellStyle{
				Padding:  pdf.Padding{Top: 5, Right: 8, Bottom: 5, Left: 8},
				Border:   cellBorder,
				FontName: "regular",
				FontSize: 10,
			},
			PageBottom:    doc.PageHeight() - 60,
			ContinuationY: 60,
		})

		// Header row — navy background with white text.
		tbl.AddRow(pdf.Row{
			Height: 24,
			Cells: []pdf.Cell{
				{Text: "Product", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold"}},
				{Text: "Description", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold"}},
				{Text: "Qty", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold"}},
				{Text: "Price", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold"}},
			},
		})

		// Group header row with colspan.
		tbl.AddRow(pdf.Row{
			Height: 20,
			Cells: []pdf.Cell{
				{Text: "Electronics", ColSpan: 4, Style: pdf.CellStyle{
					Background: &lightBlue, FontName: "bold", FontSize: 9}},
			},
		})

		// Data rows — first two share a rowspan in the Product column.
		orange := pdf.ColorOrange
		tbl.AddRow(pdf.Row{
			Cells: []pdf.Cell{
				{Text: "Wireless\nMouse", RowSpan: 2, Style: pdf.CellStyle{Background: &lightGray}},
				{Text: "Ergonomic 3-button wireless mouse, 2.4 GHz receiver included"},
				{Text: "50"},
				{Text: "€24.99"},
			},
		})
		tbl.AddRow(pdf.Row{
			Cells: []pdf.Cell{
				{Text: "Same item, bundle of 5 units — bulk discount applies"},
				{Text: "5"},
				{Text: "€109.99", Style: pdf.CellStyle{TextColor: &orange}},
			},
		})

		tbl.AddRow(pdf.Row{
			Cells: []pdf.Cell{
				{Text: "USB-C Hub"},
				{Text: "7-in-1 hub: HDMI, 3×USB-A, SD card, PD charging"},
				{Text: "120"},
				{Text: "€39.95"},
			},
		})

		// Second group header.
		tbl.AddRow(pdf.Row{
			Height: 20,
			Cells: []pdf.Cell{
				{Text: "Accessories", ColSpan: 4, Style: pdf.CellStyle{
					Background: &lightBlue, FontName: "bold", FontSize: 9}},
			},
		})

		tbl.AddRow(pdf.Row{
			Cells: []pdf.Cell{
				{Text: "Desk Mat"},
				{Text: "Extra-large anti-slip desk mat, 90×45 cm, waterproof surface"},
				{Text: "200"},
				{Text: "€19.50"},
			},
		})
		tbl.AddRow(pdf.Row{
			Cells: []pdf.Cell{
				{Text: "Cable Clips"},
				{Text: "Self-adhesive cable management clips, pack of 20"},
				{Text: "500"},
				{Text: "€6.99"},
			},
		})

		// Footer / totals row with colspan and right-aligned price.
		tbl.AddRow(pdf.Row{
			Height: 24,
			Cells: []pdf.Cell{
				{Text: "Total items: 875", ColSpan: 2, Style: pdf.CellStyle{
					FontName: "bold", Background: &lightBlue}},
				{Text: "875", Style: pdf.CellStyle{
					FontName: "bold", Background: &lightBlue,
					HAlign: pdf.HAlignRight}},
				{Text: "—", Style: pdf.CellStyle{
					FontName: "bold", Background: &lightBlue,
					HAlign: pdf.HAlignCenter}},
			},
		})

		contentY += 230 // approx height of table above + gap

		// ── Alignment showcase ────────────────────────────────────────
		contentY = section("Cell alignment — HAlign & VAlign", contentY)
		contentY = body("Cells can align text horizontally (left / centre / right)\n"+
			"and vertically (top / middle / bottom).", contentY)
		contentY += 6

		alignTbl := doc.NewTable(pdf.TableConfig{
			X: marginLeft, Y: contentY,
			ColWidths: []float64{120, 120, 120, 115},
			Border: pdf.NewUniformBorder(pdf.BorderSpec{
				Thickness: 1, Color: pdf.ColorNavy,
			}),
			DefaultCellStyle: pdf.CellStyle{
				Padding:  pdf.Padding{Top: 4, Right: 8, Bottom: 4, Left: 8},
				Border:   cellBorder,
				FontName: "regular",
				FontSize: 10,
			},
			PageBottom: doc.PageHeight() - 60,
		})

		// Header row.
		alignTbl.AddRow(pdf.Row{
			Height: 22,
			Cells: []pdf.Cell{
				{Text: "HAlign →", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold"}},
				{Text: "Left", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold",
					HAlign: pdf.HAlignCenter}},
				{Text: "Centre", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold",
					HAlign: pdf.HAlignCenter}},
				{Text: "Right", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold",
					HAlign: pdf.HAlignCenter}},
			},
		})

		// VAlign: Top row.
		alignTbl.AddRow(pdf.Row{
			Height: 50,
			Cells: []pdf.Cell{
				{Text: "VAlign: Top", Style: pdf.CellStyle{
					FontName: "bold", VAlign: pdf.VAlignMiddle}},
				{Text: "Top\nLeft", Style: pdf.CellStyle{
					VAlign: pdf.VAlignTop, HAlign: pdf.HAlignLeft}},
				{Text: "Top\nCentre", Style: pdf.CellStyle{
					VAlign: pdf.VAlignTop, HAlign: pdf.HAlignCenter}},
				{Text: "Top\nRight", Style: pdf.CellStyle{
					VAlign: pdf.VAlignTop, HAlign: pdf.HAlignRight}},
			},
		})

		// VAlign: Middle row.
		alignTbl.AddRow(pdf.Row{
			Height: 50,
			Cells: []pdf.Cell{
				{Text: "VAlign: Middle", Style: pdf.CellStyle{
					FontName: "bold", VAlign: pdf.VAlignMiddle}},
				{Text: "Mid\nLeft", Style: pdf.CellStyle{
					VAlign: pdf.VAlignMiddle, HAlign: pdf.HAlignLeft}},
				{Text: "Mid\nCentre", Style: pdf.CellStyle{
					VAlign: pdf.VAlignMiddle, HAlign: pdf.HAlignCenter}},
				{Text: "Mid\nRight", Style: pdf.CellStyle{
					VAlign: pdf.VAlignMiddle, HAlign: pdf.HAlignRight}},
			},
		})

		// VAlign: Bottom row.
		alignTbl.AddRow(pdf.Row{
			Height: 50,
			Cells: []pdf.Cell{
				{Text: "VAlign: Bottom", Style: pdf.CellStyle{
					FontName: "bold", VAlign: pdf.VAlignMiddle}},
				{Text: "Bot\nLeft", Style: pdf.CellStyle{
					VAlign: pdf.VAlignBottom, HAlign: pdf.HAlignLeft}},
				{Text: "Bot\nCentre", Style: pdf.CellStyle{
					VAlign: pdf.VAlignBottom, HAlign: pdf.HAlignCenter}},
				{Text: "Bot\nRight", Style: pdf.CellStyle{
					VAlign: pdf.VAlignBottom, HAlign: pdf.HAlignRight}},
			},
		})

		if err := alignTbl.Draw(); err != nil {
			log.Fatalf("draw alignment table: %v", err)
		}

		if err := tbl.Draw(); err != nil {
			log.Fatalf("draw table: %v", err)
		}

		// ── Page 7+: Long table that overflows across pages ────────────────
		doc.AddPage()
		contentY = 60.0
		contentY = section("Page-overflow table", contentY)
		contentY = body("The table below contains enough rows to force it\n"+
			"to continue on the next page, demonstrating page overflow.", contentY)
		contentY += 8

		longTbl := doc.NewTable(pdf.TableConfig{
			X: marginLeft, Y: contentY,
			ColWidths: []float64{100, 240, 135},
			Border: pdf.NewUniformBorder(pdf.BorderSpec{
				Thickness: 1,
				Color:     pdf.ColorNavy,
				Pattern:   pdf.PatternSolid,
			}),
			DefaultCellStyle: pdf.CellStyle{
				Padding:  pdf.Padding{Top: 5, Right: 8, Bottom: 5, Left: 8},
				Border:   cellBorder,
				FontName: "regular",
				FontSize: 10,
			},
			PageBottom:    doc.PageHeight() - 60,
			ContinuationY: 60,
		})

		// Header.
		longTbl.AddRow(pdf.Row{
			Height: 22,
			Cells: []pdf.Cell{
				{Text: "ID", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold"}},
				{Text: "Description", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold"}},
				{Text: "Status", Style: pdf.CellStyle{
					Background: &navy, TextColor: &white, FontName: "bold"}},
			},
		})

		statuses := []string{"Active", "Pending", "Inactive", "Review"}
		for i := range 30 {
			status := statuses[i%len(statuses)]
			var bg *pdf.Color
			if i%2 == 0 {
				bg = &lightBlue
			}
			longTbl.AddRow(pdf.Row{
				Background: bg,
				Cells: []pdf.Cell{
					{Text: fmt.Sprintf("%04d", i+1)},
					{Text: fmt.Sprintf("Item number %d — auto-generated row for overflow demo", i+1)},
					{Text: status},
				},
			})
		}
		if err := longTbl.Draw(); err != nil {
			log.Fatalf("draw long table: %v", err)
		}
	})

	// ── Write output ──────────────────────────────────────────────────────
	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save PDF: %v", err)
	}
	fmt.Printf("PDF written to %s  (%d pages)\n", *outPath, doc.PageCount())
}
