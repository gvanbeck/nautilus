package layout

import "github.com/gvanbeck/nautilus/pdf"

// Spacer reserves a fixed amount of vertical space without rendering anything.
//
// Use it to add breathing room between paragraphs or after headings.
type Spacer struct {
	// Width is the desired width.  0 or negative means "use all available width".
	Width float64

	// Height is the vertical space to reserve in points.
	Height float64
}

func (s *Spacer) Wrap(_ *pdf.Document, availWidth, _ float64) (float64, float64) {
	w := s.Width
	if w <= 0 || w > availWidth {
		w = availWidth
	}
	return w, s.Height
}

func (s *Spacer) Draw(_ *pdf.Document, _, _ float64) error        { return nil }
func (s *Spacer) Split(_ *pdf.Document, _, _ float64) []Flowable  { return nil }
func (s *Spacer) SpaceBefore() float64                            { return 0 }
func (s *Spacer) SpaceAfter() float64                             { return 0 }
func (s *Spacer) KeepWithNext() bool                              { return false }
func (s *Spacer) MinWidth() float64                               { return 0 }

// ── HRFlowable ───────────────────────────────────────────────────────────────

// HRFlowable draws a horizontal rule as a solid filled bar.
//
// Example — a thin grey rule centred across 80 % of the available width:
//
//	&layout.HRFlowable{
//	    Width:     0.8,
//	    Thickness: 1,
//	    Color:     pdf.ColorGray,
//	    Align:     layout.AlignCenter,
//	    Before:    6,
//	    After:     6,
//	}
type HRFlowable struct {
	// Width controls the horizontal extent of the rule.
	//   > 1.0  — absolute width in points.
	//   0..1.0 — fraction of available width (e.g. 0.8 = 80 %).
	//   0      — full available width.
	Width float64

	// Thickness is the height of the rule bar in points.  Defaults to 1.
	Thickness float64

	// Color is the fill colour of the rule bar.
	Color pdf.Color

	// Align controls horizontal placement when Width < available width.
	Align HAlign

	// Before and After are extra whitespace above and below the rule.
	Before, After float64

	// computed by Wrap
	computedW float64
	computedX float64 // offset from the x passed to Draw
}

func (hr *HRFlowable) Wrap(_ *pdf.Document, availWidth, _ float64) (float64, float64) {
	w := hr.Width
	switch {
	case w <= 0:
		w = availWidth
	case w <= 1.0:
		w = availWidth * w
	}
	if w > availWidth {
		w = availWidth
	}
	hr.computedW = w

	switch hr.Align {
	case AlignCenter:
		hr.computedX = (availWidth - w) / 2
	case AlignRight:
		hr.computedX = availWidth - w
	default:
		hr.computedX = 0
	}

	t := hr.thickness()
	return availWidth, t
}

func (hr *HRFlowable) Draw(doc *pdf.Document, x, y float64) error {
	doc.FillRect(x+hr.computedX, y, hr.computedW, hr.thickness(), hr.Color)
	return nil
}

func (hr *HRFlowable) Split(_ *pdf.Document, _, _ float64) []Flowable { return nil }
func (hr *HRFlowable) SpaceBefore() float64                           { return hr.Before }
func (hr *HRFlowable) SpaceAfter() float64                            { return hr.After }
func (hr *HRFlowable) KeepWithNext() bool                             { return false }
func (hr *HRFlowable) MinWidth() float64                              { return 0 }

func (hr *HRFlowable) thickness() float64 {
	if hr.Thickness > 0 {
		return hr.Thickness
	}
	return 1
}
