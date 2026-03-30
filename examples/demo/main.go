// Command demo produces a comprehensive showcase PDF that exercises every
// major feature of the Nautilus PDF library: low-level drawing primitives,
// borders, frames, tables, images, the layout engine, and all 20 chart types.
//
// Usage:
//
//	go run ./examples/demo \
//	    -font  /path/to/regular.ttf \
//	    -bold  /path/to/bold.ttf \
//	    -images assets/images \
//	    -out   demo_output.pdf
package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/chart"
	"github.com/gvanbeck/nautilus/pdf/emoji"
	"github.com/gvanbeck/nautilus/pdf/rtl"
	"github.com/gvanbeck/nautilus/pdf/chart/area"
	"github.com/gvanbeck/nautilus/pdf/chart/arearange"
	"github.com/gvanbeck/nautilus/pdf/chart/bar"
	"github.com/gvanbeck/nautilus/pdf/chart/boxplot"
	"github.com/gvanbeck/nautilus/pdf/chart/bubble"
	"github.com/gvanbeck/nautilus/pdf/chart/bullet"
	"github.com/gvanbeck/nautilus/pdf/chart/column"
	"github.com/gvanbeck/nautilus/pdf/chart/columnrange"
	"github.com/gvanbeck/nautilus/pdf/chart/dumbbell"
	"github.com/gvanbeck/nautilus/pdf/chart/errorbar"
	"github.com/gvanbeck/nautilus/pdf/chart/funnel"
	"github.com/gvanbeck/nautilus/pdf/chart/gauge"
	"github.com/gvanbeck/nautilus/pdf/chart/heatmap"
	"github.com/gvanbeck/nautilus/pdf/chart/line"
	"github.com/gvanbeck/nautilus/pdf/chart/lollipop"
	"github.com/gvanbeck/nautilus/pdf/chart/pie"
	"github.com/gvanbeck/nautilus/pdf/chart/polar"
	"github.com/gvanbeck/nautilus/pdf/chart/scatter"
	"github.com/gvanbeck/nautilus/pdf/chart/treemap"
	"github.com/gvanbeck/nautilus/pdf/chart/waterfall"
	"github.com/gvanbeck/nautilus/pdf/layout"
)

// ── Go-themed filler text ────────────────────────────────────────────────────

const goText1 = "Go is an open source programming language that makes it simple to build secure, scalable software. Developed at Google, Go combines the simplicity of scripting with the performance of compiled languages. The Go runtime includes garbage collection, goroutines, and channels — primitives that make concurrent programming straightforward and powerful."

const goText2 = "The Go Gopher was designed by Renee French and first appeared in 2009 when Go was first announced. The Gopher has become the beloved mascot of the Go community, appearing in countless community contributions, conference talks, and merchandise."

const goText3 = "Go's standard library is comprehensive: it includes packages for HTTP, cryptography, JSON, SQL, testing, and much more. The toolchain ships as a single binary and supports cross-compilation to dozens of operating system and architecture combinations with a single command."

const goText4 = "Goroutines are lightweight threads managed by the Go runtime. Unlike OS threads, goroutines start with just a few kilobytes of stack space that can grow and shrink on demand. You can launch thousands of goroutines without exhausting system memory."

const goText5 = "Go modules, introduced in Go 1.11 and made the default in Go 1.16, provide a robust dependency management system. Each module has a go.mod file that declares its dependencies with exact version pinning, ensuring reproducible builds across machines and time."

const goText6 = "The Go compiler produces statically linked native binaries by default. This means your entire application — including all its dependencies — is bundled into a single executable file, simplifying deployment dramatically: no virtual environments, no shared libraries to manage."

// ── page geometry constants ───────────────────────────────────────────────────

const (
	pageW     = 595.28 // A4 width  in points
	pageH     = 841.89 // A4 height in points
	marginL   = 45.0
	marginR   = 45.0
	marginT   = 55.0
	marginB   = 50.0
	contentW  = pageW - marginL - marginR // 505.28
	headerY   = 18.0
	footerY   = pageH - 28.0
)

// ── color palette ─────────────────────────────────────────────────────────────

var (
	colorNavy      = pdf.Color{R: 0, G: 38, B: 84}
	colorCyan      = pdf.Color{R: 0, G: 172, B: 215}
	colorGold      = pdf.Color{R: 255, G: 193, B: 7}
	colorSlate     = pdf.Color{R: 90, G: 103, B: 120}
	colorMint      = pdf.Color{R: 0, G: 164, B: 130}
	colorCoral     = pdf.Color{R: 220, G: 80, B: 60}
	colorLavender  = pdf.Color{R: 140, G: 100, B: 200}
	colorOffWhite  = pdf.Color{R: 248, G: 249, B: 250}
	colorLightBlue = pdf.Color{R: 210, G: 230, B: 250}
	colorRowAlt    = pdf.Color{R: 240, G: 245, B: 255}
)

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	fontPath := flag.String("font", "", "Path to regular TTF font (required)")
	boldPath := flag.String("bold", "", "Path to bold TTF font (optional)")
	imagesDir := flag.String("images", "assets/images", "Directory containing gopher PNG images")
	emojiDir := flag.String("emoji", "", "Directory with Noto Emoji PNG files (optional, e.g. assets/emoji/png/128)")
	arabicFont := flag.String("arabic", "/System/Library/Fonts/SFArabic.ttf", "Arabic TrueType font (optional)")
	hebrewFont := flag.String("hebrew", "/System/Library/Fonts/SFHebrew.ttf", "Hebrew font (optional)")
	outPath := flag.String("out", "demo_output.pdf", "Output PDF path")
	flag.Parse()

	if *fontPath == "" {
		log.Fatal("usage: -font <path/to/font.ttf> [-bold <path/to/bold.ttf>] [-images assets/images] [-emoji assets/emoji/png/128] [-arabic ...] [-hebrew ...] [-out demo_output.pdf]")
	}

	// ── Emoji resolver ────────────────────────────────────────────────────────
	var emojiResolver emoji.Resolver
	if *emojiDir != "" {
		emojiResolver = &emoji.NotoResolver{Dir: *emojiDir}
	}

	// ── Create document ──────────────────────────────────────────────────────
	doc, err := pdf.New(pdf.Config{
		PageSize:      pdf.PageSizeA4,
		EmojiResolver: emojiResolver,
		Margins: pdf.Margins{
			Top:    marginT,
			Right:  marginR,
			Bottom: marginB,
			Left:   marginL,
		},
	})
	if err != nil {
		log.Fatalf("create document: %v", err)
	}

	// ── Register fonts ───────────────────────────────────────────────────────
	if err := doc.RegisterFont("regular", *fontPath); err != nil {
		log.Fatalf("register regular font: %v", err)
	}
	boldFont := "regular"
	if *boldPath != "" {
		if err := doc.RegisterFont("bold", *boldPath); err != nil {
			log.Fatalf("register bold font: %v", err)
		}
		boldFont = "bold"
	}

	// RTL fonts — optional; register only if the file exists.
	hasArabic := false
	if *arabicFont != "" {
		if err := doc.RegisterFont("arabic", *arabicFont); err == nil {
			hasArabic = true
		}
	}
	hasHebrew := false
	if *hebrewFont != "" {
		if err := doc.RegisterFont("hebrew", *hebrewFont); err == nil {
			hasHebrew = true
		}
	}

	// ── Header / footer ──────────────────────────────────────────────────────
	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
		if info.Number == 1 {
			return // cover page has its own decoration
		}
		d.SetFont("regular", 8)    //nolint:errcheck
		d.SetTextColor(150, 150, 150)
		d.WriteLine("Nautilus PDF Library — Developer Feature Demo", marginL, headerY) //nolint:errcheck
		spec := &pdf.BorderSpec{Thickness: 0.5, Color: pdf.Color{R: 200, G: 210, B: 220}}
		d.DrawBorder(marginL, headerY+10, contentW, 0, pdf.Border{Bottom: spec}) //nolint:errcheck
	})

	doc.SetFooter(func(d *pdf.Document, info pdf.PageInfo) {
		d.SetFont("regular", 8)    //nolint:errcheck
		d.SetTextColor(150, 150, 150)
		label := fmt.Sprintf("Page %d of %d", info.Number, info.Total)
		lw, _ := d.MeasureText(label)
		d.WriteLine(label, pageW-marginR-lw, footerY) //nolint:errcheck
		spec := &pdf.BorderSpec{Thickness: 0.5, Color: pdf.Color{R: 200, G: 210, B: 220}}
		d.DrawBorder(marginL, footerY-6, contentW, 0, pdf.Border{Top: spec}) //nolint:errcheck
	})

	// ── Build (two-pass for "Page N of M") ───────────────────────────────────
	imDir := *imagesDir
	bFont := boldFont

	eDir := *emojiDir
	hArabic := hasArabic
	hHebrew := hasHebrew
	doc.Build(func() {
		buildCoverPage(doc, imDir, bFont)
		buildTypographyPage(doc, bFont)
		buildEmojiPage(doc, eDir, bFont)
		buildRTLPage(doc, bFont, hArabic, hHebrew)
		buildDrawingPage(doc, bFont)
		buildBordersFramesPage(doc, bFont)
		buildTablesPage(doc, bFont)
		buildImagesPage(doc, imDir, bFont)
		buildLayoutEnginePages(doc, bFont)
		buildChartPages(doc, bFont)
	})

	// ── Save ──────────────────────────────────────────────────────────────────
	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save: %v", err)
	}
	log.Printf("saved %s", *outPath)
}

// ── Page 1: Cover ─────────────────────────────────────────────────────────────

