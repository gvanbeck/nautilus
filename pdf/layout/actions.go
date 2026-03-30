package layout

import "github.com/gvanbeck/nautilus/pdf"

// ── PageBreak ────────────────────────────────────────────────────────────────

// PageBreak forces an immediate page break.
//
// An optional NextTemplate ID switches the page template starting from the
// new page.  Leave it empty to keep the current template.
//
// Example:
//
//	story = append(story, &layout.PageBreak{NextTemplate: "TwoColumn"})
type PageBreak struct {
	// NextTemplate, when non-empty, switches to this page template on the
	// new page.
	NextTemplate string
}

func (pb *PageBreak) Wrap(_ *pdf.Document, w, _ float64) (float64, float64) { return w, 0 }
func (pb *PageBreak) Draw(_ *pdf.Document, _, _ float64) error               { return nil }
func (pb *PageBreak) Split(_ *pdf.Document, _, _ float64) []Flowable         { return nil }
func (pb *PageBreak) SpaceBefore() float64                                   { return 0 }
func (pb *PageBreak) SpaceAfter() float64                                    { return 0 }
func (pb *PageBreak) KeepWithNext() bool                                     { return false }
func (pb *PageBreak) MinWidth() float64                                      { return 0 }
func (pb *PageBreak) apply(dt *DocTemplate) error {
	return dt.forcePageBreak(pb.NextTemplate)
}

// ── FrameBreak ───────────────────────────────────────────────────────────────

// FrameBreak forces the engine to advance to the next frame (or next page
// when the current frame is the last one on the page).
//
// It is also used internally by KeepTogether to shift content to a fresh frame.
type FrameBreak struct{}

func (fb *FrameBreak) Wrap(_ *pdf.Document, w, _ float64) (float64, float64) { return w, 0 }
func (fb *FrameBreak) Draw(_ *pdf.Document, _, _ float64) error               { return nil }
func (fb *FrameBreak) Split(_ *pdf.Document, _, _ float64) []Flowable         { return nil }
func (fb *FrameBreak) SpaceBefore() float64                                   { return 0 }
func (fb *FrameBreak) SpaceAfter() float64                                    { return 0 }
func (fb *FrameBreak) KeepWithNext() bool                                     { return false }
func (fb *FrameBreak) MinWidth() float64                                      { return 0 }
func (fb *FrameBreak) apply(dt *DocTemplate) error {
	return dt.forceFrameBreak()
}

// ── CondPageBreak ────────────────────────────────────────────────────────────

// CondPageBreak inserts a page break only when fewer than MinHeight points
// remain in the current frame.
//
// Use it to avoid starting a new section when only a sliver of the page is
// left.
//
// Example — break if less than 72 pt (1 inch) remains:
//
//	story = append(story, &layout.CondPageBreak{MinHeight: 72})
type CondPageBreak struct {
	// MinHeight is the minimum remaining frame height (in points) required
	// to avoid a page break.
	MinHeight float64
}

func (cb *CondPageBreak) Wrap(_ *pdf.Document, w, _ float64) (float64, float64) { return w, 0 }
func (cb *CondPageBreak) Draw(_ *pdf.Document, _, _ float64) error               { return nil }
func (cb *CondPageBreak) Split(_ *pdf.Document, _, _ float64) []Flowable         { return nil }
func (cb *CondPageBreak) SpaceBefore() float64                                   { return 0 }
func (cb *CondPageBreak) SpaceAfter() float64                                    { return 0 }
func (cb *CondPageBreak) KeepWithNext() bool                                     { return false }
func (cb *CondPageBreak) MinWidth() float64                                      { return 0 }
func (cb *CondPageBreak) apply(dt *DocTemplate) error {
	if dt.remainingFrameHeight() < cb.MinHeight {
		return dt.forcePageBreak("")
	}
	return nil
}

// ── NextPageTemplate ─────────────────────────────────────────────────────────

// NextPageTemplate schedules a template switch that takes effect on the NEXT
// page break.  The current page continues to use the existing template.
//
// This is how you implement "first page uses a title layout, subsequent pages
// use a two-column layout":
//
//	story = append(story,
//	    titleContent...,
//	    &layout.NextPageTemplate{TemplateID: "TwoColumn"},
//	    &layout.PageBreak{},
//	    bodyContent...,
//	)
type NextPageTemplate struct {
	// TemplateID is the ID of the PageTemplate to activate on the next page.
	TemplateID string
}

func (nt *NextPageTemplate) Wrap(_ *pdf.Document, w, _ float64) (float64, float64) { return w, 0 }
func (nt *NextPageTemplate) Draw(_ *pdf.Document, _, _ float64) error               { return nil }
func (nt *NextPageTemplate) Split(_ *pdf.Document, _, _ float64) []Flowable         { return nil }
func (nt *NextPageTemplate) SpaceBefore() float64                                   { return 0 }
func (nt *NextPageTemplate) SpaceAfter() float64                                    { return 0 }
func (nt *NextPageTemplate) KeepWithNext() bool                                     { return false }
func (nt *NextPageTemplate) MinWidth() float64                                      { return 0 }
func (nt *NextPageTemplate) apply(dt *DocTemplate) error {
	dt.scheduleTemplateSwitch(nt.TemplateID)
	return nil
}
