// Command rtl demonstrates right-to-left text rendering with the nautilus
// PDF library.
//
// It generates output.pdf with examples of:
//   - Arabic text: contextual letter shaping + lam-alef ligatures + BiDi reordering
//   - Hebrew text: BiDi reordering only (no contextual shaping needed)
//   - Multi-line word-wrapped RTL paragraphs
//   - Mixed RTL/LTR content on the same line
//   - RTL text inside a Frame
//
// # Usage
//
//	go run ./examples/rtl \
//	    -arabic /System/Library/Fonts/Supplemental/DecoTypeNaskh.ttc \
//	    -hebrew /System/Library/Fonts/SFHebrew.ttf \
//	    -latin  /Library/Fonts/Lato-Regular.ttf \
//	    -out    output.pdf
//
// All font flags are optional; the defaults shown above work on macOS.
//
// # Font requirements
//
// The Arabic font must include the Unicode Arabic Presentation Forms-B block
// (U+FE70–U+FEFF).  DecoType Naskh (macOS) and Amiri / Noto Naskh Arabic
// (free download) both qualify.
// The Hebrew font requires the Hebrew Unicode block (U+0590–U+05FF).
// SFHebrew (macOS) or any Noto Sans Hebrew font works.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gvanbeck/nautilus/pdf"
	"github.com/gvanbeck/nautilus/pdf/rtl"
)