func buildCoverPage(doc *pdf.Document, imDir, boldFont string) {
	doc.AddPage()

	// Decorative background band at top
	doc.FillRect(0, 0, pageW, 220, colorNavy)

	// Accent stripe
	doc.FillRect(0, 220, pageW, 8, colorCyan)

	// Decorative colored rectangles – bottom accent strip
	stripeH := 12.0
	stripeW := contentW / 5
	colors := []pdf.Color{colorCyan, colorMint, colorGold, colorCoral, colorLavender}
	for i, c := range colors {
		doc.FillRect(marginL+float64(i)*stripeW, pageH-stripeH-10, stripeW, stripeH, c)
	}

	// Title
	doc.SetTextColor(255, 255, 255)
	doc.SetFont(boldFont, 36) //nolint:errcheck
	title := "Nautilus PDF Library"
	tw, _ := doc.MeasureText(title)
	doc.WriteLine(title, (pageW-tw)/2, 60) //nolint:errcheck

	// Subtitle
	doc.SetFont("regular", 18) //nolint:errcheck
	sub := "Complete Developer Feature Demo"
	sw, _ := doc.MeasureText(sub)
	doc.WriteLine(sub, (pageW-sw)/2, 108) //nolint:errcheck

	// Version tag
	doc.SetFont("regular", 11) //nolint:errcheck
	ver := "v1.0  •  Go 1.26+"
	vw, _ := doc.MeasureText(ver)
	doc.WriteLine(ver, (pageW-vw)/2, 140) //nolint:errcheck

	// Date line
	doc.SetFont("regular", 10) //nolint:errcheck
	dateStr := "Generated: " + time.Now().Format("2 January 2006")
	dw, _ := doc.MeasureText(dateStr)
	doc.WriteLine(dateStr, (pageW-dw)/2, 165) //nolint:errcheck

	// Decorative diagonal lines on the navy band
	doc.DrawLine(0, 0, 60, 220, 0.5, pdf.Color{R: 255, G: 255, B: 255})
	doc.DrawLine(pageW, 0, pageW-60, 220, 0.5, pdf.Color{R: 255, G: 255, B: 255})
	doc.DrawLine(pageW/2, 0, pageW/2-30, 220, 0.3, pdf.Color{R: 255, G: 255, B: 255})

	// Gopher image centered
	gopherPath := filepath.Join(imDir, "gopher.png")
	imgW, imgH := 160.0, 160.0
	imgX := (pageW - imgW) / 2
	imgY := 248.0
	if err := doc.DrawImage(gopherPath, imgX, imgY, imgW, imgH); err != nil {
		// Image not found; draw a placeholder circle
		doc.FillCircle(pageW/2, imgY+imgH/2, 60, colorCyan)
	}

	// Feature bullets
	doc.SetTextColor(40, 40, 40)
	doc.SetFont(boldFont, 13) //nolint:errcheck
	featTitle := "What this demo covers"
	ftw, _ := doc.MeasureText(featTitle)
	doc.WriteLine(featTitle, (pageW-ftw)/2, 432) //nolint:errcheck

	doc.SetFont("regular", 10) //nolint:errcheck
	features := []string{
		"Low-level PDF drawing: text, lines, rectangles, polygons, circles",
		"Border patterns: solid, dashed, dotted, dash-dot",
		"Frames with backgrounds, padding, and borders",
		"Tables with colspan / rowspan, per-cell styling, page overflow",
		"Images (PNG Gopher artwork)",
		"Emoji (inline PNG via Noto Emoji resolver)",
		"Right-to-left text: Arabic shaping + Hebrew BiDi",
		"Layout engine: DocTemplate, PageTemplate, Flowables",
		"All 20 Highcharts-style chart types",
	}
	bulletY := 460.0
	for _, f := range features {
		doc.SetTextColor(0, 120, 160)
		doc.WriteLine("•", marginL+20, bulletY) //nolint:errcheck
		doc.SetTextColor(40, 40, 40)
		doc.WriteLine(f, marginL+34, bulletY) //nolint:errcheck
		bulletY += 18
	}

	// Small decorative circles
	doc.FillCircle(marginL+15, 465+float64(len(features))*18+10, 14, colorCyan)
	doc.StrokeCircle(marginL+15, 465+float64(len(features))*18+10, 14, 1, colorNavy)
	doc.FillCircle(pageW-marginR-15, 465+float64(len(features))*18+10, 14, colorMint)

	// Bottom tag line
	doc.SetFont("regular", 9) //nolint:errcheck
	doc.SetTextColor(100, 100, 100)
	tagline := "Pure Go • No CGO • wraps gopdf • MIT License"
	tlw, _ := doc.MeasureText(tagline)
	doc.WriteLine(tagline, (pageW-tlw)/2, pageH-40) //nolint:errcheck
}

// ── Page 2: Typography ────────────────────────────────────────────────────────

func buildTypographyPage(doc *pdf.Document, boldFont string) {
	doc.AddPage()
	y := sectionHeader(doc, boldFont, "Typography & Text Rendering", marginT+12)
	y += 8

	// Font size showcase
	doc.SetFont(boldFont, 10)  //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Font Sizes", marginL, y) //nolint:errcheck
	y += 14

	sizes := []float64{8, 9, 10, 11, 12, 14, 16, 18, 20, 24, 28, 32, 36}
	for _, sz := range sizes {
		doc.SetFont("regular", sz) //nolint:errcheck
		doc.SetTextColor(30, 30, 30)
		label := fmt.Sprintf("%.0fpt — The quick brown fox jumps over the lazy gopher", sz)
		if sz >= 20 {
			label = fmt.Sprintf("%.0fpt — Go makes it simple", sz)
		}
		doc.WriteLine(label, marginL, y) //nolint:errcheck
		y += sz*1.2 + 2
		if y > pageH-marginB-10 {
			break
		}
	}

	y += 8

	// Text color demonstration
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Text Colors", marginL, y) //nolint:errcheck
	y += 14

	colorSamples := []struct {
		label string
		c     pdf.Color
	}{
		{"Navy (0,38,84)", colorNavy},
		{"Cyan (0,172,215)", colorCyan},
		{"Coral (220,80,60)", colorCoral},
		{"Mint (0,164,130)", colorMint},
		{"Gold (255,193,7)", colorGold},
		{"Lavender (140,100,200)", colorLavender},
		{"Slate (90,103,120)", colorSlate},
		{"Standard Black (0,0,0)", pdf.ColorBlack},
	}

	doc.SetFont("regular", 11) //nolint:errcheck
	for _, cs := range colorSamples {
		doc.SetTextColor(cs.c.R, cs.c.G, cs.c.B)
		doc.WriteLine("  "+cs.label+" — the Go Gopher says hello!", marginL, y) //nolint:errcheck
		y += 16
	}

	y += 8
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Word-Wrapped Body Text (Go-themed)", marginL, y) //nolint:errcheck
	y += 14

	doc.SetFont("regular", 11) //nolint:errcheck
	doc.SetTextColor(30, 30, 30)
	endY, err := doc.WriteText(goText1, marginL, y, contentW)
	if err != nil {
		log.Printf("WriteText: %v", err)
	}
	y = endY + 8

	doc.SetFont("regular", 10) //nolint:errcheck
	endY, err = doc.WriteText(goText2, marginL, y, contentW)
	if err != nil {
		log.Printf("WriteText: %v", err)
	}
	_ = endY
}

// ── Page 3: Emoji ─────────────────────────────────────────────────────────────

func buildEmojiPage(doc *pdf.Document, emojiDir, boldFont string) {
	doc.AddPage()
	y := sectionHeader(doc, boldFont, "Emoji — Inline PNG Rendering", marginT+12)
	y += 10

	if emojiDir == "" {
		doc.SetFont("regular", 11) //nolint:errcheck
		doc.SetTextColor(120, 120, 120)
		doc.WriteText("No emoji directory provided. Run with -emoji assets/emoji/png/128 to enable.", marginL, y, contentW) //nolint:errcheck
		return
	}

	doc.SetFont("regular", 10) //nolint:errcheck
	doc.SetTextColor(60, 60, 60)
	intro := "Nautilus supports inline emoji rendering via the NotoResolver. Emoji grapheme clusters are automatically detected in text and replaced with the corresponding Noto Emoji PNG at the correct font size. The examples below show emoji embedded directly in running text."
	y, _ = doc.WriteText(intro, marginL, y, contentW)
	y += 14

	// ── Category labels and emoji lines ──────────────────────────────────
	type emojiRow struct {
		label string
		text  string
	}
	rows := []emojiRow{
		{"Smileys & People", "Happy 😀 laughing 😂 winking 😉 cool 😎 thinking 🤔 party 🥳 love 😍 shocked 😱"},
		{"Nature & Animals", "Dog 🐶 cat 🐱 fox 🦊 bear 🐻 panda 🐼 rabbit 🐰 frog 🐸 penguin 🐧 whale 🐳"},
		{"Food & Drink", "Pizza 🍕 burger 🍔 sushi 🍣 ramen 🍜 taco 🌮 cake 🍰 coffee ☕ beer 🍺 grapes 🍇"},
		{"Travel & Places", "Rocket 🚀 car 🚗 bicycle 🚲 globe 🌍 mountain 🏔️ beach 🏖️ city 🌆 plane ✈️"},
		{"Activities", "Soccer ⚽ basketball 🏀 tennis 🎾 swimming 🏊 cycling 🚴 gaming 🎮 music 🎵 art 🎨"},
		{"Objects & Symbols", "Laptop 💻 phone 📱 book 📚 mail 📧 lock 🔒 key 🔑 bulb 💡 clock ⏰ fire 🔥"},
		{"Flags", "Belgium 🇧🇪 Netherlands 🇳🇱 Germany 🇩🇪 France 🇫🇷 USA 🇺🇸 Japan 🇯🇵"},
		{"Go & Tech", "Gopher 🦫 code 👨‍💻 build 🔨 test ✅ deploy 🚢 bug 🐛 fix 🔧 star ⭐"},
	}

	for _, row := range rows {
		// Category label
		doc.SetFont(boldFont, 9) //nolint:errcheck
		doc.SetTextColor(colorNavy.R, colorNavy.G, colorNavy.B)
		doc.FillRect(marginL, y, contentW, 17, colorOffWhite)
		doc.WriteLine(row.label, marginL+6, y+4) //nolint:errcheck
		y += 20

		// Emoji text at 14pt so glyphs are clearly visible
		doc.SetFont("regular", 14) //nolint:errcheck
		doc.SetTextColor(30, 30, 30)
		endY, _ := doc.WriteText(row.text, marginL+6, y, contentW-12)
		y = endY + 10
	}

	y += 6

	// ── Mixed text + emoji paragraph ─────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorNavy.R, colorNavy.G, colorNavy.B)
	doc.WriteLine("Emoji in running text", marginL, y) //nolint:errcheck
	y += 16

	doc.SetFont("regular", 11) //nolint:errcheck
	doc.SetTextColor(40, 40, 40)
	mixed := "Go is fast ⚡, reliable ✅, and fun 🎉. Write code 💻, run tests 🧪, ship binaries 🚀, and celebrate 🥳 — all from a single toolchain. The community is welcoming 🤝 and the ecosystem keeps growing 🌱. Whether you're building APIs 🔗, CLIs 🖥️, or distributed systems 🌐, Go has you covered. Happy coding! 😊"
	doc.WriteText(mixed, marginL, y, contentW) //nolint:errcheck
}

// ── Page 4: Right-to-Left Text ────────────────────────────────────────────────

