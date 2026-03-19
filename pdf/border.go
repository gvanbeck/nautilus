package pdf

import "fmt"

// Color represents an RGB stroke or fill color.
// Each component is in the range [0, 255].
type Color struct {
	R, G, B uint8
}

// Common named colors for convenience.
var (
	ColorBlack     = Color{0, 0, 0}
	ColorWhite     = Color{255, 255, 255}
	ColorLightGray = Color{200, 200, 200}
	ColorGray      = Color{128, 128, 128}
	ColorDarkGray  = Color{64, 64, 64}
	ColorRed       = Color{220, 50, 50}
	ColorGreen     = Color{50, 180, 50}
	ColorBlue      = Color{50, 50, 220}
	ColorNavy      = Color{20, 20, 100}
	ColorOrange    = Color{230, 120, 20}
)

// BorderPattern defines the line style of a border side.
type BorderPattern int

const (
	// PatternSolid draws a continuous unbroken line.
	PatternSolid BorderPattern = iota

	// PatternDashed draws a long-dash / gap pattern.
	// Equivalent to PDF dash array [5] 2 d.
	PatternDashed

	// PatternDotted draws a short-dot / gap pattern.
	// Equivalent to PDF dash array [2 3] 11 d.
	PatternDotted

	// PatternDashDot alternates a long dash with a short dot.
	// Equivalent to PDF dash array [8 3 2 3] 0 d.
	PatternDashDot

	// PatternCustom uses the DashArray and DashPhase fields of BorderSpec.
	// The values are lengths in points: [draw, gap, draw, gap, …].
	PatternCustom
)

// BorderSpec defines the visual appearance of a single border side.
type BorderSpec struct {
	// Thickness is the line width in points.  Defaults to 1 when zero.
	Thickness float64

	// Color is the stroke color.  Defaults to black when zero-valued.
	Color Color

	// Pattern controls the line style.  Defaults to PatternSolid.
	Pattern BorderPattern

	// DashArray and DashPhase are used when Pattern == PatternCustom.
	// DashArray alternates drawn and gap lengths in points, e.g.
	// []float64{10, 4} means 10 pt dash, 4 pt gap repeating.
	// An empty DashArray with PatternCustom resets to a solid line.
	DashArray []float64
	DashPhase float64
}

// Border specifies the optional border drawn around a rectangular area.
// A nil side pointer means that side is not drawn.
//
// Use NewUniformBorder to create a border with identical sides.
// Build a Border literal to specify sides individually.
//
// Example — solid blue box, 1.5 pt:
//
//	doc.DrawBorder(50, 10, 495, 30, pdf.NewUniformBorder(pdf.BorderSpec{
//	    Thickness: 1.5,
//	    Color:     pdf.ColorBlue,
//	}))
//
// Example — dashed top and bottom only:
//
//	spec := &pdf.BorderSpec{Thickness: 0.5, Pattern: pdf.PatternDashed}
//	doc.DrawBorder(50, 10, 495, 30, pdf.Border{Top: spec, Bottom: spec})
type Border struct {
	Top    *BorderSpec
	Right  *BorderSpec
	Bottom *BorderSpec
	Left   *BorderSpec
}

// NewUniformBorder returns a Border where all four sides share the same
// appearance.  Each side gets its own copy of spec so sides can be mutated
// independently after creation.
func NewUniformBorder(spec BorderSpec) Border {
	t, r, b, l := spec, spec, spec, spec
	return Border{Top: &t, Right: &r, Bottom: &b, Left: &l}
}

// FillRect draws a solid filled rectangle at (x, y) with the given width and
// height using the specified fill color.  It is a no-op during the counting
// pass of Build.
func (d *Document) FillRect(x, y, w, h float64, color Color) {
	if d.countingMode {
		return
	}
	d.pdf.SetFillColor(color.R, color.G, color.B)
	d.pdf.RectFromUpperLeftWithStyle(x, y, w, h, "F")
}

// DrawBorder draws the border of the rectangle defined by top-left corner
// (x, y), width w, and height h on the current page.
//
// Only sides with a non-nil BorderSpec are drawn.  Each side can have a
// different thickness, color, and pattern.
//
// DrawBorder modifies the stroke color and line width of the graphics state.
// Call SetTextColor or any drawing method explicitly after DrawBorder when
// the surrounding code relies on a specific stroke state.
//
// DrawBorder is a no-op during the counting pass of Build.
//
// Example — thin gray separator below a header area:
//
//	doc.SetHeader(func(d *pdf.Document, info pdf.PageInfo) {
//	    // ... render header text ...
//	    spec := &pdf.BorderSpec{Thickness: 0.5, Color: pdf.ColorLightGray}
//	    d.DrawBorder(50, 10, 495, 30, pdf.Border{Bottom: spec})
//	})
func (d *Document) DrawBorder(x, y, w, h float64, border Border) error {
	if d.countingMode {
		return nil
	}

	type side struct {
		x1, y1, x2, y2 float64
		spec            *BorderSpec
	}

	// Define the four sides as line segments.
	sides := []side{
		{x, y, x + w, y, border.Top},           // top:    left → right
		{x + w, y, x + w, y + h, border.Right}, // right:  top  → bottom
		{x, y + h, x + w, y + h, border.Bottom},// bottom: left → right
		{x, y, x, y + h, border.Left},           // left:   top  → bottom
	}

	for _, s := range sides {
		if s.spec == nil {
			continue
		}
		if err := d.drawBorderSide(s.x1, s.y1, s.x2, s.y2, s.spec); err != nil {
			return err
		}
	}

	// Reset to solid line so subsequent drawing is not affected by a pattern
	// set by the last drawn side.
	d.pdf.SetLineType("") // empty string → PDF "[] 0 d" (solid)

	return nil
}

// drawBorderSide applies the spec's stroke settings and draws a single line.
func (d *Document) drawBorderSide(x1, y1, x2, y2 float64, spec *BorderSpec) error {
	thickness := spec.Thickness
	if thickness <= 0 {
		thickness = 1
	}

	d.pdf.SetLineWidth(thickness)
	d.pdf.SetStrokeColor(spec.Color.R, spec.Color.G, spec.Color.B)

	switch spec.Pattern {
	case PatternSolid:
		d.pdf.SetLineType("") // solid: [] 0 d
	case PatternDashed:
		d.pdf.SetLineType("dashed") // [5] 2 d
	case PatternDotted:
		d.pdf.SetLineType("dotted") // [2 3] 11 d
	case PatternDashDot:
		d.pdf.SetCustomLineType([]float64{8, 3, 2, 3}, 0)
	case PatternCustom:
		d.pdf.SetCustomLineType(spec.DashArray, spec.DashPhase)
	default:
		return fmt.Errorf("pdf: unknown border pattern %d", spec.Pattern)
	}

	d.pdf.Line(x1, y1, x2, y2)
	return nil
}