func main() {
	arabicFont := flag.String("arabic", "/System/Library/Fonts/Supplemental/Mishafi.ttf",
		"Arabic TTF/OTF font with Presentation Forms-B (U+FE70–U+FEFF);\n"+
			"\talternatives on macOS: /System/Library/Fonts/Supplemental/Arial Unicode.ttf\n"+
			"\t                       /System/Library/Fonts/Supplemental/Farisi.ttf")
	hebrewFont := flag.String("hebrew", "/System/Library/Fonts/SFHebrew.ttf",
		"Hebrew TTF/OTF font")
	latinFont := flag.String("latin", "/Library/Fonts/Lato-Regular.ttf",
		"Latin TTF/OTF font for labels and LTR content")
	outPath := flag.String("out", "output.pdf", "output PDF file path")
	flag.Parse()

	// ── Create document ───────────────────────────────────────────────────
	doc, err := pdf.New(pdf.Config{
		PageSize:         pdf.PageSizeA4,
		DefaultFontSize:  14,
		LineHeightFactor: 1.6,
	})
	if err != nil {
		log.Fatalf("create document: %v", err)
	}

	// ── Register fonts ────────────────────────────────────────────────────
	for _, reg := range []struct{ name, path string }{
		{"arabic", *arabicFont},
		{"hebrew", *hebrewFont},
		{"latin", *latinFont},
	} {
		if err := doc.RegisterFont(reg.name, reg.path); err != nil {
			log.Fatalf("register font %q from %q: %v", reg.name, reg.path, err)
		}
	}

	// ── Layout helpers ────────────────────────────────────────────────────
	const (
		marginLeft  = 60.0
		marginRight = 60.0
		fontSize    = 16.0
		labelSize   = 9.0
	)
	pageW    := doc.PageWidth()
	rightX   := pageW - marginRight
	contentW := pageW - marginLeft - marginRight

	setArabic := func() { doc.SetFont("arabic", fontSize) } //nolint
	setHebrew := func() { doc.SetFont("hebrew", fontSize) } //nolint
	setLatin  := func() { doc.SetFont("latin", fontSize) }  //nolint

	// label prints a small gray annotation above a sample.
	label := func(text string, y float64) float64 {
		doc.SetFont("latin", labelSize)  //nolint
		doc.SetTextColor(150, 150, 150)
		doc.WriteLine(text, marginLeft, y) //nolint
		return y + labelSize*1.4
	}

	// section prints a bold-style section heading.
	section := func(title string, y float64) float64 {
		doc.SetFont("latin", 11)  //nolint
		doc.SetTextColor(20, 20, 100)
		endY, _ := doc.WriteText(title, marginLeft, y, contentW)
		return endY + 6
	}

	// hline draws a thin separator.
	hline := func(y float64) float64 {
		doc.DrawBorder(marginLeft, y, contentW, 0, pdf.Border{ //nolint
			Top: &pdf.BorderSpec{Thickness: 0.4, Color: pdf.ColorLightGray, Pattern: pdf.PatternSolid},
		})
		return y + 10
	}

	// ── Page 1 ────────────────────────────────────────────────────────────
	doc.AddPage()
	y := 50.0

	// Title
	doc.SetFont("latin", 22) //nolint
	doc.SetTextColor(20, 20, 100)
	doc.WriteLine("Right-to-left text rendering", marginLeft, y) //nolint
	y += 32

	doc.SetFont("latin", 10) //nolint
	doc.SetTextColor(100, 100, 100)
	endY, _ := doc.WriteText(
		"Demonstrates Arabic contextual shaping (presentation forms + lam-alef "+
			"ligatures) and Hebrew BiDi reordering, both using WriteLineRTL and "+
			"WriteTextRTL.",
		marginLeft, y, contentW,
	)
	y = endY + 20

	// ── Section 1: Arabic single words ───────────────────────────────────
	y = section("1. Arabic — single words", y)

	words := []struct{ latin, arabic string }{
		{"kitab (book) — كتاب", "كتاب"},
		{"bab (door) — باب", "باب"},
		{"marhaba (hello) — مرحبا", "مرحبا"},
		{"lam-alef ligature — لا", "لا"},
		{"sentence — كيف حالك", "كيف حالك"},
	}

	for _, w := range words {
		y = label(w.latin, y)
		setArabic()
		doc.SetTextColor(30, 30, 30)
		doc.WriteLineRTL(rtl.Shape(w.arabic), rightX, y) //nolint
		y += fontSize*1.6 + 8
	}

	y = hline(y + 4)

	// ── Section 2: Hebrew single words ───────────────────────────────────
	y = section("2. Hebrew — single words", y)

	hebrewWords := []struct{ latin, hebrew string }{
		{"shalom (peace/hello) — שלום", "שלום"},
		{"todah (thank you) — תודה", "תודה"},
		{"Israel — ישראל", "ישראל"},
		{"sentence — שלום עולם", "שלום עולם"},
	}

	for _, w := range hebrewWords {
		y = label(w.latin, y)
		setHebrew()
		doc.SetTextColor(30, 30, 30)
		doc.WriteLineRTL(rtl.Shape(w.hebrew), rightX, y) //nolint
		y += fontSize*1.6 + 8
	}

	y = hline(y + 4)

	// ── Section 3: Mixed RTL/LTR ──────────────────────────────────────────
	y = section("3. Mixed RTL + numbers", y)
	y = label("Arabic numerals embedded in Arabic text", y)

	// Numbers in Arabic text are LTR runs within an RTL paragraph.
	// rtl.Shape + WriteLineRTL handles the run ordering correctly.
	mixed := "السعر 42 دولار"
	setArabic()
	doc.SetTextColor(30, 30, 30)
	doc.WriteLineRTL(rtl.Shape(mixed), rightX, y) //nolint
	y += fontSize*1.6 + 8

	y = label("Hebrew with Latin brand name", y)
	mixedHe := "Google הוא מנוע חיפוש"
	setHebrew()
	doc.WriteLineRTL(rtl.Shape(mixedHe), rightX, y) //nolint
	y += fontSize*1.6 + 20

	// ── Page 2 ────────────────────────────────────────────────────────────
	doc.AddPage()
	y = 50.0

	// ── Section 4: Multi-line word-wrapped RTL ────────────────────────────
	y = section("4. Multi-line word-wrapped Arabic (WriteTextRTL)", y)
	y = label("Shaping and BiDi reordering applied per line — word order preserved across breaks.", y)

	setArabic()
	doc.SetTextColor(30, 30, 30)
	arabicPara := "في البدء كان الكلمة وكانت الكلمة عند الله وكان الله هو الكلمة"
	endY, _ = doc.WriteTextRTL(arabicPara, rightX, y, contentW)
	y = endY + 16

	y = label("Hebrew paragraph:", y)
	setHebrew()
	hebrewPara := "בְּרֵאשִׁית בָּרָא אֱלֹהִים אֵת הַשָּׁמַיִם וְאֵת הָאָרֶץ"
	endY, _ = doc.WriteTextRTL(hebrewPara, rightX, y, contentW)
	y = endY + 20

	y = hline(y + 4)

	// ── Section 5: RTL inside a Frame ─────────────────────────────────────
	y = section("5. RTL text inside a Frame", y)
	y = label("Frame with RTL content — right edge of content area used as anchor.", y)

	bg := pdf.Color{R: 255, G: 252, B: 240}
	f := doc.NewFrame(pdf.FrameConfig{
		X: marginLeft, Y: y, Width: contentW,
		Background: &bg,
		Border: pdf.Border{
			Right: &pdf.BorderSpec{Thickness: 4, Color: pdf.ColorOrange, Pattern: pdf.PatternSolid},
		},
		Padding: pdf.Padding{Top: 10, Right: 16, Bottom: 10, Left: 12},
	})

	f.SetFont("arabic", fontSize) //nolint
	f.SetTextColor(30, 30, 30)
	f.WriteTextRTL("مرحبا بالعالم كيف حالك اليوم") //nolint
	f.Advance(8)
	f.SetFont("hebrew", fontSize) //nolint
	f.WriteTextRTL("שלום עולם זהו טקסט עברי בתוך מסגרת") //nolint

	if err := f.Close(); err != nil {
		log.Fatalf("close frame: %v", err)
	}
	y = f.CurrentY() + 24

	y = hline(y)

	// ── Section 6: Side-by-side LTR and RTL columns ───────────────────────
	y = section("6. Side-by-side columns — LTR left, RTL right", y)

	colW    := (contentW - 16) / 2
	colGap  := 16.0
	colYL   := y
	colYR   := y

	// Left column — Latin LTR
	fL := doc.NewFrame(pdf.FrameConfig{
		X: marginLeft, Y: colYL, Width: colW,
		Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
		Padding: pdf.UniformPadding(8),
	})
	fL.SetFont("latin", 10)  //nolint
	fL.SetTextColor(20, 20, 20)
	fL.WriteText("English (LTR)\n\nThe quick brown fox jumps over the lazy dog.") //nolint
	fL.Close()                                                                     //nolint

	// Right column — Arabic RTL
	fR := doc.NewFrame(pdf.FrameConfig{
		X: marginLeft + colW + colGap, Y: colYR, Width: colW,
		Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
		Padding: pdf.UniformPadding(8),
	})
	fR.SetFont("latin", 10)   //nolint
	fR.SetTextColor(100, 100, 100)
	fR.WriteText("Arabic (RTL)") //nolint
	fR.Advance(4)
	fR.SetFont("arabic", fontSize) //nolint
	fR.SetTextColor(20, 20, 20)
	fR.WriteTextRTL("الثعلب البني السريع يقفز فوق الكلب الكسول") //nolint
	fR.Close()                                                      //nolint

	// ── Write output ──────────────────────────────────────────────────────
	if err := doc.Save(*outPath); err != nil {
		log.Fatalf("save PDF: %v", err)
	}
	fmt.Printf("PDF written to %s  (%d pages)\n", *outPath, doc.PageCount())
	_ = setLatin
}