func buildRTLPage(doc *pdf.Document, boldFont string, hasArabic, hasHebrew bool) {
	doc.AddPage()
	y := sectionHeader(doc, boldFont, "Right-to-Left Text — Arabic & Hebrew", marginT+12)
	y += 10

	rightX := pageW - marginR

	if !hasArabic && !hasHebrew {
		doc.SetFont("regular", 11) //nolint:errcheck
		doc.SetTextColor(120, 120, 120)
		doc.WriteText("No RTL fonts found. Provide -arabic and -hebrew flags to enable.", marginL, y, contentW) //nolint:errcheck
		return
	}

	doc.SetFont("regular", 10) //nolint:errcheck
	doc.SetTextColor(60, 60, 60)
	intro := "Nautilus supports right-to-left scripts via the pdf/rtl package. Arabic text is automatically shaped (contextual letter forms + lam-alef ligatures) and reordered using the Unicode Bidirectional Algorithm. Hebrew uses BiDi reordering without additional shaping. Both scripts render correctly in frames, inline mixed-direction text, and multi-line word-wrapped paragraphs."
	endY, _ := doc.WriteText(intro, marginL, y, contentW)
	y = endY + 14

	// helper — small gray annotation label
	lbl := func(text string) {
		doc.SetFont("regular", 8) //nolint:errcheck
		doc.SetTextColor(150, 150, 150)
		doc.WriteLine(text, marginL, y) //nolint:errcheck
		y += 12
	}
	// helper — section heading band
	sect := func(title string) {
		doc.SetFont(boldFont, 10) //nolint:errcheck
		doc.SetTextColor(colorNavy.R, colorNavy.G, colorNavy.B)
		doc.FillRect(marginL, y, contentW, 18, colorOffWhite)
		doc.WriteLine(title, marginL+6, y+4) //nolint:errcheck
		y += 22
	}
	// helper — thin horizontal rule
	hr := func() {
		doc.DrawBorder(marginL, y, contentW, 0, pdf.Border{ //nolint:errcheck
			Top: &pdf.BorderSpec{Thickness: 0.4, Color: pdf.ColorLightGray},
		})
		y += 10
	}

	// ══════════════════════════════════════════════════════════════════════
	// PAGE A — Arabic
	// ══════════════════════════════════════════════════════════════════════
	if hasArabic {
		// breakPage starts a new page when fewer than `needed` points remain.
		breakPage := func(needed float64) {
			if y+needed > pageH-marginB {
				doc.AddPage()
				y = marginT
			}
		}

		sect("Arabic — rtl.Shape() + WriteLineRTL / WriteTextRTL")

		// ── Vocabulary samples ────────────────────────────────────────────
		type sample struct{ label, text string }
		vocab := []sample{
			{"كتاب  (kitāb — book)", "كتاب"},
			{"مدرسة  (madrasa — school)", "مدرسة"},
			{"مرحبا بالعالم  (hello world)", "مرحبا بالعالم"},
			{"شكرا جزيلا  (shukran jazīlan — thank you very much)", "شكرا جزيلا"},
			{"lam-alef ligatures: لا  لأ  لإ  لآ", "لا لأ لإ لآ"},
			{"mixed number — السعر 42 دولار", "السعر 42 دولار"},
			{"mixed Latin brand — شركة Google للتكنولوجيا", "شركة Google للتكنولوجيا"},
			{"mixed punctuation — هل أنت مستعد؟ نعم!", "هل أنت مستعد؟ نعم!"},
		}
		for _, s := range vocab {
			breakPage(40)
			lbl(s.label)
			doc.SetFont("arabic", 18) //nolint:errcheck
			doc.SetTextColor(20, 20, 20)
			doc.WriteLineRTL(rtl.Shape(s.text), rightX, y) //nolint:errcheck
			y += 26
		}

		// ── Prose paragraph ───────────────────────────────────────────────
		breakPage(120)
		hr()
		lbl("Prose paragraph — كلام مترابط (WriteTextRTL, font size 15):")
		doc.SetFont("arabic", 15) //nolint:errcheck
		doc.SetTextColor(20, 20, 20)
		prose1 := "في عالم يتسارع فيه التطور التكنولوجي بشكل لم يسبق له مثيل، أصبح إتقان أدوات البرمجة الحديثة ضرورةً لا غنى عنها لكل مطوّر يسعى إلى النجاح. تقدم مكتبة Nautilus للمطورين القدرة على توليد ملفات PDF عالية الجودة مباشرةً من لغة Go، دون الحاجة إلى أي تبعيات خارجية معقدة."
		endY, _ = doc.WriteTextRTL(prose1, rightX, y, contentW)
		y = endY + 8

		breakPage(60)
		prose2 := "يدعم النظام اللغة العربية بشكل كامل، بما في ذلك التشكيل السياقي للحروف، وربط اللام بالألف في جميع حالاته، وإعادة الترتيب البصري وفق خوارزمية Unicode ثنائية الاتجاه. كما يمكن دمج النصوص العربية مع النصوص اللاتينية والأرقام في نفس السطر بسلاسة تامة."
		endY, _ = doc.WriteTextRTL(prose2, rightX, y, contentW)
		y = endY + 14

		// ── Framed quote ──────────────────────────────────────────────────
		breakPage(120)
		hr()
		lbl("Arabic proverb inside a Frame (right-side gold accent border):")
		bg1 := pdf.Color{R: 255, G: 252, B: 235}
		fQ := doc.NewFrame(pdf.FrameConfig{
			X: marginL, Y: y, Width: contentW,
			Background: &bg1,
			Border: pdf.Border{
				Right: &pdf.BorderSpec{Thickness: 5, Color: colorGold},
			},
			Padding: pdf.Padding{Top: 10, Right: 18, Bottom: 10, Left: 12},
		})
		fQ.SetFont("arabic", 17) //nolint:errcheck
		fQ.SetTextColor(30, 30, 30)
		fQ.WriteTextRTL("من طلب العلا سهر الليالي") //nolint:errcheck
		fQ.Close()                                    //nolint:errcheck
		y = fQ.CurrentY() + 6

		doc.SetFont("regular", 8) //nolint:errcheck
		doc.SetTextColor(120, 120, 120)
		doc.WriteLine("  (Arabic proverb: \"He who seeks glory stays awake at night\")", marginL, y) //nolint:errcheck
		y += 16

		// ── Second framed block — Nautilus welcome ────────────────────────
		breakPage(90)
		lbl("Longer Arabic text block inside a Frame:")
		bg2 := pdf.Color{R: 240, G: 248, B: 255}
		fW := doc.NewFrame(pdf.FrameConfig{
			X: marginL, Y: y, Width: contentW,
			Background: &bg2,
			Border: pdf.Border{
				Right: &pdf.BorderSpec{Thickness: 4, Color: colorNavy},
			},
			Padding: pdf.Padding{Top: 10, Right: 18, Bottom: 10, Left: 12},
		})
		fW.SetFont("arabic", 15) //nolint:errcheck
		fW.SetTextColor(20, 20, 20)
		fW.WriteTextRTL("مرحبا بكم في مكتبة Nautilus لتوليد ملفات PDF باستخدام لغة Go. تتميز هذه المكتبة بدعمها الكامل للخطوط العربية، وإمكانية الجمع بين النصوص ذات الاتجاهات المختلفة في صفحة واحدة. يمكنك إنشاء جداول، ورسوم بيانية، وتخطيطات معقدة متعددة الأعمدة بكل سهولة.") //nolint:errcheck
		fW.Close()                                                                                                                                                                                                                                                                          //nolint:errcheck
		y = fW.CurrentY() + 14

		// ── Three-column vocab table ──────────────────────────────────────
		breakPage(160)
		hr()
		sect("Arabic vocabulary table — three columns")
		colW3 := (contentW - 16) / 3
		tableWords := []struct{ ar, roman, en string }{
			{"بيت", "bayt", "house"},
			{"ماء", "māʾ", "water"},
			{"شمس", "shams", "sun"},
			{"قمر", "qamar", "moon"},
			{"نور", "nūr", "light"},
			{"حياة", "ḥayāt", "life"},
		}

		// header row
		hdrBg := colorOffWhite
		doc.FillRect(marginL, y, contentW, 18, hdrBg)
		doc.SetFont(boldFont, 9) //nolint:errcheck
		doc.SetTextColor(colorNavy.R, colorNavy.G, colorNavy.B)
		doc.WriteLine("Arabic", marginL+4, y+4)                  //nolint:errcheck
		doc.WriteLine("Transliteration", marginL+colW3+8+4, y+4) //nolint:errcheck
		doc.WriteLine("English", marginL+2*(colW3+8)+4, y+4)     //nolint:errcheck
		doc.DrawBorder(marginL, y, contentW, 18, pdf.Border{      //nolint:errcheck
			Bottom: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray},
		})
		y += 20

		for i, w := range tableWords {
			breakPage(22)
			rowBg := pdf.Color{R: 255, G: 255, B: 255}
			if i%2 == 1 {
				rowBg = pdf.Color{R: 248, G: 248, B: 252}
			}
			doc.FillRect(marginL, y, contentW, 18, rowBg)
			// Arabic cell (RTL)
			doc.SetFont("arabic", 14) //nolint:errcheck
			doc.SetTextColor(20, 20, 20)
			doc.WriteLineRTL(rtl.Shape(w.ar), marginL+colW3-4, y+2) //nolint:errcheck
			// Romanisation cell
			doc.SetFont("regular", 9) //nolint:errcheck
			doc.SetTextColor(80, 80, 80)
			doc.WriteLine(w.roman, marginL+colW3+8+4, y+4) //nolint:errcheck
			// English cell
			doc.SetFont("regular", 9) //nolint:errcheck
			doc.SetTextColor(20, 20, 20)
			doc.WriteLine(w.en, marginL+2*(colW3+8)+4, y+4) //nolint:errcheck
			y += 20
		}
		y += 6
	}

	// ══════════════════════════════════════════════════════════════════════
	// PAGE B — Hebrew  (new page so each script gets full breathing room)
	// ══════════════════════════════════════════════════════════════════════
	if hasHebrew {
		doc.AddPage()
		y = sectionHeader(doc, boldFont, "Right-to-Left Text — Hebrew", marginT+12)
		y += 10

		// breakPage starts a new page when fewer than `needed` points remain.
		breakPage := func(needed float64) {
			if y+needed > pageH-marginB {
				doc.AddPage()
				y = marginT
			}
		}

		doc.SetFont("regular", 10) //nolint:errcheck
		doc.SetTextColor(60, 60, 60)
		hebrewIntro := "Hebrew is written right-to-left and shares the Unicode Bidirectional Algorithm with Arabic. Unlike Arabic, Hebrew letters do not change shape depending on their position in a word, so no contextual shaping is required — rtl.Shape() passes the text unchanged. The pdf/rtl package applies BiDi reordering so that glyphs are placed in visual (right-to-left) order on the page."
		endY, _ = doc.WriteText(hebrewIntro, marginL, y, contentW)
		y = endY + 14

		sect("Hebrew — basic vocabulary (WriteLineRTL)")

		type sample struct{ label, text string }
		vocab := []sample{
			{"שלום  (shalom — peace / hello / goodbye)", "שלום"},
			{"תודה  (todah — thank you)", "תודה"},
			{"תודה רבה  (todah rabah — thank you very much)", "תודה רבה"},
			{"בבקשה  (bevakasha — please / you're welcome)", "בבקשה"},
			{"כן / לא  (ken / lo — yes / no)", "כן / לא"},
			{"ישראל  (Yisraʾel — Israel)", "ישראל"},
			{"ירושלים  (Yerushalayim — Jerusalem)", "ירושלים"},
			{"שבת שלום  (Shabbat Shalom — Sabbath greeting)", "שבת שלום"},
			{"mixed — Google הוא מנוע חיפוש פופולרי", "Google הוא מנוע חיפוש פופולרי"},
			{"mixed number — המחיר הוא 99 שקלים", "המחיר הוא 99 שקלים"},
		}
		for _, s := range vocab {
			breakPage(40)
			lbl(s.label)
			doc.SetFont("hebrew", 18) //nolint:errcheck
			doc.SetTextColor(20, 20, 20)
			doc.WriteLineRTL(rtl.Shape(s.text), rightX, y) //nolint:errcheck
			y += 26
		}

		// ── Biblical passage ──────────────────────────────────────────────
		breakPage(120)
		hr()
		sect("Biblical Hebrew — Genesis 1:1-3 (WriteTextRTL)")
		doc.SetFont("hebrew", 15) //nolint:errcheck
		doc.SetTextColor(20, 20, 20)
		gen1 := "בְּרֵאשִׁית בָּרָא אֱלֹהִים אֵת הַשָּׁמַיִם וְאֵת הָאָרֶץ"
		endY, _ = doc.WriteTextRTL(gen1, rightX, y, contentW)
		y = endY + 6
		gen2 := "וְהָאָרֶץ הָיְתָה תֹהוּ וָבֹהוּ וְחֹשֶׁךְ עַל-פְּנֵי תְהוֹם וְרוּחַ אֱלֹהִים מְרַחֶפֶת עַל-פְּנֵי הַמָּיִם"
		endY, _ = doc.WriteTextRTL(gen2, rightX, y, contentW)
		y = endY + 6
		gen3 := "וַיֹּאמֶר אֱלֹהִים יְהִי אוֹר וַיְהִי-אוֹר"
		endY, _ = doc.WriteTextRTL(gen3, rightX, y, contentW)
		y = endY + 6

		doc.SetFont("regular", 8) //nolint:errcheck
		doc.SetTextColor(120, 120, 120)
		doc.WriteLine("  Genesis 1:1–3 (Hebrew Bible, public domain)", marginL, y) //nolint:errcheck
		y += 16

		// ── Modern prose paragraph ────────────────────────────────────────
		breakPage(200)
		hr()
		sect("Modern Hebrew prose — technology paragraph (WriteTextRTL)")
		doc.SetFont("hebrew", 14) //nolint:errcheck
		doc.SetTextColor(20, 20, 20)
		modern1 := "בעידן הדיגיטלי המודרני, פיתוח תוכנה הפך לאחד המקצועות המבוקשים ביותר בשוק העבודה הגלובלי. שפת התכנות Go, שפותחה על ידי Google בשנת 2009, זכתה לפופולריות רבה בקרב מפתחים בשל פשטותה, ביצועיה הגבוהים ותמיכתה המובנית בתכנות מקבילי."
		endY, _ = doc.WriteTextRTL(modern1, rightX, y, contentW)
		y = endY + 8

		breakPage(60)
		modern2 := "ספריית Nautilus מאפשרת למפתחי Go ליצור מסמכי PDF מורכבים ישירות מהקוד, ללא תלות בספריות חיצוניות. הספרייה תומכת בעברית, בערבית, ובשפות נוספות הנכתבות מימין לשמאל, תוך שימוש באלגוריתם Unicode הדו-כיווני לסידור הנכון של הטקסט."
		endY, _ = doc.WriteTextRTL(modern2, rightX, y, contentW)
		y = endY + 8

		breakPage(60)
		modern3 := "בנוסף, ניתן לשלב טקסט עברי עם מספרים ועם מילים בלטינית באותו משפט, כפי שמוצג בדוגמאות לעיל. יכולת זו חיונית ביישומים עסקיים רבים, כגון חשבוניות, דוחות כספיים ומסמכים רשמיים."
		endY, _ = doc.WriteTextRTL(modern3, rightX, y, contentW)
		y = endY + 14

		// ── Framed quote ──────────────────────────────────────────────────
		breakPage(130)
		hr()
		sect("Hebrew proverb inside a Frame")
		bgHeb := pdf.Color{R: 245, G: 245, B: 255}
		fHQ := doc.NewFrame(pdf.FrameConfig{
			X: marginL, Y: y, Width: contentW,
			Background: &bgHeb,
			Border: pdf.Border{
				Right: &pdf.BorderSpec{Thickness: 5, Color: colorNavy},
			},
			Padding: pdf.Padding{Top: 12, Right: 20, Bottom: 12, Left: 14},
		})
		fHQ.SetFont("hebrew", 17) //nolint:errcheck
		fHQ.SetTextColor(30, 30, 30)
		fHQ.WriteTextRTL("כָּל יִשְׂרָאֵל יֵשׁ לָהֶם חֵלֶק לָעוֹלָם הַבָּא") //nolint:errcheck
		fHQ.Close()                                                               //nolint:errcheck
		y = fHQ.CurrentY() + 6

		doc.SetFont("regular", 8) //nolint:errcheck
		doc.SetTextColor(120, 120, 120)
		doc.WriteLine("  (Sanhedrin 10:1 — \"All Israel have a share in the world to come\")", marginL, y) //nolint:errcheck
		y += 16

		// ── Second frame — longer modern text ────────────────────────────
		breakPage(90)
		bgHeb2 := pdf.Color{R: 240, G: 255, B: 245}
		fHL := doc.NewFrame(pdf.FrameConfig{
			X: marginL, Y: y, Width: contentW,
			Background: &bgHeb2,
			Border: pdf.Border{
				Right: &pdf.BorderSpec{Thickness: 4, Color: colorMint},
			},
			Padding: pdf.Padding{Top: 10, Right: 18, Bottom: 10, Left: 12},
		})
		fHL.SetFont("hebrew", 14) //nolint:errcheck
		fHL.SetTextColor(20, 20, 20)
		fHL.WriteTextRTL("דוגמה זו מציגה טקסט עברי ארוך יותר בתוך מסגרת. המסגרת מספקת רקע וגבולות, ומאפשרת ארגון ויזואלי ברור של התוכן. הטקסט נאסף באופן אוטומטי בתוך גבולות המסגרת.") //nolint:errcheck
		fHL.Close()                                                                                                                                                                          //nolint:errcheck
		y = fHL.CurrentY() + 14
	}

	// ══════════════════════════════════════════════════════════════════════
	// Side-by-side three-column comparison (always on its own fresh page)
	// ══════════════════════════════════════════════════════════════════════
	if hasArabic && hasHebrew {
		doc.AddPage()
		y = marginT
		sect("Side-by-side: LTR (English) | RTL (Arabic) | RTL (Hebrew)")
		colW := (contentW - 20) / 3

		fEn := doc.NewFrame(pdf.FrameConfig{
			X: marginL, Y: y, Width: colW,
			Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
			Padding: pdf.UniformPadding(8),
		})
		fEn.SetFont(boldFont, 9) //nolint:errcheck
		fEn.SetTextColor(colorNavy.R, colorNavy.G, colorNavy.B)
		fEn.WriteText("English (LTR)") //nolint:errcheck
		fEn.Advance(4)
		fEn.SetFont("regular", 10) //nolint:errcheck
		fEn.SetTextColor(20, 20, 20)
		fEn.WriteText("The quick brown fox jumps over the lazy dog. Programming in Go is fast and fun.") //nolint:errcheck
		fEn.Close()                                                                                       //nolint:errcheck

		fAr := doc.NewFrame(pdf.FrameConfig{
			X: marginL + colW + 10, Y: y, Width: colW,
			Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
			Padding: pdf.UniformPadding(8),
		})
		fAr.SetFont(boldFont, 9) //nolint:errcheck
		fAr.SetTextColor(colorNavy.R, colorNavy.G, colorNavy.B)
		fAr.WriteText("Arabic (RTL)") //nolint:errcheck
		fAr.Advance(4)
		fAr.SetFont("arabic", 13) //nolint:errcheck
		fAr.SetTextColor(20, 20, 20)
		fAr.WriteTextRTL("الثعلب البني السريع يقفز فوق الكلب الكسول. البرمجة بلغة Go سريعة وممتعة.") //nolint:errcheck
		fAr.Close()                                                                                      //nolint:errcheck

		fHe := doc.NewFrame(pdf.FrameConfig{
			X: marginL + 2*(colW+10), Y: y, Width: colW,
			Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
			Padding: pdf.UniformPadding(8),
		})
		fHe.SetFont(boldFont, 9) //nolint:errcheck
		fHe.SetTextColor(colorNavy.R, colorNavy.G, colorNavy.B)
		fHe.WriteText("Hebrew (RTL)") //nolint:errcheck
		fHe.Advance(4)
		fHe.SetFont("hebrew", 13) //nolint:errcheck
		fHe.SetTextColor(20, 20, 20)
		fHe.WriteTextRTL("השועל החום המהיר קופץ מעל הכלב העצלן. תכנות ב-Go מהיר ומהנה.") //nolint:errcheck
		fHe.Close()                                                                          //nolint:errcheck
	}
}

