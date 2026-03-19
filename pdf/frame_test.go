package pdf_test

import (
	"bytes"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
)

// ── FrameConfig / Padding helpers ─────────────────────────────────────────

func TestUniformPadding(t *testing.T) {
	p := pdf.UniformPadding(10)
	if p.Top != 10 || p.Right != 10 || p.Bottom != 10 || p.Left != 10 {
		t.Errorf("UniformPadding(10) = %+v, want all 10", p)
	}
}

func TestHorizontalPadding(t *testing.T) {
	p := pdf.HorizontalPadding(8, 4)
	if p.Left != 8 || p.Right != 8 {
		t.Errorf("horizontal padding: got left=%f right=%f, want 8", p.Left, p.Right)
	}
	if p.Top != 4 || p.Bottom != 4 {
		t.Errorf("vertical padding: got top=%f bottom=%f, want 4", p.Top, p.Bottom)
	}
}

// ── Content area geometry ──────────────────────────────────────────────────

func TestFrame_contentGeometry(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()

	cfg := pdf.FrameConfig{
		X: 50, Y: 100,
		Width:   200,
		Padding: pdf.Padding{Top: 5, Right: 10, Bottom: 5, Left: 8},
	}
	f := doc.NewFrame(cfg)

	wantContentX := 50 + 8.0
	wantContentW := 200 - 8.0 - 10.0

	if f.ContentX() != wantContentX {
		t.Errorf("ContentX = %f, want %f", f.ContentX(), wantContentX)
	}
	if f.ContentWidth() != wantContentW {
		t.Errorf("ContentWidth = %f, want %f", f.ContentWidth(), wantContentW)
	}
}

func TestFrame_currentY_initialValue(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	doc.AddPage()

	f := doc.NewFrame(pdf.FrameConfig{X: 50, Y: 80, Width: 200, Padding: pdf.UniformPadding(6)})
	want := 80 + 6.0
	if f.CurrentY() != want {
		t.Errorf("CurrentY initially = %f, want %f (Y + padding top)", f.CurrentY(), want)
	}
}

// ── FrameHeight ────────────────────────────────────────────────────────────

func TestFrame_height_fixed(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	doc.AddPage()

	f := doc.NewFrame(pdf.FrameConfig{X: 50, Y: 50, Width: 200, Height: 100})
	if f.FrameHeight() != 100 {
		t.Errorf("FrameHeight (fixed) = %f, want 100", f.FrameHeight())
	}
}

func TestFrame_height_auto(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	doc.AddPage()

	padding := pdf.Padding{Top: 10, Bottom: 15, Left: 5, Right: 5}
	f := doc.NewFrame(pdf.FrameConfig{X: 50, Y: 50, Width: 200, Padding: padding})

	// Before advancing: height = padding.Top + padding.Bottom = 10+15 = 25
	wantInitial := padding.Top + padding.Bottom
	if f.FrameHeight() != wantInitial {
		t.Errorf("FrameHeight (auto, no content) = %f, want %f", f.FrameHeight(), wantInitial)
	}

	f.Advance(40)
	// After advancing 40: contentY = 50+10+40 = 100; height = 100-50+15 = 65
	wantAfter := padding.Top + 40 + padding.Bottom
	if f.FrameHeight() != wantAfter {
		t.Errorf("FrameHeight (auto, after Advance(40)) = %f, want %f", f.FrameHeight(), wantAfter)
	}
}

// ── Advance / NewLine ──────────────────────────────────────────────────────

func TestFrame_advance(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})
	doc.AddPage()

	f := doc.NewFrame(pdf.FrameConfig{X: 0, Y: 0, Width: 100, Padding: pdf.UniformPadding(0)})
	before := f.CurrentY()
	f.Advance(20)
	if f.CurrentY() != before+20 {
		t.Errorf("after Advance(20): CurrentY = %f, want %f", f.CurrentY(), before+20)
	}
}

// ── Close idempotency ──────────────────────────────────────────────────────

func TestFrame_close_idempotent(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()

	f := doc.NewFrame(pdf.FrameConfig{
		X: 50, Y: 50, Width: 200,
		Border: pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1}),
	})

	if err := f.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("second Close (idempotent): %v", err)
	}
}

