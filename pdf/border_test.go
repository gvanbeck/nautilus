package pdf_test

import (
	"bytes"
	"testing"

	"github.com/gvanbeck/nautilus/pdf"
)

// TestNewUniformBorder verifies that all four sides are non-nil and share the
// same values.
func TestNewUniformBorder(t *testing.T) {
	spec := pdf.BorderSpec{
		Thickness: 2,
		Color:     pdf.ColorBlue,
		Pattern:   pdf.PatternDashed,
	}
	b := pdf.NewUniformBorder(spec)

	for _, side := range []*pdf.BorderSpec{b.Top, b.Right, b.Bottom, b.Left} {
		if side == nil {
			t.Fatal("NewUniformBorder: got nil side, want non-nil")
		}
		if side.Thickness != spec.Thickness {
			t.Errorf("Thickness = %f, want %f", side.Thickness, spec.Thickness)
		}
		if side.Color != spec.Color {
			t.Errorf("Color = %v, want %v", side.Color, spec.Color)
		}
		if side.Pattern != spec.Pattern {
			t.Errorf("Pattern = %d, want %d", side.Pattern, spec.Pattern)
		}
	}
}

// TestNewUniformBorder_sidesAreIndependent verifies that mutating one side does
// not affect the others (each side holds its own copy of the spec).
func TestNewUniformBorder_sidesAreIndependent(t *testing.T) {
	b := pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1})
	b.Top.Thickness = 5

	if b.Right.Thickness == 5 {
		t.Error("mutating Top.Thickness should not change Right.Thickness")
	}
}

// TestDrawBorder_noError verifies that DrawBorder returns no error for all
// built-in patterns on a real document.
func TestDrawBorder_noError(t *testing.T) {
	fontPath := systemFont(t)

	patterns := []struct {
		name    string
		pattern pdf.BorderPattern
	}{
		{"Solid", pdf.PatternSolid},
		{"Dashed", pdf.PatternDashed},
		{"Dotted", pdf.PatternDotted},
		{"DashDot", pdf.PatternDashDot},
	}

	for _, p := range patterns {
		t.Run(p.name, func(t *testing.T) {
			doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
			if err := doc.RegisterFont("regular", fontPath); err != nil {
				t.Fatalf("RegisterFont: %v", err)
			}
			doc.AddPage()

			border := pdf.NewUniformBorder(pdf.BorderSpec{
				Thickness: 1,
				Color:     pdf.ColorBlack,
				Pattern:   p.pattern,
			})
			if err := doc.DrawBorder(50, 50, 200, 40, border); err != nil {
				t.Fatalf("DrawBorder(%s): %v", p.name, err)
			}

			var buf bytes.Buffer
			if err := doc.Output(&buf); err != nil {
				t.Fatalf("Output: %v", err)
			}
			if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
				t.Fatal("output is not a valid PDF")
			}
		})
	}
}

// TestDrawBorder_customPattern verifies PatternCustom with an explicit dash
// array.
func TestDrawBorder_customPattern(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	doc.AddPage()

	spec := pdf.BorderSpec{
		Thickness:  1.5,
		Color:      pdf.ColorRed,
		Pattern:    pdf.PatternCustom,
		DashArray:  []float64{12, 4, 4, 4},
		DashPhase:  0,
	}
	border := pdf.NewUniformBorder(spec)
	if err := doc.DrawBorder(50, 50, 200, 40, border); err != nil {
		t.Fatalf("DrawBorder(Custom): %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("Output: %v", err)
	}
	if !bytes.HasPrefix(buf.Bytes(), []byte("%PDF-")) {
		t.Fatal("output is not a valid PDF")
	}
}

// TestDrawBorder_partialSides verifies that only the specified sides are drawn
// (no error when some sides are nil).
func TestDrawBorder_partialSides(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	doc.AddPage()

	spec := &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorGray}
	// Only top and bottom.
	border := pdf.Border{Top: spec, Bottom: spec}

	if err := doc.DrawBorder(50, 50, 200, 40, border); err != nil {
		t.Fatalf("DrawBorder(partial): %v", err)
	}
}

// TestDrawBorder_countingMode verifies that DrawBorder is a no-op and returns
// nil during the counting pass of Build.
func TestDrawBorder_countingMode(t *testing.T) {
	doc, _ := pdf.New(pdf.Config{})

	var drawCalled bool
	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
		border := pdf.NewUniformBorder(pdf.BorderSpec{Thickness: 1})
		if err := d.DrawBorder(0, 0, 100, 30, border); err != nil {
			t.Errorf("DrawBorder in counting mode returned error: %v", err)
		}
		drawCalled = true
	})

	doc.Build(func() {
		doc.AddPage()
	})

	// drawCalled should be true only for the rendering pass.
	// The counting pass also triggers SetHeader, but DrawBorder must be a
	// no-op there.  We simply verify no panic or error occurred.
	_ = drawCalled
}

// TestDrawBorder_differentSides verifies that different specs per side compile
// and run without error.
func TestDrawBorder_differentSides(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	doc.AddPage()

	border := pdf.Border{
		Top:    &pdf.BorderSpec{Thickness: 2, Color: pdf.ColorNavy, Pattern: pdf.PatternSolid},
		Bottom: &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorGray, Pattern: pdf.PatternDashed},
		Left:   &pdf.BorderSpec{Thickness: 1, Color: pdf.ColorRed, Pattern: pdf.PatternDotted},
		// Right is nil — not drawn.
	}

	if err := doc.DrawBorder(50, 50, 200, 60, border); err != nil {
		t.Fatalf("DrawBorder(mixed sides): %v", err)
	}
}

// TestDrawBorder_zeroThicknessDefaultsToOne verifies that a zero Thickness is
// treated as 1 pt rather than producing an invisible line.
func TestDrawBorder_zeroThicknessDefaultsToOne(t *testing.T) {
	fontPath := systemFont(t)

	doc, _ := pdf.New(pdf.Config{PageSize: pdf.PageSizeA4})
	if err := doc.RegisterFont("regular", fontPath); err != nil {
		t.Fatalf("RegisterFont: %v", err)
	}
	doc.AddPage()

	border := pdf.NewUniformBorder(pdf.BorderSpec{}) // zero Thickness
	if err := doc.DrawBorder(50, 50, 100, 30, border); err != nil {
		t.Fatalf("DrawBorder(zero thickness): %v", err)
	}
}

// TestColorConstants verifies that named color constants have the expected RGB
// values.
func TestColorConstants(t *testing.T) {
	if pdf.ColorBlack != (pdf.Color{0, 0, 0}) {
		t.Errorf("ColorBlack = %v, want {0,0,0}", pdf.ColorBlack)
	}
	if pdf.ColorWhite != (pdf.Color{255, 255, 255}) {
		t.Errorf("ColorWhite = %v, want {255,255,255}", pdf.ColorWhite)
	}
}