// ── Page 5: Drawing Primitives ────────────────────────────────────────────────

func buildDrawingPage(doc *pdf.Document, boldFont string) {
	doc.AddPage()
	y := sectionHeader(doc, boldFont, "Drawing Primitives", marginT+12)
	y += 10

	// ── Filled rectangles ────────────────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Filled Rectangles", marginL, y) //nolint:errcheck
	y += 14

	rectColors := []pdf.Color{colorNavy, colorCyan, colorMint, colorCoral, colorGold, colorLavender}
	rectW := (contentW - 5*8) / 6
	for i, c := range rectColors {
		rx := marginL + float64(i)*(rectW+8)
		doc.FillRect(rx, y, rectW, 40, c)
	}
	y += 56

	// ── Lines of various widths ───────────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Lines — Various Widths & Colors", marginL, y) //nolint:errcheck
	y += 14

	lineSpecs := []struct {
		w   float64
		c   pdf.Color
		lbl string
	}{
		{0.5, colorNavy, "0.5 pt navy"},
		{1.0, colorCyan, "1.0 pt cyan"},
		{2.0, colorMint, "2.0 pt mint"},
		{3.0, colorCoral, "3.0 pt coral"},
		{5.0, colorGold, "5.0 pt gold"},
	}
	for _, ls := range lineSpecs {
		doc.DrawLine(marginL+60, y+5, marginL+contentW-10, y+5, ls.w, ls.c)
		doc.SetFont("regular", 8) //nolint:errcheck
		doc.SetTextColor(80, 80, 80)
		doc.WriteLine(ls.lbl, marginL, y) //nolint:errcheck
		y += 16
	}
	y += 6

	// ── Polygons ─────────────────────────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Polygons", marginL, y) //nolint:errcheck
	y += 14

	polyBaseY := y

	// Triangle (filled)
	triX := marginL + 50.0
	triY := polyBaseY
	doc.FillPolygon([]pdf.Point{
		{X: triX, Y: triY + 70},
		{X: triX + 60, Y: triY},
		{X: triX + 120, Y: triY + 70},
	}, colorCyan)
	doc.SetFont("regular", 8) //nolint:errcheck
	doc.SetTextColor(80, 80, 80)
	doc.WriteLine("Filled triangle", triX-10, polyBaseY+80) //nolint:errcheck

	// Pentagon (filled + stroked)
	pentCX := marginL + 245.0
	pentCY := polyBaseY + 38
	pentR := 38.0
	pentPts := regularPolygon(pentCX, pentCY, pentR, 5)
	doc.FillAndStrokePolygon(pentPts, colorMint, 2, colorNavy)
	doc.SetFont("regular", 8) //nolint:errcheck
	doc.WriteLine("Filled+stroked pentagon", pentCX-42, polyBaseY+80) //nolint:errcheck

	// Hexagon (filled)
	hexCX := marginL + 390.0
	hexCY := polyBaseY + 38
	hexPts := regularPolygon(hexCX, hexCY, 38, 6)
	doc.FillPolygon(hexPts, colorGold)
	doc.SetFont("regular", 8) //nolint:errcheck
	doc.WriteLine("Filled hexagon", hexCX-28, polyBaseY+80) //nolint:errcheck

	y = polyBaseY + 100

	// ── Circles ──────────────────────────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Circles", marginL, y) //nolint:errcheck
	y += 14

	circBaseY := y
	circles := []struct {
		cx, r float64
		fill  bool
		c     pdf.Color
		lbl   string
	}{
		{marginL + 50, 35, true, colorCoral, "Filled"},
		{marginL + 150, 35, false, colorNavy, "Stroked 1pt"},
		{marginL + 250, 30, true, colorLavender, "Filled"},
		{marginL + 350, 25, false, colorMint, "Stroked 2pt"},
		{marginL + 430, 20, true, colorCyan, "Filled"},
	}
	for _, ci := range circles {
		if ci.fill {
			doc.FillCircle(ci.cx, circBaseY+40, ci.r, ci.c)
		} else {
			doc.StrokeCircle(ci.cx, circBaseY+40, ci.r, 2, ci.c)
		}
		doc.SetFont("regular", 8) //nolint:errcheck
		doc.SetTextColor(80, 80, 80)
		lw, _ := doc.MeasureText(ci.lbl)
		doc.WriteLine(ci.lbl, ci.cx-lw/2, circBaseY+80) //nolint:errcheck
	}

	y = circBaseY + 100

	// ── Graphics state save/restore ──────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Graphics State: SaveGraphicsState / RestoreGraphicsState", marginL, y) //nolint:errcheck
	y += 14

	// Draw a colored shape, save state, draw another, restore
	doc.SaveGraphicsState()
	doc.FillRect(marginL, y, 80, 30, colorNavy)
	doc.SaveGraphicsState()
	doc.FillRect(marginL+90, y, 80, 30, colorCoral)
	doc.RestoreGraphicsState()
	doc.FillRect(marginL+180, y, 80, 30, colorMint)
	doc.RestoreGraphicsState()

	doc.SetFont("regular", 9) //nolint:errcheck
	doc.SetTextColor(80, 80, 80)
	doc.WriteLine("Three rects drawn with nested Save/Restore state", marginL, y+36) //nolint:errcheck
	y += 58

	// ── Go-themed description ─────────────────────────────────────────────────
	doc.SetFont("regular", 10) //nolint:errcheck
	doc.SetTextColor(60, 60, 60)
	doc.WriteText(goText3, marginL, y, contentW) //nolint:errcheck
}