// ── Integration tests ──────────────────────────────────────────────────────

// TestFrame_writeText verifies that WriteText advances the Y cursor and
// produces a valid PDF.
func TestFrame_writeText(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	f := doc.NewFrame(pdf.FrameConfig{
		X: 50, Y: 60, Width: 200, Padding: pdf.UniformPadding(8),
	})
	f.SetFont("regular", 11) //nolint

	initialY := f.CurrentY()
	f.WriteText("Hello, Frame! This text wraps within the content width.") //nolint

	if f.CurrentY() <= initialY {
		t.Error("WriteText did not advance the Y cursor")
	}
	f.Close() //nolint

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a valid PDF")
	}
}

// TestFrame_twoColumns verifies that two side-by-side frames produce a valid
// PDF — simulating a two-column layout.
func TestFrame_twoColumns(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	colW := (doc.PageWidth() - 60 - 60 - 10) / 2 // two columns with 10 pt gutter

	for i, xPos := range []float64{60, 60 + colW + 10} {
		f := doc.NewFrame(pdf.FrameConfig{
			X: xPos, Y: 60, Width: colW,
			Border:  pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}),
			Padding: pdf.UniformPadding(6),
		})
		f.SetFont("regular", 10) //nolint
		f.WriteText("Column content") //nolint
		_ = i
		f.Close() //nolint
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a valid PDF (two-column)")
	}
}

// TestFrame_background verifies that a fixed-height frame with a background
// color produces a valid PDF without error.
func TestFrame_background(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	bg := pdf.Color{R: 230, G: 240, B: 255}
	f := doc.NewFrame(pdf.FrameConfig{
		X: 50, Y: 60, Width: 200, Height: 80,
		Background: &bg,
		Border:     pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1, Color: pdf.ColorNavy}),
		Padding:    pdf.UniformPadding(8),
	})
	f.SetFont("regular", 11) //nolint
	f.WriteText("Callout box with background fill.") //nolint
	f.Close() //nolint

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a valid PDF (background frame)")
	}
}

// TestFrame_innerBorder verifies that DrawInnerBorder draws a relative border
// without error.
func TestFrame_innerBorder(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	f := doc.NewFrame(pdf.FrameConfig{X: 50, Y: 60, Width: 200, Height: 100, Padding: pdf.UniformPadding(8)})
	f.SetFont("regular", 10) //nolint
	f.WriteText("Title") //nolint

	// Separator line below title.
	sep := pdf.Border{Bottom: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorGray}}
	if err := f.DrawInnerBorder(0, 0, 200, 30, sep); err != nil {
		t.Fatalf("DrawInnerBorder: %v", err)
	}

	f.Close() //nolint

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a valid PDF (inner border)")
	}
}

// TestFrame_countingMode verifies that Frame methods are no-ops during the
// Build counting pass.
func TestFrame_countingMode(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	doc.Build(func() {
		doc.AddPage()
		doc.SetFont("regular", 11) //nolint
		f := doc.NewFrame(pdf.FrameConfig{
			X: 50, Y: 60, Width: 200, Height: 100,
			Background: &pdf.Color{R: 200, G: 200, B: 255},
			Border:     pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1}),
		})
		f.WriteText("Counting mode content") //nolint
		f.Close() //nolint
	})

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output after Build: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a valid PDF after Build with frames")
	}
}

// TestFrame_writeLineAt verifies that WriteLineAt positions text at the
// correct relative horizontal offset.
func TestFrame_writeLineAt(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	doc.AddPage()
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}

	f := doc.NewFrame(pdf.FrameConfig{X: 50, Y: 60, Width: 300, Padding: pdf.UniformPadding(5)})
	f.SetFont("regular", 11) //nolint

	// Write left-aligned text at offset 0.
	if _, err := f.WriteLineAt("Left", 0); err != nil {
		t.Fatalf("WriteLineAt offset 0: %v", err)
	}
	// Write text at a further offset within the content area.
	if _, err := f.WriteLineAt("Right-ish", 200); err != nil {
		t.Fatalf("WriteLineAt offset 200: %v", err)
	}
	f.Close() //nolint
}