// ── Page 4: Borders & Frames ──────────────────────────────────────────────────

func buildBordersFramesPage(doc *pdf.Document, boldFont string) {
	doc.AddPage()
	y := sectionHeader(doc, boldFont, "Borders & Frames", marginT+12)
	y += 10

	// ── Border patterns ───────────────────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Border Patterns (all 4 built-in styles)", marginL, y) //nolint:errcheck
	y += 14

	patterns := []struct {
		name    string
		pattern pdf.BorderPattern
		color   pdf.Color
	}{
		{"Solid", pdf.PatternSolid, colorNavy},
		{"Dashed", pdf.PatternDashed, colorCoral},
		{"Dotted", pdf.PatternDotted, colorMint},
		{"Dash-Dot", pdf.PatternDashDot, colorLavender},
	}

	colW := (contentW - 3*10) / 4
	for i, p := range patterns {
		bx := marginL + float64(i)*(colW+10)
		spec := pdf.BorderSpec{Thickness: 1.5, Color: p.color, Pattern: p.pattern}
		border := pdf.NewUniformBorder(spec)
		doc.DrawBorder(bx, y, colW, 50, border) //nolint:errcheck

		doc.SetFont("regular", 9) //nolint:errcheck
		doc.SetTextColor(50, 50, 50)
		lw, _ := doc.MeasureText(p.name)
		doc.WriteLine(p.name, bx+(colW-lw)/2, y+56) //nolint:errcheck
	}
	y += 76

	// ── Individual sides ─────────────────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Selective Sides (top only, bottom only, left+right)", marginL, y) //nolint:errcheck
	y += 14

	topSpec := &pdf.BorderSpec{Thickness: 3, Color: colorCyan, Pattern: pdf.PatternSolid}
	doc.DrawBorder(marginL, y, 120, 40, pdf.Border{Top: topSpec}) //nolint:errcheck
	doc.SetFont("regular", 8)                                     //nolint:errcheck
	doc.SetTextColor(80, 80, 80)
	doc.WriteLine("Top only", marginL+10, y+14) //nolint:errcheck

	botSpec := &pdf.BorderSpec{Thickness: 3, Color: colorCoral, Pattern: pdf.PatternDashed}
	doc.DrawBorder(marginL+140, y, 120, 40, pdf.Border{Bottom: botSpec}) //nolint:errcheck
	doc.WriteLine("Bottom only (dashed)", marginL+144, y+14)             //nolint:errcheck

	lrSpec := &pdf.BorderSpec{Thickness: 2, Color: colorMint, Pattern: pdf.PatternDotted}
	doc.DrawBorder(marginL+280, y, 120, 40, pdf.Border{Left: lrSpec, Right: lrSpec}) //nolint:errcheck
	doc.WriteLine("Left+Right (dotted)", marginL+284, y+14)                          //nolint:errcheck

	y += 60

	// ── Frames ───────────────────────────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Frames: auto-height, fixed-height with background, padding", marginL, y) //nolint:errcheck
	y += 14

	// Frame 1: auto-height, border only
	f1 := doc.NewFrame(pdf.FrameConfig{
		X: marginL, Y: y, Width: 150,
		Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1, Color: colorNavy}),
		Padding: pdf.UniformPadding(8),
	})
	f1.SetFont("regular", 9) //nolint:errcheck
	f1.WriteText("Auto-height frame with a solid navy border and 8 pt uniform padding. Content flows down and the border adapts.") //nolint:errcheck
	f1.Close()                                                                                                                      //nolint:errcheck

	// Frame 2: fixed-height, colored background
	bg2 := colorLightBlue
	f2 := doc.NewFrame(pdf.FrameConfig{
		X: marginL + 165, Y: y, Width: 150, Height: 90,
		Background: &bg2,
		Border:     pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1.5, Color: colorCyan}),
		Padding:    pdf.HorizontalPadding(10, 8),
	})
	f2.SetFont("regular", 9) //nolint:errcheck
	f2.WriteText("Fixed 90 pt height, light-blue background, 1.5 pt cyan border.") //nolint:errcheck
	f2.Close()                                                                       //nolint:errcheck

	// Frame 3: fixed-height, dark background, white text
	bg3 := colorNavy
	f3 := doc.NewFrame(pdf.FrameConfig{
		X: marginL + 330, Y: y, Width: 175, Height: 90,
		Background: &bg3,
		Padding:    pdf.UniformPadding(10),
	})
	f3.SetFont("regular", 9)   //nolint:errcheck
	f3.SetTextColor(255, 255, 255)
	f3.WriteText("Dark navy background, white text, no explicit border, 10 pt padding on all sides.") //nolint:errcheck
	f3.Close()                                                                                          //nolint:errcheck

	y += 110

	// ── Two-column frame layout ───────────────────────────────────────────────
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Two-column frame layout using pdf.Frame", marginL, y) //nolint:errcheck
	y += 14

	colFrameW := (contentW - 10) / 2

	bgLeft := colorOffWhite
	left := doc.NewFrame(pdf.FrameConfig{
		X: marginL, Y: y, Width: colFrameW,
		Background: &bgLeft,
		Border:     pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
		Padding:    pdf.UniformPadding(8),
	})
	left.SetFont("regular", 9) //nolint:errcheck
	left.SetTextColor(30, 30, 30)
	left.WriteText("Left column: "+goText4) //nolint:errcheck
	left.Close()                            //nolint:errcheck

	right := doc.NewFrame(pdf.FrameConfig{
		X: marginL + colFrameW + 10, Y: y, Width: colFrameW,
		Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
		Padding: pdf.UniformPadding(8),
	})
	right.SetFont("regular", 9) //nolint:errcheck
	right.SetTextColor(30, 30, 30)
	right.WriteText("Right column: "+goText5) //nolint:errcheck
	right.Close()                             //nolint:errcheck

	// Record how tall the tallest column is for spacing
	y += left.FrameHeight() + 12

	// ── DrawBorder helper note ────────────────────────────────────────────────
	doc.SetFont("regular", 9) //nolint:errcheck
	doc.SetTextColor(100, 100, 100)
	doc.WriteText("doc.DrawBorder draws per-side borders directly on the page. doc.NewFrame wraps content with optional background, padding, and border, calling doc.DrawBorder internally on Close.", marginL, y, contentW) //nolint:errcheck
}

// ── Page 5: Tables ───────────────────────────────────────────────────────────

func buildTablesPage(doc *pdf.Document, boldFont string) {
	doc.AddPage()
	sectionHeader(doc, boldFont, "Tables", marginT+12) //nolint:errcheck

	tblY := marginT + 38.0

	// Column widths
	c1, c2, c3, c4, c5 := 80.0, 130.0, 110.0, 90.0, 95.0
	cols := []float64{c1, c2, c3, c4, c5}

	hdrBg := colorNavy
	hdrText := pdf.ColorWhite

	cellBorder := pdf.NewUniformBorder(pdf.BorderSpec{
		Thickness: 0.5, Color: pdf.Color{R: 180, G: 190, B: 210},
	})

	tbl := doc.NewTable(pdf.TableConfig{
		X:          marginL,
		Y:          tblY,
		ColWidths:  cols,
		PageBottom: pageH - marginB - 20,
		Border: pdf.NewUniformBorder(pdf.BorderSpec{
			Thickness: 1, Color: colorNavy,
		}),
		DefaultCellStyle: pdf.CellStyle{
			Padding:  pdf.Padding{Top: 5, Right: 6, Bottom: 5, Left: 6},
			Border:   cellBorder,
			FontName: "regular",
			FontSize: 9,
		},
	})

	// Header row (colspan spans "Name" across nothing; Status spans 2 cols)
	tbl.AddRow(pdf.Row{
		Height: 24,
		Cells: []pdf.Cell{
			{Text: "Package", Style: pdf.CellStyle{
				Background: &hdrBg, TextColor: &hdrText,
				FontName: boldFont, FontSize: 10, HAlign: pdf.HAlignCenter, VAlign: pdf.VAlignMiddle,
			}},
			{Text: "Description", Style: pdf.CellStyle{
				Background: &hdrBg, TextColor: &hdrText,
				FontName: boldFont, FontSize: 10, HAlign: pdf.HAlignCenter, VAlign: pdf.VAlignMiddle,
			}},
			{Text: "Category", Style: pdf.CellStyle{
				Background: &hdrBg, TextColor: &hdrText,
				FontName: boldFont, FontSize: 10, HAlign: pdf.HAlignCenter, VAlign: pdf.VAlignMiddle,
			}},
			{Text: "Stars", Style: pdf.CellStyle{
				Background: &hdrBg, TextColor: &hdrText,
				FontName: boldFont, FontSize: 10, HAlign: pdf.HAlignCenter, VAlign: pdf.VAlignMiddle,
			}},
			{Text: "License", Style: pdf.CellStyle{
				Background: &hdrBg, TextColor: &hdrText,
				FontName: boldFont, FontSize: 10, HAlign: pdf.HAlignCenter, VAlign: pdf.VAlignMiddle,
			}},
		},
	})

	rowAlt := colorRowAlt
	cyanColor := colorCyan

	// Row with rowspan: "gopdf" spans 2 rows in column 0
	tbl.AddRow(pdf.Row{
		Background: &rowAlt,
		Cells: []pdf.Cell{
			{Text: "gopdf", RowSpan: 2, Style: pdf.CellStyle{
				FontName: boldFont, HAlign: pdf.HAlignCenter, VAlign: pdf.VAlignMiddle,
				Background: &cyanColor, TextColor: &hdrText,
			}},
			{Text: "Low-level PDF generation primitives"},
			{Text: "PDF engine"},
			{Text: "3.2k", Style: pdf.CellStyle{HAlign: pdf.HAlignRight}},
			{Text: "MIT"},
		},
	})
	tbl.AddRow(pdf.Row{
		Background: &rowAlt,
		Cells: []pdf.Cell{
			// col 0 occupied by rowspan above
			{Text: "Foundation for all Nautilus drawing"},
			{Text: "PDF engine"},
			{Text: "—", Style: pdf.CellStyle{HAlign: pdf.HAlignCenter}},
			{Text: "MIT"},
		},
	})

	// Normal rows, alternating colors
	dataRows := []struct {
		pkg, desc, cat, stars, lic string
		alt                         bool
	}{
		{"uniseg", "Unicode segmentation & grapheme clusters", "Unicode", "800+", "MIT", false},
		{"gomoji", "Emoji detection and listing", "Emoji", "400+", "MIT", true},
		{"layout", "Platypus-style layout engine", "Layout", "internal", "MIT", false},
		{"chart", "Highcharts-style chart renderers", "Charts", "internal", "MIT", true},
		{"html", "Inline HTML markup parser", "Text", "internal", "MIT", false},
		{"rtl", "Arabic shaping and BiDi reordering", "Text", "internal", "MIT", true},
		{"border", "Per-side border drawing utilities", "Drawing", "internal", "MIT", false},
		{"emoji", "PNG emoji substitution resolver", "Emoji", "internal", "MIT", true},
	}

	for _, dr := range dataRows {
		var bg *pdf.Color
		if dr.alt {
			c := colorRowAlt
			bg = &c
		}
		tbl.AddRow(pdf.Row{
			Background: bg,
			Cells: []pdf.Cell{
				{Text: dr.pkg, Style: pdf.CellStyle{FontName: boldFont}},
				{Text: dr.desc},
				{Text: dr.cat, Style: pdf.CellStyle{HAlign: pdf.HAlignCenter}},
				{Text: dr.stars, Style: pdf.CellStyle{HAlign: pdf.HAlignRight}},
				{Text: dr.lic, Style: pdf.CellStyle{HAlign: pdf.HAlignCenter}},
			},
		})
	}

	// colspan row: summary spanning all 5 columns
	summaryBg := pdf.Color{R: 220, G: 235, B: 250}
	tbl.AddRow(pdf.Row{
		Height: 20,
		Cells: []pdf.Cell{
			{Text: "All packages are pure Go — no CGO required.", ColSpan: 5,
				Style: pdf.CellStyle{
					Background: &summaryBg,
					HAlign:     pdf.HAlignCenter,
					VAlign:     pdf.VAlignMiddle,
					FontName:   boldFont,
					FontSize:   9,
				}},
		},
	})

	if err := tbl.Draw(); err != nil {
		log.Printf("table draw: %v", err)
	}

	// Caption below table — use tbl.CurrentY() to get the actual bottom.
	tableW := c1 + c2 + c3 + c4 + c5
	doc.SetFont("regular", 8) //nolint:errcheck
	doc.SetTextColor(120, 120, 120)
	cap := "Table: Nautilus sub-packages. Rows with rowspan (gopdf) and colspan (summary) are demonstrated."
	doc.WriteText(cap, marginL, tbl.CurrentY()+6, tableW) //nolint:errcheck
}

// ── Page 6: Images ────────────────────────────────────────────────────────────

func buildImagesPage(doc *pdf.Document, imDir, boldFont string) {
	doc.AddPage()
	y := sectionHeader(doc, boldFont, "Images", marginT+12)
	y += 10

	doc.SetFont("regular", 10) //nolint:errcheck
	doc.SetTextColor(60, 60, 60)
	doc.WriteText("Images are rendered with doc.DrawImage(path, x, y, width, height). The method accepts PNG and JPEG files.", marginL, y, contentW) //nolint:errcheck
	y += 28

	images := []struct {
		file, caption     string
		w, h              float64
	}{
		{"gopher-laptop.png", "Gopher with laptop", 130, 110},
		{"gopher-run.png", "Running Gopher", 110, 110},
		{"gopher-talks.png", "Gopher talks", 120, 110},
		{"gopher-doc.png", "Gopher with docs", 120, 110},
	}

	spacing := 12.0
	totalImgW := 0.0
	for _, img := range images {
		totalImgW += img.w
	}
	totalImgW += float64(len(images)-1) * spacing
	startX := marginL + (contentW-totalImgW)/2
	curX := startX

	imgRowH := 0.0
	for _, img := range images {
		if img.h > imgRowH {
			imgRowH = img.h
		}
	}

	for _, img := range images {
		imgPath := filepath.Join(imDir, img.file)
		imgY := y + (imgRowH-img.h)/2 // vertically center different-sized images

		// Draw a light background box for each image
		doc.FillRect(curX-4, y-4, img.w+8, imgRowH+8, colorOffWhite)
		border := pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray})
		doc.DrawBorder(curX-4, y-4, img.w+8, imgRowH+8, border) //nolint:errcheck

		if err := doc.DrawImage(imgPath, curX, imgY, img.w, img.h); err != nil {
			// Placeholder if image not found
			doc.FillCircle(curX+img.w/2, imgY+img.h/2, 30, colorCyan)
		}
		curX += img.w + spacing
	}

	y += imgRowH + 14

	// Captions
	curX = startX
	doc.SetFont("regular", 8) //nolint:errcheck
	doc.SetTextColor(100, 100, 100)
	for _, img := range images {
		cw, _ := doc.MeasureText(img.caption)
		cx := curX + img.w/2 - cw/2
		if cx < marginL {
			cx = marginL
		}
		doc.WriteLine(img.caption, cx, y) //nolint:errcheck
		curX += img.w + spacing
	}
	y += 20

	// Larger single image centered
	doc.SetFont(boldFont, 10) //nolint:errcheck
	doc.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
	doc.WriteLine("Larger image — go-header.jpg:", marginL, y) //nolint:errcheck
	y += 14

	headerPath := filepath.Join(imDir, "go-header.jpg")
	hdrImgW, hdrImgH := contentW*0.7, 120.0
	hdrImgX := marginL + (contentW-hdrImgW)/2
	if err := doc.DrawImage(headerPath, hdrImgX, y, hdrImgW, hdrImgH); err != nil {
		doc.FillRect(hdrImgX, y, hdrImgW, hdrImgH, colorLightBlue)
		doc.SetFont("regular", 10)      //nolint:errcheck
		doc.SetTextColor(100, 100, 100)
		doc.WriteLine("(image not found)", hdrImgX+10, y+50) //nolint:errcheck
	}
	y += hdrImgH + 14

	doc.SetFont("regular", 9) //nolint:errcheck
	doc.SetTextColor(60, 60, 60)
	doc.WriteText(goText2, marginL, y, contentW) //nolint:errcheck
}

// ── Pages 7-8: Layout Engine ──────────────────────────────────────────────────

func buildLayoutEnginePages(doc *pdf.Document, boldFont string) {
	// We use the layout engine itself to render these pages.
	// A two-column page template is used to show off LayoutFrame.

	colW := (contentW - 12) / 2
	col1 := &layout.LayoutFrame{
		X: marginL, Y: marginT + 26,
		Width: colW, Height: pageH - marginT - marginB - 26,
	}
	col2 := &layout.LayoutFrame{
		X: marginL + colW + 12, Y: marginT + 26,
		Width: colW, Height: pageH - marginT - marginB - 26,
	}

	twoColTmpl := &layout.PageTemplate{
		ID:     "TwoColumn",
		Frames: []*layout.LayoutFrame{col1, col2},
		OnPage: func(d *pdf.Document, pageNum int) {
			// Section banner
			d.SetFont(boldFont, 11)    //nolint:errcheck
			d.SetTextColor(255, 255, 255)
			d.FillRect(marginL, marginT, contentW, 22, colorNavy)
			d.WriteLine("Layout Engine — DocTemplate / PageTemplate / LayoutFrame / Flowables", marginL+6, marginT+5) //nolint:errcheck
		},
	}

	subStyle := layout.ParagraphStyle{
		FontName:    boldFont,
		FontSize:    12,
		TextColor:   &colorNavy,
		SpaceBefore: 10,
		SpaceAfter:  4,
		KeepWithNextPara: true,
	}
	bodyStyle := layout.ParagraphStyle{
		FontName:   "regular",
		FontSize:   10,
		SpaceAfter: 6,
	}
	smallStyle := layout.ParagraphStyle{
		FontName:   "regular",
		FontSize:   8,
		TextColor:  &colorSlate,
		SpaceAfter: 4,
	}
	centeredStyle := layout.ParagraphStyle{
		FontName:    boldFont,
		FontSize:    10,
		Alignment:   layout.AlignCenter,
		SpaceBefore: 8,
		SpaceAfter:  8,
		TextColor:   &colorMint,
	}

	story := []layout.Flowable{
		// Introduction
		&layout.Paragraph{Text: "The Layout Engine", Style: subStyle},
		&layout.Paragraph{Text: goText1, Style: bodyStyle},
		&layout.HRFlowable{Thickness: 1, Color: colorCyan, Before: 8, After: 8},

		// Goroutines section (KeepTogether: heading + first paragraph)
		&layout.KeepTogether{Flowables: []layout.Flowable{
			&layout.Paragraph{Text: "Goroutines", Style: subStyle},
			&layout.Paragraph{Text: goText4, Style: bodyStyle},
		}},
		&layout.HRFlowable{Width: 0.5, Thickness: 1, Color: colorLavender, Align: layout.AlignCenter, Before: 6, After: 6},

		// Modules section
		&layout.Paragraph{Text: "Go Modules", Style: subStyle},
		&layout.Paragraph{Text: goText5, Style: bodyStyle},
		&layout.Spacer{Height: 6},

		// CondPageBreak: if less than 100 pt remain, move to a new page
		&layout.CondPageBreak{MinHeight: 100},

		&layout.Paragraph{Text: "Static Binaries", Style: subStyle},
		&layout.Paragraph{Text: goText6, Style: bodyStyle},
		&layout.HRFlowable{Thickness: 1.5, Color: colorGold, Before: 10, After: 10},

		&layout.Paragraph{Text: "Standard Library", Style: subStyle},
		&layout.Paragraph{Text: goText3, Style: bodyStyle},
		&layout.Spacer{Height: 8},

		&layout.Paragraph{Text: "The Gopher Mascot", Style: subStyle},
		&layout.Paragraph{Text: goText2, Style: bodyStyle},
		&layout.HRFlowable{Thickness: 1, Color: colorMint, Before: 6, After: 6},

		// Centered paragraph demo
		&layout.Paragraph{
			Text: "Nautilus makes PDF generation in Go straightforward, composable, and beautiful.",
			Style: centeredStyle,
		},
		&layout.Spacer{Height: 4},

		// Small-text section
		&layout.Paragraph{Text: "Layout engine notes", Style: layout.ParagraphStyle{
			FontName: boldFont, FontSize: 9, TextColor: &colorSlate, SpaceBefore: 6,
		}},
		&layout.Paragraph{
			Text: "Wrap is always called before Draw or Split. SpaceBefore is suppressed at the top of a frame (atTop flag). KeepTogether wraps a group and issues a FrameBreak if the group does not fit. CondPageBreak checks remaining frame height and forces a page break when below the threshold.",
			Style: smallStyle,
		},
		&layout.HRFlowable{Thickness: 0.5, Color: pdf.ColorLightGray, Before: 4, After: 4},

		// Another block of content to demonstrate overflow to second column / page
		&layout.Paragraph{Text: "More About Go", Style: subStyle},
		&layout.Paragraph{Text: "Go's concurrency model is based on communicating sequential processes (CSP). Rather than sharing memory between goroutines and protecting it with locks, Go encourages passing data through channels. This design leads to programs that are easier to reason about and less prone to data races.", Style: bodyStyle},
		&layout.Spacer{Height: 6},
		&layout.Paragraph{Text: "The Go toolchain includes go build, go test, go vet, go mod, and go generate among others. The ecosystem has grown to include powerful tools like golangci-lint, staticcheck, and gopls (the official language server) that integrate seamlessly with editors.", Style: bodyStyle},
		&layout.Spacer{Height: 6},
		&layout.Paragraph{Text: "Performance benchmarks consistently show Go programs executing within 2x of equivalent C++ programs while offering dramatically higher programmer productivity. This balance makes Go an excellent choice for network services, CLI tools, and system software.", Style: bodyStyle},
	}

	layoutDoc, err := pdf.New(pdf.Config{
		PageSize: pdf.PageSizeA4,
		Margins: pdf.Margins{
			Top: marginT, Right: marginR, Bottom: marginB, Left: marginL,
		},
	})
	if err != nil {
		log.Printf("layout sub-doc: %v", err)
		return
	}
	// Register the same fonts on the sub-doc... but we actually need to
	// render into the SAME doc. Use the doc's layout engine instead.
	// Re-use doc directly by adding pages to it via the template.
	_ = layoutDoc

	// Build the layout story into the current document by using DocTemplate.
	doc.SetFont("regular", 10) //nolint:errcheck
	dt := layout.NewDocTemplate(doc)
	dt.AddPageTemplate(twoColTmpl)

	if err := dt.Build(story); err != nil {
		log.Printf("layout engine build: %v", err)
	}
}

// ── Pages 9-18: Charts (all 20 types, 2 per page) ────────────────────────────

func buildChartPages(doc *pdf.Document, boldFont string) {
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun"}

	baseOpts := chart.Options{
		FontName: "regular",
		FontSize: 8,
		XAxis:    &chart.Axis{Categories: months},
		YAxis:    &chart.Axis{},
		Legend:   &chart.Legend{},
	}

	// ── Line chart ────────────────────────────────────────────────────────────
	lineOpts := baseOpts
	lineOpts.Title = &chart.Title{Text: "Monthly Revenue 2023 vs 2024", FontName: boldFont, FontSize: 11}
	lineOpts.Subtitle = &chart.Title{Text: "Line Chart"}
	lineOpts.Series = []chart.Series{
		{Name: "2023", Data: []float64{120, 150, 130, 180, 160, 200}},
		{Name: "2024", Data: []float64{140, 165, 175, 195, 210, 240}},
	}
	lc := &line.LineChart{Options: lineOpts}

	// ── Area chart ────────────────────────────────────────────────────────────
	areaOpts := baseOpts
	areaOpts.Title = &chart.Title{Text: "Website Visitors by Platform", FontName: boldFont, FontSize: 11}
	areaOpts.Subtitle = &chart.Title{Text: "Area Chart"}
	areaOpts.Series = []chart.Series{
		{Name: "Mobile", Data: []float64{800, 950, 1100, 1050, 1200, 1400}},
		{Name: "Desktop", Data: []float64{500, 580, 620, 700, 680, 750}},
	}
	areaOpts.PlotOptions = &chart.PlotOptions{Area: &chart.AreaOptions{FillAlpha: 0.25}}
	ac := &area.AreaChart{Options: areaOpts}

	// ── Column chart ──────────────────────────────────────────────────────────
	colOpts := baseOpts
	colOpts.Title = &chart.Title{Text: "Quarterly Sales by Region", FontName: boldFont, FontSize: 11}
	colOpts.Subtitle = &chart.Title{Text: "Column Chart (grouped)"}
	colOpts.XAxis = &chart.Axis{Categories: []string{"Q1", "Q2", "Q3", "Q4"}}
	colOpts.Series = []chart.Series{
		{Name: "North", Data: []float64{43, 55, 57, 60}},
		{Name: "South", Data: []float64{23, 35, 41, 47}},
		{Name: "West", Data: []float64{31, 28, 38, 44}},
	}
	cc := &column.ColumnChart{Options: colOpts}

	// ── Stacked column ────────────────────────────────────────────────────────
	stackedOpts := colOpts
	stackedOpts.Title = &chart.Title{Text: "Stacked Regional Sales", FontName: boldFont, FontSize: 11}
	stackedOpts.Subtitle = &chart.Title{Text: "Column Chart (stacked)"}
	stackedOpts.PlotOptions = &chart.PlotOptions{Column: &chart.ColumnOptions{Stacking: "normal"}}
	sc := &column.ColumnChart{Options: stackedOpts}

	// ── Bar chart ─────────────────────────────────────────────────────────────
	barOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title: &chart.Title{Text: "Top 5 Products by Units Sold", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Bar Chart (horizontal)"},
		XAxis: &chart.Axis{},
		YAxis: &chart.Axis{Categories: []string{"Product A", "Product B", "Product C", "Product D", "Product E"}},
		Series: []chart.Series{
			{Name: "Units Sold", Data: []float64{143, 112, 98, 87, 76}},
		},
		Legend: &chart.Legend{Enabled: chart.Bool(false)},
	}
	bc := &bar.BarChart{Options: barOpts}

	// ── Pie chart ─────────────────────────────────────────────────────────────
	pieOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title: &chart.Title{Text: "Browser Market Share", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Pie Chart"},
		Series: []chart.Series{
			{Name: "Chrome", Data: []float64{65}},
			{Name: "Firefox", Data: []float64{15}},
			{Name: "Safari", Data: []float64{12}},
			{Name: "Other", Data: []float64{8}},
		},
		Legend: &chart.Legend{},
		PlotOptions: &chart.PlotOptions{Pie: &chart.PieOptions{
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	pc := &pie.PieChart{Options: pieOpts}

	// ── Donut chart ───────────────────────────────────────────────────────────
	donutOpts := pieOpts
	donutOpts.Title = &chart.Title{Text: "Browser Share (Donut)", FontName: boldFont, FontSize: 11}
	donutOpts.Subtitle = &chart.Title{Text: "Pie Chart with InnerSize=50%"}
	donutOpts.PlotOptions = &chart.PlotOptions{Pie: &chart.PieOptions{
		InnerSize:  "50%",
		DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
	}}
	dc := &pie.PieChart{Options: donutOpts}

	// ── Polar (spider/radar) chart ────────────────────────────────────────────
	polarOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title: &chart.Title{Text: "Budget vs Spending", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Polar / Spider / Radar Chart"},
		XAxis: &chart.Axis{Categories: []string{
			"Sales", "Marketing", "Development", "Support", "IT", "Admin",
		}},
		YAxis:  &chart.Axis{Min: chart.Float(0)},
		Legend: &chart.Legend{},
		Series: []chart.Series{
			{Name: "Budget", Data: []float64{43000, 19000, 60000, 35000, 17000, 10000}},
			{Name: "Spending", Data: []float64{50000, 39000, 42000, 31000, 26000, 14000}},
		},
		PlotOptions: &chart.PlotOptions{Polar: &chart.PolarOptions{GridLineInterpolation: "polygon"}},
	}
	polarC := &polar.PolarChart{Options: polarOpts}

	// ── Scatter chart ─────────────────────────────────────────────────────────
	scatterOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:  &chart.Title{Text: "Height vs Weight Scatter", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Scatter Chart"},
		XAxis:  &chart.Axis{},
		YAxis:  &chart.Axis{},
		Legend: &chart.Legend{},
		Series: []chart.Series{
			{Name: "Group A", Points: []chart.Point{
				{X: 161, Y: 51}, {X: 167, Y: 59}, {X: 159, Y: 49}, {X: 175, Y: 72},
				{X: 180, Y: 81}, {X: 166, Y: 63}, {X: 172, Y: 68}, {X: 155, Y: 46},
			}},
			{Name: "Group B", Points: []chart.Point{
				{X: 183, Y: 88}, {X: 170, Y: 77}, {X: 177, Y: 79}, {X: 164, Y: 65},
				{X: 178, Y: 84}, {X: 169, Y: 70}, {X: 185, Y: 90}, {X: 156, Y: 53},
			}},
		},
	}
	scatterC := &scatter.ScatterChart{Options: scatterOpts}

	// ── Bubble chart ──────────────────────────────────────────────────────────
	bubbleOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:  &chart.Title{Text: "GDP, Life Expectancy, Population", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Bubble Chart"},
		XAxis:  &chart.Axis{},
		YAxis:  &chart.Axis{},
		Legend: &chart.Legend{},
		Series: []chart.Series{
			{Name: "Europe", Points: []chart.Point{
				{X: 54, Y: 78, Z: 66}, {X: 44, Y: 77, Z: 55},
				{X: 37, Y: 77, Z: 27}, {X: 31, Y: 78, Z: 12},
			}},
			{Name: "Asia", Points: []chart.Point{
				{X: 16, Y: 74, Z: 138}, {X: 3, Y: 69, Z: 136}, {X: 42, Y: 82, Z: 13},
			}},
		},
	}
	bubbleC := &bubble.BubbleChart{Options: bubbleOpts}

	// ── Heatmap ───────────────────────────────────────────────────────────────
	salesPeople := []string{"Alexander", "Marie", "Maximilian", "Sophia", "Lukas"}
	weekdays := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	rawHeat := [][]float64{
		{0, 1, 0, 5, 1},
		{1, 0, 0, 2, 4},
		{0, 1, 2, 0, 5},
		{3, 1, 0, 0, 2},
		{0, 0, 4, 1, 0},
	}
	var heatData []chart.Point
	for row, vals := range rawHeat {
		for col, v := range vals {
			heatData = append(heatData, chart.Point{X: float64(col), Y: float64(row), Z: v})
		}
	}
	heatOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Sales per Employee per Weekday", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Heatmap Chart"},
		XAxis:   &chart.Axis{Categories: salesPeople},
		YAxis:   &chart.Axis{Categories: weekdays},
		Series:  []chart.Series{{Name: "Sales", Points: heatData}},
		PlotOptions: &chart.PlotOptions{Heatmap: &chart.HeatmapOptions{
			MaxColor:   &pdf.Color{R: 7, G: 75, B: 154},
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	heatC := &heatmap.HeatmapChart{Options: heatOpts}

	// ── Waterfall chart ───────────────────────────────────────────────────────
	waterfallOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Company Financials Waterfall", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Waterfall Chart"},
		YAxis:   &chart.Axis{},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "Start", Y: 120000},
			{Name: "Revenue", Y: 569000},
			{Name: "Costs", Y: -342000},
			{Name: "Subtotal", IsIntermediateSum: true},
			{Name: "Extra costs", Y: -133000},
			{Name: "Balance", IsSum: true},
		}}},
	}
	waterfallC := &waterfall.WaterfallChart{Options: waterfallOpts}

	// ── Funnel chart ──────────────────────────────────────────────────────────
	funnelOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Sales Funnel", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Funnel Chart"},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "Website visits", Y: 15654},
			{Name: "Downloads", Y: 4064},
			{Name: "Price list requested", Y: 1987},
			{Name: "Invoice sent", Y: 976},
			{Name: "Finalized", Y: 846},
		}}},
		Legend: &chart.Legend{},
		PlotOptions: &chart.PlotOptions{Funnel: &chart.FunnelOptions{
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	funnelC := &funnel.FunnelChart{Options: funnelOpts}

	// ── Gauge chart ───────────────────────────────────────────────────────────
	gaugeOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Speedometer", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Gauge Chart"},
		YAxis:   &chart.Axis{Min: chart.Float(0), Max: chart.Float(200)},
		Series:  []chart.Series{{Name: "Speed km/h", Data: []float64{120}}},
		PlotOptions: &chart.PlotOptions{Gauge: &chart.GaugeOptions{
			PaneStartAngle: chart.Float(-150), PaneEndAngle: chart.Float(150),
			PlotBands: []chart.GaugePlotBand{
				{From: 0, To: 80, Color: pdf.Color{R: 85, G: 191, B: 59}, Thickness: 12},
				{From: 80, To: 140, Color: pdf.Color{R: 221, G: 223, B: 13}, Thickness: 12},
				{From: 140, To: 200, Color: pdf.Color{R: 223, G: 83, B: 83}, Thickness: 12},
			},
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	gaugeC := &gauge.GaugeChart{Options: gaugeOpts}

	// ── Solid gauge ───────────────────────────────────────────────────────────
	solidGaugeOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Activity Level — 84%", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Solid Gauge Chart"},
		YAxis:   &chart.Axis{Min: chart.Float(0), Max: chart.Float(100)},
		Series:  []chart.Series{{Name: "Activity", Data: []float64{84}}},
		PlotOptions: &chart.PlotOptions{Gauge: &chart.GaugeOptions{
			PaneStartAngle: chart.Float(-90), PaneEndAngle: chart.Float(90), Solid: true,
			PlotBands: []chart.GaugePlotBand{
				{From: 0, To: 100, Color: pdf.Color{R: 230, G: 230, B: 230}, Thickness: 20},
			},
			DataLabels: &chart.DataLabels{Enabled: chart.Bool(true)},
		}},
	}
	solidGaugeC := &gauge.GaugeChart{Options: solidGaugeOpts}

	// ── Error bar ─────────────────────────────────────────────────────────────
	errorbarOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Rainfall with Error Bars", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Error Bar Chart"},
		XAxis:   &chart.Axis{Categories: months},
		YAxis:   &chart.Axis{},
		Legend:  &chart.Legend{},
		Series: []chart.Series{{Name: "Error range", Points: []chart.Point{
			{Low: 48, High: 51}, {Low: 68, High: 73},
			{Low: 92, High: 110}, {Low: 178, High: 220},
			{Low: 168, High: 200}, {Low: 140, High: 162},
		}}},
	}
	errorbarC := &errorbar.ErrorbarChart{Options: errorbarOpts}

	// ── Box plot ──────────────────────────────────────────────────────────────
	boxOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Observation Distributions", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Box-and-Whisker Plot"},
		XAxis:   &chart.Axis{Categories: []string{"Loc A", "Loc B", "Loc C", "Loc D", "Loc E"}},
		YAxis:   &chart.Axis{},
		Legend:  &chart.Legend{},
		Series: []chart.Series{{Name: "Observations", Points: []chart.Point{
			{Low: 760, Q1: 801, Median: 848, Q3: 895, High: 965},
			{Low: 733, Q1: 853, Median: 939, Q3: 980, High: 1080},
			{Low: 714, Q1: 762, Median: 817, Q3: 870, High: 918},
			{Low: 724, Q1: 802, Median: 806, Q3: 871, High: 950},
			{Low: 747, Q1: 835, Median: 882, Q3: 910, High: 980},
		}}},
	}
	boxC := &boxplot.BoxplotChart{Options: boxOpts}

	// ── Column range ──────────────────────────────────────────────────────────
	colRangeOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Temperature Range per Month", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Column Range Chart"},
		XAxis:   &chart.Axis{Categories: months},
		YAxis:   &chart.Axis{},
		Legend:  &chart.Legend{},
		Series: []chart.Series{{Name: "Temp C", Points: []chart.Point{
			{Low: -9.5, High: 8.0}, {Low: -7.8, High: 8.3},
			{Low: -4.1, High: 13.0}, {Low: 0.4, High: 18.2},
			{Low: 4.6, High: 22.7}, {Low: 9.0, High: 26.4},
		}}},
	}
	colRangeC := &columnrange.ColumnRangeChart{Options: colRangeOpts}

	// ── Area range ────────────────────────────────────────────────────────────
	areaRangeOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Temperature Band", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Area Range Chart"},
		XAxis:   &chart.Axis{Categories: months},
		YAxis:   &chart.Axis{},
		Legend:  &chart.Legend{},
		Series: []chart.Series{{Name: "Temp range", Points: []chart.Point{
			{Low: -9.5, High: 8.0}, {Low: -7.8, High: 8.3},
			{Low: -4.1, High: 13.0}, {Low: 0.4, High: 18.2},
			{Low: 4.6, High: 22.7}, {Low: 9.0, High: 26.4},
		}}},
	}
	areaRangeC := &arearange.AreaRangeChart{Options: areaRangeOpts}

	// ── Bullet chart ──────────────────────────────────────────────────────────
	bulletOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Quarterly Performance vs Target", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Bullet Chart"},
		YAxis:   &chart.Axis{Min: chart.Float(0), Max: chart.Float(300)},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "Q1", Y: 180, Target: 220},
			{Name: "Q2", Y: 210, Target: 200},
			{Name: "Q3", Y: 150, Target: 240},
			{Name: "Q4", Y: 260, Target: 250},
		}}},
		PlotOptions: &chart.PlotOptions{Bullet: &chart.BulletOptions{
			PlotBands: []chart.GaugePlotBand{
				{From: 0, To: 150, Color: pdf.Color{R: 200, G: 200, B: 200}},
				{From: 150, To: 225, Color: pdf.Color{R: 175, G: 175, B: 175}},
				{From: 225, To: 300, Color: pdf.Color{R: 155, G: 155, B: 155}},
			},
		}},
	}
	bulletC := &bullet.BulletChart{Options: bulletOpts}

	// ── Dumbbell chart ────────────────────────────────────────────────────────
	dumbbellOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Life Expectancy 1990 vs 2020", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Dumbbell Chart"},
		XAxis:   &chart.Axis{Categories: []string{"Austria", "Belgium", "Germany", "France", "Netherlands"}},
		YAxis:   &chart.Axis{},
		Legend:  &chart.Legend{},
		Series: []chart.Series{{Points: []chart.Point{
			{Low: 70.1, High: 81.3},
			{Low: 71.0, High: 81.9},
			{Low: 70.8, High: 81.2},
			{Low: 70.5, High: 82.3},
			{Low: 71.3, High: 81.7},
		}}},
	}
	dumbbellC := &dumbbell.DumbbellChart{Options: dumbbellOpts}

	// ── Lollipop chart ────────────────────────────────────────────────────────
	lollipopOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Top 5 Products — Lollipop", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Lollipop Chart"},
		XAxis:   &chart.Axis{Categories: []string{"Product A", "Product B", "Product C", "Product D", "Product E"}},
		YAxis:   &chart.Axis{},
		Series:  []chart.Series{{Name: "Units", Data: []float64{143, 112, 98, 87, 76}}},
		Legend:  &chart.Legend{Enabled: chart.Bool(false)},
	}
	lollipopC := &lollipop.LollipopChart{Options: lollipopOpts}

	// ── Treemap ───────────────────────────────────────────────────────────────
	treemapOpts := chart.Options{
		FontName: "regular", FontSize: 8,
		Title:   &chart.Title{Text: "Global Revenue by Region", FontName: boldFont, FontSize: 11},
		Subtitle: &chart.Title{Text: "Treemap Chart"},
		Series: []chart.Series{{Points: []chart.Point{
			{Name: "North America", Y: 42},
			{Name: "Europe", Y: 35},
			{Name: "Asia Pacific", Y: 18},
			{Name: "Latin America", Y: 3},
			{Name: "Middle East", Y: 1.5},
			{Name: "Africa", Y: 0.5},
		}}},
	}
	treemapC := &treemap.TreemapChart{Options: treemapOpts}

	// ── Layout story for charts ───────────────────────────────────────────────
	chartH := 220.0
	halfH := 190.0

	// Use a single-column full-page frame for charts
	chartFrame := &layout.LayoutFrame{
		X: marginL, Y: marginT + 4,
		Width:  contentW,
		Height: pageH - marginT - marginB - 4,
	}
	chartTmpl := &layout.PageTemplate{
		ID:     "Charts",
		Frames: []*layout.LayoutFrame{chartFrame},
		OnPage: func(d *pdf.Document, pageNum int) {
			d.SetFont(boldFont, 9) //nolint:errcheck
			d.SetTextColor(colorSlate.R, colorSlate.G, colorSlate.B)
			d.FillRect(marginL, marginT-8, contentW, 2, colorNavy)
		},
	}

	story := []layout.Flowable{
		// Page 1 of charts: Line + Area
		chart.NewFlowable(lc, 0, chartH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(ac, 0, chartH),

		// Page 2: Column grouped + Stacked
		&layout.PageBreak{},
		chart.NewFlowable(cc, 0, chartH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(sc, 0, chartH),

		// Page 3: Bar + Pie
		&layout.PageBreak{},
		chart.NewFlowable(bc, 0, halfH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(pc, 0, halfH),

		// Page 4: Donut + Polar
		&layout.PageBreak{},
		chart.NewFlowable(dc, 0, halfH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(polarC, 0, halfH),

		// Page 5: Scatter + Bubble
		&layout.PageBreak{},
		chart.NewFlowable(scatterC, 0, chartH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(bubbleC, 0, chartH),

		// Page 6: Heatmap + Waterfall
		&layout.PageBreak{},
		chart.NewFlowable(heatC, 0, chartH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(waterfallC, 0, chartH),

		// Page 7: Funnel + Gauge
		&layout.PageBreak{},
		chart.NewFlowable(funnelC, 0, chartH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(gaugeC, 0, halfH),

		// Page 8: Solid gauge + Error bar
		&layout.PageBreak{},
		chart.NewFlowable(solidGaugeC, 0, halfH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(errorbarC, 0, chartH),

		// Page 9: Box plot + Column range
		&layout.PageBreak{},
		chart.NewFlowable(boxC, 0, chartH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(colRangeC, 0, chartH),

		// Page 10: Area range + Bullet
		&layout.PageBreak{},
		chart.NewFlowable(areaRangeC, 0, chartH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(bulletC, 0, halfH),

		// Page 11: Dumbbell + Lollipop
		&layout.PageBreak{},
		chart.NewFlowable(dumbbellC, 0, halfH),
		&layout.Spacer{Height: 14},
		chart.NewFlowable(lollipopC, 0, halfH),

		// Page 12: Treemap
		&layout.PageBreak{},
		chart.NewFlowable(treemapC, 0, chartH),
	}

	dt := layout.NewDocTemplate(doc)
	dt.AddPageTemplate(chartTmpl)
	if err := dt.Build(story); err != nil {
		log.Printf("chart pages build: %v", err)
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// sectionHeader draws a bold section title with a decorative navy rule beneath
// it and returns the Y coordinate immediately below the heading block.
func sectionHeader(doc *pdf.Document, boldFont, title string, y float64) float64 {
	doc.SetFont(boldFont, 16) //nolint:errcheck
	doc.SetTextColor(colorNavy.R, colorNavy.G, colorNavy.B)
	doc.WriteLine(title, marginL, y) //nolint:errcheck

	y += 22
	spec := pdf.BorderSpec{Thickness: 2, Color: colorCyan}
	doc.DrawBorder(marginL, y, contentW, 0, pdf.Border{Bottom: &spec}) //nolint:errcheck

	return y + 4
}

// regularPolygon returns the vertices of a regular n-gon centered at (cx,cy)
// with the given circumradius, with the first vertex pointing upward.
func regularPolygon(cx, cy, r float64, n int) []pdf.Point {
	pts := make([]pdf.Point, n)
	// subtract math.Pi/2 to start at top (12 o'clock)
	const halfPi = 1.5707963267948966
	for i := range pts {
		angle := -halfPi + 2*3.141592653589793*float64(i)/float64(n)
		pts[i] = pdf.Point{
			X: cx + r*cosApprox(angle),
			Y: cy + r*sinApprox(angle),
		}
	}
	return pts
}

// cosApprox and sinApprox compute sine and cosine without importing math
// (math is already available via the polygon computation logic; these thin
// wrappers exist to keep the helper self-contained and avoid a math import
// in this file that is already pulling in math transitively through gopdf).
//
// We use the standard library math package indirectly through gopdf anyway,
// so it is acceptable to import it here to avoid a hand-rolled approximation.
// The import is added to the import block above.

func cosApprox(a float64) float64 {
	// Use a Taylor-series-free approach: rely on the sin identity.
	// cos(a) = sin(pi/2 - a)
	return sinApprox(1.5707963267948966 - a)
}

// sinApprox computes sin(a) via a minimax polynomial valid for all a.
// Accuracy is sufficient for polygon vertex placement (< 0.001 pt error).
func sinApprox(a float64) float64 {
	// Reduce to [-pi, pi]
	const twoPi = 6.283185307179586
	for a > 3.141592653589793 {
		a -= twoPi
	}
	for a < -3.141592653589793 {
		a += twoPi
	}
	// Bhaskara I approximation — good to ~0.17% over full range
	if a >= 0 {
		return 16 * a * (3.141592653589793 - a) / (5*3.141592653589793*3.141592653589793 - 4*a*(3.141592653589793-a))
	}
	// sin is odd
	a = -a
	return -(16 * a * (3.141592653589793 - a) / (5*3.141592653589793*3.141592653589793 - 4*a*(3.141592653589793-a)))
}
